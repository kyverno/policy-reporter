package resources

import (
	"context"
	"strings"
	"time"

	"github.com/go-openapi/inflect"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/metadata"
	gocache "zgo.at/zcache/v2"

	"github.com/kyverno/policy-reporter/pkg/kubernetes/retry"
)

type Client interface {
	Get(context.Context, *corev1.ObjectReference) (map[string]string, error)
}

type k8sClient struct {
	client metadata.Interface
	cache  *gocache.Cache[string, map[string]string]
}

func (c *k8sClient) Get(ctx context.Context, ref *corev1.ObjectReference) (map[string]string, error) {
	if ref == nil || ref.Name == "" || ref.Kind == "" || ref.APIVersion == "" {
		return nil, nil
	}

	groupVersion, err := schema.ParseGroupVersion(ref.APIVersion)
	if err != nil {
		return nil, err
	}

	gvr := groupVersion.WithResource(
		inflect.Pluralize(strings.ToLower(ref.Kind)),
	)

	cacheKey := strings.Join([]string{
		ref.APIVersion,
		ref.Kind,
		ref.Namespace,
		ref.Name,
		string(ref.UID),
	}, "/")

	if labels, ok := c.cache.Get(cacheKey); ok {
		return labels, nil
	}

	var resource metadata.ResourceInterface
	if ref.Namespace != "" {
		resource = c.client.Resource(gvr).Namespace(ref.Namespace)
	} else {
		resource = c.client.Resource(gvr)
	}

	object, err := retry.Retry(func() (*metav1.PartialObjectMetadata, error) {
		return resource.Get(ctx, ref.Name, metav1.GetOptions{})
	})
	if err != nil {
		return nil, err
	}

	resourceLabels := object.GetLabels()
	if resourceLabels == nil {
		resourceLabels = map[string]string{}
	}

	c.cache.SetWithExpire(cacheKey, resourceLabels, 5*time.Second)

	return resourceLabels, nil
}

func NewClient(
	client metadata.Interface,
	cache *gocache.Cache[string, map[string]string],
) Client {
	if cache == nil {
		cache = gocache.New[string, map[string]string](15*time.Second, 5*time.Second)
	}

	return &k8sClient{
		client: client,
		cache:  cache,
	}
}
