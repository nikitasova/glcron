package tui

import (
	"fmt"
	"glcron/internal/models"
	"glcron/internal/services"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// App constants
const AppName = "glcron"

// AppVersion is set at build time via -ldflags
var AppVersion = "dev"

// Screen represents the current view
type Screen int

const (
	ScreenConfigList Screen = iota
	ScreenScheduleList
	ScreenEditSchedule
	ScreenNewSchedule
	ScreenEditConfig
	ScreenNewConfig
)

// Model is the main application model
type Model struct {
	// Services
	configService services.ConfigServiceInterface
	gitlabService services.GitLabServiceInterface

	// State
	screen            Screen
	width             int
	height            int
	configs           []models.Config
	currentConfigIdx  int
	schedules         []models.Schedule
	filteredSchedules []models.Schedule
	branches          []string
	statusMsg         string
	statusType        string
	loading           bool

	// Sub-models
	configList   ConfigListModel
	scheduleList ScheduleListModel
	scheduleForm ScheduleFormModel
	configForm   ConfigFormModel
}

// NewModel creates a new application model
func NewModel() Model {
	configService := services.NewConfigService()
	gitlabService := services.NewGitLabService()

	m := Model{
		configService:    configService,
		gitlabService:    gitlabService,
		screen:           ScreenConfigList,
		currentConfigIdx: -1,
		branches:         []string{"main", "master"},
	}

	m.configList = NewConfigListModel()
	m.scheduleList = NewScheduleListModel()
	m.scheduleForm = NewScheduleFormModel()
	m.configForm = NewConfigFormModel()

	return m
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		tea.EnterAltScreen,
		m.loadConfigs,
	)
}

