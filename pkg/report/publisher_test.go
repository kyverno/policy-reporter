package report_test

import (
	"sync"
	"testing"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"openreports.io/apis/openreports.io/v1alpha1"

	"github.com/kyverno/policy-reporter/pkg/report"
)

func Test_PublishLifecycleEvents(t *testing.T) {
	var event report.LifecycleEvent

	wg := sync.WaitGroup{}
	wg.Add(2)

	publisher := report.NewEventPublisher()
	publisher.RegisterListener("test", func(le report.LifecycleEvent) {
		event = le
		wg.Done()
	})

	publisher.Publish(report.LifecycleEvent{Type: report.Added, PolicyReport: &v1alpha1.Report{
		ObjectMeta: v1.ObjectMeta{
			Name:      "polr-test",
			Namespace: "test",
		},
	}})

	publisher.Publish(report.LifecycleEvent{Type: report.Updated, PolicyReport: &v1alpha1.Report{
		ObjectMeta: v1.ObjectMeta{
			Name:      "polr-test",
			Namespace: "test",
		},
	}})

	wg.Wait()

	if event.Type != report.Updated {
		t.Error("Expected Event to be published to the listener")
	}
}

func Test_PublishDeleteLifecycleEvents(t *testing.T) {
	var event report.LifecycleEvent

	wg := sync.WaitGroup{}
	wg.Add(2)

	publisher := report.NewEventPublisher()
	publisher.RegisterListener("test", func(le report.LifecycleEvent) {
		event = le
		wg.Done()
	})

	publisher.Publish(report.LifecycleEvent{Type: report.Added, PolicyReport: &v1alpha1.Report{
		ObjectMeta: v1.ObjectMeta{
			Name:      "polr-test",
			Namespace: "test",
		},
	}})
	publisher.Publish(report.LifecycleEvent{Type: report.Deleted, PolicyReport: &v1alpha1.Report{
		ObjectMeta: v1.ObjectMeta{
			Name:      "polr-test",
			Namespace: "test",
		},
	}})

	wg.Wait()

	if event.Type != report.Deleted {
		t.Error("Expected Event to be published to the listener")
	}
}

func Test_GetReisteredListeners(t *testing.T) {
	publisher := report.NewEventPublisher()
	publisher.RegisterListener("test", func(le report.LifecycleEvent) {})

	if len(publisher.GetListener()) != 1 {
		t.Error("Expected to get one registered listener back")
	}
}

func Test_UnreisteredListeners(t *testing.T) {
	publisher := report.NewEventPublisher()
	publisher.RegisterListener("test", func(le report.LifecycleEvent) {})
	publisher.UnregisterListener("test")

	if len(publisher.GetListener()) != 0 {
		t.Error("Expected to get 0 listeners back after unregistration")
	}
}
