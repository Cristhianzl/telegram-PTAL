package config

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Setting describes one user-tunable option: how to show it, how to validate
// it, and what it is for.
//
// This table is the single place a setting is defined. `ptal config` renders
// itself from it, so adding an option here makes it settable, listable and
// documented at once.
type Setting struct {
	Key     string
	Summary string
	// Validate rejects a value before it reaches the file, so a typo becomes
	// an error now rather than a daemon that fails to start later.
	Validate func(string) error
	// Secret hides the value when listing.
	Secret bool
}

var settings = []Setting{
	{
		Key:      "POLL_INTERVAL",
		Summary:  "How often to check GitHub. Minimum 30s.",
		Validate: validateDuration(30 * time.Second),
	},
	{
		Key:      "MAX_AGE_DAYS",
		Summary:  "Ignore pull requests idle for more than N days. 0 disables.",
		Validate: validateInt(0, 3650),
	},
	{
		Key:      "MAX_PER_HOUR",
		Summary:  "Ceiling on messages per hour.",
		Validate: validateInt(1, 1000),
	},
	{
		Key:      "QUIET_HOURS",
		Summary:  `Deliver without a sound in this range, e.g. "23:00-08:00".`,
		Validate: validateQuietHours,
	},
	{
		Key:      "MUTE_EVENTS",
		Summary:  "Event types to never send. Run `ptal events` for the list.",
		Validate: validateEventList,
	},
	{
		Key:      "ALERT_ON",
		Summary:  "Send only these event types. Empty means all of them.",
		Validate: validateEventList,
	},
	{
		Key:      "WATCH_REPOS",
		Summary:  "Limit to repositories or organizations, comma-separated.",
		Validate: nil,
	},
	{
		Key:      "IGNORE_AUTHORS",
		Summary:  "Authors to skip entirely, comma-separated.",
		Validate: nil,
	},
	{
		Key:      "INCLUDE_TEAM_REVIEWS",
		Summary:  "Include reviews requested from your team, not just you.",
		Validate: validateBool,
	},
	{
		Key:      "IGNORE_DRAFTS",
		Summary:  "Draft pull requests do not count as review requests.",
		Validate: validateBool,
	},
	{
		Key:      "USE_GH_CLI",
		Summary:  "Use the GitHub CLI's token instead of GH_PAT_TOKEN.",
		Validate: validateBool,
	},
	{
		Key:      "PUBLIC_ONLY",
		Summary:  "Skip GraphQL and use public search only.",
		Validate: validateBool,
	},
	{
		Key:      "GITHUB_LOGIN",
		Summary:  "Your GitHub username. Discovered from the token when empty.",
		Validate: nil,
	},
}

// Settings returns the tunable options, in a stable order.
func Settings() []Setting { return settings }

// FindSetting looks a setting up by key, case-insensitively, so both
// `poll-interval` and `POLL_INTERVAL` work on the command line.
func FindSetting(key string) (Setting, bool) {
	normalized := NormalizeKey(key)
	for _, s := range settings {
		if s.Key == normalized {
			return s, true
		}
	}
	return Setting{}, false
}

// NormalizeKey accepts the shape people actually type - lowercase, hyphenated -
// and turns it into the environment variable name.
func NormalizeKey(key string) string {
	return strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(key), "-", "_"))
}

// SettingKeys lists the valid keys, for error messages.
func SettingKeys() []string {
	out := make([]string, 0, len(settings))
	for _, s := range settings {
		out = append(out, s.Key)
	}
	sort.Strings(out)
	return out
}

// Effective reports the value currently in force for a key, which is not
// always what the file says: environment variables win, and unset options
// fall back to a default.
func (c *Config) Effective(key string) string {
	switch NormalizeKey(key) {
	case "POLL_INTERVAL":
		return c.PollInterval.String()
	case "MAX_AGE_DAYS":
		return strconv.Itoa(c.MaxAgeDays)
	case "MAX_PER_HOUR":
		return strconv.Itoa(c.MaxPerHour)
	case "QUIET_HOURS":
		if c.QuietHours == "" {
			return "(disabled)"
		}
		return c.QuietHours
	case "WATCH_REPOS":
		if len(c.WatchRepos) == 0 {
			return "(everything)"
		}
		return strings.Join(c.WatchRepos, ",")
	case "MUTE_EVENTS":
		if len(c.MuteEvents) == 0 {
			return "(nothing muted)"
		}
		return strings.Join(c.MuteEvents, ",")
	case "ALERT_ON":
		if len(c.AlertOn) == 0 {
			return "(all events)"
		}
		return strings.Join(c.AlertOn, ",")
	case "IGNORE_AUTHORS":
		return strings.Join(c.IgnoreAuthors, ",")
	case "INCLUDE_TEAM_REVIEWS":
		return strconv.FormatBool(c.IncludeTeamReviews)
	case "IGNORE_DRAFTS":
		return strconv.FormatBool(c.IgnoreDrafts)
	case "USE_GH_CLI":
		return strconv.FormatBool(strings.Contains(c.TokenSource, "CLI"))
	case "PUBLIC_ONLY":
		return strconv.FormatBool(c.PublicOnly)
	case "GITHUB_LOGIN":
		if c.Login == "" {
			return "(from token)"
		}
		return c.Login
	}
	return ""
}

func validateDuration(min time.Duration) func(string) error {
	return func(v string) error {
		d, err := time.ParseDuration(v)
		if err != nil {
			return fmt.Errorf("not a duration: %q (try 2m, 30s, 1h)", v)
		}
		if d < min {
			return fmt.Errorf("%s is below the minimum of %s", d, min)
		}
		return nil
	}
}

func validateInt(min, max int) func(string) error {
	return func(v string) error {
		n, err := strconv.Atoi(v)
		if err != nil {
			return fmt.Errorf("not a number: %q", v)
		}
		if n < min || n > max {
			return fmt.Errorf("%d is outside the range %d-%d", n, min, max)
		}
		return nil
	}
}

func validateBool(v string) error {
	if v != "true" && v != "false" {
		return fmt.Errorf("must be true or false, got %q", v)
	}
	return nil
}

// validateEventList is wired up by the engine package, which owns the event
// names. Keeping the hook here avoids config importing engine, which would
// make the dependency run the wrong way.
var eventListValidator func(string) error

// SetEventListValidator installs the validator for event-name settings.
func SetEventListValidator(fn func(string) error) { eventListValidator = fn }

func validateEventList(v string) error {
	if v == "" || eventListValidator == nil {
		return nil
	}
	return eventListValidator(v)
}

func validateQuietHours(v string) error {
	if v == "" {
		return nil
	}
	from, to, ok := strings.Cut(v, "-")
	if !ok {
		return fmt.Errorf("must look like 23:00-08:00, got %q", v)
	}
	for _, clock := range []string{from, to} {
		if _, err := time.Parse("15:04", strings.TrimSpace(clock)); err != nil {
			return fmt.Errorf("%q is not a HH:MM time", strings.TrimSpace(clock))
		}
	}
	return nil
}