func (m Model) loadConfigs() tea.Msg {
	configFile, err := m.configService.Load()
	if err != nil {
		return errMsg{err}
	}
	return configsLoadedMsg{configs: configFile.Configs}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.screen == ScreenConfigList {
				return m, tea.Quit
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		contentHeight := m.height - 4
		m.configList.SetSize(m.width-2, contentHeight)
		m.scheduleList.SetSize(m.width-2, contentHeight)
		m.scheduleForm.SetSize(m.width-2, contentHeight)
		m.configForm.SetSize(m.width-2, contentHeight)

	case configsLoadedMsg:
		m.configs = msg.configs
		m.configList.SetItems(m.configs)
		m.statusMsg = ""

	case schedulesLoadedMsg:
		m.schedules = msg.schedules
		m.filteredSchedules = msg.schedules
		m.scheduleList.SetItems(m.filteredSchedules)
		m.loading = false
		m.statusMsg = ""

	case branchesLoadedMsg:
		m.branches = msg.branches
		m.scheduleForm.SetBranches(m.branches)

	case configSelectedMsg:
		m.schedules = msg.schedules
		m.filteredSchedules = msg.schedules
		m.branches = msg.branches
		m.scheduleList.SetItems(m.filteredSchedules)
		m.scheduleForm.SetBranches(m.branches)
		m.loading = false
		m.statusMsg = ""
		m.screen = ScreenScheduleList

		// Save config with updated ProjectID
		if msg.updatedConfig != nil && m.currentConfigIdx >= 0 && m.currentConfigIdx < len(m.configs) {
			m.configs[m.currentConfigIdx] = *msg.updatedConfig
			m.configList.SetItems(m.configs)
			// Save to file
			configFile := &models.ConfigFile{Configs: m.configs}
			_ = m.configService.Save(configFile)
		}

	case schedulesSavedMsg:
		m.schedules = msg.schedules
		m.filteredSchedules = msg.schedules
		m.scheduleList.SetItems(m.filteredSchedules)
		m.loading = false
		m.statusMsg = msg.message
		m.statusType = "success"
		m.screen = ScreenScheduleList
		return m, ClearStatusAfter(10 * time.Second)

	case configSavedMsg:
		m.configs = msg.configs
		m.configList.SetItems(m.configs)
		m.loading = false
		m.statusMsg = msg.message
		m.statusType = "success"
		m.screen = ScreenConfigList
		return m, ClearStatusAfter(10 * time.Second)

	case errMsg:
		m.statusMsg = msg.err.Error()
		m.statusType = "error"
		m.loading = false
		return m, ClearStatusAfter(10 * time.Second)

	case statusMsg:
		m.statusMsg = msg.text
		m.statusType = msg.msgType
		return m, ClearStatusAfter(10 * time.Second)

	case clearStatusMsg:
		m.statusMsg = ""
		m.statusType = ""

	case navigateMsg:
		return m.handleNavigation(msg)

	case saveScheduleMsg:
		return m.handleSaveSchedule(msg)

	case createScheduleMsg:
		return m.handleCreateSchedule(msg)

	case deleteScheduleMsg:
		return m.handleDeleteSchedule(msg)

	case toggleScheduleMsg:
		return m.handleToggleSchedule(msg)

	case runScheduleMsg:
		return m.handleRunSchedule(msg)

	case refreshSchedulesMsg:
		return m.handleRefreshSchedules()

	case saveConfigMsg:
		return m.handleSaveConfig(msg)

	case deleteConfigMsg:
		return m.handleDeleteConfig(msg)

	case selectConfigMsg:
		return m.handleSelectConfig(msg)
	}

	switch m.screen {
	case ScreenConfigList:
		var cmd tea.Cmd
		m.configList, cmd = m.configList.Update(msg)
		cmds = append(cmds, cmd)

	case ScreenScheduleList:
		var cmd tea.Cmd
		m.scheduleList, cmd = m.scheduleList.Update(msg)
		cmds = append(cmds, cmd)

	case ScreenEditSchedule, ScreenNewSchedule:
		var cmd tea.Cmd
		m.scheduleForm, cmd = m.scheduleForm.Update(msg)
		cmds = append(cmds, cmd)

	case ScreenEditConfig, ScreenNewConfig:
		var cmd tea.Cmd
		m.configForm, cmd = m.configForm.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// Build the bordered grid like tview does
	return m.renderGrid()
}

// renderGrid creates a bordered layout matching tview's Grid with SetBorders(true)
func (m Model) renderGrid() string {
	w := m.width
	h := m.height

	// Header content
	header := m.renderHeader()

	// Main content based on current screen
	var content string
	switch m.screen {
	case ScreenConfigList:
		content = m.configList.View()
	case ScreenScheduleList:
		content = m.scheduleList.View()
	case ScreenEditSchedule, ScreenNewSchedule:
		content = m.scheduleForm.View()
	case ScreenEditConfig, ScreenNewConfig:
		content = m.configForm.View()
	}

	// Footer content (legend)
	footer := m.renderLegend()

	// Build the grid with box-drawing characters
	var sb strings.Builder
	borderH := strings.Repeat(BorderTop, w-2)

	// Top border: ┌──────────────────────┐
	sb.WriteString(BorderTopLeft + borderH + BorderTopRight + "\n")

	// Header row with borders: │ header │
	sb.WriteString(BorderLeft + padToWidth(header, w-2) + BorderRight + "\n")

	// Separator: ├──────────────────────┤
	sb.WriteString("├" + borderH + "┤\n")

	// Content area
	contentHeight := h - 6 // top border + header + sep + sep + footer + bottom border
	contentLines := strings.Split(content, "\n")

	for i := 0; i < contentHeight; i++ {
		line := ""
		if i < len(contentLines) {
			line = contentLines[i]
		}
		// Ensure line fits and pad it
		line = truncateOrPad(line, w-2)
		sb.WriteString(BorderLeft + line + BorderRight + "\n")
	}

	// Separator before footer: ├──────────────────────┤
	sb.WriteString("├" + borderH + "┤\n")

	// Footer row: │ legend │
	sb.WriteString(BorderLeft + padToWidth(footer, w-2) + BorderRight + "\n")

	// Bottom border: └──────────────────────┘
	sb.WriteString(BorderBottomLeft + borderH + BorderBottomRight)

	return sb.String()
}

func (m Model) renderHeader() string {
	// Format: " glcron 0.1.0 - configname" on left, status on right
	orange := TitleStyle
	green := GreenStyle

	left := " " + orange.Render(AppName) + " " + AppVersion
	if m.currentConfigIdx >= 0 && m.currentConfigIdx < len(m.configs) {
		left += " - " + green.Render(m.configs[m.currentConfigIdx].Name)
	}

	right := ""
	if m.statusMsg != "" {
		var style lipgloss.Style
		switch m.statusType {
		case "success":
			style = GreenStyle
		case "error":
			style = RedStyle
		case "warning":
			style = YellowStyle
		default:
			style = lipgloss.NewStyle()
		}
		right = style.Render(m.statusMsg) + " "
	}

	// Calculate the visible widths (without ANSI codes)
	leftWidth := lipgloss.Width(left)
	rightWidth := lipgloss.Width(right)
	padding := m.width - 2 - leftWidth - rightWidth
	if padding < 0 {
		padding = 0
	}

	return left + strings.Repeat(" ", padding) + right
}

func (m Model) renderLegend() string {
	yellow := YellowStyle

	var items []string

	switch m.screen {
	case ScreenConfigList:
		items = []string{
			yellow.Render("↑↓") + " Navigate",
			yellow.Render("Enter") + " Select",
			yellow.Render("c") + " Create",
			yellow.Render("e") + " Edit",
			yellow.Render("d") + " Delete",
			yellow.Render("q") + " Quit",
		}
	case ScreenScheduleList:
		items = []string{
			yellow.Render("↑↓") + " Navigate",
			yellow.Render("/") + " Search",
			yellow.Render("e") + " Edit",
			yellow.Render("c") + " Create",
			yellow.Render("d") + " Delete",
			yellow.Render("r") + " Run Pipeline",
			yellow.Render("A") + " Toggle",
			yellow.Render("u") + " Update",
			yellow.Render("o") + " Configs",
			yellow.Render("q") + " Quit",
		}
	case ScreenEditSchedule, ScreenNewSchedule:
		items = []string{
			yellow.Render("↑↓") + " Navigate",
			yellow.Render("Enter") + " Select/Toggle",
			yellow.Render("Ctrl+S") + " Save",
			yellow.Render("Esc") + " Cancel",
		}
	case ScreenEditConfig, ScreenNewConfig:
		items = []string{
			yellow.Render("↑↓") + " Navigate",
			yellow.Render("Tab") + " Next",
			yellow.Render("Ctrl+S") + " Save",
			yellow.Render("Esc") + " Cancel",
		}
	}

	legend := strings.Join(items, "  │  ")

	// Center the legend (using visible width for proper centering)
	width := m.width - 2
	legendWidth := lipgloss.Width(legend)
	if legendWidth >= width {
		return legend
	}
	padding := (width - legendWidth) / 2
	return strings.Repeat(" ", padding) + legend
}

// Navigation handlers
func (m Model) handleNavigation(msg navigateMsg) (tea.Model, tea.Cmd) {
	switch msg.screen {
	case ScreenConfigList:
		m.screen = ScreenConfigList
		m.configList.SetItems(m.configs)

	case ScreenScheduleList:
		m.screen = ScreenScheduleList
		m.scheduleList.SetItems(m.filteredSchedules)

	case ScreenEditSchedule:
		m.screen = ScreenEditSchedule
		m.scheduleForm.SetSchedule(msg.schedule, m.branches, false)

	case ScreenNewSchedule:
		m.screen = ScreenNewSchedule
		m.scheduleForm.SetSchedule(nil, m.branches, true)

	case ScreenEditConfig:
		m.screen = ScreenEditConfig
		m.configForm.SetConfig(msg.config, false)

	case ScreenNewConfig:
		m.screen = ScreenNewConfig
		m.configForm.SetConfig(nil, true)
	}

	return m, nil
}

func (m Model) handleSelectConfig(msg selectConfigMsg) (tea.Model, tea.Cmd) {
	if msg.index < 0 || msg.index >= len(m.configs) {
		return m, nil
	}

	m.currentConfigIdx = msg.index
	m.loading = true
	m.statusMsg = "Connecting..."
	m.statusType = "warning"

	gitlabService := m.gitlabService
	config := m.configs[m.currentConfigIdx]

	return m, func() tea.Msg {
		if err := gitlabService.SetConfig(&config); err != nil {
			return errMsg{err}
		}

		schedules, err := gitlabService.GetSchedules()
		if err != nil {
			return errMsg{err}
		}

		branches, _ := gitlabService.GetBranches()
		branchNames := make([]string, len(branches))
		for i, b := range branches {
			branchNames[i] = b.Name
		}

		return configSelectedMsg{
			schedules:     schedules,
			branches:      branchNames,
			updatedConfig: &config, // Contains ProjectID from API
		}
	}
}

func (m Model) handleSaveSchedule(msg saveScheduleMsg) (tea.Model, tea.Cmd) {
	m.loading = true
	m.statusMsg = "Saving..."
	m.statusType = "warning"

	gitlabService := m.gitlabService

	return m, func() tea.Msg {
		req := &models.ScheduleUpdateRequest{
			Description:  &msg.description,
			Cron:         &msg.cron,
			CronTimezone: &msg.timezone,
			Ref:          &msg.branch,
			Active:       &msg.active,
		}

		if _, err := gitlabService.UpdateSchedule(msg.id, req); err != nil {
			return errMsg{err}
		}

		schedules, _ := gitlabService.GetSchedules()
		return schedulesSavedMsg{schedules: schedules, message: "Schedule saved!"}
	}
}

func (m Model) handleCreateSchedule(msg createScheduleMsg) (tea.Model, tea.Cmd) {
	m.loading = true
	m.statusMsg = "Creating..."
	m.statusType = "warning"

	gitlabService := m.gitlabService

	return m, func() tea.Msg {
		req := &models.ScheduleCreateRequest{
			Description:  msg.description,
			Cron:         msg.cron,
			CronTimezone: msg.timezone,
			Ref:          msg.branch,
			Active:       msg.active,
			Variables:    msg.variables,
		}

		if _, err := gitlabService.CreateSchedule(req); err != nil {
			return errMsg{err}
		}

		schedules, _ := gitlabService.GetSchedules()
		return schedulesSavedMsg{schedules: schedules, message: "Schedule created!"}
	}
}

func (m Model) handleDeleteSchedule(msg deleteScheduleMsg) (tea.Model, tea.Cmd) {
	m.loading = true
	m.statusMsg = "Deleting..."
	m.statusType = "warning"

	gitlabService := m.gitlabService

	return m, func() tea.Msg {
		if err := gitlabService.DeleteSchedule(msg.id); err != nil {
			return errMsg{err}
		}

		schedules, _ := gitlabService.GetSchedules()
		return schedulesSavedMsg{schedules: schedules, message: "Schedule deleted!"}
	}
}

func (m Model) handleToggleSchedule(msg toggleScheduleMsg) (tea.Model, tea.Cmd) {
	gitlabService := m.gitlabService

	return m, func() tea.Msg {
		req := &models.ScheduleUpdateRequest{
			Active: &msg.active,
		}

		if _, err := gitlabService.UpdateSchedule(msg.id, req); err != nil {
			return errMsg{err}
		}

		schedules, _ := gitlabService.GetSchedules()
		return schedulesLoadedMsg{schedules: schedules}
	}
}

func (m Model) handleRunSchedule(msg runScheduleMsg) (tea.Model, tea.Cmd) {
	m.loading = true
	m.statusMsg = "Running pipeline..."
	m.statusType = "warning"

	gitlabService := m.gitlabService

	return m, func() tea.Msg {
		if err := gitlabService.RunSchedule(msg.id); err != nil {
			return errMsg{err}
		}

		schedules, _ := gitlabService.GetSchedules()
		return schedulesSavedMsg{schedules: schedules, message: "Pipeline started!"}
	}
}

func (m Model) handleRefreshSchedules() (tea.Model, tea.Cmd) {
	m.loading = true
	m.statusMsg = "Refreshing..."
	m.statusType = "warning"

	gitlabService := m.gitlabService

	return m, func() tea.Msg {
		schedules, err := gitlabService.GetSchedules()
		if err != nil {
			return errMsg{err}
		}
		return schedulesSavedMsg{schedules: schedules, message: "Schedules refreshed!"}
	}
}

func (m Model) handleSaveConfig(msg saveConfigMsg) (tea.Model, tea.Cmd) {
	m.loading = true
	m.statusMsg = "Validating..."
	m.statusType = "warning"

	gitlabService := m.gitlabService
	configService := m.configService

	return m, func() tea.Msg {
		config := models.Config{
			Name:       msg.name,
			ProjectURL: msg.url,
			Token:      msg.token,
		}

		if err := gitlabService.ValidateConfig(&config); err != nil {
			return errMsg{err}
		}

		var err error
		if msg.isNew {
			err = configService.AddConfig(config)
		} else {
			err = configService.UpdateConfig(msg.index, config)
		}

		if err != nil {
			return errMsg{err}
		}

		configFile, _ := configService.Load()
		return configSavedMsg{configs: configFile.Configs, message: "Configuration saved!"}
	}
}

func (m Model) handleDeleteConfig(msg deleteConfigMsg) (tea.Model, tea.Cmd) {
	configService := m.configService

	return m, func() tea.Msg {
		if err := configService.DeleteConfig(msg.index); err != nil {
			return errMsg{err}
		}

		configFile, _ := configService.Load()
		return configSavedMsg{configs: configFile.Configs, message: "Configuration deleted!"}
	}
}

// Helper functions
func padToWidth(s string, width int) string {
	visibleWidth := lipgloss.Width(s)
	if visibleWidth >= width {
		return s
	}
	return s + strings.Repeat(" ", width-visibleWidth)
}

func truncateOrPad(s string, width int) string {
	// Remove any trailing newline
	s = strings.TrimRight(s, "\n")

	visibleWidth := lipgloss.Width(s)
	if visibleWidth > width {
		// Truncate - this is tricky with ANSI codes
		// For simplicity, truncate the raw string
		if len(s) > width {
			return s[:width]
		}
		return s
	}
	return s + strings.Repeat(" ", width-visibleWidth)
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func truncateStr(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func padRight(s string, width int) string {
	visibleWidth := lipgloss.Width(s)
	if visibleWidth >= width {
		return s
	}
	return s + strings.Repeat(" ", width-visibleWidth)
}

func padLeft(s string, width int) string {
	visibleWidth := lipgloss.Width(s)
	if visibleWidth >= width {
		return s
	}
	return strings.Repeat(" ", width-visibleWidth) + s
}

func formatShortTime(seconds int64) string {
	if seconds < 0 {
		return "past"
	}
	if seconds < 60 {
		return "<1m"
	}
	if seconds < 3600 {
		return fmt.Sprintf("%dm", seconds/60)
	}
	if seconds < 86400 {
		return fmt.Sprintf("%dh", seconds/3600)
	}
	return fmt.Sprintf("%dd", seconds/86400)
}
