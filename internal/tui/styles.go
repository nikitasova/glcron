package tui

import "github.com/charmbracelet/lipgloss"

// Colors - matching default terminal colors that tview uses
var (
	ColorOrange = lipgloss.Color("208") // Orange
	ColorGreen  = lipgloss.Color("10")  // Green
	ColorRed    = lipgloss.Color("9")   // Red
	ColorYellow = lipgloss.Color("11")  // Yellow
	ColorBlue   = lipgloss.Color("12")  // Blue
	ColorGray   = lipgloss.Color("8")   // Gray
	ColorWhite  = lipgloss.Color("15")  // White
)

// Common styles used across the app
var (
	// Text styles
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorOrange)

	LabelStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorOrange)

	SelectedStyle = lipgloss.NewStyle().
			Reverse(true)

	GreenStyle = lipgloss.NewStyle().
			Foreground(ColorGreen)

	RedStyle = lipgloss.NewStyle().
			Foreground(ColorRed)

	YellowStyle = lipgloss.NewStyle().
			Foreground(ColorYellow)

	BlueStyle = lipgloss.NewStyle().
			Foreground(ColorBlue)

	GrayStyle = lipgloss.NewStyle().
			Foreground(ColorGray)

	// Border characters for manual box drawing (matching tview style)
	BorderTop         = "─"
	BorderBottom      = "─"
	BorderLeft        = "│"
	BorderRight       = "│"
	BorderTopLeft     = "┌"
	BorderTopRight    = "┐"
	BorderBottomLeft  = "└"
	BorderBottomRight = "┘"

	// Cursor style for text inputs
	CursorStyle = lipgloss.NewStyle().
			Foreground(ColorWhite).
			Background(ColorOrange)
)
