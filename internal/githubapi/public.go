package githubapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// PublicClient reads pull requests through the REST search API, optionally
// without authentication.
//
// It exists for a concrete reason: an organization may have an enterprise
// policy that blocks the user's token — GitHub answers FORBIDDEN and returns
// empty nodes — even when the repository is public and readable by anyone.
// Without a token, the same data comes back normally.
//
// The price is losing what only authenticated GraphQL delivers: CI state,
// review decision and merge status.
type PublicClient struct {
	http     *http.Client
	endpoint string

	// Token is optional. Empty means an anonymous request.
	Token string
	// Login is the user whose pull requests matter. Without a token there is
	// no @me, so the login must be explicit.
	Login string
	// Repos limits the search; empty searches all of GitHub.
	Repos []string
	// IgnoreAuthors removes bot-authored pull requests.
	IgnoreAuthors []string
	// MaxAgeDays drops pull requests with no recent activity. Zero disables.
	MaxAgeDays int
	// IncludeTeamReviews includes review requests sent to your team.
	IncludeTeamReviews bool

	// Rate limit state, read from the headers on every response.
	rateRemaining int
	rateReset     time.Time
	rateKnown     bool
}

// RateRemaining reports how many searches still fit in the current window.
// It returns -1 until a response has been observed.
func (c *PublicClient) RateRemaining() int {
	if !c.rateKnown {
		return -1
	}
	if time.Now().After(c.rateReset) {
		return -1
	}
	return c.rateRemaining
}

// NewPublic creates the REST search client.
func NewPublic(login string) *PublicClient {
	return &PublicClient{
		http:     &http.Client{Timeout: 30 * time.Second},
		endpoint: "https://api.github.com/search/issues",
		Login:    login,
	}
}

// SetEndpoint points at another server, used by tests.
func (c *PublicClient) SetEndpoint(u string) { c.endpoint = u }

