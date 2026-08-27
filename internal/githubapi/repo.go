package githubapi

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// RepoPullRequests lists every open pull request in one repository,
// regardless of who they belong to.
//
// This is a deliberate exception to the rest of the package, which only ever
// looks at what is tied to the authenticated user. It exists for on-demand
// queries — "what is open on this project right now" — and never feeds the
// alerting loop, which would otherwise start reporting other people's work.
func (c *Client) RepoPullRequests(ctx context.Context, repo string, limit int) ([]*PullRequest, error) {
	if !strings.Contains(repo, "/") {
		return nil, fmt.Errorf("expected owner/name, got %q", repo)
	}
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	query := fmt.Sprintf(`query {
  search(query: %q, type: ISSUE, first: %d) {
    issueCount
    nodes { ...prFields }
  }
}
%s`, "is:open is:pr archived:false repo:"+repo, limit, prFieldsFragment)

	body, err := c.do(ctx, query)
	if err != nil {
		return nil, err
	}

	var resp graphQLResponse
	if err := unmarshalGraphQL(body, &resp); err != nil {
		return nil, err
	}
	if err := graphQLError(resp); err != nil {
		return nil, err
	}

	var result searchResult
	if err := unmarshalRaw(resp.Data["search"], &result); err != nil {
		return nil, fmt.Errorf("unreadable search response: %w", err)
	}

	out := make([]*PullRequest, 0, len(result.Nodes))
	for _, n := range result.Nodes {
		if n.ID == "" {
			continue
		}
		out = append(out, convert(n))
	}
	return out, nil
}

// RepoPullRequestsPublic is the same query through REST search, for when the
// GraphQL path is unavailable.
func (c *PublicClient) RepoPullRequests(ctx context.Context, repo string, limit int) ([]*PullRequest, error) {
	if !strings.Contains(repo, "/") {
		return nil, fmt.Errorf("expected owner/name, got %q", repo)
	}

	items, err := c.search(ctx, "is:open is:pr archived:false repo:"+repo)
	if err != nil {
		return nil, err
	}
	if limit > 0 && len(items) > limit {
		items = items[:limit]
	}

	out := make([]*PullRequest, 0, len(items))
	for _, it := range items {
		pr := &PullRequest{
			ID:       it.NodeID,
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
		out = append(out, pr)
	}
	return out, nil
}

// RepoPullRequests on the Source picks whichever path is currently working.
func (s *Source) RepoPullRequests(ctx context.Context, repo string, limit int) ([]*PullRequest, error) {
	if s.mode == ModeRich && s.rich != nil {
		prs, err := s.rich.RepoPullRequests(ctx, repo, limit)
		if err == nil {
			return prs, nil
		}
		var policyErr *PolicyError
		if !asPolicy(err, &policyErr) {
			return nil, err
		}
		s.switchMode(ModePublic, policyErr.Msg)
	}
	return s.public.RepoPullRequests(ctx, repo, limit)
}

// ResolveRepo turns what someone typed into an "owner/name" pair.
//
// People refer to a project by its short name, not its full path, so a bare
// word is matched against the repositories already being watched before being
// rejected. That way `/prs myproject` works without anyone having to remember
// the owner.
func ResolveRepo(input string, watched []string) (string, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", fmt.Errorf("no repository given")
	}
	if strings.Contains(input, "/") {
		return input, nil
	}

	var matches []string
	for _, w := range watched {
		if !strings.Contains(w, "/") {
			continue
		}
		name := w[strings.Index(w, "/")+1:]
		if strings.EqualFold(name, input) || strings.EqualFold(w, input) {
			return w, nil
		}
		if strings.Contains(strings.ToLower(w), strings.ToLower(input)) {
			matches = append(matches, w)
		}
	}

	switch len(matches) {
	case 1:
		return matches[0], nil
	case 0:
		return "", fmt.Errorf("no watched repository matches %q\n"+
			"  Use the full owner/name, or add it to WATCH_REPOS", input)
	default:
		return "", fmt.Errorf("%q matches several repositories: %s\n"+
			"  Use the full owner/name", input, strings.Join(matches, ", "))
	}
}
