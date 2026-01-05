package tui

import (
	"fmt"
	"glcron/internal/models"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ConfigListModel struct {
	configs     []models.Config
	cursor      int
	width       int
	height      int
	deletePopup *ConfirmPopup
	deleteIndex int
}

func NewConfigListModel() ConfigListModel {
	return ConfigListModel{}
}

func (m *ConfigListModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

func (m *ConfigListModel) SetItems(configs []models.Config) {
	m.configs = configs
	if m.cursor >= len(m.configs) && len(m.configs) > 0 {
		m.cursor = len(m.configs) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

func (m ConfigListModel) Update(msg tea.Msg) (ConfigListModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle delete popup
		if m.deletePopup != nil {
			switch msg.String() {
			case "left", "h":
				m.deletePopup.SelectYes()
			case "right", "l":
				m.deletePopup.SelectNo()
			case "y", "Y":
				m.deletePopup = nil
				idx := m.deleteIndex
				return m, func() tea.Msg {
					return deleteConfigMsg{index: idx}
				}
			case "n", "N", "esc":
				m.deletePopup = nil
			case "enter":
				if m.deletePopup.IsYesSelected() {
					m.deletePopup = nil
					idx := m.deleteIndex
					return m, func() tea.Msg {
						return deleteConfigMsg{index: idx}
					}
				} else {
					m.deletePopup = nil
				}
			}
			return m, nil
		}

		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.configs)-1 {
				m.cursor++
			}
		case "enter":
			if m.cursor < len(m.configs) {
				return m, func() tea.Msg {
					return selectConfigMsg{index: m.cursor}
				}
			}
		case "c":
			return m, Navigate(ScreenNewConfig)
		case "e":
			if m.cursor < len(m.configs) {
				config := m.configs[m.cursor]
				return m, NavigateToEditConfig(&config, m.cursor)
			}
		case "d":
			if m.cursor < len(m.configs) {
				config := m.configs[m.cursor]
				m.deleteIndex = m.cursor
				m.deletePopup = NewConfirmPopup(
					"Delete Config",
					fmt.Sprintf("Delete \"%s\"?", config.Name),
				).WithWidth(45)
			}
		}
	}
	return m, nil
}

func (m ConfigListModel) View() string {
	// Show delete popup as full screen if active
	if m.deletePopup != nil {
		return m.deletePopup.View(m.width, m.height)
	}

	if len(m.configs) == 0 {
		return m.renderEmptyState()
	}

	// Split into left (2/3) and right (1/3) columns
	leftWidth := (m.width * 2) / 3
	rightWidth := m.width - leftWidth - 1 // -1 for the separator

	// Build the two columns
	leftLines := m.renderTable(leftWidth)
	rightLines := m.renderDetailsPanel(rightWidth)

	// Combine columns side by side
	var result []string
	maxLines := maxInt(len(leftLines), len(rightLines))

	for i := 0; i < maxLines; i++ {
		left := ""
		if i < len(leftLines) {
			left = leftLines[i]
		}
		left = padToWidth(left, leftWidth)

		right := ""
		if i < len(rightLines) {
			right = rightLines[i]
		}
		right = padToWidth(right, rightWidth)

		result = append(result, left+"│"+right)
	}

	return strings.Join(result, "\n")
}

func (m ConfigListModel) renderEmptyState() string {
	var lines []string
	lines = append(lines, "")
	lines = append(lines, "")
	lines = append(lines, "")
	lines = append(lines, TitleStyle.Render("No configurations found"))
	lines = append(lines, "")
	lines = append(lines, "Press "+YellowStyle.Render("c")+" to add a new configuration")

	// Center each line
	var result []string
	for _, line := range lines {
		lineWidth := lipgloss.Width(line)
		padding := (m.width - lineWidth) / 2
		if padding < 0 {
			padding = 0
		}
		result = append(result, strings.Repeat(" ", padding)+line)
	}

	return strings.Join(result, "\n")
}

func (m ConfigListModel) renderTable(width int) []string {
	// Column widths
	const colName = 25

	var lines []string
	indent := "   "

	// Table header
	headerRow := indent + padRight("Name", colName) + "Project URL"
	lines = append(lines, TitleStyle.Render(headerRow))

	// Table rows
	for i, config := range m.configs {
		name := padRight(truncateStr(config.Name, colName-2), colName)
		url := truncateURL(config.ProjectURL, width-colName-5)
		row := indent + name + url

		if i == m.cursor {
			lines = append(lines, SelectedStyle.Render(row))
		} else {
			lines = append(lines, row)
		}
	}

	return lines
}

func (m ConfigListModel) renderDetailsPanel(width int) []string {
	// Use shared styles
	title := TitleStyle
	label := YellowStyle.Bold(true)

	var lines []string

	// Top border with title
	boxTitle := " Details "
	titleWidth := lipgloss.Width(boxTitle)
	borderLen := width - titleWidth - 4
	if borderLen < 0 {
		borderLen = 0
	}
	lines = append(lines, BorderTopLeft+BorderTop+boxTitle+strings.Repeat(BorderTop, borderLen)+BorderTop+BorderTopRight)

	// Content
	var content []string
	contentWidth := width - 6 // Account for borders and padding

	if m.cursor < len(m.configs) {
		config := m.configs[m.cursor]

		content = append(content, title.Render(config.Name))
		content = append(content, "")

		// Project URL - wrap if too long
		content = append(content, label.Render("Project URL"))
		url := config.ProjectURL
		for len(url) > 0 {
			lineLen := contentWidth - 2
			if lineLen > len(url) {
				lineLen = len(url)
			}
			content = append(content, "  "+url[:lineLen])
			url = url[lineLen:]
		}
		content = append(content, "")

		// Token
		content = append(content, label.Render("Token"))
		maskedToken := "****"
		if len(config.Token) > 8 {
			maskedToken = config.Token[:4] + "..." + config.Token[len(config.Token)-4:]
		}
		content = append(content, "  "+maskedToken)
		content = append(content, "")

		// Project ID - show "Not fetched" if 0
		content = append(content, label.Render("Project ID"))
		if config.ProjectID > 0 {
			content = append(content, fmt.Sprintf("  %d", config.ProjectID))
		} else {
			content = append(content, GrayStyle.Render("  Not fetched yet"))
		}
	}

	// Render content with borders
	for _, line := range content {
		paddedLine := " " + padToWidth(line, width-4) + " "
		lines = append(lines, BorderLeft+paddedLine+BorderRight)
	}

	// Fill remaining space
	for len(lines) < m.height-2 {
		lines = append(lines, BorderLeft+strings.Repeat(" ", width-2)+BorderRight)
	}

	lines = append(lines, BorderBottomLeft+strings.Repeat(BorderTop, width-2)+BorderBottomRight)

	return lines
}

func truncateURL(s string, maxLen int) string {
	if maxLen < 6 {
		return s
	}
	if len(s) <= maxLen {
		return s
	}
	return "..." + s[len(s)-maxLen+3:]
}

