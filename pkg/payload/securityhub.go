package payload

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/securityhub/types"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/payload/scutils"
)

var schema = toPointer("2018-10-08")

func toPointer[T any](v T) *T {
	return &v
}

func (p *PolicyReportResultPayload) ToSecurityHubFindings(scConf scutils.SecurityHubConfig) (*types.AwsSecurityFinding, error) {
	if !shouldSendresult(p.Result) {
		return nil, fmt.Errorf("invalid result to send")
	}

	generator := p.Result.Policy
	if generator == "" {
		generator = p.Result.Rule
	}

	title := generator
	if p.Result.HasResource() {
		title = fmt.Sprintf("%s: %s", title, p.Result.ResourceString())
	}

	t := time.Unix(p.Result.Timestamp.Seconds, int64(p.Result.Timestamp.Nanos))

	return &types.AwsSecurityFinding{
		Id:            toPointer(p.Result.GetID()),
		AwsAccountId:  &scConf.AccountID,
		SchemaVersion: schema,
		ProductArn:    &scConf.ProductARN,
		GeneratorId:   toPointer(fmt.Sprintf("%s/%s", p.Result.Source, generator)),
		Types:         []string{mapType(p.Result.Source)},
		CreatedAt:     toPointer(t.Format("2006-01-02T15:04:05.999999999Z07:00")),
		UpdatedAt:     toPointer(t.Format("2006-01-02T15:04:05.999999999Z07:00")),
		Severity: &types.Severity{
			Label: mapSeverity(p.Result.Severity),
		},
		Title:       &title,
		Description: &p.Result.Message,
		ProductName: &scConf.ProductName,
		CompanyName: &scConf.CompanyName,
		Compliance: &types.Compliance{
			Status: types.ComplianceStatusFailed,
		},
		Workflow: &types.Workflow{
			Status: types.WorkflowStatusNew,
		},
		Resources: []types.Resource{
			{
				Type:      toPointer("Other"),
				Region:    &scConf.Region,
				Partition: types.PartitionAws,
				Id:        mapResourceID(p.Result),
				Details: &types.ResourceDetails{
					Other: p.mapOtherDetails(),
				},
			},
		},
		RecordState: types.RecordStateActive,
	}, nil
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

func mapType(source string) string {
	if source == "" {
		return "Software and Configuration Checks/Kubernetes Policies"
	}
	return "Software and Configuration Checks/Kubernetes Policies/" + source
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

func (p *PolicyReportResultPayload) mapOtherDetails() map[string]string {
	details := map[string]string{
		"Source":   p.Result.Source,
		"Category": p.Result.Category,
		"Policy":   p.Result.Policy,
		"Rule":     p.Result.Rule,
		"Result":   string(p.Result.Result),
	}

	if len(p.Result.Properties) > 0 {
		for property, value := range p.Result.Properties {
			details[property] = value
		}
	}

	if p.Result.HasResource() {
		res := p.Result.GetResource()

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

func shouldSendresult(r v1alpha2.PolicyReportResult) bool {
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
}
