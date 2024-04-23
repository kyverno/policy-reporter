package securityhub

import (
	"context"
	"fmt"
	"time"

	hub "github.com/aws/aws-sdk-go-v2/service/securityhub"
	"github.com/aws/aws-sdk-go-v2/service/securityhub/types"
	"go.uber.org/zap"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/target"
)

// Options to configure the S3 target
type Options struct {
	target.ClientOptions
	CustomFields map[string]string
	Client       *hub.Client
	AccountID    string
	Region       string
	ProductName  string
	Delay        time.Duration
	Cleanup      bool
}

type client struct {
	target.BaseClient
	customFields map[string]string
	hub          *hub.Client
	accountID    string
	region       string
	productName  string
	delay        time.Duration
	cleanup      bool
}

func (c *client) Send(result v1alpha2.PolicyReportResult) {
	generator := result.Policy
	if generator == "" {
		generator = result.Rule
	}

	title := generator
	if result.HasResource() {
		title = fmt.Sprintf("%s: %s", title, result.ResourceString())
	}

	t := time.Unix(result.Timestamp.Seconds, int64(result.Timestamp.Nanos))

	var accID *string
	if c.accountID != "" {
		accID = toPointer(c.accountID)
	}

	res, err := c.hub.BatchImportFindings(context.TODO(), &hub.BatchImportFindingsInput{
		Findings: []types.AwsSecurityFinding{
			{
				Id:            toPointer(result.GetID()),
				AwsAccountId:  accID,
				SchemaVersion: toPointer("2018-10-08"),
				ProductArn:    toPointer("arn:aws:securityhub:" + c.region + ":" + c.accountID + ":product/" + c.accountID + "/default"),
				GeneratorId:   toPointer(fmt.Sprintf("%s/%s", result.Source, generator)),
				Types:         []string{mapType(result.Source)},
				CreatedAt:     toPointer(t.Format("2006-01-02T15:04:05.999999999Z07:00")),
				UpdatedAt:     toPointer(t.Format("2006-01-02T15:04:05.999999999Z07:00")),
				Severity: &types.Severity{
					Label: mapSeverity(result.Severity),
				},
				Title:       &title,
				Description: &result.Message,
				ProductFields: map[string]string{
					"Product Name": c.productName,
				},
				Compliance: &types.Compliance{
					Status: types.ComplianceStatusFailed,
				},
				Resources: []types.Resource{
					{
						Type:      toPointer("Other"),
						Region:    &c.region,
						Partition: types.PartitionAws,
						Id:        mapResourceID(result),
						Details: &types.ResourceDetails{
							Other: c.mapOtherDetails(result),
						},
					},
				},
				RecordState: types.RecordStateActive,
			},
		},
	})
	if err != nil {
		zap.L().Error(c.Name()+": PUSH FAILED", zap.Error(err), zap.Any("response", res))
		return
	}

	zap.L().Info(c.Name()+": PUSH OK", zap.Int32("successCount", *res.SuccessCount), zap.Int32("failedCount", *res.FailedCount))
}

func (c *client) CleanUp(ctx context.Context, report v1alpha2.ReportInterface) {
	if !c.cleanup {
		return
	}

	resourceIds := toResourceIDFilter(report)
	if len(resourceIds) == 0 {
		return
	}

	findings, err := c.hub.GetFindings(ctx, &hub.GetFindingsInput{
		Filters: &types.AwsSecurityFindingFilters{
			Region: []types.StringFilter{
				{
					Comparison: types.StringFilterComparisonEquals,
					Value:      &c.region,
				},
			},
			Type: []types.StringFilter{
				{
					Comparison: types.StringFilterComparisonPrefix,
					Value:      toPointer(mapType(report.GetSource())),
				},
			},
			ResourceId: resourceIds,
			RecordState: []types.StringFilter{
				{
					Comparison: types.StringFilterComparisonEquals,
					Value:      toPointer("ACTIVE"),
				},
			},
		},
	})
	if err != nil {
		zap.L().Error(c.Name()+": failed to get findings", zap.Error(err))
		return
	}

	if len(findings.Findings) == 0 {
		time.Sleep(c.delay)
		return
	}

	mapping := make(map[string]types.AwsSecurityFinding, len(findings.Findings))
	for _, f := range findings.Findings {
		mapping[*f.Id] = f
	}

	for _, r := range report.GetResults() {
		if !c.BaseClient.Validate(report, r) {
			continue
		}

		delete(mapping, r.GetID())
	}

	if len(mapping) == 0 {
		time.Sleep(c.delay)
		return
	}

	list := make([]types.AwsSecurityFinding, 0)
	for _, f := range mapping {
		f.UpdatedAt = toPointer(time.Now().Format("2006-01-02T15:04:05.999999999Z07:00"))
		f.RecordState = types.RecordStateArchived
		f.Workflow = &types.Workflow{
			Status: types.WorkflowStatusResolved,
		}

		list = append(list, f)
	}

	if _, err = c.hub.BatchImportFindings(ctx, &hub.BatchImportFindingsInput{Findings: list}); err != nil {
		zap.L().Error(c.Name()+": failed to batch archived findings", zap.Error(err))
		time.Sleep(c.delay)
		return
	}

	zap.L().Info(c.Name()+": Findings updated", zap.Int("count", len(list)))
	time.Sleep(c.delay)
}

