package kubernetes

import (
	"context"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

// ConfigMapAdapter provides simplified APIs for ConfigMap Resources
type ConfigMapAdapter interface {
	// GetConfig return a single ConfigMap by name if exist
	GetConfig(ctx context.Context, name string) (*apiv1.ConfigMap, error)
	// WatchConfigs calls its ConfigMapCallback whenever a ConfigMap was added, modified or deleted
	WatchConfigs(ctx context.Context, cb ConfigMapCallback) error
}

// ConfigMapCallback is used by WatchConfigs
type ConfigMapCallback = func(watch.EventType, *apiv1.ConfigMap)

type cmAdapter struct {
	api v1.ConfigMapInterface
}

func (c cmAdapter) GetConfig(ctx context.Context, name string) (*apiv1.ConfigMap, error) {
	return c.api.Get(ctx, name, metav1.GetOptions{})
}

func (c cmAdapter) WatchConfigs(ctx context.Context, cb ConfigMapCallback) error {
	for {
		watch, err := c.api.Watch(ctx, metav1.ListOptions{})
		if err != nil {
			return err
		}

		for event := range watch.ResultChan() {
			if cm, ok := event.Object.(*apiv1.ConfigMap); ok {
				cb(event.Type, cm)
			}
		}
	}
}

// NewConfigMapAdapter creates a new ConfigMapClient
func NewConfigMapAdapter(api v1.ConfigMapInterface) ConfigMapAdapter {
	return &cmAdapter{api}
}
