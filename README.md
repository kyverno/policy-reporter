# PolicyReporter
[![CI](https://github.com/fjogeleit/policy-reporter/actions/workflows/ci.yaml/badge.svg)](https://github.com/fjogeleit/policy-reporter/actions/workflows/ci.yaml) [![Go Report Card](https://goreportcard.com/badge/github.com/fjogeleit/policy-reporter)](https://goreportcard.com/report/github.com/fjogeleit/policy-reporter) [![Coverage Status](https://coveralls.io/repos/github/fjogeleit/policy-reporter/badge.svg?branch=main)](https://coveralls.io/github/fjogeleit/policy-reporter?branch=main)

## Motivation

Kyverno ships with two types of validation. You can either enforce a rule or audit it. If you don't want to block developers or if you want to try out a new rule, you can use the audit functionality. The audit configuration creates [PolicyReports](https://kyverno.io/docs/policy-reports/) which you can access with `kubectl`. Because I can't find a simple solution to get a general overview of this PolicyReports and PolicyReportResults, I created this tool to send information from PolicyReports to different targets like [Grafana Loki](https://grafana.com/oss/loki/), [Elasticsearch](https://www.elastic.co/de/elasticsearch/) or [Slack](https://slack.com/). This tool provides by default an HTTP server with Prometheus Metrics on `http://localhost:2112/metrics` about ReportPolicy Summaries and ReportPolicyRules.

This project is in an early stage. Please let me know if anything did not work as expected or if you want to send your audits to other targets then Loki.

## Getting Started
* [Installation with Helm v3](#installation-with-helm-v3)
  * [Grafana Loki](#installation-with-loki)
  * [Elasticsearch](#installation-with-elasticsearch)
  * [Slack](#installation-with-slack)
  * [Discord](#installation-with-discord)
  * [Customization](#customization)
* [Configure Policy Priorities](#configure-policy-priorities)
* [Configure Monitoring](#monitoring)
* [Policy Report UI](#policy-report-ui)

## Installation with Helm v3

Installation via Helm Repository

### Add the Helm repository

```bash
helm repo add policy-reporter https://fjogeleit.github.io/policy-reporter
helm repo update
```

### Basic Installation - Provides Prometheus Metrics

```bash
helm install policy-reporter policy-reporter/policy-reporter -n policy-reporter --create-namespace
```

#### Example

![Prometheus Metrics](https://github.com/fjogeleit/policy-reporter/blob/main/docs/images/prometheus.png?raw=true)

### Installation with Loki

```bash
helm install policy-reporter policy-reporter/policy-reporter --set target.loki.host=http://loki:3100 -n policy-reporter --create-namespace
```
#### Additional configurations for Loki

* Configure `target.loki.minimumPriority` to send only results with the configured minimumPriority or above, empty means all results. (info < warning < error)
* Configure `target.loki.skipExistingOnStartup` to skip all results who already existed before the PolicyReporter started (default: `true`).

```yaml
target:
  loki:
    host: ""
    minimumPriority: ""
    skipExistingOnStartup: true
```

#### Example

![Grafana Loki](https://github.com/fjogeleit/policy-reporter/blob/main/docs/images/grafana-loki.png?raw=true)

### Installation with Elasticsearch

```bash
helm install policy-reporter policy-reporter/policy-reporter --set target.elasticsearch.host=http://elasticsearch:3100 -n policy-reporter --create-namespace
```

#### Additional configurations for Elasticsearch

* Configure `target.elasticsearch.index` to customize the elasticsearch index.
* Configure `target.elasticsearch.rotation` is added as suffix to the index. Possible values are `daily`, `monthly`, `annually` and `none`.
* Configure `target.elasticsearch.minimumPriority` to send only results with the configured minimumPriority or above, empty means all results. (info < warning < error)
* Configure `target.elasticsearch.skipExistingOnStartup` to skip all results who already existed before the PolicyReporter started (default: `true`).

```yaml
target:
  elasticsearch:
    host: ""
    index: "policy-reporter"
    rotation: "daily"
    minimumPriority: ""
    skipExistingOnStartup: true
```

#### Example

![Elasticsearch](https://github.com/fjogeleit/policy-reporter/blob/main/docs/images/elasticsearch.png?raw=true)

### Installation with Slack

```bash
helm install policy-reporter policy-reporter/policy-reporter --set target.slack.webhook=http://hook.slack -n policy-reporter --create-namespace
```

#### Additional configurations for Slack

* Configure `target.slack.minimumPriority` to send only results with the configured minimumPriority or above, empty means all results. (info < warning < error)
* Configure `target.slack.skipExistingOnStartup` to skip all results who already existed before the PolicyReporter started (default: `true`).

```yaml
target:
  slack:
    webhook: ""
    minimumPriority: ""
    skipExistingOnStartup: true
```

#### Example

![Slack](https://github.com/fjogeleit/policy-reporter/blob/main/docs/images/slack.png?raw=true)

### Installation with Discord

```bash
helm install policy-reporter policy-reporter/policy-reporter --set target.discord.webhook=http://hook.discord -n policy-reporter --create-namespace
```

#### Additional configurations for Discord

* Configure `target.discord.minimumPriority` to send only results with the configured minimumPriority or above, empty means all results. (info < warning < error)
* Configure `target.discord.skipExistingOnStartup` to skip all results who already existed before the PolicyReporter started (default: `true`).

```yaml
target:
  discord:
    webhook: ""
    minimumPriority: ""
    skipExistingOnStartup: true
```

#### Example

![Discord](https://github.com/fjogeleit/policy-reporter/blob/main/docs/images/discord.png?raw=true)

### Customization

You can combine multiple targets by setting the required `host` or `webhook` configuration for your targets of choice. For all possible configurations checkout the `./charts/policy-reporter/values.yaml` to change any available configuration.

## Configure Policy Priorities

By default kyverno PolicyReports has no priority or severity for policies. So every passed rule validation will be processed as notice, a failed validation is processed as error. To customize this you can configure a mapping from policies to fail priorities. So you can send them as debug, info or warnings instead of errors. To configure the priorities enale the required `Role` and `RoleBinding` by setting `policyPriorities.enabled` to `true` and create a ConfigMap in the `policy-reporter` namespace with the name `policy-reporter-priorities`. Configure each priority as value with the __Policyname__ as key and the __Priority__ as value. This Configuration is loaded and synchronized during runtime. Any change to this configmap will automaticly synchronized, no new deployment needed.

A special Policyname `default` is supported. The `default` configuration can be used to set a global default priority instead of `error`.

### Enable the required Role and RoleBinding

```bash
helm install policy-reporter policy-reporter/policy-reporter --set policyPriorities.enabled=true -n policy-reporter --create-namespace
```

### Create the ConfigMap
```bash
kubectl create configmap policy-reporter-priorities --from-literal check-label-app=warning --from-literal require-ns-labels=warning -n policy-reporter
```

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: policy-reporter-priorities
  namespace: policy-reporter
data:
  default: debug
  check-label-app: warning
  require-ns-labels: warning
```

## Monitoring

The Helm Chart includes optional Sub Chart for the [MonitoringStack](https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-prometheus-stack). The provided Dashboards working without Loki and are only based on the Prometheus Metrics.

* Enable the Monitoring by setting `monitoring.enabled` to `true`.
    * Change the `namespace` to your required monitoring namespace by changing `monitoring.namespace` (default: `cattle-dashboards`)
    * With `monitoring.serviceMonitor.labels` you can add additional labels to the `ServiceMonitor`. This helps to match the `serviceMonitorSelector` configuration of your Prometheus resource

### Example Installation
```bash
helm install policy-reporter policy-reporter/policy-reporter --set monitoring.enabled=true --set monitoring.namespace=cattle-dashboards -n policy-reporter --create-namespace
```

### Customized Dashboards

The Monitoring Subchart offers several values for changing the height or disabling different components of the individual dashboards.

To change a value of this subchart you have to prefix each option with `monitoring.`

#### Example

```bash
helm install policy-reporter policy-reporter/policy-reporter --set monitoring.enabled=true --set monitoring.policyReportDetails.secondStatusRow.enabled=false -n policy-reporter --create-namespace
```

#### PolicyReport Details Dashboard

| Value | Default |
|-------|---------|
| policyReportDetails.firstStatusRow.height | 6 |
| policyReportDetails.secondStatusRow.enabled | true |
| policyReportDetails.secondStatusRow.height | 2 |
| policyReportDetails.statusTimeline.enabled | true |
| policyReportDetails.statusTimeline.height | 8 |
| policyReportDetails.passTable.enabled | true |
| policyReportDetails.passTable.height | 8 |
| policyReportDetails.failTable.enabled | true |
| policyReportDetails.failTable.height | 8 |
| policyReportDetails.warningTable.enabled | true |
| policyReportDetails.warningTable.height | 4 |
| policyReportDetails.errorTable.enabled | true |
| policyReportDetails.errorTable.height | 4 |

#### ClusterPolicyReport Details Dashboard

| Value | Default |
|-------|---------|
| clusterPolicyReportDetails.statusRow.height | 6 |
| clusterPolicyReportDetails.statusTimeline.enabled | true |
| clusterPolicyReportDetails.statusTimeline.height | 8 |
| clusterPolicyReportDetails.passTable.enabled | true |
| clusterPolicyReportDetails.passTable.height | 8 |
| clusterPolicyReportDetails.failTable.enabled | true |
| clusterPolicyReportDetails.failTable.height | 8 |
| clusterPolicyReportDetails.warningTable.enabled | true |
| clusterPolicyReportDetails.warningTable.height | 4 |
| clusterPolicyReportDetails.errorTable.enabled | true |
| clusterPolicyReportDetails.errorTable.height | 4 |

#### PolicyReport Overview Dashboard

| Value | Default |
|-------|---------|
| policyReportOverview.failingSummaryRow.height | 8 |
| policyReportOverview.failingTimeline.height | 10 |
| policyReportOverview.failingPolicyRuleTable.height | 10 |
| policyReportOverview.failingClusterPolicyRuleTable.height | 10 |

#### Grafana Dashboard Import

If you are not using the MonitoringStack you can import the dashboards from [Grafana](https://grafana.com/orgs/policyreporter/dashboards)

### Dashboard Preview

![PolicyReporter Grafana Dashboard](https://github.com/fjogeleit/policy-reporter/blob/main/docs/images/policy-reports-dashboard.png?raw=true)

![PolicyReporter Details Grafana Dashboard](https://github.com/fjogeleit/policy-reporter/blob/main/docs/images/policy-details.png?raw=true)

![ClusterPolicyReporter Details Grafana Dashboard](https://github.com/fjogeleit/policy-reporter/blob/main/docs/images/cluster-policy-details.png?raw=true)

## Policy Report UI

If you don't have any supported Monitoring solution running, you can use the standalone [Policy Report UI](https://github.com/fjogeleit/policy-reporter-ui).

The UI is provided as optional Helm Sub Chart and can be enabled by setting `ui.enabled` to `true`. 

### Installation

```bash
helm install policy-reporter policy-reporter/policy-reporter --set ui.enabled=true -n policy-reporter --create-namespace
```

### Access it with Port Forward on localhost

```bash
kubectl port-forward service/policy-reporter-ui 8082:8080 -n policy-reporter
```

Open `http://localhost:8082/` in your browser.

### Example

The UI is an optional application and provides three different views with informations about the validation status of your audit policies.

![Dashboard](https://github.com/fjogeleit/policy-reporter-ui/blob/main/docs/images/dashboard.png?raw=true)

![Policy Reports](https://github.com/fjogeleit/policy-reporter-ui/blob/main/docs/images/policy-report.png?raw=true)

![ClusterPolicyReports](https://github.com/fjogeleit/policy-reporter-ui/blob/main/docs/images/cluster-policy-report.png?raw=true)

## Example Helm values.yaml

Example Helm `values.yaml` with the integrated Policy Reporter UI, Loki as target and customized Grafana Dashboards enabled.

```yaml
ui:
  enabled: true

policyPriorities:
  enabled: true

target:
  loki:
    host: "http://loki.loki-stack.svc.cluster.local:3100"
    minimumPriority: "warning"
    skipExistingOnStartup: true

monitoring:
  enabled: true

  policyReportDetails:
    firstStatusRow:
      height: 6
    secondStatusRow:
      enabled: false
      height: 2
    statusTimeline:
      enabled: true
      height: 8
    passTable:
      enabled: true
      height: 8
    failTable:
      enabled: true
      height: 8
    warningTable:
      enabled: false
      height: 4
    errorTable:
      enabled: false
      height: 4

  clusterPolicyReportDetails:
    statusRow:
      height: 6
    statusTimeline:
      enabled: true
      height: 8
    passTable:
      enabled: true
      height: 8
    failTable:
      enabled: true
      height: 8
    warningTable:
      enabled: false
      height: 4
    errorTable:
      enabled: false
      height: 4

  policyReportOverview:
    failingSummaryRow:
      height: 8
    failingTimeline:
      height: 10
    failingPolicyRuleTable:
      height: 10
    failingClusterPolicyRuleTable:
      height: 10
```

# Todos
* ~~Support for ClusterPolicyReports~~
* ~~Additional Targets~~
* Filter
