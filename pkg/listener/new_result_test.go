package listener_test

import (
	"testing"
	"time"

	"github.com/kyverno/policy-reporter/pkg/listener"
	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/patrickmn/go-cache"
)

func Test_ResultListener(t *testing.T) {
	t.Run("Publish Result", func(t *testing.T) {
		var called *report.Result

		slistener := listener.NewResultListener(true, cache.New(cache.DefaultExpiration, 5*time.Minute), time.Now())
		slistener.RegisterListener(func(r *report.Result, b bool) {
			called = r
		})

		slistener.Listen(report.LifecycleEvent{Type: report.Updated, NewPolicyReport: preport2, OldPolicyReport: preport1})

		if called.GetIdentifier() != result2.GetIdentifier() {
			t.Error("Expected Listener to be called with Result2")
		}
	})

	t.Run("Ignore Delete Event", func(t *testing.T) {
		var called bool

		slistener := listener.NewResultListener(true, cache.New(cache.DefaultExpiration, 5*time.Minute), time.Now())
		slistener.RegisterListener(func(r *report.Result, b bool) {
			called = true
		})

		slistener.Listen(report.LifecycleEvent{Type: report.Deleted, NewPolicyReport: preport2, OldPolicyReport: preport1})

		if called {
			t.Error("Expected Listener not be called on Deleted event")
		}
	})

	t.Run("Ignore Added Results created before startup", func(t *testing.T) {
		var called bool

		slistener := listener.NewResultListener(true, cache.New(cache.DefaultExpiration, 5*time.Minute), time.Now())
		slistener.RegisterListener(func(r *report.Result, b bool) {
			called = true
		})

		slistener.Listen(report.LifecycleEvent{Type: report.Added, NewPolicyReport: preport2, OldPolicyReport: preport1})

		if called {
			t.Error("Expected Listener not be called on Deleted event")
		}
	})

	t.Run("Ignore CacheResults", func(t *testing.T) {
		var called bool

		rcache := cache.New(cache.DefaultExpiration, 5*time.Minute)
		rcache.SetDefault(result2.ID, true)

		slistener := listener.NewResultListener(true, rcache, time.Now())
		slistener.RegisterListener(func(r *report.Result, b bool) {
			called = true
		})

		slistener.Listen(report.LifecycleEvent{Type: report.Updated, NewPolicyReport: preport2, OldPolicyReport: preport1})

		if called {
			t.Error("Expected Listener not be called on cached results")
		}
	})

	t.Run("Early Return if Rsults are empty", func(t *testing.T) {
		var called bool

		rcache := cache.New(cache.DefaultExpiration, 5*time.Minute)
		rcache.SetDefault(result2.ID, true)

		slistener := listener.NewResultListener(true, rcache, time.Now())
		slistener.RegisterListener(func(r *report.Result, b bool) {
			called = true
		})

		slistener.Listen(report.LifecycleEvent{Type: report.Updated, NewPolicyReport: preport2, OldPolicyReport: preport1})

		if called {
			t.Error("Expected Listener not be called with empty results")
		}
	})
}
