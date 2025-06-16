package slack_test

import (
	"net/http"
	"testing"

	"github.com/kyverno/policy-reporter/pkg/fixtures"
	"github.com/kyverno/policy-reporter/pkg/openreports"
	"github.com/kyverno/policy-reporter/pkg/target"
	"github.com/kyverno/policy-reporter/pkg/target/slack"
)

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

func Test_SlackTarget(t *testing.T) {
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

		client := slack.NewClient(slack.Options{
			ClientOptions: target.ClientOptions{
				Name: "Slack",
			},
			Webhook:      "http://hook.slack:80",
			CustomFields: map[string]string{"Cluster": "Name"},
			HTTPClient:   testClient{callback, 200},
		})
		client.Send(&openreports.ORResultAdapter{ReportResult: &fixtures.CompleteTargetSendResult})
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

		client := slack.NewClient(slack.Options{
			ClientOptions: target.ClientOptions{
				Name: "Slack",
			},
			Webhook:    "http://hook.slack:80",
			HTTPClient: testClient{callback, 200},
		})
		client.Send(&openreports.ORResultAdapter{ReportResult: &fixtures.MinimalTargetSendResult})
	})

	t.Run("Send enforce Result", func(t *testing.T) {
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

		client := slack.NewClient(slack.Options{
			ClientOptions: target.ClientOptions{
				Name: "Slack",
			},
			Webhook:    "http://hook.slack:80",
			HTTPClient: testClient{callback, 200},
		})
		client.Send(&openreports.ORResultAdapter{ReportResult: &fixtures.EnforceTargetSendResult})
	})

	t.Run("Send incomplete Result", func(t *testing.T) {
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

		client := slack.NewClient(slack.Options{
			ClientOptions: target.ClientOptions{
				Name: "Slack",
			},
			Webhook:      "http://hook.slack:80",
			CustomFields: map[string]string{"Cluster": "Name"},
			HTTPClient:   testClient{callback, 200},
		})
		client.Send(&openreports.ORResultAdapter{ReportResult: &fixtures.MissingUIDSendResult})
	})

	t.Run("Send incomplete Result2", func(t *testing.T) {
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

		client := slack.NewClient(slack.Options{
			ClientOptions: target.ClientOptions{
				Name: "Slack",
			},
			Webhook:      "http://hook.slack:80",
			CustomFields: map[string]string{"Cluster": "Name"},
			HTTPClient:   testClient{callback, 200},
		})
		client.Send(&openreports.ORResultAdapter{ReportResult: &fixtures.MissingAPIVersionSendResult})
	})

	t.Run("Name", func(t *testing.T) {
		client := slack.NewClient(slack.Options{
			ClientOptions: target.ClientOptions{
				Name: "Slack",
			},
			Webhook:      "http://hook.slack:80",
			CustomFields: map[string]string{"Cluster": "Name"},
			HTTPClient:   testClient{},
		})

		if client.Name() != "Slack" {
			t.Errorf("Unexpected Name %s", client.Name())
		}
	})
}
