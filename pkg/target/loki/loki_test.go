package loki_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/kyverno/policy-reporter/pkg/target/loki"
)

var completeResult = &report.Result{
	Message:   "validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/",
	Policy:    "require-requests-and-limits-required",
	Rule:      "autogen-check-for-requests-and-limits",
	Timestamp: time.Date(2021, time.February, 23, 15, 10, 0, 0, time.UTC),
	Priority:  report.WarningPriority,
	Status:    report.Fail,
	Severity:  report.High,
	Category:  "resources",
	Scored:    true,
	Source:    "Kyverno",
	Resource: &report.Resource{
		APIVersion: "v1",
		Kind:       "Deployment",
		Name:       "nginx",
		Namespace:  "default",
		UID:        "536ab69f-1b3c-4bd9-9ba4-274a56188409",
	},
	Properties: map[string]string{"version": "1.2.0"},
}

var minimalResult = &report.Result{
	Message:  "validation error: label required. Rule app-label-required failed at path /spec/template/spec/containers/0/resources/requests/",
	Policy:   "app-label-requirement",
	Priority: report.WarningPriority,
	Status:   report.Fail,
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

func Test_LokiTarget(t *testing.T) {
	t.Run("Send Complete Result", func(t *testing.T) {
		callback := func(req *http.Request) {
			if contentType := req.Header.Get("Content-Type"); contentType != "application/json" {
				t.Errorf("Unexpected Content-Type: %s", contentType)
			}

			if agend := req.Header.Get("User-Agent"); agend != "Policy-Reporter" {
				t.Errorf("Unexpected Host: %s", agend)
			}

			if url := req.URL.String(); url != "http://localhost:3100/api/prom/push" {
				t.Errorf("Unexpected Host: %s", url)
			}

			expectedLine := fmt.Sprintf("[%s] %s", strings.ToUpper(completeResult.Priority.String()), completeResult.Message)
			labels, line := convertAndValidateBody(req, t)
			if line != expectedLine {
				t.Errorf("Unexpected LineContent: %s", line)
			}
			if !strings.Contains(labels, "policy=\""+completeResult.Policy+"\"") {
				t.Error("Missing Content for Label 'policy'")
			}
			if !strings.Contains(labels, "status=\""+completeResult.Status+"\"") {
				t.Error("Missing Content for Label 'status'")
			}
			if !strings.Contains(labels, "priority=\""+completeResult.Priority.String()+"\"") {
				t.Error("Missing Content for Label 'priority'")
			}
			if !strings.Contains(labels, "source=\"policy-reporter\"") {
				t.Error("Missing Content for Label 'policy-reporter'")
			}
			if !strings.Contains(labels, "rule=\""+completeResult.Rule+"\"") {
				t.Error("Missing Content for Label 'rule'")
			}
			if !strings.Contains(labels, "category=\""+completeResult.Category+"\"") {
				t.Error("Missing Content for Label 'category'")
			}
			if !strings.Contains(labels, "severity=\""+completeResult.Severity+"\"") {
				t.Error("Missing Content for Label 'severity'")
			}

			res := completeResult.Resource
			if !strings.Contains(labels, "kind=\""+res.Kind+"\"") {
				t.Error("Missing Content for Label 'kind'")
			}
			if !strings.Contains(labels, "name=\""+res.Name+"\"") {
				t.Error("Missing Content for Label 'name'")
			}
			if !strings.Contains(labels, "uid=\""+res.UID+"\"") {
				t.Error("Missing Content for Label 'uid'")
			}
			if !strings.Contains(labels, "namespace=\""+res.Namespace+"\"") {
				t.Error("Missing Content for Label 'namespace'")
			}
			if !strings.Contains(labels, "version=\""+completeResult.Properties["version"]+"\"") {
				t.Error("Missing Content for Label 'version'")
			}
		}

		loki := loki.NewClient("http://localhost:3100", "", []string{}, false, testClient{callback, 200})
		loki.Send(completeResult)
	})

	t.Run("Send Minimal Result", func(t *testing.T) {
		callback := func(req *http.Request) {
			if contentType := req.Header.Get("Content-Type"); contentType != "application/json" {
				t.Errorf("Unexpected Content-Type: %s", contentType)
			}

			if agend := req.Header.Get("User-Agent"); agend != "Policy-Reporter" {
				t.Errorf("Unexpected Host: %s", agend)
			}

			if url := req.URL.String(); url != "http://localhost:3100/api/prom/push" {
				t.Errorf("Unexpected Host: %s", url)
			}

			expectedLine := fmt.Sprintf("[%s] %s", strings.ToUpper(minimalResult.Priority.String()), minimalResult.Message)
			labels, line := convertAndValidateBody(req, t)
			if line != expectedLine {
				t.Errorf("Unexpected LineContent: %s", line)
			}
			if !strings.Contains(labels, "policy=\""+minimalResult.Policy+"\"") {
				t.Error("Missing Content for Label 'policy'")
			}
			if !strings.Contains(labels, "status=\""+minimalResult.Status+"\"") {
				t.Error("Missing Content for Label 'status'")
			}
			if !strings.Contains(labels, "priority=\""+minimalResult.Priority.String()+"\"") {
				t.Error("Missing Content for Label 'priority'")
			}
			if !strings.Contains(labels, "source=\"policy-reporter\"") {
				t.Error("Missing Content for Label 'policy-reporter'")
			}
			if strings.Contains(labels, "rule") {
				t.Error("Unexpected Label 'rule'")
			}
			if strings.Contains(labels, "category") {
				t.Error("Unexpected Label 'category'")
			}
			if strings.Contains(labels, "severity") {
				t.Error("Unexpected 'severity'")
			}
			if strings.Contains(labels, "kind") {
				t.Error("Unexpected Label 'kind'")
			}
			if strings.Contains(labels, "name") {
				t.Error("Unexpected 'name'")
			}
			if strings.Contains(labels, "uid") {
				t.Error("Unexpected 'uid'")
			}
			if strings.Contains(labels, "namespace") {
				t.Error("Unexpected 'namespace'")
			}
		}

		loki := loki.NewClient("http://localhost:3100", "", []string{}, false, testClient{callback, 200})
		loki.Send(minimalResult)
	})
	t.Run("Name", func(t *testing.T) {
		client := loki.NewClient("http://localhost:9200", "", []string{}, true, testClient{})

		if client.Name() != "Loki" {
			t.Errorf("Unexpected Name %s", client.Name())
		}
	})
}

func convertAndValidateBody(req *http.Request, t *testing.T) (string, string) {
	payload := make(map[string]interface{})

	err := json.NewDecoder(req.Body).Decode(&payload)
	if err != nil {
		t.Fatal(err)
	}

	streamsContent, ok := payload["streams"]
	if !ok {
		t.Errorf("Expected payload key 'streams' is missing")
	}

	streams := streamsContent.([]interface{})
	if len(streams) != 1 {
		t.Errorf("Expected one streams entry")
	}

	firstStream := streams[0].(map[string]interface{})
	entriesContent, ok := firstStream["entries"]
	if !ok {
		t.Errorf("Expected stream key 'entries' is missing")
	}
	labels, ok := firstStream["labels"]
	if !ok {
		t.Errorf("Expected stream key 'labels' is missing")
	}

	entryContent := entriesContent.([]interface{})[0]
	entry := entryContent.(map[string]interface{})
	_, ok = entry["ts"]
	if !ok {
		t.Errorf("Expected entry key 'ts' is missing")
	}
	line, ok := entry["line"]
	if !ok {
		t.Errorf("Expected entry key 'line' is missing")
	}

	return labels.(string), line.(string)
}
