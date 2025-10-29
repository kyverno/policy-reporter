package securityhub_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	hub "github.com/aws/aws-sdk-go-v2/service/securityhub"
	"github.com/aws/aws-sdk-go-v2/service/securityhub/types"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/fixtures"
	"github.com/kyverno/policy-reporter/pkg/target/securityhub"
)

type client struct {
	batched bool
	fetched bool

	send     func(findings []types.AwsSecurityFinding)
	findings []types.AwsSecurityFinding
}

func (c *client) BatchImportFindings(ctx context.Context, params *hub.BatchImportFindingsInput, optFns ...func(*hub.Options)) (*hub.BatchImportFindingsOutput, error) {
	c.batched = true

	if c.send != nil {
		c.send(params.Findings)
	}

	return &hub.BatchImportFindingsOutput{
		SuccessCount: aws.Int32(1),
		FailedCount:  aws.Int32(0),
	}, nil
}

func (c *client) GetFindings(ctx context.Context, params *hub.GetFindingsInput, optFns ...func(*hub.Options)) (*hub.GetFindingsOutput, error) {
	c.fetched = true
	return &hub.GetFindingsOutput{
		Findings: c.findings,
	}, nil
}

func (c *client) BatchUpdateFindings(ctx context.Context, params *hub.BatchUpdateFindingsInput, optFns ...func(*hub.Options)) (*hub.BatchUpdateFindingsOutput, error) {
	c.batched = true

	return &hub.BatchUpdateFindingsOutput{}, nil
}

func TestSecurityHub(t *testing.T) {
	t.Run("send result", func(t *testing.T) {
		c := securityhub.NewClient(securityhub.Options{
			AccountID:   "accountId",
			Region:      "eu-central-1",
			ProductName: "Policy Reporter",
			CompanyName: "Kyverno",
			Client: &client{
				send: func(findings []types.AwsSecurityFinding) {
					if len(findings) != 1 {
						t.Error("expected to get one finding")
						return
					}

					finding := findings[0]

					if *finding.AwsAccountId != "accountId" {
						t.Errorf("unexpected accountId: %s", *finding.AwsAccountId)
					}
					if *finding.Id != fixtures.CompleteTargetSendResult.GetID() {
						t.Errorf("unexpected id: %s", *finding.Id)
					}
					if *finding.ProductArn != "arn:aws:securityhub:eu-central-1:accountId:product/accountId/default" {
						t.Errorf("unexpected product arn: %s", *finding.ProductArn)
					}
					if *finding.ProductName != "Policy Reporter" {
						t.Errorf("unexpected product name: %s", *finding.ProductName)
					}
					if *finding.CompanyName != "Kyverno" {
						t.Errorf("unexpected company name: %s", *finding.CompanyName)
					}
				},
			},
		})

		c.Send(fixtures.DefaultPolicyReport, fixtures.CompleteTargetSendResult)
	})
	t.Run("clean up disabled", func(t *testing.T) {
		h := &client{}

		c := securityhub.NewClient(securityhub.Options{
			AccountID:   "accountId",
			Region:      "eu-central-1",
			ProductName: "Policy Reporter",
			CompanyName: "Kyverno",
			Client:      h,
			Synchronize: false,
		})

		c.CleanUp(context.TODO(), fixtures.DefaultPolicyReport)

		if h.fetched {
			t.Error("expected fetch was not called")
		}
		if h.batched {
			t.Error("expected batch was not called")
		}
	})
	t.Run("findings without results", func(t *testing.T) {
		h := &client{}

		c := securityhub.NewClient(securityhub.Options{
			AccountID:   "accountId",
			Region:      "eu-central-1",
			ProductName: "Policy Reporter",
			CompanyName: "Kyverno",
			Client:      h,
			Synchronize: true,
		})

		c.CleanUp(context.TODO(), fixtures.DefaultPolicyReport)

		if !h.fetched {
			t.Error("expected fetch was called")
		}
		if h.batched {
			t.Error("expected batch was not called")
		}
	})
	t.Run("findings with existing result", func(t *testing.T) {
		h := &client{
			findings: []types.AwsSecurityFinding{
				{
					Id: aws.String(fixtures.DefaultPolicyReport.GetResults()[0].GetID()),
				},
			},
		}

		c := securityhub.NewClient(securityhub.Options{
			AccountID:   "accountId",
			Region:      "eu-central-1",
			ProductName: "Policy Reporter",
			CompanyName: "Kyverno",
			Client:      h,
			Synchronize: true,
		})

		c.CleanUp(context.TODO(), fixtures.DefaultPolicyReport)

		if !h.fetched {
			t.Error("expected fetch was called")
		}
		if h.batched {
			t.Error("expected batch was not called")
		}
	})
	t.Run("findings with not existing result", func(t *testing.T) {
		h := &client{
			findings: []types.AwsSecurityFinding{
				{
					Id: aws.String("not-existing-result"),
				},
			},
		}

		c := securityhub.NewClient(securityhub.Options{
			AccountID:   "accountId",
			Region:      "eu-central-1",
			ProductName: "Policy Reporter",
			CompanyName: "Kyverno",
			Client:      h,
			Synchronize: true,
		})

		c.CleanUp(context.TODO(), fixtures.DefaultPolicyReport)

		if !h.fetched {
			t.Error("expected fetch was called")
		}
		if !h.batched {
			t.Error("expected batch was called")
		}
	})
	t.Run("MapSeverity", func(t *testing.T) {
		if securityhub.MapSeverity(v1alpha2.SeverityInfo) != types.SeverityLabelInformational {
			t.Error("unexpected severity mapping")
		}
		if securityhub.MapSeverity(v1alpha2.SeverityLow) != types.SeverityLabelLow {
			t.Error("unexpected severity mapping")
		}
		if securityhub.MapSeverity(v1alpha2.SeverityMedium) != types.SeverityLabelMedium {
			t.Error("unexpected severity mapping")
		}
		if securityhub.MapSeverity(v1alpha2.SeverityHigh) != types.SeverityLabelHigh {
			t.Error("unexpected severity mapping")
		}
		if securityhub.MapSeverity(v1alpha2.SeverityCritical) != types.SeverityLabelCritical {
			t.Error("unexpected severity mapping")
		}
		if securityhub.MapSeverity("") != types.SeverityLabelInformational {
			t.Error("unexpected severity mapping")
		}
	})
}
