package pagerduty

import (
	"fmt"
	"strings"
	"time"
	"context"

	"github.com/PagerDuty/go-pagerduty"
	"go.uber.org/zap"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/target"
	"github.com/kyverno/policy-reporter/pkg/target/formatting"
)

// RetryConfig defines the retry behavior for failed PagerDuty API calls
type RetryConfig struct {
	MaxRetries  int           // Maximum number of retry attempts
	InitialWait time.Duration // Initial wait time between retries
	MaxWait     time.Duration // Maximum wait time between retries
}

// TimeoutConfig defines timeouts for PagerDuty API operations
type TimeoutConfig struct {
	Create  time.Duration // Timeout for incident creation
	Resolve time.Duration // Timeout for incident resolution
}

// Options configures the PagerDuty client behavior
type Options struct {
	target.ClientOptions
	APIToken     string            // PagerDuty API authentication token
	ServiceID    string            // PagerDuty service to create incidents for
	ClusterName  string            // Kubernetes cluster name for incident correlation
	CustomFields map[string]string // Additional fields to include in incidents
	Retry        RetryConfig      // Retry configuration for API calls
}

// Add default retry configuration
var defaultRetryConfig = RetryConfig{
	MaxRetries:  3,
	InitialWait: 1 * time.Second,
	MaxWait:     10 * time.Second,
}

var defaultTimeoutConfig = TimeoutConfig{
	Create:  30 * time.Second,
	Resolve: 30 * time.Second,
}

// client implements the PagerDuty incident management integration
type client struct {
	target.BaseClient
	client        *pagerduty.Client
	serviceID     string
	clusterName   string
	customFields  map[string]string
	retryConfig   RetryConfig
	timeoutConfig TimeoutConfig
}

// incidentMetadata contains fields used to correlate and deduplicate incidents
type incidentMetadata struct {
	ClusterName  string // Kubernetes cluster name
	PolicyName   string // Policy that triggered the incident
	RuleName     string // Specific rule that was violated
	ResourceKind string // Kind of resource that violated the policy
	ResourceNS   string // Namespace of the violating resource
	ResourceName string // Name of the violating resource
}

// retryWithExponentialBackoff attempts an operation with exponential backoff
// It will retry failed operations up to MaxRetries times, with increasing delays
func (p *client) retryWithExponentialBackoff(ctx context.Context, operation func() error) error {
	var lastErr error
	wait := p.retryConfig.InitialWait

	for i := 0; i <= p.retryConfig.MaxRetries; i++ {
		select {
		case <-ctx.Done():
			return fmt.Errorf("operation timed out: %w", ctx.Err())
		default:
			if err := operation(); err != nil {
				lastErr = err
				if i == p.retryConfig.MaxRetries {
					break
				}

				zap.L().Warn("PagerDuty API call failed, retrying",
					zap.Int("attempt", i+1),
					zap.Int("maxRetries", p.retryConfig.MaxRetries),
					zap.Duration("wait", wait),
					zap.Duration("maxWait", p.retryConfig.MaxWait),
					zap.String("target", p.Name()),
					zap.Error(err),
				)

				timer := time.NewTimer(wait)
				select {
				case <-ctx.Done():
					timer.Stop()
					return fmt.Errorf("operation timed out during retry: %w", ctx.Err())
				case <-timer.C:
				}

				wait *= 2
				if wait > p.retryConfig.MaxWait {
					wait = p.retryConfig.MaxWait
				}
				continue
			}
			return nil
		}
	}

	return fmt.Errorf("operation failed after %d retries: %w", p.retryConfig.MaxRetries, lastErr)
}

// findExistingIncident searches for an open incident matching the given result
// Returns the matching incident if found, nil otherwise
func (p *client) findExistingIncident(ctx context.Context, result v1alpha2.PolicyReportResult) (*pagerduty.Incident, error) {
	// Query for open incidents with matching metadata
	opts := pagerduty.ListIncidentsOptions{
		ServiceIDs: []string{p.serviceID},
		Statuses:   []string{"triggered", "acknowledged"},
	}
	
	incidents, err := p.client.ListIncidents(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list incidents: %w", err)
	}

	metadata := incidentMetadata{
		ClusterName: p.clusterName,
		PolicyName:  result.Policy,
		RuleName:    result.Rule,
	}
	if result.HasResource() {
		res := result.GetResource()
		metadata.ResourceKind = res.Kind
		metadata.ResourceNS = res.Namespace 
		metadata.ResourceName = res.Name
	}

	// Find matching incident by checking custom fields
	for _, incident := range incidents {
		if matchesMetadata(incident, metadata) {
			return &incident, nil
		}
	}

	return nil, nil
}

// resolveIncident marks an incident as resolved with the given resolution message
func (p *client) resolveIncident(ctx context.Context, incidentID string, resolution string) error {
	incident := pagerduty.ManageIncidentsOptions{
		ID: incidentID,
		Incidents: []pagerduty.ManageIncident{
			{
				Status:     "resolved",
				Resolution: resolution,
			},
		},
	}

	return p.retryWithExponentialBackoff(ctx, func() error {
		return p.client.ManageIncidents("policy-reporter", &incident)
	})
}

