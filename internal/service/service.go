// Package service registers PTAL to start with the system.
//
// Linux and macOS each have their own mechanism, and neither requires
// administrator privileges: user-mode systemd and a LaunchAgent.
//
// Windows is not supported. Its service model differs enough that supporting
// it properly means a Windows Service or a scheduled task, both with their own
// elevation and console-window behavior — and untested code for a platform
// nobody here runs is worse than no code at all. Use WSL2, where the Linux
// path applies unchanged.
package service

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// Name identifies the service on all three systems.
const Name = "ptal"

// Info describes the installation state.
type Info struct {
	Installed bool
	Running   bool
	Manager   string
	UnitPath  string
	Detail    string
}

// Manager returns the name of the current system's service manager.
func Manager() string {
	switch runtime.GOOS {
	case "linux":
		return "systemd (user)"
	case "darwin":
		return "launchd (LaunchAgent)"
	}
	return runtime.GOOS
}

// Install registers the service and starts it.
func Install() error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("locating the executable: %w", err)
	}
	exe, _ = filepath.Abs(exe)

	// The daemon must find the .env; record the current working directory.
	wd, err := os.Getwd()
	if err != nil {
		wd = filepath.Dir(exe)
	}

	switch runtime.GOOS {
	case "linux":
		return installSystemd(exe, wd)
	case "darwin":
		return installLaunchd(exe, wd)
	}
	return unsupported()
}

// Uninstall removes the service registration.
func Uninstall() error {
	switch runtime.GOOS {
	case "linux":
		return uninstallSystemd()
	case "darwin":
		return uninstallLaunchd()
	}
	return unsupported()
}

// Restart reloads the daemon so a configuration change takes effect.
// Configuration is read once at startup, so changing a setting without this
// would appear to do nothing until the next reboot.
func Restart() error {
	switch runtime.GOOS {
	case "linux":
		return run("systemctl", "--user", "restart", Name+".service")
	case "darwin":
		target := fmt.Sprintf("gui/%d/%s", os.Getuid(), launchdLabel)
		return run("launchctl", "kickstart", "-k", target)
	}
	return unsupported()
}

// Status describes whether the service is installed and running.
func Status() Info {
	info := Info{Manager: Manager()}
	switch runtime.GOOS {
	case "linux":
		info.UnitPath = systemdUnitPath()
		info.Installed = fileExists(info.UnitPath)
		out, _ := exec.Command("systemctl", "--user", "is-active", Name+".service").Output()
		info.Running = strings.TrimSpace(string(out)) == "active"
		info.Detail = strings.TrimSpace(string(out))
	case "darwin":
		info.UnitPath = launchdPlistPath()
		info.Installed = fileExists(info.UnitPath)
		out, _ := exec.Command("launchctl", "list").Output()
		info.Running = strings.Contains(string(out), launchdLabel)
	}
	return info
}

// ---------- Linux / systemd ----------

func systemdUnitPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "systemd", "user", Name+".service")
}

const systemdTemplate = `[Unit]
Description=PTAL - your pull requests on Telegram
Documentation=https://github.com/Cristhianzl/telegram-PTAL
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=%s run
WorkingDirectory=%s
Restart=always
RestartSec=10
# The daemon needs no privilege beyond networking and its own HOME.
NoNewPrivileges=true
PrivateTmp=true

[Install]
WantedBy=default.target
`

// renderSystemdUnit is separated from installation so the unit file can be
// asserted on without touching the system.
func renderSystemdUnit(exe, wd string) string {
	return fmt.Sprintf(systemdTemplate, exe, wd)
}

func installSystemd(exe, wd string) error {
	path := systemdUnitPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(path, []byte(renderSystemdUnit(exe, wd)), 0o644); err != nil {
		return err
	}

	if err := run("systemctl", "--user", "daemon-reload"); err != nil {
		return err
	}
	if err := run("systemctl", "--user", "enable", "--now", Name+".service"); err != nil {
		return err
	}

	// Without linger the service dies at logout, the most common cause of
	// "it stopped working on its own".
	if user := os.Getenv("USER"); user != "" {
		if err := run("loginctl", "enable-linger", user); err != nil {
			return fmt.Errorf("service installed, but enabling linger failed: %w\n"+
				"run manually: loginctl enable-linger %s", err, user)
		}
	}
	return nil
}

