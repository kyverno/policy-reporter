package leaderelection_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/kubernetes/typed/coordination/v1/fake"

	"github.com/kyverno/policy-reporter/pkg/leaderelection"
)

func TestClient(t *testing.T) {
	client := leaderelection.New(&fake.FakeCoordinationV1{}, "policy-reporter", "namespace", "pod-123", time.Second, time.Second, time.Second, true)

	if client == nil {
		t.Fatal("failed to create leaderelection client")
	}

	var isLeader bool
	client.RegisterOnNew(func(currentID, lockID string) {
		isLeader = currentID == lockID
	})

	client.RegisterOnStart(func(c context.Context) {})
	client.RegisterOnStop(func() {})

	lock := client.CreateLock()

	assert.Equal(t, "policy-reporter", lock.LeaseMeta.Name)
	assert.Equal(t, "namespace", lock.LeaseMeta.Namespace)
	assert.Equal(t, "pod-123", lock.LockConfig.Identity)

	assert.False(t, isLeader)

	config := client.CreateConfig()

	config.Callbacks.OnNewLeader("pod-123")

	assert.True(t, isLeader)
	assert.Equal(t, time.Second, config.LeaseDuration)
	assert.Equal(t, time.Second, config.RenewDeadline)
	assert.Equal(t, time.Second, config.RetryPeriod)
	assert.True(t, config.ReleaseOnCancel)
}
