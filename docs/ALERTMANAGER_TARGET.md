# AlertManager Integration

This document explains how to set up the AlertManager integration with Policy Reporter to receive policy violation alerts.

## Overview

The AlertManager integration allows Policy Reporter to send policy violation notifications to AlertManager, which can then route and manage alerts based on its own configuration. This integration enables you to:

- Monitor and get notified about Kubernetes policy violations
- Route alerts to different notification channels (email, Slack, etc.) based on severity, source, or other criteria
- Use AlertManager's features like silencing, grouping, and inhibition for policy violation alerts

## Configuration

### 1. Policy Reporter TargetConfig

Create a TargetConfig to connect Policy Reporter to AlertManager:

```yaml
apiVersion: policyreporter.kyverno.io/v1alpha1
kind: TargetConfig
metadata:
  name: alertmanager-target
spec:
  alertManager:
    host: http://alertmanager.monitoring:9093
    skipTLS: true
    minimumSeverity: warning
    sources:
      - kyverno
      - policy-reporter
```

Configuration options:

| Option | Description | Required | Default |
|--------|-------------|----------|---------|
| `host` | URL of the AlertManager API | Yes | - |
| `skipTLS` | Skip TLS verification | No | `false` |
| `certificate` | Path to custom certificate | No | - |
| `minimumSeverity` | Minimum severity level to send (info, warning, error, critical) | No | - |
| `sources` | List of sources to accept alerts from | No | All sources |
| `headers` | Custom HTTP headers | No | - |

### 2. AlertManager Configuration

Configure AlertManager to receive and properly route the alerts:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: alertmanager-config
data:
  alertmanager.yml: |
    global:
      resolve_timeout: 5m
    route:
      group_by: ['alertname', 'severity', 'policy', 'rule', 'source']
      group_wait: 10s
      group_interval: 10s
      repeat_interval: 1h
      receiver: 'default'
    receivers:
    - name: 'default'
      webhook_configs:
      - url: 'http://your-webhook-endpoint'
        send_resolved: true
```

## Alert Format

Policy Reporter sends alerts to AlertManager with the following structure:

### Labels

- `alertname`: "PolicyReporterViolation"
- `severity`: The policy severity (info, warning, medium, high, critical)
- `status`: The result status (pass, fail, warn, error, skip)
- `source`: The policy source (kyverno, policy-reporter, etc.)
- `policy`: The policy name
- `rule`: The rule name

### Annotations

- `message`: The violation message
- `category`: The policy category (if available)
- `resource`: The resource string (namespace/kind/name)
- `resource_kind`: The resource kind
- `resource_name`: The resource name
- `resource_namespace`: The resource namespace
- `resource_apiversion`: The resource API version

Additional properties from the policy result are also included as annotations.

## Verification

After setting up the integration, verify it's working by:

1. Port-forwarding to the AlertManager service:
   ```bash
   kubectl port-forward svc/alertmanager -n monitoring 9093:9093
   ```

2. Accessing the AlertManager UI at http://localhost:9093

3. Checking for alerts via the API:
   ```bash
   curl -s http://localhost:9093/api/v2/alerts | jq
   ```

## Troubleshooting

If you're not seeing alerts in AlertManager:

1. Verify the TargetConfig is correctly configured and applied
2. Check Policy Reporter logs for any errors when sending alerts
3. Ensure AlertManager is accessible from the Policy Reporter pod
4. Verify the AlertManager configuration has the correct receiver setup 