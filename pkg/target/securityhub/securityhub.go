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
}

type client struct {
	target.BaseClient
	customFields map[string]string
	hub          *hub.Client
	accountID    string
	region       string
	productName  string
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
				Id:            &result.ID,
				AwsAccountId:  accID,
				SchemaVersion: toPointer("2018-10-08"),
				ProductArn:    toPointer("arn:aws:securityhub:" + c.region + ":" + c.accountID + ":product/" + c.accountID + "/default"),
				GeneratorId:   toPointer(fmt.Sprintf("%s/%s", result.Source, generator)),
				Types:         []string{"Software and Configuration Checks"},
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
		if res.Kind != "" {
			return toPointer(string(res.UID))
		}

		return toPointer(result.ResourceString())
	}

	return &result.ID
}
