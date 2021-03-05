package kubernetes_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/fjogeleit/policy-reporter/pkg/kubernetes"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/fake"
	clientv1 "k8s.io/client-go/kubernetes/typed/core/v1"
	testcore "k8s.io/client-go/testing"
)

var configMap = &v1.ConfigMap{
	TypeMeta: metav1.TypeMeta{
		Kind:       "ConfigMap",
		APIVersion: "v1",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name: "policy-reporter-priorities",
	},
	Data: map[string]string{
		"default": "warning",
	},
}

func Test_GetConfigMap(t *testing.T) {
	_, cmAPI := newFakeAPI()
	cmAPI.Create(context.Background(), configMap, metav1.CreateOptions{})

	cmClient := kubernetes.NewConfigMapAdapter(cmAPI)

	cm, err := cmClient.GetConfig(context.Background(), "policy-reporter-priorities")
	if err != nil {
		t.Fatalf("Unexpected Error: %s", err)
	}

	if cm.Name != "policy-reporter-priorities" {
		t.Error("Unexpted ConfigMapReturned")
	}
	if priority, ok := cm.Data["default"]; !ok || priority != "warning" {
		t.Error("Unexpted default priority")
	}
}

func Test_WatchConfigMap(t *testing.T) {
	client, cmAPI := newFakeAPI()

	watcher := watch.NewFake()
	client.PrependWatchReactor("configmaps", testcore.DefaultWatchReactor(watcher, nil))

	cmClient := kubernetes.NewConfigMapAdapter(cmAPI)

	wg := sync.WaitGroup{}
	wg.Add(1)

	go cmClient.WatchConfigs(context.Background(), func(et watch.EventType, cm *v1.ConfigMap) {
		defer wg.Done()

		if cm.Name != "policy-reporter-priorities" {
			t.Error("Unexpted ConfigMapReturned")
		}
		if priority, ok := cm.Data["default"]; !ok || priority != "warning" {
			t.Error("Unexpted default priority")
		}
	})

	watcher.Add(configMap)

	wg.Wait()
}

func Test_WatchConfigMapError(t *testing.T) {
	client, cmAPI := newFakeAPI()
	client.PrependWatchReactor("configmaps", testcore.DefaultWatchReactor(watch.NewFake(), errors.New("")))

	cmClient := kubernetes.NewConfigMapAdapter(cmAPI)

	err := cmClient.WatchConfigs(context.Background(), func(et watch.EventType, cm *v1.ConfigMap) {})
	if err == nil {
		t.Error("Watch Error should stop execution")
	}
}

func newFakeAPI() (*fake.Clientset, clientv1.ConfigMapInterface) {
	client := fake.NewSimpleClientset()
	return client, client.CoreV1().ConfigMaps("policy-reporter")
}
