package listener_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kyverno/policy-reporter/pkg/fixtures"
	"github.com/kyverno/policy-reporter/pkg/listener"
	"github.com/kyverno/policy-reporter/pkg/openreports"
	"github.com/kyverno/policy-reporter/pkg/target"
)

func Test_ScopeResultsListener(t *testing.T) {
	t.Run("Send Results", func(t *testing.T) {
		c := &client{validated: true, batchSend: true}
		slistener := listener.NewSendScopeResultsListener(target.NewCollection(&target.Target{Client: c}))
		slistener(preport1, []openreports.ResultAdapter{fixtures.FailResult}, false)

		assert.True(t, c.Called, "Expected Send to be called")
	})
	t.Run("Don't Send Result when validation fails", func(t *testing.T) {
		c := &client{validated: false, batchSend: true}
		slistener := listener.NewSendScopeResultsListener(target.NewCollection(&target.Target{Client: c}))
		slistener(preport1, []openreports.ResultAdapter{fixtures.FailResult}, false)

		assert.False(t, c.Called, "Expected Send not to be called")
	})
	t.Run("Don't Send pre existing Result when skipExistingOnStartup is true", func(t *testing.T) {
		c := &client{skipExistingOnStartup: true, batchSend: true}
		slistener := listener.NewSendScopeResultsListener(target.NewCollection(&target.Target{Client: c}))
		slistener(preport1, []openreports.ResultAdapter{fixtures.FailResult}, true)

		if c.Called {
			t.Error("Expected Send not to be called")
		}
	})
}
