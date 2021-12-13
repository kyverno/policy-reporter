package report_test

import (
	"sync"
	"testing"

	"github.com/kyverno/policy-reporter/pkg/report"
)

func Test_PublishLifecycleEvents(t *testing.T) {
	eventChan := make(chan report.LifecycleEvent)

	var event report.LifecycleEvent

	wg := sync.WaitGroup{}
	wg.Add(1)

	publisher := report.NewEventPublisher()
	publisher.RegisterListener(func(le report.LifecycleEvent) {
		event = le
		wg.Done()
	})

	go func() {
		eventChan <- report.LifecycleEvent{Type: report.Updated, NewPolicyReport: &report.PolicyReport{}, OldPolicyReport: &report.PolicyReport{}}

		close(eventChan)
	}()

	publisher.Publish(eventChan)

	wg.Wait()

	if event.Type != report.Updated {
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
