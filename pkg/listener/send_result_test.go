package listener_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kyverno/policy-reporter/pkg/fixtures"
	"github.com/kyverno/policy-reporter/pkg/listener"
	"github.com/kyverno/policy-reporter/pkg/openreports"
	"github.com/kyverno/policy-reporter/pkg/target"
)

type client struct {
	Called                bool
	skipExistingOnStartup bool
	validated             bool
	cleanupCalled         bool
	batchSend             bool
	cleanup               bool
}

func (c *client) Send(result *openreports.ORResultAdapter) {
	c.Called = true
}

func (c *client) MinimumSeverity() string {
	return openreports.SeverityInfo
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

func (c client) Validate(rep openreports.ReportInterface, result *openreports.ORResultAdapter) bool {
	return c.validated
}

func (c *client) Reset(_ context.Context) error {
	return nil
}

func (c *client) SendHeartbeat() {}

func (c *client) CleanUp(_ context.Context, _ openreports.ReportInterface) {
	c.cleanupCalled = true
}

func (c *client) BatchSend(_ openreports.ReportInterface, _ []*openreports.ORResultAdapter) {
	c.Called = true
}

func (c *client) Type() target.ClientType {
	if c.cleanup {
		return target.SyncSend
	}
	if c.batchSend {
		return target.BatchSend
	}

	return target.SingleSend
}

func Test_SendResultListener(t *testing.T) {
	t.Run("Send Result", func(t *testing.T) {
		c := &client{validated: true}
		slistener := listener.NewSendResultListener(target.NewCollection(&target.Target{Client: c}))
		slistener(&openreports.ORReportAdapter{Report: preport1}, fixtures.FailResult, false)

		assert.True(t, c.Called, "Expected Send to be called")
	})
	t.Run("Don't Send Result when validation fails", func(t *testing.T) {
		c := &client{validated: false}
		slistener := listener.NewSendResultListener(target.NewCollection(&target.Target{Client: c}))
		slistener(&openreports.ORReportAdapter{Report: preport1}, fixtures.FailResult, false)

		assert.False(t, c.Called, "Expected Send not to be called")
	})
	t.Run("Don't Send pre existing Result when skipExistingOnStartup is true", func(t *testing.T) {
		c := &client{skipExistingOnStartup: true}
		slistener := listener.NewSendResultListener(target.NewCollection(&target.Target{Client: c}))
		slistener(&openreports.ORReportAdapter{Report: preport1}, fixtures.FailResult, true)

		assert.False(t, c.Called, "Expected Send not to be called")
	})
}
