package securityhub

import (
	"fmt"
	"time"

	"go.uber.org/zap"

	hub "github.com/aws/aws-sdk-go/service/securityhub"
	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/target"
)

// Options to configure the S3 target
type Options struct {
	target.ClientOptions
	CustomFields map[string]string
	Client       *hub.SecurityHub
	AccountID    string
	Region       string
}

type client struct {
	target.BaseClient
	customFields map[string]string
	hub          *hub.SecurityHub
	accountID    string
	region       string
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

	res, err := c.hub.BatchImportFindings(&hub.BatchImportFindingsInput{
		Findings: []*hub.AwsSecurityFinding{
			{
				Id:            &result.ID,
				AwsAccountId:  &c.accountID,
				SchemaVersion: toPointer("2018-10-08"),
				ProductArn:    toPointer("arn:aws:securityhub:" + c.region + ":" + c.accountID + ":product/" + c.accountID + "/default"),
				GeneratorId:   toPointer(fmt.Sprintf("%s/%s", result.Source, generator)),
				Types:         []*string{toPointer("Software and Configuration Checks")},
				CreatedAt:     toPointer(t.Format("2006-01-02T15:04:05.999999999Z07:00")),
				UpdatedAt:     toPointer(t.Format("2006-01-02T15:04:05.999999999Z07:00")),
				Severity: &hub.Severity{
					Label: mapSeverity(result.Severity),
				},
				Title:       &title,
				Description: &result.Message,
				ProductFields: map[string]*string{
					"Product Name": toPointer("Policy Reporter"),
				},
				Resources: []*hub.Resource{
					{
						Type:      toPointer("Other"),
						Region:    &c.region,
						Partition: toPointer("aws"),
						Id:        mapResourceID(result),
						Details: &hub.ResourceDetails{
							Other: c.mapOtherDetails(result),
						},
					},
				},
				RecordState: toPointer(hub.RecordStateActive),
			},
		},
	})
	if err != nil {
		zap.L().Error(c.Name()+": PUSH FAILED", zap.Error(err), zap.Any("response", res))
		return
	}

	zap.L().Info(c.Name()+": PUSH OK", zap.Int64("successCount", *res.SuccessCount), zap.Int64("failedCount", *res.FailedCount))
}

func (c *client) mapOtherDetails(result v1alpha2.PolicyReportResult) map[string]*string {
	details := map[string]*string{
		"Source":   &result.Source,
		"Category": &result.Category,
		"Policy":   &result.Policy,
		"Rule":     &result.Rule,
		"Result":   toPointer(string(result.Result)),
		"Priority": toPointer(result.Priority.String()),
	}

	if len(c.customFields) > 0 {
		for property, value := range c.customFields {
			details[property] = &value
		}

		for property, value := range result.Properties {
			details[property] = &value
		}
	}

	if result.HasResource() {
		res := result.GetResource()

		if res.APIVersion != "" {
			details["Resource APIVersion"] = &res.APIVersion
		}
		if res.Kind != "" {
			details["Resource Kind"] = &res.Kind
		}
		if res.Namespace != "" {
			details["Resource Namespace"] = &res.Namespace
		}
		if res.Name != "" {
			details["Resource Name"] = &res.Name
		}
		if res.UID != "" {
			details["Resource UID"] = toPointer(string(res.UID))
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
	}
}

func toPointer[T any](value T) *T {
	return &value
}

func mapSeverity(s v1alpha2.PolicySeverity) *string {
	switch s {
	case v1alpha2.SeverityInfo:
		return toPointer(hub.SeverityLabelInformational)
	case v1alpha2.SeverityLow:
		return toPointer(hub.SeverityLabelLow)
	case v1alpha2.SeverityMedium:
		return toPointer(hub.SeverityLabelMedium)
	case v1alpha2.SeverityHigh:
		return toPointer(hub.SeverityLabelHigh)
	case v1alpha2.SeverityCritical:
		return toPointer(hub.SeverityLabelCritical)
	default:
		return toPointer(hub.SeverityLabelInformational)
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
