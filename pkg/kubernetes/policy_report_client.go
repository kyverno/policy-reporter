package kubernetes

import (
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/metadata"
	"k8s.io/client-go/metadata/metadatainformer"
	"k8s.io/client-go/tools/cache"

	pr "github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/report"
)

var (
	polrResource  = pr.SchemeGroupVersion.WithResource("policyreports")
	cpolrResource = pr.SchemeGroupVersion.WithResource("clusterpolicyreports")
)

type k8sPolicyReportClient struct {
	queue         *Queue
	metaClient    metadata.Interface
	synced        bool
	mx            *sync.Mutex
	reportFilter  *report.MetaFilter
	stopChan      chan struct{}
	polrInformer  cache.SharedIndexInformer
	cpolrInformer cache.SharedIndexInformer
	periodicSync  bool
	syncInterval  time.Duration
}

func (k *k8sPolicyReportClient) HasSynced() bool {
	return k.synced
}

func (k *k8sPolicyReportClient) Stop() {
	close(k.stopChan)
}

func (k *k8sPolicyReportClient) Sync(stopper chan struct{}) error {
	factory := metadatainformer.NewSharedInformerFactory(k.metaClient, 15*time.Minute)

	k.polrInformer = k.configureInformer(factory.ForResource(polrResource).Informer())

	if !k.reportFilter.DisableClusterReports() {
		k.cpolrInformer = k.configureInformer(factory.ForResource(cpolrResource).Informer())
	}

	factory.Start(stopper)

	if !cache.WaitForCacheSync(stopper, k.polrInformer.HasSynced) {
		return fmt.Errorf("failed to sync policy reports")
	}

	if k.cpolrInformer != nil && !cache.WaitForCacheSync(stopper, k.cpolrInformer.HasSynced) {
		return fmt.Errorf("failed to sync cluster policy reports")
	}

	k.synced = true

	zap.L().Info("policy report informer sync completed")

	return nil
}

func (k *k8sPolicyReportClient) Run(worker int, stopper chan struct{}) error {
	k.stopChan = stopper
	if err := k.Sync(stopper); err != nil {
		return err
	}

	// Initial sync of existing reports
	go k.syncExistingReports()

	// Periodic sync if enabled
	if k.periodicSync {
		zap.L().Info("policy report periodic sync enabled",
			zap.String("interval", k.syncInterval.String()))
		ticker := time.NewTicker(k.syncInterval)
		go func() {
			for {
				select {
				case <-ticker.C:
					k.syncExistingReports()
				case <-stopper:
					ticker.Stop()
					return
				}
			}
		}()
	} else {
		zap.L().Info("policy report periodic sync disabled")
	}

	k.queue.Run(worker, stopper)
	return nil
}

func (k *k8sPolicyReportClient) syncExistingReports() {
	total := 0

	if k.polrInformer != nil {
		items := k.polrInformer.GetStore().List()
		for _, item := range items {
			if meta, ok := item.(*v1.PartialObjectMetadata); ok {
				if k.reportFilter.AllowReport(meta) {
					k.queue.Add(meta)
					total++
				}
			}
		}
	}

	if k.cpolrInformer != nil {
		items := k.cpolrInformer.GetStore().List()
		for _, item := range items {
			if meta, ok := item.(*v1.PartialObjectMetadata); ok {
				if k.reportFilter.AllowReport(meta) {
					k.queue.Add(meta)
					total++
				}
			}
		}
	}

	zap.L().Info("syncing existing policy reports", zap.Int("count", total))
}

func (k *k8sPolicyReportClient) configureInformer(informer cache.SharedIndexInformer) cache.SharedIndexInformer {
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if item, ok := obj.(*v1.PartialObjectMetadata); ok {
				if k.reportFilter.AllowReport(item) {
					k.queue.Add(item)
				}
			}
		},
		DeleteFunc: func(obj interface{}) {
			if item, ok := obj.(*v1.PartialObjectMetadata); ok {
				if k.reportFilter.AllowReport(item) {
					k.queue.Add(item)
				}
			}
		},
		UpdateFunc: func(_, newObj interface{}) {
			if item, ok := newObj.(*v1.PartialObjectMetadata); ok {
				if k.reportFilter.AllowReport(item) {
					k.queue.Add(item)
				}
			}
		},
	})

	informer.SetWatchErrorHandler(func(_ *cache.Reflector, _ error) {
		k.synced = false
	})

	return informer
}

// NewPolicyReportClient new Client for Policy Report Kubernetes API
func NewPolicyReportClient(metaClient metadata.Interface, reportFilter *report.MetaFilter, queue *Queue, periodicSync bool, syncInterval time.Duration) report.PolicyReportClient {
	return &k8sPolicyReportClient{
		metaClient:    metaClient,
		mx:            &sync.Mutex{},
		queue:         queue,
		reportFilter:  reportFilter,
		polrInformer:  nil,
		cpolrInformer: nil,
		periodicSync:  periodicSync,
		syncInterval:  syncInterval,
	}
}
