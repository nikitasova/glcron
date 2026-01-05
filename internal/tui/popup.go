package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ConfirmPopup represents a reusable confirmation dialog
type ConfirmPopup struct {
	Title       string
	Message     []string // Multiple lines of message
	YesText     string
	NoText      string
	YesSelected bool
	Width       int
}

// NewConfirmPopup creates a new confirmation popup with defaults
func NewConfirmPopup(title string, message ...string) *ConfirmPopup {
	return &ConfirmPopup{
		Title:       title,
		Message:     message,
		YesText:     "Yes",
		NoText:      "No",
		YesSelected: true,
		Width:       50,
	}
}

// WithButtons sets custom button text
func (p *ConfirmPopup) WithButtons(yes, no string) *ConfirmPopup {
	p.YesText = yes
	p.NoText = no
	return p
}

// WithWidth sets the dialog width
func (p *ConfirmPopup) WithWidth(width int) *ConfirmPopup {
	p.Width = width
	return p
}

// SelectYes selects the Yes button
func (p *ConfirmPopup) SelectYes() {
	p.YesSelected = true
}

// SelectNo selects the No button
func (p *ConfirmPopup) SelectNo() {
	p.YesSelected = false
}

// Toggle switches between Yes and No
func (p *ConfirmPopup) Toggle() {
	p.YesSelected = !p.YesSelected
}

// IsYesSelected returns true if Yes is selected
func (p *ConfirmPopup) IsYesSelected() bool {
	return p.YesSelected
}

// Render renders the popup dialog box
func (p *ConfirmPopup) Render() []string {
	dialogWidth := p.Width
	contentWidth := dialogWidth - 4 // borders + padding

	var lines []string

	// Top border with title
	title := " " + p.Title + " "
	titleLen := len(title)
	leftPad := (dialogWidth - 2 - titleLen) / 2
	rightPad := dialogWidth - 2 - titleLen - leftPad
	topBorder := BorderTopLeft + strings.Repeat(BorderTop, leftPad) + title + strings.Repeat(BorderTop, rightPad) + BorderTopRight
	lines = append(lines, topBorder)

	// Empty line after title
	lines = append(lines, BorderLeft+strings.Repeat(" ", dialogWidth-2)+BorderRight)

	// Message lines (centered)
	for _, msg := range p.Message {
		// Word wrap if needed
		wrapped := wrapText(msg, contentWidth)
		for _, line := range wrapped {
			padding := (dialogWidth - 2 - lipgloss.Width(line)) / 2
			if padding < 0 {
				padding = 0
			}
			paddedLine := strings.Repeat(" ", padding) + line
			paddedLine = padToWidth(paddedLine, dialogWidth-2)
			lines = append(lines, BorderLeft+paddedLine+BorderRight)
		}
	}

	// Empty line before buttons
	lines = append(lines, BorderLeft+strings.Repeat(" ", dialogWidth-2)+BorderRight)

	// Buttons
	yesBtn := "[ " + p.YesText + " ]"
	noBtn := "[ " + p.NoText + " ]"
	if p.YesSelected {
		yesBtn = SelectedStyle.Render("[ " + p.YesText + " ]")
		noBtn = GrayStyle.Render("[ " + p.NoText + " ]")
	} else {
		yesBtn = GrayStyle.Render("[ " + p.YesText + " ]")
		noBtn = SelectedStyle.Render("[ " + p.NoText + " ]")
	}

	buttonsText := yesBtn + "   " + noBtn
	buttonsWidth := lipgloss.Width(buttonsText)
	buttonsPadding := (dialogWidth - 2 - buttonsWidth) / 2
	if buttonsPadding < 0 {
		buttonsPadding = 0
	}
	buttonLine := padToWidth(strings.Repeat(" ", buttonsPadding)+buttonsText, dialogWidth-2)
	lines = append(lines, BorderLeft+buttonLine+BorderRight)

	// Empty line after buttons
	lines = append(lines, BorderLeft+strings.Repeat(" ", dialogWidth-2)+BorderRight)

	// Bottom border
	lines = append(lines, BorderBottomLeft+strings.Repeat(BorderTop, dialogWidth-2)+BorderBottomRight)

	return lines
}

// View renders the popup centered on a full screen
func (p *ConfirmPopup) View(screenWidth, screenHeight int) string {
	dialogLines := p.Render()
	dialogWidth := p.Width
	dialogHeight := len(dialogLines)

	// Calculate vertical centering
	topPadding := (screenHeight - dialogHeight) / 2
	if topPadding < 0 {
		topPadding = 0
	}

	// Calculate horizontal centering
	leftPadding := (screenWidth - dialogWidth) / 2
	if leftPadding < 0 {
		leftPadding = 0
	}

	var lines []string

	// Add top padding (empty lines)
	for i := 0; i < topPadding; i++ {
		lines = append(lines, strings.Repeat(" ", screenWidth))
	}

	// Add dialog lines with horizontal padding
	for _, dialogLine := range dialogLines {
		line := strings.Repeat(" ", leftPadding) + dialogLine
		// Pad to full width
		if lipgloss.Width(line) < screenWidth {
			line += strings.Repeat(" ", screenWidth-lipgloss.Width(line))
		}
		lines = append(lines, line)
	}

	// Add bottom padding (empty lines to fill screen)
	for len(lines) < screenHeight {
		lines = append(lines, strings.Repeat(" ", screenWidth))
	}

	return strings.Join(lines, "\n")
}

// wrapText wraps text to fit within maxWidth
func wrapText(text string, maxWidth int) []string {
	if len(text) <= maxWidth {
		return []string{text}
	}

	var lines []string
	remaining := text

	for len(remaining) > 0 {
		if len(remaining) <= maxWidth {
			lines = append(lines, remaining)
			break
		}

		// Find last space within maxWidth
		cutPoint := maxWidth
		for i := maxWidth; i > 0; i-- {
			if remaining[i] == ' ' {
				cutPoint = i
				break
			}
		}

		lines = append(lines, remaining[:cutPoint])
		remaining = strings.TrimLeft(remaining[cutPoint:], " ")
	}

	return lines
}

