package helper

import (
	"crypto/x509"
	"fmt"
	"os"
)

// LoadCertPool reads the PEM encoded CA certificates in path into a new pool.
// It returns an error when the file has no valid PEM certificate, because an
// empty pool would make every TLS handshake fail with "unknown authority".
func LoadCertPool(path string) (*x509.CertPool, error) {
	caCert, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("no valid PEM certificate found in %s", path)
	}

	return pool, nil
}
