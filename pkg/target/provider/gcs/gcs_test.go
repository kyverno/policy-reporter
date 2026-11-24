package gcs

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/option"
)

type fakeGCS struct {
	server *httptest.Server

	mu          sync.Mutex
	tokenCalls  int
	uploadAuths []string
}

func newFakeGCS(t *testing.T) *fakeGCS {
	t.Helper()

	f := &fakeGCS{}
	f.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)

		f.mu.Lock()
		defer f.mu.Unlock()

		w.Header().Set("Content-Type", "application/json")

		if r.URL.Path == "/token" {
			f.tokenCalls++
			_, _ = w.Write([]byte(`{"access_token":"test-token","token_type":"Bearer","expires_in":3600}`))
			return
		}

		f.uploadAuths = append(f.uploadAuths, r.Header.Get("Authorization"))
		_, _ = w.Write([]byte(`{"bucket":"bucket","name":"report.json"}`))
	}))
	t.Cleanup(f.server.Close)

	return f
}

func (f *fakeGCS) serviceAccountJSON(t *testing.T) string {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})

	data, err := json.Marshal(map[string]string{
		"type":           "service_account",
		"project_id":     "project",
		"private_key_id": "key-id",
		"private_key":    string(keyPEM),
		"client_email":   "policy-reporter@project.iam.gserviceaccount.com",
		"client_id":      "123",
		"token_uri":      f.server.URL + "/token",
	})
	require.NoError(t, err)

	return string(data)
}

func (f *fakeGCS) upload(t *testing.T, c Client) {
	t.Helper()

	require.NotNil(t, c)
	require.NoError(t, c.Upload(bytes.NewBufferString(`{"policy":"test"}`), "prefix/report.json"))
}

func (f *fakeGCS) endpoint() option.ClientOption {
	return option.WithEndpoint(f.server.URL + "/storage/v1/")
}

func TestUploadUsesServiceAccountCredentials(t *testing.T) {
	gcs := newFakeGCS(t)

	gcs.upload(t, newClient(context.Background(), gcs.serviceAccountJSON(t), "bucket", gcs.endpoint()))

	assert.Equal(t, 1, gcs.tokenCalls)
	require.NotEmpty(t, gcs.uploadAuths)
	for _, auth := range gcs.uploadAuths {
		assert.Equal(t, "Bearer test-token", auth)
	}
}

func TestUploadUsesApplicationDefaultCredentials(t *testing.T) {
	gcs := newFakeGCS(t)

	path := filepath.Join(t.TempDir(), "credentials.json")
	require.NoError(t, os.WriteFile(path, []byte(gcs.serviceAccountJSON(t)), 0o600))
	t.Setenv("GOOGLE_APPLICATION_CREDENTIALS", path)

	gcs.upload(t, newClient(context.Background(), "", "bucket", gcs.endpoint()))

	assert.Equal(t, 1, gcs.tokenCalls)
	require.NotEmpty(t, gcs.uploadAuths)
	for _, auth := range gcs.uploadAuths {
		assert.Equal(t, "Bearer test-token", auth)
	}
}

func TestNewClientRejectsInvalidCredentials(t *testing.T) {
	assert.Nil(t, NewClient(context.Background(), `{"type": "authorized_user"}`, "bucket"))
	assert.Nil(t, NewClient(context.Background(), strings.Repeat("{", 2), "bucket"))
}
