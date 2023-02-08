package cmd

import (
	"flag"

	"github.com/spf13/cobra"

	"github.com/kyverno/policy-reporter/cmd/send"
)

func newSendCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send",
		Short: "Send different kinds of email reports",
	}

	// For local usage
	cmd.PersistentFlags().StringP("kubeconfig", "k", "", "absolute path to the kubeconfig file")
	cmd.PersistentFlags().StringP("config", "c", "", "target configuration file")
	cmd.PersistentFlags().StringP("template-dir", "t", "./templates", "template directory for email reports")
	cmd.AddCommand(send.NewSummaryCMD())
	cmd.AddCommand(send.NewViolationsCMD())

	flag.Parse()

	return cmd
}
