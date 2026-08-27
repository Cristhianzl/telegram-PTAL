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
  run         Keep running and alerting (what the service executes)
  install     Register to start with the system
  uninstall   Remove the registration
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

	var err error
	switch os.Args[1] {
	case "setup":
		err = cmdSetup(ctx)
	case "config":
		err = cmdConfig(configArgs())
	case "events":
		err = cmdEvents()
	case "doctor":
		err = cmdDoctor(ctx)
	case "once":
		err = cmdOnce(ctx)
	case "panel":
		err = cmdPanel(ctx)
	case "run":
		err = cmdRun(ctx)
	case "install":
		err = cmdInstall()
	case "uninstall":
		err = cmdUninstall()
	case "status":
		err = cmdStatus()
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