func (c *client) mapOtherDetails(result v1alpha2.PolicyReportResult) map[string]string {
	details := map[string]string{
		"Source":   result.Source,
		"Category": result.Category,
		"Policy":   result.Policy,
		"Rule":     result.Rule,
		"Result":   string(result.Result),
		"Priority": result.Priority.String(),
	}

	if len(c.customFields) > 0 {
		for property, value := range c.customFields {
			details[property] = value
		}

		for property, value := range result.Properties {
			details[property] = value
		}
	}

	if result.HasResource() {
		res := result.GetResource()

		if res.APIVersion != "" {
			details["Resource APIVersion"] = res.APIVersion
		}
		if res.Kind != "" {
			details["Resource Kind"] = res.Kind
		}
		if res.Namespace != "" {
			details["Resource Namespace"] = res.Namespace
		}
		if res.Name != "" {
			details["Resource Name"] = res.Name
		}
		if res.UID != "" {
			details["Resource UID"] = string(res.UID)
		}
	}

	return details
}

// NewClient creates a new S3.client to send Results to S3.
func NewClient(options Options) target.Client {
	return &client{
		target.NewBaseClient(options.ClientOptions),
		options.CustomFields,
		options.Client,
		options.AccountID,
		options.Region,
		options.ProductName,
		options.Delay,
		options.Cleanup,
	}
}

func toPointer[T any](value T) *T {
	return &value
}

func mapSeverity(s v1alpha2.PolicySeverity) types.SeverityLabel {
	switch s {
	case v1alpha2.SeverityInfo:
		return types.SeverityLabelInformational
	case v1alpha2.SeverityLow:
		return types.SeverityLabelLow
	case v1alpha2.SeverityMedium:
		return types.SeverityLabelMedium
	case v1alpha2.SeverityHigh:
		return types.SeverityLabelHigh
	case v1alpha2.SeverityCritical:
		return types.SeverityLabelCritical
	default:
		return types.SeverityLabelInformational
	}
}

func mapResourceID(result v1alpha2.PolicyReportResult) *string {
	if result.HasResource() {
		res := result.GetResource()
		if res.UID != "" {
			return toPointer(string(res.UID))
		}

		return toPointer(result.ResourceString())
	}

	return toPointer(result.GetID())
}

func mapType(source string) string {
	if source == "" {
		return "Software and Configuration Checks/Kubernetes Policies"
	}

	return "Software and Configuration Checks/Kubernetes Policies/" + source
}

func toResourceIDFilter(report v1alpha2.ReportInterface) []types.StringFilter {
	res := report.GetScope()
	if res != nil {
		var value string
		if res.UID != "" {
			value = string(res.UID)
		} else {
			value = v1alpha2.ToResourceString(res)
		}

		return []types.StringFilter{
			{
				Comparison: types.StringFilterComparisonEquals,
				Value:      toPointer(value),
			},
		}
	}

	if len(report.GetResults()) == 0 {
		return []types.StringFilter{
			{
				Comparison: types.StringFilterComparisonEquals,
				Value:      toPointer(report.GetName()),
			},
		}
	}

	list := map[string]bool{}
	for _, result := range report.GetResults() {
		list[*mapResourceID(result)] = true
	}

	filter := make([]types.StringFilter, 0, len(list))
	for id := range list {
		filter = append(filter, types.StringFilter{
			Comparison: types.StringFilterComparisonEquals,
			Value:      toPointer(id),
		})
	}

	return filter
}
