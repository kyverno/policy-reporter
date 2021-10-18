package kubernetes

import (
	"context"
	"log"
	"sync"

	"github.com/kyverno/policy-reporter/pkg/report"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
)

var (
	policyReportAlphaV1 = schema.GroupVersionResource{
		Group:    "wgpolicyk8s.io",
		Version:  "v1alpha1",
		Resource: "policyreports",
	}
	policyReportAlphaV2 = schema.GroupVersionResource{
		Group:    "wgpolicyk8s.io",
		Version:  "v1alpha2",
		Resource: "policyreports",
	}

	clusterPolicyReportAlphaV1 = schema.GroupVersionResource{
		Group:    "wgpolicyk8s.io",
		Version:  "v1alpha1",
		Resource: "clusterpolicyreports",
	}
	clusterPolicyReportAlphaV2 = schema.GroupVersionResource{
		Group:    "wgpolicyk8s.io",
		Version:  "v1alpha2",
		Resource: "clusterpolicyreports",
	}
)

// WatchEvent of PolicyReports
type WatchEvent struct {
	Report report.PolicyReport
	Type   watch.EventType
}

// PolicyReportAdapter translates API responses to an internal struct
type PolicyReportAdapter interface {
	WatchPolicyReports(ctx context.Context) (chan WatchEvent, error)
	GetFoundResources() map[string]string
}

type k8sPolicyReportAdapter struct {
	client dynamic.Interface
	found  map[string]string
	mapper Mapper
	mx     *sync.Mutex
}

func (k *k8sPolicyReportAdapter) GetFoundResources() map[string]string {
	return k.found
}

func (k *k8sPolicyReportAdapter) WatchPolicyReports(ctx context.Context) (chan WatchEvent, error) {
	events := make(chan WatchEvent)

	pr := []schema.GroupVersionResource{
		policyReportAlphaV2,
		policyReportAlphaV1,
	}

	cpor := []schema.GroupVersionResource{
		clusterPolicyReportAlphaV2,
		clusterPolicyReportAlphaV1,
	}

	for _, versions := range [][]schema.GroupVersionResource{pr, cpor} {
		go func(vs []schema.GroupVersionResource) {
			for {
				for _, resource := range vs {
					k.WatchCRD(ctx, resource, events)
				}
			}

		}(versions)
	}

	return events, nil
}

func (k *k8sPolicyReportAdapter) WatchCRD(ctx context.Context, r schema.GroupVersionResource, events chan WatchEvent) {
	for {
		w, err := k.client.Resource(r).Watch(ctx, metav1.ListOptions{})
		if err != nil {
			k.mx.Lock()
			delete(k.found, r.String())
			k.mx.Unlock()

			return
		}

		log.Printf("[INFO] Resource registered: %s\n", r.String())

		k.mx.Lock()
		k.found[r.String()] = r.String()
		k.mx.Unlock()

		for result := range w.ResultChan() {
			if item, ok := result.Object.(*unstructured.Unstructured); ok {
				preport := k.mapper.MapPolicyReport(item.Object)
				events <- WatchEvent{preport, result.Type}
			}
		}
	}
}

// NewPolicyReportAdapter new Adapter for Policy Report Kubernetes API
func NewPolicyReportAdapter(dynamic dynamic.Interface, mapper Mapper) PolicyReportAdapter {
	return &k8sPolicyReportAdapter{
		client: dynamic,
		mapper: mapper,
		mx:     &sync.Mutex{},
		found:  make(map[string]string),
	}
}
