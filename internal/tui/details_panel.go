package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// DetailsPanel is a reusable bordered panel for displaying details
type DetailsPanel struct {
	title   string
	width   int
	height  int
	content []string
}

// NewDetailsPanel creates a new details panel
func NewDetailsPanel(title string, width, height int) *DetailsPanel {
	return &DetailsPanel{
		title:  title,
		width:  width,
		height: height,
	}
}

// SetTitle sets the panel title
func (p *DetailsPanel) SetTitle(title string) {
	p.title = title
}

// SetSize sets the panel dimensions
func (p *DetailsPanel) SetSize(width, height int) {
	p.width = width
	p.height = height
}

// SetContent sets the content lines
func (p *DetailsPanel) SetContent(content []string) {
	p.content = content
}

// AddLine adds a single line to the content
func (p *DetailsPanel) AddLine(line string) {
	p.content = append(p.content, line)
}

// AddEmptyLine adds an empty line
func (p *DetailsPanel) AddEmptyLine() {
	p.content = append(p.content, "")
}

// AddLabel adds a styled label line
func (p *DetailsPanel) AddLabel(text string) {
	p.content = append(p.content, YellowStyle.Bold(true).Render(text))
}

// AddValue adds an indented value line
func (p *DetailsPanel) AddValue(text string) {
	p.content = append(p.content, "  "+text)
}

// AddLabelValue adds a label followed by a value on the next line
func (p *DetailsPanel) AddLabelValue(label, value string) {
	p.AddLabel(label)
	p.AddValue(value)
}

// AddKeyValue adds a key-value pair on a single line
func (p *DetailsPanel) AddKeyValue(key, value string, keyStyle lipgloss.Style) {
	p.content = append(p.content, "  "+keyStyle.Render(key+":")+" "+value)
}

// Clear clears the content
func (p *DetailsPanel) Clear() {
	p.content = nil
}

// Render renders the panel as a slice of strings
func (p *DetailsPanel) Render() []string {
	var lines []string

	// Top border with title
	boxTitle := " " + p.title + " "
	titleWidth := lipgloss.Width(boxTitle)
	borderLen := p.width - titleWidth - 4
	if borderLen < 0 {
		borderLen = 0
	}
	lines = append(lines, BorderTopLeft+BorderTop+boxTitle+strings.Repeat(BorderTop, borderLen)+BorderTop+BorderTopRight)

	// Render content with borders
	for _, line := range p.content {
		paddedLine := " " + padToWidth(line, p.width-4) + " "
		lines = append(lines, BorderLeft+paddedLine+BorderRight)
	}

	// Fill remaining space
	for len(lines) < p.height-1 {
		lines = append(lines, BorderLeft+strings.Repeat(" ", p.width-2)+BorderRight)
	}

	// Bottom border
	lines = append(lines, BorderBottomLeft+strings.Repeat(BorderTop, p.width-2)+BorderBottomRight)

	return lines
}

// RenderString renders the panel as a single string
func (p *DetailsPanel) RenderString() string {
	return strings.Join(p.Render(), "\n")
}

// WrapText wraps text to fit within a given width, returning multiple lines
func WrapText(text string, width int) []string {
	if len(text) <= width {
		return []string{text}
	}

	var lines []string
	for len(text) > 0 {
		lineLen := width
		if lineLen > len(text) {
			lineLen = len(text)
		}
		lines = append(lines, text[:lineLen])
		text = text[lineLen:]
	}
	return lines
}

// Note: padToWidth is defined in model.go and shared across components
