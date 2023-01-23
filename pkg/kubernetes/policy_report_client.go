package kubernetes

import (
	"fmt"
	"sync"
	"time"

	"github.com/kyverno/policy-reporter/pkg/report"

	pr "github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/crd/client/clientset/versioned"
	"github.com/kyverno/policy-reporter/pkg/crd/client/informers/externalversions"
	"github.com/kyverno/policy-reporter/pkg/crd/client/informers/externalversions/policyreport/v1alpha2"
	"k8s.io/client-go/tools/cache"
)

type k8sPolicyReportClient struct {
	debouncer    Debouncer
	fatcory      externalversions.SharedInformerFactory
	v1alpha2     v1alpha2.Interface
	synced       bool
	mx           *sync.Mutex
	reportFilter *report.Filter
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
				if k.reportFilter.AllowReport(item) {
					k.debouncer.Add(report.LifecycleEvent{NewPolicyReport: item, OldPolicyReport: nil, Type: report.Added})
				}
			}
		},
		DeleteFunc: func(obj interface{}) {
			if item, ok := obj.(*pr.PolicyReport); ok {
				if k.reportFilter.AllowReport(item) {
					k.debouncer.Add(report.LifecycleEvent{NewPolicyReport: item, OldPolicyReport: nil, Type: report.Deleted})
				}
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			if item, ok := newObj.(*pr.PolicyReport); ok {
				oldItem := oldObj.(*pr.PolicyReport)

				if k.reportFilter.AllowReport(item) {
					k.debouncer.Add(report.LifecycleEvent{NewPolicyReport: item, OldPolicyReport: oldItem, Type: report.Updated})
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
				if k.reportFilter.AllowReport(item) {
					k.debouncer.Add(report.LifecycleEvent{NewPolicyReport: item, OldPolicyReport: nil, Type: report.Added})
				}
			}
		},
		DeleteFunc: func(obj interface{}) {
			if item, ok := obj.(*pr.ClusterPolicyReport); ok {
				if k.reportFilter.AllowReport(item) {
					k.debouncer.Add(report.LifecycleEvent{NewPolicyReport: item, OldPolicyReport: nil, Type: report.Deleted})
				}
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			if item, ok := newObj.(*pr.ClusterPolicyReport); ok {
				oldItem := oldObj.(*pr.ClusterPolicyReport)

				if k.reportFilter.AllowReport(item) {
					k.debouncer.Add(report.LifecycleEvent{NewPolicyReport: item, OldPolicyReport: oldItem, Type: report.Updated})
				}
			}
		},
	})

	cpolrInformer.SetWatchErrorHandler(func(_ *cache.Reflector, _ error) {
		k.synced = false
	})

	return cpolrInformer
}

// NewPolicyReportClient new Client for Policy Report Kubernetes API
func NewPolicyReportClient(client versioned.Interface, reportFilter *report.Filter, publisher report.EventPublisher) report.PolicyReportClient {
	fatcory := externalversions.NewSharedInformerFactory(client, time.Hour)
	v1alpha2 := fatcory.Wgpolicyk8s().V1alpha2()

	return &k8sPolicyReportClient{
		fatcory:      fatcory,
		v1alpha2:     v1alpha2,
		mx:           &sync.Mutex{},
		debouncer:    NewDebouncer(time.Minute, publisher),
		reportFilter: reportFilter,
	}
}
