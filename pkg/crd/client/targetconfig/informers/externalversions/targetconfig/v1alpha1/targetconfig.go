/*
Copyright The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by informer-gen. DO NOT EDIT.

package v1alpha1

import (
	"context"
	time "time"

	targetconfigv1alpha1 "github.com/kyverno/policy-reporter/pkg/crd/api/targetconfig/v1alpha1"
	versioned "github.com/kyverno/policy-reporter/pkg/crd/client/targetconfig/clientset/versioned"
	internalinterfaces "github.com/kyverno/policy-reporter/pkg/crd/client/targetconfig/informers/externalversions/internalinterfaces"
	v1alpha1 "github.com/kyverno/policy-reporter/pkg/crd/client/targetconfig/listers/targetconfig/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// TargetConfigInformer provides access to a shared informer and lister for
// TargetConfigs.
type TargetConfigInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.TargetConfigLister
}

type targetConfigInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewTargetConfigInformer constructs a new informer for TargetConfig type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewTargetConfigInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredTargetConfigInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredTargetConfigInformer constructs a new informer for TargetConfig type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredTargetConfigInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.Wgpolicyk8sV1alpha1().TargetConfigs(namespace).List(context.TODO(), options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.Wgpolicyk8sV1alpha1().TargetConfigs(namespace).Watch(context.TODO(), options)
			},
		},
		&targetconfigv1alpha1.TargetConfig{},
		resyncPeriod,
		indexers,
	)
}

func (f *targetConfigInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredTargetConfigInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *targetConfigInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&targetconfigv1alpha1.TargetConfig{}, f.defaultInformer)
}

func (f *targetConfigInformer) Lister() v1alpha1.TargetConfigLister {
	return v1alpha1.NewTargetConfigLister(f.Informer().GetIndexer())
}
