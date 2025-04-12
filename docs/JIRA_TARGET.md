# JIRA Target for Policy Reporter

This guide explains how to configure Policy Reporter to send policy violations to JIRA as issues.

## Configuration via Helm Values

You can configure the JIRA target directly in the Helm values when installing or upgrading Policy Reporter:

```yaml
target:
  jira:
    # JIRA server URL
    host: "https://your-jira-instance.atlassian.net"
    # JIRA credentials - username and either API token or password
    username: "your-jira-username"
    apiToken: "your-jira-api-token"  # Preferred auth method
    # password: "your-jira-password" # Not recommended for production
    
    # JIRA project key (required)
    projectKey: "POL"
    # JIRA issue type (defaults to "Bug" if not specified)
    issueType: "Bug"
    
    # TLS configuration
    skipTLS: false
    certificate: ""
    
    # Minimum severity to report to JIRA
    # Values: "", "info", "warning", "error", "critical"
    minimumSeverity: "warning"
    
    # List of policy sources to include
    sources: []
    
    # Skip existing policy violations on startup
    skipExistingOnStartup: true
    
    # Filter policy violations
    filter:
      namespaces:
        include: []
        exclude: []
      policies:
        include: []
        exclude: []
      reportLabels:
        include: []
        exclude: []
      
    # Custom fields to add to JIRA issues
    customFields:
      # Map JIRA custom field IDs to values
      # customfield_10001: "Policy Violation"
      # customfield_10002: "Kubernetes"
```

## Configuration via TargetConfig CRD

You can also configure the JIRA target using Kubernetes custom resources by enabling the TargetConfig CRD:

```yaml
target:
  crd: true
```

Then apply a TargetConfig resource:

```yaml
apiVersion: policyreporter.io/v1alpha1
kind: TargetConfig
metadata:
  name: policy-reporter-jira
spec:
  jira:
    host: "https://your-jira-instance.atlassian.net"
    username: "your-jira-username"
    apiToken: "your-jira-api-token"
    projectKey: "POL"
    issueType: "Bug"
    minimumSeverity: "warning"
    filter:
      namespaces:
        include:
          - "default"
          - "kube-system"
      policies:
        include:
          - "require-pod-probes"
          - "require-resources-*"
```

## Using Secrets for Authentication

For secure credential management, you can store JIRA credentials in a Kubernetes Secret and reference it from your configuration:

1. Create a Secret with your JIRA credentials:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: jira-credentials
  namespace: policy-reporter
type: Opaque
stringData:
  username: "your-jira-username"
  apiToken: "your-jira-api-token"
  host: "https://your-jira-instance.atlassian.net"
  projectKey: "POL"
```

2. Reference the Secret in your configuration:

```yaml
target:
  jira:
    secretRef: "jira-credentials"
```

## Issue Format

Policy violations reported to JIRA will have the following format:

- **Summary**: Policy Violation: [policy-name]
- **Issue Type**: Bug (configurable)
- **Description**: Contains detailed information about the policy violation, including:
  - Policy name
  - Severity
  - Status
  - Category (if available)
  - Source (if available)
  - Resource details (kind, name, namespace)
  - Violation message
  - Additional properties
- **Labels**: 
  - policy-reporter
  - policy-violation
  - policy-[policy-name]
  - severity-[severity-level]

## Troubleshooting

If issues are not being created in JIRA:

1. Check the Policy Reporter logs for any errors related to JIRA communication
2. Verify your authentication credentials are correct
3. Ensure the configured JIRA project exists and the user has permission to create issues
4. Confirm the issue type specified exists in your JIRA project
5. Check that policy violations meet the configured minimum severity threshold 