package tui

import (
	"fmt"
	"glcron/internal/models"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// QuickRunPipelinesListLimit defines how many pipelines to fetch from API
// This value can be adjusted as needed
const QuickRunPipelinesListLimit = 10

// Column widths for pipeline list
const (
	ColStatus       = 12
	ColGap          = 2
	ColPipeline     = 10
	ColPipelineName = 22
	ColBranch       = 16
	ColTriggered    = 16
	ColStages       = 20
)

type QuickRunField int

const (
	QuickRunFieldBranch QuickRunField = iota
	QuickRunFieldVariables
	QuickRunFieldStart
	QuickRunFieldCancel
)

type QuickRunModel struct {
	width  int
	height int

	// Form state
	focusedField  QuickRunField
	showingForm   bool
	branch        string
	branchIdx     int
	branches      []string
	variables     []models.Variable
	varInputs     []textinput.Model
	focusedVarIdx int

	// Popup state
	showingPopup bool
	popupCursor  int

	// Pipeline list
	pipelines        []models.PipelineWithJobs
	selectedPipeline int
	scrollOffset     int

}

func NewQuickRunModel() QuickRunModel {
	return QuickRunModel{
		branch:       "main",
		branches:     []string{"main", "master"},
		showingForm:  false,
		focusedField: QuickRunFieldBranch,
	}
}

func (m *QuickRunModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

func (m *QuickRunModel) SetBranches(branches []string) {
	m.branches = branches
	// Find current branch in the list
	for i, b := range m.branches {
		if b == m.branch {
			m.branchIdx = i
			break
		}
	}
}

func (m *QuickRunModel) SetPipelines(pipelines []models.PipelineWithJobs) {
	m.pipelines = pipelines
	if m.selectedPipeline >= len(m.pipelines) && len(m.pipelines) > 0 {
		m.selectedPipeline = len(m.pipelines) - 1
	}
}

func (m *QuickRunModel) GetBranch() string {
	return m.branch
}

func (m *QuickRunModel) GetVariables() []models.Variable {
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
	return vars
}

func (m *QuickRunModel) rebuildVarInputs() {
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

func (m *QuickRunModel) Reset() {
	m.showingForm = false
	m.showingPopup = false
	m.focusedField = QuickRunFieldBranch
	m.selectedPipeline = 0
	m.scrollOffset = 0
	// Don't reset variables - keep them for next run
}

// adjustScroll ensures the selected pipeline is visible
func (m *QuickRunModel) adjustScroll() {
	visibleRows := m.getVisiblePipelineRows()
	if visibleRows <= 0 {
		visibleRows = 5
	}

	// Scroll up if needed
	if m.selectedPipeline < m.scrollOffset {
		m.scrollOffset = m.selectedPipeline
	}

	// Scroll down if needed
	if m.selectedPipeline >= m.scrollOffset+visibleRows {
		m.scrollOffset = m.selectedPipeline - visibleRows + 1
	}
}

// getVisiblePipelineRows calculates how many pipeline rows can be shown
func (m *QuickRunModel) getVisiblePipelineRows() int {
	// Account for borders, header, instructions, etc.
	// Height - title - instructions - empty - header - separator - bottom border
	return m.height - 7
}

func (m *QuickRunModel) ShowForm() {
	m.showingForm = true
	m.focusedField = QuickRunFieldBranch
	m.rebuildVarInputs()
}

func (m QuickRunModel) Update(msg tea.Msg) (QuickRunModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.showingPopup {
			return m.handlePopupKey(msg)
		}

		if m.showingForm {
			return m.handleFormKey(msg)
		}

		// Main view (pipeline list)
		switch msg.String() {
		case "q":
			return m, tea.Quit
		case "esc":
			return m, Navigate(ScreenScheduleList)
		case "R":
			m.ShowForm()
		case "u":
			// Manual refresh (status shown via global log panel)
			return m, func() tea.Msg {
				return refreshPipelinesMsg{}
			}
		case "up", "k":
			if m.selectedPipeline > 0 {
				m.selectedPipeline--
				m.adjustScroll()
			}
		case "down", "j":
			if m.selectedPipeline < len(m.pipelines)-1 {
				m.selectedPipeline++
				m.adjustScroll()
			}
		}
	}

	return m, nil
}

func (m QuickRunModel) handlePopupKey(msg tea.KeyMsg) (QuickRunModel, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.showingPopup = false
	case "up", "k":
		if m.popupCursor > 0 {
			m.popupCursor--
		}
	case "down", "j":
		if m.popupCursor < len(m.branches)-1 {
			m.popupCursor++
		}
	case "enter":
		m.showingPopup = false
		m.branch = m.branches[m.popupCursor]
		m.branchIdx = m.popupCursor
	}
	return m, nil
}

func (m QuickRunModel) handleFormKey(msg tea.KeyMsg) (QuickRunModel, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.showingForm = false
		return m, nil
	case "tab", "down":
		m.nextField()
	case "shift+tab", "up":
		m.prevField()
	case "enter":
		return m.handleEnter()
	case "left":
		if m.focusedField == QuickRunFieldCancel {
			m.focusedField = QuickRunFieldStart
		} else if m.focusedField == QuickRunFieldVariables {
			return m.handleInputKey(msg)
		}
	case "right":
		if m.focusedField == QuickRunFieldStart {
			m.focusedField = QuickRunFieldCancel
		} else if m.focusedField == QuickRunFieldVariables {
			return m.handleInputKey(msg)
		}
	default:
		return m.handleInputKey(msg)
	}

	return m, nil
}

