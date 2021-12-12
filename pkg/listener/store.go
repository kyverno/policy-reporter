package listener

import (
	"log"

	"github.com/kyverno/policy-reporter/pkg/report"
)

func NewStoreListener(store report.PolicyReportStore) report.PolicyReportListener {
	return func(event report.LifecycleEvent) {
		if event.Type == report.Deleted {
			logOnError("remove", event.NewPolicyReport.Name, store.Remove(event.NewPolicyReport.GetIdentifier()))
			return
		}

		if event.Type == report.Updated {
			logOnError("update", event.NewPolicyReport.Name, store.Update(event.NewPolicyReport))
			return
		}

		logOnError("add", event.NewPolicyReport.Name, store.Add(event.NewPolicyReport))
	}
}

func logOnError(operation, name string, err error) {
	if err != nil {
		log.Printf("[ERROR] Failed to %s Policy Report %s (%s)\n", operation, name, err.Error())
	}
}
