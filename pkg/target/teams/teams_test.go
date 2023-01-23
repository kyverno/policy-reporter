package teams_test

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/target"
	"github.com/kyverno/policy-reporter/pkg/target/teams"
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

var minimalErrorResult = v1alpha2.PolicyReportResult{
	Message:  "validation error: label required. Rule app-label-required failed at path /spec/template/spec/containers/0/resources/requests/",
	Policy:   "app-label-requirement",
	Priority: v1alpha2.ErrorPriority,
	Result:   v1alpha2.StatusFail,
	Scored:   true,
}

var minimalResult = v1alpha2.PolicyReportResult{
	Message:  "validation error: label required. Rule app-label-required failed at path /spec/template/spec/containers/0/resources/requests/",
	Policy:   "app-label-requirement",
	Priority: v1alpha2.CriticalPriority,
	Result:   v1alpha2.StatusFail,
	Scored:   true,
}

var minimalInfoResult = v1alpha2.PolicyReportResult{
	Message:  "validation error: label required. Rule app-label-required failed at path /spec/template/spec/containers/0/resources/requests/",
	Policy:   "app-label-requirement",
	Priority: v1alpha2.InfoPriority,
	Result:   v1alpha2.StatusFail,
	Scored:   true,
}

var minimalDebugResult = v1alpha2.PolicyReportResult{
	Message:  "validation error: label required. Rule app-label-required failed at path /spec/template/spec/containers/0/resources/requests/",
	Policy:   "app-label-requirement",
	Priority: v1alpha2.DebugPriority,
	Result:   v1alpha2.StatusFail,
	Scored:   true,
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

func Test_TeamsTarget(t *testing.T) {
	t.Run("Send Complete Result", func(t *testing.T) {
		callback := func(req *http.Request) {
			if contentType := req.Header.Get("Content-Type"); contentType != "application/json; charset=utf-8" {
				t.Errorf("Unexpected Content-Type: %s", contentType)
			}

			if agend := req.Header.Get("User-Agent"); agend != "Policy-Reporter" {
				t.Errorf("Unexpected Host: %s", agend)
			}

			if url := req.URL.String(); url != "http://hook.teams:80" {
				t.Errorf("Unexpected Host: %s", url)
			}

			payload := make(map[string]interface{})

			err := json.NewDecoder(req.Body).Decode(&payload)
			if err != nil {
				t.Fatal(err)
			}

			if payload["themeColor"] != "f2c744" {
				t.Errorf("Unexpected ThemeColor %s", payload["themeColor"])
			}
		}

		client := teams.NewClient(teams.Options{
			ClientOptions: target.ClientOptions{
				Name: "Teams",
			},
			Webhook:      "http://hook.teams:80",
			CustomFields: map[string]string{"Cluster": "Name"},
			HTTPClient:   testClient{callback, 200},
		})
		client.Send(completeResult)
	})

	t.Run("Send Minimal Result", func(t *testing.T) {
		callback := func(req *http.Request) {
			if contentType := req.Header.Get("Content-Type"); contentType != "application/json; charset=utf-8" {
				t.Errorf("Unexpected Content-Type: %s", contentType)
			}

			if agend := req.Header.Get("User-Agent"); agend != "Policy-Reporter" {
				t.Errorf("Unexpected Host: %s", agend)
			}

			if url := req.URL.String(); url != "http://hook.teams:80" {
				t.Errorf("Unexpected Host: %s", url)
			}

			payload := make(map[string]interface{})

			err := json.NewDecoder(req.Body).Decode(&payload)
			if err != nil {
				t.Fatal(err)
			}

			if payload["themeColor"] != "b80707" {
				t.Errorf("Unexpected ThemeColor %s", payload["themeColor"])
			}
		}

		client := teams.NewClient(teams.Options{
			ClientOptions: target.ClientOptions{
				Name: "Teams",
			},
			Webhook:      "http://hook.teams:80",
			CustomFields: map[string]string{"Cluster": "Name"},
			HTTPClient:   testClient{callback, 200},
		})
		client.Send(minimalResult)
	})
	t.Run("Send Minimal InfoResult", func(t *testing.T) {
		callback := func(req *http.Request) {
			payload := make(map[string]interface{})

			err := json.NewDecoder(req.Body).Decode(&payload)
			if err != nil {
				t.Fatal(err)
			}

			if payload["themeColor"] != "36a64f" {
				t.Errorf("Unexpected ThemeColor %s", payload["themeColor"])
			}
		}

		client := teams.NewClient(teams.Options{
			ClientOptions: target.ClientOptions{
				Name: "Teams",
			},
			Webhook:      "http://hook.teams:80",
			CustomFields: map[string]string{"Cluster": "Name"},
			HTTPClient:   testClient{callback, 200},
		})
		client.Send(minimalInfoResult)
	})
	t.Run("Send Minimal ErrorResult", func(t *testing.T) {
		callback := func(req *http.Request) {
			payload := make(map[string]interface{})

			err := json.NewDecoder(req.Body).Decode(&payload)
			if err != nil {
				t.Fatal(err)
			}

			if payload["themeColor"] != "e20b0b" {
				t.Errorf("Unexpected ThemeColor %s", payload["themeColor"])
			}
		}

		client := teams.NewClient(teams.Options{
			ClientOptions: target.ClientOptions{
				Name: "Teams",
			},
			Webhook:      "http://hook.teams:80",
			CustomFields: map[string]string{"Cluster": "Name"},
			HTTPClient:   testClient{callback, 200},
		})
		client.Send(minimalErrorResult)
	})
	t.Run("Send Minimal Debug Result", func(t *testing.T) {
		callback := func(req *http.Request) {
			if contentType := req.Header.Get("Content-Type"); contentType != "application/json; charset=utf-8" {
				t.Errorf("Unexpected Content-Type: %s", contentType)
			}

			if agend := req.Header.Get("User-Agent"); agend != "Policy-Reporter" {
				t.Errorf("Unexpected Host: %s", agend)
			}

			if url := req.URL.String(); url != "http://hook.teams:80" {
				t.Errorf("Unexpected Host: %s", url)
			}

			payload := make(map[string]interface{})

			err := json.NewDecoder(req.Body).Decode(&payload)
			if err != nil {
				t.Fatal(err)
			}

			if payload["themeColor"] != "68c2ff" {
				t.Errorf("Unexpected ThemeColor %s", payload["themeColor"])
			}
		}

		client := teams.NewClient(teams.Options{
			ClientOptions: target.ClientOptions{
				Name: "Teams",
			},
			Webhook:      "http://hook.teams:80",
			CustomFields: map[string]string{"Cluster": "Name"},
			HTTPClient:   testClient{callback, 200},
		})
		client.Send(minimalDebugResult)
	})
	t.Run("Name", func(t *testing.T) {
		client := teams.NewClient(teams.Options{
			ClientOptions: target.ClientOptions{
				Name: "Teams",
			},
			Webhook:      "http://hook.teams:80",
			CustomFields: map[string]string{"Cluster": "Name"},
			HTTPClient:   testClient{},
		})

		if client.Name() != "Teams" {
			t.Errorf("Unexpected Name %s", client.Name())
		}
	})
}
