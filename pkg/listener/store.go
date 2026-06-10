package listener

import (
	"context"
	"errors"

	"go.uber.org/zap"
	"k8s.io/client-go/util/retry"

	"github.com/kyverno/policy-reporter/pkg/report"
)

const Store = "store_listener"

func NewStoreListener(store report.PolicyReportStore) report.PolicyReportListener {
	return func(ctx context.Context, event report.LifecycleEvent) {
		err := retry.OnError(retry.DefaultRetry, func(err error) bool {
			return !errors.Is(err, context.DeadlineExceeded)
		}, func() error {
			if event.Type == report.Deleted {
				return store.Remove(ctx, event.PolicyReport.GetID())
			}

			return store.Update(ctx, event.PolicyReport)
		})

		logOnError(event.Type.String(), event.PolicyReport.GetName(), err)
	}
}

func logOnError(operation, name string, err error) {
	if err != nil {
		zap.L().Error("failed to "+operation+" policy report", zap.String("name", name), zap.Error(err))
	}
}
