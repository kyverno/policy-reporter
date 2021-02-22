# PolicyReporter

## Motivation

Kyverno ships with two types of validation. You can either enforce a rule or audit it. If you don't want to block developers or if you want to try out a new rule, you can use the audit functionality. The audit configuration creates [PolicyReports](https://kyverno.io/docs/policy-reports/) which you can access with `kubectl`. Because I can't find a simple solution to get a general overview of this PolicyReports and PolicyReportResults, I created this tool to send information from PolicyReports to [Grafana Loki](https://grafana.com/oss/loki/). As additional feature this tool provides an http server with Prometheus Metrics about ReportPolicy Summaries and ReportPolicyRules.

This project is in an early stage. Please let me know if anything did not work as expected or if you want to send your audits to other targets then Loki.

## Installation with Helm v3

Installation via Helm Repository

```bash
helm repo add policy-reporter https://fjogeleit.github.io/policy-reporter
helm install policy-reporter policy-reporter/policy-reporter --set loki.host=http://lokihost:3100 -n policy-reporter --create-namespace
```
You can also customize the `./charts/policy-reporter/values.yaml` to change the default configurations.

### Additional configurations for Loki

Configure `loki.minimumPriority` to send only results with the configured minimumPriority or above, empty means all results. (info < warning < error)
Configure `loki.skipExistingOnStartup` to skip all results who already existed before the PolicyReporter started. Can be used after the first deployment to prevent duplicated events.

```yaml
loki:
  minimumPriority: ""
  skipExistingOnStartup: false
```

### Configure Policy Priorities

By default kyverno PolicyReports has no priority or severity for policies. So every passed rule validation will be processed as notice, a failed validation is processed as error. To customize this you can configure a mapping from policies to fail priorities. So you can send them as warnings instead of errors. To configure the priorities create a ConfigMap in the `policy-reporter` namespace with the name `policy-reporter-config`. This ConfigMap have to have a property `config.yaml` with the map as YAML content. See the Example for Detailes.

#### Example

```yaml
# config.yaml
policy_priorities:
  check-label-app: warning
  require-ns-labels: warning
```

```bash
kubectl create configmap policy-reporter-config --from-file=config.yaml -n policy-reporter
```

## Monitoring

The Helm Chart includes optional Manifests for the [MonitoringStack](https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-prometheus-stack). The provided Dashboard works without Loki

* Enable a ServiceMonitor by setting `metrics.serviceMonitor` to `true`.
* Enable a basic Dashboard as ConfigMap by setting `metrics.dashboard.enabled` to `true`.
    * Change the namespace to your required monitoring namespace by changing `metrics.dashboard.namespace` (default: cattle-dashboards)


If you are not using the MonitoringStack you can get the dashboard configuration from this [Gist](https://gist.github.com/fjogeleit/bf540421fd28989fc92841177be972bc)

Example Installation
```bash
helm install policy-reporter policy-reporter/policy-reporter --set metrics.serviceMonitor=true --set metrics.dashboard.enabled=true -n policy-reporter --create-namespace
```

#### Dashboard Preview

![PolicyReporter Grafana Dashboard](https://github.com/fjogeleit/policy-reporter/blob/main/docs/images/policy-reports-dashboard.png?raw=true)

## Example Outputs

![Grafana Loki](https://github.com/fjogeleit/policy-reporter/blob/main/docs/images/grafana-loki.png?raw=true)

![Prometheus Metrics](https://github.com/fjogeleit/policy-reporter/blob/main/docs/images/prometheus.png?raw=true)

# Todos
* ~~Support for ClusterPolicyReports~~
* Additional Targets
