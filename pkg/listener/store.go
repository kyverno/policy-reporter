package listener

import (
	"log"

	"github.com/kyverno/policy-reporter/pkg/report"
)

const Store = "store_listener"

func NewStoreListener(store report.PolicyReportStore) report.PolicyReportListener {
	return func(event report.LifecycleEvent) {
		if event.Type == report.Deleted {
			logOnError("remove", event.PolicyReport.GetName(), store.Remove(event.PolicyReport.GetID()))
			return
		}

		if event.Type == report.Updated {
			logOnError("update", event.PolicyReport.GetName(), store.Update(event.PolicyReport))
			return
		}

		logOnError("add", event.PolicyReport.GetName(), store.Add(event.PolicyReport))
	}
}

func logOnError(operation, name string, err error) {
	if err != nil {
		log.Printf("[ERROR] Failed to %s Policy Report %s (%s)\n", operation, name, err.Error())
	}
}
