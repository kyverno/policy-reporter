package listener

import (
	"context"

	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/kyverno/policy-reporter/pkg/target"
)

const CleanUpListener = "cleanup_listener"

func NewCleanupListener(ctx context.Context, targets *target.Collection) report.PolicyReportListener {
	return func(event report.LifecycleEvent) {
		if event.Type == report.Added {
			return
		}

		for _, handler := range targets.SyncClients() {
			handler.CleanUp(ctx, event.PolicyReport)
		}
	}
}
