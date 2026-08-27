// Package config loads and validates the credentials and preferences that
// PTAL needs to run.
//
// Resolution order: process environment variables always win, then the .env
// file found at the first existing path among $PTAL_CONFIG, the current
// directory, the executable's directory, and the OS configuration directory.
package config

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Config holds everything a single cycle needs.
type Config struct {
	GitHubToken   string
	TelegramToken string
	TelegramChat  string

	// PollInterval is the gap between full GitHub searches.
	PollInterval time.Duration
	// BatchWindow is how long non-urgent events wait to be grouped.
	BatchWindow time.Duration
	// MaxPerHour caps messages per hour, guarding against floods.
	MaxPerHour int
	// QuietHours is a "23:00-08:00" range delivered without a sound.
	QuietHours string
	// IgnoreAuthors are logins excluded from searches. Bots by default.
	IgnoreAuthors []string
	// IgnoreDrafts suppresses review requests on draft pull requests.
	IgnoreDrafts bool
	// IncludeTeamReviews includes pull requests where the review was
	// requested from your team rather than from you by name. Off by default.
	IncludeTeamReviews bool
	// AlertOn, when non-empty, is the only set of event kinds delivered.
	AlertOn []string
	// MuteEvents are event kinds never delivered. Applied after AlertOn.
	MuteEvents []string
	// MaxAgeDays drops pull requests with no activity in the last N days.
	// Zero disables the filter. This is what separates what is alive from an
	// archive of pull requests abandoned years ago.
	MaxAgeDays int

	// Login is the GitHub user to watch. Discovered from the token when
	// empty, but required in public mode, where there is no @me.
	Login string
	// WatchRepos limits watching to repositories ("owner/name") or whole
	// organizations (bare name). Empty watches everything.
	WatchRepos []string
	// PublicOnly skips the GraphQL path, useful when an organization policy
	// rejects the token but the repositories are public.
	PublicOnly bool
	// TokenSource records where the credential came from, for `doctor`.
	TokenSource string

	// ReviewRepos are the repositories where the review button is offered.
	// Empty disables reviewing entirely: running an agent over a branch is
	// something to opt into per repository, not a default.
	ReviewRepos []string
	// ReviewRulesDir holds the .claude directory the review follows.
	ReviewRulesDir string
	// ReviewModel overrides the model used for reviews.
	ReviewModel string
	// ReviewTimeout bounds a single review.
	ReviewTimeout time.Duration

	// SourcePath is the .env file the configuration came from, so `setup`
	// can write back into it.
	SourcePath string
	// StatePath is where state persists between runs.
	StatePath string
}

// Defaults chosen to fit comfortably inside GitHub's rate limit without
// turning the chat into a wall of notifications.
const (
	DefaultPollInterval = 2 * time.Minute
	DefaultBatchWindow  = 60 * time.Second
	DefaultMaxPerHour   = 30
	// DefaultMaxAgeDays drops pull requests idle for more than two weeks.
	// Ones abandoned years ago are still technically open and would drown
	// out the few that actually need you.
	DefaultMaxAgeDays = 14
)

var defaultIgnoredAuthors = []string{
	"app/dependabot",
	"app/renovate",
	"app/github-actions",
}

// ErrNoTelegramChat means `setup` has not been run yet.
var ErrNoTelegramChat = errors.New("TELEGRAM_CHAT_ID is not set: run `ptal setup`")

