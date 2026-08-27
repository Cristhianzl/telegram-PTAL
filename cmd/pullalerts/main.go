// Command pullalerts watches the pull requests tied to your GitHub user and
// tells you about them on Telegram.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

// version is replaced at build time for releases.
var version = "0.1.0-dev"

const usage = `pullalerts - your GitHub pull requests on Telegram

USAGE
  pullalerts <command>

COMMANDS
  setup       Connect Telegram and discover your chat automatically
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
		fmt.Printf("pullalerts %s\n", version)
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
