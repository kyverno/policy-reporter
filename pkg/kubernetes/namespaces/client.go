package namespaces

import (
	"context"

	gocache "github.com/patrickmn/go-cache"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/kyverno/policy-reporter/pkg/helper"
	"github.com/kyverno/policy-reporter/pkg/kubernetes"
)

type Client interface {
	List(context.Context, map[string]string) ([]string, error)
}

type k8sClient struct {
	client v1.NamespaceInterface
	cache  *gocache.Cache
}

func (c *k8sClient) List(ctx context.Context, selector map[string]string) ([]string, error) {
	labelSelector := metav1.FormatLabelSelector(&metav1.LabelSelector{MatchLabels: selector})
	if cached, ok := c.cache.Get(labelSelector); ok {
		return cached.([]string), nil
	}

	list, err := kubernetes.Retry(func() ([]string, error) {
		namespaces, err := c.client.List(ctx, metav1.ListOptions{LabelSelector: labelSelector})
		if err != nil {
			return nil, err
		}

		return helper.Map(namespaces.Items, func(ns corev1.Namespace) string {
			return ns.Name
		}), nil
	})
	if err != nil {
		return nil, err
	}

	c.cache.Set(labelSelector, list, 0)

	return list, nil
}

func NewClient(secretClient v1.NamespaceInterface, cache *gocache.Cache) Client {
	return &k8sClient{
		client: secretClient,
		cache:  cache,
	}
}