func (m *QuickRunModel) nextField() {
	m.blurCurrent()
	switch m.focusedField {
	case QuickRunFieldBranch:
		m.focusedField = QuickRunFieldVariables
		m.focusedVarIdx = 0
	case QuickRunFieldVariables:
		if m.focusedVarIdx < len(m.varInputs)-1 {
			m.focusedVarIdx++
		} else {
			m.focusedField = QuickRunFieldStart
		}
	case QuickRunFieldStart:
		m.focusedField = QuickRunFieldCancel
	case QuickRunFieldCancel:
		m.focusedField = QuickRunFieldBranch
	}
	m.focusCurrent()
}

func (m *QuickRunModel) prevField() {
	m.blurCurrent()
	switch m.focusedField {
	case QuickRunFieldBranch:
		m.focusedField = QuickRunFieldCancel
	case QuickRunFieldVariables:
		if m.focusedVarIdx > 0 {
			m.focusedVarIdx--
		} else {
			m.focusedField = QuickRunFieldBranch
		}
	case QuickRunFieldStart:
		m.focusedField = QuickRunFieldVariables
		m.focusedVarIdx = len(m.varInputs) - 1
	case QuickRunFieldCancel:
		m.focusedField = QuickRunFieldStart
	}
	m.focusCurrent()
}

func (m *QuickRunModel) blurCurrent() {
	for i := range m.varInputs {
		m.varInputs[i].Blur()
	}
}

func (m *QuickRunModel) focusCurrent() {
	if m.focusedField == QuickRunFieldVariables && m.focusedVarIdx < len(m.varInputs) {
		m.varInputs[m.focusedVarIdx].Focus()
	}
}

func (m QuickRunModel) handleEnter() (QuickRunModel, tea.Cmd) {
	switch m.focusedField {
	case QuickRunFieldBranch:
		m.showingPopup = true
		m.popupCursor = m.branchIdx
	case QuickRunFieldVariables:
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
	case QuickRunFieldStart:
		m.showingForm = false
		return m, func() tea.Msg {
			return quickRunPipelineMsg{
				branch:    m.branch,
				variables: m.GetVariables(),
			}
		}
	case QuickRunFieldCancel:
		m.showingForm = false
	}
	return m, nil
}

func (m QuickRunModel) handleInputKey(msg tea.KeyMsg) (QuickRunModel, tea.Cmd) {
	var cmd tea.Cmd
	if m.focusedField == QuickRunFieldVariables && m.focusedVarIdx < len(m.varInputs) {
		m.varInputs[m.focusedVarIdx], cmd = m.varInputs[m.focusedVarIdx].Update(msg)
		m.updateVariablesFromInputs()
	}
	return m, cmd
}

func (m *QuickRunModel) updateVariablesFromInputs() {
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
		if m.focusedField == QuickRunFieldVariables {
			m.varInputs[m.focusedVarIdx].Focus()
		}
	}
}

func (m QuickRunModel) View() string {
	if m.showingForm {
		return m.renderWithForm()
	}
	return m.renderPipelineList()
}

