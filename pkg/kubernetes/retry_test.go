package kubernetes_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/kyverno/policy-reporter/pkg/kubernetes"
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

type ns struct {
	maxRetry int
	try      int
	err      bool
	v1.NamespaceInterface
}

func (s *ns) List(ctx context.Context, opts metav1.ListOptions) (*corev1.NamespaceList, error) {
	if !s.err {
		if s.try >= s.maxRetry {
			return s.NamespaceInterface.List(ctx, opts)
		}

		s.try++
	}

	return nil, errors.New("error")
}

func TestRetry(t *testing.T) {
	t.Run("direct success", func(t *testing.T) {
		client := &ns{NamespaceInterface: newFakeClient()}

		list, err := kubernetes.Retry(func() (*corev1.NamespaceList, error) {
			return client.List(context.Background(), metav1.ListOptions{})
		})

		assert.Nil(t, err)
		assert.Equal(t, 2, len(list.Items))
	})

	t.Run("retry success", func(t *testing.T) {
		client := &ns{maxRetry: 1, NamespaceInterface: newFakeClient()}

		list, err := kubernetes.Retry(func() (*corev1.NamespaceList, error) {
			return client.List(context.Background(), metav1.ListOptions{})
		})

		assert.Nil(t, err)
		assert.Equal(t, 2, len(list.Items))
	})

	t.Run("retry error", func(t *testing.T) {
		client := &ns{NamespaceInterface: newFakeClient(), err: true}

		_, err := kubernetes.Retry(func() (*corev1.NamespaceList, error) {
			return client.List(context.Background(), metav1.ListOptions{})
		})

		assert.NotNil(t, err)
	})

	t.Run("retry timeout", func(t *testing.T) {
		try := 0

		_, err := kubernetes.Retry(func() (any, error) {
			try++

			return nil, &kerr.StatusError{
				ErrStatus: metav1.Status{Reason: metav1.StatusReasonTimeout},
			}
		})

		assert.Equal(t, 5, try)
		assert.NotNil(t, err)
	})

	t.Run("retry server timeout", func(t *testing.T) {
		try := 0

		_, err := kubernetes.Retry(func() (any, error) {
			try++

			return nil, &kerr.StatusError{
				ErrStatus: metav1.Status{Reason: metav1.StatusReasonServerTimeout},
			}
		})

		assert.Equal(t, 5, try)
		assert.NotNil(t, err)
	})

	t.Run("retry service unavailable", func(t *testing.T) {
		try := 0

		_, err := kubernetes.Retry(func() (any, error) {
			try++

			return nil, &kerr.StatusError{
				ErrStatus: metav1.Status{Reason: metav1.StatusReasonServiceUnavailable},
			}
		})

		assert.Equal(t, 5, try)
		assert.NotNil(t, err)
	})

	t.Run("retry ignore other status", func(t *testing.T) {
		try := 0

		_, err := kubernetes.Retry(func() (any, error) {
			try++

			return nil, &kerr.StatusError{
				ErrStatus: metav1.Status{Reason: metav1.StatusReasonForbidden},
			}
		})

		assert.Equal(t, 1, try)
		assert.NotNil(t, err)
	})
}
