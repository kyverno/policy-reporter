package alertmanager

import (
	"time"

	"go.uber.org/zap"
	reportsv1alpha1 "openreports.io/apis/openreports.io/v1alpha1"

	"github.com/kyverno/policy-reporter/pkg/openreports"
	"github.com/kyverno/policy-reporter/pkg/target"
	targethttp "github.com/kyverno/policy-reporter/pkg/target/http"
)

// Alert represents an AlertManager alert
type Alert struct {
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	StartsAt    time.Time         `json:"startsAt"`
	EndsAt      time.Time         `json:"endsAt"`
}

// Options to configure the AlertManager target
type Options struct {
	target.ClientOptions
	Host         string
	Headers      map[string]string
	CustomFields map[string]string
	HTTPClient   targethttp.Client
}

type client struct {
	target.BaseClient
	host         string
	headers      map[string]string
	customFields map[string]string
	client       targethttp.Client
}

// Ensure the client type implements the Client interface
var _ Client = (*client)(nil)

func (a *client) Send(result reportsv1alpha1.ReportResult) {
	zap.L().Debug("Sending policy violation to AlertManager",
		zap.String("policy", result.Policy),
		zap.String("rule", result.Rule),
		zap.String("severity", string(result.Severity)),
		zap.String("status", string(result.Result)),
		zap.String("source", result.Source),
		zap.String("message", result.Description))

	alert := a.createAlert(result)
	a.sendAlerts([]Alert{alert})
}

func (a *client) BatchSend(report openreports.ReportInterface, results []reportsv1alpha1.ReportResult) {
	zap.L().Debug("Batch sending policy violations to AlertManager",
		zap.Int("count", len(results)),
		zap.String("reportName", report.GetName()),
		zap.String("reportNamespace", report.GetNamespace()))

	alerts := make([]Alert, 0, len(results))
	for _, result := range results {
		zap.L().Debug("Processing policy violation for AlertManager",
			zap.String("policy", result.Policy),
			zap.String("rule", result.Rule),
			zap.String("severity", string(result.Severity)),
			zap.String("status", string(result.Result)),
			zap.String("source", result.Source))

		alerts = append(alerts, a.createAlert(result))
	}
	a.sendAlerts(alerts)
}

func (a *client) createAlert(result reportsv1alpha1.ReportResult) Alert {
	labels := map[string]string{
		"alertname": "PolicyReporterViolation",
		"severity":  string(result.Severity),
		"status":    string(result.Result),
		"source":    result.Source,
		"policy":    result.Policy,
		"rule":      result.Rule,
	}

	annotations := map[string]string{
		"message": result.Description,
	}

	// Add resource information if available
	if result.HasResource() {
		resource := result.GetResource()
		resourceString := result.ResourceString()

		// Add resource identifiers to annotations
		annotations["resource"] = resourceString

		if resource.Kind != "" {
			annotations["resource_kind"] = resource.Kind
		}

		if resource.Name != "" {
			annotations["resource_name"] = resource.Name
		}

		if resource.Namespace != "" {
			annotations["resource_namespace"] = resource.Namespace
		}

		if resource.APIVersion != "" {
			annotations["resource_apiversion"] = resource.APIVersion
		}
	}

	if result.Category != "" {
		annotations["category"] = result.Category
	}

	for k, v := range result.Properties {
		annotations[k] = v
	}

	for k, v := range a.customFields {
		annotations[k] = v
	}

	startsAt := time.Now()
	return Alert{
		Labels:      labels,
		Annotations: annotations,
		StartsAt:    startsAt,
		EndsAt:      startsAt.Add(24 * time.Hour),
	}
}

func (a *client) sendAlerts(alerts []Alert) {
	zap.L().Debug("Sending alerts to AlertManager",
		zap.Int("alertCount", len(alerts)),
		zap.String("endpoint", a.host+"/api/v2/alerts"))

	req, err := targethttp.CreateJSONRequest("POST", a.host+"/api/v2/alerts", alerts)
	if err != nil {
		zap.L().Error("Failed to create request", zap.Error(err))
		return
	}

	for key, value := range a.headers {
		req.Header.Set(key, value)
	}

	resp, err := a.client.Do(req)
	targethttp.ProcessHTTPResponse(a.Name(), resp, err)
}

func (a *client) Type() target.ClientType {
	return target.BatchSend
}

// NewClient creates a new AlertManager client to send policy violations
func NewClient(options Options) Client {
	zap.L().Debug("Creating AlertManager client",
		zap.String("host", options.Host),
		zap.Int("headers", len(options.Headers)))

	return &client{
		target.NewBaseClient(options.ClientOptions),
		options.Host,
		options.Headers,
		options.CustomFields,
		options.HTTPClient,
	}
}