func (m QuickRunModel) renderPipelineList() string {
	var lines []string

	// Title (status is shown in main app header)
	title := " 🚀 Quick Pipeline Run "
	titleWidth := lipgloss.Width(title)
	
	borderLen := m.width - titleWidth - 4
	if borderLen < 0 {
		borderLen = 0
	}
	lines = append(lines, BorderTopLeft+BorderTop+title+strings.Repeat(BorderTop, borderLen)+BorderTop+BorderTopRight)

	// Instructions
	instructionLine := " Press " + YellowStyle.Render("R") + " to start a new pipeline run, " +
		YellowStyle.Render("u") + " to update, " +
		YellowStyle.Render("↑↓") + " to navigate, " +
		YellowStyle.Render("Esc") + " to go back"
	lines = append(lines, "│"+padToWidth(" "+instructionLine, m.width-2)+"│")
	lines = append(lines, "│"+strings.Repeat(" ", m.width-2)+"│")

	// Pipeline list header
	headerStyle := lipgloss.NewStyle().Foreground(ColorOrange)
	headerContent := headerStyle.Render(
		padRight("Status", ColStatus) +
			strings.Repeat(" ", ColGap) +
			padRight("#ID", ColPipeline) +
			padRight("Name", ColPipelineName) +
			padRight("Branch", ColBranch) +
			padRight("Triggered", ColTriggered) +
			"Stages")
	lines = append(lines, "│ "+padToWidth(headerContent, m.width-4)+" │")
	lines = append(lines, "│"+strings.Repeat("─", m.width-2)+"│")

	// Calculate visible area
	visibleRows := m.getVisiblePipelineRows()
	needsScroll := NeedsScrollbar(len(m.pipelines), visibleRows)

	// First, collect all content lines (without scrollbar)
	var contentLines []string

	if len(m.pipelines) == 0 {
		contentLines = append(contentLines, "  "+GrayStyle.Render("No pipelines found. Press R to run a new pipeline."))
	} else {
		endIdx := m.scrollOffset + visibleRows
		if endIdx > len(m.pipelines) {
			endIdx = len(m.pipelines)
		}

		for i := m.scrollOffset; i < endIdx; i++ {
			p := m.pipelines[i]
			pipelineContentLines := m.renderPipelineRowContent(p, i == m.selectedPipeline)
			contentLines = append(contentLines, pipelineContentLines...)
		}
	}

	// Calculate scrollbar
	scrollbarHeight := visibleRows
	scrollbar := RenderScrollbar(ScrollbarConfig{
		TotalItems:   len(m.pipelines),
		VisibleItems: visibleRows,
		ScrollOffset: m.scrollOffset,
		Height:       scrollbarHeight,
	})

	// Apply scrollbar to content lines
	for i := 0; i < visibleRows; i++ {
		content := ""
		if i < len(contentLines) {
			content = contentLines[i]
		}

		scrollChar := " "
		if needsScroll && i < len(scrollbar) {
			scrollChar = scrollbar[i]
		}

		lines = append(lines, "│"+padToWidth(content, m.width-3)+scrollChar+"│")
	}

	// Bottom border
	lines = append(lines, BorderBottomLeft+strings.Repeat(BorderTop, m.width-2)+BorderBottomRight)

	return strings.Join(lines, "\n")
}

