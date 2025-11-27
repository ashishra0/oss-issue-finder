package state

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ashishra0/issue-finder/pkg/types"
)

type Manager struct {
	statePath string
}

func NewManager(statePath string) *Manager {
	return &Manager{
		statePath: statePath,
	}
}

// Load reads the state from disk
func (m *Manager) Load() types.State {
	data, err := os.ReadFile(m.statePath)

	if os.IsNotExist(err) {
		return types.State{
			ProcessedIssues: make(map[string]bool),
			AllMatches:      []types.IssueMatch{},
		}
	}

	if err != nil {
		log.Printf("Error reading state: %v", err)
		return types.State{
			ProcessedIssues: make(map[string]bool),
			AllMatches:      []types.IssueMatch{},
		}
	}

	var state types.State
	err = json.Unmarshal(data, &state)
	if err != nil {
		log.Printf("Error parsing state: %v", err)
		return types.State{
			ProcessedIssues: make(map[string]bool),
			AllMatches:      []types.IssueMatch{},
		}
	}

	return state
}

// Save writes the state to disk
func (m *Manager) Save(state types.State) error {
	state.LastRun = time.Now().Format(time.RFC3339)

	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("error marshaling state: %w", err)
	}

	err = os.WriteFile(m.statePath, data, 0644)
	if err != nil {
		return fmt.Errorf("error saving state: %w", err)
	}

	return nil
}

// FilterNewIssues filters out already processed issues and marks new ones
func (m *Manager) FilterNewIssues(state *types.State, issues []map[string]any) []map[string]any {
	newIssues := []map[string]any{}

	for _, issue := range issues {
		issueKey := fmt.Sprintf("%s/%d", issue["repo"], issue["number"])

		if !state.ProcessedIssues[issueKey] {
			newIssues = append(newIssues, issue)
			state.ProcessedIssues[issueKey] = true
		}
	}

	return newIssues
}

// AddMatches adds new matches to state and maintains history limit
func (m *Manager) AddMatches(state *types.State, matches []types.IssueMatch, maxMatches int) {
	state.AllMatches = append(matches, state.AllMatches...)

	if len(state.AllMatches) > maxMatches {
		state.AllMatches = state.AllMatches[:maxMatches]
	}
}

// Clear removes all processed issues from state
func (m *Manager) Clear() error {
	state := types.State{
		ProcessedIssues: make(map[string]bool),
		AllMatches:      []types.IssueMatch{},
	}

	return m.Save(state)
}

// GetStats returns statistics about the current state
func (m *Manager) GetStats() (int, int, error) {
	state := m.Load()
	return len(state.ProcessedIssues), len(state.AllMatches), nil
}
