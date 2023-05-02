package listener

import (
	"context"

	"go.uber.org/zap"

	"github.com/kyverno/policy-reporter/pkg/report"
)

const Store = "store_listener"

func NewStoreListener(ctx context.Context, store report.PolicyReportStore) report.PolicyReportListener {
	return func(event report.LifecycleEvent) {
		if event.Type == report.Deleted {
			logOnError("remove", event.PolicyReport.GetName(), store.Remove(ctx, event.PolicyReport.GetID()))
			return
		}

		if event.Type == report.Updated {
			logOnError("update", event.PolicyReport.GetName(), store.Update(ctx, event.PolicyReport))
			return
		}

		logOnError("add", event.PolicyReport.GetName(), store.Update(ctx, event.PolicyReport))
	}
}

func logOnError(operation, name string, err error) {
	if err != nil {
		zap.L().Error("failed to "+operation+" policy report", zap.String("name", name), zap.Error(err))
	}
}
