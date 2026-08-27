package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A bad value must be caught before it reaches the file. Writing it and
// letting the daemon fail at startup would turn a typo into silence.
func TestValidationRejectsBadValues(t *testing.T) {
	cases := []struct {
		key, value string
	}{
		{"POLL_INTERVAL", "5"},          // no unit
		{"POLL_INTERVAL", "1s"},         // below the floor
		{"POLL_INTERVAL", "soon"},       // not a duration
		{"MAX_AGE_DAYS", "-1"},          // out of range
		{"MAX_AGE_DAYS", "many"},        // not a number
		{"MAX_PER_HOUR", "0"},           // would silence everything
		{"QUIET_HOURS", "23:00"},        // missing the end
		{"QUIET_HOURS", "25:00-08:00"},  // not a clock time
		{"IGNORE_DRAFTS", "yes"},        // not a bool
	}

	for _, c := range cases {
		setting, ok := FindSetting(c.key)
		if !ok {
			t.Fatalf("%s is not a known setting", c.key)
		}
		if setting.Validate == nil {
			t.Fatalf("%s has no validator", c.key)
		}
		if err := setting.Validate(c.value); err == nil {
			t.Errorf("%s=%q should have been rejected", c.key, c.value)
		}
	}
}

func TestValidationAcceptsGoodValues(t *testing.T) {
	cases := []struct{ key, value string }{
		{"POLL_INTERVAL", "30s"},
		{"POLL_INTERVAL", "2m"},
		{"POLL_INTERVAL", "1h"},
		{"MAX_AGE_DAYS", "0"},
		{"MAX_AGE_DAYS", "365"},
		{"MAX_PER_HOUR", "1"},
		{"QUIET_HOURS", "23:00-08:00"},
		{"QUIET_HOURS", ""},
		{"IGNORE_DRAFTS", "false"},
	}

	for _, c := range cases {
		setting, _ := FindSetting(c.key)
		if err := setting.Validate(c.value); err != nil {
			t.Errorf("%s=%q should be valid: %v", c.key, c.value, err)
		}
	}
}

// People type `poll-interval`, not `POLL_INTERVAL`.
func TestKeysAreForgiving(t *testing.T) {
	for _, typed := range []string{"poll-interval", "POLL_INTERVAL", "Poll-Interval", " poll_interval "} {
		if _, ok := FindSetting(typed); !ok {
			t.Errorf("%q should resolve to POLL_INTERVAL", typed)
		}
	}
	if _, ok := FindSetting("nonsense"); ok {
		t.Error("an unknown key must not resolve")
	}
}

// Writing a setting must not destroy the comments in .env - the template
// ships with explanations that are the only documentation many users read.
func TestSetValuePreservesCommentsAndOrder(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	original := `# Telegram
TELEGRAM_BOT_TOKEN=abc

# How often to check
POLL_INTERVAL=2m

# Trailing comment
`
	if err := os.WriteFile(path, []byte(original), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PTAL_CONFIG", path)

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if err := cfg.SetValue("POLL_INTERVAL", "10m"); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(path)
	content := string(data)

	if !strings.Contains(content, "POLL_INTERVAL=10m") {
		t.Errorf("the value was not written:\n%s", content)
	}
	if strings.Contains(content, "POLL_INTERVAL=2m") {
		t.Errorf("the old value is still there:\n%s", content)
	}
	for _, comment := range []string{"# Telegram", "# How often to check", "# Trailing comment"} {
		if !strings.Contains(content, comment) {
			t.Errorf("comment %q was destroyed:\n%s", comment, content)
		}
	}
	if !strings.Contains(content, "TELEGRAM_BOT_TOKEN=abc") {
		t.Errorf("an unrelated setting was lost:\n%s", content)
	}
}

func TestSetValueAppendsWhenKeyIsAbsent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	if err := os.WriteFile(path, []byte("TELEGRAM_BOT_TOKEN=abc\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PTAL_CONFIG", path)

	cfg, _ := Load()
	if err := cfg.SetValue("MAX_AGE_DAYS", "3"); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(path)
	if !strings.Contains(string(data), "MAX_AGE_DAYS=3") {
		t.Errorf("the new key was not appended:\n%s", data)
	}
}

// `ptal config` shows what is actually in force, which is not the file value
// when a default applies or an environment variable overrides it.
func TestEffectiveShowsWhatIsInForce(t *testing.T) {
	t.Setenv("PTAL_CONFIG", filepath.Join(t.TempDir(), "absent"))

	cfg, _ := Load()

	if got := cfg.Effective("poll-interval"); got != DefaultPollInterval.String() {
		t.Errorf("poll-interval = %q, want the default %s", got, DefaultPollInterval)
	}
	if got := cfg.Effective("watch-repos"); got != "(everything)" {
		t.Errorf("an empty repo list should read as %q, got %q", "(everything)", got)
	}
	if got := cfg.Effective("mute-events"); got != "(nothing muted)" {
		t.Errorf("mute-events = %q", got)
	}
}

// Every setting needs a summary: `ptal config` renders itself from this
// table, and a blank line there is a setting nobody can understand.
func TestEverySettingIsDocumented(t *testing.T) {
	for _, s := range Settings() {
		if strings.TrimSpace(s.Summary) == "" {
			t.Errorf("%s has no summary", s.Key)
		}
		if s.Key != strings.ToUpper(s.Key) {
			t.Errorf("%s should be an upper-case environment variable name", s.Key)
		}
	}
}
