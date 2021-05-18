package kubernetes

import (
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/watch"
)

type debouncer struct {
	events  map[string]WatchEvent
	channel chan WatchEvent
	mutx    *sync.Mutex
}

func (d *debouncer) Add(e WatchEvent) {
	_, ok := d.events[e.Report.GetIdentifier()]
	if e.Type != watch.Modified && ok {
		d.mutx.Lock()
		delete(d.events, e.Report.GetIdentifier())
		d.mutx.Unlock()
	}

	if e.Type != watch.Modified {
		d.channel <- e
		return
	}

	if len(e.Report.Results) == 0 && !ok {
		d.mutx.Lock()
		d.events[e.Report.GetIdentifier()] = e
		d.mutx.Unlock()

		go func() {
			time.Sleep(1 * time.Minute)

			d.mutx.Lock()
			if event, ok := d.events[e.Report.GetIdentifier()]; ok {
				d.channel <- event
				delete(d.events, e.Report.GetIdentifier())
			}
			d.mutx.Unlock()
		}()

		return
	}

	if ok {
		d.mutx.Lock()
		d.events[e.Report.GetIdentifier()] = e
		d.mutx.Unlock()

		return
	}

	d.channel <- e
}

func (d *debouncer) ReportChan() chan WatchEvent {
	return d.channel
}

func newDebouncer() *debouncer {
	return &debouncer{
		events:  make(map[string]WatchEvent),
		mutx:    new(sync.Mutex),
		channel: make(chan WatchEvent),
	}
}
