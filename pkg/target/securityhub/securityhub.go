package securityhub

import (
	"context"
	"fmt"
	"strings"
	"time"

	hub "github.com/aws/aws-sdk-go-v2/service/securityhub"
	"github.com/aws/aws-sdk-go-v2/service/securityhub/types"
	"go.uber.org/zap"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/helper"
	"github.com/kyverno/policy-reporter/pkg/target"
)

var schema = toPointer("2018-10-08")

type HubClient interface {
	BatchImportFindings(ctx context.Context, params *hub.BatchImportFindingsInput, optFns ...func(*hub.Options)) (*hub.BatchImportFindingsOutput, error)
	GetFindings(ctx context.Context, params *hub.GetFindingsInput, optFns ...func(*hub.Options)) (*hub.GetFindingsOutput, error)
	BatchUpdateFindings(ctx context.Context, params *hub.BatchUpdateFindingsInput, optFns ...func(*hub.Options)) (*hub.BatchUpdateFindingsOutput, error)
}

type PolrClient interface {
	Get(ctx context.Context, name, namespace string) (v1alpha2.ReportInterface, error)
}

// Options to configure the SecurityHub target
type Options struct {
	target.ClientOptions
	CustomFields map[string]string
	Client       HubClient
	AccountID    string
	Region       string
	ProductName  string
	CompanyName  string
	Delay        time.Duration
	Cleanup      bool
}

type client struct {
	target.BaseClient
	customFields map[string]string
	hub          HubClient
	accountID    string
	region       string
	productName  string
	companyName  string
	delay        time.Duration
	cleanup      bool
	arn          *string
	synced       bool
}

func (c *client) mapFindings(polr v1alpha2.ReportInterface, results []v1alpha2.PolicyReportResult) []types.AwsSecurityFinding {
	var accID *string
	if c.accountID != "" {
		accID = toPointer(c.accountID)
	}

	return helper.Map(results, func(result v1alpha2.PolicyReportResult) types.AwsSecurityFinding {
		generator := result.Policy
		if generator == "" {
			generator = result.Rule
		}

		title := generator
		if result.HasResource() {
			title = fmt.Sprintf("%s: %s", title, result.ResourceString())
		}

		t := time.Unix(result.Timestamp.Seconds, int64(result.Timestamp.Nanos))

		return types.AwsSecurityFinding{
			Id:            toPointer(result.GetID()),
			AwsAccountId:  accID,
			SchemaVersion: schema,
			ProductArn:    c.arn,
			GeneratorId:   toPointer(fmt.Sprintf("%s/%s", result.Source, generator)),
			Types:         []string{mapType(result.Source)},
			CreatedAt:     toPointer(t.Format("2006-01-02T15:04:05.999999999Z07:00")),
			UpdatedAt:     toPointer(t.Format("2006-01-02T15:04:05.999999999Z07:00")),
			Severity: &types.Severity{
				Label: MapSeverity(result.Severity),
			},
			Title:       &title,
			Description: &result.Message,
			ProductName: &c.productName,
			CompanyName: &c.companyName,
			Compliance: &types.Compliance{
				Status: types.ComplianceStatusFailed,
			},
			Workflow: &types.Workflow{
				Status: types.WorkflowStatusNew,
			},
			Resources: []types.Resource{
				{
					Type:      toPointer("Other"),
					Region:    &c.region,
					Partition: types.PartitionAws,
					Id:        mapResourceID(result),
					Details: &types.ResourceDetails{
						Other: c.mapOtherDetails(polr, result),
					},
				},
			},
			RecordState: types.RecordStateActive,
		}
	})
}

func (c *client) Send(result v1alpha2.PolicyReportResult) {
	c.BatchSend(&v1alpha2.PolicyReport{}, []v1alpha2.PolicyReportResult{result})
}

func filterResults(results []v1alpha2.PolicyReportResult) []v1alpha2.PolicyReportResult {
	return helper.Filter(results, func(r v1alpha2.PolicyReportResult) bool {
		if r.Result == v1alpha2.StatusFail {
			return true
		}
		if r.Result == v1alpha2.StatusWarn {
			return true
		}
		if r.Result == v1alpha2.StatusError {
			return true
		}

		return false
	})
}

