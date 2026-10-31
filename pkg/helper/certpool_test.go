package helper_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/kyverno/policy-reporter/pkg/helper"
)

func writeFile(t *testing.T, content []byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "ca.crt")
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func newCACertPEM(t *testing.T) []byte {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "test-ca"},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: true,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
}

func TestLoadCertPool(t *testing.T) {
	t.Run("valid certificate", func(t *testing.T) {
		pool, err := helper.LoadCertPool(writeFile(t, newCACertPEM(t)))
		assert.NoError(t, err)
		assert.NotNil(t, pool)
	})

	t.Run("missing file", func(t *testing.T) {
		pool, err := helper.LoadCertPool(filepath.Join(t.TempDir(), "missing.crt"))
		assert.Error(t, err)
		assert.Nil(t, pool)
	})

	t.Run("empty file", func(t *testing.T) {
		pool, err := helper.LoadCertPool(writeFile(t, nil))
		assert.ErrorContains(t, err, "no valid PEM certificate")
		assert.Nil(t, pool)
	})

	t.Run("no PEM certificate", func(t *testing.T) {
		pool, err := helper.LoadCertPool(writeFile(t, []byte("not a certificate")))
		assert.ErrorContains(t, err, "no valid PEM certificate")
		assert.Nil(t, pool)
	})
}
