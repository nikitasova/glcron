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

// OverlayOnContent overlays the popup dialog onto existing content
func (p *ConfirmPopup) OverlayOnContent(content string, screenWidth, screenHeight int) string {
	contentLines := strings.Split(content, "\n")
	dialogLines := p.Render()

	dialogWidth := p.Width
	dialogHeight := len(dialogLines)

	// Center the dialog
	startX := (screenWidth - dialogWidth) / 2
	startY := (screenHeight - dialogHeight) / 2

	if startX < 0 {
		startX = 0
	}
	if startY < 0 {
		startY = 0
	}

	// Ensure content has enough lines
	for len(contentLines) < startY+dialogHeight {
		contentLines = append(contentLines, strings.Repeat(" ", screenWidth))
	}

	// Overlay dialog onto content
	for i, dialogLine := range dialogLines {
		lineIdx := startY + i
		if lineIdx >= 0 && lineIdx < len(contentLines) {
			contentLines[lineIdx] = insertDialogIntoLine(contentLines[lineIdx], dialogLine, startX, dialogWidth)
		}
	}

	return strings.Join(contentLines, "\n")
}

// insertDialogIntoLine inserts dialog content into a line at specified position
func insertDialogIntoLine(line, dialogLine string, startX, dialogWidth int) string {
	lineRunes := []rune(line)

	// Ensure line is wide enough
	for len(lineRunes) < startX+dialogWidth {
		lineRunes = append(lineRunes, ' ')
	}

	// Insert dialog into line
	if startX < len(lineRunes) && startX+dialogWidth <= len(lineRunes) {
		return string(lineRunes[:startX]) + dialogLine + string(lineRunes[startX+dialogWidth:])
	} else if startX < len(lineRunes) {
		return string(lineRunes[:startX]) + dialogLine
	}
	return line + strings.Repeat(" ", startX-len(lineRunes)) + dialogLine
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

