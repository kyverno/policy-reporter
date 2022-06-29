package report

import (
	"sync"
)

type EventPublisher interface {
	// RegisterListener register Handlers called on each PolicyReport watch.Event
	RegisterListener(PolicyReportListener)
	// GetListener returns a list of all registered Listeners
	GetListener() []PolicyReportListener
	// Process LifecycleEvent with all registered listeners
	Publish(event LifecycleEvent)
}

type lifecycleEventPublisher struct {
	listeners     []PolicyReportListener
	listenerCount int
}

func (p *lifecycleEventPublisher) RegisterListener(listener PolicyReportListener) {
	p.listeners = append(p.listeners, listener)
	p.listenerCount++
}

func (p *lifecycleEventPublisher) GetListener() []PolicyReportListener {
	return p.listeners
}

func (p *lifecycleEventPublisher) Publish(event LifecycleEvent) {
	g := sync.WaitGroup{}
	g.Add(len(p.listeners))
	for _, listener := range p.listeners {
		go func(li PolicyReportListener, ev LifecycleEvent) {
			li(ev)

			g.Done()
		}(listener, event)
	}

	g.Wait()
}

func NewEventPublisher() EventPublisher {
	return &lifecycleEventPublisher{}
}
