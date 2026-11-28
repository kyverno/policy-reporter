package kubernetes

import (
	"sync"
	"time"

	"github.com/kyverno/policy-reporter/pkg/report"
)

type Debouncer interface {
	Add(e report.LifecycleEvent)
}

// window is one debounce period of one report. Its pointer identity is the token
// a timer compares with the pending entry, so a timer whose Stop lost the race
// cannot publish the event of a newer window of the same report.
type window struct {
	timer *time.Timer
	event report.LifecycleEvent
}

// publication keeps the listener calls of one report in FIFO order: events
// queued while a drain runs are published by that drain, never concurrently.
type publication struct {
	events []report.LifecycleEvent
}

type debouncer struct {
	waitDuration time.Duration
	mutx         sync.Mutex
	pending      map[string]*window
	publications map[string]*publication
	publisher    report.EventPublisher
}

func (d *debouncer) Add(event report.LifecycleEvent) {
	id := event.PolicyReport.GetID()

	if event.Type == report.Updated && len(event.PolicyReport.GetResults()) == 0 {
		d.mutx.Lock()
		if w, ok := d.pending[id]; ok {
			// Coalesce into the open window: latest event, original deadline.
			persisted := w.event.Persisted
			w.event = event
			if w.event.Persisted == nil {
				w.event.Persisted = persisted
			}
			d.mutx.Unlock()
			return
		}

		w := &window{event: event}
		d.pending[id] = w
		w.timer = time.AfterFunc(d.waitDuration, func() { d.expire(id, w) })
		d.mutx.Unlock()
		return
	}

	d.mutx.Lock()
	// Every other event of the report supersedes its pending empty update.
	if w, ok := d.pending[id]; ok {
		delete(d.pending, id)
		w.timer.Stop()
	}
	claimed := d.enqueueLocked(id, event)
	d.mutx.Unlock()

	if claimed {
		d.drain(id)
	}
}

func (d *debouncer) expire(id string, w *window) {
	d.mutx.Lock()
	if d.pending[id] != w {
		d.mutx.Unlock()
		return
	}
	delete(d.pending, id)
	claimed := d.enqueueLocked(id, w.event)
	d.mutx.Unlock()

	if claimed {
		d.drain(id)
	}
}

// enqueueLocked appends the event to the publication order of the report and
// reports whether the caller owns draining it. Map membership is the ownership:
// the entry exists from the owner's enqueue until its drain finds the queue empty
// and deletes it, so a second caller always finds the entry and never drains.
func (d *debouncer) enqueueLocked(id string, event report.LifecycleEvent) bool {
	p, ok := d.publications[id]
	if !ok {
		p = &publication{}
		d.publications[id] = p
	}
	p.events = append(p.events, event)
	return !ok
}

// drain publishes the queued events of one report in queue order without holding
// the debouncer lock, so listeners of other reports never wait for a listener.
func (d *debouncer) drain(id string) {
	for {
		d.mutx.Lock()
		p := d.publications[id]
		if p == nil || len(p.events) == 0 {
			delete(d.publications, id)
			d.mutx.Unlock()
			return
		}
		event := p.events[0]
		p.events[0] = report.LifecycleEvent{}
		p.events = p.events[1:]
		d.mutx.Unlock()

		d.publisher.Publish(event)
	}
}

func NewDebouncer(waitDuration time.Duration, publisher report.EventPublisher) Debouncer {
	return &debouncer{
		waitDuration: waitDuration,
		pending:      make(map[string]*window),
		publications: make(map[string]*publication),
		publisher:    publisher,
	}
}
