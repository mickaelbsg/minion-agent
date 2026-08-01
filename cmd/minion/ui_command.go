package main

import (
	"log"
	"minion/internal/ui"
)

func runUI(configPath, section string) {
	if err := ui.Run(configPath, section); err != nil {
		log.Fatalf("failed to start UI: %v", err)
	}
}
