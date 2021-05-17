package report_test

import (
	"testing"

	"github.com/fjogeleit/policy-reporter/pkg/report"
)

func Test_PolicyReportStore(t *testing.T) {
	store := report.NewPolicyReportStore()

	t.Run("Add/Get", func(t *testing.T) {
		_, ok := store.Get(preport.GetType(), preport.GetIdentifier())
		if ok == true {
			t.Fatalf("Should not be found in empty Store")
		}

		store.Add(preport)
		_, ok = store.Get(preport.GetType(), preport.GetIdentifier())
		if ok == false {
			t.Errorf("Should be found in Store after adding report to the store")
		}
	})

	t.Run("List", func(t *testing.T) {
		items := store.List(preport.GetType())
		if len(items) != 1 {
			t.Errorf("Should return List with the added Report")
		}
	})

	t.Run("Delete/Get", func(t *testing.T) {
		_, ok := store.Get(preport.GetType(), preport.GetIdentifier())
		if ok == false {
			t.Errorf("Should be found in Store after adding report to the store")
		}

		store.Remove(preport.GetType(), preport.GetIdentifier())
		_, ok = store.Get(preport.GetType(), preport.GetIdentifier())
		if ok == true {
			t.Fatalf("Should not be found after Remove report from Store")
		}
	})
}
