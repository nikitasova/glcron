package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// HelpItem represents a single help entry
type HelpItem struct {
	Key         string
	Description string
}

// HelpSection represents a group of related help items
type HelpSection struct {
	Title string
	Items []HelpItem
}

// HelpModel manages the help screen state
type HelpModel struct {
	width        int
	height       int
	visible      bool
	parentScreen Screen
}

// NewHelpModel creates a new help model
func NewHelpModel() HelpModel {
	return HelpModel{
		visible: false,
	}
}

// SetSize sets the help screen dimensions
func (m *HelpModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// Show displays the help screen for a given parent screen
func (m *HelpModel) Show(parentScreen Screen) {
	m.visible = true
	m.parentScreen = parentScreen
}

// Hide hides the help screen
func (m *HelpModel) Hide() {
	m.visible = false
}

// IsVisible returns whether the help screen is visible
func (m *HelpModel) IsVisible() bool {
	return m.visible
}

// GetParentScreen returns the screen that opened help
func (m *HelpModel) GetParentScreen() Screen {
	return m.parentScreen
}

// GetHelpSections returns the help sections for the current parent screen
func (m *HelpModel) GetHelpSections() []HelpSection {
	switch m.parentScreen {
	case ScreenConfigList:
		return m.getConfigListHelp()
	case ScreenScheduleList:
		return m.getScheduleListHelp()
	case ScreenEditSchedule, ScreenNewSchedule:
		return m.getScheduleFormHelp()
	case ScreenEditConfig, ScreenNewConfig:
		return m.getConfigFormHelp()
	case ScreenQuickRun:
		return m.getQuickRunHelp()
	default:
		return m.getGeneralHelp()
	}
}

func (m *HelpModel) getConfigListHelp() []HelpSection {
	return []HelpSection{
		{
			Title: "Navigation",
			Items: []HelpItem{
				{Key: "↑/k", Description: "Move up"},
				{Key: "↓/j", Description: "Move down"},
				{Key: "Enter", Description: "Select configuration"},
			},
		},
		{
			Title: "Actions",
			Items: []HelpItem{
				{Key: "n", Description: "New configuration"},
				{Key: "e", Description: "Edit configuration"},
				{Key: "d", Description: "Delete configuration"},
			},
		},
		{
			Title: "General",
			Items: []HelpItem{
				{Key: "h", Description: "Show this help"},
				{Key: "q", Description: "Quit application"},
			},
		},
	}
}

func (m *HelpModel) getScheduleListHelp() []HelpSection {
	return []HelpSection{
		{
			Title: "Navigation",
			Items: []HelpItem{
				{Key: "↑/k", Description: "Move up"},
				{Key: "↓/j", Description: "Move down"},
				{Key: "Tab", Description: "Switch between All/Mine"},
			},
		},
		{
			Title: "Schedule Actions",
			Items: []HelpItem{
				{Key: "n", Description: "New schedule"},
				{Key: "e", Description: "Edit schedule"},
				{Key: "d", Description: "Delete schedule"},
				{Key: "Space", Description: "Toggle active/inactive"},
				{Key: "r", Description: "Run pipeline now"},
				{Key: "o", Description: "Take ownership"},
			},
		},
		{
			Title: "Other",
			Items: []HelpItem{
				{Key: "R", Description: "Quick Run (ad-hoc pipeline)"},
				{Key: "u", Description: "Refresh list"},
				{Key: "h", Description: "Show this help"},
				{Key: "Esc", Description: "Back to configs"},
				{Key: "q", Description: "Quit application"},
			},
		},
	}
}

func (m *HelpModel) getScheduleFormHelp() []HelpSection {
	return []HelpSection{
		{
			Title: "Navigation",
			Items: []HelpItem{
				{Key: "Tab", Description: "Next field"},
				{Key: "Shift+Tab", Description: "Previous field"},
				{Key: "↑/↓", Description: "Navigate in lists"},
			},
		},
		{
			Title: "Fields",
			Items: []HelpItem{
				{Key: "Enter", Description: "Open dropdown / Add variable"},
				{Key: "Space", Description: "Toggle active checkbox"},
				{Key: "Backspace", Description: "Delete variable"},
			},
		},
		{
			Title: "Actions",
			Items: []HelpItem{
				{Key: "Enter on Save", Description: "Save schedule"},
				{Key: "h", Description: "Show this help"},
				{Key: "Esc", Description: "Cancel and go back"},
			},
		},
	}
}

func (m *HelpModel) getConfigFormHelp() []HelpSection {
	return []HelpSection{
		{
			Title: "Navigation",
			Items: []HelpItem{
				{Key: "Tab", Description: "Next field"},
				{Key: "Shift+Tab", Description: "Previous field"},
			},
		},
		{
			Title: "Actions",
			Items: []HelpItem{
				{Key: "Enter on Save", Description: "Save configuration"},
				{Key: "h", Description: "Show this help"},
				{Key: "Esc", Description: "Cancel and go back"},
			},
		},
	}
}

func (m *HelpModel) getQuickRunHelp() []HelpSection {
	return []HelpSection{
		{
			Title: "Pipeline List",
			Items: []HelpItem{
				{Key: "↑/k", Description: "Move up"},
				{Key: "↓/j", Description: "Move down"},
				{Key: "R", Description: "Open run form"},
				{Key: "u", Description: "Refresh pipeline list"},
			},
		},
		{
			Title: "Run Form",
			Items: []HelpItem{
				{Key: "Tab", Description: "Next field"},
				{Key: "Shift+Tab", Description: "Previous field"},
				{Key: "Enter", Description: "Open branch selector / Start pipeline"},
				{Key: "Esc", Description: "Close form"},
			},
		},
		{
			Title: "General",
			Items: []HelpItem{
				{Key: "h", Description: "Show this help"},
				{Key: "Esc", Description: "Back to schedule list"},
				{Key: "q", Description: "Quit application"},
			},
		},
	}
}

func (m *HelpModel) getGeneralHelp() []HelpSection {
	return []HelpSection{
		{
			Title: "General",
			Items: []HelpItem{
				{Key: "h", Description: "Show/hide help"},
				{Key: "Esc", Description: "Go back / Close"},
				{Key: "q", Description: "Quit application"},
			},
		},
	}
}

// View renders the help screen
func (m *HelpModel) View() string {
	if !m.visible {
		return ""
	}

	sections := m.GetHelpSections()

	var lines []string
	contentWidth := m.width - 2 // Account for help box borders (│ on each side)

	// Title
	title := " ⌨ Keyboard Shortcuts "
	titleWidth := lipgloss.Width(title)
	borderLen := contentWidth - titleWidth
	if borderLen < 0 {
		borderLen = 0
	}
	lines = append(lines, BorderTopLeft+strings.Repeat(BorderTop, borderLen/2)+title+strings.Repeat(BorderTop, borderLen-borderLen/2)+BorderTopRight)

	// Screen indicator
	screenName := m.getScreenName()
	screenLine := "  " + GrayStyle.Render("Screen: ") + YellowStyle.Render(screenName)
	lines = append(lines, "│"+padToWidth(screenLine, contentWidth)+"│")
	lines = append(lines, "│"+strings.Repeat(" ", contentWidth)+"│")

	// Render sections
	keyStyle := lipgloss.NewStyle().Foreground(ColorOrange).Bold(true)
	sectionStyle := lipgloss.NewStyle().Foreground(ColorBlue).Bold(true)

	for i, section := range sections {
		// Section title
		sectionTitle := "  " + sectionStyle.Render("─── "+section.Title+" ───")
		lines = append(lines, "│"+padToWidth(sectionTitle, contentWidth)+"│")

		// Items
		for _, item := range section.Items {
			keyPart := keyStyle.Render(padRight(item.Key, 14))
			descPart := item.Description
			itemLine := "    " + keyPart + "  " + descPart
			lines = append(lines, "│"+padToWidth(itemLine, contentWidth)+"│")
		}

		// Add spacing between sections (except last)
		if i < len(sections)-1 {
			lines = append(lines, "│"+strings.Repeat(" ", contentWidth)+"│")
		}
	}

	// Footer
	lines = append(lines, "│"+strings.Repeat(" ", contentWidth)+"│")
	footerLine := "  " + GrayStyle.Render("Press ") + YellowStyle.Render("h") + GrayStyle.Render(" or ") + YellowStyle.Render("Esc") + GrayStyle.Render(" to close")
	lines = append(lines, "│"+padToWidth(footerLine, contentWidth)+"│")

	// Fill remaining space
	for len(lines) < m.height-1 {
		lines = append(lines, "│"+strings.Repeat(" ", contentWidth)+"│")
	}

	// Bottom border
	lines = append(lines, BorderBottomLeft+strings.Repeat(BorderTop, contentWidth)+BorderBottomRight)

	return strings.Join(lines, "\n")
}

func (m *HelpModel) getScreenName() string {
	switch m.parentScreen {
	case ScreenConfigList:
		return "Configuration List"
	case ScreenScheduleList:
		return "Schedule List"
	case ScreenEditSchedule:
		return "Edit Schedule"
	case ScreenNewSchedule:
		return "New Schedule"
	case ScreenEditConfig:
		return "Edit Configuration"
	case ScreenNewConfig:
		return "New Configuration"
	case ScreenQuickRun:
		return "Quick Pipeline Run"
	default:
		return "Unknown"
	}
}
