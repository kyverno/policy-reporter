package cmd

import (
	"github.com/spf13/cobra"
)

// NewCLI creates a new instance of the root CLI
func NewCLI() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "policyreporter",
		Short: "Generates PolicyReport Metrics and Send Results to different targets",
		Long: `Generates Prometheus Metrics from PolicyReports, ClusterPolicyReports and PolicyReportResults.
		Sends notifications to different targets like Grafana's Loki.`,
	}

	rootCmd.AddCommand(newRunCMD())
	rootCmd.AddCommand(newSendCMD())

	return rootCmd
}
