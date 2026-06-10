package send

import (
	"flag"
	"strings"
	"sync"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/kyverno/policy-reporter/pkg/config"
	"github.com/kyverno/policy-reporter/pkg/email/summary"
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
			if c.K8sClient.Kubeconfig != "" {
				k8sConfig, err = clientcmd.BuildConfigFromFlags("", c.K8sClient.Kubeconfig)
			} else {
				k8sConfig, err = rest.InClusterConfig()
			}
			if err != nil {
				return err
			}

			resolver := config.NewResolver(c, k8sConfig)
			logger, err := resolver.Logger()
			if err != nil {
				return err
			}

			generator, err := resolver.SummaryGenerator()
			if err != nil {
				return err
			}

			data, err := generator.GenerateData(cmd.Context())
			if err != nil {
				logger.Error("failed to generate report data", zap.Error(err))
				return err
			}

			reporter := resolver.SummaryReporter()

			wg := &sync.WaitGroup{}
			wg.Add(1 + len(c.EmailReports.Summary.Channels))

			go func() {
				defer wg.Done()

				if len(c.EmailReports.Summary.To) == 0 {
					logger.Info("skipped - no email configured")
					return
				}

				report, err := reporter.Report(data, c.EmailReports.Summary.Format)
				if err != nil {
					logger.Error("failed to create report", zap.Error(err))
					return
				}

				err = resolver.EmailClient().Send(report, c.EmailReports.Summary.To)
				if err != nil {
					logger.Error("failed to send report", zap.Error(err))
					return
				}

				logger.Sugar().Infof("email sent to %s\n", strings.Join(c.EmailReports.Summary.To, ", "))
			}()

			nsclient, err := resolver.NamespaceClient()
			if err != nil {
				logger.Error("failed to get namespace client", zap.Error(err))
				return err
			}

			for _, ch := range c.EmailReports.Summary.Channels {
				go func(channel config.EmailReport) {
					defer wg.Done()

					if len(channel.To) == 0 {
						logger.Info("skipped - no channel email configured")
						return
					}

					sources := summary.FilterSources(data, config.EmailReportFilterFromConfig(nsclient, channel.Filter), !channel.Filter.DisableClusterReports)
					if len(sources) == 0 {
						logger.Info("skip email - no results to send")
						return
					}

					report, err := reporter.Report(sources, channel.Format)
					if err != nil {
						logger.Error("failed to create report", zap.Error(err))
						return
					}

					err = resolver.EmailClient().Send(report, channel.To)
					if err != nil {
						logger.Error("failed to send report", zap.Error(err))
						return
					}

					logger.Sugar().Infof("email sent to %s\n", strings.Join(channel.To, ", "))
				}(ch)
			}

			wg.Wait()

			return nil
		},
	}

	flag.Parse()

	return cmd
}