// renderPipelineRowContent returns content lines without borders (for use with scrollbar)
func (m QuickRunModel) renderPipelineRowContent(p models.PipelineWithJobs, selected bool) []string {
	var lines []string

	// Status icon
	statusIcon, statusStyle := getStatusIconAndStyle(p.Pipeline.Status)

	// Triggered by - show user or source type
	triggeredBy := getTriggerInfo(p.Pipeline.Source, p.Pipeline.User)

	// Pipeline ID with hyperlink
	pipelineID := fmt.Sprintf("#%d", p.Pipeline.ID)
	if p.Pipeline.WebURL != "" {
		// Create clickable hyperlink using OSC 8 escape sequence
		pipelineID = fmt.Sprintf("\x1b]8;;%s\x1b\\%s\x1b]8;;\x1b\\", p.Pipeline.WebURL, pipelineID)
	}

	// Pipeline name based on source
	pipelineName := getPipelineDisplayName(p.Pipeline.Source, p.Pipeline.Name)

	// Stages visualization with trigger indicator
	stagesStr := m.renderStagesWithTrigger(p.Pipeline.Source, p.Stages)

	// Main row with gap between Status and ID
	rowContent := " " +
		statusStyle.Render(padRight(statusIcon+" "+p.Pipeline.Status, ColStatus)) +
		strings.Repeat(" ", ColGap) +
		padRight(pipelineID, ColPipeline) +
		padRight(truncateStr(pipelineName, ColPipelineName-2), ColPipelineName) +
		padRight(truncateStr(p.Pipeline.Ref, ColBranch-2), ColBranch) +
		padRight(truncateStr(triggeredBy, ColTriggered-2), ColTriggered) +
		stagesStr

	if selected {
		lines = append(lines, SelectedStyle.Render(rowContent))
	} else {
		lines = append(lines, rowContent)
	}

	// Show stage details for selected pipeline (under Stages column, reverse order)
	if selected && len(p.Stages) > 0 {
		// Calculate indent to align under Stages column
		stagesIndent := 1 + ColStatus + ColGap + ColPipeline + ColPipelineName + ColBranch + ColTriggered
		indent := strings.Repeat(" ", stagesIndent)

		// Show trigger source info if triggered by another pipeline
		if isTriggerSource(p.Pipeline.Source) {
			// Show the upstream project name if available, otherwise show source type
			triggerSource := p.UpstreamProjectName
			if triggerSource == "" {
				triggerSource = getTriggerSourceDescription(p.Pipeline.Source)
			}
			triggerInfo := indent + "↳ triggered by: " + BlueStyle.Render(triggerSource)
			if p.Pipeline.User != nil && p.Pipeline.User.Username != "" {
				triggerInfo += " " + GrayStyle.Render("(@"+p.Pipeline.User.Username+")")
			}
			lines = append(lines, triggerInfo)
			// Gap after trigger info
			lines = append(lines, "")
		}

		// Show stages in reverse order (from last to first)
		for i := len(p.Stages) - 1; i >= 0; i-- {
			stage := p.Stages[i]
			stageIcon, stageStyle := getStatusIconAndStyle(stage.Status)
			stageLine := indent + "- " + stage.Name + ": " + stageStyle.Render(stageIcon+" "+stage.Status)
			lines = append(lines, stageLine)
		}
		// pb-1: padding bottom
		lines = append(lines, "")
	}

	return lines
}

// getTriggerInfo returns who/what triggered the pipeline
func getTriggerInfo(source string, user *models.User) string {
	// If we have a user, show their username
	if user != nil && user.Username != "" {
		return "@" + user.Username
	}
	
	// Otherwise show the source type
	return getSourceDisplayName(source)
}

// getSourceDisplayName returns a human-readable name for the pipeline source
func getSourceDisplayName(source string) string {
	switch source {
	case "push":
		return "push"
	case "web":
		return "web"
	case "trigger":
		return "API trigger"
	case "schedule":
		return "schedule"
	case "pipeline":
		return "pipeline"
	case "parent_pipeline":
		return "parent pipeline"
	case "cross_project_pipeline":
		return "cross-project"
	case "merge_request_event":
		return "merge request"
	case "external_pull_request_event":
		return "external PR"
	case "chat":
		return "chat"
	case "webide":
		return "Web IDE"
	case "external":
		return "external"
	case "ondemand_dast_scan":
		return "DAST scan"
	case "ondemand_dast_validation":
		return "DAST validation"
	default:
		if source == "" {
			return "-"
		}
		return source
	}
}

// getPipelineDisplayName returns a display name for the pipeline
func getPipelineDisplayName(source, name string) string {
	// If we have a name (commit title), use it for non-scheduled pipelines
	if name != "" {
		return name
	}
	
	// Fallback based on source
	switch source {
	case "schedule":
		return "Scheduled"
	case "merge_request_event":
		return "Merge request"
	case "push":
		return "Push"
	case "web":
		return "Manual (web)"
	case "trigger", "pipeline", "parent_pipeline", "cross_project_pipeline":
		return "Triggered"
	default:
		if source == "" {
			return "-"
		}
		return source
	}
}

// getTriggerSourceDescription returns a descriptive text for trigger sources
func getTriggerSourceDescription(source string) string {
	switch source {
	case "trigger":
		return "external API trigger"
	case "pipeline":
		return "upstream pipeline"
	case "parent_pipeline":
		return "parent project pipeline"
	case "cross_project_pipeline":
		return "cross-project pipeline"
	default:
		return source
	}
}

