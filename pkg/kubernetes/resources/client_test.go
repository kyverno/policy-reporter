package resources

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/metadata/fake"
	gocache "zgo.at/zcache/v2"
)

func TestClient_Get(t *testing.T) {
	scheme := runtime.NewScheme()
	metav1.AddMetaToScheme(scheme)

	ref := &corev1.ObjectReference{
		APIVersion: "v1",
		Kind:       "Pod",
		Namespace:  "default",
		Name:       "nginx",
	}

	object := &metav1.PartialObjectMetadata{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nginx",
			Namespace: "default",
			Labels: map[string]string{
				"app":  "frontend",
				"tier": "web",
			},
		},
	}

	client := fake.NewSimpleMetadataClient(scheme, object)

	c := NewClient(
		client,
		gocache.New[string, map[string]string](
			gocache.NoExpiration,
			gocache.NoExpiration,
		),
	)

	got, err := c.Get(context.Background(), ref)

	require.NoError(t, err)
	assert.Equal(t, map[string]string{
		"app":  "frontend",
		"tier": "web",
	}, got)
}
