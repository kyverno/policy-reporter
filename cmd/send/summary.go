package send

import (
	"log"
	"sync"

	"github.com/kyverno/policy-reporter/pkg/config"
	"github.com/kyverno/policy-reporter/pkg/email/summary"

	"github.com/spf13/cobra"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func NewSummaryCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "summary",
		Short: "Send a summary e-mail to the configured emails",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := config.Load(cmd)
			if err != nil {
				return err
			}

			var k8sConfig *rest.Config
			if c.Kubeconfig != "" {
				k8sConfig, err = clientcmd.BuildConfigFromFlags("", c.Kubeconfig)
			} else {
				k8sConfig, err = rest.InClusterConfig()
			}
			if err != nil {
				return err
			}

			resolver := config.NewResolver(c, k8sConfig)

			generator, err := resolver.SummaryGenerator()
			if err != nil {
				return err
			}

			data, err := generator.GenerateData(cmd.Context())
			if err != nil {
				log.Printf("[ERROR] failed to generate report data: %s\n", err)
				return err
			}

			reporter := resolver.SummaryReport()

			wg := &sync.WaitGroup{}
			wg.Add(1 + len(c.EmailReports.Summary.Channels))

			go func() {
				defer wg.Done()

				if len(c.EmailReports.Summary.To) == 0 {
					return
				}

				report, err := reporter.Report(data, c.EmailReports.Summary.Format)
				if err != nil {
					log.Printf("[ERROR] failed to create report: %s\n", err)
					return
				}

				err = resolver.EmailClient().Send(report, c.EmailReports.Summary.To)
				if err != nil {
					log.Printf("[ERROR] failed to send report: %s\n", err)
				}
			}()

			for _, ch := range c.EmailReports.Violations.Channels {
				go func(channel config.EmailReport) {
					defer wg.Done()

					if len(channel.To) == 0 {
						return
					}

					sources := summary.FilterSources(data, config.EmailReportFilterFromConfig(channel.Filter), !channel.DisableClusterReports)
					if len(sources) == 0 {
						log.Printf("[INFO] skip email - no results to send")
						return
					}

					report, err := reporter.Report(sources, channel.Format)
					if err != nil {
						log.Printf("[ERROR] failed to create report: %s\n", err)
						return
					}

					err = resolver.EmailClient().Send(report, channel.To)
					if err != nil {
						log.Printf("[ERROR] failed to send report: %s\n", err)
					}
				}(ch)
			}

			wg.Wait()

			return nil
		},
	}

	return cmd
}
