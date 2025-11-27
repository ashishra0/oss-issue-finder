package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
	Long:  `Manage your issue finder configuration file.`,
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a configuration file",
	Long: `Create a default configuration file at ~/.issue-finder.yaml

This will create a template configuration file with example values
that you can edit to match your profile.`,
	Example: `  # Create config in default location
  issue-finder config init

  # Create config in custom location
  issue-finder config init --config ~/my-config.yaml`,
	RunE: runConfigInit,
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Long:  `Display the current configuration values being used.`,
	Example: `  # Show current config
  issue-finder config show

  # Show config from custom location
  issue-finder config show --config ~/my-config.yaml`,
	RunE: runConfigShow,
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configShowCmd)
}

func runConfigInit(cmd *cobra.Command, args []string) error {
	configPath := cfgFile
	if configPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("error finding home directory: %w", err)
		}
		configPath = filepath.Join(home, ".issue-finder.yaml")
	}

	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("config file already exists at %s\nUse a different path or delete the existing file first", configPath)
	}

	defaultConfig := `# Issue Finder Configuration
# Edit this file to match your profile

profile:
  name: "Your Name"
  skills:
    - Go
    - Python
    - PostgreSQL
  interests:
    - Backend development
    - Databases
    - Web frameworks
  experience_years: 5

preferences:
  # Where to write results
  output_path: "~/issues.md"

  # Where to store state
  state_path: "~/.issue-finder-state.json"

  # Enable/disable notifications
  notify_on_completion: true

  # Maximum number of opportunities to keep
  max_matches: 100

api:
  # Environment variable names for API keys
  anthropic_key_env: "ANTHROPIC_API_KEY"
  github_token_env: "GITHUB_TOKEN"
`

	err := os.WriteFile(configPath, []byte(defaultConfig), 0644)
	if err != nil {
		return fmt.Errorf("error writing config file: %w", err)
	}

	fmt.Printf("Created configuration file at: %s\n", configPath)
	fmt.Println("\nEdit this file to customize your profile, then run:")
	fmt.Println("  issue-finder search")

	return nil
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	fmt.Println("Current Configuration:")
	fmt.Println("=====================")
	fmt.Println()

	if viper.ConfigFileUsed() != "" {
		fmt.Printf("Config file: %s\n\n", viper.ConfigFileUsed())
	} else {
		fmt.Println("No config file loaded (using defaults and flags)")
	}

	fmt.Println("Profile:")
	fmt.Printf("  Name: %s\n", viper.GetString("profile.name"))

	skills := viper.GetStringSlice("profile.skills")
	if len(skills) > 0 {
		fmt.Printf("  Skills: %v\n", skills)
	} else {
		fmt.Println("  Skills: (none set)")
	}

	interests := viper.GetStringSlice("profile.interests")
	if len(interests) > 0 {
		fmt.Printf("  Interests: %v\n", interests)
	} else {
		fmt.Println("  Interests: (none set)")
	}

	fmt.Printf("  Experience: %d years\n", viper.GetInt("profile.experience_years"))
	fmt.Println()

	fmt.Println("Preferences:")
	outputPath := viper.GetString("preferences.output_path")
	if outputPath == "" {
		home, _ := os.UserHomeDir()
		outputPath = filepath.Join(home, "issues.md")
	}
	fmt.Printf("  Output path: %s\n", outputPath)

	statePath := viper.GetString("preferences.state_path")
	if statePath == "" {
		home, _ := os.UserHomeDir()
		statePath = filepath.Join(home, ".issue-finder-state.json")
	}
	fmt.Printf("  State path: %s\n", statePath)

	maxMatches := viper.GetInt("preferences.max_matches")
	if maxMatches == 0 {
		maxMatches = 100
	}
	fmt.Printf("  Max matches: %d\n", maxMatches)
	fmt.Printf("  Notify on completion: %v\n", viper.GetBool("preferences.notify_on_completion"))
	fmt.Println()

	fmt.Println("API:")
	fmt.Printf("  Anthropic key env: %s\n", getEnvVarName("api.anthropic_key_env", "ANTHROPIC_API_KEY"))
	fmt.Printf("  GitHub token env: %s\n", getEnvVarName("api.github_token_env", "GITHUB_TOKEN"))

	anthropicKey := os.Getenv(getEnvVarName("api.anthropic_key_env", "ANTHROPIC_API_KEY"))
	githubToken := os.Getenv(getEnvVarName("api.github_token_env", "GITHUB_TOKEN"))

	fmt.Println()
	fmt.Println("Environment Variables:")
	if anthropicKey != "" {
		fmt.Printf("  ANTHROPIC_API_KEY: Set (%d characters)\n", len(anthropicKey))
	} else {
		fmt.Println("  ANTHROPIC_API_KEY: Not set")
	}

	if githubToken != "" {
		fmt.Printf("  GITHUB_TOKEN: Set (%d characters)\n", len(githubToken))
	} else {
		fmt.Println("  GITHUB_TOKEN: Not set")
	}

	return nil
}

func getEnvVarName(viperKey, defaultValue string) string {
	value := viper.GetString(viperKey)
	if value == "" {
		return defaultValue
	}
	return value
}
