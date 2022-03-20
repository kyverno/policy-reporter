package kubernetes

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/kyverno/policy-reporter/pkg/report"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
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

type k8sPolicyReportClient struct {
	debouncer             Debouncer
	client                dynamic.Interface
	found                 map[string]string
	mapper                Mapper
	mx                    *sync.Mutex
	restartWatchOnFailure time.Duration
	reportFilter          report.Filter
}

func (k *k8sPolicyReportClient) GetFoundResources() map[string]string {
	return k.found
}

func (k *k8sPolicyReportClient) WatchPolicyReports(ctx context.Context) <-chan report.LifecycleEvent {
	schemas := [][]schema.GroupVersionResource{}

	pr := []schema.GroupVersionResource{
		policyReportAlphaV2,
		policyReportAlphaV1,
	}

	cpor := []schema.GroupVersionResource{
		clusterPolicyReportAlphaV2,
		clusterPolicyReportAlphaV1,
	}

	schemas = append(schemas, pr)
	if !k.reportFilter.DisableClusterReports() {
		schemas = append(schemas, cpor)
	}

	for _, versions := range schemas {
		go func(vs []schema.GroupVersionResource) {
			for {
				factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(k.client, 30*time.Minute, corev1.NamespaceAll, nil)
				for _, resource := range vs {
					k.watchCRD(ctx, resource, factory)
				}

				time.Sleep(2 * time.Second)
			}

		}(versions)
	}

	for {
		if !k.reportFilter.DisableClusterReports() && len(k.found) == 2 {
			break
		} else if k.reportFilter.DisableClusterReports() && len(k.found) == 1 {
			break
		}
	}

	return k.debouncer.ReportChan()
}

func (k *k8sPolicyReportClient) watchCRD(ctx context.Context, r schema.GroupVersionResource, factory dynamicinformer.DynamicSharedInformerFactory) {
	informer := factory.ForResource(r).Informer()

	ctx, cancel := context.WithCancel(ctx)

	informer.SetWatchErrorHandler(func(c *cache.Reflector, err error) {
		k.mx.Lock()
		delete(k.found, r.String())
		k.mx.Unlock()
		cancel()

		log.Printf("[WARNING] Resource registration failed: %s\n", r.String())
	})

	go k.handleCRDRegistration(ctx, informer, r)

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if item, ok := obj.(*unstructured.Unstructured); ok {
				preport := k.mapper.MapPolicyReport(item.Object)
				if k.reportFilter.AllowReport(preport) {
					k.debouncer.Add(report.LifecycleEvent{NewPolicyReport: preport, OldPolicyReport: &report.PolicyReport{}, Type: report.Added})
				}
			}
		},
		DeleteFunc: func(obj interface{}) {
			if item, ok := obj.(*unstructured.Unstructured); ok {
				preport := k.mapper.MapPolicyReport(item.Object)
				if k.reportFilter.AllowReport(preport) {
					k.debouncer.Add(report.LifecycleEvent{NewPolicyReport: preport, OldPolicyReport: &report.PolicyReport{}, Type: report.Deleted})
				}
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			if item, ok := newObj.(*unstructured.Unstructured); ok {
				preport := k.mapper.MapPolicyReport(item.Object)

				var oreport *report.PolicyReport
				if oldItem, ok := oldObj.(*unstructured.Unstructured); ok {
					oreport = k.mapper.MapPolicyReport(oldItem.Object)
				}

				if k.reportFilter.AllowReport(preport) {
					k.debouncer.Add(report.LifecycleEvent{NewPolicyReport: preport, OldPolicyReport: oreport, Type: report.Updated})
				}
			}
		},
	})

	informer.Run(ctx.Done())
}

func (k *k8sPolicyReportClient) handleCRDRegistration(ctx context.Context, informer cache.SharedIndexInformer, r schema.GroupVersionResource) {
	ticker := time.NewTicker(k.restartWatchOnFailure)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if informer.HasSynced() {
				k.mx.Lock()
				k.found[r.String()] = r.String()
				k.mx.Unlock()

				log.Printf("[INFO] Resource registered: %s\n", r.String())
				return
			}
		}
	}
}

// NewPolicyReportAdapter new Adapter for Policy Report Kubernetes API
func NewPolicyReportClient(dynamic dynamic.Interface, mapper Mapper, restartWatchOnFailure time.Duration, reportFilter report.Filter) report.PolicyReportClient {
	return &k8sPolicyReportClient{
		client:                dynamic,
		mapper:                mapper,
		mx:                    &sync.Mutex{},
		found:                 make(map[string]string),
		debouncer:             NewDebouncer(time.Minute),
		restartWatchOnFailure: restartWatchOnFailure,
		reportFilter:          reportFilter,
	}
}
