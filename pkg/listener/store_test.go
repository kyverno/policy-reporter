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
		slistener(report.LifecycleEvent{Type: report.Added, NewPolicyReport: preport1, OldPolicyReport: &report.PolicyReport{}})

		if _, ok := store.Get(preport1.ID); !ok {
			t.Error("Expected Report to be stored")
		}
	})
	t.Run("Update Modified Report", func(t *testing.T) {
		slistener := listener.NewStoreListener(store)
		slistener(report.LifecycleEvent{Type: report.Updated, NewPolicyReport: preport2, OldPolicyReport: preport1})

		if preport, ok := store.Get(preport2.ID); !ok && len(preport.Results) == 2 {
			t.Error("Expected Report to be updated")
		}
	})
	t.Run("Remove Deleted Report", func(t *testing.T) {
		slistener := listener.NewStoreListener(store)
		slistener(report.LifecycleEvent{Type: report.Deleted, NewPolicyReport: preport2, OldPolicyReport: &report.PolicyReport{}})

		if _, ok := store.Get(preport2.ID); ok {
			t.Error("Expected Report to be removed")
		}
	})
}
