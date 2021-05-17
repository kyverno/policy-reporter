package kubernetes

import (
	"sync"
	"time"

	"github.com/fjogeleit/policy-reporter/pkg/report"
	"k8s.io/apimachinery/pkg/watch"
)

type policyReportEvent struct {
	report    report.PolicyReport
	eventType watch.EventType
}

type policyReportEventDebouncer struct {
	events       map[string]policyReportEvent
	channel      chan policyReportEvent
	mutx         *sync.Mutex
	debounceTime time.Duration
}

func (d *policyReportEventDebouncer) Add(e policyReportEvent) {
	_, ok := d.events[e.report.GetIdentifier()]
	if e.eventType != watch.Modified && ok {
		d.mutx.Lock()
		delete(d.events, e.report.GetIdentifier())
		d.mutx.Unlock()
	}

	if e.eventType != watch.Modified {
		d.channel <- e
		return
	}

	if len(e.report.Results) == 0 && !ok {
		d.mutx.Lock()
		d.events[e.report.GetIdentifier()] = e
		d.mutx.Unlock()

		go func() {
			time.Sleep(d.debounceTime * time.Second)

			d.mutx.Lock()
			if event, ok := d.events[e.report.GetIdentifier()]; ok {
				d.channel <- event
				delete(d.events, e.report.GetIdentifier())
			}
			d.mutx.Unlock()
		}()

		return
	}

	if ok {
		d.mutx.Lock()
		d.events[e.report.GetIdentifier()] = e
		d.mutx.Unlock()

		return
	}

	d.channel <- e
}

func (d *policyReportEventDebouncer) Reset() {
	d.mutx.Lock()
	d.events = make(map[string]policyReportEvent)
	d.mutx.Unlock()
}

func (d *policyReportEventDebouncer) ReportChan() chan policyReportEvent {
	return d.channel
}

type clusterPolicyReportEvent struct {
	report    report.ClusterPolicyReport
	eventType watch.EventType
}

type clusterPolicyReportEventDebouncer struct {
	events       map[string]clusterPolicyReportEvent
	channel      chan clusterPolicyReportEvent
	mutx         *sync.Mutex
	debounceTime time.Duration
}

func (d *clusterPolicyReportEventDebouncer) Add(e clusterPolicyReportEvent) {
	_, ok := d.events[e.report.GetIdentifier()]
	if e.eventType != watch.Modified && ok {
		d.mutx.Lock()
		delete(d.events, e.report.GetIdentifier())
		d.mutx.Unlock()
	}

	if e.eventType != watch.Modified {
		d.channel <- e
		return
	}

	if len(e.report.Results) == 0 && !ok {
		d.mutx.Lock()
		d.events[e.report.GetIdentifier()] = e
		d.mutx.Unlock()

		go func() {
			time.Sleep(d.debounceTime * time.Second)

			d.mutx.Lock()
			if event, ok := d.events[e.report.GetIdentifier()]; ok {
				d.channel <- event
				delete(d.events, e.report.GetIdentifier())
			}
			d.mutx.Unlock()
		}()

		return
	}

	if ok {
		d.mutx.Lock()
		d.events[e.report.GetIdentifier()] = e
		d.mutx.Unlock()

		return
	}

	d.channel <- e
}

func (d *clusterPolicyReportEventDebouncer) Reset() {
	d.mutx.Lock()
	d.events = make(map[string]clusterPolicyReportEvent)
	d.mutx.Unlock()
}

func (d *clusterPolicyReportEventDebouncer) ReportChan() chan clusterPolicyReportEvent {
	return d.channel
}
