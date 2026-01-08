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

// SelectPopup represents a reusable selection list popup
type SelectPopup struct {
	Title        string
	Options      []string
	Cursor       int
	Width        int
	VisibleItems int
}

// NewSelectPopup creates a new selection popup
func NewSelectPopup(title string, options []string) *SelectPopup {
	return &SelectPopup{
		Title:        title,
		Options:      options,
		Cursor:       0,
		Width:        40,
		VisibleItems: 10,
	}
}

// WithWidth sets the popup width
func (p *SelectPopup) WithWidth(width int) *SelectPopup {
	p.Width = width
	return p
}

// WithVisibleItems sets the number of visible items
func (p *SelectPopup) WithVisibleItems(n int) *SelectPopup {
	p.VisibleItems = n
	return p
}

// SetCursor sets the current cursor position
func (p *SelectPopup) SetCursor(idx int) {
	if idx >= 0 && idx < len(p.Options) {
		p.Cursor = idx
	}
}

// MoveUp moves cursor up
func (p *SelectPopup) MoveUp() {
	if p.Cursor > 0 {
		p.Cursor--
	}
}

// MoveDown moves cursor down
func (p *SelectPopup) MoveDown() {
	if p.Cursor < len(p.Options)-1 {
		p.Cursor++
	}
}

// Selected returns the currently selected option
func (p *SelectPopup) Selected() string {
	if p.Cursor >= 0 && p.Cursor < len(p.Options) {
		return p.Options[p.Cursor]
	}
	return ""
}

// SelectedIndex returns the currently selected index
func (p *SelectPopup) SelectedIndex() int {
	return p.Cursor
}

// Render renders the popup as a list of lines
func (p *SelectPopup) Render() []string {
	popupWidth := p.Width

	// Calculate max width based on options
	for _, opt := range p.Options {
		if len(opt)+6 > popupWidth {
			popupWidth = len(opt) + 6
		}
	}

	var lines []string

	// Top border with title
	title := " " + p.Title + " "
	titleLen := lipgloss.Width(title)
	borderLen := popupWidth - titleLen - 4
	if borderLen < 0 {
		borderLen = 0
	}
	lines = append(lines, "┌─"+title+strings.Repeat("─", borderLen)+"─┐")

	// Calculate visible range
	visibleItems := p.VisibleItems
	start := p.Cursor - visibleItems/2
	if start < 0 {
		start = 0
	}
	end := start + visibleItems
	if end > len(p.Options) {
		end = len(p.Options)
		start = end - visibleItems
		if start < 0 {
			start = 0
		}
	}

	// Show scroll indicator at top
	if start > 0 {
		scrollLine := padToWidth("  ▲ more above", popupWidth-2)
		lines = append(lines, "│"+GrayStyle.Render(scrollLine)+"│")
	}

	// Options
	for i := start; i < end; i++ {
		item := p.Options[i]
		if len(item) > popupWidth-4 {
			item = item[:popupWidth-7] + "..."
		}
		itemPadded := padRight(item, popupWidth-4)
		if i == p.Cursor {
			lines = append(lines, "│ "+SelectedStyle.Render(itemPadded)+" │")
		} else {
			lines = append(lines, "│ "+itemPadded+" │")
		}
	}

	// Show scroll indicator at bottom
	if end < len(p.Options) {
		scrollLine := padToWidth("  ▼ more below", popupWidth-2)
		lines = append(lines, "│"+GrayStyle.Render(scrollLine)+"│")
	}

	// Bottom border
	lines = append(lines, "└"+strings.Repeat("─", popupWidth-2)+"┘")

	return lines
}

// RenderOverlay renders the popup as an overlay on existing content
func (p *SelectPopup) RenderOverlay(bgLines []string, startX, startY int) []string {
	popupLines := p.Render()

	var result []string
	for i, line := range bgLines {
		if i >= startY && i < startY+len(popupLines) {
			popupIdx := i - startY
			popupLine := popupLines[popupIdx]

			// Overlay the popup
			before := ""
			if startX > 0 && len(line) > startX {
				before = line[:startX]
			}
			after := ""
			afterStart := startX + lipgloss.Width(popupLine)
			if afterStart < len(line) {
				after = line[afterStart:]
			}
			result = append(result, before+popupLine+after)
		} else {
			result = append(result, line)
		}
	}

	return result
}

