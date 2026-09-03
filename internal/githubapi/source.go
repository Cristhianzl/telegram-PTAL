package githubapi

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// The rich path is retried on a widening interval: often at first, because a
// keyring usually unlocks within a minute or two of someone logging in, then
// rarely, because a genuine policy rejection will not fix itself.
const (
	retryRichInitial = time.Minute
	retryRichMax     = 10 * time.Minute
)

// Mode indicates where the data is coming from.
type Mode string

const (
	// ModeRich uses authenticated GraphQL: it carries CI state, review
	// decision and merge status.
	ModeRich Mode = "rich"
	// ModePublic uses REST search, possibly anonymous. It works even when an
	// organization policy rejects the token, but only sees public
	// repositories and carries fewer fields.
	ModePublic Mode = "public"
)

// Description explains the mode in one line, for `doctor` and the logs.
func (m Mode) Description() string {
	switch m {
	case ModeRich:
		return "authenticated GraphQL (with CI, approvals and conflicts)"
	case ModePublic:
		return "public search (no CI, approvals or private repositories)"
	}
	return string(m)
}

// Source fetches pull requests through the best available path and degrades
// on its own when the preferred path is blocked.
type Source struct {
	rich   *Client
	public *PublicClient

	mode  Mode
	login string

	// degradedAt is when the rich path was last given up on, and retryAfter
	// how long to wait before trying it again.
	//
	// Degrading permanently was a real failure: a daemon that starts at boot
	// finds the system keyring still locked, cannot read the GitHub CLI's
	// token, falls back to a weaker one, gets rejected by an organization
	// policy, and then stays in reduced mode until someone restarts it by
	// hand - long after the keyring unlocked.
	degradedAt time.Time
	retryAfter time.Duration
	// refreshToken re-reads the preferred credential, for the same reason.
	refreshToken func() string

	// The search parameters, kept so the rich client can be rebuilt on retry.
	repos              []string
	ignoreAuthors      []string
	maxAgeDays         int
	includeTeamReviews bool

	// OnModeChange fires when the mode changes, so the daemon can say so.
	OnModeChange func(from, to Mode, reason string)
}

// NewSource assembles the data source. With no token it starts in public mode.
//
// The token is handed to both layers deliberately: an organization can block
// GraphQL and still accept REST search, and there the token is worth a lot —
// it triples the request budget, from ten to thirty per minute.
func NewSource(token, login string, repos, ignoreAuthors []string, maxAgeDays int, includeTeamReviews bool) *Source {
	s := &Source{
		public: NewPublic(login),
		mode:   ModePublic,
		login:  login,
	}
	s.public.Token = token
	s.public.Repos = repos
	s.public.IgnoreAuthors = ignoreAuthors
	s.public.MaxAgeDays = maxAgeDays
	s.public.IncludeTeamReviews = includeTeamReviews
	s.repos, s.ignoreAuthors = repos, ignoreAuthors
	s.maxAgeDays, s.includeTeamReviews = maxAgeDays, includeTeamReviews

	s.retryAfter = retryRichInitial
	// Starting without a credential is the same situation as degrading into
	// one: the daemon boots before the system keyring unlocks, so the token
	// simply is not readable yet. Marking the moment is what lets the retry
	// pick it up later instead of running blind for the whole session.
	if token == "" {
		s.degradedAt = time.Now()
	}

	if token != "" {
		s.rich = New(token)
		s.rich.Repos = repos
		s.rich.IgnoreAuthors = ignoreAuthors
		s.rich.MaxAgeDays = maxAgeDays
		s.rich.IncludeTeamReviews = includeTeamReviews
		s.mode = ModeRich
	}
	return s
}

// Mode reports the current mode.
func (s *Source) Mode() Mode { return s.mode }

// Login returns the user being watched.
func (s *Source) Login() string { return s.login }

// Resolve discovers the token owner's login when it was not supplied.
// Without a login there is no way to run public search, which has no @me.
func (s *Source) Resolve(ctx context.Context) error {
	if s.login != "" {
		return nil
	}
	if s.rich == nil {
		return fmt.Errorf("set GITHUB_LOGIN in .env: without a token your user cannot be discovered")
	}
	login, err := s.rich.Viewer(ctx)
	if err != nil {
		return err
	}
	s.login = login
	s.public.Login = login
	return nil
}

