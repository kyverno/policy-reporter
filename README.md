# PolicyReporter

PolicyReporter is a simple tool to watch for PolicyReports in your cluster.

It uses this resources to create Prometheus Metrics from it. It also provides a configuration to push rule validation results to Grafana Loki.

## Installation with Helm v3

Clone the repository and use the following command:

```bash
git clone https://github.com/fjogeleit/policy-reporter.git

cd policy-reporter

helm install policy-reporter ./charts/policy-reporter --set loki=http://lokihost:3100 -n policy-reporter --create-namespace
```
You can also customize the `./charts/policy-reporter/values.yaml` to change the default configurations.

### Configure policyPriorities

By default kyverno PolicyReports has no priority or severity for policies. So every passed rule validation will be processed as notice, a failed validation is processed as error. To customize this you can configure a mapping from policies to fail priorities. So you can send them as warnings instead of errors.

```yaml
# values.yaml
# policyPriorities example diff

policyPriorities:
    check-label-app: warning
```

## Example Outputs

![Grafana Loki](https://github.com/fjogeleit/policy-reporter/blob/main/docs/images/grafana-loki.png?raw=true)

![Prometheus Metrics](https://github.com/fjogeleit/policy-reporter/blob/main/docs/images/prometheus.png?raw=true)
