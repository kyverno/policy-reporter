package listener_test

import (
	"context"
	"testing"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/fixtures"
	"github.com/kyverno/policy-reporter/pkg/listener"
	"github.com/kyverno/policy-reporter/pkg/target"
)

type client struct {
	Called                bool
	skipExistingOnStartup bool
	validated             bool
	cleanupCalled         bool
	batchSendCalled       bool
}

func (c *client) Send(result v1alpha2.PolicyReportResult) {
	c.Called = true
}

func (c *client) MinimumPriority() string {
	return v1alpha2.InfoPriority.String()
}

func (c *client) Name() string {
	return "test"
}

func (c *client) Sources() []string {
	return []string{}
}

func (c *client) SkipExistingOnStartup() bool {
	return c.skipExistingOnStartup
}

func (c client) Validate(rep v1alpha2.ReportInterface, result v1alpha2.PolicyReportResult) bool {
	return c.validated
}

func (c *client) CleanUp(_ context.Context, _ v1alpha2.ReportInterface) {
	c.cleanupCalled = true
}

func (c *client) BatchSend(_ v1alpha2.ReportInterface, _ []v1alpha2.PolicyReportResult) {
	c.batchSendCalled = true
}

func (c *client) SupportsBatchSend() bool {
	return false
}

func Test_SendResultListener(t *testing.T) {
	t.Run("Send Result", func(t *testing.T) {
		c := &client{validated: true}
		slistener := listener.NewSendResultListener(target.NewCollection(&target.Target{Client: c}))
		slistener(preport1, fixtures.FailResult, false)

		if !c.Called {
			t.Error("Expected Send to be called")
		}
	})
	t.Run("Don't Send Result when validation fails", func(t *testing.T) {
		c := &client{validated: false}
		slistener := listener.NewSendResultListener(target.NewCollection(&target.Target{Client: c}))
		slistener(preport1, fixtures.FailResult, false)

		if c.Called {
			t.Error("Expected Send not to be called")
		}
	})
	t.Run("Don't Send pre existing Result when skipExistingOnStartup is true", func(t *testing.T) {
		c := &client{skipExistingOnStartup: true}
		slistener := listener.NewSendResultListener(target.NewCollection(&target.Target{Client: c}))
		slistener(preport1, fixtures.FailResult, true)

		if c.Called {
			t.Error("Expected Send not to be called")
		}
	})
}
