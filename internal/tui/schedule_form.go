package tui

import (
	"glcron/internal/models"
	"glcron/internal/services"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type FormField int

const (
	FieldDescription FormField = iota
	FieldCron
	FieldTimezone
	FieldBranch
	FieldActive
	FieldVariables
	FieldSave
	FieldCancel
)

type ScheduleFormModel struct {
	isNew        bool
	scheduleID   int
	focusedField FormField
	width        int
	height       int

	descInput textinput.Model
	cronInput textinput.Model

	timezone    string
	branch      string
	active      bool
	timezoneIdx int
	branchIdx   int
	branches    []string

	variables     []models.Variable
	varInputs     []textinput.Model
	focusedVarIdx int

	showingPopup bool
	popupType    string
	popupCursor  int
	popupOptions []string
}

func NewScheduleFormModel() ScheduleFormModel {
	descInput := textinput.New()
	descInput.Placeholder = "Schedule description"
	descInput.CharLimit = 130
	descInput.Width = 50
	descInput.Cursor.Style = CursorStyle
	descInput.Focus()

	cronInput := textinput.New()
	cronInput.Placeholder = "0 0 * * *"
	cronInput.CharLimit = 50
	cronInput.Width = 20
	cronInput.Cursor.Style = CursorStyle

	return ScheduleFormModel{
		descInput:    descInput,
		cronInput:    cronInput,
		timezone:     "UTC",
		branch:       "main",
		active:       true,
		branches:     []string{"main", "master"},
		focusedField: FieldDescription,
	}
}

func (m *ScheduleFormModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

func (m *ScheduleFormModel) SetBranches(branches []string) {
	m.branches = branches
}

func (m *ScheduleFormModel) SetSchedule(schedule *models.Schedule, branches []string, isNew bool) {
	m.isNew = isNew
	m.branches = branches

	if schedule != nil {
		m.scheduleID = schedule.ID
		m.descInput.SetValue(schedule.Description)
		m.cronInput.SetValue(schedule.Cron)
		m.timezone = schedule.CronTimezone
		m.branch = schedule.Ref
		m.active = schedule.Active
		m.variables = make([]models.Variable, len(schedule.Variables))
		copy(m.variables, schedule.Variables)
	} else {
		m.scheduleID = 0
		m.descInput.SetValue("")
		m.cronInput.SetValue("0 0 * * *")
		m.timezone = "UTC"
		m.branch = "main"
		m.active = true
		m.variables = []models.Variable{}
	}

	for i, tz := range services.CommonTimezones {
		if tz == m.timezone {
			m.timezoneIdx = i
			break
		}
	}
	for i, b := range m.branches {
		if b == m.branch {
			m.branchIdx = i
			break
		}
	}

	m.rebuildVarInputs()
	m.focusedField = FieldDescription
	m.descInput.Focus()
	m.showingPopup = false
}

func (m *ScheduleFormModel) rebuildVarInputs() {
	m.varInputs = make([]textinput.Model, len(m.variables)+1)
	for i, v := range m.variables {
		ti := textinput.New()
		ti.SetValue(v.Key + "=" + v.Value)
		ti.Width = 40
		ti.Cursor.Style = CursorStyle
		m.varInputs[i] = ti
	}
	ti := textinput.New()
	ti.Placeholder = "KEY=value"
	ti.Width = 40
	ti.Cursor.Style = CursorStyle
	m.varInputs[len(m.variables)] = ti
}

func (m ScheduleFormModel) Update(msg tea.Msg) (ScheduleFormModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.showingPopup {
			return m.handlePopupKey(msg)
		}

		switch msg.String() {
		case "esc":
			return m, Navigate(ScreenScheduleList)
		case "tab", "down":
			m.nextField()
		case "shift+tab", "up":
			m.prevField()
		case "enter":
			return m.handleEnter()
		case "ctrl+s":
			return m.save()
		case "left":
			// Only switch buttons if on button fields, otherwise pass to text input
			if m.focusedField == FieldCancel {
				m.focusedField = FieldSave
			} else if m.focusedField == FieldDescription || m.focusedField == FieldCron || m.focusedField == FieldVariables {
				return m.handleInputKey(msg)
			}
		case "right":
			// Only switch buttons if on button fields, otherwise pass to text input
			if m.focusedField == FieldSave {
				m.focusedField = FieldCancel
			} else if m.focusedField == FieldDescription || m.focusedField == FieldCron || m.focusedField == FieldVariables {
				return m.handleInputKey(msg)
			}
		default:
			return m.handleInputKey(msg)
		}
	}

	return m, nil
}

func (m ScheduleFormModel) handlePopupKey(msg tea.KeyMsg) (ScheduleFormModel, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.showingPopup = false
	case "up", "k":
		if m.popupCursor > 0 {
			m.popupCursor--
		}
	case "down", "j":
		if m.popupCursor < len(m.popupOptions)-1 {
			m.popupCursor++
		}
	case "enter":
		m.showingPopup = false
		if m.popupType == "timezone" {
			m.timezone = m.popupOptions[m.popupCursor]
			m.timezoneIdx = m.popupCursor
		} else {
			m.branch = m.popupOptions[m.popupCursor]
			m.branchIdx = m.popupCursor
		}
	}
	return m, nil
}