func uninstallSystemd() error {
	_ = run("systemctl", "--user", "disable", "--now", Name+".service")
	if err := os.Remove(systemdUnitPath()); err != nil && !os.IsNotExist(err) {
		return err
	}
	return run("systemctl", "--user", "daemon-reload")
}

// ---------- macOS / launchd ----------

const launchdLabel = "dev.ptal.agent"

func launchdPlistPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Library", "LaunchAgents", launchdLabel+".plist")
}

// KeepAlive with NetworkState is exactly "only run when there is internet";
// launchd handles this natively.
const launchdTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key>              <string>%s</string>
  <key>ProgramArguments</key>
  <array>
    <string>%s</string>
    <string>run</string>
  </array>
  <key>WorkingDirectory</key>   <string>%s</string>
  <key>RunAtLoad</key>          <true/>
  <key>KeepAlive</key>
  <dict>
    <key>NetworkState</key>     <true/>
    <key>SuccessfulExit</key>   <false/>
  </dict>
  <key>ThrottleInterval</key>   <integer>10</integer>
  <key>StandardOutPath</key>    <string>%s/ptal.log</string>
  <key>StandardErrorPath</key>  <string>%s/ptal.log</string>
</dict>
</plist>
`

// renderLaunchdPlist is separated from installation for the same reason as
// renderSystemdUnit.
func renderLaunchdPlist(exe, wd, logDir string) string {
	return fmt.Sprintf(launchdTemplate, launchdLabel, exe, wd, logDir, logDir)
}

func installLaunchd(exe, wd string) error {
	path := launchdPlistPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	logDir := logDirectory()
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(path, []byte(renderLaunchdPlist(exe, wd, logDir)), 0o644); err != nil {
		return err
	}

	target := fmt.Sprintf("gui/%d", os.Getuid())
	// bootout before bootstrap makes installation idempotent.
	_ = run("launchctl", "bootout", target+"/"+launchdLabel)
	if err := run("launchctl", "bootstrap", target, path); err != nil {
		return err
	}
	return run("launchctl", "enable", target+"/"+launchdLabel)
}

func uninstallLaunchd() error {
	target := fmt.Sprintf("gui/%d", os.Getuid())
	_ = run("launchctl", "bootout", target+"/"+launchdLabel)
	if err := os.Remove(launchdPlistPath()); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// ---------- Windows / Agendador de Tarefas ----------

func installWindows(exe, wd string) error {
	// /sc onlogon needs neither elevation nor a password; the service starts
	// when the user logs in, the right behavior for a personal app.
	cmd := fmt.Sprintf(`"%s" run`, exe)
	if err := run("schtasks", "/create", "/tn", Name, "/sc", "onlogon",
		"/rl", "limited", "/f", "/tr", cmd); err != nil {
		return err
	}
	return run("schtasks", "/run", "/tn", Name)
}

func uninstallWindows() error {
	_ = run("schtasks", "/end", "/tn", Name)
	return run("schtasks", "/delete", "/tn", Name, "/f")
}

// ---------- helpers ----------

func logDirectory() string {
	home, _ := os.UserHomeDir()
	if runtime.GOOS == "darwin" {
		return filepath.Join(home, "Library", "Logs")
	}
	return filepath.Join(home, ".local", "state", "ptal")
}

// unsupported explains what to do rather than only naming the platform.
func unsupported() error {
	return fmt.Errorf("PTAL supports Linux and macOS, not %s.\n"+
		"  On Windows, run it inside WSL2 - the Linux path works unchanged", runtime.GOOS)
}

func run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg == "" {
			msg = err.Error()
		}
		return fmt.Errorf("%s %s: %s", name, strings.Join(args, " "), msg)
	}
	return nil
}

func fileExists(p string) bool {
	st, err := os.Stat(p)
	return err == nil && !st.IsDir()
}
