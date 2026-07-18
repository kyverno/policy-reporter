package replicasets

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/kubernetes/typed/apps/v1"

	"github.com/kyverno/policy-reporter/pkg/kubernetes/retry"
)

type Client interface {
	Get(scope *corev1.ObjectReference) (*appsv1.ReplicaSet, error)
}

type k8sClient struct {
	client v1.ReplicaSetsGetter
}

func (c *k8sClient) Get(scope *corev1.ObjectReference) (*appsv1.ReplicaSet, error) {
	return retry.Retry(func() (*appsv1.ReplicaSet, error) {
		return c.client.ReplicaSets(scope.Namespace).Get(context.Background(), scope.Name, metav1.GetOptions{})
	})
}

func NewClient(client v1.ReplicaSetsGetter) Client {
	return &k8sClient{
		client: client,
	}
}
