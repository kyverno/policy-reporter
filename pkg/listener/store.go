package listener

import (
	"log"

	"github.com/kyverno/policy-reporter/pkg/report"
)

const Store = "store_listener"

func NewStoreListener(store report.PolicyReportStore) report.PolicyReportListener {
	return func(event report.LifecycleEvent) {
		if event.Type == report.Deleted {
			logOnError("remove", event.NewPolicyReport.GetName(), store.Remove(event.NewPolicyReport.GetID()))
			return
		}

		if event.Type == report.Updated {
			logOnError("update", event.NewPolicyReport.GetName(), store.Update(event.NewPolicyReport))
			return
		}

		logOnError("add", event.NewPolicyReport.GetName(), store.Add(event.NewPolicyReport))
	}
}

func logOnError(operation, name string, err error) {
	if err != nil {
		log.Printf("[ERROR] Failed to %s Policy Report %s (%s)\n", operation, name, err.Error())
	}
}
