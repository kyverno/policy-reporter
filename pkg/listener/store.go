package listener

import (
	"go.uber.org/zap"

	"github.com/kyverno/policy-reporter/pkg/report"
)

const Store = "store_listener"

func NewStoreListener(store report.PolicyReportStore, logger *zap.Logger) report.PolicyReportListener {
	return func(event report.LifecycleEvent) {
		if event.Type == report.Deleted {
			logOnError(logger, "remove", event.PolicyReport.GetName(), store.Remove(event.PolicyReport.GetID()))
			return
		}

		if event.Type == report.Updated {
			logOnError(logger, "update", event.PolicyReport.GetName(), store.Update(event.PolicyReport))
			return
		}

		logOnError(logger, "add", event.PolicyReport.GetName(), store.Add(event.PolicyReport))
	}
}

func logOnError(logger *zap.Logger, operation, name string, err error) {
	if logger != nil && err != nil {
		logger.Error("failed to "+operation+" policy report", zap.String("name", name), zap.Error(err))
	}
}
