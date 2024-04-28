package namespaces_test

import (
	"context"
	"errors"
	"testing"

	gocache "github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/kyverno/policy-reporter/pkg/kubernetes/namespaces"
)

func newFakeClient() v1.NamespaceInterface {
	return fake.NewSimpleClientset(
		&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "default",
				Labels: map[string]string{
					"team": "team-a",
					"name": "default",
				},
			},
		},
		&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "user",
				Labels: map[string]string{
					"team": "team-a",
					"name": "user",
				},
			},
		},
	).CoreV1().Namespaces()
}

type nsErrorClient struct {
	v1.NamespaceInterface
}

func (s *nsErrorClient) List(ctx context.Context, opts metav1.ListOptions) (*corev1.NamespaceList, error) {
	return nil, errors.New("error")
}

func TestClient(t *testing.T) {
	t.Run("read from api", func(t *testing.T) {
		client := namespaces.NewClient(newFakeClient(), gocache.New(gocache.DefaultExpiration, gocache.DefaultExpiration))

		list, err := client.List(context.Background(), map[string]string{"name": "default"})

		assert.Nil(t, err)
		assert.Equal(t, 1, len(list))
	})

	t.Run("read from cache", func(t *testing.T) {
		fake := newFakeClient()
		cache := gocache.New(gocache.NoExpiration, gocache.NoExpiration)

		client := namespaces.NewClient(fake, cache)

		list, err := client.List(context.Background(), map[string]string{"team": "team-a"})

		assert.Nil(t, err)
		assert.Equal(t, 2, len(list))

		fake.Create(context.Background(), &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "finance",
				Labels: map[string]string{
					"team": "team-a",
					"name": "finance",
				},
			},
		}, metav1.CreateOptions{})

		list, err = client.List(context.Background(), map[string]string{"team": "team-a"})

		assert.Nil(t, err)
		assert.Equal(t, 2, len(list))

		cache.Flush()

		list, err = client.List(context.Background(), map[string]string{"team": "team-a"})

		assert.Nil(t, err)
		assert.Equal(t, 3, len(list))
	})
	t.Run("return error", func(t *testing.T) {
		client := namespaces.NewClient(&nsErrorClient{NamespaceInterface: newFakeClient()}, gocache.New(gocache.DefaultExpiration, gocache.DefaultExpiration))

		_, err := client.List(context.Background(), map[string]string{"team": "team-a"})

		assert.NotNil(t, err)
		assert.Equal(t, "error", err.Error())
	})
}
