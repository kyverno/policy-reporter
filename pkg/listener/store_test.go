package listener_test

import (
	"testing"

	"github.com/kyverno/policy-reporter/pkg/listener"
	"github.com/kyverno/policy-reporter/pkg/report"
)

func Test_StoreListener(t *testing.T) {
	store := report.NewPolicyReportStore()

	t.Run("Save New Report", func(t *testing.T) {
		slistener := listener.NewStoreListener(store)
		slistener(report.LifecycleEvent{Type: report.Added, PolicyReport: preport1})

		if _, ok := store.Get(preport1.GetID()); !ok {
			t.Error("Expected Report to be stored")
		}
	})
	t.Run("Update Modified Report", func(t *testing.T) {
		slistener := listener.NewStoreListener(store)
		slistener(report.LifecycleEvent{Type: report.Updated, PolicyReport: preport2})

		if preport, ok := store.Get(preport2.GetID()); !ok && len(preport.GetResults()) == 2 {
			t.Error("Expected Report to be updated")
		}
	})
	t.Run("Remove Deleted Report", func(t *testing.T) {
		slistener := listener.NewStoreListener(store)
		slistener(report.LifecycleEvent{Type: report.Deleted, PolicyReport: preport2})

		if _, ok := store.Get(preport2.GetID()); ok {
			t.Error("Expected Report to be removed")
		}
	})
}