func (m *ScheduleFormModel) nextField() {
	m.blurCurrent()
	switch m.focusedField {
	case FieldDescription:
		m.focusedField = FieldCron
	case FieldCron:
		m.focusedField = FieldTimezone
	case FieldTimezone:
		m.focusedField = FieldBranch
	case FieldBranch:
		m.focusedField = FieldActive
	case FieldActive:
		m.focusedField = FieldVariables
		m.focusedVarIdx = 0
	case FieldVariables:
		if m.focusedVarIdx < len(m.varInputs)-1 {
			m.focusedVarIdx++
		} else {
			m.focusedField = FieldSave
		}
	case FieldSave:
		m.focusedField = FieldCancel
	case FieldCancel:
		m.focusedField = FieldDescription
	}
	m.focusCurrent()
}

func (m *ScheduleFormModel) prevField() {
	m.blurCurrent()
	switch m.focusedField {
	case FieldDescription:
		m.focusedField = FieldCancel
	case FieldCron:
		m.focusedField = FieldDescription
	case FieldTimezone:
		m.focusedField = FieldCron
	case FieldBranch:
		m.focusedField = FieldTimezone
	case FieldActive:
		m.focusedField = FieldBranch
	case FieldVariables:
		if m.focusedVarIdx > 0 {
			m.focusedVarIdx--
		} else {
			m.focusedField = FieldActive
		}
	case FieldSave:
		m.focusedField = FieldVariables
		m.focusedVarIdx = len(m.varInputs) - 1
	case FieldCancel:
		m.focusedField = FieldSave
	}
	m.focusCurrent()
}

func (m *ScheduleFormModel) blurCurrent() {
	m.descInput.Blur()
	m.cronInput.Blur()
	for i := range m.varInputs {
		m.varInputs[i].Blur()
	}
}

func (m *ScheduleFormModel) focusCurrent() {
	switch m.focusedField {
	case FieldDescription:
		m.descInput.Focus()
	case FieldCron:
		m.cronInput.Focus()
	case FieldVariables:
		if m.focusedVarIdx < len(m.varInputs) {
			m.varInputs[m.focusedVarIdx].Focus()
		}
	}
}

func (m ScheduleFormModel) handleEnter() (ScheduleFormModel, tea.Cmd) {
	switch m.focusedField {
	case FieldTimezone:
		m.showingPopup = true
		m.popupType = "timezone"
		m.popupOptions = services.CommonTimezones
		m.popupCursor = m.timezoneIdx
	case FieldBranch:
		m.showingPopup = true
		m.popupType = "branch"
		m.popupOptions = m.branches
		m.popupCursor = m.branchIdx
	case FieldActive:
		m.active = !m.active
	case FieldVariables:
		if m.focusedVarIdx == len(m.varInputs)-1 {
			val := m.varInputs[m.focusedVarIdx].Value()
			if val != "" {
				key, value := parseKeyValue(val)
				if key != "" {
					m.variables = append(m.variables, models.Variable{
						Key:          key,
						Value:        value,
						VariableType: "env_var",
					})
					m.rebuildVarInputs()
					m.focusedVarIdx = len(m.varInputs) - 1
					m.varInputs[m.focusedVarIdx].Focus()
				}
			}
		}
	case FieldSave:
		return m.save()
	case FieldCancel:
		return m, Navigate(ScreenScheduleList)
	}
	return m, nil
}

