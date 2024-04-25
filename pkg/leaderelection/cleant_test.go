package leaderelection_test

import (
	"context"
	"testing"
	"time"

	"github.com/kyverno/policy-reporter/pkg/leaderelection"
	"k8s.io/client-go/kubernetes/typed/coordination/v1/fake"
)

func TestClient(t *testing.T) {
	client := leaderelection.New(&fake.FakeCoordinationV1{}, "policy-reporter", "namespace", "pod-123", time.Second, time.Second, time.Second, false)

	if client == nil {
		t.Fatal("failed to create leaderelection client")
	}

	client.RegisterOnNew(func(currentID, lockID string) {})
	client.RegisterOnStart(func(c context.Context) {})
	client.RegisterOnStop(func() {})

	lock := client.CreateLock()

	if lock.LeaseMeta.Name != "policy-reporter" {
		t.Error("unexpected lease name")
	}
	if lock.LeaseMeta.Namespace != "namespace" {
		t.Error("unexpected lease namespace")
	}
	if lock.LockConfig.Identity != "pod-123" {
		t.Error("unexpected lease identity")
	}
}
