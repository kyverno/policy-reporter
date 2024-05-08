package pods

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/kyverno/policy-reporter/pkg/kubernetes"
)

type Client interface {
	Get(scope *corev1.ObjectReference) (*corev1.Pod, error)
}

type k8sClient struct {
	client v1.CoreV1Interface
}

func (c *k8sClient) Get(scope *corev1.ObjectReference) (*corev1.Pod, error) {
	return kubernetes.Retry(func() (*corev1.Pod, error) {
		return c.client.Pods(scope.Namespace).Get(context.Background(), scope.Name, metav1.GetOptions{})
	})
}

func NewClient(client v1.CoreV1Interface) Client {
	return &k8sClient{
		client: client,
	}
}
