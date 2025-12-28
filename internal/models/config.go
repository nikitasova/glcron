package models

// Config represents a GitLab project configuration
type Config struct {
	Name       string `json:"name"`        // Custom display name
	ProjectURL string `json:"project_url"` // GitLab project URL (https://gitlab.company.com/group/project)
	Token      string `json:"token"`       // GitLab personal access token
	ProjectID  int    `json:"project_id"`  // GitLab project ID (extracted from API)
	BaseURL    string `json:"base_url"`    // Base GitLab API URL
}

// ConfigFile represents the configuration file structure
type ConfigFile struct {
	Configs []Config `json:"configs"`
}
