# AlertManager Target Test

This is a simple test program to verify the AlertManager target integration.

## Prerequisites

- Running AlertManager instance (default: http://localhost:9093)
- Go environment

## How to run

```bash
# From the project root
cd test/alertmanager
go run main.go
```

## Expected results

1. The test sends two types of alerts to the AlertManager:
   - A single test alert with policy "test-policy"
   - Two batch alerts with policies "batch-policy-1" and "batch-policy-2"

2. Check the AlertManager UI to verify the alerts are received with:
   - Correct labels (severity, status, source, policy, rule)
   - Correct annotations (message, category, custom properties)
   - Custom fields (environment: test, test: true)

## Troubleshooting

- If AlertManager is running on a different URL, edit `alertManagerURL` in main.go
- Ensure the AlertManager API endpoint is accessible
- Check AlertManager logs for any errors 