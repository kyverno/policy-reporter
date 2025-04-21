package securityhub

import (
	"context"
	"time"

	hub "github.com/aws/aws-sdk-go-v2/service/securityhub"
	"github.com/aws/aws-sdk-go-v2/service/securityhub/types"
	"go.uber.org/zap"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/helper"
	"github.com/kyverno/policy-reporter/pkg/payload"
	"github.com/kyverno/policy-reporter/pkg/payload/scutils"
	"github.com/kyverno/policy-reporter/pkg/target"
)

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
	Synchronize  bool
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
	synchronize  bool
	arn          *string
}

func (c *client) Send(result payload.Payload) {
	c.BatchSend(nil, []payload.Payload{result})
}

func (c *client) BatchSend(polr v1alpha2.ReportInterface, results []payload.Payload) {
	var (
		accountID  *string
		newResults []payload.Payload
	)

	if c.accountID != "" {
		accountID = toPointer(c.accountID)
	}

	scConf := scutils.SecurityHubConfig{
		AccountID:   *accountID,
		ProductName: c.productName,
		CompanyName: c.companyName,
		ProductARN:  *c.arn,
		Region:      c.region,
	}

	// go over the results, add custom fields to them and append them to the findings if they are not nil
	fs := []types.AwsSecurityFinding{}
	for _, r := range results {
		if len(c.customFields) > 0 {
			if err := r.AddCustomFields(c.customFields); err != nil {
				zap.L().Error(c.Name()+": Error adding custom fields", zap.Error(err))
				return
			}
		}
		f, err := r.ToSecurityHubFindings(scConf)
		if err != nil {
			zap.L().Error(c.Name()+": Skipping result: ", zap.Any("resultID", r.GetID()), zap.Error(err))
			continue
		}

		fs = append(fs, *f)
	}

	// transform the findings to filters
	filters := scutils.ToResourceIDFilter(fs)

	// fetch findings from SH with the filters
	list, err := c.getFindingsByIDs(context.Background(), filters, "")
	if err != nil {
		zap.L().Error(c.Name()+": failed to get findings", zap.Error(err))
		return
	}

	findings := helper.Map(list, func(f types.AwsSecurityFinding) types.AwsSecurityFindingIdentifier {
		return types.AwsSecurityFindingIdentifier{
			Id:         f.Id,
			ProductArn: f.ProductArn,
		}
	})

	// create a new variable with all the results in case they are all new
	newResults = results

	// update the existing findings and get the ones remaining that were not there before
	if len(findings) > 0 {
		updated, err := c.batchUpdate(context.Background(), findings, types.WorkflowStatusNew)
		if err != nil {
			zap.L().Error(c.Name()+": PUSH FAILED", zap.Error(err))
			return
		} else if updated > 0 {
			zap.L().Info(c.Name()+": PUSH OK", zap.Int("updated", updated))
		}

		// build a map of the existing findings, we iterate over the list variable which contains the results returned from the SH query
		mapping := make(map[string]bool, len(list))
		for _, f := range list {
			mapping[*f.Id] = true
		}

		// get the payloads that were not included in the updated list and put them in an array of payload
		newResults = helper.Filter(results, func(result payload.Payload) bool {
			return !mapping[result.GetID()]
		})
	}

	if len(newResults) == 0 {
		return
	}

	newfindings := []types.AwsSecurityFinding{}
	// no need to check for nil here since we are sure there is a value because we already skipped nil ones
	for _, r := range newResults {
		f, err := r.ToSecurityHubFindings(scConf)
		if err != nil {
			zap.L().Error(c.Name()+": Skipping result: ", zap.Any("resultID", r.GetID()), zap.Error(err))
			continue
		}

		newfindings = append(newfindings, *f)
	}

	// import new findings
	res, err := c.hub.BatchImportFindings(context.Background(), &hub.BatchImportFindingsInput{
		Findings: newfindings,
	})
	if err != nil {
		zap.L().Error(c.Name()+": PUSH FAILED", zap.Error(err), zap.Any("response", res))
		return
	}

	zap.L().Info(c.Name()+": PUSH OK", zap.Int32("imported", *res.SuccessCount), zap.Int32("failed", *res.FailedCount))
}

func (c *client) Reset(ctx context.Context) error {
	if !c.synchronize {
		return nil
	}

	zap.L().Info(c.Name() + ": START SYNC")

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
	if !c.synchronize {
		return
	}

	zap.L().Info(c.Name()+": start cleanup", zap.String("report", report.GetKey()))

	if report.GetSource() != "" {
		if !c.BaseClient.ValidateReport(report) {
			return
		}
	}

	resourceIds := toResourceIDFilter(report, report.GetResults())

	findings, err := c.getFindingsByIDs(ctx, resourceIds, "")
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
		zap.L().Error(c.Name()+": failed to batch resolve findings", zap.Error(err))
		return
	}

	zap.L().Info(c.Name()+": CLEANUP OK", zap.Int("count", count), zap.String("report", report.GetKey()))
}

func (c *client) getFindings(ctx context.Context) ([]types.AwsSecurityFinding, error) {
	list := make([]types.AwsSecurityFinding, 0)

	var token *string

	for {
		resp, err := c.hub.GetFindings(ctx, &hub.GetFindingsInput{
			NextToken: token,
			Filters:   c.BaseFilter(),
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

func (c *client) getFindingsByIDs(ctx context.Context, resources []types.StringFilter, status string) ([]types.AwsSecurityFinding, error) {
	list := make([]types.AwsSecurityFinding, 0)

	chunks := helper.ChunkSlice(resources, 20)

	for _, res := range chunks {
		filter := c.BaseFilter()
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

func (c *client) BaseFilter() *types.AwsSecurityFindingFilters {
	// references to the report were removed here
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

	return filter
}

func (c *client) Type() target.ClientType {
	if !c.synchronize {
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
		options.Synchronize,
		toPointer("arn:aws:securityhub:" + options.Region + ":" + options.AccountID + ":product/" + options.AccountID + "/default"),
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
