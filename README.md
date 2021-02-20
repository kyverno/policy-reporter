# PolicyReporter

## Motivation

Kyverno ships with two types of validation. You can either enforce a rule or audit it. If you don't want to block developers or if you want to try out what a new rule you can audit it. The audit configuration creates PolicyReports which you can describe or read with over `kubectl` but it's not that easy to get a good overview. To solve this problem this tool sends informations from PolicyReports to Loki and provide a Metrics endpoint to get metrics about summaries and each rule result.

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

# Todos
* Support for ClusterPolicyReports
* Additional Targets