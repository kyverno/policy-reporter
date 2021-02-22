package kubernetes

import (
	"context"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type CoreClient interface {
	GetConfig(ctx context.Context, name string) (*apiv1.ConfigMap, error)
	WatchConfigs(ctx context.Context, cb ConfigMapCallback) error
}

type ConfigMapCallback = func(watch.EventType, *apiv1.ConfigMap)

type coreClient struct {
	cmClient v1.ConfigMapInterface
}

func (c coreClient) GetConfig(ctx context.Context, name string) (*apiv1.ConfigMap, error) {
	return c.cmClient.Get(ctx, name, metav1.GetOptions{})
}

func (c coreClient) WatchConfigs(ctx context.Context, cb ConfigMapCallback) error {
	watch, err := c.cmClient.Watch(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	for event := range watch.ResultChan() {
		if cm, ok := event.Object.(*apiv1.ConfigMap); ok {
			cb(event.Type, cm)
		}
	}

	return nil
}

func NewCoreClient(kubeconfig, namespace string) (CoreClient, error) {
	var config *rest.Config
	var err error

	if kubeconfig != "" {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	} else {
		config, err = rest.InClusterConfig()
	}
	if err != nil {
		return nil, err
	}

	client, err := v1.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &coreClient{
		cmClient: client.ConfigMaps(namespace),
	}, nil
}
