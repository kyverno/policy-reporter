package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
)

func TestEmailClient_MicrosoftGraphMailer(t *testing.T) {
	t.Run("fallback to microsoft_graph_mailer when graphAPI is disabled", func(t *testing.T) {
		cfg := &Config{
			EmailReports: EmailReports{
				GraphAPI: GraphAPI{
					Enabled: false,
				},
				MicrosoftGraphMailer: GraphAPI{
					Enabled:  true,
					Tenant:   "mail-tenant",
					ClientID: "mail-client",
					Password: "mail-secret",
					UserID:   "mail-user",
				},
			},
		}
		resolver := NewResolver(cfg, &rest.Config{})
		client := resolver.EmailClient()
		assert.NotNil(t, client)
	})

	t.Run("resolve clientSecret from k8s secret", func(t *testing.T) {
		cfg := &Config{
			Namespace: "default",
			EmailReports: EmailReports{
				GraphAPI: GraphAPI{
					Enabled:  true,
					Tenant:   "tenant",
					ClientID: "client",
					SecretRef: SecretRef{
						Name: "graph-secret",
						Key:  "my-key",
					},
					UserID: "user",
				},
			},
		}
		fakeClientset := fake.NewClientset(&corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "graph-secret",
				Namespace: "default",
			},
			Data: map[string][]byte{
				"my-key": []byte("resolved-secret-value"),
			},
		})

		resolver := NewResolver(cfg, &rest.Config{})
		resolver.k8sClient = fakeClientset

		client := resolver.EmailClient()
		assert.NotNil(t, client)
	})

	t.Run("resolve clientSecret from k8s secret default key", func(t *testing.T) {
		cfg := &Config{
			Namespace: "default",
			EmailReports: EmailReports{
				GraphAPI: GraphAPI{
					Enabled:  true,
					Tenant:   "tenant",
					ClientID: "client",
					SecretRef: SecretRef{
						Name: "graph-secret",
					},
					UserID: "user",
				},
			},
		}
		fakeClientset := fake.NewClientset(&corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "graph-secret",
				Namespace: "default",
			},
			Data: map[string][]byte{
				"secret": []byte("default-secret-value"),
			},
		})

		resolver := NewResolver(cfg, &rest.Config{})
		resolver.k8sClient = fakeClientset

		client := resolver.EmailClient()
		assert.NotNil(t, client)
	})

	t.Run("custom endpoints and alternatives", func(t *testing.T) {
		cfg := &Config{
			EmailReports: EmailReports{
				GraphAPI: GraphAPI{
					Enabled:         true,
					Tenant:          "tenant",
					ClientID:        "client-alt",
					Password:        "secret",
					UserID:          "user-alt",
					AzureADEndpoint: "https://login.microsoftonline.de",
					GraphEndpoint:   "https://graph.microsoft.de",
				},
			},
		}
		resolver := NewResolver(cfg, &rest.Config{})
		client := resolver.EmailClient()
		assert.NotNil(t, client)
	})
}
