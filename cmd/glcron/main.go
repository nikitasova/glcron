package main

import (
	"fmt"
	"glcron/internal/tui"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Create the bubbletea program
	p := tea.NewProgram(
		tui.NewModel(),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	// Run the program
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
