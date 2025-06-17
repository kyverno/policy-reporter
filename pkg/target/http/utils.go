package http

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"net"
	"net/http"
	"os"
	"time"

	"go.uber.org/zap"

	"github.com/kyverno/policy-reporter/pkg/openreports"
)

// CreateJSONRequest for the given configuration
func CreateJSONRequest(method, host string, payload interface{}) (*http.Request, error) {
	body := new(bytes.Buffer)

	json.NewEncoder(body).Encode(payload)

	req, err := http.NewRequest(method, host, body)
	if err != nil {
		zap.L().Error("failed to create request", zap.Error(err))
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("User-Agent", "Policy-Reporter")

	return req, nil
}

// ProcessHTTPResponse Logs Error or Success messages
func ProcessHTTPResponse(target string, resp *http.Response, err error) {
	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()

	if err != nil {
		zap.L().Error(target+": PUSH FAILED", zap.Error(err))
	} else if resp.StatusCode >= 400 {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)

		zap.L().Error(target+": PUSH FAILED", zap.Int("statusCode", resp.StatusCode), zap.String("body", buf.String()))
	} else {
		zap.L().Info(target + ": PUSH OK")
	}
}

func NewJSONResult(r openreports.ORResultAdapter) Result {
	res := Resource{}
	if r.HasResource() {
		resOb := r.GetResource()

		res.Namespace = resOb.Namespace
		res.APIVersion = resOb.APIVersion
		res.Kind = resOb.Kind
		res.Name = resOb.Name
		res.UID = string(resOb.UID)
	}
	return Result{
		Message:           r.Description,
		Policy:            r.Policy,
		Rule:              r.Rule,
		Status:            string(r.Result),
		Severity:          string(r.Severity),
		Category:          r.Category,
		Scored:            r.Scored,
		Properties:        r.Properties,
		Resource:          res,
		CreationTimestamp: time.Unix(r.Timestamp.Seconds, int64(r.Timestamp.Nanos)),
		Source:            r.Source,
	}
}

func NewClient(certificatePath string, skipTLS bool) *http.Client {
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 60 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: skipTLS,
		},
		Proxy: http.ProxyFromEnvironment,
	}

	client := &http.Client{
		Transport: NewLoggingRoundTripper(transport),
		Timeout:   30 * time.Second,
	}

	if certificatePath != "" {
		caCert, err := os.ReadFile(certificatePath)
		if err != nil {
			zap.L().Error("failed to read certificate", zap.String("path", certificatePath))
			return client
		}

		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		transport.TLSClientConfig.RootCAs = caCertPool
	}

	return client
}
