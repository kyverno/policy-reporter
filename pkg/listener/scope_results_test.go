package listener_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/kyverno/policy-reporter/pkg/cache"
	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/fixtures"
	"github.com/kyverno/policy-reporter/pkg/listener"
	"github.com/kyverno/policy-reporter/pkg/target"
)

func Test_ScopeResultsListener(t *testing.T) {
	t.Run("Send Results", func(t *testing.T) {
		c := &client{validated: true, batchSend: true}
		resultCache := cache.NewInMermoryCache(time.Hour, time.Hour)
		c.SetCache(resultCache)

		slistener := listener.NewSendScopeResultsListener(target.NewCollection(&target.Target{Client: c}))
		slistener(preport1, []v1alpha2.PolicyReportResult{fixtures.FailResult})

		assert.True(t, c.Called, "Expected Send to be called")
	})
	t.Run("Don't Send Result when validation fails", func(t *testing.T) {
		c := &client{validated: false, batchSend: true}
		resultCache := cache.NewInMermoryCache(time.Hour, time.Hour)
		c.SetCache(resultCache)

		slistener := listener.NewSendScopeResultsListener(target.NewCollection(&target.Target{Client: c}))
		slistener(preport1, []v1alpha2.PolicyReportResult{fixtures.FailResult})

		assert.False(t, c.Called, "Expected Send not to be called")
	})
	t.Run("Don't Send pre existing Result when skipExistingOnStartup is true", func(t *testing.T) {
		c := &client{skipExistingOnStartup: true, batchSend: true}
		resultCache := cache.NewInMermoryCache(time.Hour, time.Hour)
		resultCache.AddReport(preport1)
		c.SetCache(resultCache)

		slistener := listener.NewSendScopeResultsListener(target.NewCollection(&target.Target{Client: c}))
		slistener(preport1, []v1alpha2.PolicyReportResult{fixtures.FailResult})

		if c.Called {
			t.Error("Expected Send not to be called")
		}
	})
}
