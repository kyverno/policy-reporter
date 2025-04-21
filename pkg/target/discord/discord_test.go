package discord_test

import (
	"net/http"
	"testing"

	"github.com/kyverno/policy-reporter/pkg/fixtures"
	"github.com/kyverno/policy-reporter/pkg/payload"
	"github.com/kyverno/policy-reporter/pkg/target"
	"github.com/kyverno/policy-reporter/pkg/target/discord"
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

func Test_LokiTarget(t *testing.T) {
	t.Run("Send Complete Result", func(t *testing.T) {
		callback := func(req *http.Request) {
			if contentType := req.Header.Get("Content-Type"); contentType != "application/json; charset=utf-8" {
				t.Errorf("Unexpected Content-Type: %s", contentType)
			}

			if agend := req.Header.Get("User-Agent"); agend != "Policy-Reporter" {
				t.Errorf("Unexpected Host: %s", agend)
			}

			if url := req.URL.String(); url != "http://hook.discord:80" {
				t.Errorf("Unexpected Host: %s", url)
			}
		}

		client := discord.NewClient(discord.Options{
			ClientOptions: target.ClientOptions{
				Name: "Discord",
			},
			Webhook:    "http://hook.discord:80",
			HTTPClient: testClient{callback, 200},
		})
		client.Send(&payload.PolicyReportResultPayload{Result: fixtures.CompleteTargetSendResult})
	})

	t.Run("Send Minimal Result", func(t *testing.T) {
		callback := func(req *http.Request) {
			if contentType := req.Header.Get("Content-Type"); contentType != "application/json; charset=utf-8" {
				t.Errorf("Unexpected Content-Type: %s", contentType)
			}

			if agend := req.Header.Get("User-Agent"); agend != "Policy-Reporter" {
				t.Errorf("Unexpected Host: %s", agend)
			}

			if url := req.URL.String(); url != "http://hook.discord:80" {
				t.Errorf("Unexpected Host: %s", url)
			}
		}

		client := discord.NewClient(discord.Options{
			ClientOptions: target.ClientOptions{
				Name: "Discord",
			},
			Webhook:    "http://hook.discord:80",
			HTTPClient: testClient{callback, 200},
		})
		client.Send(&payload.PolicyReportResultPayload{Result: fixtures.MinimalTargetSendResult})
	})
	t.Run("Name", func(t *testing.T) {
		client := discord.NewClient(discord.Options{
			ClientOptions: target.ClientOptions{
				Name: "Discord",
			},
			Webhook:    "http://hook.discord:80",
			HTTPClient: testClient{},
		})

		if client.Name() != "Discord" {
			t.Errorf("Unexpected Name %s", client.Name())
		}
	})
}
