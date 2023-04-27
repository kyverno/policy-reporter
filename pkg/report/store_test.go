package report_test

import (
	"context"
	"testing"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/report"
)

var ctx = context.Background()

func Test_PolicyReportStore(t *testing.T) {
	store := report.NewPolicyReportStore()
	store.CreateSchemas(ctx)

	t.Run("Add/Get", func(t *testing.T) {
		_, err := store.Get(ctx, preport.GetID())
		if err == nil {
			t.Fatalf("Should not be found in empty Store")
		}

		store.Add(ctx, preport)
		_, err = store.Get(ctx, preport.GetID())
		if err != nil {
			t.Errorf("Should be found in Store after adding report to the store")
		}
	})

	t.Run("Update/Get", func(t *testing.T) {
		ureport := &v1alpha2.PolicyReport{
			ObjectMeta: v1.ObjectMeta{
				Name:              "polr-test",
				Namespace:         "test",
				CreationTimestamp: v1.Now(),
			},
			Results: make([]v1alpha2.PolicyReportResult, 0),
			Summary: v1alpha2.PolicyReportSummary{Skip: 1},
		}

		store.Add(ctx, preport)
		r, _ := store.Get(ctx, preport.GetID())
		if r.GetSummary().Skip != 0 {
			t.Errorf("Expected Summary.Skip to be 0")
		}

		store.Update(ctx, ureport)
		r2, _ := store.Get(ctx, preport.GetID())
		if r2.GetSummary().Skip != 1 {
			t.Errorf("Expected Summary.Skip to be 1 after update")
		}
	})

	t.Run("Delete/Get", func(t *testing.T) {
		_, err := store.Get(ctx, preport.GetID())
		if err != nil {
			t.Errorf("Should be found in Store after adding report to the store")
		}

		store.Remove(ctx, preport.GetID())
		_, err = store.Get(ctx, preport.GetID())
		if err == nil {
			t.Fatalf("Should not be found after Remove report from Store")
		}
	})

	t.Run("CleanUp", func(t *testing.T) {
		store.Add(ctx, preport)

		store.CleanUp(ctx)
		_, err := store.Get(ctx, preport.GetID())
		if err == nil {
			t.Fatalf("Should have no results after CleanUp")
		}
	})
}
