package ui_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/target"
	"github.com/kyverno/policy-reporter/pkg/target/ui"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var seconds = time.Date(2021, time.February, 23, 15, 10, 0, 0, time.UTC).Unix()

var completeResult = v1alpha2.PolicyReportResult{
	Message:   "validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/",
	Policy:    "require-requests-and-limits-required",
	Rule:      "autogen-check-for-requests-and-limits",
	Timestamp: v1.Timestamp{Seconds: seconds},
	Priority:  v1alpha2.WarningPriority,
	Result:    v1alpha2.StatusFail,
	Severity:  v1alpha2.SeverityHigh,
	Category:  "resources",
	Scored:    true,
	Source:    "Kyverno",
	Resources: []corev1.ObjectReference{{
		APIVersion: "v1",
		Kind:       "Deployment",
		Name:       "nginx",
		Namespace:  "default",
		UID:        "536ab69f-1b3c-4bd9-9ba4-274a56188409",
	}},
	Properties: map[string]string{"version": "1.2.0"},
}

type testClient struct {
	callback   func(req *http.Request)
	statusCode int
}

func (c testClient) Do(req *http.Request) (*http.Response, error) {
	c.callback(req)

	return &http.Response{
		StatusCode: c.statusCode,
	}, nil
}

func Test_UITarget(t *testing.T) {
	t.Run("Send", func(t *testing.T) {
		callback := func(req *http.Request) {
			if contentType := req.Header.Get("Content-Type"); contentType != "application/json; charset=utf-8" {
				t.Errorf("Unexpected Content-Type: %s", contentType)
			}

			if agend := req.Header.Get("User-Agent"); agend != "Policy-Reporter" {
				t.Errorf("Unexpected Host: %s", agend)
			}

			if url := req.URL.String(); url != "http://localhost:8080/api/push" {
				t.Errorf("Unexpected Host: %s", url)
			}
		}

		client := ui.NewClient(ui.Options{
			ClientOptions: target.ClientOptions{
				Name: "UI",
			},
			Host:       "http://localhost:8080",
			HTTPClient: testClient{callback, 200},
		})
		client.Send(completeResult)
	})
	t.Run("Name", func(t *testing.T) {
		client := ui.NewClient(ui.Options{
			ClientOptions: target.ClientOptions{
				Name: "UI",
			},
			Host:       "http://localhost:8080",
			HTTPClient: testClient{},
		})

		if client.Name() != "UI" {
			t.Errorf("Unexpected Name %s", client.Name())
		}
	})
}
