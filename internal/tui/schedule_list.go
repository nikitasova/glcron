package tui

import (
	"fmt"
	"glcron/internal/models"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ScheduleListModel struct {
	schedules []models.Schedule
	filtered  []models.Schedule
	cursor    int
	width     int
	height    int
	search    textinput.Model
	searching bool
}

func NewScheduleListModel() ScheduleListModel {
	ti := textinput.New()
	ti.Placeholder = "Search..."
	ti.CharLimit = 40
	ti.Width = 30

	return ScheduleListModel{
		search: ti,
	}
}

func (m *ScheduleListModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

func (m *ScheduleListModel) SetItems(schedules []models.Schedule) {
	m.schedules = schedules
	m.filtered = schedules
	if m.cursor >= len(m.filtered) && len(m.filtered) > 0 {
		m.cursor = len(m.filtered) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

func (m ScheduleListModel) Update(msg tea.Msg) (ScheduleListModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.searching {
			switch msg.String() {
			case "enter", "esc":
				m.searching = false
				m.search.Blur()
				return m, nil
			default:
				var cmd tea.Cmd
				m.search, cmd = m.search.Update(msg)
				m.filterSchedules()
				return m, cmd
			}
		}

		switch msg.String() {
		case "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.filtered)-1 {
				m.cursor++
			}
		case "enter", "e":
			if m.cursor < len(m.filtered) {
				schedule := m.filtered[m.cursor]
				return m, NavigateToEdit(&schedule)
			}
		case "c":
			return m, Navigate(ScreenNewSchedule)
		case "d":
			if m.cursor < len(m.filtered) {
				schedule := m.filtered[m.cursor]
				return m, func() tea.Msg {
					return deleteScheduleMsg{id: schedule.ID}
				}
			}
		case "A":
			if m.cursor < len(m.filtered) {
				schedule := m.filtered[m.cursor]
				return m, func() tea.Msg {
					return toggleScheduleMsg{id: schedule.ID, active: !schedule.Active}
				}
			}
		case "/":
			m.searching = true
			m.search.Focus()
			return m, textinput.Blink
		case "u":
			// Refresh schedules from GitLab
			return m, func() tea.Msg {
				return refreshSchedulesMsg{}
			}
		case "o":
			return m, Navigate(ScreenConfigList)
		}
	}

	return m, nil
}

func (m *ScheduleListModel) filterSchedules() {
	query := strings.ToLower(m.search.Value())
	if query == "" {
		m.filtered = m.schedules
		return
	}

	var filtered []models.Schedule
	for _, s := range m.schedules {
		if strings.Contains(strings.ToLower(s.Description), query) ||
			strings.Contains(strings.ToLower(s.Ref), query) ||
			strings.Contains(strings.ToLower(s.Cron), query) {
			filtered = append(filtered, s)
		}
	}
	m.filtered = filtered
	m.cursor = 0
}

func (m ScheduleListModel) View() string {
	// Split into left (2/3) and right (1/3) columns
	leftWidth := (m.width * 2) / 3
	rightWidth := m.width - leftWidth - 1

	leftLines := m.renderLeftColumn(leftWidth)
	rightLines := m.renderDetailsPanel(rightWidth)

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

func (m ScheduleListModel) renderLeftColumn(width int) []string {
	// Define styles
	headerStyle := lipgloss.NewStyle().Foreground(ColorOrange)
	grayStyle := lipgloss.NewStyle().Foreground(ColorGray)
	greenStyle := GreenStyle
	redStyle := RedStyle
	selectedStyle := SelectedStyle

	// Column widths - Description is wider now
	const (
		colActive      = 3
		colDescription = 50
		colCron        = 15
		colBranch      = 18
		colStatus      = 10
	)

	var lines []string
	indent := "   "

	// Search row
	searchIcon := headerStyle.Render("🔍 ")
	searchField := m.search.View()
	counter := grayStyle.Render(fmt.Sprintf("  %d/%d", len(m.filtered), len(m.schedules)))
	lines = append(lines, indent+searchIcon+searchField+counter)
	lines = append(lines, "")

	// Table header - render as plain text first, not styled
	headerRow := indent +
		padRight("", colActive) +
		padRight("Description", colDescription) +
		padRight("Cron", colCron) +
		padRight("Branch", colBranch) +
		padRight("Status", colStatus) +
		"Next"
	lines = append(lines, headerStyle.Render(headerRow))

	// Table rows
	for i, schedule := range m.filtered {
		// Active indicator (filled = active, empty = inactive)
		activeIcon := "○"
		activeStyle := grayStyle
		if schedule.Active {
			activeIcon = "●"
			activeStyle = greenStyle
		}

		// Pipeline status based on LastPipeline from GitLab
		// ○ gray (unknown/no pipeline), ● green (success), ● red (failed)
		// ◐ yellow (running/pending), ○ gray (canceled/skipped)
		statusIcon := "○"
		statusStyle := grayStyle
		if schedule.LastPipeline != nil && schedule.LastPipeline.Status != "" {
			switch schedule.LastPipeline.Status {
			case "success":
				statusIcon = "●"
				statusStyle = greenStyle
			case "failed":
				statusIcon = "●"
				statusStyle = redStyle
			case "running", "pending":
				statusIcon = "◐"
				statusStyle = lipgloss.NewStyle().Foreground(ColorYellow)
			case "canceled", "skipped", "manual", "scheduled":
				statusIcon = "○"
				statusStyle = grayStyle
			}
		}

		// Next run time
		nextRun := "-"
		if schedule.NextRunAt != nil {
			nextRun = formatRelativeTime(*schedule.NextRunAt)
		}

		// Build row columns - truncate description (no wrapping in table)
		colActiveStr := padRight(activeIcon, colActive)
		colDescStr := padRight(truncateStr(schedule.Description, colDescription-2), colDescription)
		colCronStr := padRight(truncateStr(schedule.Cron, colCron-2), colCron)
		colBranchStr := padRight(truncateStr(schedule.Ref, colBranch-2), colBranch)
		colStatusStr := padRight(statusIcon, colStatus)
		colNextStr := nextRun

		if i == m.cursor {
			// Selected row - rectangle highlight
			plainRow := indent + colActiveStr + colDescStr + colCronStr + colBranchStr + colStatusStr + colNextStr
			lines = append(lines, selectedStyle.Render(plainRow))
		} else {
			// Normal row
			row := indent +
				activeStyle.Render(colActiveStr) +
				colDescStr +
				colCronStr +
				colBranchStr +
				statusStyle.Render(colStatusStr) +
				colNextStr
			lines = append(lines, row)
		}
	}

	return lines
}

func (m ScheduleListModel) renderDetailsPanel(width int) []string {
	// Use shared styles
	title := TitleStyle
	label := YellowStyle.Bold(true)
	green := GreenStyle
	red := RedStyle
	blue := BlueStyle
	gray := GrayStyle

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
	if m.cursor < len(m.filtered) {
		s := m.filtered[m.cursor]

		// Active status
		activeIcon := "○"
		activeText := "Inactive"
		activeStyle := red
		if s.Active {
			activeIcon = "●"
			activeText = "Active"
			activeStyle = green
		}

		// Wrap description if too long
		descMaxWidth := width - 8 // Account for borders and padding
		desc := s.Description
		if len(desc) > descMaxWidth {
			// Wrap to multiple lines
			for len(desc) > 0 {
				lineLen := descMaxWidth
				if lineLen > len(desc) {
					lineLen = len(desc)
				}
				content = append(content, title.Render(desc[:lineLen]))
				desc = desc[lineLen:]
			}
			content = append(content, activeStyle.Render(activeIcon))
		} else {
			content = append(content, title.Render(s.Description)+" "+activeStyle.Render(activeIcon))
		}
		content = append(content, "")

		content = append(content, label.Render("Status"))
		content = append(content, "  "+activeStyle.Render(activeIcon+" "+activeText))
		content = append(content, "")

		content = append(content, label.Render("Schedule"))
		content = append(content, "  "+blue.Render("Cron:")+" "+s.Cron)
		content = append(content, "  "+blue.Render("Timezone:")+" "+s.CronTimezone)
		nextRun := "Not scheduled"
		if s.NextRunAt != nil {
			nextRun = formatDetailTime(*s.NextRunAt)
		}
		content = append(content, "  "+blue.Render("Next run:")+" "+nextRun)
		content = append(content, "")

		content = append(content, label.Render("Target"))
		content = append(content, "  "+blue.Render("Branch:")+" "+s.Ref)
		content = append(content, "")

		content = append(content, label.Render("Last Pipeline"))
		pipelineStatus := gray.Render("○ No pipeline")
		if s.LastPipeline != nil && s.LastPipeline.Status != "" {
			yellow := lipgloss.NewStyle().Foreground(ColorYellow)
			switch s.LastPipeline.Status {
			case "success":
				pipelineStatus = green.Render("● Passed")
			case "failed":
				pipelineStatus = red.Render("● Failed")
			case "running":
				pipelineStatus = yellow.Render("◐ Running")
			case "pending":
				pipelineStatus = yellow.Render("◐ Pending")
			case "canceled":
				pipelineStatus = gray.Render("○ Canceled")
			case "skipped":
				pipelineStatus = gray.Render("○ Skipped")
			case "manual":
				pipelineStatus = blue.Render("○ Manual")
			default:
				pipelineStatus = gray.Render("○ " + s.LastPipeline.Status)
			}
		}
		content = append(content, "  "+pipelineStatus)
		content = append(content, "")

		content = append(content, label.Render("Variables"))
		if len(s.Variables) == 0 {
			content = append(content, "  "+gray.Render("No variables"))
		} else {
			for _, v := range s.Variables {
				content = append(content, "  "+label.Render("▶")+" "+blue.Render(v.Key)+": "+gray.Render("*****"))
			}
		}
		content = append(content, "")

		content = append(content, label.Render("Owner"))
		ownerInfo := "Unknown"
		if s.Owner.Username != "" {
			ownerInfo = fmt.Sprintf("%s (@%s)", s.Owner.Name, s.Owner.Username)
		}
		content = append(content, "  "+ownerInfo)
	}

	for _, line := range content {
		paddedLine := " " + padToWidth(line, width-4) + " "
		lines = append(lines, "│"+paddedLine+"│")
	}

	for len(lines) < m.height-2 {
		lines = append(lines, "│"+strings.Repeat(" ", width-2)+"│")
	}

	lines = append(lines, "└"+strings.Repeat("─", width-2)+"┘")

	return lines
}

func formatRelativeTime(t time.Time) string {
	now := time.Now()
	diff := t.Sub(now)

	if diff < 0 {
		return "past"
	}
	if diff < time.Minute {
		return "<1m"
	}
	if diff < time.Hour {
		return fmt.Sprintf("%dm", int(diff.Minutes()))
	}
	if diff < 24*time.Hour {
		return fmt.Sprintf("%dh", int(diff.Hours()))
	}
	return fmt.Sprintf("%dd", int(diff.Hours()/24))
}

func formatDetailTime(t time.Time) string {
	now := time.Now()
	diff := t.Sub(now)

	if diff < 0 {
		return "Past due"
	}
	if diff < time.Minute {
		return "< 1 minute"
	}
	if diff < time.Hour {
		mins := int(diff.Minutes())
		if mins == 1 {
			return "in 1 minute"
		}
		return fmt.Sprintf("in %d minutes", mins)
	}
	if diff < 24*time.Hour {
		hours := int(diff.Hours())
		if hours == 1 {
			return "in 1 hour"
		}
		return fmt.Sprintf("in %d hours", hours)
	}
	days := int(diff.Hours() / 24)
	if days == 1 {
		return "in 1 day"
	}
	return fmt.Sprintf("in %d days", days)
}
