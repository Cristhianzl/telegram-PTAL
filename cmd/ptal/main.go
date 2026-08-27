// Command ptal watches the pull requests tied to your GitHub user and
// tells you about them on Telegram.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Cristhianzl/telegram-PTAL/internal/config"
	"github.com/Cristhianzl/telegram-PTAL/internal/engine"
)

// version is replaced at build time for releases.
var version = "0.1.0-dev"

const usage = `ptal - your GitHub pull requests on Telegram

USAGE
  ptal <command>

COMMANDS
  setup       Connect Telegram and discover your chat automatically
  config      Read and change settings (ptal config poll-interval 5m)
  events      List alert types and which are on
  doctor      Diagnose token, chat, connectivity and service
  once        Run a single cycle and print the result
  panel       Send a panel with the current state to Telegram
  repo        List every open PR in a repository (ptal repo owner/name)
  run         Keep running and alerting (what the service executes)
  install     Register to start with the system
  uninstall   Remove the registration

  start       Start the daemon
  stop        Stop the daemon (it returns at the next boot)
  restart     Restart the daemon
  pause       Stop alerting for a while (ptal pause 2h)
  resume      Start alerting again
  status      Show the service and the last sync
  version     Print the version

Configuration lives in a .env file. See .env.example for the options.
`

func main() {
	log.SetFlags(log.Ltime)

	// The engine owns the event names; config validates against them without
	// importing it, so the dependency does not run backwards.
	config.SetEventListValidator(func(v string) error {
		_, err := engine.ParseKinds(v)
		return err
	})
	if len(os.Args) < 2 {
		fmt.Print(usage)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Commands that take no arguments must say so rather than silently
	// ignoring them: `ptal install --help` used to install the service.
	rest := os.Args[2:]
	noArgs := func(name string) error {
		if len(rest) > 0 {
			return fmt.Errorf("%s takes no arguments, got %q\n\n%s", name, rest[0], usage)
		}
		return nil
	}

	var err error
	switch os.Args[1] {
	case "setup":
		if err = noArgs("setup"); err == nil {
			err = cmdSetup(ctx)
		}
	case "config":
		err = cmdConfig(configArgs())
	case "events":
		if err = noArgs("events"); err == nil {
			err = cmdEvents()
		}
	case "doctor":
		if err = noArgs("doctor"); err == nil {
			err = cmdDoctor(ctx)
		}
	case "once":
		if err = noArgs("once"); err == nil {
			err = cmdOnce(ctx)
		}
	case "panel":
		if err = noArgs("panel"); err == nil {
			err = cmdPanel(ctx)
		}
	case "repo":
		err = cmdRepo(rest)
	case "run":
		if err = noArgs("run"); err == nil {
			err = cmdRun(ctx)
		}
	case "start":
		if err = noArgs("start"); err == nil {
			err = cmdStart()
		}
	case "stop":
		if err = noArgs("stop"); err == nil {
			err = cmdStop()
		}
	case "restart":
		if err = noArgs("restart"); err == nil {
			err = cmdRestart()
		}
	case "pause":
		err = cmdPause(rest)
	case "resume":
		if err = noArgs("resume"); err == nil {
			err = cmdResume()
		}
	case "install":
		if err = noArgs("install"); err == nil {
			err = cmdInstall()
		}
	case "uninstall":
		if err = noArgs("uninstall"); err == nil {
			err = cmdUninstall()
		}
	case "status":
		if err = noArgs("status"); err == nil {
			err = cmdStatus()
		}
	case "version", "--version", "-v":
		fmt.Printf("ptal %s\n", version)
	case "help", "--help", "-h":
		fmt.Print(usage)
	default:
		fmt.Printf("unknown command: %s\n\n%s", os.Args[1], usage)
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "\n✗ %v\n", err)
		os.Exit(1)
	}
}
