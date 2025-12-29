package tui

import (
	"glcron/internal/models"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Data loading messages
type configsLoadedMsg struct {
	configs []models.Config
}

type schedulesLoadedMsg struct {
	schedules []models.Schedule
}

type branchesLoadedMsg struct {
	branches []string
}

type errMsg struct {
	err error
}

// Combined messages for async operations
type configSelectedMsg struct {
	schedules     []models.Schedule
	branches      []string
	updatedConfig *models.Config
}

type schedulesSavedMsg struct {
	schedules []models.Schedule
	message   string
}

type configSavedMsg struct {
	configs []models.Config
	message string
}

// Status messages
type statusMsg struct {
	text    string
	msgType string // "success", "error", "warning"
}

type clearStatusMsg struct{}

// Navigation messages
type navigateMsg struct {
	screen   Screen
	schedule *models.Schedule
	config   *models.Config
}

// Config actions
type selectConfigMsg struct {
	index int
}

type saveConfigMsg struct {
	index int
	name  string
	url   string
	token string
	isNew bool
}

type deleteConfigMsg struct {
	index int
}

// Schedule actions
type saveScheduleMsg struct {
	id          int
	description string
	cron        string
	timezone    string
	branch      string
	active      bool
	variables   []models.Variable
}

type createScheduleMsg struct {
	description string
	cron        string
	timezone    string
	branch      string
	active      bool
	variables   []models.Variable
}

type deleteScheduleMsg struct {
	id int
}

type toggleScheduleMsg struct {
	id     int
	active bool
}

type runScheduleMsg struct {
	id int
}

type refreshSchedulesMsg struct{}

// Helper to create navigate command
func Navigate(screen Screen) tea.Cmd {
	return func() tea.Msg {
		return navigateMsg{screen: screen}
	}
}

func NavigateToEdit(schedule *models.Schedule) tea.Cmd {
	return func() tea.Msg {
		return navigateMsg{screen: ScreenEditSchedule, schedule: schedule}
	}
}

func NavigateToEditConfig(config *models.Config) tea.Cmd {
	return func() tea.Msg {
		return navigateMsg{screen: ScreenEditConfig, config: config}
	}
}

// ClearStatusAfter returns a command that clears the status message after the specified duration
func ClearStatusAfter(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return clearStatusMsg{}
	})
}
