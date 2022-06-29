package report_test

import (
	"sync"
	"testing"

	"github.com/kyverno/policy-reporter/pkg/report"
)

func Test_PublishLifecycleEvents(t *testing.T) {
	var event report.LifecycleEvent

	wg := sync.WaitGroup{}
	wg.Add(1)

	publisher := report.NewEventPublisher()
	publisher.RegisterListener(func(le report.LifecycleEvent) {
		event = le
		wg.Done()
	})

	publisher.Publish(report.LifecycleEvent{Type: report.Updated, NewPolicyReport: report.PolicyReport{ID: "UID"}, OldPolicyReport: report.PolicyReport{ID: "UID"}})

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
	publisher.RegisterListener(func(le report.LifecycleEvent) {
		event = le
		wg.Done()
	})

	publisher.Publish(report.LifecycleEvent{Type: report.Updated, NewPolicyReport: report.PolicyReport{ID: "UID"}, OldPolicyReport: report.PolicyReport{ID: "UID"}})
	publisher.Publish(report.LifecycleEvent{Type: report.Deleted, NewPolicyReport: report.PolicyReport{ID: "UID"}})

	wg.Wait()

	if event.Type != report.Deleted {
		t.Error("Expected Event to be published to the listener")
	}
}

func Test_GetReisteredListeners(t *testing.T) {
	publisher := report.NewEventPublisher()
	publisher.RegisterListener(func(le report.LifecycleEvent) {})

	if len(publisher.GetListener()) != 1 {
		t.Error("Expected to get one registered listener back")
	}
}
