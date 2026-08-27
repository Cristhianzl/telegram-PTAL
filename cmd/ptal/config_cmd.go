package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/Cristhianzl/telegram-PTAL/internal/config"
	"github.com/Cristhianzl/telegram-PTAL/internal/service"
)

const configUsage = `ptal config - read and change settings

USAGE
  ptal config                  List every setting and its current value
  ptal config <key>            Print one value
  ptal config <key> <value>    Change a setting

EXAMPLES
  ptal config poll-interval 5m
  ptal config max-age-days 3
  ptal config quiet-hours 23:00-08:00
  ptal config watch-repos octocat/hello-world,acme

Keys are case-insensitive and accept hyphens, so poll-interval and
POLL_INTERVAL are the same setting. Changes are written to your .env file
and take effect on the next cycle; the service is restarted for you.
`

func cmdConfig(args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	switch len(args) {
	case 0:
		return listSettings(cfg)
	case 1:
		if args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
			fmt.Print(configUsage)
			return nil
		}
		return printSetting(cfg, args[0])
	case 2:
		return setSetting(cfg, args[0], args[1])
	default:
		// A space in an unquoted value is the likely cause, and the error
		// should say so rather than just counting arguments.
		return fmt.Errorf("expected at most two arguments, got %d\n"+
			"  If the value contains spaces, quote it: ptal config quiet-hours \"23:00-08:00\"", len(args))
	}
}

func listSettings(cfg *config.Config) error {
	if cfg.SourcePath != "" {
		fmt.Printf("%s\n\n", cfg.SourcePath)
	} else {
		fmt.Printf("no .env found - showing defaults\n\n")
	}

	for _, s := range config.Settings() {
		fmt.Printf("  %-22s %s\n", strings.ToLower(strings.ReplaceAll(s.Key, "_", "-")),
			cfg.Effective(s.Key))
		fmt.Printf("  %-22s \033[2m%s\033[0m\n\n", "", s.Summary)
	}

	fmt.Println("Change one with:  ptal config <key> <value>")
	return nil
}

func printSetting(cfg *config.Config, key string) error {
	if _, ok := config.FindSetting(key); !ok {
		return unknownKey(key)
	}
	fmt.Println(cfg.Effective(key))
	return nil
}

func setSetting(cfg *config.Config, key, value string) error {
	setting, ok := config.FindSetting(key)
	if !ok {
		return unknownKey(key)
	}
	if setting.Validate != nil {
		if err := setting.Validate(value); err != nil {
			return fmt.Errorf("%s: %w", setting.Key, err)
		}
	}

	previous := cfg.Effective(setting.Key)
	if err := cfg.SetValue(setting.Key, value); err != nil {
		return fmt.Errorf("writing to %s: %w", cfg.SourcePath, err)
	}

	fmt.Printf("✓ %s: %s → %s\n", strings.ToLower(strings.ReplaceAll(setting.Key, "_", "-")),
		previous, value)

	// A setting the daemon already read is only real once it restarts, so
	// doing it here avoids a change that silently does nothing.
	if info := service.Status(); info.Running {
		if err := service.Restart(); err != nil {
			fmt.Printf("• Restart the service to apply: %v\n", err)
			return nil
		}
		fmt.Println("✓ service restarted")
	}
	return nil
}

func unknownKey(key string) error {
	return fmt.Errorf("unknown setting %q\n  Valid keys: %s\n  Run `ptal config` to see them with their current values",
		key, strings.Join(config.SettingKeys(), ", "))
}

// configArgs returns everything after the subcommand name.
func configArgs() []string {
	if len(os.Args) < 3 {
		return nil
	}
	return os.Args[2:]
}
