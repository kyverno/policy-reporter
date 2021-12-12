package kubernetes

import (
	"sync"
	"time"

	"github.com/kyverno/policy-reporter/pkg/report"
)

type Debouncer interface {
	Add(e report.LifecycleEvent)
	ReportChan() <-chan report.LifecycleEvent
	Close()
}

type debouncer struct {
	waitDuration time.Duration
	events       map[string]report.LifecycleEvent
	channel      chan report.LifecycleEvent
	mutx         *sync.Mutex
}

func (d *debouncer) Add(event report.LifecycleEvent) {
	_, ok := d.events[event.NewPolicyReport.GetIdentifier()]
	if event.Type != report.Updated && ok {
		d.mutx.Lock()
		delete(d.events, event.NewPolicyReport.GetIdentifier())
		d.mutx.Unlock()
	}

	if event.Type != report.Updated {
		d.channel <- event
		return
	}

	if len(event.NewPolicyReport.Results) == 0 && !ok {
		d.mutx.Lock()
		d.events[event.NewPolicyReport.GetIdentifier()] = event
		d.mutx.Unlock()

		go func() {
			time.Sleep(d.waitDuration)

			d.mutx.Lock()
			if event, ok := d.events[event.NewPolicyReport.GetIdentifier()]; ok {
				d.channel <- event
				delete(d.events, event.NewPolicyReport.GetIdentifier())
			}
			d.mutx.Unlock()
		}()

		return
	}

	if ok {
		d.mutx.Lock()
		d.events[event.NewPolicyReport.GetIdentifier()] = event
		d.mutx.Unlock()

		return
	}

	d.channel <- event
}

func (d *debouncer) ReportChan() <-chan report.LifecycleEvent {
	return d.channel
}

func (d *debouncer) Close() {
	close(d.channel)
}

func NewDebouncer(waitDuration time.Duration) Debouncer {
	return &debouncer{
		waitDuration: waitDuration,
		events:       make(map[string]report.LifecycleEvent),
		mutx:         new(sync.Mutex),
		channel:      make(chan report.LifecycleEvent),
	}
}
