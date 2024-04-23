package listener_test

import (
	"testing"

	"github.com/kyverno/policy-reporter/pkg/listener"
	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/kyverno/policy-reporter/pkg/target"
)

func Test_CleanupListener(t *testing.T) {
	t.Run("Execute Cleanup Handler", func(t *testing.T) {
		c := &client{}

		slistener := listener.NewCleanupListener(ctx, []target.Client{c})
		slistener(report.LifecycleEvent{Type: report.Added, PolicyReport: preport1})

		if !c.cleanupCalled {
			t.Error("expected cleanup method was called")
		}
	})
}
