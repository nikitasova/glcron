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
	schedules     []models.Schedule
	filtered      []models.Schedule
	cursor        int
	width         int
	height        int
	search        textinput.Model
	searching     bool
	confirmDelete bool // showing delete confirmation dialog
	confirmYes    bool // true = Yes selected, false = No selected
	deleteID      int  // ID of schedule to delete
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
		// Handle confirmation dialog
		if m.confirmDelete {
			switch msg.String() {
			case "left", "h":
				m.confirmYes = true
			case "right", "l":
				m.confirmYes = false
			case "y", "Y":
				m.confirmYes = true
				m.confirmDelete = false
				return m, func() tea.Msg {
					return deleteScheduleMsg{id: m.deleteID}
				}
			case "n", "N", "esc":
				m.confirmDelete = false
				m.confirmYes = true // reset to default
			case "enter":
				if m.confirmYes {
					m.confirmDelete = false
					return m, func() tea.Msg {
						return deleteScheduleMsg{id: m.deleteID}
					}
				} else {
					m.confirmDelete = false
					m.confirmYes = true // reset to default
				}
			}
			return m, nil
		}

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
			} else if len(m.filtered) > 0 {
				m.cursor = len(m.filtered) - 1 // wrap to last
			}
		case "down", "j":
			if m.cursor < len(m.filtered)-1 {
				m.cursor++
			} else if len(m.filtered) > 0 {
				m.cursor = 0 // wrap to first
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
				m.confirmDelete = true
				m.confirmYes = true // Yes selected by default
				m.deleteID = schedule.ID
			}
		case "A":
			if m.cursor < len(m.filtered) {
				schedule := m.filtered[m.cursor]
				return m, func() tea.Msg {
					return toggleScheduleMsg{id: schedule.ID, active: !schedule.Active}
				}
			}
		case "r":
			if m.cursor < len(m.filtered) {
				schedule := m.filtered[m.cursor]
				return m, func() tea.Msg {
					return runScheduleMsg{id: schedule.ID}
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
		case "y":
			// Yonk - copy current schedule to new
			if m.cursor < len(m.filtered) {
				schedule := &m.filtered[m.cursor]
				return m, NavigateToYonk(schedule)
			}
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

	content := strings.Join(result, "\n")

	// Overlay confirmation dialog if active
	if m.confirmDelete {
		content = m.overlayConfirmDialog(content)
	}

	return content
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
		colStatus      = 8
		colNext        = 8
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
		padRight("Next", colNext)
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
		colNextStr := padRight(truncateStr(nextRun, colNext-2), colNext)

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
			content = append(content, "  "+gray.Render("(No variables)"))
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

func (m ScheduleListModel) overlayConfirmDialog(content string) string {
	lines := strings.Split(content, "\n")

	// Dialog dimensions
	dialogWidth := 50
	contentWidth := dialogWidth - 4 // Account for borders and padding

	// Get schedule description for the message
	scheduleName := ""
	for _, s := range m.filtered {
		if s.ID == m.deleteID {
			scheduleName = s.Description
			break
		}
	}

	// Wrap the description if too long
	var descLines []string
	question := fmt.Sprintf("Delete \"%s\"?", scheduleName)
	if len(question) <= contentWidth {
		descLines = append(descLines, question)
	} else {
		// Wrap the description - first add "Delete?" then the name on separate lines
		descLines = append(descLines, "Delete?")
		descLines = append(descLines, "") // gap
		remaining := scheduleName
		for len(remaining) > 0 {
			lineLen := contentWidth
			if lineLen > len(remaining) {
				lineLen = len(remaining)
			}
			descLines = append(descLines, remaining[:lineLen])
			remaining = remaining[lineLen:]
		}
	}

	// Calculate dialog height based on content
	dialogHeight := 4 + len(descLines) // top border + empty line + desc lines + buttons + bottom border

	// Center the dialog
	startX := (m.width - dialogWidth) / 2
	startY := (len(lines) - dialogHeight) / 2

	if startX < 0 {
		startX = 0
	}
	if startY < 0 {
		startY = 0
	}

	// Build dialog lines
	borderH := strings.Repeat(BorderTop, dialogWidth-2)
	title := " Delete Schedule "
	titlePadding := (dialogWidth - 2 - len(title)) / 2
	topBorder := BorderTopLeft + strings.Repeat(BorderTop, titlePadding) + title + strings.Repeat(BorderTop, dialogWidth-2-titlePadding-len(title)) + BorderTopRight

	// Buttons - centered
	yesBtn := "[ Yes ]"
	noBtn := "[ No ]"
	if m.confirmYes {
		yesBtn = SelectedStyle.Render("[ Yes ]")
		noBtn = GrayStyle.Render("[ No ]")
	} else {
		yesBtn = GrayStyle.Render("[ Yes ]")
		noBtn = SelectedStyle.Render("[ No ]")
	}
	buttonsText := yesBtn + "   " + noBtn
	buttonsWidth := lipgloss.Width(buttonsText)
	buttonsPadding := (dialogWidth - 2 - buttonsWidth) / 2
	if buttonsPadding < 0 {
		buttonsPadding = 0
	}
	buttonsPadded := padToWidth(strings.Repeat(" ", buttonsPadding)+buttonsText, dialogWidth-2)

	var dialogLines []string
	dialogLines = append(dialogLines, topBorder)
	dialogLines = append(dialogLines, BorderLeft+strings.Repeat(" ", dialogWidth-2)+BorderRight)

	// Add description lines (centered)
	for _, descLine := range descLines {
		descPadding := (dialogWidth - 2 - len(descLine)) / 2
		if descPadding < 0 {
			descPadding = 0
		}
		descPadded := padToWidth(strings.Repeat(" ", descPadding)+descLine, dialogWidth-2)
		dialogLines = append(dialogLines, BorderLeft+descPadded+BorderRight)
	}

	dialogLines = append(dialogLines, BorderLeft+strings.Repeat(" ", dialogWidth-2)+BorderRight)
	dialogLines = append(dialogLines, BorderLeft+buttonsPadded+BorderRight)
	dialogLines = append(dialogLines, BorderBottomLeft+borderH+BorderBottomRight)

	// Overlay dialog onto content
	for i, dialogLine := range dialogLines {
		lineIdx := startY + i
		if lineIdx >= 0 && lineIdx < len(lines) {
			line := lines[lineIdx]
			// Replace portion of line with dialog
			newLine := m.insertDialogLine(line, dialogLine, startX, dialogWidth)
			lines[lineIdx] = newLine
		}
	}

	return strings.Join(lines, "\n")
}

func (m ScheduleListModel) insertDialogLine(line, dialogLine string, startX, dialogWidth int) string {
	// Convert to runes for proper handling
	lineRunes := []rune(line)

	// Ensure line is wide enough
	for len(lineRunes) < startX+dialogWidth {
		lineRunes = append(lineRunes, ' ')
	}

	// Simple approach: rebuild the line
	// Note: This is simplified and may not handle ANSI codes perfectly
	result := ""
	lineWidth := lipgloss.Width(line)

	// Pad line if needed
	if lineWidth < startX {
		result = line + strings.Repeat(" ", startX-lineWidth)
		result += dialogLine
	} else {
		// We need to insert the dialog into the line
		// For simplicity, we'll use a character-based approach
		// This works best when the background has no ANSI codes at the insertion point
		if startX < len(lineRunes) && startX+dialogWidth <= len(lineRunes) {
			result = string(lineRunes[:startX]) + dialogLine + string(lineRunes[startX+dialogWidth:])
		} else if startX < len(lineRunes) {
			result = string(lineRunes[:startX]) + dialogLine
		} else {
			result = line + strings.Repeat(" ", startX-len(lineRunes)) + dialogLine
		}
	}

	return result
}
