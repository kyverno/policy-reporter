package kubernetes

import (
	"fmt"
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
	debouncer    Debouncer
	fatcory      externalversions.SharedInformerFactory
	v1alpha2     v1alpha2.Interface
	synced       bool
	mapper       Mapper
	mx           *sync.Mutex
	reportFilter report.Filter
}

func (k *k8sPolicyReportClient) HasSynced() bool {
	return k.synced
}

func (k *k8sPolicyReportClient) Run(stopper chan struct{}) error {
	var cpolrInformer cache.SharedIndexInformer

	polrInformer := k.configurePolicyReport()

	if !k.reportFilter.DisableClusterReports() {
		cpolrInformer = k.configureClusterPolicyReport()
	}

	k.fatcory.Start(stopper)

	if !cache.WaitForCacheSync(stopper, polrInformer.HasSynced) {
		return fmt.Errorf("failed to sync policy reports")
	}

	if cpolrInformer != nil && !cache.WaitForCacheSync(stopper, cpolrInformer.HasSynced) {
		return fmt.Errorf("failed to sync cluster policy reports")
	}

	k.synced = true

	return nil
}

func (k *k8sPolicyReportClient) configurePolicyReport() cache.SharedIndexInformer {
	polrInformer := k.v1alpha2.PolicyReports().Informer()
	polrInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if item, ok := obj.(*pr.PolicyReport); ok {
				preport := k.mapper.MapPolicyReport(item)
				if k.reportFilter.AllowReport(preport) {
					k.debouncer.Add(report.LifecycleEvent{NewPolicyReport: preport, OldPolicyReport: report.PolicyReport{}, Type: report.Added})
				}
			}
		},
		DeleteFunc: func(obj interface{}) {
			if item, ok := obj.(*pr.PolicyReport); ok {
				preport := k.mapper.MapPolicyReport(item)
				if k.reportFilter.AllowReport(preport) {
					k.debouncer.Add(report.LifecycleEvent{NewPolicyReport: preport, OldPolicyReport: report.PolicyReport{}, Type: report.Deleted})
				}
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			if item, ok := newObj.(*pr.PolicyReport); ok {
				preport := k.mapper.MapPolicyReport(item)

				var oreport report.PolicyReport
				if oldItem, ok := oldObj.(*pr.PolicyReport); ok {
					oreport = k.mapper.MapPolicyReport(oldItem)
				}

				if k.reportFilter.AllowReport(preport) {
					k.debouncer.Add(report.LifecycleEvent{NewPolicyReport: preport, OldPolicyReport: oreport, Type: report.Updated})
				}
			}
		},
	})

	polrInformer.SetWatchErrorHandler(func(_ *cache.Reflector, _ error) {
		k.synced = false
	})

	return polrInformer
}

func (k *k8sPolicyReportClient) configureClusterPolicyReport() cache.SharedIndexInformer {
	cpolrInformer := k.v1alpha2.ClusterPolicyReports().Informer()
	cpolrInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if item, ok := obj.(*pr.ClusterPolicyReport); ok {
				preport := k.mapper.MapClusterPolicyReport(item)
				if k.reportFilter.AllowReport(preport) {
					k.debouncer.Add(report.LifecycleEvent{NewPolicyReport: preport, OldPolicyReport: report.PolicyReport{}, Type: report.Added})
				}
			}
		},
		DeleteFunc: func(obj interface{}) {
			if item, ok := obj.(*pr.ClusterPolicyReport); ok {
				preport := k.mapper.MapClusterPolicyReport(item)
				if k.reportFilter.AllowReport(preport) {
					k.debouncer.Add(report.LifecycleEvent{NewPolicyReport: preport, OldPolicyReport: report.PolicyReport{}, Type: report.Deleted})
				}
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			if item, ok := newObj.(*pr.ClusterPolicyReport); ok {
				preport := k.mapper.MapClusterPolicyReport(item)

				var oreport report.PolicyReport
				if oldItem, ok := oldObj.(*pr.ClusterPolicyReport); ok {
					oreport = k.mapper.MapClusterPolicyReport(oldItem)
				}

				if k.reportFilter.AllowReport(preport) {
					k.debouncer.Add(report.LifecycleEvent{NewPolicyReport: preport, OldPolicyReport: oreport, Type: report.Updated})
				}
			}
		},
	})

	cpolrInformer.SetWatchErrorHandler(func(_ *cache.Reflector, _ error) {
		k.synced = false
	})

	return cpolrInformer
}

// NewPolicyReportAdapter new Adapter for Policy Report Kubernetes API
func NewPolicyReportClient(client versioned.Interface, mapper Mapper, reportFilter report.Filter, publisher report.EventPublisher) report.PolicyReportClient {
	fatcory := externalversions.NewSharedInformerFactory(client, time.Hour)
	v1alpha2 := fatcory.Wgpolicyk8s().V1alpha2()

	return &k8sPolicyReportClient{
		fatcory:      fatcory,
		v1alpha2:     v1alpha2,
		mapper:       mapper,
		mx:           &sync.Mutex{},
		debouncer:    NewDebouncer(time.Minute, publisher),
		reportFilter: reportFilter,
	}
}