func (m ScheduleFormModel) handleInputKey(msg tea.KeyMsg) (ScheduleFormModel, tea.Cmd) {
	var cmd tea.Cmd
	switch m.focusedField {
	case FieldDescription:
		m.descInput, cmd = m.descInput.Update(msg)
	case FieldCron:
		m.cronInput, cmd = m.cronInput.Update(msg)
	case FieldVariables:
		if m.focusedVarIdx < len(m.varInputs) {
			m.varInputs[m.focusedVarIdx], cmd = m.varInputs[m.focusedVarIdx].Update(msg)
			m.updateVariablesFromInputs()
		}
	}
	return m, cmd
}

func (m *ScheduleFormModel) updateVariablesFromInputs() {
	var newVars []models.Variable
	for i := 0; i < len(m.varInputs)-1; i++ {
		val := m.varInputs[i].Value()
		if val != "" {
			key, value := parseKeyValue(val)
			if key != "" {
				newVars = append(newVars, models.Variable{Key: key, Value: value, VariableType: "env_var"})
			}
		}
	}
	if len(newVars) < len(m.variables) {
		m.variables = newVars
		oldIdx := m.focusedVarIdx
		m.rebuildVarInputs()
		m.focusedVarIdx = minInt(oldIdx, len(m.varInputs)-1)
		if m.focusedField == FieldVariables {
			m.varInputs[m.focusedVarIdx].Focus()
		}
	}
}

func (m ScheduleFormModel) save() (ScheduleFormModel, tea.Cmd) {
	var vars []models.Variable
	for _, input := range m.varInputs {
		val := input.Value()
		if val != "" {
			key, value := parseKeyValue(val)
			if key != "" {
				vars = append(vars, models.Variable{Key: key, Value: value, VariableType: "env_var"})
			}
		}
	}

	if m.isNew {
		return m, func() tea.Msg {
			return createScheduleMsg{
				description: m.descInput.Value(),
				cron:        m.cronInput.Value(),
				timezone:    m.timezone,
				branch:      m.branch,
				active:      m.active,
				variables:   vars,
			}
		}
	}

	return m, func() tea.Msg {
		return saveScheduleMsg{
			id:          m.scheduleID,
			description: m.descInput.Value(),
			cron:        m.cronInput.Value(),
			timezone:    m.timezone,
			branch:      m.branch,
			active:      m.active,
			variables:   vars,
		}
	}
}

