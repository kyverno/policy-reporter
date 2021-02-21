# PolicyReporter

Install with Helm

```bash
helm repo add policy-reporter https://fjogeleit.github.io/policy-reporter
helm install policy-reporter policy-reporter/policy-reporter --set loki=http://lokihost:3100 -n policy-reporter --create-namespace
```