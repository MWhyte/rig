package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mrwhyte/rig/pkg/ui"
)

func main() {
	// Create the model
	m, err := ui.NewModel()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing: %v\n", err)
		os.Exit(1)
	}

	// Ensure cleanup on exit
	defer func() {
		if err := m.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Error cleaning up: %v\n", err)
		}
	}()

	// Start the program with alternate screen buffer
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}
