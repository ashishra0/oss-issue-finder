package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/ashishra0/issue-finder/internal/ai"
	"github.com/ashishra0/issue-finder/internal/github"
	"github.com/ashishra0/issue-finder/internal/output"
	"github.com/ashishra0/issue-finder/internal/state"
	"github.com/ashishra0/issue-finder/pkg/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	skills     []string
	interests  []string
	experience int
	outputPath string
	statePath  string
	noNotify   bool
)

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search for OSS contribution opportunities",
	Long: `Search GitHub for open source contribution opportunities that match
your skills, interests, and experience level.

The command will:
1. Search GitHub for relevant issues
2. Filter out already processed issues
3. Send new issues to Anthropic AI for evaluation
4. Write results to a markdown file

You can provide your profile via command-line flags or a config file.`,
	Example: `  # Search with inline preferences
  issue-finder search --skills "Go,Python,PostgreSQL" --interests "Backend,Databases" --experience 5

  # Use config file
  issue-finder search --config ~/.my-profile.yaml

  # Specify custom output location
  issue-finder search --output ~/my-contributions.md`,
	RunE: runSearch,
}

func init() {
	rootCmd.AddCommand(searchCmd)

	searchCmd.Flags().StringSliceVar(&skills, "skills", []string{}, "Your technical skills (comma-separated)")
	searchCmd.Flags().StringSliceVar(&interests, "interests", []string{}, "Your areas of interest (comma-separated)")
	searchCmd.Flags().IntVar(&experience, "experience", 0, "Years of experience")
	searchCmd.Flags().StringVar(&outputPath, "output", "", "Output file path (default: ~/contributions.md)")
	searchCmd.Flags().StringVar(&statePath, "state", "", "State file path (default: ~/.issue-finder-state.json)")
	searchCmd.Flags().BoolVar(&noNotify, "no-notify", false, "Disable desktop notifications")

	viper.BindPFlag("profile.skills", searchCmd.Flags().Lookup("skills"))
	viper.BindPFlag("profile.interests", searchCmd.Flags().Lookup("interests"))
	viper.BindPFlag("profile.experience_years", searchCmd.Flags().Lookup("experience"))
	viper.BindPFlag("preferences.output_path", searchCmd.Flags().Lookup("output"))
	viper.BindPFlag("preferences.state_path", searchCmd.Flags().Lookup("state"))
}

func runSearch(cmd *cobra.Command, args []string) error {
	profile := loadProfile()

	if err := validateProfile(profile); err != nil {
		return err
	}

	home, _ := os.UserHomeDir()

	outputFile := viper.GetString("preferences.output_path")
	if outputFile == "" {
		outputFile = filepath.Join(home, "contributions.md")
	}
	outputFile = expandPath(outputFile)

	stateFile := viper.GetString("preferences.state_path")
	if stateFile == "" {
		stateFile = filepath.Join(home, ".issue-finder-state.json")
	}
	stateFile = expandPath(stateFile)

	githubToken := os.Getenv("GITHUB_TOKEN")
	if githubToken == "" {
		return fmt.Errorf("GITHUB_TOKEN environment variable not set")
	}

	anthropicKey := os.Getenv("ANTHROPIC_API_KEY")
	if anthropicKey == "" {
		return fmt.Errorf("ANTHROPIC_API_KEY environment variable not set")
	}

	progress := output.NewProgressFormatter(quiet)
	progress.PrintHeader(profile.Name, profile.Skills, profile.Interests, profile.ExperienceYears)

	stateMgr := state.NewManager(stateFile)
	currentState := stateMgr.Load()

	progress.Step(1, "Searching GitHub for relevant issues...")

	ghClient := github.NewClient(githubToken)
	recentIssues := ghClient.FetchRelevantIssues(profile)

	progress.Detail(fmt.Sprintf("Found %d issues across multiple queries", len(recentIssues)))
	progress.EmptyLine()

	progress.Step(2, "Filtering processed issues...")

	newIssues := stateMgr.FilterNewIssues(&currentState, recentIssues)

	progress.Detail(fmt.Sprintf("%d already evaluated, %d new issues to process",
		len(recentIssues)-len(newIssues), len(newIssues)))
	progress.EmptyLine()

	var newMatchesCount int

	if len(newIssues) > 0 {
		progress.Step(3, "Evaluating with AI...")
		progress.Detail(fmt.Sprintf("Sending %d issues to Anthropic AI for evaluation...", len(newIssues)))

		evaluator := ai.NewEvaluator(anthropicKey)
		matches := evaluator.EvaluateIssues(profile, newIssues)

		newMatchesCount = len(matches)

		progress.Detail(fmt.Sprintf("Received %d high-quality matches", len(matches)))
		progress.EmptyLine()

		maxMatches := viper.GetInt("preferences.max_matches")
		if maxMatches == 0 {
			maxMatches = 100
		}

		stateMgr.AddMatches(&currentState, matches, maxMatches)

		log.Printf("Claude found %d good matches", len(matches))
	} else {
		progress.Step(3, "Evaluating with AI...")
		progress.Detail("No new issues to evaluate")
		progress.EmptyLine()
	}

	progress.Step(4, "Writing results...")

	err := output.WriteMarkdownFile(outputFile, currentState)
	if err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	err = stateMgr.Save(currentState)
	if err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	progress.Detail(fmt.Sprintf("Updated %s", outputFile))
	progress.Detail(fmt.Sprintf("Total opportunities: %d (%d new, %d from history)",
		len(currentState.AllMatches),
		newMatchesCount,
		len(currentState.AllMatches)-newMatchesCount))

	progress.Success(fmt.Sprintf("Done! Run 'open %s' to view results.", outputFile))

	// Send notification if enabled (default true unless explicitly disabled)
	notifyEnabled := viper.GetBool("preferences.notify_on_completion")
	if !viper.IsSet("preferences.notify_on_completion") {
		notifyEnabled = true // Default to true if not set in config
	}

	if !noNotify && notifyEnabled {
		var notifMsg string
		if newMatchesCount > 0 {
			notifMsg = fmt.Sprintf("Found %d new opportunities", newMatchesCount)
		} else {
			notifMsg = "No new opportunities found"
		}

		if !quiet {
			progress.Detail("Sending notification...")
		}

		err := output.SendNotification("Issue Finder", notifMsg)
		if err != nil {
			if !quiet {
				progress.Warning(fmt.Sprintf("Failed to send notification: %v", err))
			}
		}
	}

	return nil
}

func loadProfile() types.UserProfile {
	return types.UserProfile{
		Name:            viper.GetString("profile.name"),
		Skills:          viper.GetStringSlice("profile.skills"),
		Interests:       viper.GetStringSlice("profile.interests"),
		ExperienceYears: viper.GetInt("profile.experience_years"),
	}
}

func validateProfile(profile types.UserProfile) error {
	if len(profile.Skills) == 0 {
		return fmt.Errorf("no skills specified\n  Use --skills flag or set profile.skills in config file\n  Example: contribution-finder search --skills \"Go,Python\"")
	}

	if profile.ExperienceYears < 0 || profile.ExperienceYears > 50 {
		return fmt.Errorf("experience years must be between 0 and 50")
	}

	return nil
}

func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	return path
}
