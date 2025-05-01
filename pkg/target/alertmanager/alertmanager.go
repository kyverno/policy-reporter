package alertmanager

import (
	"encoding/json"
	"os"
	"time"

	"go.uber.org/zap"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
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

func (a *client) Send(result v1alpha2.PolicyReportResult) {
	zap.L().Info("Sending policy violation to AlertManager",
		zap.String("policy", result.Policy),
		zap.String("rule", result.Rule),
		zap.String("severity", string(result.Severity)),
		zap.String("status", string(result.Result)),
		zap.String("source", result.Source),
		zap.String("message", result.Message))

	alert := a.createAlert(result)
	a.sendAlerts([]Alert{alert})
}

func (a *client) BatchSend(report v1alpha2.ReportInterface, results []v1alpha2.PolicyReportResult) {
	zap.L().Info("Batch sending policy violations to AlertManager",
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

// SendTestAlert sends a test alert to the AlertManager to verify connectivity
// File can be a path to a JSON file containing alerts or empty for a default test alert
func (a *client) SendTestAlert(file string) error {
	zap.L().Info("Sending test alert to AlertManager",
		zap.String("host", a.host),
		zap.String("file", file))

	var alerts []Alert

	// Use provided file or create a default test alert
	if file != "" {
		// Try to read the test file
		data, err := os.ReadFile(file)
		if err != nil {
			zap.L().Error("Failed to read test alert file",
				zap.String("file", file),
				zap.Error(err))
			return err
		}

		if err := json.Unmarshal(data, &alerts); err != nil {
			zap.L().Error("Failed to parse test alert file",
				zap.String("file", file),
				zap.Error(err))
			return err
		}

		zap.L().Info("Loaded test alert from file",
			zap.String("file", file),
			zap.Int("alertCount", len(alerts)))
	} else {
		// Create a default test alert
		now := time.Now()
		alerts = []Alert{
			{
				Labels: map[string]string{
					"alertname": "PolicyReporterTest",
					"severity":  "info",
					"source":    "policy-reporter",
				},
				Annotations: map[string]string{
					"summary":     "Policy Reporter Test Alert",
					"description": "This is a test alert sent from Policy Reporter to verify AlertManager connectivity",
				},
				StartsAt: now,
				EndsAt:   now.Add(1 * time.Hour),
			},
		}
		zap.L().Info("Created default test alert")
	}

	// Send the test alert(s)
	a.sendAlerts(alerts)
	return nil
}

func (a *client) createAlert(result v1alpha2.PolicyReportResult) Alert {
	labels := map[string]string{
		"severity": string(result.Severity),
		"status":   string(result.Result),
		"source":   result.Source,
		"policy":   result.Policy,
		"rule":     result.Rule,
	}

	annotations := map[string]string{
		"message": result.Message,
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
	zap.L().Info("Sending alerts to AlertManager",
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
	return &client{
		target.NewBaseClient(options.ClientOptions),
		options.Host,
		options.Headers,
		options.CustomFields,
		options.HTTPClient,
	}
}
