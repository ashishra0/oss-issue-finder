package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ashishra0/issue-finder/internal/state"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var stateCmd = &cobra.Command{
	Use:   "state",
	Short: "Manage state",
	Long:  `Manage the issue finder state (processed issues and matches).`,
}

var stateClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear processed issues",
	Long: `Clear all processed issues from the state file.

This will cause the next search to re-evaluate all issues,
even ones that have been processed before.

Useful when you've changed your profile significantly or want
to get fresh matches for previously seen issues.`,
	Example: `  # Clear state
  issue-finder state clear

  # Clear state from custom location
  issue-finder state clear --state ~/.my-state.json`,
	RunE: runStateClear,
}

var stateShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show state statistics",
	Long:  `Display statistics about the current state (processed issues and matches).`,
	Example: `  # Show state stats
  issue-finder state show

  # Show stats from custom location
  issue-finder state show --state ~/.my-state.json`,
	RunE: runStateShow,
}

func init() {
	rootCmd.AddCommand(stateCmd)
	stateCmd.AddCommand(stateClearCmd)
	stateCmd.AddCommand(stateShowCmd)

	stateClearCmd.Flags().StringVar(&statePath, "state", "", "State file path (default: ~/.issue-finder-state.json)")
	stateShowCmd.Flags().StringVar(&statePath, "state", "", "State file path (default: ~/.issue-finder-state.json)")
}

func runStateClear(cmd *cobra.Command, args []string) error {
	stateFile := getStatePath()

	if _, err := os.Stat(stateFile); os.IsNotExist(err) {
		fmt.Printf("No state file found at: %s\n", stateFile)
		fmt.Println("Nothing to clear.")
		return nil
	}

	fmt.Printf("This will clear all processed issues from: %s\n", stateFile)
	fmt.Print("Are you sure? (y/N): ")

	var response string
	fmt.Scanln(&response)

	if response != "y" && response != "Y" {
		fmt.Println("Cancelled.")
		return nil
	}

	stateMgr := state.NewManager(stateFile)
	err := stateMgr.Clear()
	if err != nil {
		return fmt.Errorf("error clearing state: %w", err)
	}

	fmt.Println("State cleared successfully.")
	fmt.Println("The next search will re-evaluate all issues.")

	return nil
}

func runStateShow(cmd *cobra.Command, args []string) error {
	stateFile := getStatePath()

	if _, err := os.Stat(stateFile); os.IsNotExist(err) {
		fmt.Printf("No state file found at: %s\n", stateFile)
		fmt.Println("Run a search first to create the state file.")
		return nil
	}

	stateMgr := state.NewManager(stateFile)
	processedCount, matchesCount, err := stateMgr.GetStats()
	if err != nil {
		return fmt.Errorf("error reading state: %w", err)
	}

	currentState := stateMgr.Load()

	fmt.Println("State Statistics:")
	fmt.Println("=================")
	fmt.Println()
	fmt.Printf("State file: %s\n", stateFile)
	fmt.Printf("Last run: %s\n", currentState.LastRun)
	fmt.Println()
	fmt.Printf("Processed issues: %d\n", processedCount)
	fmt.Printf("Saved matches: %d\n", matchesCount)
	fmt.Println()

	if matchesCount > 0 {
		fmt.Println("Recent matches:")
		limit := 5
		if matchesCount < limit {
			limit = matchesCount
		}

		for i := 0; i < limit; i++ {
			match := currentState.AllMatches[i]
			fmt.Printf("  %d. [%s] %s\n", i+1, match.Repo, match.Title)
			fmt.Printf("     %s\n", match.URL)
		}

		if matchesCount > limit {
			fmt.Printf("\n  ... and %d more (view in contributions.md)\n", matchesCount-limit)
		}
	}

	return nil
}

func getStatePath() string {
	if statePath != "" {
		return expandPath(statePath)
	}

	stateFile := viper.GetString("preferences.state_path")
	if stateFile != "" {
		return expandPath(stateFile)
	}

	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".issue-finder-state.json")
}