// Load resolves the .env file, applies environment overrides and validates.
func Load() (*Config, error) {
	path := findEnvFile()
	values := map[string]string{}
	if path != "" {
		v, err := parseEnvFile(path)
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", path, err)
		}
		values = v
	}

	get := func(key string) string {
		if v, ok := os.LookupEnv(key); ok && strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
		return strings.TrimSpace(values[key])
	}

	c := &Config{
		GitHubToken:        firstNonEmpty(get("GH_PAT_TOKEN"), get("GITHUB_TOKEN")),
		TokenSource:        ".env file",
		TelegramToken:      get("TELEGRAM_BOT_TOKEN"),
		TelegramChat:       get("TELEGRAM_CHAT_ID"),
		PollInterval:       DefaultPollInterval,
		BatchWindow:        DefaultBatchWindow,
		MaxPerHour:         DefaultMaxPerHour,
		MaxAgeDays:         DefaultMaxAgeDays,
		QuietHours:         get("QUIET_HOURS"),
		Login:              get("GITHUB_LOGIN"),
		PublicOnly:         get("PUBLIC_ONLY") == "true",
		IncludeTeamReviews: get("INCLUDE_TEAM_REVIEWS") == "true",
		IgnoreDrafts:       get("IGNORE_DRAFTS") != "false",
		SourcePath:         path,
	}

	if d, err := time.ParseDuration(get("POLL_INTERVAL")); err == nil && d >= 30*time.Second {
		c.PollInterval = d
	}
	if d, err := time.ParseDuration(get("BATCH_WINDOW")); err == nil && d >= 0 {
		c.BatchWindow = d
	}
	if n, err := strconv.Atoi(get("MAX_PER_HOUR")); err == nil && n > 0 {
		c.MaxPerHour = n
	}
	if n, err := strconv.Atoi(get("MAX_AGE_DAYS")); err == nil && n >= 0 {
		c.MaxAgeDays = n
	}

	if raw := get("IGNORE_AUTHORS"); raw != "" {
		c.IgnoreAuthors = splitList(raw)
	} else {
		c.IgnoreAuthors = append([]string(nil), defaultIgnoredAuthors...)
	}
	c.WatchRepos = splitList(get("WATCH_REPOS"))
	c.AlertOn = splitList(get("ALERT_ON"))
	c.ReviewRepos = splitList(get("REVIEW_REPOS"))
	c.ReviewRulesDir = get("REVIEW_RULES_DIR")
	c.ReviewModel = get("REVIEW_MODEL")
	c.ReviewTimeout = 15 * time.Minute
	if d, err := time.ParseDuration(get("REVIEW_TIMEOUT")); err == nil && d > time.Minute {
		c.ReviewTimeout = d
	}
	c.MuteEvents = splitList(get("MUTE_EVENTS"))

	// With no token in the file, try the GitHub CLI: anyone already using
	// `gh` does not need to create a credential at all. USE_GH_CLI=true
	// prefers it even when a token exists, which matters when that token is
	// rejected by an organization policy.
	if c.GitHubToken == "" || get("USE_GH_CLI") == "true" {
		if cliToken := GitHubCLIToken(); cliToken != "" {
			c.GitHubToken = cliToken
			c.TokenSource = "GitHub CLI (gh auth token)"
		}
	}

	c.StatePath = get("STATE_PATH")
	if c.StatePath == "" {
		c.StatePath = defaultStatePath(path)
	}
	c.PublicOnly = c.PublicOnly || get("PUBLIC_ONLY") == "true"

	return c, nil
}

// Validate checks only what a full cycle cannot run without.
func (c *Config) Validate() error {
	var missing []string
	if c.GitHubToken == "" && c.Login == "" {
		missing = append(missing, "GH_PAT_TOKEN or GITHUB_LOGIN")
	}
	if c.TelegramToken == "" {
		missing = append(missing, "TELEGRAM_BOT_TOKEN")
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing from .env: %s", strings.Join(missing, ", "))
	}
	if c.TelegramChat == "" {
		return ErrNoTelegramChat
	}
	return nil
}

// SaveChatID writes the discovered chat ID back into .env.
func (c *Config) SaveChatID(chatID string) error {
	if err := c.SetValue("TELEGRAM_CHAT_ID", chatID); err != nil {
		return err
	}
	c.TelegramChat = chatID
	return nil
}

// DefaultConfigPath is where configuration belongs when none exists yet:
// the OS configuration directory, not the working directory. Writing to the
// working directory would put the file wherever the user happened to be
// standing, and the daemon would then only find it from there.
func DefaultConfigPath() string {
	if dir, err := os.UserConfigDir(); err == nil {
		return filepath.Join(dir, "ptal", ".env")
	}
	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, ".config", "ptal", ".env")
	}
	return ".env"
}

