package service

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// The unit file is what stands between "works while I am logged in" and
// "starts with the computer". A malformed one fails at boot, far away from
// anyone who could debug it, so the template is worth asserting on directly.
func TestSystemdUnitIsWellFormed(t *testing.T) {
	unit := renderSystemdUnit("/home/someone/.local/bin/ptal", "/home/someone/projects")

	for _, section := range []string{"[Unit]", "[Service]", "[Install]"} {
		if !strings.Contains(unit, section) {
			t.Errorf("missing section %s:\n%s", section, unit)
		}
	}

	// Restart=always plus a delay is what survives a laptop losing network
	// or GitHub returning 500s for a while.
	if !strings.Contains(unit, "Restart=always") {
		t.Error("the service must restart on failure")
	}
	if !strings.Contains(unit, "RestartSec=") {
		t.Error("restarting without a delay would spin on a persistent failure")
	}

	// WantedBy=default.target is what makes `systemctl --user enable` work.
	if !strings.Contains(unit, "WantedBy=default.target") {
		t.Error("without an [Install] target the unit cannot be enabled")
	}

	if !strings.Contains(unit, "ExecStart=/home/someone/.local/bin/ptal run") {
		t.Errorf("ExecStart does not point at the binary:\n%s", unit)
	}
	if !strings.Contains(unit, "WorkingDirectory=/home/someone/projects") {
		t.Error("without a working directory the daemon cannot find its .env")
	}
}

// KeepAlive/NetworkState is the macOS equivalent of "only run when there is
// internet". Dropping it would leave the daemon spinning on failures while
// the machine is offline.
func TestLaunchdPlistKeepsAliveOnNetwork(t *testing.T) {
	plist := renderLaunchdPlist("/usr/local/bin/ptal", "/Users/someone", "/Users/someone/Library/Logs")

	if !strings.Contains(plist, "<?xml") || !strings.Contains(plist, "</plist>") {
		t.Error("the plist is not a complete XML document")
	}
	for _, key := range []string{"RunAtLoad", "KeepAlive", "NetworkState", "SuccessfulExit"} {
		if !strings.Contains(plist, "<key>"+key+"</key>") {
			t.Errorf("missing key %s:\n%s", key, plist)
		}
	}
	if !strings.Contains(plist, "<string>/usr/local/bin/ptal</string>") {
		t.Error("the plist does not point at the binary")
	}
	if !strings.Contains(plist, "<string>run</string>") {
		t.Error("the plist does not pass the `run` argument")
	}
}

// Paths must land in the per-user location on every platform: writing to a
// system directory would require privileges this tool deliberately never asks
// for.
func TestPathsStayInsideTheUserHome(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("no home directory in this environment")
	}

	for name, path := range map[string]string{
		"systemd unit":  systemdUnitPath(),
		"launchd plist": launchdPlistPath(),
		"log directory": logDirectory(),
	} {
		if path == "" {
			continue
		}
		if !strings.HasPrefix(filepath.Clean(path), filepath.Clean(home)) {
			t.Errorf("%s escapes the home directory: %s", name, path)
		}
	}
}

func TestManagerNamesEveryPlatform(t *testing.T) {
	got := Manager()
	if got == "" || got == runtime.GOOS {
		t.Errorf("Manager() = %q, want a human-readable service manager name", got)
	}
}

// Status must answer on any platform without panicking, including when
// nothing is installed - `doctor` calls it before anything exists.
func TestStatusIsSafeWhenNothingIsInstalled(t *testing.T) {
	info := Status()

	if info.Manager == "" {
		t.Error("Status must always name the manager")
	}
	if info.Running && !info.Installed {
		t.Error("a service cannot be running without being installed")
	}
}
