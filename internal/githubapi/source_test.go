package githubapi

import (
	"context"
	"errors"
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

// People refer to a project by its short name, not its full path. Matching
// against the watched repositories is what makes `/prs myproject` work
// without anyone having to remember the owner.
func TestResolveRepoAcceptsShortNames(t *testing.T) {
	watched := []string{"octocat/hello-world", "acme/api", "acme"}

	cases := []struct{ input, want string }{
		{"octocat/hello-world", "octocat/hello-world"}, // already complete
		{"hello-world", "octocat/hello-world"},         // bare name
		{"HELLO-WORLD", "octocat/hello-world"},         // case-insensitive
		{"api", "acme/api"},
		{"other/repo", "other/repo"}, // unwatched but explicit
	}
	for _, c := range cases {
		got, err := ResolveRepo(c.input, watched)
		if err != nil {
			t.Errorf("ResolveRepo(%q) errored: %v", c.input, err)
			continue
		}
		if got != c.want {
			t.Errorf("ResolveRepo(%q) = %q, want %q", c.input, got, c.want)
		}
	}
}

// An ambiguous short name must ask rather than guess: picking one silently
// would show the wrong project's pull requests.
func TestResolveRepoRefusesToGuess(t *testing.T) {
	watched := []string{"acme/api-gateway", "acme/api-worker"}

	_, err := ResolveRepo("api", watched)

	if err == nil {
		t.Fatal("an ambiguous name should be refused, not guessed")
	}
	for _, want := range []string{"api-gateway", "api-worker"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the error should list the candidates, got: %v", err)
		}
	}
}

func TestResolveRepoRejectsUnknownNames(t *testing.T) {
	if _, err := ResolveRepo("nothing", []string{"acme/api"}); err == nil {
		t.Error("an unmatched bare name should be an error")
	}
	if _, err := ResolveRepo("", nil); err == nil {
		t.Error("an empty name should be an error")
	}
}

// The repository query deliberately carries no user filter: asking about a
// project means asking about everyone's work on it.
func TestRepoQueryHasNoUserFilter(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("q")
		for _, forbidden := range []string{"@me", "author:", "review-requested:", "mentions:"} {
			if strings.Contains(q, forbidden) {
				t.Errorf("the repository query must not filter by user, found %q in %q", forbidden, q)
			}
		}
		if !strings.Contains(q, "repo:acme/api") {
			t.Errorf("the query should scope to the repository: %q", q)
		}
		fmt.Fprint(w, `{"total_count":0,"items":[]}`)
	}))
	defer srv.Close()

	c := NewPublic("alice")
	c.SetEndpoint(srv.URL)

	if _, err := c.RepoPullRequests(context.Background(), "acme/api", "", 10); err != nil {
		t.Fatal(err)
	}
}

func TestRepoPullRequestsRejectsBareName(t *testing.T) {
	if _, err := NewPublic("alice").RepoPullRequests(context.Background(), "api", "", 10); err == nil {
		t.Error("a name without an owner should be rejected before any request")
	}
}

// A second argument narrows the listing to one author, which is what makes
// "show me mine in this project" possible.
func TestRepoQueryFiltersByAuthor(t *testing.T) {
	if got := repoQuery("acme/api", ""); strings.Contains(got, "author:") {
		t.Errorf("without an author there should be no filter: %q", got)
	}
	got := repoQuery("acme/api", "alice")
	if !strings.Contains(got, "author:alice") {
		t.Errorf("query = %q, want an author filter", got)
	}
	if !strings.Contains(got, "repo:acme/api") {
		t.Errorf("query = %q, should still scope to the repository", got)
	}
}

// "me" is resolved locally, because the public search path is
// unauthenticated and GitHub has no @me to expand there.
func TestMeResolvesToTheConfiguredLogin(t *testing.T) {
	var seen string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = r.URL.Query().Get("q")
		fmt.Fprint(w, `{"total_count":0,"items":[]}`)
	}))
	defer srv.Close()

	src := NewSource("", "alice", nil, nil, 0, false)
	src.public.SetEndpoint(srv.URL)

	if _, err := src.RepoPullRequests(context.Background(), "acme/api", "me", 10); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(seen, "author:alice") {
		t.Errorf("\"me\" should expand to the configured login, query was %q", seen)
	}
	if strings.Contains(seen, "@me") {
		t.Errorf("@me must not reach an unauthenticated search: %q", seen)
	}
}