// Send processes a policy result and manages corresponding PagerDuty incidents
// - Creates new incidents for failing policy results
// - Resolves incidents when policies pass or are deleted
// - Deduplicates incidents for the same policy violation
func (p *client) Send(result v1alpha2.PolicyReportResult) {
	ctx, cancel := context.WithTimeout(context.Background(), p.timeoutConfig.Create)
	defer cancel()

	// For pass results or deletions, resolve any existing incidents
	if result.Result == v1alpha2.StatusPass || result.Result == "" {
		incident, err := p.findExistingIncident(ctx, result)
		if err != nil {
			zap.L().Error("failed to check for existing incidents", zap.Error(err))
			return
		}
		if incident != nil {
			resolution := "Policy violation has been resolved"
			if result.Result == "" {
				resolution = "Policy or resource has been deleted"
			}
			if err := p.resolveIncident(ctx, incident.ID, resolution); err != nil {
				zap.L().Error("failed to resolve incident", zap.Error(err))
			}
		}
		return
	}

	if result.Result != v1alpha2.StatusFail {
		return
	}

	// Check for existing incident before creating new one
	incident, err := p.findExistingIncident(ctx, result)
	if err != nil {
		zap.L().Error("failed to check for existing incidents", zap.Error(err))
		return
	}
	if incident != nil {
		return // Incident already exists
	}

	// Create new incident with metadata
	details := map[string]interface{}{
		"cluster":   p.clusterName,
		"policy":    result.Policy,
		"rule":      result.Rule,
		"message":   result.Message,
		"severity":  result.Severity,
	}

	if result.HasResource() {
		res := result.GetResource()
		details["resource"] = formatting.ResourceString(res)
		details["resource_kind"] = res.Kind
		details["resource_namespace"] = res.Namespace
		details["resource_name"] = res.Name
	}

	for k, v := range p.customFields {
		details[k] = v
	}

	for k, v := range result.Properties {
		details[k] = v
	}

	incidentOpts := pagerduty.CreateIncidentOptions{
		Type:    "incident",
		Title:   fmt.Sprintf("[%s] Policy Violation: %s - Rule: %s", p.clusterName, result.Policy, result.Rule),
		Service: &pagerduty.APIReference{ID: p.serviceID, Type: "service_reference"},
		Body: &pagerduty.APIDetails{
			Type:    "incident_body",
			Details: details,
		},
		Urgency: mapSeverityToUrgency(result.Severity),
	}

	// Create the incident with retry
	err = p.retryWithExponentialBackoff(ctx, func() error {
		_, err := p.client.CreateIncident("policy-reporter", &incidentOpts)
		return err
	})

	if err != nil {
		zap.L().Error("failed to create incident",
			zap.String("policy", result.Policy),
			zap.String("rule", result.Rule),
			zap.Error(err),
		)
	}
}

// mapSeverityToUrgency converts policy severity levels to PagerDuty urgency levels
func mapSeverityToUrgency(severity v1alpha2.PolicySeverity) string {
	switch severity {
	case v1alpha2.SeverityCritical, v1alpha2.SeverityHigh:
		return "high"
	default:
		return "low"
	}
}

// SetClient replaces the internal PagerDuty client, primarily used for testing
func (p *client) SetClient(c interface{}) {
	if pdClient, ok := c.(interface {
		CreateIncident(string, *pagerduty.CreateIncidentOptions) (*pagerduty.Incident, error)
		ManageIncidents(string, *pagerduty.ManageIncidentsOptions) error
	}); ok {
		p.client = pdClient
	}
}

// NewClient creates a new PagerDuty client with the given configuration
func NewClient(options Options) target.Client {
	if options.Retry.MaxRetries == 0 {
		options.Retry = defaultRetryConfig
	}

	clusterName := options.ClusterName
	if clusterName == "" {
		clusterName = "default"
	}

	return &client{
		BaseClient:    target.NewBaseClient(options.ClientOptions),
		client:        pagerduty.NewClient(options.APIToken),
		serviceID:     options.ServiceID,
		clusterName:   clusterName,
		customFields:  options.CustomFields,
		retryConfig:   options.Retry,
		timeoutConfig: defaultTimeoutConfig,
	}
}

// matchesMetadata checks if an incident's metadata matches the given fields
// Used to correlate and deduplicate incidents across policy results
func matchesMetadata(incident pagerduty.Incident, metadata incidentMetadata) bool {
	details, ok := incident.Body.Details.(map[string]interface{})
	if !ok {
		return false
	}

	// Check policy and rule match
	if details["policy"] != metadata.PolicyName || details["rule"] != metadata.RuleName {
		return false
	}

	// Check cluster match
	if details["cluster"] != metadata.ClusterName {
		return false
	}

	// If resource info exists, check that too
	if metadata.ResourceKind != "" {
		if details["resource_kind"] != metadata.ResourceKind ||
		   details["resource_namespace"] != metadata.ResourceNS ||
		   details["resource_name"] != metadata.ResourceName {
			return false
		}
	}

	return true
} 