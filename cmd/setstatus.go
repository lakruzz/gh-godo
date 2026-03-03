// Package cmd provides command-line interface commands for the gh-godo extension.
package cmd

import (
	"github.com/lakruzz/gh-godo/cmd/setstatus"
	"github.com/spf13/cobra"
)

var (
	statusSHA         string
	statusState       string
	statusContext     string
	statusDescription string
	statusTargetURL   string
	statusRepo        string
)

var setstatusCmd = &cobra.Command{
	Use:   "set-status",
	Short: "Set a commit status during a workflow run",
	Long: `Set the status of a commit using the GitHub Commit Statuses API.

The state must be one of: error, failure, pending, success.

Usage:
  godo set-status --sha <commit-sha> --state <state> [--context <context>] [--description <description>] [--url <target-url>] [--repo <owner/repo>]

Examples:
  godo set-status --sha abc1234 --state success --context "ci/build" --description "Build passed"
  godo set-status --sha abc1234 --state failure --context "ci/test" --description "Tests failed" --url https://example.com/logs`,
	RunE: func(_ *cobra.Command, _ []string) error {
		return setstatus.Run(statusSHA, statusState, statusContext, statusDescription, statusTargetURL, statusRepo)
	},
}

func init() {
	rootCmd.AddCommand(setstatusCmd)

	setstatusCmd.Flags().StringVarP(&statusSHA, "sha", "s", "", "Commit SHA to set the status for (required)")
	setstatusCmd.Flags().StringVarP(&statusState, "state", "S", "", "Status state: error, failure, pending, or success (required)")
	setstatusCmd.Flags().StringVarP(&statusContext, "context", "c", "default", "A string label to differentiate this status from others (optional)")
	setstatusCmd.Flags().StringVarP(&statusDescription, "description", "d", "", "A short description of the status (optional)")
	setstatusCmd.Flags().StringVarP(&statusTargetURL, "url", "u", "", "URL to associate with the status (optional)")
	setstatusCmd.Flags().StringVarP(&statusRepo, "repo", "r", "", "Repository in owner/repo format (optional, defaults to current repo)")
	_ = setstatusCmd.MarkFlagRequired("sha")
	_ = setstatusCmd.MarkFlagRequired("state")
}
