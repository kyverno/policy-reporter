package listener_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/kyverno/policy-reporter/pkg/cache"
	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/fixtures"
	"github.com/kyverno/policy-reporter/pkg/listener"
	"github.com/kyverno/policy-reporter/pkg/report"
)

func Test_ResultListener(t *testing.T) {
	t.Run("Publish Result", func(t *testing.T) {
		var called v1alpha2.PolicyReportResult

		slistener := listener.NewResultListener(cache.NewInMermoryCache(time.Minute, time.Minute), time.Now())
		slistener.RegisterListener(func(_ v1alpha2.ReportInterface, r v1alpha2.PolicyReportResult) {
			called = r
		})

		slistener.Listen(report.LifecycleEvent{Type: report.Added, PolicyReport: preport1})
		slistener.Listen(report.LifecycleEvent{Type: report.Updated, PolicyReport: preport2})

		assert.Equal(t, called.GetID(), fixtures.FailPodResult.GetID(), "Expected Listener to be called with FailPodResult")
	})

	t.Run("Ignore Delete Event", func(t *testing.T) {
		var called bool

		slistener := listener.NewResultListener(cache.NewInMermoryCache(time.Minute, time.Minute), time.Now())
		slistener.RegisterListener(func(_ v1alpha2.ReportInterface, r v1alpha2.PolicyReportResult) {
			called = true
		})

		slistener.Listen(report.LifecycleEvent{Type: report.Deleted, PolicyReport: preport2})

		assert.False(t, called, "Expected Listener not be called on Deleted event")
	})

	t.Run("Early Return if Results are empty", func(t *testing.T) {
		var called bool

		slistener := listener.NewResultListener(cache.NewInMermoryCache(time.Minute, time.Minute), time.Now())
		slistener.RegisterListener(func(_ v1alpha2.ReportInterface, r v1alpha2.PolicyReportResult) {
			called = true
		})

		slistener.Listen(report.LifecycleEvent{Type: report.Updated, PolicyReport: preport3})

		assert.False(t, called, "Expected Listener not be called with empty results")
	})

	t.Run("Skip process events when no listeners registered", func(t *testing.T) {
		c := cache.NewInMermoryCache(time.Minute, time.Minute)

		slistener := listener.NewResultListener(c, time.Now())
		slistener.Listen(report.LifecycleEvent{Type: report.Added, PolicyReport: preport2})

		assert.Greater(t, len(c.GetResults(preport2.GetID())), 0, "Expected cached report was found")
	})

	t.Run("UnregisterListener removes all listeners", func(t *testing.T) {
		var called bool

		slistener := listener.NewResultListener(cache.NewInMermoryCache(time.Minute, time.Minute), time.Now())
		slistener.RegisterListener(func(_ v1alpha2.ReportInterface, r v1alpha2.PolicyReportResult) {
			called = true
		})

		slistener.UnregisterListener()

		slistener.Listen(report.LifecycleEvent{Type: report.Updated, PolicyReport: preport2})

		assert.False(t, called, "Expected Listener not called because it was unregistered")
	})
}
