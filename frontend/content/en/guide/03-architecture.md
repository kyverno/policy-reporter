---
title: Architecture
description: ''
position: 3
category: 'Guide'
---

Policy Reporter consists of up to 3 components.

<img src="/images/policy-reporter.svg" style="background-color: #fff; padding: 15px; width: 100%;">

## Policy Reporter

This is the core application and watches for PolicyReporter and ClusterPolicyReporter CRD resources in the cluster. Policy Reporter uses internal Listeners to react on incomming events and apply its logic on it.

* __MetricsListener__ (optional) creates metrics based on new, updated and removed resources
* __StoreListener__ (optional) persists new resources and every change of an existing resource in an internal representation in the included SQLite Database
* __ResultsListener__ checks each new and updated report for new results and publishes them to all registered __PolicyResultListeners__
* __SendResultListener__ is an PolicyResultListener and sends all new results to the configured targets

Processed information is available over the embedded HTTP Server as API endpoints. See [API Reference](/core/07-api-reference) for details.

## Policy Reporter Kyverno Plugin

This component watches for the Kyverno (Cluster)Policy CRDs like the Policy Reporter core application for the (Cluster)PolicyReport CRDs. The collected data is transformed into a internal format and available over the embedded HTTP Server as API endpoints.

This component is independent from the Policy Reporter core application and only consumed by the Policy Reporter UI.

## Policy Reporter UI

This component is an optional, standalone UI for information provided by the Policy Report core application (and Policy Reporter Kyverno Plugin). The intention was to provide a simple alternative to external monitoring solutions such as Grafana, which are not always available. The UI is a <a href="https://nuxtjs.org/" target="_blank">NuxtJS</a> based single-page application and is served over an Golang based static file server. This server also proxies all requests made by the UI to the Policy Reporter API.

### Kyverno Plugin

The Kyverno integration is an optional Plugin. If enabled it provides additional views about Kyverno policies. This information are provided by the *Policy Reporter Kyverno Plugin* which will also be proxied.
