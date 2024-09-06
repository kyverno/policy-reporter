package listener_test

import (
	"testing"

	"github.com/kyverno/policy-reporter/pkg/listener"
	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/kyverno/policy-reporter/pkg/target"
	"github.com/stretchr/testify/assert"
)

func Test_CleanupListener(t *testing.T) {
	t.Run("Execute Cleanup Handler", func(t *testing.T) {
		c := &client{cleanup: true}

		slistener := listener.NewCleanupListener(ctx, target.NewCollection(&target.Target{Client: c}))
		slistener(report.LifecycleEvent{Type: report.Deleted, PolicyReport: preport1})

		assert.True(t, c.cleanupCalled, "expected cleanup method was called")
	})
}
