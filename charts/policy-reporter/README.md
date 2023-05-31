# Policy Reporter

![Version: v2.19.2](https://img.shields.io/badge/Version-v2.19.2-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: v2.15.2](https://img.shields.io/badge/AppVersion-v2.15.2-informational?style=flat-square)

## Motivation

Kyverno ships with two types of validation. You can either enforce a rule or audit it. If you don't want to block developers or if you want to try out a new rule, you can use the audit functionality. The audit configuration creates [PolicyReports](https://kyverno.io/docs/policy-reports/) which you can access with `kubectl`. Because I can't find a simple solution to get a general overview of this PolicyReports and PolicyReportResults, I created this tool to send information about PolicyReports to different targets like [Grafana Loki](https://grafana.com/oss/loki/), [Elasticsearch](https://www.elastic.co/de/elasticsearch/) or [Slack](https://slack.com/). 

## Documentation

You can find detailed Information and Screens about Features and Configurations in the [Documentation](https://kyverno.github.io/policy-reporter/guide/02-getting-started#core--policy-reporter-ui).

## Getting Started

## Installation with Helm v3

Installation via Helm Repository

### Add the Helm repository
```bash
helm repo add policy-reporter https://kyverno.github.io/policy-reporter
helm repo update
```

### Basic Installation

The basic installation provides an Prometheus Metrics Endpoint and different REST APIs, for more details have a look at the [Documentation](https://kyverno.github.io/policy-reporter/guide/02-getting-started).

```bash
helm install policy-reporter policy-reporter/policy-reporter -n policy-reporter --create-namespace
```

## Policy Reporter UI

You can use the Policy Reporter as standalone Application along with the optional UI SubChart.

### Installation with Policy Reporter UI and Kyverno Plugin enabled
```bash
helm install policy-reporter policy-reporter/policy-reporter --set kyvernoPlugin.enabled=true --set ui.enabled=true --set ui.plugins.kyverno=true -n policy-reporter --create-namespace
kubectl port-forward service/policy-reporter-ui 8082:8080 -n policy-reporter
```
Open `http://localhost:8082/` in your browser.

Check the [Documentation](https://kyverno.github.io/policy-reporter/guide/02-getting-started#core--policy-reporter-ui) for Screens and additional Information

## Resources

* [[Video] 37. #EveryoneCanContribute cafe: Policy reporter for Kyverno](https://youtu.be/1mKywg9f5Fw)
* [[Video] Rawkode Live: Hands on Policy Reporter](https://www.youtube.com/watch?v=ZrOtTELNLyg)
* [[Blog] Monitor Security and Best Practices with Kyverno and Policy Reporter](https://blog.webdev-jogeleit.de/blog/monitor-security-with-kyverno-and-policy-reporter/)
