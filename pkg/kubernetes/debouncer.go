package kubernetes

import (
	"sync"
	"time"

	"github.com/kyverno/policy-reporter/pkg/report"
)

type Debouncer interface {
	Add(e report.LifecycleEvent)
	ReportGroups() *report.Group
}

type debouncer struct {
	waitDuration time.Duration
	events       map[string]report.LifecycleEvent
	channel      *report.Group
	mutx         *sync.Mutex
}

func (d *debouncer) Add(event report.LifecycleEvent) {
	cached, ok := d.events[event.NewPolicyReport.GetIdentifier()]
	if event.Type != report.Updated && ok {
		d.mutx.Lock()
		delete(d.events, event.NewPolicyReport.GetIdentifier())
		d.mutx.Unlock()
	}

	if event.Type == report.Added {
		d.channel.Register(event.NewPolicyReport.ID)
		d.channel.AddEvent(event)
		return
	}

	if event.Type == report.Deleted {
		d.channel.AddEvent(event)
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
				d.channel.AddEvent(event)
				delete(d.events, event.NewPolicyReport.GetIdentifier())
			}
			d.mutx.Unlock()
		}()

		return
	}

	if ok {
		d.mutx.Lock()
		event.OldPolicyReport = cached.OldPolicyReport
		d.events[event.NewPolicyReport.GetIdentifier()] = event
		d.mutx.Unlock()

		return
	}

	d.channel.AddEvent(event)
}

func (d *debouncer) ReportGroups() *report.Group {
	return d.channel
}

func NewDebouncer(waitDuration time.Duration) Debouncer {
	return &debouncer{
		waitDuration: waitDuration,
		events:       make(map[string]report.LifecycleEvent),
		mutx:         new(sync.Mutex),
		channel:      report.NewGroup(),
	}
}
