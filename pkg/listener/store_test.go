package listener_test

import (
	"context"
	"testing"

	"github.com/kyverno/policy-reporter/pkg/listener"
	"github.com/kyverno/policy-reporter/pkg/report"
)

var ctx = context.Background()

func Test_StoreListener(t *testing.T) {
	store := report.NewPolicyReportStore()

	t.Run("Save New Report", func(t *testing.T) {
		slistener := listener.NewStoreListener(ctx, store)
		slistener(report.LifecycleEvent{Type: report.Added, PolicyReport: preport1})

		if _, err := store.Get(ctx, preport1.GetID()); err != nil {
			t.Error("Expected Report to be stored")
		}
	})
	t.Run("Update Modified Report", func(t *testing.T) {
		slistener := listener.NewStoreListener(ctx, store)
		slistener(report.LifecycleEvent{Type: report.Updated, PolicyReport: preport2})

		if preport, err := store.Get(ctx, preport2.GetID()); err != nil && len(preport.GetResults()) == 2 {
			t.Error("Expected Report to be updated")
		}
	})
	t.Run("Remove Deleted Report", func(t *testing.T) {
		slistener := listener.NewStoreListener(ctx, store)
		slistener(report.LifecycleEvent{Type: report.Deleted, PolicyReport: preport2})

		if _, err := store.Get(ctx, preport2.GetID()); err == nil {
			t.Error("Expected Report to be removed")
		}
	})
}
