package pagerduty

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/PagerDuty/go-pagerduty"
	"go.uber.org/zap"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/target"
	"github.com/kyverno/policy-reporter/pkg/target/formatting"
)

// Options to configure the PagerDuty target
type Options struct {
	target.ClientOptions
	APIToken     string
	ServiceID    string
	CustomFields map[string]string
}

type client struct {
	target.BaseClient
	client       *pagerduty.Client
	serviceID    string
	customFields map[string]string
	// Track active incidents by policy+resource
	incidents sync.Map
}

// Create a unique key for tracking incidents
func incidentKey(result v1alpha2.PolicyReportResult) string {
	key := result.Policy
	if result.HasResource() {
		res := result.GetResource()
		key = fmt.Sprintf("%s/%s/%s/%s", 
			result.Policy,
			res.Kind,
			res.Namespace,
			res.Name,
		)
	}
	return key
}

func (p *client) Send(result v1alpha2.PolicyReportResult) {
	key := incidentKey(result)

	if result.Result == v1alpha2.StatusPass {
		// Check if we have an active incident to resolve
		if incidentID, ok := p.incidents.Load(key); ok {
			p.resolveIncident(incidentID.(string))
			p.incidents.Delete(key)
		}
		return
	}

	if result.Result != v1alpha2.StatusFail {
		// Only create incidents for failed policies
		return
	}

	// Check if we already have an incident for this policy/resource
	if _, exists := p.incidents.Load(key); exists {
		// Incident already exists, no need to create another
		return
	}

	details := map[string]interface{}{
		"policy":   result.Policy,
		"rule":     result.Rule,
		"message":  result.Message,
		"severity": result.Severity,
	}

	if result.HasResource() {
		res := result.GetResource()
		details["resource"] = formatting.ResourceString(res)
	}

	for k, v := range p.customFields {
		details[k] = v
	}

	for k, v := range result.Properties {
		details[k] = v
	}

	incident := pagerduty.CreateIncidentOptions{
		Type:    "incident",
		Title:   fmt.Sprintf("Policy Violation: %s", result.Policy),
		Service: &pagerduty.APIReference{ID: p.serviceID, Type: "service_reference"},
		Body: &pagerduty.APIDetails{
			Type:    "incident_body",
			Details: details,
		},
		Urgency: mapSeverityToUrgency(result.Severity),
	}

	resp, err := p.client.CreateIncident("policy-reporter", &incident)
	if err != nil {
		zap.L().Error("failed to create PagerDuty incident", 
			zap.String("policy", result.Policy),
			zap.Error(err),
		)
		return
	}

	// Store the incident ID for later resolution
	p.incidents.Store(key, resp.Id)

	zap.L().Info("PagerDuty incident created", 
		zap.String("policy", result.Policy),
		zap.String("severity", string(result.Severity)),
		zap.String("incidentId", resp.Id),
	)
}

func (p *client) resolveIncident(incidentID string) {
	incident := pagerduty.ManageIncidentsOptions{
		ID: incidentID,
		Incidents: []pagerduty.ManageIncident{
			{
				Status: "resolved",
				Resolution: "Policy violation has been resolved",
			},
		},
	}

	if err := p.client.ManageIncidents("policy-reporter", &incident); err != nil {
		zap.L().Error("failed to resolve PagerDuty incident",
			zap.String("incidentId", incidentID),
			zap.Error(err),
		)
		return
	}

	zap.L().Info("PagerDuty incident resolved",
		zap.String("incidentId", incidentID),
	)
}

func (p *client) Type() target.ClientType {
	return target.SingleSend
}

func mapSeverityToUrgency(severity v1alpha2.PolicySeverity) string {
	switch severity {
	case v1alpha2.SeverityCritical, v1alpha2.SeverityHigh:
		return "high"
	default:
		return "low"
	}
}

// SetClient allows replacing the PagerDuty client for testing
func (p *client) SetClient(c interface{}) {
	if pdClient, ok := c.(interface {
		CreateIncident(string, *pagerduty.CreateIncidentOptions) (*pagerduty.Incident, error)
		ManageIncidents(string, *pagerduty.ManageIncidentsOptions) error
	}); ok {
		p.client = pdClient
	}
}

// NewClient creates a new PagerDuty client
func NewClient(options Options) target.Client {
	return &client{
		target.NewBaseClient(options.ClientOptions),
		pagerduty.NewClient(options.APIToken),
		options.ServiceID,
		options.CustomFields,
		sync.Map{},
	}
} 