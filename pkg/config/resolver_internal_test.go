package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"

	"github.com/kyverno/policy-reporter/pkg/email"
)

func TestEmailClient_GraphAPI(t *testing.T) {
	graphClientType := email.NewGraphAPIClient("", "", "", "", email.GraphAPIClientOptions{})

	t.Run("returns graph api client when enabled", func(t *testing.T) {
		cfg := &Config{
			EmailReports: EmailReports{
				GraphAPI: GraphAPI{
					Enabled:      true,
					Tenant:       "tenant",
					ClientID:     "client",
					ClientSecret: "secret",
					UserID:       "user",
				},
			},
		}
		resolver := NewResolver(cfg, &rest.Config{})

		assert.IsType(t, graphClientType, resolver.EmailClient())
	})

	t.Run("returns smtp client when graph api is disabled", func(t *testing.T) {
		cfg := &Config{
			EmailReports: EmailReports{
				GraphAPI: GraphAPI{Enabled: false},
			},
		}
		resolver := NewResolver(cfg, &rest.Config{})

		assert.IsType(t, &email.Client{}, resolver.EmailClient())
	})

	t.Run("custom endpoints", func(t *testing.T) {
		cfg := &Config{
			EmailReports: EmailReports{
				GraphAPI: GraphAPI{
					Enabled:         true,
					Tenant:          "tenant",
					ClientID:        "client-alt",
					ClientSecret:    "secret",
					UserID:          "user-alt",
					AzureADEndpoint: "https://login.microsoftonline.de",
					GraphEndpoint:   "https://graph.microsoft.de",
				},
			},
		}
		resolver := NewResolver(cfg, &rest.Config{})

		assert.IsType(t, graphClientType, resolver.EmailClient())
	})
}

func TestEmailReportControllerLoadsJobTemplate(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "job-template.yaml")
	err := os.WriteFile(path, []byte(`spec:
  template:
    spec:
      restartPolicy: Never
      containers:
        - name: policy-reporter
          image: local/policy-reporter:test
`), 0o600)
	assert.NoError(t, err)

	resolver := NewResolver(&Config{
		Namespace: "policy-reporter",
		EmailReports: EmailReports{CRD: EmailReportCRD{
			JobTemplate: path,
		}},
	}, &rest.Config{Host: "https://example.test"})
	resolver.k8sClient = fake.NewClientset()

	controller, err := resolver.EmailReportController()
	assert.NoError(t, err)
	assert.NotNil(t, controller)
}

func TestGraphAPIClientSecret(t *testing.T) {
	t.Run("resolves clientSecret from referenced secret", func(t *testing.T) {
		graphAPI := GraphAPI{
			ClientSecret: "inline-secret",
			SecretRef:    "graph-secret",
		}

		resolver := NewResolver(&Config{Namespace: "default"}, &rest.Config{})
		resolver.k8sClient = fake.NewClientset(&corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "graph-secret",
				Namespace: "default",
			},
			Data: map[string][]byte{
				"clientSecret": []byte("resolved-secret-value"),
			},
		})

		assert.Equal(t, "resolved-secret-value", resolver.graphAPIClientSecret(graphAPI))
	})

	t.Run("falls back to inline clientSecret without secretRef", func(t *testing.T) {
		graphAPI := GraphAPI{ClientSecret: "inline-secret"}

		resolver := NewResolver(&Config{Namespace: "default"}, &rest.Config{})

		assert.Equal(t, "inline-secret", resolver.graphAPIClientSecret(graphAPI))
	})

	t.Run("falls back to inline clientSecret when secret is missing", func(t *testing.T) {
		graphAPI := GraphAPI{
			ClientSecret: "inline-secret",
			SecretRef:    "missing-secret",
		}

		resolver := NewResolver(&Config{Namespace: "default"}, &rest.Config{})
		resolver.k8sClient = fake.NewClientset()

		assert.Equal(t, "inline-secret", resolver.graphAPIClientSecret(graphAPI))
	})

	t.Run("falls back to inline clientSecret when key is missing", func(t *testing.T) {
		graphAPI := GraphAPI{
			ClientSecret: "inline-secret",
			SecretRef:    "graph-secret",
		}

		resolver := NewResolver(&Config{Namespace: "default"}, &rest.Config{})
		resolver.k8sClient = fake.NewClientset(&corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "graph-secret",
				Namespace: "default",
			},
			Data: map[string][]byte{
				"other": []byte("value"),
			},
		})

		assert.Equal(t, "inline-secret", resolver.graphAPIClientSecret(graphAPI))
	})
}