func TestMeWithoutALoginIsAnError(t *testing.T) {
	src := NewSource("", "", nil, nil, 0, false)

	_, err := src.RepoPullRequests(context.Background(), "acme/api", "me", 10)

	if err == nil {
		t.Fatal("resolving \"me\" without a login should fail")
	}
	if !strings.Contains(err.Error(), "GITHUB_LOGIN") {
		t.Errorf("the error should say how to fix it: %v", err)
	}
}

// Degrading used to be permanent. A daemon starting at boot finds the system
// keyring locked, cannot read the CLI's token, falls back to a weaker one,
// gets rejected by an organization policy — and then stayed in reduced mode
// until someone restarted it by hand, long after the keyring unlocked.
func TestDegradedModeIsRetried(t *testing.T) {
	graphql := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, policyBlockedResponse)
	}))
	defer graphql.Close()
	rest := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"total_count":0,"items":[]}`)
	}))
	defer rest.Close()

	src := NewSource("weak-token", "alice", nil, nil, 0, false)
	src.rich.SetEndpoint(graphql.URL)
	src.public.SetEndpoint(rest.URL)

	if _, err := src.Fetch(context.Background()); err != nil {
		t.Fatal(err)
	}
	if src.Mode() != ModePublic {
		t.Fatalf("should have degraded, mode = %s", src.Mode())
	}

	// Before the backoff elapses, nothing changes.
	src.maybeRetryRich()
	if src.Mode() != ModePublic {
		t.Error("the rich path must not be retried immediately")
	}

	// Once it has, and a better credential is available, the rich path
	// comes back.
	src.degradedAt = time.Now().Add(-2 * retryRichMax)
	src.SetTokenRefresher(func() string { return "the-good-token" })

	src.maybeRetryRich()

	if src.Mode() != ModeRich {
		t.Errorf("mode = %s, want the rich path restored once a better token appeared", src.Mode())
	}
	if src.public.Token != "the-good-token" {
		t.Errorf("the refreshed token should also be used for search, got %q", src.public.Token)
	}
}

// Retrying must not lose the search settings, or the restored path would
// silently start watching the wrong repositories.
func TestRetryKeepsSearchSettings(t *testing.T) {
	src := NewSource("weak", "alice", []string{"acme/api"}, []string{"app/bot"}, 7, true)
	src.degradedAt = time.Now().Add(-2 * retryRichMax)
	src.mode = ModePublic
	src.SetTokenRefresher(func() string { return "better" })

	src.maybeRetryRich()

	if src.rich.MaxAgeDays != 7 {
		t.Errorf("MaxAgeDays = %d, want 7", src.rich.MaxAgeDays)
	}
	if !src.rich.IncludeTeamReviews {
		t.Error("IncludeTeamReviews was lost on retry")
	}
	if len(src.rich.Repos) != 1 || src.rich.Repos[0] != "acme/api" {
		t.Errorf("Repos = %v", src.rich.Repos)
	}
	if len(src.rich.IgnoreAuthors) != 1 {
		t.Errorf("IgnoreAuthors = %v", src.rich.IgnoreAuthors)
	}
}

// GitHub answers 422 when a search names a repository the caller cannot see.
// "Validation Failed" sends people hunting for a syntax error that is not
// there, so the message has to say what actually happened.
func TestInvisibleRepositoryExplainsItself(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprint(w, `{"message":"Validation Failed","errors":[{"code":"invalid"}]}`)
	}))
	defer srv.Close()

	c := NewPublic("alice")
	c.SetEndpoint(srv.URL)

	_, err := c.RepoPullRequests(context.Background(), "private/repo", "", 10)
	if err == nil {
		t.Fatal("expected an error")
	}

	var notVisible *NotVisibleError
	if !errors.As(err, &notVisible) {
		t.Fatalf("error should be a NotVisibleError, got %T: %v", err, err)
	}
	if strings.Contains(err.Error(), "Validation Failed") {
		t.Errorf("the raw GitHub wording should not be forwarded: %v", err)
	}
	for _, want := range []string{"private", "token"} {
		if !strings.Contains(strings.ToLower(err.Error()), want) {
			t.Errorf("the message should mention %q: %v", want, err)
		}
	}
}

// Starting with no credential is not a permanent condition: the daemon boots
// before the system keyring unlocks. Without this, a daemon that came up
// blind stayed blind for the whole session — searching anonymously, unable to
// see any private repository, and never trying again.
func TestSourceStartedWithoutTokenStillRetries(t *testing.T) {
	src := NewSource("", "alice", nil, nil, 0, false)

	if src.Mode() != ModePublic {
		t.Fatalf("mode = %s, want public with no token", src.Mode())
	}
	if src.Authenticated() {
		t.Error("Authenticated() must be false with no token")
	}
	if src.degradedAt.IsZero() {
		t.Fatal("starting without a credential must be marked, or the retry never fires")
	}

	// The keyring unlocks a few minutes later.
	src.degradedAt = time.Now().Add(-2 * retryRichMax)
	src.SetTokenRefresher(func() string { return "now-readable" })

	src.maybeRetryRich()

	if src.Mode() != ModeRich {
		t.Errorf("mode = %s, want rich once the credential became readable", src.Mode())
	}
	if !src.Authenticated() {
		t.Error("the recovered token must also be used for search")
	}
}

// When the credential is still unavailable, the retry backs off again rather
// than shelling out to the CLI on every cycle.
func TestRetryBacksOffWhenStillUnavailable(t *testing.T) {
	src := NewSource("", "alice", nil, nil, 0, false)
	src.degradedAt = time.Now().Add(-2 * retryRichMax)
	src.SetTokenRefresher(func() string { return "" })

	src.maybeRetryRich()

	if src.Mode() != ModePublic {
		t.Errorf("mode = %s, should stay public when nothing is readable", src.Mode())
	}
	if time.Since(src.degradedAt) > time.Minute {
		t.Error("the backoff window should have been reset, not left expired")
	}
}

// The backoff must never make a person wait. It exists to stop the
// background loop shelling out to the CLI every couple of minutes; when
// someone explicitly asks for something, reading the keyring again costs
// milliseconds and the alternative is telling them to go check `doctor`
// while a working credential sits one call away.
func TestEnsureCredentialIgnoresTheBackoff(t *testing.T) {
	src := NewSource("", "alice", nil, nil, 0, false)
	src.SetTokenRefresher(func() string { return "now-readable" })

	// Well inside the backoff window: the background retry would skip.
	src.degradedAt = time.Now()
	src.maybeRetryRich()
	if src.Mode() != ModePublic {
		t.Fatal("the background retry should have respected the backoff")
	}

	// A person asking gets a fresh attempt regardless.
	src.EnsureCredential()

	if src.Mode() != ModeRich {
		t.Errorf("mode = %s, want rich after an explicit request", src.Mode())
	}
	if !src.Authenticated() {
		t.Error("the credential should now be in use for search")
	}
}

func TestEnsureCredentialIsCheapWhenAlreadyHealthy(t *testing.T) {
	calls := 0
	src := NewSource("already-have-one", "alice", nil, nil, 0, false)
	src.SetTokenRefresher(func() string { calls++; return "another" })

	src.EnsureCredential()

	if calls != 0 {
		t.Errorf("a healthy source should not re-read the credential, called %d times", calls)
	}
}

// The first retry comes quickly, because a keyring usually unlocks within a
// minute or two of someone logging in. Waiting ten minutes for the first
// attempt left a daemon blind for hours after a boot.
func TestRetryStartsShortThenWidens(t *testing.T) {
	src := NewSource("", "alice", nil, nil, 0, false)
	src.SetTokenRefresher(func() string { return "" }) // never becomes readable

	if src.retryAfter != retryRichInitial {
		t.Errorf("first interval = %s, want %s", src.retryAfter, retryRichInitial)
	}

	for range 10 {
		src.degradedAt = time.Now().Add(-2 * src.retryAfter)
		src.maybeRetryRich()
	}

	if src.retryAfter != retryRichMax {
		t.Errorf("interval = %s, want it capped at %s", src.retryAfter, retryRichMax)
	}
}

// Recovering resets the interval, so a later hiccup is retried promptly
// rather than inheriting a widened backoff.
func TestSuccessfulRetryResetsTheInterval(t *testing.T) {
	src := NewSource("", "alice", nil, nil, 0, false)
	src.SetTokenRefresher(func() string { return "" })

	src.degradedAt = time.Now().Add(-2 * src.retryAfter)
	src.maybeRetryRich()
	widened := src.retryAfter

	src.SetTokenRefresher(func() string { return "readable" })
	src.degradedAt = time.Now().Add(-2 * widened)
	src.maybeRetryRich()

	if src.retryAfter != retryRichInitial {
		t.Errorf("interval = %s, want it reset to %s after recovering", src.retryAfter, retryRichInitial)
	}
}
