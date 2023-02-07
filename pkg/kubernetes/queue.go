package kubernetes

import (
	"context"
	"log"
	"strings"
	"time"

	pr "github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/crd/client/clientset/versioned/typed/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/report"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

type Queue struct {
	queue     workqueue.RateLimitingInterface
	client    v1alpha2.Wgpolicyk8sV1alpha2Interface
	debouncer Debouncer
}

func (q *Queue) Add(event report.Event, obj *v1.PartialObjectMetadata) error {
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		return err
	}

	q.queue.Add(event.String() + ":" + key)

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

	event, namespace, name, err := splitKey(key)
	if err != nil {
		q.queue.Forget(key)
		return true
	}

	var polr pr.ReportInterface

	if event == report.Deleted {
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

		q.debouncer.Add(report.LifecycleEvent{Type: event, PolicyReport: polr})

		return true
	}

	if namespace == "" {
		polr, err = q.client.ClusterPolicyReports().Get(context.Background(), name, v1.GetOptions{})
	} else {
		polr, err = q.client.PolicyReports(namespace).Get(context.Background(), name, v1.GetOptions{})
	}

	q.handleErr(err, key)

	q.debouncer.Add(report.LifecycleEvent{Type: event, PolicyReport: polr})

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

func splitKey(key interface{}) (report.Event, string, string, error) {
	parts := strings.Split(key.(string), ":")

	namespace, name, err := cache.SplitMetaNamespaceKey(parts[1])

	var ev report.Event
	switch parts[0] {
	case report.Updated.String():
		ev = report.Updated
	case report.Deleted.String():
		ev = report.Deleted
	default:
		ev = report.Added
	}

	return ev, namespace, name, err
}

func NewQueue(debouncer Debouncer, queue workqueue.RateLimitingInterface, client v1alpha2.Wgpolicyk8sV1alpha2Interface) *Queue {
	return &Queue{
		debouncer: debouncer,
		queue:     queue,
		client:    client,
	}
}
