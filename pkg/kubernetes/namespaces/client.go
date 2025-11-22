package namespaces

import (
	"context"
	"strings"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	gocache "zgo.at/zcache/v2"

	"github.com/kyverno/policy-reporter/pkg/helper"
	"github.com/kyverno/policy-reporter/pkg/kubernetes"
)

type Client interface {
	List(context.Context, map[string]string) ([]string, error)
}

type k8sClient struct {
	client v1.NamespaceInterface
	cache  *gocache.Cache[string, []string]
}

func (c *k8sClient) List(ctx context.Context, selector map[string]string) ([]string, error) {
	s := labels.NewSelector()
	for l, v := range selector {
		var err error
		var req *labels.Requirement

		if strings.Contains(v, ",") {
			req, err = labels.NewRequirement(l, selection.In, helper.Map(strings.Split(v, ","), func(val string) string {
				return strings.TrimSpace(val)
			}))
		} else if v == "*" {
			req, err = labels.NewRequirement(l, selection.Exists, nil)
		} else if v == "!*" {
			req, err = labels.NewRequirement(l, selection.DoesNotExist, nil)
		} else {
			req, err = labels.NewRequirement(l, selection.Equals, []string{v})
		}
		if err != nil {
			zap.L().Error("failed to create selector requirement", zap.Error(err), zap.String("label", l), zap.String("value", v))
			continue
		}

		s = s.Add(*req)
	}

	zap.L().Debug("created label selector for namespace resolution", zap.String("selector", s.String()))

	if cached, ok := c.cache.Get(s.String()); ok {
		return cached, nil
	}

	list, err := kubernetes.Retry(func() ([]string, error) {
		namespaces, err := c.client.List(ctx, metav1.ListOptions{LabelSelector: s.String()})
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

	c.cache.Set(s.String(), list)

	return list, nil
}

func NewClient(secretClient v1.NamespaceInterface, cache *gocache.Cache[string, []string]) Client {
	return &k8sClient{
		client: secretClient,
		cache:  cache,
	}
}
