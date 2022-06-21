package report

import (
	"fmt"
	"log"
	"sync"
	"time"
)

type Group struct {
	reportsChan chan string
	eventGroups map[string]chan LifecycleEvent
}

func (g *Group) Register(reportID string) {
	g.eventGroups[reportID] = make(chan LifecycleEvent)
	g.reportsChan <- reportID
}

func (g *Group) ChannelAdded() <-chan string {
	return g.reportsChan
}

func (g *Group) AddEvent(event LifecycleEvent) {
	if channel, ok := g.eventGroups[event.NewPolicyReport.ID]; ok {
		channel <- event
	}
}

func (g *Group) Listen(reportID string) (<-chan LifecycleEvent, error) {
	if channel, ok := g.eventGroups[reportID]; ok {
		return channel, nil
	}

	return nil, fmt.Errorf("channel for ReportID %s not found", reportID)
}

func (g *Group) ListenWithRetry(reportID string, retry int) (<-chan LifecycleEvent, error) {
	var count int
	var err error
	for count <= retry {
		channel, err := g.Listen(reportID)
		if err == nil {
			return channel, nil
		}
		count++
		time.Sleep(1 * time.Second)
	}

	return nil, err
}

func (g *Group) Close(reportID string) error {
	if channel, ok := g.eventGroups[reportID]; ok {
		close(channel)
		delete(g.eventGroups, reportID)
		return nil
	}

	return fmt.Errorf("channel for ReportID %s not found", reportID)
}

func (g *Group) CloseRegisterChannel() {
	close(g.reportsChan)
}

func (g *Group) CloseAll() {
	for reportID, channel := range g.eventGroups {
		close(channel)
		delete(g.eventGroups, reportID)
	}

	close(g.reportsChan)
}

func NewGroup() *Group {
	return &Group{
		reportsChan: make(chan string),
		eventGroups: make(map[string]chan LifecycleEvent),
	}
}

type EventPublisher interface {
	// RegisterListener register Handlers called on each PolicyReport watch.Event
	RegisterListener(PolicyReportListener)
	// GetListener returns a list of all registered Listeners
	GetListener() []PolicyReportListener
	// Publish events to the registered listeners
	Publish(reportChannels *Group)
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

func (p *lifecycleEventPublisher) Publish(reportChannels *Group) {
	for channelName := range reportChannels.ChannelAdded() {
		go func(cName string) {
			channel, err := reportChannels.ListenWithRetry(cName, 3)
			if err != nil {
				log.Println(err.Error())
			}

			for event := range channel {
				wg := sync.WaitGroup{}
				wg.Add(p.listenerCount)

				for _, listener := range p.listeners {
					go func(li PolicyReportListener, ev LifecycleEvent) {
						li(ev)
						wg.Done()
					}(listener, event)
				}

				wg.Wait()

				if event.Type == Deleted {
					reportChannels.Close(cName)
				}
			}
		}(channelName)
	}
}

func NewEventPublisher() EventPublisher {
	return &lifecycleEventPublisher{}
}
