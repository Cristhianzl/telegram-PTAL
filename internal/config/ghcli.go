package config

import (
	"context"
	"os/exec"
	"strings"
	"time"
)

// GitHubCLIToken returns the token of the GitHub CLI's active account, if any.
//
// This exists because anyone already using `gh` usually holds a better
// credential than the one they would create by hand: the CLI token is OAuth,
// carries the `repo` scope, and passes organization policies that reject
// fine-grained tokens. Reusing it removes the most tedious setup step.
//
// An empty string only means "not available" — it is never an error.
func GitHubCLIToken() string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// `gh` keeps the token in the system keyring. In a background service
	// the keyring may be locked and the command fails, which is why there is
	// always a fallback path through .env.
	out, err := exec.CommandContext(ctx, "gh", "auth", "token").Output()
	if err != nil {
		return ""
	}
	token := strings.TrimSpace(string(out))
	for _, prefix := range []string{"gho_", "ghp_", "ghu_"} {
		if strings.HasPrefix(token, prefix) {
			return token
		}
	}
	return ""
}

// HasGitHubCLI reports whether `gh` is installed and authenticated.
func HasGitHubCLI() bool { return GitHubCLIToken() != "" }
