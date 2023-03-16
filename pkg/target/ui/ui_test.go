package ui_test

import (
	"net/http"
	"testing"

	"github.com/kyverno/policy-reporter/pkg/fixtures"
	"github.com/kyverno/policy-reporter/pkg/target"
	"github.com/kyverno/policy-reporter/pkg/target/ui"
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
				Name:   "UI",
				Logger: fixtures.Logger,
			},
			Host:       "http://localhost:8080",
			HTTPClient: testClient{callback, 200},
		})
		client.Send(fixtures.CompleteTargetSendResult)
	})
	t.Run("Name", func(t *testing.T) {
		client := ui.NewClient(ui.Options{
			ClientOptions: target.ClientOptions{
				Name:   "UI",
				Logger: fixtures.Logger,
			},
			Host:       "http://localhost:8080",
			HTTPClient: testClient{},
		})

		if client.Name() != "UI" {
			t.Errorf("Unexpected Name %s", client.Name())
		}
	})
}
