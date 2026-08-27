package githubapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	defaultEndpoint = "https://api.github.com/graphql"
	perBucket       = 50
	userAgent       = "ptal/0.1 (+https://github.com/Cristhianzl/telegram-PTAL)"
)

// Client talks to GitHub's GraphQL API on behalf of a single user.
type Client struct {
	token    string
	endpoint string
	http     *http.Client

	// IgnoreAuthors is applied inside the search itself, so bot pull
	// requests never travel over the wire.
	IgnoreAuthors []string
	// Repos limits the search to repositories ("owner/name") or whole
	// organizations (bare name). Empty searches all of GitHub.
	Repos []string
	// MaxAgeDays drops pull requests with no recent activity. Zero disables.
	MaxAgeDays int
	// IncludeTeamReviews includes pull requests where the review was
	// requested from your team rather than from you by name. Off by default:
	// on large teams this is the single largest source of noise.
	IncludeTeamReviews bool
}

// New creates a client with a timeout suited to a long-running daemon.
func New(token string) *Client {
	return &Client{
		token:    token,
		endpoint: defaultEndpoint,
		http:     &http.Client{Timeout: 30 * time.Second},
	}
}

// SetEndpoint points the client at another server. It exists so tests can
// exercise the whole pipeline against a mock.
func (c *Client) SetEndpoint(url string) { c.endpoint = url }

// buildQuery assembles a single query with one search alias per bucket.
// Six searches in one request cost about 5 of the 5,000 points per hour.
func (c *Client) buildQuery() string {
	var exclude strings.Builder
	for _, r := range c.Repos {
		if r = strings.TrimSpace(r); r == "" {
			continue
		}
		if strings.Contains(r, "/") {
			exclude.WriteString(" repo:")
		} else {
			exclude.WriteString(" org:")
		}
		exclude.WriteString(r)
	}
	for _, a := range c.IgnoreAuthors {
		exclude.WriteString(" -author:")
		exclude.WriteString(a)
	}
	exclude.WriteString(UpdatedSinceQualifier(c.MaxAgeDays, time.Now()))

	var b strings.Builder
	b.WriteString("query {\n  viewer { login }\n  rateLimit { limit cost remaining resetAt }\n")
	for _, bucket := range AllBuckets {
		fmt.Fprintf(&b,
			"  b_%s: search(query: %q, type: ISSUE, first: %d) {\n"+
				"    pageInfo { hasNextPage }\n    nodes { ...prFields }\n  }\n",
			bucket, bucket.query(c.IncludeTeamReviews)+exclude.String(), perBucket)
	}
	b.WriteString("}\n")
	b.WriteString(prFieldsFragment)
	return b.String()
}

const prFieldsFragment = `
fragment prFields on PullRequest {
  id
  number
  title
  url
  isDraft
  createdAt
  updatedAt
  additions
  deletions
  repository { nameWithOwner }
  author { login }
  reviewDecision
  mergeable
  comments { totalCount }
  reviews { totalCount }
  commits(last: 1) {
    nodes { commit { statusCheckRollup { state } } }
  }
}
`

// Raw GraphQL response, mirroring the fragment above.
type rawPR struct {
	ID         string `json:"id"`
	Number     int    `json:"number"`
	Title      string `json:"title"`
	URL        string `json:"url"`
	IsDraft    bool   `json:"isDraft"`
	CreatedAt  string `json:"createdAt"`
	UpdatedAt  string `json:"updatedAt"`
	Additions  int    `json:"additions"`
	Deletions  int    `json:"deletions"`
	Repository struct {
		NameWithOwner string `json:"nameWithOwner"`
	} `json:"repository"`
	Author *struct {
		Login string `json:"login"`
	} `json:"author"`
	ReviewDecision string `json:"reviewDecision"`
	Mergeable      string `json:"mergeable"`
	Comments       struct {
		TotalCount int `json:"totalCount"`
	} `json:"comments"`
	Reviews struct {
		TotalCount int `json:"totalCount"`
	} `json:"reviews"`
	Commits struct {
		Nodes []struct {
			Commit struct {
				StatusCheckRollup *struct {
					State string `json:"state"`
				} `json:"statusCheckRollup"`
			} `json:"commit"`
		} `json:"nodes"`
	} `json:"commits"`
}

type searchResult struct {
	PageInfo struct {
		HasNextPage bool `json:"hasNextPage"`
	} `json:"pageInfo"`
	Nodes []rawPR `json:"nodes"`
}

