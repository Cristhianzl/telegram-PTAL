package main

import (
	"fmt"
	"strings"

	"github.com/Cristhianzl/telegram-PTAL/internal/config"
	"github.com/Cristhianzl/telegram-PTAL/internal/engine"
)

// cmdEvents lists every alert type and whether it is currently delivered.
//
// Without this the mute settings are unusable: you cannot turn off an alert
// whose name you have no way to discover.
func cmdEvents() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	allow, err := engine.ParseKinds(strings.Join(cfg.AlertOn, ","))
	if err != nil {
		return fmt.Errorf("ALERT_ON: %w", err)
	}
	mute, err := engine.ParseKinds(strings.Join(cfg.MuteEvents, ","))
	if err != nil {
		return fmt.Errorf("MUTE_EVENTS: %w", err)
	}
	filter := engine.NewKindFilter(allow, mute)

	fmt.Print("Alert types:\n\n")
	for _, k := range engine.AllKinds {
		mark, label := "\033[32m●\033[0m", "on "
		if !filter.Allows(k) {
			mark, label = "\033[2m○\033[0m", "off"
		}
		fmt.Printf("  %s %s  %-20s %s\n", mark, label, k, k.Summary())
	}

	fmt.Println()
	if muted := filter.Muted(); len(muted) > 0 {
		names := make([]string, 0, len(muted))
		for _, k := range muted {
			names = append(names, string(k))
		}
		fmt.Printf("Muted: %s\n\n", strings.Join(names, ", "))
	}

	fmt.Println("Turn one off:      ptal config mute-events checks_failed")
	fmt.Println("Turn several off:  ptal config mute-events checks_failed,checks_fixed,gone")
	fmt.Println("Only want a few:   ptal config alert-on review_requested,mentioned")
	fmt.Println("Back to all:       ptal config mute-events \"\"")
	return nil
}
