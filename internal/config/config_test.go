package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeEnv(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PULLALERTS_CONFIG", path)
	return path
}

func TestParsesRealWorldEnvFile(t *testing.T) {
	writeEnv(t, `# a comment
GH_PAT_TOKEN=ghp_abc123

export TELEGRAM_BOT_TOKEN="123:XYZ"
TELEGRAM_CHAT_ID='456'
WATCH_REPOS=octocat/hello-world, acme/app ,
IGNORE_DRAFTS=false
line_without_equals
`)

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}

	if cfg.GitHubToken != "ghp_abc123" {
		t.Errorf("token = %q", cfg.GitHubToken)
	}
	// Quotes around the value must not become part of the token.
	if cfg.TelegramToken != "123:XYZ" {
		t.Errorf("telegram token = %q - quotes were not stripped", cfg.TelegramToken)
	}
	if cfg.TelegramChat != "456" {
		t.Errorf("chat = %q", cfg.TelegramChat)
	}
	if len(cfg.WatchRepos) != 2 || cfg.WatchRepos[1] != "acme/app" {
		t.Errorf("repos = %v - whitespace and a trailing comma should be ignored", cfg.WatchRepos)
	}
	if cfg.IgnoreDrafts {
		t.Error("IGNORE_DRAFTS=false should turn the option off")
	}
}

// Environment variables must beat the file: that is how the GitHub Actions
// mode works, passing everything through secrets.
func TestEnvironmentOverridesFile(t *testing.T) {
	writeEnv(t, "GH_PAT_TOKEN=from_file\n")
	t.Setenv("GH_PAT_TOKEN", "from_environment")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.GitHubToken != "from_environment" {
		t.Errorf("token = %q, want the environment value", cfg.GitHubToken)
	}
}

func TestDefaultsAreSane(t *testing.T) {
	writeEnv(t, "GH_PAT_TOKEN=x\n")

	cfg, _ := Load()

	if cfg.PollInterval != DefaultPollInterval {
		t.Errorf("interval = %s", cfg.PollInterval)
	}
	if cfg.MaxPerHour != DefaultMaxPerHour {
		t.Errorf("hourly ceiling = %d", cfg.MaxPerHour)
	}
	if !cfg.IgnoreDrafts {
		t.Error("drafts should be ignored by default")
	}
	if len(cfg.IgnoreAuthors) == 0 {
		t.Error("bots should be ignored by default")
	}
}

// Too short an interval would turn the daemon into a 429 generator.
func TestPollIntervalFloor(t *testing.T) {
	writeEnv(t, "GH_PAT_TOKEN=x\nPOLL_INTERVAL=1s\n")

	cfg, _ := Load()

	if cfg.PollInterval < DefaultPollInterval {
		t.Errorf("interval = %s, values under 30s should be rejected", cfg.PollInterval)
	}
}

func TestValidateRequiresChatID(t *testing.T) {
	writeEnv(t, "GH_PAT_TOKEN=x\nTELEGRAM_BOT_TOKEN=y\n")

	cfg, _ := Load()

	if err := cfg.Validate(); err != ErrNoTelegramChat {
		t.Errorf("error = %v, want ErrNoTelegramChat", err)
	}
}

// Setup writes the discovered chat ID without destroying the rest of the file.
func TestSaveChatIDPreservesTheFile(t *testing.T) {
	path := writeEnv(t, "# my file\nGH_PAT_TOKEN=x\nTELEGRAM_BOT_TOKEN=y\n")

	cfg, _ := Load()
	if err := cfg.SaveChatID("999"); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(path)
	content := string(data)

	if !strings.Contains(content, "TELEGRAM_CHAT_ID=999") {
		t.Errorf("chat ID was not written:\n%s", content)
	}
	for _, keep := range []string{"# my file", "GH_PAT_TOKEN=x", "TELEGRAM_BOT_TOKEN=y"} {
		if !strings.Contains(content, keep) {
			t.Errorf("setup erased %q:\n%s", keep, content)
		}
	}

	// The file holds two secrets: it must not be world-readable.
	st, _ := os.Stat(path)
	if perm := st.Mode().Perm(); perm != 0o600 {
		t.Errorf("permissions = %o, want 600", perm)
	}
}

func TestSaveChatIDReplacesExistingValue(t *testing.T) {
	path := writeEnv(t, "TELEGRAM_CHAT_ID=old\n")

	cfg, _ := Load()
	if err := cfg.SaveChatID("new"); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(path)
	if strings.Contains(string(data), "old") {
		t.Errorf("the old value remained in the file:\n%s", data)
	}
	if strings.Count(string(data), "TELEGRAM_CHAT_ID") != 1 {
		t.Errorf("the key was duplicated:\n%s", data)
	}
}
