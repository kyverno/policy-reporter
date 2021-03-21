package ui_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/fjogeleit/policy-reporter/pkg/report"
	"github.com/fjogeleit/policy-reporter/pkg/target/ui"
)

var completeResult = report.Result{
	Message:  "validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/",
	Policy:   "require-requests-and-limits-required",
	Rule:     "autogen-check-for-requests-and-limits",
	Priority: report.WarningPriority,
	Status:   report.Fail,
	Severity: report.Heigh,
	Category: "resources",
	Scored:   true,
	Resources: []report.Resource{
		{
			APIVersion: "v1",
			Kind:       "Deployment",
			Name:       "nginx",
			Namespace:  "default",
			UID:        "536ab69f-1b3c-4bd9-9ba4-274a56188409",
		},
	},
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

			if url := req.URL.String(); url != "http://localhost:9200/policy-reporter-"+time.Now().Format("2006")+"/event" {
				t.Errorf("Unexpected Host: %s", url)
			}
		}

		client := ui.NewClient("http://localhost:8080", "", false, testClient{callback, 200})
		client.Send(completeResult)
	})
	t.Run("Send with ingored Priority", func(t *testing.T) {
		callback := func(req *http.Request) {
			t.Errorf("Unexpected Call")
		}

		client := ui.NewClient("http://localhost:8080", "", false, testClient{callback, 200})
		client.Send(completeResult)
	})
	t.Run("SkipExistingOnStartup", func(t *testing.T) {
		callback := func(req *http.Request) {
			t.Errorf("Unexpected Call")
		}

		client := ui.NewClient("http://localhost:8080", "", false, testClient{callback, 200})

		if !client.SkipExistingOnStartup() {
			t.Error("Should return configured SkipExistingOnStartup")
		}
	})
	t.Run("Name", func(t *testing.T) {
		client := ui.NewClient("http://localhost:8080", "", false, testClient{})

		if client.Name() != "UI" {
			t.Errorf("Unexpected Name %s", client.Name())
		}
	})
	t.Run("MinimumPriority", func(t *testing.T) {
		client := ui.NewClient("http://localhost:8080", "debug", false, testClient{})

		if client.MinimumPriority() != "debug" {
			t.Errorf("Unexpected MinimumPriority %s", client.MinimumPriority())
		}
	})
}