// ReviewEnabledFor reports whether reviewing is allowed in a repository.
//
// The check is an allowlist rather than a flag because the blast radius is
// per repository: reviewing your own project is very different from running
// an agent over a branch a stranger opened.
func (c *Config) ReviewEnabledFor(repo string) bool {
	for _, allowed := range c.ReviewRepos {
		if strings.EqualFold(allowed, repo) {
			return true
		}
		// A bare organization name enables every repository under it.
		if !strings.Contains(allowed, "/") &&
			strings.HasPrefix(strings.ToLower(repo), strings.ToLower(allowed)+"/") {
			return true
		}
	}
	return false
}

// Dir returns the directory holding the configuration, which is what the
// service installer pins as its working directory.
func (c *Config) Dir() string {
	if c.SourcePath != "" {
		return filepath.Dir(c.SourcePath)
	}
	return filepath.Dir(DefaultConfigPath())
}

// SetValue writes one key into the .env file, preserving comments, blank
// lines and the order of the entries already there. Rewriting the file from
// the parsed map instead would silently discard every comment the user has,
// including the ones the template ships with.
func (c *Config) SetValue(key, value string) error {
	path := c.SourcePath
	if path == "" {
		path = DefaultConfigPath()
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			return err
		}
		// Remember it, so a second write in the same run lands in the same
		// file rather than creating another one.
		c.SourcePath = path
	}

	var lines []string
	if data, err := os.ReadFile(path); err == nil {
		lines = strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	} else if !os.IsNotExist(err) {
		return err
	}

	replaced := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, key+"=") ||
			strings.HasPrefix(trimmed, "export "+key+"=") {
			lines[i] = key + "=" + value
			replaced = true
			break
		}
	}
	if !replaced {
		lines = append(lines, key+"="+value)
	}

	return writeFilePrivate(path, []byte(strings.Join(lines, "\n")+"\n"))
}

// writeFilePrivate writes atomically with restrictive permissions: the file
// holds two secrets and must not be readable by other users.
func writeFilePrivate(path string, data []byte) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		os.Remove(tmp)
		return err
	}
	return os.Chmod(path, 0o600)
}

func findEnvFile() string {
	// An explicit PTAL_CONFIG is honored even when the file is absent.
	// Falling through to the next candidate would silently load a different
	// configuration than the one the caller named - which in a test run means
	// reading the developer's own credentials.
	if p := os.Getenv("PTAL_CONFIG"); p != "" {
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p
		}
		return ""
	}

	var candidates []string
	if wd, err := os.Getwd(); err == nil {
		candidates = append(candidates, filepath.Join(wd, ".env"))
	}
	if exe, err := os.Executable(); err == nil {
		candidates = append(candidates, filepath.Join(filepath.Dir(exe), ".env"))
	}
	if dir, err := os.UserConfigDir(); err == nil {
		candidates = append(candidates, filepath.Join(dir, "ptal", ".env"))
	}
	for _, p := range candidates {
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p
		}
	}
	return ""
}

// defaultStatePath keeps state next to the .env, so installing the service
// does not change which file the daemon reads.
func defaultStatePath(envPath string) string {
	if envPath != "" {
		return filepath.Join(filepath.Dir(envPath), "state.json")
	}
	if dir, err := os.UserConfigDir(); err == nil {
		return filepath.Join(dir, "ptal", "state.json")
	}
	return "state.json"
}

// parseEnvFile understands the subset of .env that matters: KEY=VALUE, #
// comments, optional quotes, and an `export` prefix.
func parseEnvFile(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	out := map[string]string{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, found := strings.Cut(strings.TrimPrefix(line, "export "), "=")
		if !found {
			continue
		}
		out[strings.TrimSpace(key)] = unquote(strings.TrimSpace(value))
	}
	return out, sc.Err()
}

func unquote(v string) string {
	if len(v) >= 2 {
		if (v[0] == '"' && v[len(v)-1] == '"') || (v[0] == '\'' && v[len(v)-1] == '\'') {
			return v[1 : len(v)-1]
		}
	}
	return v
}

func splitList(raw string) []string {
	var out []string
	for _, item := range strings.Split(raw, ",") {
		if item = strings.TrimSpace(item); item != "" {
			out = append(out, item)
		}
	}
	return out
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
