# PolicyReporter

## Motivation

Kyverno ships with two types of validation. You can either enforce a rule or audit it. If you don't want to block developers or if you want to try out a new rule, you can use the audit functionality. The audit configuration creates [PolicyReports](https://kyverno.io/docs/policy-reports/) which you can access with `kubectl`. Because I can't find a simple solution to get a general overview of this PolicyReports and PolicyReportResults, I created this tool to send information from PolicyReports to [Grafana Loki](https://grafana.com/oss/loki/). As additional feature this tool provides an http server with Prometheus Metrics about ReportPolicy Summaries and ReportPolicyRules.

This project is in an early stage. Please let me know if anything did not work as expected or if you want so send your audits to other targets then Loki.

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