// restItem is one result from the REST issue search.
type restItem struct {
	Number    int    `json:"number"`
	Title     string `json:"title"`
	HTMLURL   string `json:"html_url"`
	State     string `json:"state"`
	Draft     bool   `json:"draft"`
	Comments  int    `json:"comments"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	NodeID    string `json:"node_id"`
	RepoURL   string `json:"repository_url"`
	User      *struct {
		Login string `json:"login"`
	} `json:"user"`
	Assignees []struct {
		Login string `json:"login"`
	} `json:"assignees"`
	PullRequest *struct {
		HTMLURL string `json:"html_url"`
	} `json:"pull_request"`
}

type restSearchResponse struct {
	TotalCount        int        `json:"total_count"`
	IncompleteResults bool       `json:"incomplete_results"`
	Items             []restItem `json:"items"`
	Message           string     `json:"message"`
}

// Fetch builds the picture with one explicit search per bucket.
//
// Using exact qualifiers instead of a broad `involves:` avoids a subtle bug:
// `involves:` also matches pull requests you merely commented on once, which
// would inflate "you were mentioned" with dozens of irrelevant results.
//
// That is four requests per cycle. At a two-minute interval it works out to
// two per minute, against an anonymous limit of ten per minute.
func (c *PublicClient) Fetch(ctx context.Context) (*Snapshot, error) {
	if c.Login == "" {
		return nil, fmt.Errorf("a GitHub login is required for public search")
	}

	snap := &Snapshot{
		Viewer:    c.Login,
		PRs:       map[string]*PullRequest{},
		FetchedAt: time.Now().UTC(),
	}

	reviewQualifier := "user-review-requested:"
	if c.IncludeTeamReviews {
		reviewQualifier = "review-requested:"
	}

	searches := []struct {
		qualifier string
		bucket    Bucket
	}{
		{reviewQualifier + c.Login, BucketReviewRequested},
		{"author:" + c.Login, BucketAuthored},
		{"assignee:" + c.Login, BucketAssigned},
		{"mentions:" + c.Login, BucketMentioned},
	}

	for _, s := range searches {
		items, err := c.search(ctx, c.buildQuery(s.qualifier))
		if err != nil {
			return nil, err
		}
		for _, it := range items {
			c.merge(snap, it, s.bucket)
		}
	}

	return snap, nil
}

// buildQuery adds the shared filters: open pull requests only, no archived
// repositories, limited to the chosen repositories, and no bots.
func (c *PublicClient) buildQuery(who string) string {
	parts := []string{"is:open", "is:pr", "archived:false", who}
	for _, r := range c.Repos {
		if r = strings.TrimSpace(r); r == "" {
			continue
		}
		// A value without "/" is treated as a whole organization.
		if strings.Contains(r, "/") {
			parts = append(parts, "repo:"+r)
		} else {
			parts = append(parts, "org:"+r)
		}
	}
	for _, a := range c.IgnoreAuthors {
		parts = append(parts, "-author:"+a)
	}
	if q := UpdatedSinceQualifier(c.MaxAgeDays, time.Now()); q != "" {
		parts = append(parts, strings.TrimSpace(q))
	}
	return strings.Join(parts, " ")
}

func (c *PublicClient) merge(snap *Snapshot, it restItem, bucket Bucket) {
	id := it.NodeID
	if id == "" {
		id = it.HTMLURL
	}
	pr, exists := snap.PRs[id]
	if !exists {
		pr = &PullRequest{
			ID:       id,
			Repo:     repoFromURL(it.RepoURL),
			Number:   it.Number,
			Title:    strings.TrimSpace(it.Title),
			URL:      it.HTMLURL,
			IsDraft:  it.Draft,
			Comments: it.Comments,
		}
		if it.User != nil {
			pr.Author = it.User.Login
		}
		pr.CreatedAt, _ = time.Parse(time.RFC3339, it.CreatedAt)
		pr.UpdatedAt, _ = time.Parse(time.RFC3339, it.UpdatedAt)
		snap.PRs[id] = pr
	}
	if !pr.InBucket(bucket) {
		pr.Buckets = append(pr.Buckets, bucket)
	}
}

// search runs a paginated search up to 100 results, which is more than
// enough for one person's pull requests.
func (c *PublicClient) search(ctx context.Context, query string) ([]restItem, error) {
	u := fmt.Sprintf("%s?q=%s&per_page=100&sort=updated&order=desc",
		c.endpoint, url.QueryEscape(query))

	const maxAttempts = 3
	var lastErr error

	for attempt := range maxAttempts {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(time.Duration(1<<attempt) * time.Second):
			}
		}
		if err := c.waitForBudget(ctx); err != nil {
			return nil, err
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Accept", "application/vnd.github+json")
		req.Header.Set("User-Agent", userAgent)
		if c.Token != "" {
			req.Header.Set("Authorization", "Bearer "+c.Token)
		}

		resp, err := c.http.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		body, readErr := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
		resp.Body.Close()
		if readErr != nil {
			lastErr = readErr
			continue
		}

		c.observeRate(resp.Header)

		switch {
		case resp.StatusCode == http.StatusOK:
			var parsed restSearchResponse
			if err := json.Unmarshal(body, &parsed); err != nil {
				return nil, fmt.Errorf("unreadable search response: %w", err)
			}
			return parsed.Items, nil

		case resp.StatusCode == http.StatusUnauthorized:
			return nil, &AuthError{Msg: apiMessage(body)}

		case resp.StatusCode == http.StatusForbidden, resp.StatusCode == http.StatusTooManyRequests:
			if remaining := resp.Header.Get("X-RateLimit-Remaining"); remaining == "0" {
				return nil, &RateLimitError{RetryAfter: retryAfter(resp), Msg: apiMessage(body)}
			}
			return nil, fmt.Errorf("search denied: %s", apiMessage(body))

		case resp.StatusCode >= 500:
			lastErr = fmt.Errorf("GitHub returned %d", resp.StatusCode)

		default:
			return nil, fmt.Errorf("GitHub returned %d: %s", resp.StatusCode, apiMessage(body))
		}
	}
	return nil, fmt.Errorf("search failed after %d attempts: %w", maxAttempts, lastErr)
}

// repoFromURL extracts "owner/repo" from the repository API URL, avoiding an
// extra request just to learn the name.
func repoFromURL(apiURL string) string {
	const marker = "/repos/"
	i := strings.Index(apiURL, marker)
	if i < 0 {
		return ""
	}
	return strings.Trim(apiURL[i+len(marker):], "/")
}

// observeRate records what the headers say about the remaining budget, so the
// client can slow down before earning a 403.
func (c *PublicClient) observeRate(h http.Header) {
	if v := h.Get("X-RateLimit-Remaining"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.rateRemaining = n
			c.rateKnown = true
		}
	}
	if v := h.Get("X-RateLimit-Reset"); v != "" {
		if ts, err := strconv.ParseInt(v, 10, 64); err == nil {
			c.rateReset = time.Unix(ts, 0)
		}
	}
}

// waitForBudget holds the next search when the window's budget is spent.
//
// GitHub search is limited per minute — ten without a token, thirty with one —
// and waiting for the reset beats taking a 403 and losing the cycle.
func (c *PublicClient) waitForBudget(ctx context.Context) error {
	if !c.rateKnown || c.rateRemaining > 0 {
		return nil
	}
	wait := time.Until(c.rateReset)
	if wait <= 0 || wait > 2*time.Minute {
		return nil
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(wait + time.Second):
		c.rateRemaining = 1 // measured optimism: the window has rolled over
		return nil
	}
}
