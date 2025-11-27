package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	quiet   bool
)

var rootCmd = &cobra.Command{
	Use:   "issue-finder",
	Short: "Find GitHub OSS contribution opportunities matching your skills",
	Long: `Issue Finder is an AI-powered CLI tool that searches GitHub
for open source issue opportunities matching your skills, interests,
and experience level.

It uses Claude AI to evaluate issues and suggest the best matches,
saving you hours of manual searching.`,
	Version: "1.0.0",
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.issue-finder.yaml)")
	rootCmd.PersistentFlags().BoolVar(&quiet, "quiet", false, "suppress progress output")
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error finding home directory: %v\n", err)
			os.Exit(1)
		}

		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".issue-finder")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil && !quiet {
		fmt.Fprintf(os.Stderr, "Using config file: %s\n", viper.ConfigFileUsed())
	}
}
