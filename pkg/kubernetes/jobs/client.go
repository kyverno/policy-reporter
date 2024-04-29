package jobs

import (
	"context"

	gocache "github.com/patrickmn/go-cache"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/kubernetes/typed/batch/v1"

	"github.com/kyverno/policy-reporter/pkg/kubernetes"
)

type Client interface {
	Get(scope *corev1.ObjectReference) (*batchv1.Job, error)
}

type k8sClient struct {
	client v1.BatchV1Interface
	cache  *gocache.Cache
}

func (c *k8sClient) Get(scope *corev1.ObjectReference) (*batchv1.Job, error) {
	if cached, ok := c.cache.Get(string(scope.UID)); ok {
		return cached.(*batchv1.Job), nil
	}

	pod, err := kubernetes.Retry(func() (*batchv1.Job, error) {
		return c.client.Jobs(scope.Namespace).Get(context.Background(), scope.Name, metav1.GetOptions{})
	})
	if err != nil {
		return nil, err
	}

	c.cache.Set(string(scope.UID), pod, 0)

	return pod, nil
}

func NewClient(client v1.BatchV1Interface, cache *gocache.Cache) Client {
	return &k8sClient{
		client: client,
		cache:  cache,
	}
}
