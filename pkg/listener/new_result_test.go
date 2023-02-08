package listener_test

import (
	"testing"
	"time"

	"github.com/kyverno/policy-reporter/pkg/cache"
	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/fixtures"
	"github.com/kyverno/policy-reporter/pkg/listener"
	"github.com/kyverno/policy-reporter/pkg/report"
)

func Test_ResultListener(t *testing.T) {
	t.Run("Publish Result", func(t *testing.T) {
		var called v1alpha2.PolicyReportResult

		slistener := listener.NewResultListener(true, cache.New(), time.Now())
		slistener.RegisterListener(func(_ v1alpha2.ReportInterface, r v1alpha2.PolicyReportResult, b bool) {
			called = r
		})

		slistener.Listen(report.LifecycleEvent{Type: report.Added, PolicyReport: preport1})
		slistener.Listen(report.LifecycleEvent{Type: report.Updated, PolicyReport: preport2})

		if called.GetID() != fixtures.FailPodResult.GetID() {
			t.Error("Expected Listener to be called with FailPodResult")
		}
	})

	t.Run("Ignore Delete Event", func(t *testing.T) {
		var called bool

		slistener := listener.NewResultListener(true, cache.New(), time.Now())
		slistener.RegisterListener(func(_ v1alpha2.ReportInterface, r v1alpha2.PolicyReportResult, b bool) {
			called = true
		})

		slistener.Listen(report.LifecycleEvent{Type: report.Deleted, PolicyReport: preport2})

		if called {
			t.Error("Expected Listener not be called on Deleted event")
		}
	})

	t.Run("Ignore Added Results created before startup", func(t *testing.T) {
		var called bool

		slistener := listener.NewResultListener(true, cache.New(), time.Now())
		slistener.RegisterListener(func(_ v1alpha2.ReportInterface, r v1alpha2.PolicyReportResult, b bool) {
			called = true
		})

		slistener.Listen(report.LifecycleEvent{Type: report.Added, PolicyReport: preport2})

		if called {
			t.Error("Expected Listener not be called on Deleted event")
		}
	})

	t.Run("Ignore CacheResults", func(t *testing.T) {
		var called bool

		rcache := cache.New()

		slistener := listener.NewResultListener(true, rcache, time.Now())
		slistener.RegisterListener(func(_ v1alpha2.ReportInterface, r v1alpha2.PolicyReportResult, b bool) {
			called = true
		})

		slistener.Listen(report.LifecycleEvent{Type: report.Added, PolicyReport: preport2})
		slistener.Listen(report.LifecycleEvent{Type: report.Updated, PolicyReport: preport2})

		if called {
			t.Error("Expected Listener not be called on cached results")
		}
	})

	t.Run("Early Return if Results are empty", func(t *testing.T) {
		var called bool

		rcache := cache.New()

		slistener := listener.NewResultListener(true, rcache, time.Now())
		slistener.RegisterListener(func(_ v1alpha2.ReportInterface, r v1alpha2.PolicyReportResult, b bool) {
			called = true
		})

		slistener.Listen(report.LifecycleEvent{Type: report.Updated, PolicyReport: preport3})

		if called {
			t.Error("Expected Listener not be called with empty results")
		}
	})
}
