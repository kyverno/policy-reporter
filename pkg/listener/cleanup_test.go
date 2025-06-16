package listener_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kyverno/policy-reporter/pkg/listener"
	"github.com/kyverno/policy-reporter/pkg/openreports"
	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/kyverno/policy-reporter/pkg/target"
)

func Test_CleanupListener(t *testing.T) {
	t.Run("Execute Cleanup Handler", func(t *testing.T) {
		c := &client{cleanup: true}

		slistener := listener.NewCleanupListener(ctx, target.NewCollection(&target.Target{Client: c}))
		slistener(report.LifecycleEvent{Type: report.Deleted, PolicyReport: &openreports.ORReportAdapter{Report: preport1}})

		assert.True(t, c.cleanupCalled, "expected cleanup method was called")
	})
}

func Test_Cleanup_Listener_Skip_Added(t *testing.T) {
	t.Run("Execute Cleanup Handler", func(t *testing.T) {
		c := &client{cleanup: true}

		slistener := listener.NewCleanupListener(ctx, target.NewCollection(&target.Target{Client: c}))
		slistener(report.LifecycleEvent{Type: report.Added, PolicyReport: &openreports.ORReportAdapter{Report: preport1}})

		assert.False(t, c.cleanupCalled, "expected cleanup execution was skipped")
	})
}