// SetTokenRefresher installs a callback that re-reads the preferred
// credential. It is consulted when retrying the rich path, so a token that
// was unavailable at startup can still be picked up later.
func (s *Source) SetTokenRefresher(fn func() string) { s.refreshToken = fn }

// maybeRetryRich retries the authenticated path if the backoff has elapsed.
// This is the background path, called on every cycle.
func (s *Source) maybeRetryRich() {
	if s.mode != ModePublic || s.degradedAt.IsZero() {
		return
	}
	if time.Since(s.degradedAt) < s.retryAfter {
		return
	}
	s.tryRich()
}

// EnsureCredential tries to obtain a credential right now, ignoring the
// backoff.
//
// The backoff exists to stop the background loop shelling out to the CLI
// every couple of minutes. It should never make a person wait: when someone
// explicitly asks for something, reading the keyring again costs a few
// milliseconds, and the alternative is telling them to go check `doctor`
// while a working credential sits one call away.
func (s *Source) EnsureCredential() {
	if s.Authenticated() && s.mode == ModeRich {
		return
	}
	s.tryRich()
}

// tryRich attempts to restore the authenticated path, picking up a credential
// that became readable since the last attempt.
func (s *Source) tryRich() {
	if s.refreshToken != nil {
		if token := s.refreshToken(); token != "" && token != s.public.Token {
			s.public.Token = token
			s.rich = New(token)
			s.rich.Repos = s.repos
			s.rich.IgnoreAuthors = s.ignoreAuthors
			s.rich.MaxAgeDays = s.maxAgeDays
			s.rich.IncludeTeamReviews = s.includeTeamReviews
		}
	}

	if s.rich != nil {
		s.degradedAt = time.Time{}
		s.retryAfter = retryRichInitial
		s.switchMode(ModeRich, "retrying the authenticated path")
		return
	}

	// Still nothing readable. Widen the interval rather than trying on every
	// cycle, which would shell out to the CLI every two minutes.
	s.degradedAt = time.Now()
	if s.retryAfter *= 2; s.retryAfter > retryRichMax {
		s.retryAfter = retryRichMax
	}
}

// Authenticated reports whether searches currently carry a credential.
// Running anonymously is not an error, but it cannot see private
// repositories, and that is worth surfacing rather than discovering through
// a confusing 422.
func (s *Source) Authenticated() bool { return s.public.Token != "" }

// Fetch retrieves the picture of your pull requests, falling back to public
// mode if an organization policy blocks the token.
func (s *Source) Fetch(ctx context.Context) (*Snapshot, error) {
	s.maybeRetryRich()

	if s.mode == ModeRich && s.rich != nil {
		snap, err := s.rich.Fetch(ctx)
		if err == nil {
			return snap, nil
		}

		var policyErr *PolicyError
		if errors.As(err, &policyErr) {
			// The token is valid, but the organization rejects this kind of
			// token. The same data is usually publicly readable.
			s.degradedAt = time.Now()
			s.switchMode(ModePublic, policyErr.Msg)
			return s.fetchPublic(ctx)
		}
		return nil, err
	}
	return s.fetchPublic(ctx)
}

func (s *Source) fetchPublic(ctx context.Context) (*Snapshot, error) {
	if s.public.Login == "" {
		s.public.Login = s.login
	}
	if s.public.Login == "" {
		return nil, fmt.Errorf("public search needs your GitHub login")
	}
	return s.public.Fetch(ctx)
}

// UsePublicOnly forces public mode, skipping GraphQL.
//
// The token is still used for REST search unless dropToken says otherwise:
// even when the organization rejects GraphQL, the token usually still works
// for search and is worth keeping for the larger request budget.
func (s *Source) UsePublicOnly(dropToken bool) {
	if dropToken {
		s.public.Token = ""
	}
	s.rich = nil
	s.switchMode(ModePublic, "public mode set in the configuration")
}

// RateRemaining reports how many search requests still fit in the window.
func (s *Source) RateRemaining() int { return s.public.RateRemaining() }

func (s *Source) switchMode(to Mode, reason string) {
	if s.mode == to {
		return
	}
	from := s.mode
	s.mode = to
	if s.OnModeChange != nil {
		s.OnModeChange(from, to, reason)
	}
}

// PublicClientForTest exposes the public search client so tests can point it
// at a local server.
func (s *Source) PublicClientForTest() *PublicClient { return s.public }
