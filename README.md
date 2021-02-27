# PolicyReporter
[![CI](https://github.com/fjogeleit/policy-reporter/actions/workflows/ci.yaml/badge.svg)](https://github.com/fjogeleit/policy-reporter/actions/workflows/ci.yaml) [![Go Report Card](https://goreportcard.com/badge/github.com/fjogeleit/policy-reporter)](https://goreportcard.com/report/github.com/fjogeleit/policy-reporter) [![Coverage Status](https://coveralls.io/repos/github/fjogeleit/policy-reporter/badge.svg?branch=main)](https://coveralls.io/github/fjogeleit/policy-reporter?branch=main)

## Motivation

Kyverno ships with two types of validation. You can either enforce a rule or audit it. If you don't want to block developers or if you want to try out a new rule, you can use the audit functionality. The audit configuration creates [PolicyReports](https://kyverno.io/docs/policy-reports/) which you can access with `kubectl`. Because I can't find a simple solution to get a general overview of this PolicyReports and PolicyReportResults, I created this tool to send information from PolicyReports to different targets like [Grafana Loki](https://grafana.com/oss/loki/). This tool provides by default an HTTP server with Prometheus Metrics on `http://localhost:2112/metrics` about ReportPolicy Summaries and ReportPolicyRules.

This project is in an early stage. Please let me know if anything did not work as expected or if you want to send your audits to other targets then Loki.

## Installation with Helm v3

Installation via Helm Repository

### Add the Helm repository

```bash
helm repo add policy-reporter https://fjogeleit.github.io/policy-reporter
```

### Basic Installation - Provides Prometheus Metrics

```bash
helm install policy-reporter policy-reporter/policy-reporter -n policy-reporter --create-namespace
```

### Installation with Loki

```bash
helm install policy-reporter policy-reporter/policy-reporter --set loki.host=http://loki:3100 -n policy-reporter --create-namespace
```

### Installation with Elasticsearch

```bash
helm install policy-reporter policy-reporter/policy-reporter --set elasticsearch.host=http://elasticsearch:3100 -n policy-reporter --create-namespace
```

You can also customize the `./charts/policy-reporter/values.yaml` to change the default configurations.

### Additional configurations for Loki

* Configure `loki.minimumPriority` to send only results with the configured minimumPriority or above, empty means all results. (info < warning < error)
* Configure `loki.skipExistingOnStartup` to skip all results who already existed before the PolicyReporter started (default: `true`).

```yaml
loki:
  minimumPriority: ""
  skipExistingOnStartup: true
```

### Additional configurations for Elasticsearch

* Configure `elasticsearch.index` to customize the elasticsearch index.
* Configure `elasticsearch.rotation` is added as suffix to the index. Possible values are `daily`, `monthly`, `annually` and `none`.
* Configure `elasticsearch.minimumPriority` to send only results with the configured minimumPriority or above, empty means all results. (info < warning < error)
* Configure `elasticsearch.skipExistingOnStartup` to skip all results who already existed before the PolicyReporter started (default: `true`).

```yaml
elasticsearch:
  index: "policy-reporter"
  rotation: "daily"
  minimumPriority: ""
  skipExistingOnStartup: true
```

### Configure Policy Priorities

By default kyverno PolicyReports has no priority or severity for policies. So every passed rule validation will be processed as notice, a failed validation is processed as error. To customize this you can configure a mapping from policies to fail priorities. So you can send them as warnings instead of errors. To configure the priorities create a ConfigMap in the `policy-reporter` namespace with the name `policy-reporter-priorities`. Configure each priority as value with the Policyname as key and the Priority as value. This Configuration is loaded and synchronized during runtime. Any change to this configmap will automaticly synchronized, no new deployment needed.

#### Example

```bash
kubectl create configmap policy-reporter-priorities --from-literal check-label-app=warning --from-literal require-ns-labels=warning -n policy-reporter
```

## Monitoring

The Helm Chart includes optional Manifests for the [MonitoringStack](https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-prometheus-stack). The provided Dashboard works without Loki

* Enable a ServiceMonitor by setting `metrics.serviceMonitor` to `true`.
* Enable a basic Dashboard as ConfigMap by setting `metrics.dashboard.enabled` to `true`.
    * Change the namespace to your required monitoring namespace by changing `metrics.dashboard.namespace` (default: cattle-dashboards)


If you are not using the MonitoringStack you can import the dashboard from [Grafana](https://grafana.com/grafana/dashboards/13968)

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