func (c *client) BatchSend(polr v1alpha2.ReportInterface, results []v1alpha2.PolicyReportResult) {
	results = filterResults(results)
	if len(results) == 0 {
		return
	}

	list, err := c.getFindingsByIDs(context.Background(), polr, toResourceIDFilter(polr, results), "")
	if err != nil {
		zap.L().Error(c.Name()+": failed to get findings", zap.Error(err))
		return
	}

	list = filterFindings(list, results)
	findings := helper.Map(list, func(f types.AwsSecurityFinding) types.AwsSecurityFindingIdentifier {
		return types.AwsSecurityFindingIdentifier{
			Id:         f.Id,
			ProductArn: f.ProductArn,
		}
	})

	if len(findings) > 0 {
		updated, err := c.batchUpdate(context.Background(), findings, types.WorkflowStatusNew)
		if err != nil {
			zap.L().Error(c.Name()+": PUSH FAILED", zap.Error(err))
			return
		} else if updated > 0 {
			zap.L().Info(c.Name()+": PUSH OK", zap.Int("updated", updated))
		}

		mapping := make(map[string]bool, len(list))
		for _, f := range list {
			mapping[*f.Id] = true
		}

		results = helper.Filter(results, func(result v1alpha2.PolicyReportResult) bool {
			return !mapping[result.GetID()]
		})
	}

	if len(results) == 0 {
		return
	}

	res, err := c.hub.BatchImportFindings(context.Background(), &hub.BatchImportFindingsInput{
		Findings: c.mapFindings(polr, results),
	})
	if err != nil {
		zap.L().Error(c.Name()+": PUSH FAILED", zap.Error(err), zap.Any("response", res))
		return
	}

	zap.L().Info(c.Name()+": PUSH OK", zap.Int32("imported", *res.SuccessCount), zap.Int32("failed", *res.FailedCount), zap.String("report", polr.GetKey()))
}

func (c *client) Sync(ctx context.Context) error {
	if !c.cleanup {
		return nil
	}
	defer zap.L().Info(c.Name() + ": START SYNC")

	list, err := c.getFindings(ctx)
	if err != nil {
		zap.L().Error(c.Name()+": failed to get findings", zap.Error(err))
		return err
	}

	if len(list) == 0 {
		zap.L().Info(c.Name() + ": no findings to sync")
		return nil
	}

	findings := helper.Map(list, func(f types.AwsSecurityFinding) types.AwsSecurityFindingIdentifier {
		return types.AwsSecurityFindingIdentifier{
			Id:         f.Id,
			ProductArn: f.ProductArn,
		}
	})

	count, err := c.batchUpdate(ctx, findings, types.WorkflowStatusResolved)
	if err != nil {
		zap.L().Error(c.Name()+": failed to sync findings", zap.Error(err))
		return err
	}

	zap.L().Info(c.Name()+": FINISHED SYNC", zap.Int("updated", count))

	return nil
}

func (c *client) CleanUp(ctx context.Context, report v1alpha2.ReportInterface) {
	if !c.cleanup {
		return
	}

	zap.L().Info(c.Name()+": start cleanup", zap.String("report", report.GetKey()))

	if report.GetSource() != "" {
		if !c.BaseClient.ValidateReport(report) {
			return
		}
	}

	resourceIds := toResourceIDFilter(report, report.GetResults())

	findings, err := c.getFindingsByIDs(ctx, report, resourceIds, "")
	if err != nil {
		zap.L().Error(c.Name()+": failed to get findings", zap.Error(err))
		return
	}
	defer time.Sleep(c.delay)

	if len(findings) == 0 {
		return
	}

	mapping := make(map[string]types.AwsSecurityFinding, len(findings))
	for _, f := range findings {
		mapping[*f.Id] = f
	}

	for _, r := range report.GetResults() {
		if !c.BaseClient.Validate(report, r) {
			continue
		}

		delete(mapping, r.GetID())
	}

	if len(mapping) == 0 {
		return
	}

	list := make([]types.AwsSecurityFindingIdentifier, 0, len(mapping))
	for _, f := range mapping {
		list = append(list, types.AwsSecurityFindingIdentifier{
			Id:         f.Id,
			ProductArn: f.ProductArn,
		})
	}

	count, err := c.batchUpdate(ctx, list, types.WorkflowStatusResolved)
	if err != nil {
		zap.L().Error(c.Name()+": failed to batch archived findings", zap.Error(err))
		return
	}

	zap.L().Info(c.Name()+": CLEANUP OK", zap.Int("count", count), zap.String("report", report.GetKey()))
}

