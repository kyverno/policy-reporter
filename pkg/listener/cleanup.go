package listener

import (
	"context"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/helper"
	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/kyverno/policy-reporter/pkg/target"
)

const CleanUpListener = "cleanup_listener"

func NewCleanupListener(ctx context.Context, targets *target.Collection) report.PolicyReportListener {
	return func(event report.LifecycleEvent) {
		for _, handler := range targets.Clients() {
			if event.Type != report.Deleted {
				filtered := helper.Filter(event.PolicyReport.GetResults(), func(result v1alpha2.PolicyReportResult) bool {
					return handler.Validate(event.PolicyReport, result)
				})

				if len(filtered) == 0 {
					return
				}
			}

			handler.CleanUp(ctx, event.PolicyReport)
		}
	}
}
