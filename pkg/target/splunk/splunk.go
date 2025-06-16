package splunk

import (
	"encoding/json"

	"go.uber.org/zap"

	"github.com/kyverno/policy-reporter/pkg/openreports"
	"github.com/kyverno/policy-reporter/pkg/target"
	"github.com/kyverno/policy-reporter/pkg/target/http"
)

const policyReporterSource = "Policy-Reporter"

type splunkRequest struct {
	Event      http.Result `json:"event"`
	SourceType string      `json:"sourcetype"`
}

type Options struct {
	target.ClientOptions
	Host         string
	CustomFields map[string]string
	HTTPClient   http.Client
	Headers      map[string]string
	Token        string
}

type client struct {
	target.BaseClient
	host         string
	customFields map[string]string
	headers      map[string]string
	client       http.Client
	token        string
}

func (c *client) Send(result *openreports.ORResultAdapter) {
	c.sendAndLogResult(splunkRequest{
		Event:      http.NewJSONResult(result),
		SourceType: policyReporterSource,
	})
}

func (c *client) BatchSend(rep openreports.ReportInterface, results []*openreports.ORResultAdapter) {
	srs := ""
	for _, res := range results {
		sr := splunkRequest{
			Event:      http.NewJSONResult(res),
			SourceType: policyReporterSource,
		}
		srString, err := json.Marshal(sr)
		if err != nil {
			zap.L().Error(c.Name()+"Error marhsalling the JSON to a splunk request:", zap.Error(err))
			return
		}
		srs = srs + string(srString)
	}

	c.sendAndLogResult(srs)
}

func (c *client) sendAndLogResult(payload interface{}) {
	req, err := http.CreateJSONRequest("POST", c.host, payload)
	if err != nil {
		zap.L().Error(c.Name()+": PUSH FAILED", zap.Error(err))
		return
	}

	for k, v := range c.headers {
		req.Header.Set(k, v)
	}

	resp, err := c.client.Do(req)

	http.ProcessHTTPResponse(c.Name(), resp, err)
	zap.L().Info(c.Name() + ": PUSH OK")
}

func (c *client) Type() string {
	return target.BatchSend
}

func NewClient(options Options) target.Client {
	return &client{
		target.NewBaseClient(options.ClientOptions),
		options.Host,
		options.CustomFields,
		options.Headers,
		options.HTTPClient,
		options.Token,
	}
}
