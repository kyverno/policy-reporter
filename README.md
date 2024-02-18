# Policy Reporter
[![CI](https://github.com/kyverno/policy-reporter/actions/workflows/ci.yaml/badge.svg)](https://github.com/kyverno/policy-reporter/actions/workflows/ci.yaml) [![Go Report Card](https://goreportcard.com/badge/github.com/kyverno/policy-reporter)](https://goreportcard.com/report/github.com/kyverno/policy-reporter) [![Coverage Status](https://coveralls.io/repos/github/kyverno/policy-reporter/badge.svg?branch=main)](https://coveralls.io/github/kyverno/policy-reporter?branch=main)

## Motivation

Kyverno ships with two types of validation. You can either enforce a rule or audit it. If you don't want to block developers or if you want to try out a new rule, you can use the audit functionality. The audit configuration creates [PolicyReports](https://kyverno.io/docs/policy-reports/) which you can access with `kubectl`. Because I can't find a simple solution to get a general overview of this PolicyReports and PolicyReportResults, I created this tool to send information about PolicyReports to different targets like [Grafana Loki](https://grafana.com/oss/loki/), [Elasticsearch](https://www.elastic.co/de/elasticsearch/) or [Slack](https://slack.com/). 

Policy Reporter provides also a Prometheus Metrics API as well as an standalone mode along with the [Policy Reporter UI](https://kyverno.github.io/policy-reporter/guide/getting-started#core--policy-reporter-ui).

This project is in an early stage. Please let me know if anything did not work as expected or if you want to send your audits to unsupported targets.

## Preview Feature Docs

Documentation for upcoming features and changes for the new Policy Reporter UI v2 are located in [Docs](https://github.com/kyverno/policy-reporter/tree/3.x/docs)

* [Basic Installation](https://github.com/kyverno/policy-reporter/blob/3.x/docs/TUTORIALS.md)
* [OAUth2 / OpenIDConnect](https://github.com/kyverno/policy-reporter/blob/3.x/docs/UI_AUTH.md)

## Documentation

You can find detailed Information and Screens about Features and Configurations in the [Documentation](https://kyverno.github.io/policy-reporter).

## Getting Started

## Installation with Helm v3

Installation via Helm Repository

### Add the Helm repository
```bash
helm repo add policy-reporter https://kyverno.github.io/policy-reporter
helm repo update
```

### Basic Installation

The basic installation provides optional Prometheus Metrics and/or optional REST APIs, for more details have a look at the [Documentation](https://kyverno.github.io/policy-reporter/guide/getting-started).

```bash
helm install policy-reporter policy-reporter/policy-reporter -n policy-reporter --set metrics.enabled=true --set rest.enabled=true --create-namespace
```

### Installation without Helm or Kustomize

To install Policy Reporter without Helm or Kustomize have a look at [manifests](https://github.com/kyverno/policy-reporter/tree/main/manifest).

## Policy Reporter UI

You can use the Policy Reporter as standalone Application along with the optional UI SubChart.

### Installation with Policy Reporter UI and Kyverno Plugin enabled
```bash
helm install policy-reporter policy-reporter/policy-reporter --set kyvernoPlugin.enabled=true --set ui.enabled=true --set ui.plugins.kyverno=true -n policy-reporter --create-namespace
kubectl port-forward service/policy-reporter-ui 8082:8080 -n policy-reporter
```
Open `http://localhost:8082/` in your browser.

Check the [Documentation](https://kyverno.github.io/policy-reporter/guide/getting-started#core--policy-reporter-ui) for Screens and additional Information

## Targets

Policy Reporter supports the following [Targets](https://kyverno.github.io/policy-reporter/core/targets) to send new (Cluster)PolicyReport Results too:
* [Grafana Loki](https://kyverno.github.io/policy-reporter/core/targets#grafana-loki)
* [Elasticsearch](https://kyverno.github.io/policy-reporter/core/targets#elasticsearch)
* [Microsoft Teams](https://kyverno.github.io/policy-reporter/core/targets#microsoft-teams)
* [Slack](https://kyverno.github.io/policy-reporter/core/targets#slack)
* [Discord](https://kyverno.github.io/policy-reporter/core/targets#discord)
* [Policy Reporter UI](https://kyverno.github.io/policy-reporter/core/targets#policy-reporter-ui)
* [Webhook](https://kyverno.github.io/policy-reporter/core/targets#webhook)
* [S3](https://kyverno.github.io/policy-reporter/core/targets#s3-compatible-storage)
* [AWS Kinesis compatible Services](https://kyverno.github.io/policy-reporter/core/targets#kinesis-compatible-services)
* [AWS SecurityHub](https://kyverno.github.io/policy-reporter/core/targets#aws-securityhub)
* [Google Cloud Storage](https://kyverno.github.io/policy-reporter/core/targets/#google-cloud-storage)
* [Telegram](https://kyverno.github.io/policy-reporter/core/targets#telegram)
* [Google Chat](https://kyverno.github.io/policy-reporter/core/targets#google-chat)

## Monitoring

The Helm Chart includes optional SubChart for [Prometheus Operator](https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-prometheus-stack) Integration. The provided Dashboards working without Loki and are only based on the Prometheus Metrics.

Have a look into the [Documentation](https://kyverno.github.io/policy-reporter/guide/helm-chart-core/#configure-the-servicemonitor) for details.

### Grafana Dashboard Import

If you are not using the MonitoringStack you can import the dashboards from [Grafana](https://grafana.com/orgs/policyreporter/dashboards)

## Resources

* [[Video] 37. #EveryoneCanContribute cafe: Policy reporter for Kyverno](https://youtu.be/1mKywg9f5Fw)
* [[Video] Rawkode Live: Hands on Policy Reporter](https://www.youtube.com/watch?v=ZrOtTELNLyg)
* [[Blog] Monitor Security and Best Practices with Kyverno and Policy Reporter](https://blog.webdev-jogeleit.de/blog/monitor-security-with-kyverno-and-policy-reporter/)
