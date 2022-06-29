package webhook_test

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/kyverno/policy-reporter/pkg/target"
	"github.com/kyverno/policy-reporter/pkg/target/webhook"
)

var completeResult = report.Result{
	Message:  "validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/",
	Policy:   "require-requests-and-limits-required",
	Rule:     "autogen-check-for-requests-and-limits",
	Priority: report.WarningPriority,
	Status:   report.Fail,
	Severity: report.High,
	Category: "resources",
	Scored:   true,
	Resource: report.Resource{
		APIVersion: "v1",
		Kind:       "Deployment",
		Name:       "nginx",
		Namespace:  "default",
		UID:        "536ab69f-1b3c-4bd9-9ba4-274a56188409",
	},
}

type testClient struct {
	callback   func(req *http.Request) error
	statusCode int
}

func (c testClient) Do(req *http.Request) (*http.Response, error) {
	err := c.callback(req)

	return &http.Response{
		StatusCode: c.statusCode,
		Body:       io.NopCloser(strings.NewReader("")),
	}, err
}

func Test_UITarget(t *testing.T) {
	t.Run("Send", func(t *testing.T) {
		callback := func(req *http.Request) error {
			if contentType := req.Header.Get("Content-Type"); contentType != "application/json; charset=utf-8" {
				t.Errorf("Unexpected Content-Type: %s", contentType)
			}

			if agend := req.Header.Get("User-Agent"); agend != "Policy-Reporter" {
				t.Errorf("Unexpected Host: %s", agend)
			}

			if url := req.URL.String(); url != "http://localhost:8080/webhook" {
				t.Errorf("Unexpected Host: %s", url)
			}

			if value := req.Header.Get("X-Code"); value != "1234" {
				t.Errorf("Unexpected Header X-Code: %s", value)
			}

			return nil
		}

		client := webhook.NewClient("HTTP", "http://localhost:8080/webhook", map[string]string{"X-Code": "1234"}, false, &target.Filter{}, testClient{callback, 200})
		client.Send(completeResult)
	})
	t.Run("Name", func(t *testing.T) {
		client := webhook.NewClient("HTTP", "http://localhost:8080/webhook", map[string]string{"X-Code": "1234"}, false, &target.Filter{}, testClient{})

		if client.Name() != "HTTP" {
			t.Errorf("Unexpected Name %s", client.Name())
		}
	})
}
