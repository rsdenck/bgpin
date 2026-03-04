package tui

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Config holds TUI configuration
type Config struct {
	RefreshInterval string
	FocusASN        int
	StartWithFlows  bool
}

// Start initializes and runs the TUI
func Start(config Config) error {
	// Parse refresh interval
	refreshDuration, err := time.ParseDuration(config.RefreshInterval)
	if err != nil {
		refreshDuration = 2 * time.Second
	}

	// Create initial model
	m := NewModel(config, refreshDuration)

	// Create program with alt screen
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())

	// Start background data fetching
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go m.startDataFetching(ctx)

	// Run the program
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to run TUI: %w", err)
	}

	return nil
}