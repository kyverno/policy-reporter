package target

import (
	"context"
	"strings"

	"github.com/kyverno/go-wildcard"
	"go.uber.org/zap"
	"openreports.io/apis/openreports.io/v1alpha1"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/kubernetes/namespaces"
	"github.com/kyverno/policy-reporter/pkg/openreports"
	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/kyverno/policy-reporter/pkg/validate"
)

type ClientType = string

const (
	SingleSend ClientType = "single"
	BatchSend  ClientType = "batch"
	SyncSend   ClientType = "sync"
)

// Client for a provided Target
type Client interface {
	// Send the given Result to the configured Target
	Send(result openreports.ResultAdapter)
	// BatchSend the given Results of a single PolicyReport to the configured Target
	BatchSend(report openreports.ReportInterface, results []openreports.ResultAdapter)
	// SkipExistingOnStartup skips already existing PolicyReportResults on startup
	SkipExistingOnStartup() bool
	// Name is a unique identifier for each Target
	Name() string
	// Validate if a result should send
	Validate(rep openreports.ReportInterface, result openreports.ResultAdapter) bool
	// MinimumSeverity for a triggered Result to send to this target
	MinimumSeverity() string
	// Sources of the Results which should send to this target, empty means all sources
	Sources() []string
	// Type for the given target
	Type() ClientType
	// CleanUp old results if supported by the target
	CleanUp(context.Context, openreports.ReportInterface)
	// Reset the current state in the related target
	Reset(context.Context) error
	// SendHeartbeat sends a periodic keepalive message
	SendHeartbeat()
}

type ResultFilterFactory struct {
	client namespaces.Client
}

func (rf *ResultFilterFactory) CreateFilter(namespace, severity, status, policy, sources validate.RuleSets, minimumSeverity string) *report.ResultFilter {
	f := report.NewResultFilter()
	f.Sources = sources.Include
	f.MinimumSeverity = minimumSeverity

	if namespace.Count() > 0 {
		f.AddValidation(func(r openreports.ResultAdapter) bool {
			if r.GetResource() == nil {
				return true
			}

			return validate.Namespace(r.GetResource().Namespace, namespace)
		})
	}

	if len(namespace.Selector) > 0 {
		f.AddValidation(func(r openreports.ResultAdapter) bool {
			if r.GetResource() == nil || r.GetResource().Namespace == "" {
				return true
			}

			namespaces, err := rf.client.List(context.Background(), namespace.Selector)
			if err != nil {
				zap.L().Error("failed to resolve namespace selector", zap.Error(err))
				return false
			}

			return validate.Namespace(r.GetResource().Namespace, validate.RuleSets{Include: namespaces})
		})
	}

	if minimumSeverity != "" {
		f.AddValidation(func(r openreports.ResultAdapter) bool {
			return openreports.SeverityLevel[r.Severity] >= openreports.SeverityLevel[v1alpha1.ResultSeverity(f.MinimumSeverity)]
		})
	}

	if sources.Count() > 0 {
		f.AddValidation(func(r openreports.ResultAdapter) bool {
			return validate.MatchRuleSet(r.Source, sources)
		})
	}

	if policy.Count() > 0 {
		f.AddValidation(func(r openreports.ResultAdapter) bool {
			return validate.MatchRuleSet(r.Policy, policy)
		})
	}

	if severity.Count() > 0 {
		f.AddValidation(func(r openreports.ResultAdapter) bool {
			return validate.ContainsRuleSet(string(r.Severity), severity)
		})
	}

	if status.Count() > 0 {
		f.AddValidation(func(r openreports.ResultAdapter) bool {
			return validate.ContainsRuleSet(string(r.Result), status)
		})
	}

	return f
}

func NewReportFilter(labels, sources validate.RuleSets) *report.ReportFilter {
	f := report.NewReportFilter()

	if labels.Count() > 0 {
		f.AddValidation(func(r openreports.ReportInterface) bool {
			if len(labels.Include) > 0 {
				for _, label := range labels.Include {
					parts := strings.Split(label, ":")
					if len(parts) == 1 {
						parts = append(parts, "*")
					}

					labelName := strings.TrimSpace(parts[0])
					labelValue := strings.TrimSpace(parts[1])

					for key, value := range r.GetLabels() {
						if labelName == key && wildcard.Match(labelValue, value) {
							return true
						}
					}
				}

				return false
			} else if len(labels.Exclude) > 0 {
				for _, label := range labels.Exclude {
					parts := strings.Split(label, ":")
					if len(parts) == 1 {
						parts = append(parts, "*")
					}

					labelName := strings.TrimSpace(parts[0])
					labelValue := strings.TrimSpace(parts[1])

					for key, value := range r.GetLabels() {
						if labelName == key && wildcard.Match(labelValue, value) {
							return false
						}
					}
				}
			}

			return true
		})
	}

	if sources.Count() > 0 {
		f.AddValidation(func(r openreports.ReportInterface) bool {
			source := r.GetSource()
			if source == "" {
				return true
			}

			return validate.MatchRuleSet(source, sources)
		})
	}

	return f
}

func NewResultFilterFactory(client namespaces.Client) *ResultFilterFactory {
	return &ResultFilterFactory{client: client}
}

type BaseClient struct {
	name                  string
	skipExistingOnStartup bool
	resultFilter          *report.ResultFilter
	reportFilter          *report.ReportFilter
}

type ClientOptions struct {
	Name                  string
	SkipExistingOnStartup bool
	ResultFilter          *report.ResultFilter
	ReportFilter          *report.ReportFilter
}

func (c *BaseClient) Name() string {
	return c.name
}

func (c *BaseClient) MinimumSeverity() string {
	if c.resultFilter == nil {
		return v1alpha2.SeverityInfo
	}

	return c.resultFilter.MinimumSeverity
}

func (c *BaseClient) Sources() []string {
	if c.resultFilter == nil {
		return make([]string, 0)
	}

	return c.resultFilter.Sources
}

func (c *BaseClient) Validate(rep openreports.ReportInterface, result openreports.ResultAdapter) bool {
	if !c.ValidateReport(rep) {
		return false
	}

	if c.resultFilter != nil && !c.resultFilter.Validate(result) {
		return false
	}

	return true
}

func (c *BaseClient) ValidateReport(rep openreports.ReportInterface) bool {
	if rep == nil {
		return false
	}

	if c.reportFilter != nil && !c.reportFilter.Validate(rep) {
		return false
	}

	return true
}

func (c *BaseClient) SkipExistingOnStartup() bool {
	return c.skipExistingOnStartup
}

func (c *BaseClient) Reset(_ context.Context) error {
	return nil
}

func (c *BaseClient) CleanUp(_ context.Context, _ openreports.ReportInterface) {}

func (c *BaseClient) BatchSend(_ openreports.ReportInterface, _ []openreports.ResultAdapter) {}

func (c *BaseClient) SendHeartbeat() {} // Default no-op implementation

func NewBaseClient(options ClientOptions) BaseClient {
	return BaseClient{options.Name, options.SkipExistingOnStartup, options.ResultFilter, options.ReportFilter}
}
