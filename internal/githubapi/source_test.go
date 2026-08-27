package githubapi

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// GitHub's real response when an enterprise policy blocks the token: status
// 200, a correct `issueCount`, and every node empty.
const policyBlockedResponse = `{
  "data": {
    "viewer": {"login": "alice"},
    "rateLimit": {"cost": 1, "remaining": 4999},
    "b_review_requested": {"pageInfo": {"hasNextPage": false}, "nodes": [null]}
  },
  "errors": [
    {"type": "FORBIDDEN",
     "message": "The 'AcmeCorp' enterprise forbids access via a fine-grained personal access tokens if the token's lifetime is greater than 366 days."}
  ]
}`

const publicSearchResponse = `{
  "total_count": 1,
  "items": [{
    "number": 42,
    "title": "Corrige o parser",
    "html_url": "https://github.com/acme/app/pull/42",
    "repository_url": "https://api.github.com/repos/acme/app",
    "node_id": "PR_kwABC",
    "draft": false,
    "comments": 3,
    "created_at": "2026-08-20T10:00:00Z",
    "updated_at": "2026-08-26T18:30:00Z",
    "user": {"login": "bob"},
    "assignees": []
  }]
}`

// This is the scenario public mode exists for: the token is valid, but the
// organization rejects it, and the same data is reachable without it.
func TestSourceFallsBackWhenOrganizationBlocksToken(t *testing.T) {
	graphql := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, policyBlockedResponse)
	}))
	defer graphql.Close()

	rest := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Only the review search returns a result, keeping the test predictable.
		if strings.Contains(r.URL.Query().Get("q"), "review-requested") {
			fmt.Fprint(w, publicSearchResponse)
			return
		}
		fmt.Fprint(w, `{"total_count":0,"items":[]}`)
	}))
	defer rest.Close()

	src := NewSource("github_pat_qualquer", "alice", nil, nil, 0, false)
	src.rich.SetEndpoint(graphql.URL)
	src.public.SetEndpoint(rest.URL)

	var switched bool
	src.OnModeChange = func(from, to Mode, reason string) { switched = true }

	snap, err := src.Fetch(context.Background())
	if err != nil {
		t.Fatalf("the fallback should have absorbed the block: %v", err)
	}
	if !switched {
		t.Error("the mode change should have been announced")
	}
	if src.Mode() != ModePublic {
		t.Errorf("mode = %s, want %s", src.Mode(), ModePublic)
	}
	if len(snap.PRs) != 1 {
		t.Fatalf("expected 1 pull request from public search, got %d", len(snap.PRs))
	}
	for _, pr := range snap.PRs {
		if pr.Repo != "acme/app" || pr.Number != 42 {
			t.Errorf("badly converted pull request: %+v", pr)
		}
		if !pr.InBucket(BucketReviewRequested) {
			t.Error("the pull request should be in the review bucket")
		}
	}
}

// An invalid credential differs from a policy: there is no fallback for it.
func TestBadCredentialsIsFatal(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, `{"message":"Bad credentials"}`)
	}))
	defer srv.Close()

	src := NewSource("ghp_invalido", "alice", nil, nil, 0, false)
	src.rich.SetEndpoint(srv.URL)

	if _, err := src.Fetch(context.Background()); err == nil {
		t.Fatal("an invalid credential should fail, not fall back to public mode")
	} else if !strings.Contains(err.Error(), "credential") {
		t.Errorf("unclear error: %v", err)
	}
}

// The repository filter must actually reach the query that is sent.
func TestRepoFilterReachesTheQuery(t *testing.T) {
	c := New("t")
	c.Repos = []string{"octocat/hello-world", "acme"}
	q := c.buildQuery()

	if !strings.Contains(q, "repo:octocat/hello-world") {
		t.Error("missing the repo: qualifier for owner/name")
	}
	if !strings.Contains(q, "org:acme") {
		t.Error("a bare name should become org:")
	}
}

func TestPublicQueryUsesExplicitQualifiers(t *testing.T) {
	c := NewPublic("alice")
	c.Repos = []string{"acme/app"}
	c.IgnoreAuthors = []string{"app/dependabot"}

	q := c.buildQuery("review-requested:alice")

	for _, want := range []string{"is:open", "is:pr", "archived:false",
		"review-requested:alice", "repo:acme/app", "-author:app/dependabot"} {
		if !strings.Contains(q, want) {
			t.Errorf("query missing %q: %s", want, q)
		}
	}
	// `involves:` would bring in pull requests the user merely commented on.
	if strings.Contains(q, "involves:") {
		t.Error("the search should not use involves:, which inflates results")
	}
}

func TestRepoFromURL(t *testing.T) {
	got := repoFromURL("https://api.github.com/repos/octocat/hello-world")
	if got != "octocat/hello-world" {
		t.Errorf("repoFromURL = %q", got)
	}
	if repoFromURL("lixo") != "" {
		t.Error("URL inesperada deveria devolver vazio")
	}
}

// Pull requests abandoned years ago stay open and drown out what matters.
// The filter must go into the search itself, not a later pass.
func TestAgeFilterReachesBothClients(t *testing.T) {
	now := time.Date(2026, 8, 27, 12, 0, 0, 0, time.UTC)

	if got := UpdatedSinceQualifier(5, now); got != " updated:>=2026-08-22" {
		t.Errorf("qualifier = %q", got)
	}
	if got := UpdatedSinceQualifier(0, now); got != "" {
		t.Errorf("zero should disable the filter, got %q", got)
	}
	if got := UpdatedSinceQualifier(-3, now); got != "" {
		t.Errorf("a negative value should disable the filter, got %q", got)
	}

	rich := New("t")
	rich.MaxAgeDays = 5
	if !strings.Contains(rich.buildQuery(), "updated:>=") {
		t.Error("the GraphQL query did not receive the age filter")
	}

	pub := NewPublic("alice")
	pub.MaxAgeDays = 5
	if !strings.Contains(pub.buildQuery("author:alice"), "updated:>=") {
		t.Error("the REST search did not receive the age filter")
	}
}

// On large teams, `review-requested:` matches requests sent to the whole
// team. On a real account that was the difference between 61 pull requests
// and 1: sixty were "your team may review", not "you were asked to review".
func TestReviewQualifierExcludesTeamRequestsByDefault(t *testing.T) {
	// GraphQL
	rich := New("t")
	q := rich.buildQuery()
	if !strings.Contains(q, "user-review-requested:@me") {
		t.Error("the default should ask only for reviews requested by name")
	}
	if strings.Contains(q, " review-requested:@me") {
		t.Error("the broad qualifier leaked into the default query")
	}

	rich.IncludeTeamReviews = true
	if !strings.Contains(rich.buildQuery(), "review-requested:@me") {
		t.Error("with IncludeTeamReviews the broad qualifier should return")
	}

	// Busca REST
	pub := NewPublic("alice")
	if q := pub.buildQuery("user-review-requested:alice"); !strings.Contains(q, "user-review-requested") {
		t.Errorf("REST search: %s", q)
	}
}

func TestBucketQueryHonorsTeamFlag(t *testing.T) {
	if got := BucketReviewRequested.query(false); !strings.Contains(got, "user-review-requested") {
		t.Errorf("default = %q", got)
	}
	if got := BucketReviewRequested.query(true); strings.Contains(got, "user-review-requested") {
		t.Errorf("with teams = %q, should use the broad qualifier", got)
	}
	// The other buckets do not change with the flag.
	if BucketAuthored.query(false) != BucketAuthored.query(true) {
		t.Error("only the review bucket depends on the team flag")
	}
}