type graphQLResponse struct {
	Data   map[string]json.RawMessage `json:"data"`
	Errors []struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"errors"`
}

// Viewer returns the login of the token's owner. Used by `doctor` and to
// make clear whose pull requests the messages are about.
func (c *Client) Viewer(ctx context.Context) (string, error) {
	body, err := c.do(ctx, `query { viewer { login } }`)
	if err != nil {
		return "", err
	}
	var resp graphQLResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", err
	}
	if err := graphQLError(resp); err != nil {
		return "", err
	}
	var v struct {
		Login string `json:"login"`
	}
	if err := json.Unmarshal(resp.Data["viewer"], &v); err != nil {
		return "", err
	}
	return v.Login, nil
}

// Fetch retrieves the full picture of pull requests tied to the token's
// user. A pull request appearing in more than one bucket is recorded once,
// accumulating every bucket it belongs to.
func (c *Client) Fetch(ctx context.Context) (*Snapshot, error) {
	body, err := c.do(ctx, c.buildQuery())
	if err != nil {
		return nil, err
	}

	var resp graphQLResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unreadable response from GitHub: %w", err)
	}
	if err := graphQLError(resp); err != nil {
		return nil, err
	}

	snap := &Snapshot{
		PRs:       map[string]*PullRequest{},
		FetchedAt: time.Now().UTC(),
	}

	if raw, ok := resp.Data["viewer"]; ok {
		var v struct {
			Login string `json:"login"`
		}
		if json.Unmarshal(raw, &v) == nil {
			snap.Viewer = v.Login
		}
	}
	if raw, ok := resp.Data["rateLimit"]; ok {
		var rl struct {
			Cost      int `json:"cost"`
			Remaining int `json:"remaining"`
		}
		if json.Unmarshal(raw, &rl) == nil {
			snap.RateCost, snap.RateRemaining = rl.Cost, rl.Remaining
		}
	}

	for _, bucket := range AllBuckets {
		raw, ok := resp.Data["b_"+string(bucket)]
		if !ok {
			continue
		}
		var sr searchResult
		if err := json.Unmarshal(raw, &sr); err != nil {
			return nil, fmt.Errorf("unreadable bucket %s: %w", bucket, err)
		}
		for _, n := range sr.Nodes {
			// Searching with type:ISSUE returns empty nodes for plain issues;
			// the fragment only fills PullRequest, so an empty ID filters them.
			if n.ID == "" {
				continue
			}
			pr, exists := snap.PRs[n.ID]
			if !exists {
				pr = convert(n)
				snap.PRs[n.ID] = pr
			}
			if !pr.InBucket(bucket) {
				pr.Buckets = append(pr.Buckets, bucket)
			}
		}
	}

	return snap, nil
}

func convert(n rawPR) *PullRequest {
	pr := &PullRequest{
		ID:             n.ID,
		Repo:           n.Repository.NameWithOwner,
		Number:         n.Number,
		Title:          strings.TrimSpace(n.Title),
		URL:            n.URL,
		IsDraft:        n.IsDraft,
		ReviewDecision: n.ReviewDecision,
		Mergeable:      n.Mergeable,
		Additions:      n.Additions,
		Deletions:      n.Deletions,
		Comments:       n.Comments.TotalCount,
		Reviews:        n.Reviews.TotalCount,
	}
	if n.Author != nil {
		pr.Author = n.Author.Login
	}
	pr.CreatedAt, _ = time.Parse(time.RFC3339, n.CreatedAt)
	pr.UpdatedAt, _ = time.Parse(time.RFC3339, n.UpdatedAt)
	if len(n.Commits.Nodes) > 0 {
		if r := n.Commits.Nodes[0].Commit.StatusCheckRollup; r != nil {
			pr.Checks = r.State
		}
	}
	return pr
}

// do performs the request, retrying transient failures. Credential errors
// are never retried: repeating a 401 never helps.
func (c *Client) do(ctx context.Context, query string) ([]byte, error) {
	payload, err := json.Marshal(map[string]string{"query": query})
	if err != nil {
		return nil, err
	}

	const maxAttempts = 4
	var lastErr error

	for attempt := range maxAttempts {
		if attempt > 0 {
			wait := time.Duration(1<<attempt) * time.Second
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(wait):
			}
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(payload))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+c.token)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", userAgent)

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

		switch {
		case resp.StatusCode == http.StatusOK:
			return body, nil

		case resp.StatusCode == http.StatusUnauthorized:
			return nil, &AuthError{Msg: apiMessage(body)}

		case resp.StatusCode == http.StatusForbidden, resp.StatusCode == http.StatusTooManyRequests:
			// A 403 from GitHub is almost always a secondary rate limit,
			// not a permission problem.
			return nil, &RateLimitError{RetryAfter: retryAfter(resp), Msg: apiMessage(body)}

		case resp.StatusCode >= 500:
			lastErr = fmt.Errorf("GitHub returned %d", resp.StatusCode)

		default:
			return nil, fmt.Errorf("GitHub returned %d: %s", resp.StatusCode, apiMessage(body))
		}
	}
	return nil, fmt.Errorf("GitHub unreachable after %d attempts: %w", maxAttempts, lastErr)
}

func retryAfter(resp *http.Response) time.Duration {
	if v := resp.Header.Get("Retry-After"); v != "" {
		if secs, err := strconv.Atoi(v); err == nil {
			return time.Duration(secs) * time.Second
		}
	}
	if v := resp.Header.Get("X-RateLimit-Reset"); v != "" {
		if ts, err := strconv.ParseInt(v, 10, 64); err == nil {
			if d := time.Until(time.Unix(ts, 0)); d > 0 {
				return d
			}
		}
	}
	return time.Minute
}

func apiMessage(body []byte) string {
	var e struct {
		Message string `json:"message"`
	}
	if json.Unmarshal(body, &e) == nil && e.Message != "" {
		return e.Message
	}
	if len(body) > 200 {
		body = body[:200]
	}
	return strings.TrimSpace(string(body))
}

// graphQLError converts the errors the API returns inside a 200 response.
func graphQLError(resp graphQLResponse) error {
	if len(resp.Errors) == 0 {
		return nil
	}
	msgs := make([]string, 0, len(resp.Errors))
	for _, e := range resp.Errors {
		msgs = append(msgs, e.Message)
		low := strings.ToLower(e.Message)
		if strings.Contains(low, "bad credentials") {
			return &AuthError{Msg: e.Message}
		}
		// Enterprise policy: the token is valid, but the organization
		// rejects this kind of token. The data may still be publicly
		// readable, so this is not fatal.
		if e.Type == "FORBIDDEN" &&
			(strings.Contains(low, "forbids access") ||
				strings.Contains(low, "fine-grained") ||
				strings.Contains(low, "personal access token")) {
			return &PolicyError{Msg: e.Message}
		}
		if e.Type == "FORBIDDEN" {
			return &AuthError{Msg: e.Message}
		}
	}
	return &GraphQLErrors{Messages: msgs}
}
