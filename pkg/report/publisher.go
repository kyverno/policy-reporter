package report

import (
	"sync"
)

type EventPublisher interface {
	// RegisterListener register Handlers called on each PolicyReport Event
	RegisterListener(string, PolicyReportListener)
	// UnregisterListener removes an registered listener
	UnregisterListener(string)
	// GetListener returns a list of all registered Listeners
	GetListener() map[string]PolicyReportListener
	// Publish Process LifecycleEvent with all registered listeners
	Publish(event LifecycleEvent)
	// RegisterPostListener register Handlers called on each PolicyReport Event after all regular listeners are finished
	RegisterPostListener(string, PolicyReportListener)
	// UnregisterPostListener removes an registered post listener
	UnregisterPostListener(string)
}

type lifecycleEventPublisher struct {
	postListeners map[string]PolicyReportListener
	listeners     map[string]PolicyReportListener
	listenerCount int
}

func (p *lifecycleEventPublisher) RegisterListener(name string, listener PolicyReportListener) {
	p.listeners[name] = listener
	p.listenerCount++
}

func (p *lifecycleEventPublisher) UnregisterListener(name string) {
	if _, ok := p.listeners[name]; ok {
		delete(p.listeners, name)
		p.listenerCount--
	}
}

func (p *lifecycleEventPublisher) RegisterPostListener(name string, listener PolicyReportListener) {
	p.postListeners[name] = listener
}

func (p *lifecycleEventPublisher) UnregisterPostListener(name string) {
	delete(p.postListeners, name)
}

func (p *lifecycleEventPublisher) GetListener() map[string]PolicyReportListener {
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

	for _, listener := range p.postListeners {
		listener(event)
	}
}

func NewEventPublisher() EventPublisher {
	return &lifecycleEventPublisher{
		postListeners: make(map[string]func(LifecycleEvent)),
		listeners:     make(map[string]func(LifecycleEvent)),
		listenerCount: 0,
	}
}
