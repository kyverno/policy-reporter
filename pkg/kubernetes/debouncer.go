package kubernetes

import (
	"sync"
	"time"

	"github.com/kyverno/policy-reporter/pkg/report"
)

type Debouncer interface {
	Add(e report.LifecycleEvent)
}

type debouncer struct {
	waitDuration time.Duration
	events       map[string]report.LifecycleEvent
	publisher    report.EventPublisher
	mutx         *sync.Mutex
}

func (d *debouncer) Add(event report.LifecycleEvent) {
	cached, ok := d.events[event.NewPolicyReport.GetID()]
	if event.Type != report.Updated && ok {
		d.mutx.Lock()
		delete(d.events, event.NewPolicyReport.GetID())
		d.mutx.Unlock()
	}

	if event.Type == report.Added || event.Type == report.Deleted {
		d.publisher.Publish(event)
		return
	}

	if len(event.NewPolicyReport.GetResults()) == 0 && !ok {
		d.mutx.Lock()
		d.events[event.NewPolicyReport.GetID()] = event
		d.mutx.Unlock()

		go func() {
			time.Sleep(d.waitDuration)

			d.mutx.Lock()
			if event, ok := d.events[event.NewPolicyReport.GetID()]; ok {
				d.publisher.Publish(event)
				delete(d.events, event.NewPolicyReport.GetID())
			}
			d.mutx.Unlock()
		}()

		return
	}

	if ok {
		d.mutx.Lock()
		event.OldPolicyReport = cached.OldPolicyReport
		d.events[event.NewPolicyReport.GetID()] = event
		d.mutx.Unlock()

		return
	}

	d.publisher.Publish(event)
}

func NewDebouncer(waitDuration time.Duration, publisher report.EventPublisher) Debouncer {
	return &debouncer{
		waitDuration: waitDuration,
		events:       make(map[string]report.LifecycleEvent),
		mutx:         new(sync.Mutex),
		publisher:    publisher,
	}
}
