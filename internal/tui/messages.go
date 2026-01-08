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
	currentUser   *models.User
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
	screen      Screen
	schedule    *models.Schedule
	config      *models.Config
	configIndex int
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

type saveScheduleWithOwnershipMsg struct {
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

type takeOwnershipMsg struct {
	id int
}

type ownershipTakenMsg struct {
	schedule  *models.Schedule
	schedules []models.Schedule
}

type refreshSchedulesMsg struct{}

// Quick Run messages
type quickRunPipelineMsg struct {
	branch    string
	variables []models.Variable
}

type pipelineCreatedMsg struct {
	pipeline *models.Pipeline
}

type pipelinesLoadedMsg struct {
	pipelines []models.PipelineWithJobs
}

type refreshPipelinesMsg struct{}

type pipelineTickMsg struct{}

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

func NavigateToEditConfig(config *models.Config, index int) tea.Cmd {
	return func() tea.Msg {
		return navigateMsg{screen: ScreenEditConfig, config: config, configIndex: index}
	}
}

func NavigateToYonk(schedule *models.Schedule) tea.Cmd {
	return func() tea.Msg {
		// Create a copy with "[Copy]" prefix
		copied := &models.Schedule{
			Description:  "[Copy] " + schedule.Description,
			Ref:          schedule.Ref,
			Cron:         schedule.Cron,
			CronTimezone: schedule.CronTimezone,
			Active:       schedule.Active,
			Variables:    make([]models.Variable, len(schedule.Variables)),
		}
		copy(copied.Variables, schedule.Variables)
		return navigateMsg{screen: ScreenNewSchedule, schedule: copied}
	}
}

// ClearStatusAfter returns a command that clears the status message after the specified duration
func ClearStatusAfter(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return clearStatusMsg{}
	})
}
