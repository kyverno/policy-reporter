package listener_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"openreports.io/apis/openreports.io/v1alpha1"

	"github.com/kyverno/policy-reporter/pkg/cache"
	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/fixtures"
	"github.com/kyverno/policy-reporter/pkg/listener"
	"github.com/kyverno/policy-reporter/pkg/report"
)

func Test_ResultListener(t *testing.T) {
	t.Run("Publish Result", func(t *testing.T) {
		var called v1alpha1.ReportResult

		slistener := listener.NewResultListener(true, cache.NewInMermoryCache(time.Minute, time.Minute), time.Now())
		slistener.RegisterListener(func(_ v1alpha1.ReportInterface, r v1alpha1.ReportResult, b bool) {
			called = r
		})

		slistener.Listen(report.LifecycleEvent{Type: report.Added, PolicyReport: preport1})
		slistener.Listen(report.LifecycleEvent{Type: report.Updated, PolicyReport: preport2})

		assert.Equal(t, called.GetID(), fixtures.FailPodResult.GetID(), "Expected Listener to be called with FailPodResult")
	})

	t.Run("Ignore Delete Event", func(t *testing.T) {
		var called bool

		slistener := listener.NewResultListener(true, cache.NewInMermoryCache(time.Minute, time.Minute), time.Now())
		slistener.RegisterListener(func(_ v1alpha1.ReportInterface, r v1alpha1.ReportResult, b bool) {
			called = true
		})

		slistener.Listen(report.LifecycleEvent{Type: report.Deleted, PolicyReport: preport2})

		assert.False(t, called, "Expected Listener not be called on Deleted event")
	})

	t.Run("Ignore Added Results created before startup", func(t *testing.T) {
		var called bool

		slistener := listener.NewResultListener(true, cache.NewInMermoryCache(time.Minute, time.Minute), time.Now())
		slistener.RegisterListener(func(_ v1alpha1.ReportInterface, r v1alpha1.ReportResult, b bool) {
			called = true
		})

		slistener.Listen(report.LifecycleEvent{Type: report.Added, PolicyReport: preport2})

		assert.False(t, called, "Expected Listener not be called on Deleted event")
	})

	t.Run("Ignore CacheResults", func(t *testing.T) {
		var called bool

		slistener := listener.NewResultListener(true, cache.NewInMermoryCache(time.Minute, time.Minute), time.Now())
		slistener.RegisterListener(func(_ v1alpha1.ReportInterface, r v1alpha1.ReportResult, b bool) {
			called = true
		})

		slistener.Listen(report.LifecycleEvent{Type: report.Added, PolicyReport: preport2})
		slistener.Listen(report.LifecycleEvent{Type: report.Updated, PolicyReport: preport2})

		assert.False(t, called, "Expected Listener not be called on cached results")
	})

	t.Run("Early Return if Results are empty", func(t *testing.T) {
		var called bool

		slistener := listener.NewResultListener(true, cache.NewInMermoryCache(time.Minute, time.Minute), time.Now())
		slistener.RegisterListener(func(_ v1alpha1.ReportInterface, r v1alpha1.ReportResult, b bool) {
			called = true
		})

		slistener.Listen(report.LifecycleEvent{Type: report.Updated, PolicyReport: preport3})

		assert.False(t, called, "Expected Listener not be called with empty results")
	})

	t.Run("Skip process events when no listeners registered", func(t *testing.T) {
		c := cache.NewInMermoryCache(time.Minute, time.Minute)

		slistener := listener.NewResultListener(true, c, time.Now())
		slistener.Listen(report.LifecycleEvent{Type: report.Added, PolicyReport: preport2})

		assert.Greater(t, len(c.GetResults(preport2.GetID())), 0, "Expected cached report was found")
	})

	t.Run("UnregisterListener removes all listeners", func(t *testing.T) {
		var called bool

		slistener := listener.NewResultListener(true, cache.NewInMermoryCache(time.Minute, time.Minute), time.Now())
		slistener.RegisterListener(func(_ v1alpha1.ReportInterface, r v1alpha1.ReportResult, b bool) {
			called = true
		})

		slistener.UnregisterListener()

		slistener.Listen(report.LifecycleEvent{Type: report.Updated, PolicyReport: preport2})

		assert.False(t, called, "Expected Listener not called because it was unregistered")
	})
	t.Run("ignore results with past timestamps", func(t *testing.T) {
		var called bool

		slistener := listener.NewResultListener(true, cache.NewInMermoryCache(time.Minute, time.Minute), time.Now())
		slistener.RegisterListener(func(_ v1alpha1.ReportInterface, r v1alpha1.ReportResult, b bool) {
			called = true
		})

		rep := &v1alpha1.Report{
			Results: make([]v1alpha1.ReportResult, 0),
		}
		rep.Results = append(rep.Results, v1alpha1.ReportResult{
			Result:    v1alpha2.StatusFail,
			Timestamp: v1.Timestamp{Seconds: time.Now().Add(-24 * time.Hour).Unix()},
		})

		slistener.Listen(report.LifecycleEvent{Type: report.Updated, PolicyReport: rep})

		assert.False(t, called, "Expected Listener not called because it was unregistered")
	})

	t.Run("Publish Scoped Report", func(t *testing.T) {
		var called []v1alpha1.ReportResult

		slistener := listener.NewResultListener(true, cache.NewInMermoryCache(time.Minute, time.Minute), time.Now())
		slistener.RegisterScopeListener(func(_ v1alpha1.ReportInterface, r []v1alpha1.ReportResult, b bool) {
			called = r
		})

		slistener.Listen(report.LifecycleEvent{Type: report.Added, PolicyReport: scopereport1})

		assert.Equal(t, called[0].GetID(), fixtures.FailResult.GetID(), "Expected Listener to be called")
	})

	t.Run("Unregister Scope Listener", func(t *testing.T) {
		var called []v1alpha1.ReportResult

		slistener := listener.NewResultListener(true, cache.NewInMermoryCache(time.Minute, time.Minute), time.Now())
		slistener.RegisterScopeListener(func(_ v1alpha1.ReportInterface, r []v1alpha1.ReportResult, b bool) {
			called = r
		})

		slistener.UnregisterScopeListener()

		slistener.Listen(report.LifecycleEvent{Type: report.Added, PolicyReport: scopereport1})

		assert.Len(t, called, 0, "Expected listener was unregistered")
	})

	t.Run("Publish Scoped Report to Sync Target", func(t *testing.T) {
		var called v1alpha1.ReportInterface

		slistener := listener.NewResultListener(true, cache.NewInMermoryCache(time.Minute, time.Minute), time.Now())
		slistener.RegisterSyncListener(func(r v1alpha1.ReportInterface) {
			called = r
		})

		slistener.Listen(report.LifecycleEvent{Type: report.Added, PolicyReport: scopereport1})

		assert.Equal(t, called.GetName(), scopereport1.Name, "Expected Listener to be called")
	})

	t.Run("Publish Scoped Report to Sync Target", func(t *testing.T) {
		var called v1alpha1.ReportInterface

		slistener := listener.NewResultListener(true, cache.NewInMermoryCache(time.Minute, time.Minute), time.Now())
		slistener.RegisterSyncListener(func(r v1alpha1.ReportInterface) {
			called = r
		})

		slistener.UnregisterSyncListener()

		slistener.Listen(report.LifecycleEvent{Type: report.Added, PolicyReport: scopereport1})

		assert.Nil(t, called, "Expected Listener was unregistered")
	})

	t.Run("Check Validation Logic", func(t *testing.T) {
		slistener := listener.NewResultListener(true, cache.NewInMermoryCache(time.Minute, time.Minute), time.Now())

		assert.True(t, slistener.Validate(fixtures.FailPodResult))
		assert.True(t, slistener.Validate(fixtures.WarnPodResult))
		assert.True(t, slistener.Validate(fixtures.ErrorPodResult))
		assert.False(t, slistener.Validate(fixtures.PassPodResult))
		assert.False(t, slistener.Validate(fixtures.SkipPodResult))
	})
}
