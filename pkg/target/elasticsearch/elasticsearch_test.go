package elasticsearch_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/fjogeleit/policy-reporter/pkg/report"
	"github.com/fjogeleit/policy-reporter/pkg/target/elasticsearch"
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
		}

		loki := elasticsearch.NewClient("http://localhost:9200", "policy-reporter", "annually", "", false, testClient{callback, 200})
		loki.Send(completeResult)
	})
	t.Run("Send with Monthly Result", func(t *testing.T) {
		callback := func(req *http.Request) {
			if url := req.URL.String(); url != "http://localhost:9200/policy-reporter-"+time.Now().Format("2006.01")+"/event" {
				t.Errorf("Unexpected Host: %s", url)
			}
		}

		loki := elasticsearch.NewClient("http://localhost:9200", "policy-reporter", "monthly", "", false, testClient{callback, 200})
		loki.Send(completeResult)
	})
	t.Run("Send with Monthly Result", func(t *testing.T) {
		callback := func(req *http.Request) {
			if url := req.URL.String(); url != "http://localhost:9200/policy-reporter-"+time.Now().Format("2006.01.02")+"/event" {
				t.Errorf("Unexpected Host: %s", url)
			}
		}

		loki := elasticsearch.NewClient("http://localhost:9200", "policy-reporter", "daily", "", false, testClient{callback, 200})
		loki.Send(completeResult)
	})
	t.Run("Send with None Result", func(t *testing.T) {
		callback := func(req *http.Request) {
			if url := req.URL.String(); url != "http://localhost:9200/policy-reporter/event" {
				t.Errorf("Unexpected Host: %s", url)
			}
		}

		loki := elasticsearch.NewClient("http://localhost:9200", "policy-reporter", "none", "", false, testClient{callback, 200})
		loki.Send(completeResult)
	})
}
