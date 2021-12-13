package report

import "sync"

type EventPublisher interface {
	// RegisterListener register Handlers called on each PolicyReport watch.Event
	RegisterListener(PolicyReportListener)
	// GetListener returns a list of all registered Listeners
	GetListener() []PolicyReportListener
	// Publish events to the registered listeners
	Publish(eventChan <-chan LifecycleEvent)
}

type lifecycleEventPublisher struct {
	listeners []PolicyReportListener
}

func (p *lifecycleEventPublisher) RegisterListener(listener PolicyReportListener) {
	p.listeners = append(p.listeners, listener)
}

func (p *lifecycleEventPublisher) GetListener() []PolicyReportListener {
	return p.listeners
}

func (p *lifecycleEventPublisher) Publish(eventChan <-chan LifecycleEvent) {
	for event := range eventChan {
		wg := sync.WaitGroup{}
		wg.Add(len(p.listeners))

		for _, listener := range p.listeners {
			go func(li PolicyReportListener, ev LifecycleEvent) {
				li(event)
				wg.Done()
			}(listener, event)
		}

		wg.Wait()
	}
}

func NewEventPublisher() EventPublisher {
	return &lifecycleEventPublisher{}
}
