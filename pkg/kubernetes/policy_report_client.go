package kubernetes

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/kyverno/policy-reporter/pkg/report"

	pr "github.com/kyverno/kyverno/api/policyreport/v1alpha2"
	"github.com/kyverno/kyverno/pkg/client/clientset/versioned"
	"github.com/kyverno/kyverno/pkg/client/informers/externalversions"
	"github.com/kyverno/kyverno/pkg/client/informers/externalversions/policyreport/v1alpha2"
	"k8s.io/client-go/tools/cache"
)

type k8sPolicyReportClient struct {
	debouncer             Debouncer
	client                versioned.Interface
	found                 map[string]string
	mapper                Mapper
	mx                    *sync.Mutex
	restartWatchOnFailure time.Duration
	reportFilter          report.Filter
}

func (k *k8sPolicyReportClient) GetFoundResources() map[string]string {
	return k.found
}

func (k *k8sPolicyReportClient) WatchPolicyReports(ctx context.Context) *report.Group {
	factory := externalversions.NewSharedInformerFactory(k.client, 0).Wgpolicyk8s().V1alpha2()

	go func(f v1alpha2.Interface) {
		informer := factory.PolicyReports().Informer()

		for {
			k.watchPolicyReport(ctx, informer, "policyreport.wgpolicyk8s.io")
			time.Sleep(k.restartWatchOnFailure)
		}
	}(factory)

	if !k.reportFilter.DisableClusterReports() {
		informer := factory.ClusterPolicyReports().Informer()

		go func(f v1alpha2.Interface) {
			for {
				k.watchClusterPolicyReport(ctx, informer, "clusterpolicyreport.wgpolicyk8s.io")
				time.Sleep(k.restartWatchOnFailure)
			}
		}(factory)
	}

	for {
		if !k.reportFilter.DisableClusterReports() && len(k.found) == 2 {
			break
		} else if k.reportFilter.DisableClusterReports() && len(k.found) == 1 {
			break
		}
	}

	return k.debouncer.ReportGroups()
}

func (k *k8sPolicyReportClient) watchPolicyReport(ctx context.Context, informer cache.SharedIndexInformer, crd string) {
	ctx = k.addErrorWatchHandler(ctx, informer, crd)

	go k.handleCRDRegistration(ctx, informer, crd)

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if item, ok := obj.(*pr.PolicyReport); ok {
				preport := k.mapper.MapPolicyReport(item)
				if k.reportFilter.AllowReport(preport) {
					k.debouncer.Add(report.LifecycleEvent{NewPolicyReport: preport, OldPolicyReport: &report.PolicyReport{}, Type: report.Added})
				}
			}
		},
		DeleteFunc: func(obj interface{}) {
			if item, ok := obj.(*pr.PolicyReport); ok {
				preport := k.mapper.MapPolicyReport(item)
				if k.reportFilter.AllowReport(preport) {
					k.debouncer.Add(report.LifecycleEvent{NewPolicyReport: preport, OldPolicyReport: &report.PolicyReport{}, Type: report.Deleted})
				}
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			if item, ok := newObj.(*pr.PolicyReport); ok {
				preport := k.mapper.MapPolicyReport(item)

				var oreport *report.PolicyReport
				if oldItem, ok := oldObj.(*pr.PolicyReport); ok {
					oreport = k.mapper.MapPolicyReport(oldItem)
				}

				if k.reportFilter.AllowReport(preport) {
					k.debouncer.Add(report.LifecycleEvent{NewPolicyReport: preport, OldPolicyReport: oreport, Type: report.Updated})
				}
			}
		},
	})

	informer.Run(ctx.Done())
}

func (k *k8sPolicyReportClient) watchClusterPolicyReport(ctx context.Context, informer cache.SharedIndexInformer, crd string) {
	ctx = k.addErrorWatchHandler(ctx, informer, crd)

	go k.handleCRDRegistration(ctx, informer, crd)

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if item, ok := obj.(*pr.ClusterPolicyReport); ok {
				preport := k.mapper.MapClusterPolicyReport(item)
				if k.reportFilter.AllowReport(preport) {
					k.debouncer.Add(report.LifecycleEvent{NewPolicyReport: preport, OldPolicyReport: &report.PolicyReport{}, Type: report.Added})
				}
			}
		},
		DeleteFunc: func(obj interface{}) {
			if item, ok := obj.(*pr.ClusterPolicyReport); ok {
				preport := k.mapper.MapClusterPolicyReport(item)
				if k.reportFilter.AllowReport(preport) {
					k.debouncer.Add(report.LifecycleEvent{NewPolicyReport: preport, OldPolicyReport: &report.PolicyReport{}, Type: report.Deleted})
				}
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			if item, ok := newObj.(*pr.ClusterPolicyReport); ok {
				preport := k.mapper.MapClusterPolicyReport(item)

				var oreport *report.PolicyReport
				if oldItem, ok := oldObj.(*pr.ClusterPolicyReport); ok {
					oreport = k.mapper.MapClusterPolicyReport(oldItem)
				}

				if k.reportFilter.AllowReport(preport) {
					k.debouncer.Add(report.LifecycleEvent{NewPolicyReport: preport, OldPolicyReport: oreport, Type: report.Updated})
				}
			}
		},
	})

	informer.Run(ctx.Done())
}

func (k *k8sPolicyReportClient) handleCRDRegistration(ctx context.Context, informer cache.SharedIndexInformer, crd string) {
	ticker := time.NewTicker(k.restartWatchOnFailure)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if informer.HasSynced() {
				k.mx.Lock()
				k.found[crd] = crd
				k.mx.Unlock()

				log.Printf("[INFO] Resource registered: %s\n", crd)
				return
			}
		}
	}
}

func (k *k8sPolicyReportClient) addErrorWatchHandler(ctx context.Context, informer cache.SharedIndexInformer, crd string) context.Context {
	ctx, cancel := context.WithCancel(ctx)

	informer.SetWatchErrorHandler(func(c *cache.Reflector, err error) {
		k.mx.Lock()
		delete(k.found, crd)
		k.mx.Unlock()
		cancel()

		log.Printf("[WARNING] Resource registration failed: %s\n", crd)
	})

	return ctx
}

// NewPolicyReportAdapter new Adapter for Policy Report Kubernetes API
func NewPolicyReportClient(client versioned.Interface, mapper Mapper, restartWatchOnFailure time.Duration, reportFilter report.Filter) report.PolicyReportClient {
	return &k8sPolicyReportClient{
		client:                client,
		mapper:                mapper,
		mx:                    &sync.Mutex{},
		found:                 make(map[string]string),
		debouncer:             NewDebouncer(time.Minute),
		restartWatchOnFailure: restartWatchOnFailure,
		reportFilter:          reportFilter,
	}
}
