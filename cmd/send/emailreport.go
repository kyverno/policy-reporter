package send

import (
	"context"
	"flag"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/kyverno/policy-reporter/pkg/config"
	policyreporterv1alpha1 "github.com/kyverno/policy-reporter/pkg/crd/api/targetconfig/v1alpha1"
	crds "github.com/kyverno/policy-reporter/pkg/crd/client/clientset/versioned"
	"github.com/kyverno/policy-reporter/pkg/emailreport"
)

func NewEmailReportCMD() *cobra.Command {
	var namespace, name, uid string
	cmd := &cobra.Command{
		Use:   "email-report",
		Short: "Send a namespaced EmailReport",
		RunE: func(cmd *cobra.Command, _ []string) error {
			c, err := config.Load(cmd)
			if err != nil {
				return err
			}
			logger, err := config.SetupLogger(c)
			if err != nil {
				return err
			}
			if err := config.SetupMemLimit(cmd.Context(), c); err != nil {
				return err
			}

			k8sConfig, err := emailReportKubernetesConfig(c)
			if err != nil {
				return err
			}
			client, err := crds.NewForConfig(k8sConfig)
			if err != nil {
				return err
			}
			resolver := config.NewResolver(c, k8sConfig)
			runner := emailreport.NewRunner(client.PolicyreporterV1alpha1(), func(ctx context.Context, report *policyreporterv1alpha1.EmailReport) error {
				return deliverEmailReport(ctx, resolver, report)
			})
			if err := runner.Run(cmd.Context(), namespace, name, types.UID(uid)); err != nil {
				logger.Error("failed to send EmailReport", zap.String("namespace", namespace), zap.String("name", name), zap.Error(err))
				return err
			}
			logger.Info("EmailReport sent", zap.String("namespace", namespace), zap.String("name", name))
			return nil
		},
	}

	cmd.Flags().StringVar(&namespace, "namespace", "", "EmailReport namespace")
	cmd.Flags().StringVar(&name, "name", "", "EmailReport name")
	cmd.Flags().StringVar(&uid, "uid", "", "EmailReport UID")
	_ = cmd.MarkFlagRequired("namespace")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("uid")
	cmd.PersistentFlags().Bool("auto-memory-enabled", true, "Enable automatic GOMEMLIMIT configuration based on container or system memory.")
	cmd.PersistentFlags().Float64("auto-memory-ratio", 0.9, "The ratio of reserved GOMEMLIMIT memory to the detected maximum container or system memory. Must be greater than 0 and less than or equal to 1.")
	flag.Parse()
	return cmd
}

func emailReportKubernetesConfig(c *config.Config) (*rest.Config, error) {
	if c.K8sClient.Kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", c.K8sClient.Kubeconfig)
	}
	return rest.InClusterConfig()
}

func deliverEmailReport(ctx context.Context, resolver *config.Resolver, report *policyreporterv1alpha1.EmailReport) error {
	var renderedReportErr error
	switch report.Spec.Type {
	case policyreporterv1alpha1.EmailReportTypeSummary:
		generator, err := resolver.SummaryGeneratorForNamespace(report.Namespace)
		if err != nil {
			return err
		}
		data, err := generator.GenerateData(ctx)
		if err != nil {
			return err
		}
		rendered, err := resolver.SummaryReporter().Report(data, "html")
		if err != nil {
			return err
		}
		renderedReportErr = resolver.EmailClient().Send(rendered, report.Spec.To)
	case policyreporterv1alpha1.EmailReportTypeViolations:
		generator, err := resolver.ViolationsGeneratorForNamespace(report.Namespace)
		if err != nil {
			return err
		}
		data, err := generator.GenerateData(ctx)
		if err != nil {
			return err
		}
		rendered, err := resolver.ViolationsReporter().Report(data, "html")
		if err != nil {
			return err
		}
		renderedReportErr = resolver.EmailClient().Send(rendered, report.Spec.To)
	default:
		return fmt.Errorf("unsupported report type %q", report.Spec.Type)
	}
	if renderedReportErr != nil {
		return renderedReportErr
	}
	zap.L().Info("email sent", zap.String("to", strings.Join(report.Spec.To, ", ")))
	return nil
}