func (m ScheduleFormModel) View() string {
	leftWidth := (m.width * 3) / 5
	rightWidth := m.width - leftWidth - 1

	leftLines := m.renderFormPanel(leftWidth)
	rightLines := m.renderHelpPanel(rightWidth)

	// If popup is showing, overlay it
	if m.showingPopup {
		return m.renderWithPopup(leftLines, rightLines, leftWidth, rightWidth)
	}

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

func (m ScheduleFormModel) renderFormPanel(width int) []string {
	// Use shared styles
	label := LabelStyle
	green := GreenStyle
	selected := SelectedStyle

	var lines []string

	// Title with icon
	title := " ✏️ Edit Schedule "
	if m.isNew {
		title = " ➕ New Schedule "
	}
	titleWidth := lipgloss.Width(title)
	borderLen := width - titleWidth - 4
	if borderLen < 0 {
		borderLen = 0
	}
	lines = append(lines, BorderTopLeft+BorderTop+title+strings.Repeat(BorderTop, borderLen)+BorderTop+BorderTopRight)

	// Form fields with spacing
	var content []string

	// Label column width for alignment
	labelWidth := 22

	// Description field - cursor shows when focused
	descLabel := label.Render(padRight("  Description", labelWidth))
	descValue := m.descInput.View()
	content = append(content, descLabel+" "+descValue)
	content = append(content, "") // Gap

	// Cron Expression field - cursor shows when focused
	cronLabel := label.Render(padRight("  Cron Expression", labelWidth))
	cronValue := m.cronInput.View()
	content = append(content, cronLabel+" "+cronValue)
	content = append(content, "") // Gap

	// Timezone dropdown
	tzLabel := label.Render(padRight("  Timezone", labelWidth))
	tzValue := m.timezone + " ▾"
	if m.focusedField == FieldTimezone {
		content = append(content, tzLabel+" "+selected.Render(" "+tzValue+" "))
	} else {
		content = append(content, tzLabel+" "+tzValue)
	}
	content = append(content, "") // Gap

	// Target Branch dropdown
	branchLabel := label.Render(padRight("  Target Branch", labelWidth))
	branchValue := m.branch + " ▾"
	if m.focusedField == FieldBranch {
		content = append(content, branchLabel+" "+selected.Render(" "+branchValue+" "))
	} else {
		content = append(content, branchLabel+" "+branchValue)
	}
	content = append(content, "") // Gap

	// Active checkbox
	activeLabel := label.Render(padRight("  Active", labelWidth))
	if m.active {
		if m.focusedField == FieldActive {
			content = append(content, activeLabel+" "+selected.Render(" [X] "))
		} else {
			content = append(content, activeLabel+" ["+green.Render("X")+"]")
		}
	} else {
		if m.focusedField == FieldActive {
			content = append(content, activeLabel+" "+selected.Render(" [ ] "))
		} else {
			content = append(content, activeLabel+" [ ]")
		}
	}
	content = append(content, "") // Gap

	// Variables section - cursor shows when focused
	content = append(content, label.Render("  Variables"))
	for _, input := range m.varInputs {
		varValue := input.View()
		content = append(content, "    "+varValue)
	}
	content = append(content, "") // Gap

	// Buttons
	saveBtn := " 💾 Save "
	cancelBtn := " ✖ Cancel "
	if m.focusedField == FieldSave {
		content = append(content, "  "+selected.Render(saveBtn)+"   "+cancelBtn)
	} else if m.focusedField == FieldCancel {
		content = append(content, "  "+saveBtn+"   "+selected.Render(cancelBtn))
	} else {
		content = append(content, "  "+saveBtn+"   "+cancelBtn)
	}

	// Render content with borders
	for _, line := range content {
		paddedLine := " " + padToWidth(line, width-4) + " "
		lines = append(lines, "│"+paddedLine+"│")
	}

	// Fill remaining space
	for len(lines) < m.height-2 {
		lines = append(lines, "│"+strings.Repeat(" ", width-2)+"│")
	}

	lines = append(lines, "└"+strings.Repeat("─", width-2)+"┘")

	return lines
}

func (m ScheduleFormModel) renderHelpPanel(width int) []string {
	// Use shared styles
	heading := TitleStyle
	highlight := YellowStyle
	example := GreenStyle
	info := BlueStyle

	var lines []string

	// Title
	title := " Help "
	titleWidth := lipgloss.Width(title)
	borderLen := width - titleWidth - 4
	if borderLen < 0 {
		borderLen = 0
	}
	lines = append(lines, BorderTopLeft+BorderTop+title+strings.Repeat(BorderTop, borderLen)+BorderTop+BorderTopRight)

	var content []string

	content = append(content, heading.Render("Cron Expression Format"))
	content = append(content, "")
	content = append(content, highlight.Render("┌───────────── minute (0-59)"))
	content = append(content, highlight.Render("│ ┌───────────── hour (0-23)"))
	content = append(content, highlight.Render("│ │ ┌───────────── day (1-31)"))
	content = append(content, highlight.Render("│ │ │ ┌───────────── month (1-12)"))
	content = append(content, highlight.Render("│ │ │ │ ┌───────────── weekday (0-6)"))
	content = append(content, highlight.Render("│ │ │ │ │"))
	content = append(content, example.Render("* * * * *"))
	content = append(content, "")

	content = append(content, info.Render("Examples:"))
	content = append(content, "  "+example.Render("0 8 * * 1-5")+"  Weekdays 8 AM")
	content = append(content, "  "+example.Render("0 0 * * *")+"    Daily midnight")
	content = append(content, "  "+example.Render("*/15 * * * *")+" Every 15 min")
	content = append(content, "  "+example.Render("0 */2 * * *")+"  Every 2 hours")
	content = append(content, "")

	content = append(content, heading.Render("Variables"))
	content = append(content, "")
	content = append(content, "Enter as "+example.Render("KEY=value")+":")
	content = append(content, "  "+example.Render("DEPLOY_ENV=production"))
	content = append(content, "")

	content = append(content, heading.Render("Keyboard"))
	content = append(content, "")
	content = append(content, "  "+highlight.Render("↑/↓")+"       Navigate")
	content = append(content, "  "+highlight.Render("Enter")+"     Select/Toggle")
	content = append(content, "  "+highlight.Render("Ctrl+S")+"    Save")
	content = append(content, "  "+highlight.Render("Esc")+"       Cancel")

	// Render content with borders
	for _, line := range content {
		paddedLine := " " + padToWidth(line, width-4) + " "
		lines = append(lines, "│"+paddedLine+"│")
	}

	// Fill remaining space
	for len(lines) < m.height-2 {
		lines = append(lines, "│"+strings.Repeat(" ", width-2)+"│")
	}

	lines = append(lines, "└"+strings.Repeat("─", width-2)+"┘")

	return lines
}

func (m ScheduleFormModel) renderWithPopup(leftLines, rightLines []string, leftWidth, rightWidth int) string {
	selectedStyle := lipgloss.NewStyle().Reverse(true)

	// Build background first
	var bg []string
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
		bg = append(bg, left+"│"+right)
	}

	// Build popup - calculate width based on longest option
	title := " Timezone "
	if m.popupType == "branch" {
		title = " Branch "
	}

	// Find max option width
	maxOptionWidth := 30
	for _, opt := range m.popupOptions {
		if len(opt) > maxOptionWidth {
			maxOptionWidth = len(opt)
		}
	}
	popupWidth := maxOptionWidth + 6 // 2 for borders, 4 for padding
	if popupWidth < 35 {
		popupWidth = 35
	}
	if popupWidth > leftWidth-15 {
		popupWidth = leftWidth - 15
	}

	var popup []string
	titleWidth := lipgloss.Width(title)
	borderLen := popupWidth - titleWidth - 4
	if borderLen < 0 {
		borderLen = 0
	}
	popup = append(popup, "┌─"+title+strings.Repeat("─", borderLen)+"─┐")

	visibleItems := 10
	start := m.popupCursor - visibleItems/2
	if start < 0 {
		start = 0
	}
	end := start + visibleItems
	if end > len(m.popupOptions) {
		end = len(m.popupOptions)
		start = maxInt(0, end-visibleItems)
	}

	for i := start; i < end; i++ {
		item := m.popupOptions[i]
		// Truncate if too long
		if len(item) > popupWidth-4 {
			item = item[:popupWidth-7] + "..."
		}
		itemPadded := padRight(item, popupWidth-4)
		if i == m.popupCursor {
			popup = append(popup, "│ "+selectedStyle.Render(itemPadded)+" │")
		} else {
			popup = append(popup, "│ "+itemPadded+" │")
		}
	}

	popup = append(popup, "└"+strings.Repeat("─", popupWidth-2)+"┘")

	// Position popup below the field
	popupStartY := 6
	if m.popupType == "branch" {
		popupStartY = 10
	}
	popupStartX := 25

	// Build result - overlay popup on background while preserving structure
	var result []string
	for i := 0; i < len(bg); i++ {
		if i >= popupStartY && i < popupStartY+len(popup) {
			popupIdx := i - popupStartY
			// Build the line: left border + popup content + remaining space + separator + right panel
			line := "│ " + strings.Repeat(" ", popupStartX-3) + popup[popupIdx]
			// Pad to fill left panel
			lineWidth := lipgloss.Width(line)
			if lineWidth < leftWidth {
				line = line + strings.Repeat(" ", leftWidth-lineWidth)
			}
			// Add separator and right panel content
			rightContent := ""
			if i < len(rightLines) {
				rightContent = rightLines[i]
			}
			rightContent = padToWidth(rightContent, rightWidth)
			result = append(result, line+"│"+rightContent)
		} else {
			result = append(result, bg[i])
		}
	}

	return strings.Join(result, "\n")
}

func parseKeyValue(text string) (key, value string) {
	text = strings.TrimSpace(text)
	if text == "" {
		return "", ""
	}
	idx := strings.Index(text, "=")
	if idx <= 0 {
		return text, ""
	}
	key = strings.TrimSpace(text[:idx])
	if idx+1 < len(text) {
		value = strings.TrimSpace(text[idx+1:])
	}
	return key, value
}
