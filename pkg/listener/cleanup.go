package listener

import (
	"context"

	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/kyverno/policy-reporter/pkg/target"
)

const CleanUpListener = "cleanup_listener"

func NewCleanupListener(ctx context.Context, handlers []target.Client) report.PolicyReportListener {
	return func(event report.LifecycleEvent) {
		for _, handler := range handlers {
			handler.CleanUp(ctx, event.PolicyReport)
		}
	}
}
