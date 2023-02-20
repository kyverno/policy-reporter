package send

import (
	"log"
	"strings"
	"sync"

	"github.com/spf13/cobra"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/kyverno/policy-reporter/pkg/config"
	"github.com/kyverno/policy-reporter/pkg/email/violations"
)

func NewViolationsCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "violations",
		Short: "Send a violations e-mail to the configured emails",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := config.Load(cmd)
			if err != nil {
				return err
			}

			var k8sConfig *rest.Config
			if c.K8sClient.Kubeconfig != "" {
				k8sConfig, err = clientcmd.BuildConfigFromFlags("", c.K8sClient.Kubeconfig)
			} else {
				k8sConfig, err = rest.InClusterConfig()
			}
			if err != nil {
				return err
			}

			resolver := config.NewResolver(c, k8sConfig)

			generator, err := resolver.ViolationsGenerator()
			if err != nil {
				return err
			}

			data, err := generator.GenerateData(cmd.Context())
			if err != nil {
				log.Printf("[ERROR] failed to generate report data: %s\n", err)
				return err
			}

			reporter := resolver.ViolationsReporter()

			wg := &sync.WaitGroup{}
			wg.Add(1 + len(c.EmailReports.Violations.Channels))

			go func() {
				defer wg.Done()

				if len(c.EmailReports.Violations.To) == 0 {
					log.Print("[INFO] skipped - no email configured")
					return
				}

				report, err := reporter.Report(data, c.EmailReports.Violations.Format)
				if err != nil {
					log.Printf("[ERROR] failed to create report: %s\n", err)
					return
				}

				err = resolver.EmailClient().Send(report, c.EmailReports.Violations.To)
				if err != nil {
					log.Printf("[ERROR] failed to send report: %s\n", err)
					return
				}

				log.Printf("[INFO] email sent to %s\n", strings.Join(c.EmailReports.Violations.To, ", "))
			}()

			for _, ch := range c.EmailReports.Violations.Channels {
				go func(channel config.EmailReport) {
					defer wg.Done()

					if len(channel.To) == 0 {
						log.Print("[INFO] skipped - no channel email configured")
						return
					}

					sources := violations.FilterSources(data, config.EmailReportFilterFromConfig(channel.Filter), !channel.Filter.DisableClusterReports)
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
						return
					}

					log.Printf("[INFO] email sent to %s\n", strings.Join(channel.To, ", "))
				}(ch)
			}

			wg.Wait()

			return nil
		},
	}

	return cmd
}
