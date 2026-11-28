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
	mutx         *sync.RWMutex
}

func (d *debouncer) Add(event report.LifecycleEvent) {
	d.mutx.RLock()
	_, ok := d.events[event.PolicyReport.GetID()]
	d.mutx.RUnlock()
	if event.Type != report.Updated && ok {
		d.mutx.Lock()
		delete(d.events, event.PolicyReport.GetID())
		d.mutx.Unlock()
	}

	if event.Type != report.Updated {
		d.publisher.Publish(event)
		return
	}

	if len(event.PolicyReport.GetResults()) == 0 {
		d.mutx.Lock()
		id := event.PolicyReport.GetID()
		_, pending := d.events[id]
		// Always keep the latest empty revision so the pending timer publishes
		// it instead of an older event scheduled by a previous update.
		d.events[id] = event
		if !pending {
			go func() {
				time.Sleep(d.waitDuration)

				d.mutx.Lock()
				if event, ok := d.events[id]; ok {
					d.publisher.Publish(event)
					delete(d.events, id)
				}
				d.mutx.Unlock()
			}()
		}
		d.mutx.Unlock()

		return
	}

	if len(event.PolicyReport.GetResults()) > 0 && ok {
		d.mutx.Lock()
		delete(d.events, event.PolicyReport.GetID())
		d.mutx.Unlock()
	}

	d.publisher.Publish(event)
}

func NewDebouncer(waitDuration time.Duration, publisher report.EventPublisher) Debouncer {
	return &debouncer{
		waitDuration: waitDuration,
		events:       make(map[string]report.LifecycleEvent),
		mutx:         new(sync.RWMutex),
		publisher:    publisher,
	}
}