func (c *client) mapOtherDetails(polr v1alpha2.ReportInterface, result v1alpha2.PolicyReportResult) map[string]string {
	details := map[string]string{
		"Source":   result.Source,
		"Category": result.Category,
		"Policy":   result.Policy,
		"Rule":     result.Rule,
		"Result":   string(result.Result),
		"Report":   polr.GetKey(),
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

func (c *client) getFindings(ctx context.Context) ([]types.AwsSecurityFinding, error) {
	list := make([]types.AwsSecurityFinding, 0)

	var token *string

	for {
		resp, err := c.hub.GetFindings(ctx, &hub.GetFindingsInput{
			NextToken: token,
			Filters:   c.BaseFilter(nil),
		})
		if err != nil {
			return nil, err
		}

		if len(resp.Findings) == 0 {
			return list, nil
		}

		list = append(list, resp.Findings...)
		if resp.NextToken == nil {
			return list, nil
		}

		token = resp.NextToken
	}
}

func (c *client) batchUpdate(ctx context.Context, findings []types.AwsSecurityFindingIdentifier, status types.WorkflowStatus) (int, error) {
	if len(findings) == 0 {
		return 0, nil
	}

	chunks := helper.ChunkSlice(findings, 100)

	var updated int
	for _, chunk := range chunks {
		response, err := c.hub.BatchUpdateFindings(ctx, &hub.BatchUpdateFindingsInput{
			FindingIdentifiers: chunk,
			Workflow: &types.WorkflowUpdate{
				Status: status,
			},
		})
		if err != nil {
			return updated, err
		}

		updated += len(response.ProcessedFindings)
	}

	return updated, nil
}

func (c *client) getFindingsByIDs(ctx context.Context, report v1alpha2.ReportInterface, resources []types.StringFilter, status string) ([]types.AwsSecurityFinding, error) {
	list := make([]types.AwsSecurityFinding, 0)

	chunks := helper.ChunkSlice(resources, 20)

	for _, res := range chunks {
		filter := c.BaseFilter(report)
		if len(res) > 0 {
			filter.ResourceId = res
		}

		if status != "" {
			filter.WorkflowStatus = []types.StringFilter{
				{
					Comparison: types.StringFilterComparisonEquals,
					Value:      toPointer(status),
				},
			}
		}

		var token *string

		for {
			resp, err := c.hub.GetFindings(ctx, &hub.GetFindingsInput{
				NextToken: token,
				Filters:   filter,
			})
			if err != nil {
				return nil, err
			}

			if len(resp.Findings) == 0 {
				break
			}

			list = append(list, resp.Findings...)
			if resp.NextToken == nil {
				break
			}

			token = resp.NextToken
		}
	}

	return list, nil
}

func (c *client) BaseFilter(report v1alpha2.ReportInterface) *types.AwsSecurityFindingFilters {
	source := ""
	if report != nil {
		source = report.GetSource()
	}

	filter := &types.AwsSecurityFindingFilters{
		ProductArn: []types.StringFilter{
			{
				Comparison: types.StringFilterComparisonEquals,
				Value:      c.arn,
			},
		},
		AwsAccountId: []types.StringFilter{
			{
				Comparison: types.StringFilterComparisonEquals,
				Value:      &c.accountID,
			},
		},
		Region: []types.StringFilter{
			{
				Comparison: types.StringFilterComparisonEquals,
				Value:      &c.region,
			},
		},
		Type: []types.StringFilter{
			{
				Comparison: types.StringFilterComparisonPrefix,
				Value:      toPointer(mapType(source)),
			},
		},
		ProductName: []types.StringFilter{
			{
				Comparison: types.StringFilterComparisonEquals,
				Value:      &c.productName,
			},
		},
		RecordState: []types.StringFilter{
			{
				Comparison: types.StringFilterComparisonEquals,
				Value:      toPointer("ACTIVE"),
			},
		},
	}

	if report != nil {
		filter.ResourceDetailsOther = []types.MapFilter{
			{
				Comparison: types.MapFilterComparisonEquals,
				Key:        toPointer("Report"),
				Value:      toPointer(report.GetKey()),
			},
		}
	}

	return filter
}

func (c *client) Type() target.ClientType {
	if !c.cleanup {
		return target.BatchSend
	}

	return target.SyncSend
}

// NewClient creates a new SecurityHub.client to send Results to SecurityHub.
func NewClient(options Options) *client {
	if options.Delay == 0 {
		options.Delay = 2 * time.Second
	}

	return &client{
		target.NewBaseClient(options.ClientOptions),
		options.CustomFields,
		options.Client,
		options.AccountID,
		options.Region,
		options.ProductName,
		options.CompanyName,
		options.Delay,
		options.Cleanup,
		toPointer("arn:aws:securityhub:" + options.Region + ":" + options.AccountID + ":product/" + options.AccountID + "/default"),
		false,
	}
}

func toPointer[T any](value T) *T {
	return &value
}

func MapSeverity(s v1alpha2.PolicySeverity) types.SeverityLabel {
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

func toResourceIDFilter(report v1alpha2.ReportInterface, results []v1alpha2.PolicyReportResult) []types.StringFilter {
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

	if len(results) == 0 {
		return []types.StringFilter{
			{
				Comparison: types.StringFilterComparisonEquals,
				Value:      toPointer(report.GetName()),
			},
		}
	}

	list := map[string]bool{}
	for _, result := range results {
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

func splitPolrKey(key string) (string, string) {
	parts := strings.Split(key, "/")
	if len(parts) == 1 {
		return parts[0], ""
	}

	return parts[1], parts[0]
}

func filterFindings(findings []types.AwsSecurityFinding, results []v1alpha2.PolicyReportResult) []types.AwsSecurityFinding {
	filtered := make([]types.AwsSecurityFinding, 0, len(findings))

	mapping := make(map[string]bool, len(results))
	for _, r := range results {
		mapping[r.GetID()] = true
	}

	for _, finding := range findings {
		if _, ok := mapping[*finding.Id]; ok {
			filtered = append(filtered, finding)
		}
	}

	return filtered
}
