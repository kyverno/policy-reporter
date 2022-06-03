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

	groups := report.NewGroup()

	go func() {
		groups.Register("UID")

		groups.AddEvent(report.LifecycleEvent{Type: report.Updated, NewPolicyReport: &report.PolicyReport{ID: "UID"}, OldPolicyReport: &report.PolicyReport{ID: "UID"}})

		groups.CloseAll()
	}()

	publisher.Publish(groups)

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

	groups := report.NewGroup()

	go func() {
		groups.Register("UID")

		groups.AddEvent(report.LifecycleEvent{Type: report.Updated, NewPolicyReport: &report.PolicyReport{ID: "UID"}, OldPolicyReport: &report.PolicyReport{ID: "UID"}})
		groups.AddEvent(report.LifecycleEvent{Type: report.Deleted, NewPolicyReport: &report.PolicyReport{ID: "UID"}})

		groups.CloseRegisterChannel()
	}()

	publisher.Publish(groups)

	wg.Wait()

	if event.Type != report.Deleted {
		t.Error("Expected Event to be published to the listener")
	}
	if _, err := groups.Listen("UID"); err == nil {
		t.Error("Expected report to be deleted")
	}
}

func Test_GetReisteredListeners(t *testing.T) {
	publisher := report.NewEventPublisher()
	publisher.RegisterListener(func(le report.LifecycleEvent) {})

	if len(publisher.GetListener()) != 1 {
		t.Error("Expected to get one registered listener back")
	}
}

func Test_ListenUknownChannel(t *testing.T) {
	reportChannel := report.NewGroup()
	_, err := reportChannel.Listen("test")

	if err == nil {
		t.Error("Expected to get a not found error")
	}
}
