package target

import (
	"strings"

	"github.com/kyverno/go-wildcard"
	"go.uber.org/zap"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/helper"
	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/kyverno/policy-reporter/pkg/validate"
)

// Client for a provided Target
type Client interface {
	// Send the given Result to the configured Target
	Send(result v1alpha2.PolicyReportResult)
	// SkipExistingOnStartup skips already existing PolicyReportResults on startup
	SkipExistingOnStartup() bool
	// Name is a unique identifier for each Target
	Name() string
	// Validate if a result should send
	Validate(rep v1alpha2.ReportInterface, result v1alpha2.PolicyReportResult) bool
	// MinimumPriority for a triggered Result to send to this target
	MinimumPriority() string
	// Sources of the Results which should send to this target, empty means all sources
	Sources() []string
}

func NewResultFilter(namespace, priority, policy validate.RuleSets, minimumPriority string, sources []string) *report.ResultFilter {
	f := report.NewResultFilter()
	f.Sources = sources
	f.MinimumPriority = minimumPriority

	if len(sources) > 0 {
		f.AddValidation(func(r v1alpha2.PolicyReportResult) bool {
			return helper.Contains(r.Source, sources)
		})
	}

	if namespace.Count() > 0 {
		f.AddValidation(func(r v1alpha2.PolicyReportResult) bool {
			if r.GetResource() == nil {
				return true
			}

			return validate.Namespace(r.GetResource().Namespace, namespace)
		})
	}

	if minimumPriority != "" {
		f.AddValidation(func(r v1alpha2.PolicyReportResult) bool {
			return r.Priority >= v1alpha2.NewPriority(f.MinimumPriority)
		})
	}

	if policy.Count() > 0 {
		f.AddValidation(func(r v1alpha2.PolicyReportResult) bool {
			return validate.MatchRuleSet(r.Policy, policy)
		})
	}

	if priority.Count() > 0 {
		f.AddValidation(func(r v1alpha2.PolicyReportResult) bool {
			return validate.ContainsRuleSet(r.Priority.String(), priority)
		})
	}

	return f
}

func NewReportFilter(labels validate.RuleSets) *report.ReportFilter {
	f := report.NewReportFilter()
	if labels.Count() > 0 {
		f.AddValidation(func(r v1alpha2.ReportInterface) bool {
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

	return f
}

type BaseClient struct {
	name                  string
	skipExistingOnStartup bool
	resultFilter          *report.ResultFilter
	reportFilter          *report.ReportFilter
	logger                *zap.Logger
}

type ClientOptions struct {
	Name                  string
	SkipExistingOnStartup bool
	ResultFilter          *report.ResultFilter
	ReportFilter          *report.ReportFilter
	Logger                *zap.Logger
}

func (c *BaseClient) Name() string {
	return c.name
}

func (c *BaseClient) Logger() *zap.Logger {
	return c.logger
}

func (c *BaseClient) MinimumPriority() string {
	if c.resultFilter == nil {
		return v1alpha2.DefaultPriority.String()
	}

	return c.resultFilter.MinimumPriority
}

func (c *BaseClient) Sources() []string {
	if c.resultFilter == nil {
		return make([]string, 0)
	}

	return c.resultFilter.Sources
}

func (c *BaseClient) Validate(rep v1alpha2.ReportInterface, result v1alpha2.PolicyReportResult) bool {
	if rep == nil {
		return false
	}

	if c.reportFilter != nil && !c.reportFilter.Validate(rep) {
		return false
	}

	if c.resultFilter != nil && !c.resultFilter.Validate(result) {
		return false
	}

	return true
}

func (c *BaseClient) SkipExistingOnStartup() bool {
	return c.skipExistingOnStartup
}

func NewBaseClient(options ClientOptions) BaseClient {
	return BaseClient{options.Name, options.SkipExistingOnStartup, options.ResultFilter, options.ReportFilter, options.Logger}
}
