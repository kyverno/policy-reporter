package report_test

import (
	"testing"
	"time"

	"github.com/kyverno/policy-reporter/pkg/report"
)

func Test_PolicyReportStore(t *testing.T) {
	store := report.NewPolicyReportStore()
	store.CreateSchemas()

	t.Run("Add/Get", func(t *testing.T) {
		_, ok := store.Get(preport.GetIdentifier())
		if ok == true {
			t.Fatalf("Should not be found in empty Store")
		}

		store.Add(preport)
		_, ok = store.Get(preport.GetIdentifier())
		if ok == false {
			t.Errorf("Should be found in Store after adding report to the store")
		}
	})

	t.Run("Update/Get", func(t *testing.T) {
		ureport := &report.PolicyReport{
			ID:                "24cfa233af033d104cd6ce0ff9a5a875c71a5844",
			Name:              "polr-test",
			Namespace:         "test",
			Results:           make(map[string]*report.Result),
			Summary:           &report.Summary{Skip: 1},
			CreationTimestamp: time.Now(),
		}

		store.Add(preport)
		r, _ := store.Get(preport.GetIdentifier())
		if r.Summary.Skip != 0 {
			t.Errorf("Expected Summary.Skip to be 0")
		}

		store.Update(ureport)
		r2, _ := store.Get(preport.GetIdentifier())
		if r2.Summary.Skip != 1 {
			t.Errorf("Expected Summary.Skip to be 1 after update")
		}
	})

	t.Run("Delete/Get", func(t *testing.T) {
		_, ok := store.Get(preport.GetIdentifier())
		if ok == false {
			t.Errorf("Should be found in Store after adding report to the store")
		}

		store.Remove(preport.GetIdentifier())
		_, ok = store.Get(preport.GetIdentifier())
		if ok == true {
			t.Fatalf("Should not be found after Remove report from Store")
		}
	})

	t.Run("CleanUp", func(t *testing.T) {
		store.Add(preport)

		store.CleanUp()
		_, ok := store.Get(preport.GetIdentifier())
		if ok == true {
			t.Fatalf("Should have no results after CleanUp")
		}
	})
}