// getStatusIconAndStyle returns the icon and style for a given status
func getStatusIconAndStyle(status string) (string, lipgloss.Style) {
	switch status {
	case "success":
		return "●", GreenStyle
	case "failed":
		return "●", RedStyle
	case "running":
		return "◐", YellowStyle
	case "pending":
		return "◐", YellowStyle
	case "canceled":
		return "○", GrayStyle
	case "skipped":
		return "○", GrayStyle
	default:
		return "○", GrayStyle
	}
}

func (m QuickRunModel) renderStagesCompact(stages []models.StageInfo) string {
	if len(stages) == 0 {
		return GrayStyle.Render("(no stages)")
	}

	var parts []string
	for _, stage := range stages {
		icon, style := getStatusIconAndStyle(stage.Status)
		parts = append(parts, style.Render(icon))
	}

	// Join with spaced dashes: o — o — o — o
	return strings.Join(parts, " — ")
}

// renderStagesWithTrigger renders stages with trigger indicator if pipeline was triggered
func (m QuickRunModel) renderStagesWithTrigger(source string, stages []models.StageInfo) string {
	stagesStr := m.renderStagesCompact(stages)
	
	// Check if pipeline was triggered by another pipeline/project
	// Sources that indicate external trigger: "trigger", "pipeline", "parent_pipeline", "cross_project_pipeline"
	switch source {
	case "trigger", "pipeline", "parent_pipeline", "cross_project_pipeline":
		// Show trigger indicator: ○ → stages
		return GrayStyle.Render("○") + " → " + stagesStr
	default:
		return stagesStr
	}
}

func (m QuickRunModel) renderWithForm() string {
	// Split vertically: top form, bottom pipeline list
	formHeight := 15
	listHeight := m.height - formHeight - 2

	formLines := m.renderFormPanel(m.width, formHeight)
	listLines := m.renderPipelineListPanel(m.width, listHeight)

	var result []string
	result = append(result, formLines...)
	result = append(result, listLines...)

	// If showing branch popup, overlay it
	if m.showingPopup {
		return m.renderWithPopup(result)
	}

	return strings.Join(result, "\n")
}

