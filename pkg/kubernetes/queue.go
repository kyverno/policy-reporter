package kubernetes

import (
	"context"
	"log"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	pr "github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/crd/client/clientset/versioned/typed/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/report"
)

type Queue struct {
	queue     workqueue.RateLimitingInterface
	client    v1alpha2.Wgpolicyk8sV1alpha2Interface
	debouncer Debouncer
	cache     Cache
}

func (q *Queue) Add(obj *v1.PartialObjectMetadata) error {
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		return err
	}

	q.queue.Add(key)

	return nil
}

func (q *Queue) Run(workers int, stopCh chan struct{}) {
	defer runtime.HandleCrash()

	defer q.queue.ShutDown()

	for i := 0; i < workers; i++ {
		go wait.Until(q.runWorker, time.Second, stopCh)
	}

	<-stopCh
}

func (q *Queue) runWorker() {
	for q.processNextItem() {
	}
}

func (q *Queue) processNextItem() bool {
	key, quit := q.queue.Get()
	if quit {
		return false
	}

	defer q.queue.Done(key)

	namespace, name, err := cache.SplitMetaNamespaceKey(key.(string))
	if err != nil {
		q.queue.Forget(key)
		return true
	}

	var polr pr.ReportInterface

	if namespace == "" {
		polr, err = q.client.ClusterPolicyReports().Get(context.Background(), name, v1.GetOptions{})
	} else {
		polr, err = q.client.PolicyReports(namespace).Get(context.Background(), name, v1.GetOptions{})
	}

	if errors.IsNotFound(err) {
		if namespace == "" {
			polr = &pr.ClusterPolicyReport{
				ObjectMeta: v1.ObjectMeta{
					Name: name,
				},
			}
		} else {
			polr = &pr.PolicyReport{
				ObjectMeta: v1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
			}
		}

		q.debouncer.Add(report.LifecycleEvent{Type: report.Deleted, PolicyReport: polr})
		q.cache.RemoveItem(key.(string))

		return true
	}

	event := report.Added
	if _, ok := q.cache.GetItem(key.(string)); ok {
		event = report.Updated
	}

	q.handleErr(err, key)

	q.debouncer.Add(report.LifecycleEvent{Type: event, PolicyReport: polr})
	q.cache.AddItem(key.(string), nil)

	return true
}

func (q *Queue) handleErr(err error, key interface{}) {
	if err == nil {
		q.queue.Forget(key)
		return
	}

	if q.queue.NumRequeues(key) < 5 {
		log.Printf("[ERROR] process report %v: %v", key, err)

		q.queue.AddRateLimited(key)
		return
	}

	q.queue.Forget(key)

	runtime.HandleError(err)
	log.Printf("[WARNING] Dropping report %q out of the queue: %v", key, err)
}

func NewQueue(cache Cache, debouncer Debouncer, queue workqueue.RateLimitingInterface, client v1alpha2.Wgpolicyk8sV1alpha2Interface) *Queue {
	return &Queue{
		debouncer: debouncer,
		queue:     queue,
		client:    client,
		cache:     cache,
	}
}
