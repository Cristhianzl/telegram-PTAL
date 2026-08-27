package main

import (
	"fmt"
	"time"

	"github.com/Cristhianzl/telegram-PTAL/internal/config"
	"github.com/Cristhianzl/telegram-PTAL/internal/service"
	"github.com/Cristhianzl/telegram-PTAL/internal/store"
)

func cmdStart() error {
	if info := service.Status(); !info.Installed {
		return fmt.Errorf("the service is not installed - run `ptal install` first")
	}
	if err := service.Start(); err != nil {
		return err
	}
	fmt.Println("✓ started")
	return nil
}

// cmdStop halts the daemon but leaves it registered, so it returns at the
// next boot. That is almost always what someone means by "stop"; removing it
// entirely is `uninstall`.
func cmdStop() error {
	info := service.Status()
	if !info.Installed {
		return fmt.Errorf("the service is not installed, so there is nothing to stop")
	}
	if !info.Running {
		fmt.Println("• already stopped")
		return nil
	}
	if err := service.Stop(); err != nil {
		return err
	}
	fmt.Println("✓ stopped")
	fmt.Println("  It will start again at the next boot. To prevent that: ptal uninstall")
	fmt.Println("  To silence alerts without stopping:                    ptal pause 2h")
	return nil
}

func cmdRestart() error {
	if err := service.Restart(); err != nil {
		return err
	}
	fmt.Println("✓ restarted")
	return nil
}

const pauseUsage = `ptal pause - stop alerting for a while

USAGE
  ptal pause <duration>     Suppress alerts for the given time
  ptal pause                Show whether alerts are paused
  ptal resume               Start alerting again

EXAMPLES
  ptal pause 2h
  ptal pause 30m
  ptal pause 1h30m

Pausing keeps the daemon running. It carries on checking GitHub and
recording what changed, it just does not message you - so when the pause
ends you are not buried under everything you missed.

To stop the daemon entirely instead, use ptal stop.
`

func cmdPause(args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	state, err := store.Load(cfg.StatePath)
	if err != nil {
		return err
	}

	if len(args) == 0 {
		if paused, until := state.Paused(); paused {
			fmt.Printf("paused until %s (%s from now)\n",
				until.Local().Format("15:04"), time.Until(until).Round(time.Minute))
		} else {
			fmt.Println("not paused")
			fmt.Println("\nPause with:  ptal pause 2h")
		}
		return nil
	}
	if args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
		fmt.Print(pauseUsage)
		return nil
	}

	d, err := time.ParseDuration(args[0])
	if err != nil {
		return fmt.Errorf("not a duration: %q (try 30m, 2h, 1h30m)", args[0])
	}
	if d <= 0 {
		return fmt.Errorf("the duration must be positive; to lift a pause use `ptal resume`")
	}

	until := time.Now().Add(d)
	state.Pause(until)
	if err := state.Save(); err != nil {
		return fmt.Errorf("saving the pause: %w", err)
	}

	fmt.Printf("✓ paused for %s, until %s\n", d, until.Local().Format("15:04"))
	fmt.Println("  Still checking GitHub, just not messaging you.")
	fmt.Println("  Resume early with: ptal resume")
	return nil
}

func cmdResume() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	state, err := store.Load(cfg.StatePath)
	if err != nil {
		return err
	}

	paused, _ := state.Paused()
	if !paused {
		fmt.Println("• not paused")
		return nil
	}
	state.Resume()
	if err := state.Save(); err != nil {
		return err
	}
	fmt.Println("✓ resumed")
	return nil
}

// pauseState is a line for `status` and `doctor`, so a forgotten pause is
// never a mystery about why nothing arrives.
func pauseState(state *store.State) string {
	if paused, until := state.Paused(); paused {
		return fmt.Sprintf("paused until %s (%s left)",
			until.Local().Format("15:04"), time.Until(until).Round(time.Minute))
	}
	return ""
}

