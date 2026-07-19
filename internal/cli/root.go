package cli

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "cogniq",
	Short: "CogniQ - One context. Every AI.",
	Long:  `CogniQ is a local-first context engine that indexes repositories and exposes them through a unified interface for AI assistants.`,
}

func Execute() error {
	return rootCmd.Execute()
}