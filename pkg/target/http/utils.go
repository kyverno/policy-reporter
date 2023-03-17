package http

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
)

// CreateJSONRequest for the given configuration
func CreateJSONRequest(target, method, host string, payload interface{}) (*http.Request, error) {
	body := new(bytes.Buffer)

	json.NewEncoder(body).Encode(payload)

	req, err := http.NewRequest(method, host, body)
	if err != nil {
		zap.L().Error(target+": PUSH FAILED", zap.Error(err))
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

func NewJSONResult(r v1alpha2.PolicyReportResult) Result {
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
		Message:           r.Message,
		Policy:            r.Policy,
		Rule:              r.Rule,
		Priority:          r.Priority.String(),
		Status:            string(r.Result),
		Severity:          string(r.Severity),
		Category:          r.Category,
		Scored:            r.Scored,
		Resource:          res,
		CreationTimestamp: time.Unix(r.Timestamp.Seconds, int64(r.Timestamp.Nanos)),
	}
}

func NewClient(certificatePath string, skipTLS bool) *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: skipTLS,
	}

	client := &http.Client{
		Transport: transport,
	}

	if certificatePath != "" {
		caCert, err := ioutil.ReadFile(certificatePath)
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
