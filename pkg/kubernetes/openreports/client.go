package orclient

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/openreports/reports-api/apis/openreports.io/v1alpha1"
	"go.uber.org/zap"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/metadata"
	"k8s.io/client-go/metadata/metadatainformer"
	"k8s.io/client-go/tools/cache"

	prcache "github.com/kyverno/policy-reporter/pkg/cache"
	"github.com/kyverno/policy-reporter/pkg/report"
)

var (
	OpenreportsReport  = v1alpha1.SchemeGroupVersion.WithResource("reports")
	OpenreportsCReport = v1alpha1.SchemeGroupVersion.WithResource("clusterreports")
)

type openreportsClient struct {
	queue        *ORQueue
	metaClient   metadata.Interface
	synced       atomic.Bool
	mx           *sync.Mutex
	reportFilter *report.MetaFilter
	stopChan     chan struct{}
	periodicSync bool
	syncInterval time.Duration
	cache        prcache.Cache
}

func (k *openreportsClient) HasSynced() bool {
	return k.synced.Load()
}

func (k *openreportsClient) Stop() {
	k.mx.Lock()
	defer k.mx.Unlock()
	if k.stopChan != nil {
		select {
		case <-k.stopChan:
		default:
			close(k.stopChan)
		}
	}
}

func (k *openreportsClient) HasProcessedInitialReports() bool {
	return k.queue.initial.Complete()
}

func (k *openreportsClient) Sync(stopper chan struct{}) error {
	factory := metadatainformer.NewSharedInformerFactory(k.metaClient, 15*time.Minute)
	type sourceInformer struct {
		source       string
		informer     cache.SharedIndexInformer
		registration cache.ResourceEventHandlerRegistration
	}
	resources := []sourceInformer{}
	namespaced := factory.ForResource(OpenreportsReport).Informer()
	registration, err := k.configureInformer(namespaced, "reports")
	if err != nil {
		return err
	}
	resources = append(resources, sourceInformer{"reports", namespaced, registration})
	if !k.reportFilter.DisableClusterReports() {
		cluster := factory.ForResource(OpenreportsCReport).Informer()
		registration, err := k.configureInformer(cluster, "clusterreports")
		if err != nil {
			return err
		}
		resources = append(resources, sourceInformer{"clusterreports", cluster, registration})
	}
	factory.Start(stopper)
	for _, resource := range resources {
		synced := resource.informer.HasSynced
		if k.queue.initial != nil {
			synced = resource.registration.HasSynced
		}
		if !cache.WaitForCacheSync(stopper, synced) {
			return fmt.Errorf("failed to sync openreports %s", resource.source)
		}
		k.queue.initial.Seal(resource.source)
	}
	// Pending keys may have disappeared during an informer restart. Fetch them
	// again so NotFound can resolve their initial deletion instead of hanging.
	for _, key := range k.queue.initial.Pending() {
		k.queue.queue.Add(key)
	}
	k.synced.Store(true)
	zap.L().Info("openreports informer sync completed")
	return nil
}

func (k *openreportsClient) Run(worker int, stopper chan struct{}) error {
	k.mx.Lock()
	k.stopChan = stopper
	k.mx.Unlock()
	if err := k.Sync(stopper); err != nil {
		return err
	}

	// Periodic sync if enabled - just stop the informer to trigger restart
	if k.periodicSync {
		zap.L().Info("openreports periodic sync enabled",
			zap.String("interval", k.syncInterval.String()))
		ticker := time.NewTicker(k.syncInterval)
		go func() {
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					zap.L().Info("triggering openreports sync - clearing cache and stopping informer")
					if k.cache != nil {
						k.cache.Clear()
						zap.L().Info("result cache cleared for openreports periodic sync")
					}
					k.mx.Lock()
					select {
					case <-stopper:
					default:
						close(stopper)
					}
					k.mx.Unlock()
					return
				case <-stopper:
					ticker.Stop()
					return
				}
			}
		}()
	} else {
		zap.L().Info("openreports periodic sync disabled")
	}

	k.queue.Run(worker, stopper)
	return nil
}

func (k *openreportsClient) configureInformer(informer cache.SharedIndexInformer, source string) (cache.ResourceEventHandlerRegistration, error) {
	registration, err := informer.AddEventHandler(cache.ResourceEventHandlerDetailedFuncs{
		AddFunc: func(obj interface{}, initial bool) {
			if item, ok := obj.(*v1.PartialObjectMetadata); ok {
				if k.reportFilter.AllowReport(item) {
					if initial {
						key, err := cache.MetaNamespaceKeyFunc(item)
						if err == nil {
							k.queue.initial.Track(source, key)
						}
					}
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
		k.synced.Store(false)
	})

	return registration, err
}

// NewPolicyReportClient new Client for Policy Report Kubernetes API
func NewOpenreportsClient(metaClient metadata.Interface, reportFilter *report.MetaFilter, queue *ORQueue, periodicSync bool, syncInterval time.Duration, cache prcache.Cache, waitForInitialReports bool) report.PolicyReportClient {
	if waitForInitialReports {
		sources := []string{"reports"}
		if !reportFilter.DisableClusterReports() {
			sources = append(sources, "clusterreports")
		}
		queue.initial = report.NewInitialReports(sources...)
	}
	return &openreportsClient{
		metaClient:   metaClient,
		mx:           &sync.Mutex{},
		queue:        queue,
		reportFilter: reportFilter,
		periodicSync: periodicSync,
		syncInterval: syncInterval,
		cache:        cache,
	}
}