func (m QuickRunModel) renderWithPopup(bgLines []string) string {
	selectedStyle := lipgloss.NewStyle().Reverse(true)

	// Build popup
	title := " Branch "
	popupWidth := 40
	for _, opt := range m.branches {
		if len(opt)+6 > popupWidth {
			popupWidth = len(opt) + 6
		}
	}
	if popupWidth > m.width-20 {
		popupWidth = m.width - 20
	}

	var popup []string
	titleLen := lipgloss.Width(title)
	borderLen := popupWidth - titleLen - 4
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
	if end > len(m.branches) {
		end = len(m.branches)
		start = maxInt(0, end-visibleItems)
	}

	// Show scroll indicator at top
	if start > 0 {
		scrollLine := padRight("  ▲ more above", popupWidth-4)
		popup = append(popup, "│ "+GrayStyle.Render(scrollLine)+" │")
	}

	for i := start; i < end; i++ {
		item := m.branches[i]
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

	// Show scroll indicator at bottom
	if end < len(m.branches) {
		scrollLine := padRight("  ▼ more below", popupWidth-4)
		popup = append(popup, "│ "+GrayStyle.Render(scrollLine)+" │")
	}

	popup = append(popup, "└"+strings.Repeat("─", popupWidth-2)+"┘")

	// Position popup - reconstruct lines completely to avoid ANSI slicing issues
	// bgLines have format: │<content>│ (form panel borders)
	// Main grid will wrap with additional │...│
	popupStartY := 2
	popupStartX := 22
	innerWidth := m.width - 2 // Width inside the form panel borders (│ on each side)

	var result []string
	for i, line := range bgLines {
		if i >= popupStartY && i < popupStartY+len(popup) {
			popupIdx := i - popupStartY
			popupLine := popup[popupIdx]
			popupLineWidth := lipgloss.Width(popupLine)

			// Build: │ + padding + popup + padding + │
			beforePadding := strings.Repeat(" ", popupStartX)
			remainingWidth := innerWidth - popupStartX - popupLineWidth
			if remainingWidth < 0 {
				remainingWidth = 0
			}
			afterPadding := strings.Repeat(" ", remainingWidth)
			newLine := "│" + beforePadding + popupLine + afterPadding + "│"
			result = append(result, newLine)
		} else {
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}

func (m QuickRunModel) renderFormPanel(width, height int) []string {
	label := LabelStyle
	selected := SelectedStyle

	var lines []string

	// Title
	title := " 🚀 New Pipeline Run "
	titleWidth := lipgloss.Width(title)
	borderLen := width - titleWidth - 4
	if borderLen < 0 {
		borderLen = 0
	}
	lines = append(lines, BorderTopLeft+BorderTop+title+strings.Repeat(BorderTop, borderLen)+BorderTop+BorderTopRight)

	var content []string
	labelWidth := 18

	// Branch field
	branchLabel := label.Render(padRight("  Branch", labelWidth))
	branchValue := m.branch + " ▾"
	if m.focusedField == QuickRunFieldBranch {
		content = append(content, branchLabel+" "+selected.Render(" "+branchValue+" "))
	} else {
		content = append(content, branchLabel+" "+branchValue)
	}
	content = append(content, "")

	// Variables section
	content = append(content, label.Render("  Variables"))
	for _, input := range m.varInputs {
		varValue := input.View()
		content = append(content, "    "+varValue)
	}
	content = append(content, "")

	// Buttons
	startBtn := " 🚀 Start Pipeline "
	cancelBtn := " ✖ Cancel "
	if m.focusedField == QuickRunFieldStart {
		content = append(content, "  "+selected.Render(startBtn)+"   "+cancelBtn)
	} else if m.focusedField == QuickRunFieldCancel {
		content = append(content, "  "+startBtn+"   "+selected.Render(cancelBtn))
	} else {
		content = append(content, "  "+startBtn+"   "+cancelBtn)
	}

	// Render content with borders
	for _, line := range content {
		paddedLine := " " + padToWidth(line, width-4) + " "
		lines = append(lines, "│"+paddedLine+"│")
	}

	// Fill remaining space
	for len(lines) < height-1 {
		lines = append(lines, "│"+strings.Repeat(" ", width-2)+"│")
	}

	lines = append(lines, "├"+strings.Repeat("─", width-2)+"┤")

	return lines
}

func (m QuickRunModel) renderPipelineListPanel(width, height int) []string {
	var lines []string

	// Header
	headerStyle := lipgloss.NewStyle().Foreground(ColorOrange)
	headerContent := headerStyle.Render("Recent Pipelines")
	lines = append(lines, "│ "+padToWidth(headerContent, width-4)+" │")
	lines = append(lines, "│"+strings.Repeat("─", width-2)+"│")

	// Pipeline rows
	if len(m.pipelines) == 0 {
		emptyLine := "│" + padToWidth("  "+GrayStyle.Render("No pipelines yet."), width-2) + "│"
		lines = append(lines, emptyLine)
	} else {
		for i, p := range m.pipelines {
			statusIcon, statusStyle := getStatusIconAndStyle(p.Pipeline.Status)

			// Triggered by
			triggeredBy := getTriggerInfo(p.Pipeline.Source, p.Pipeline.User)

			pipelineID := fmt.Sprintf("#%d", p.Pipeline.ID)

			// Pipeline name
			pipelineName := getPipelineDisplayName(p.Pipeline.Source, p.Pipeline.Name)

			stagesStr := m.renderStagesWithTrigger(p.Pipeline.Source, p.Stages)

			rowContent := fmt.Sprintf(" %s  %s %s %s %s %s",
				statusStyle.Render(statusIcon),
				padRight(truncateStr(pipelineID, 8), 9),
				padRight(truncateStr(pipelineName, 16), 17),
				padRight(truncateStr(p.Pipeline.Ref, 12), 13),
				padRight(truncateStr(triggeredBy, 12), 13),
				stagesStr,
			)

			if i == m.selectedPipeline {
				lines = append(lines, "│"+SelectedStyle.Render(padToWidth(rowContent, width-2))+"│")
			} else {
				lines = append(lines, "│"+padToWidth(rowContent, width-2)+"│")
			}
		}
	}

	// Fill remaining space
	for len(lines) < height-1 {
		lines = append(lines, "│"+strings.Repeat(" ", width-2)+"│")
	}

	lines = append(lines, BorderBottomLeft+strings.Repeat(BorderTop, width-2)+BorderBottomRight)

	return lines
}

