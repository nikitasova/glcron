package services

import (
	"encoding/json"
	"fmt"
	"glcron/internal/models"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// GitLabServiceInterface defines the interface for GitLab API operations
type GitLabServiceInterface interface {
	SetConfig(config *models.Config) error
	GetSchedules() ([]models.Schedule, error)
	GetSchedule(id int) (*models.Schedule, error)
	CreateSchedule(req *models.ScheduleCreateRequest) (*models.Schedule, error)
	UpdateSchedule(id int, req *models.ScheduleUpdateRequest) (*models.Schedule, error)
	DeleteSchedule(id int) error
	RunSchedule(id int) error
	TakeOwnership(id int) (*models.Schedule, error)
	GetCurrentUser() (*models.User, error)
	GetBranches() ([]models.Branch, error)
	CreateVariable(scheduleID int, variable *models.Variable) error
	UpdateVariable(scheduleID int, variable *models.Variable) error
	DeleteVariable(scheduleID int, key string) error
	ValidateConfig(config *models.Config) error
}

// GitLabService handles GitLab API interactions
type GitLabService struct {
	baseURL   string
	projectID int
	token     string
	client    *http.Client
}

// NewGitLabService creates a new GitLabService
func NewGitLabService() GitLabServiceInterface {
	return &GitLabService{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SetConfig sets the GitLab configuration
func (g *GitLabService) SetConfig(config *models.Config) error {
	if config == nil {
		return fmt.Errorf("config is nil")
	}

	// Parse project URL to extract base URL and project path
	baseURL, projectPath, err := parseProjectURL(config.ProjectURL)
	if err != nil {
		return err
	}

	g.baseURL = baseURL
	g.token = config.Token

	// Get project ID from API
	projectID, err := g.getProjectID(projectPath)
	if err != nil {
		return err
	}

	g.projectID = projectID
	config.ProjectID = projectID
	config.BaseURL = baseURL

	return nil
}

// ValidateConfig validates a configuration without setting it
func (g *GitLabService) ValidateConfig(config *models.Config) error {
	if config == nil {
		return fmt.Errorf("config is nil")
	}

	if config.Name == "" {
		return fmt.Errorf("name is required")
	}

	if config.ProjectURL == "" {
		return fmt.Errorf("project URL is required")
	}

	if config.Token == "" {
		return fmt.Errorf("token is required")
	}

	// Validate URL format
	_, _, err := parseProjectURL(config.ProjectURL)
	if err != nil {
		return err
	}

	// Try to get project ID to validate credentials
	tempService := &GitLabService{
		client: g.client,
		token:  config.Token,
	}
	baseURL, projectPath, _ := parseProjectURL(config.ProjectURL)
	tempService.baseURL = baseURL

	_, err = tempService.getProjectID(projectPath)
	if err != nil {
		return fmt.Errorf("failed to validate config: %v", err)
	}

	return nil
}

// parseProjectURL extracts base URL and project path from GitLab URL
func parseProjectURL(projectURL string) (baseURL, projectPath string, err error) {
	// Remove trailing slash
	projectURL = strings.TrimSuffix(projectURL, "/")

	// Parse URL
	parsed, err := url.Parse(projectURL)
	if err != nil {
		return "", "", fmt.Errorf("invalid URL: %v", err)
	}

	if parsed.Host == "" {
		return "", "", fmt.Errorf("invalid URL: missing host")
	}

	// Base URL is scheme + host
	baseURL = fmt.Sprintf("%s://%s", parsed.Scheme, parsed.Host)

	// Project path is everything after the host
	projectPath = strings.TrimPrefix(parsed.Path, "/")

	if projectPath == "" {
		return "", "", fmt.Errorf("invalid URL: missing project path")
	}

	return baseURL, projectPath, nil
}

// getProjectID gets the project ID from the project path
func (g *GitLabService) getProjectID(projectPath string) (int, error) {
	// URL encode the project path
	encodedPath := url.PathEscape(projectPath)

	resp, err := g.doRequest("GET", fmt.Sprintf("/api/v4/projects/%s", encodedPath), nil)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("failed to get project: %s - %s", resp.Status, string(body))
	}

	var project struct {
		ID int `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&project); err != nil {
		return 0, fmt.Errorf("failed to decode project: %v", err)
	}

	return project.ID, nil
}

// GetSchedules fetches all pipeline schedules
func (g *GitLabService) GetSchedules() ([]models.Schedule, error) {
	resp, err := g.doRequest("GET", fmt.Sprintf("/api/v4/projects/%d/pipeline_schedules", g.projectID), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get schedules: %s - %s", resp.Status, string(body))
	}

	var schedules []models.Schedule
	if err := json.NewDecoder(resp.Body).Decode(&schedules); err != nil {
		return nil, fmt.Errorf("failed to decode schedules: %v", err)
	}

	// Fetch details for each schedule to get variables and last pipeline
	for i := range schedules {
		details, err := g.GetSchedule(schedules[i].ID)
		if err == nil {
			schedules[i].Variables = details.Variables
			schedules[i].LastPipeline = details.LastPipeline
		}
	}

	return schedules, nil
}

// GetSchedule fetches a single schedule with full details
func (g *GitLabService) GetSchedule(id int) (*models.Schedule, error) {
	resp, err := g.doRequest("GET", fmt.Sprintf("/api/v4/projects/%d/pipeline_schedules/%d", g.projectID, id), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get schedule: %s - %s", resp.Status, string(body))
	}

	var schedule models.Schedule
	if err := json.NewDecoder(resp.Body).Decode(&schedule); err != nil {
		return nil, fmt.Errorf("failed to decode schedule: %v", err)
	}

	return &schedule, nil
}

// CreateSchedule creates a new pipeline schedule
func (g *GitLabService) CreateSchedule(req *models.ScheduleCreateRequest) (*models.Schedule, error) {
	// Build form data
	data := url.Values{}
	data.Set("description", req.Description)
	data.Set("ref", req.Ref)
	data.Set("cron", req.Cron)
	if req.CronTimezone != "" {
		data.Set("cron_timezone", req.CronTimezone)
	}
	data.Set("active", fmt.Sprintf("%t", req.Active))

	resp, err := g.doRequest("POST", fmt.Sprintf("/api/v4/projects/%d/pipeline_schedules", g.projectID), strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to create schedule: %s - %s", resp.Status, string(body))
	}

	var schedule models.Schedule
	if err := json.NewDecoder(resp.Body).Decode(&schedule); err != nil {
		return nil, fmt.Errorf("failed to decode schedule: %v", err)
	}

	// Add variables if any
	for _, v := range req.Variables {
		if err := g.CreateVariable(schedule.ID, &v); err != nil {
			// Log error but continue
			fmt.Printf("Warning: failed to create variable %s: %v\n", v.Key, err)
		}
	}

	return &schedule, nil
}

// UpdateSchedule updates an existing pipeline schedule
func (g *GitLabService) UpdateSchedule(id int, req *models.ScheduleUpdateRequest) (*models.Schedule, error) {
	// Build form data
	data := url.Values{}
	if req.Description != nil {
		data.Set("description", *req.Description)
	}
	if req.Ref != nil {
		data.Set("ref", *req.Ref)
	}
	if req.Cron != nil {
		data.Set("cron", *req.Cron)
	}
	if req.CronTimezone != nil {
		data.Set("cron_timezone", *req.CronTimezone)
	}
	if req.Active != nil {
		data.Set("active", fmt.Sprintf("%t", *req.Active))
	}

	resp, err := g.doRequest("PUT", fmt.Sprintf("/api/v4/projects/%d/pipeline_schedules/%d", g.projectID, id), strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to update schedule: %s - %s", resp.Status, string(body))
	}

	var schedule models.Schedule
	if err := json.NewDecoder(resp.Body).Decode(&schedule); err != nil {
		return nil, fmt.Errorf("failed to decode schedule: %v", err)
	}

	return &schedule, nil
}

// DeleteSchedule deletes a pipeline schedule
func (g *GitLabService) DeleteSchedule(id int) error {
	resp, err := g.doRequest("DELETE", fmt.Sprintf("/api/v4/projects/%d/pipeline_schedules/%d", g.projectID, id), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete schedule: %s - %s", resp.Status, string(body))
	}

	return nil
}

// RunSchedule triggers a pipeline schedule to run immediately
func (g *GitLabService) RunSchedule(id int) error {
	resp, err := g.doRequest("POST", fmt.Sprintf("/api/v4/projects/%d/pipeline_schedules/%d/play", g.projectID, id), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to run schedule: %s - %s", resp.Status, string(body))
	}

	return nil
}

// TakeOwnership takes ownership of a pipeline schedule
func (g *GitLabService) TakeOwnership(id int) (*models.Schedule, error) {
	resp, err := g.doRequest("POST", fmt.Sprintf("/api/v4/projects/%d/pipeline_schedules/%d/take_ownership", g.projectID, id), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to take ownership: %s - %s", resp.Status, string(body))
	}

	var schedule models.Schedule
	if err := json.NewDecoder(resp.Body).Decode(&schedule); err != nil {
		return nil, fmt.Errorf("failed to decode schedule: %v", err)
	}

	return &schedule, nil
}

// GetCurrentUser fetches the current authenticated user
func (g *GitLabService) GetCurrentUser() (*models.User, error) {
	resp, err := g.doRequest("GET", "/api/v4/user", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get current user: %s - %s", resp.Status, string(body))
	}

	var user models.User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("failed to decode user: %v", err)
	}

	return &user, nil
}

// GetBranches fetches all branches
func (g *GitLabService) GetBranches() ([]models.Branch, error) {
	var allBranches []models.Branch
	page := 1
	perPage := 100

	for {
		resp, err := g.doRequest("GET", fmt.Sprintf("/api/v4/projects/%d/repository/branches?page=%d&per_page=%d", g.projectID, page, perPage), nil)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return nil, fmt.Errorf("failed to get branches: %s - %s", resp.Status, string(body))
		}

		var branches []models.Branch
		if err := json.NewDecoder(resp.Body).Decode(&branches); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("failed to decode branches: %v", err)
		}
		resp.Body.Close()

		allBranches = append(allBranches, branches...)

		// Check if we got all branches
		if len(branches) < perPage {
			break
		}
		page++

		// Safety limit
		if page > 10 {
			break
		}
	}

	return allBranches, nil
}

// CreateVariable creates a new variable for a schedule
func (g *GitLabService) CreateVariable(scheduleID int, variable *models.Variable) error {
	data := url.Values{}
	data.Set("key", variable.Key)
	data.Set("value", variable.Value)
	if variable.VariableType != "" {
		data.Set("variable_type", variable.VariableType)
	}

	resp, err := g.doRequest("POST", fmt.Sprintf("/api/v4/projects/%d/pipeline_schedules/%d/variables", g.projectID, scheduleID), strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create variable: %s - %s", resp.Status, string(body))
	}

	return nil
}

// UpdateVariable updates an existing variable
func (g *GitLabService) UpdateVariable(scheduleID int, variable *models.Variable) error {
	data := url.Values{}
	data.Set("value", variable.Value)
	if variable.VariableType != "" {
		data.Set("variable_type", variable.VariableType)
	}

	resp, err := g.doRequest("PUT", fmt.Sprintf("/api/v4/projects/%d/pipeline_schedules/%d/variables/%s", g.projectID, scheduleID, url.PathEscape(variable.Key)), strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update variable: %s - %s", resp.Status, string(body))
	}

	return nil
}

// DeleteVariable deletes a variable
func (g *GitLabService) DeleteVariable(scheduleID int, key string) error {
	resp, err := g.doRequest("DELETE", fmt.Sprintf("/api/v4/projects/%d/pipeline_schedules/%d/variables/%s", g.projectID, scheduleID, url.PathEscape(key)), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete variable: %s - %s", resp.Status, string(body))
	}

	return nil
}

// doRequest performs an HTTP request
func (g *GitLabService) doRequest(method, path string, body io.Reader) (*http.Response, error) {
	reqURL := g.baseURL + path

	req, err := http.NewRequest(method, reqURL, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("PRIVATE-TOKEN", g.token)
	if body != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	return g.client.Do(req)
}

// Common timezones for the dropdown
var CommonTimezones = []string{
	"UTC",
	"America/New_York",
	"America/Chicago",
	"America/Denver",
	"America/Los_Angeles",
	"America/Toronto",
	"America/Vancouver",
	"America/Sao_Paulo",
	"Europe/London",
	"Europe/Paris",
	"Europe/Berlin",
	"Europe/Amsterdam",
	"Europe/Madrid",
	"Europe/Rome",
	"Europe/Moscow",
	"Europe/Kiev",
	"Asia/Dubai",
	"Asia/Kolkata",
	"Asia/Singapore",
	"Asia/Hong_Kong",
	"Asia/Shanghai",
	"Asia/Tokyo",
	"Asia/Seoul",
	"Australia/Sydney",
	"Australia/Melbourne",
	"Pacific/Auckland",
}

// ValidateCronExpression validates a cron expression
func ValidateCronExpression(cron string) error {
	// Basic validation: should have 5 fields
	fields := strings.Fields(cron)
	if len(fields) != 5 {
		return fmt.Errorf("cron expression must have exactly 5 fields")
	}

	// Validate each field
	patterns := []struct {
		name    string
		pattern string
	}{
		{"minute", `^(\*|([0-9]|[1-5][0-9])(-([0-9]|[1-5][0-9]))?(,([0-9]|[1-5][0-9])(-([0-9]|[1-5][0-9]))?)*(/[0-9]+)?)$`},
		{"hour", `^(\*|([0-9]|1[0-9]|2[0-3])(-([0-9]|1[0-9]|2[0-3]))?(,([0-9]|1[0-9]|2[0-3])(-([0-9]|1[0-9]|2[0-3]))?)*(/[0-9]+)?)$`},
		{"day of month", `^(\*|([1-9]|[12][0-9]|3[01])(-([1-9]|[12][0-9]|3[01]))?(,([1-9]|[12][0-9]|3[01])(-([1-9]|[12][0-9]|3[01]))?)*(/[0-9]+)?)$`},
		{"month", `^(\*|([1-9]|1[0-2])(-([1-9]|1[0-2]))?(,([1-9]|1[0-2])(-([1-9]|1[0-2]))?)*(/[0-9]+)?)$`},
		{"day of week", `^(\*|[0-6](-[0-6])?(,[0-6](-[0-6])?)*(/[0-9]+)?)$`},
	}

	for i, p := range patterns {
		matched, err := regexp.MatchString(p.pattern, fields[i])
		if err != nil || !matched {
			return fmt.Errorf("invalid %s field: %s", p.name, fields[i])
		}
	}

	return nil
}
