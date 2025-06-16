package elasticsearch_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/kyverno/policy-reporter/pkg/fixtures"
	"github.com/kyverno/policy-reporter/pkg/openreports"
	"github.com/kyverno/policy-reporter/pkg/target"
	"github.com/kyverno/policy-reporter/pkg/target/elasticsearch"
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

func Test_ElasticsearchTarget(t *testing.T) {
	t.Run("Send with Annually Result", func(t *testing.T) {
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

			if req.Header.Get("Authorization") == "" {
				t.Error("Expected Authentication header for BasicAuth is set")
			}
		}

		client := elasticsearch.NewClient(elasticsearch.Options{
			ClientOptions: target.ClientOptions{
				Name: "Elasticsearch",
			},
			Host:         "http://localhost:9200",
			Username:     "username",
			Password:     "password",
			ApiKey:       "ApiKey",
			Index:        "policy-reporter",
			Rotation:     elasticsearch.Annually,
			HTTPClient:   testClient{callback, 200},
			CustomFields: map[string]string{"cluster": "name"},
		})
		client.Send(&openreports.ORResultAdapter{ReportResult: &fixtures.CompleteTargetSendResult})

		if len(fixtures.CompleteTargetSendResult.Properties) > 1 {
			t.Error("expected customFields are not added to the actuel result")
		}
	})
	t.Run("Send with Monthly Result", func(t *testing.T) {
		callback := func(req *http.Request) {
			if url := req.URL.String(); url != "http://localhost:9200/policy-reporter-"+time.Now().Format("2006.01")+"/event" {
				t.Errorf("Unexpected Host: %s", url)
			}

			if req.Header.Get("Authorization") != "" {
				t.Error("Expected Authentication header is not set")
			}
		}

		client := elasticsearch.NewClient(elasticsearch.Options{
			ClientOptions: target.ClientOptions{
				Name: "Elasticsearch",
			},
			Host:       "http://localhost:9200",
			Index:      "policy-reporter",
			Rotation:   elasticsearch.Monthly,
			HTTPClient: testClient{callback, 200},
		})
		client.Send(&openreports.ORResultAdapter{ReportResult: &fixtures.CompleteTargetSendResult})
	})
	t.Run("Send with Monthly Result", func(t *testing.T) {
		callback := func(req *http.Request) {
			if url := req.URL.String(); url != "http://localhost:9200/policy-reporter-"+time.Now().Format("2006.01.02")+"/event" {
				t.Errorf("Unexpected Host: %s", url)
			}
		}

		client := elasticsearch.NewClient(elasticsearch.Options{
			ClientOptions: target.ClientOptions{
				Name: "Elasticsearch",
			},
			Host:       "http://localhost:9200",
			Index:      "policy-reporter",
			Rotation:   elasticsearch.Daily,
			HTTPClient: testClient{callback, 200},
		})
		client.Send(&openreports.ORResultAdapter{ReportResult: &fixtures.CompleteTargetSendResult})
	})
	t.Run("Send with None Result", func(t *testing.T) {
		callback := func(req *http.Request) {
			if url := req.URL.String(); url != "http://localhost:9200/policy-reporter/event" {
				t.Errorf("Unexpected Host: %s", url)
			}
		}

		client := elasticsearch.NewClient(elasticsearch.Options{
			ClientOptions: target.ClientOptions{
				Name: "Elasticsearch",
			},
			Host:       "http://localhost:9200",
			Index:      "policy-reporter",
			Rotation:   elasticsearch.None,
			HTTPClient: testClient{callback, 200},
		})
		client.Send(&openreports.ORResultAdapter{ReportResult: &fixtures.CompleteTargetSendResult})
	})
	t.Run("Name", func(t *testing.T) {
		client := elasticsearch.NewClient(elasticsearch.Options{
			ClientOptions: target.ClientOptions{
				Name: "Elasticsearch",
			},
			Host:       "http://localhost:9200",
			Index:      "policy-reporter",
			Rotation:   elasticsearch.None,
			HTTPClient: testClient{},
		})

		if client.Name() != "Elasticsearch" {
			t.Errorf("Unexpected Name %s", client.Name())
		}
	})
}
