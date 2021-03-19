# Policy Reporter
[![CI](https://github.com/fjogeleit/policy-reporter/actions/workflows/ci.yaml/badge.svg)](https://github.com/fjogeleit/policy-reporter/actions/workflows/ci.yaml) [![Go Report Card](https://goreportcard.com/badge/github.com/fjogeleit/policy-reporter)](https://goreportcard.com/report/github.com/fjogeleit/policy-reporter) [![Coverage Status](https://coveralls.io/repos/github/fjogeleit/policy-reporter/badge.svg?branch=main)](https://coveralls.io/github/fjogeleit/policy-reporter?branch=main)

## Motivation

Kyverno ships with two types of validation. You can either enforce a rule or audit it. If you don't want to block developers or if you want to try out a new rule, you can use the audit functionality. The audit configuration creates [PolicyReports](https://kyverno.io/docs/policy-reports/) which you can access with `kubectl`. Because I can't find a simple solution to get a general overview of this PolicyReports and PolicyReportResults, I created this tool to send information about PolicyReports to different targets like [Grafana Loki](https://grafana.com/oss/loki/), [Elasticsearch](https://www.elastic.co/de/elasticsearch/) or [Slack](https://slack.com/). 

Policy Reporter provides also a Prometheus Metrics API as well as an standalone mode along with the [Policy Reporter UI](#policy-report-ui).

This project is in an early stage. Please let me know if anything did not work as expected or if you want to send your audits to unsupported targets.

## Documentation

You can find detailed Information about Features and Configurations in the [Documentation](https://github.com/fjogeleit/policy-reporter/wiki).

## Getting Started

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

## Policy Reporter UI

You can use the Policy Reporter as standalone Application along with the optional UI SubChart.

### Installation with Policy Reporter UI enabled
```bash
helm install policy-reporter policy-reporter/policy-reporter --set ui.enabled=true -n policy-reporter --create-namespace
kubectl port-forward service/policy-reporter-ui 8082:8080 -n policy-reporter
```
Open `http://localhost:8082/` in your browser.

Check the [Documentation](https://github.com/fjogeleit/policy-reporter/wiki/policy-reporter-ui) for Screens and additional Information

## Targets

Policy Reporter supports the following [Targets](https://github.com/fjogeleit/policy-reporter/wiki/targets) to send new (Cluster)PolicyReport Results too:
* [Grafana Loki](https://github.com/fjogeleit/policy-reporter/wiki/grafana-loki)
* [Elasticsearch](https://github.com/fjogeleit/policy-reporter/wiki/elasticsearch)
* [Slack](https://github.com/fjogeleit/policy-reporter/wiki/slack)
* [Discord](https://github.com/fjogeleit/policy-reporter/wiki/discord)
* [MS Teams](https://github.com/fjogeleit/policy-reporter/wiki/ms-teams)


## Monitoring

The Helm Chart includes optional SubChart for [Prometheus Operator](https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-prometheus-stack) Integration. The provided Dashboards working without Loki and are only based on the Prometheus Metrics.

Have a look into the [Documentation](https://github.com/fjogeleit/policy-reporter/wiki/prometheus-operator-integration) for details.

### Grafana Dashboard Import

If you are not using the MonitoringStack you can import the dashboards from [Grafana](https://grafana.com/orgs/policyreporter/dashboards)
