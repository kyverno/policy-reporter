package teams_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/fjogeleit/policy-reporter/pkg/report"
	"github.com/fjogeleit/policy-reporter/pkg/target/teams"
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

var minimalErrorResult = report.Result{
	Message:  "validation error: label required. Rule app-label-required failed at path /spec/template/spec/containers/0/resources/requests/",
	Policy:   "app-label-requirement",
	Priority: report.ErrorPriority,
	Status:   report.Fail,
	Scored:   true,
}

var minimalResult = report.Result{
	Message:  "validation error: label required. Rule app-label-required failed at path /spec/template/spec/containers/0/resources/requests/",
	Policy:   "app-label-requirement",
	Priority: report.CriticalPriority,
	Status:   report.Fail,
	Scored:   true,
}

var minimalInfoResult = report.Result{
	Message:  "validation error: label required. Rule app-label-required failed at path /spec/template/spec/containers/0/resources/requests/",
	Policy:   "app-label-requirement",
	Priority: report.InfoPriority,
	Status:   report.Fail,
	Scored:   true,
}

var minimalDebugResult = report.Result{
	Message:  "validation error: label required. Rule app-label-required failed at path /spec/template/spec/containers/0/resources/requests/",
	Policy:   "app-label-requirement",
	Priority: report.DebugPriority,
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

		client := teams.NewClient("http://hook.teams:80", "", false, testClient{callback, 200})
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

		client := teams.NewClient("http://hook.teams:80", "", false, testClient{callback, 200})
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

		client := teams.NewClient("http://hook.teams:80", "", false, testClient{callback, 200})
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

		client := teams.NewClient("http://hook.teams:80", "", false, testClient{callback, 200})
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

		client := teams.NewClient("http://hook.teams:80", "", false, testClient{callback, 200})
		client.Send(minimalDebugResult)
	})
	t.Run("Send with ingored Priority", func(t *testing.T) {
		callback := func(req *http.Request) {
			t.Errorf("Unexpected Call")
		}

		client := teams.NewClient("http://localhost:9200", "error", false, testClient{callback, 200})
		client.Send(completeResult)
	})
	t.Run("SkipExistingOnStartup", func(t *testing.T) {
		callback := func(req *http.Request) {
			t.Errorf("Unexpected Call")
		}

		client := teams.NewClient("http://localhost:9200", "", true, testClient{callback, 200})

		if !client.SkipExistingOnStartup() {
			t.Error("Should return configured SkipExistingOnStartup")
		}
	})
	t.Run("Name", func(t *testing.T) {
		client := teams.NewClient("http://localhost:9200", "", true, testClient{})

		if client.Name() != "Teams" {
			t.Errorf("Unexpected Name %s", client.Name())
		}
	})
	t.Run("MinimumPriority", func(t *testing.T) {
		client := teams.NewClient("http://localhost:9200", "debug", true, testClient{})

		if client.MinimumPriority() != "debug" {
			t.Errorf("Unexpected MinimumPriority %s", client.MinimumPriority())
		}
	})
}
