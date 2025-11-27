package types

import "time"

// UserProfile represents a developer's profile used for matching
type UserProfile struct {
	Name            string   `json:"name" yaml:"name"`
	Skills          []string `json:"skills" yaml:"skills"`
	Interests       []string `json:"interests" yaml:"interests"`
	ExperienceYears int      `json:"experience_years" yaml:"experience_years"`
}

// IssueMatch represents a GitHub issue that matches the user's profile
type IssueMatch struct {
	Repo        string   `json:"repo"`
	IssueNumber int      `json:"issue_number"`
	Title       string   `json:"title"`
	URL         string   `json:"url"`
	MatchReason string   `json:"match_reason"`
	Effort      string   `json:"estimated_effort"`
	Labels      []string `json:"labels"`
	CreatedAt   string   `json:"created_at"`
	FoundAt     string   `json:"found_at"`
}

// State represents the persistent state of the issue finder
type State struct {
	LastRun         string          `json:"last_run"`
	ProcessedIssues map[string]bool `json:"processed_issues"`
	AllMatches      []IssueMatch    `json:"all_matches"`
}

// GitHubIssue represents a GitHub issue from the API
type GitHubIssue struct {
	Number    int       `json:"number"`
	Title     string    `json:"title"`
	URL       string    `json:"html_url"`
	Labels    []Label   `json:"labels"`
	CreatedAt time.Time `json:"created_at"`
	Body      string    `json:"body"`
	RepoURL   string    `json:"repository_url"`
}

// Label represents a GitHub label
type Label struct {
	Name string `json:"name"`
}

// Config represents the application configuration
type Config struct {
	Profile     UserProfile       `yaml:"profile"`
	Preferences PreferencesConfig `yaml:"preferences"`
	API         APIConfig         `yaml:"api"`
}

// PreferencesConfig represents user preferences
type PreferencesConfig struct {
	OutputPath         string `yaml:"output_path"`
	StatePath          string `yaml:"state_path"`
	NotifyOnCompletion bool   `yaml:"notify_on_completion"`
	MaxMatches         int    `yaml:"max_matches"`
}

// APIConfig represents API configuration
type APIConfig struct {
	AnthropicKeyEnv string `yaml:"anthropic_key_env"`
	GitHubTokenEnv  string `yaml:"github_token_env"`
}
