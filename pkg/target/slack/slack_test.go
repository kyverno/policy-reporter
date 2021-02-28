package slack_test

import (
	"net/http"
	"testing"

	"github.com/fjogeleit/policy-reporter/pkg/report"
	"github.com/fjogeleit/policy-reporter/pkg/target/slack"
)

var completeResult = report.Result{
	Message:  "validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/",
	Policy:   "require-requests-and-limits-required",
	Rule:     "autogen-check-for-requests-and-limits",
	Priority: report.ErrorPriority,
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

var minimalResult = report.Result{
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
			if contentType := req.Header.Get("Content-Type"); contentType != "application/json; charset=utf-8" {
				t.Errorf("Unexpected Content-Type: %s", contentType)
			}

			if agend := req.Header.Get("User-Agent"); agend != "Policy-Reporter" {
				t.Errorf("Unexpected Host: %s", agend)
			}

			if url := req.URL.String(); url != "http://hook.slack:80" {
				t.Errorf("Unexpected Host: %s", url)
			}
		}

		slack := slack.NewClient("http://hook.slack:80", "", false, testClient{callback, 200})
		slack.Send(minimalResult)
	})

	t.Run("Send Minimal Result", func(t *testing.T) {
		callback := func(req *http.Request) {
			if contentType := req.Header.Get("Content-Type"); contentType != "application/json; charset=utf-8" {
				t.Errorf("Unexpected Content-Type: %s", contentType)
			}

			if agend := req.Header.Get("User-Agent"); agend != "Policy-Reporter" {
				t.Errorf("Unexpected Host: %s", agend)
			}

			if url := req.URL.String(); url != "http://hook.slack:80" {
				t.Errorf("Unexpected Host: %s", url)
			}
		}

		slack := slack.NewClient("http://hook.slack:80", "", false, testClient{callback, 200})
		slack.Send(minimalResult)
	})
}
