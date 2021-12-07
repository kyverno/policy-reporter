---
title: Helm Chart
description: 'Policy Reporter Helm Chart configuration'
position: 4
category: Guide
---

The available Helm Chart is the easiest way to install and configure Policy Reporter. Depending on your usage it provides several values to configure the features you are using and disable anything else.

<alert>

The following explanations focus on feature based configurations. Additional configurations like __resource limits__, __nodeSelector__, etc. are also possible for each component. See the `values.yaml` of the related chart for reference.

</alert>

## Usage

### Add the Policy Reporter Helm Repository

```bash
helm repo add policy-reporter https://kyverno.github.io/policy-reporter
helm repo update
```
### Installation

```bash
helm install policy-reporter policy-reporter/policy-reporter -f values.yaml --create-namespace -n policy-reporter
```

## Configuration

By default all available features are disabled. So it's up to the user to enable or configure the features you need. This approach reduces the required resources to a bare minimum.

<alert>

See the complete <a href="https://raw.githubusercontent.com/kyverno/policy-reporter/main/charts/policy-reporter/values.yaml" target="_blank">values.yaml</a> for reference.

</alert>

### Enable Metrics Endpoint

Metrics will be available at `http://policy-reporter:8080/metrics`

```yaml
metrics:
  enabled: true
```

### Enable REST Endpoints

REST APIs will be available at `http://policy-reporter:8080/v1/` see [API Reference](/core/07-api-reference) for details.

```yaml
rest:
  enabled: true
```

### Priority Mapping

Define a custom mapping for fail results based on the related __policy__, you can also overwrite the default priority for fail results without a severity. See [priority mapping](/core/08-priority-mapping) for details.

```yaml
policyPriorities:
  # used for all fail results without severity or concrete mapping
  default: warning
  # used for all fail results of the require-ns-labels policy independent of the severity
  require-ns-labels: error
```

### Enable Targets / Notification

Policy Reporter supports different targets to send notifications too. You can configure as many targets as you like, you can also configure different targets for different priorities or sources (like Kyverno, Kube Bench or Falco).

As example you can configure Grafana Loki by providing an accessable host:

```yaml
target:
  loki:
    # loki host address
    host: "http://loki.loki-stack.3100"
    # minimum priority "" < info < warning < critical < error
    minimumPriority: "warning"
    # Skip already existing PolicyReportResults on startup
    skipExistingOnStartup: true
    # Send only results from the given sources
    source:
    - kyverno
    - falco
```

<alert>

See [Targets](/core/06-targets) for all available targets and how to configure them.

</alert>

### Enable NetworkPolicy

If enabled the Helm Chart creates an NetworkPolicy to allow Policy Reporter egress traffic to the Kubernetes API (defaults to port `6443`) and ingress traffic to the Policy Reporter REST API from the Policy Reporter UI. Ingress- and Egress rules for additional targets or monitoring tools can be extended as needed.

```yaml
networkPolicy:
  enabled: true
  egress:
  - to:
    ports:
    - protocol: TCP
      port: 6443
  ingress: []
```

## Subcharts

Extend Policy Reporter with the __Policy Reporter UI__ and __Kyverno Plugin__ subcharts. The __Monitoring__ subchart helps you to link Policy Reporter to your <a href="https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-prometheus-stack" target="_blank">Prometheus Operator</a> setup.

## Policy Reporter UI

<alert>

Each value is prefixed with `ui` to be clear that it has to be configured under this subchart.

</alert>

<alert>

See the complete <a href="https://raw.githubusercontent.com/kyverno/policy-reporter/main/charts/policy-reporter/charts/ui/values.yaml" target="_blank">values.yaml</a> for reference.

</alert>


Enable the Policy Reporter UI subchart and the required Policy Reporter REST API.

```yaml
ui:
  enabled: true
```

### Display Mode

Policy Reporter supports a light- and dark mode. The default depends on your System/Browser configuration. To enforce a default mode configure it with:

```yaml
ui:
  displayMode: dark # available opions: dark / light
```

### Log size

Configure the maximal size of the Log in Policy Reporter UI

```yaml
ui:
  logSize: 200 # 200 items are stored in memory
```

### Kyverno Plugin integration

<alert>

Requires the `kyvernoPlugin` subchart to be enabled.

</alert>

Enable the UI integration of the Kyverno Plugin to get additional __views__ about Kyverno Policies.

```yaml
ui:
  plugins:
    kyverno: true
```

### Ingress

Serve the UI over a hostname with the integrated __Ingress__ support.

```yaml
ui:
 ingress:
    enabled: true
    annotations:
      nginx.ingress.kubernetes.io/rewrite-target: /$1
    hosts:
      - host: domain.com
        paths:
        - "/(.*)"
```

### Enable NetworkPolicy

If enabled it creates an additional NetworkPolicy to allow ingress traffic to the Policy Reporter UI on the default Port `8080` and egress traffic to Policy Reporter and the Kyverno Plugin if enabled. Egress rules can be extended as needed.

```yaml
ui:
  networkPolicy:
    enabled: true
    egress: []
```

## Kyverno Plugin

<alert>

Each value is prefixed with `kyvernoPlugin` to be clear that it has to be configured under this subchart.

</alert>

<alert>

Other than Policy Reporter, the __metrics__ as well as the __REST__ API are enabled by default.

</alert>

<alert>

See the complete <a href="https://raw.githubusercontent.com/kyverno/policy-reporter/main/charts/policy-reporter/charts/kyvernoPlugin/values.yaml" target="_blank">values.yaml</a> for reference.

</alert>

Enable the Policy Reporter Kyverno Plugin subchart.

```yaml
kyvernoPlugin:
  enabled: true
```

### Disable Metrics Endpoint

```yaml
kyvernoPlugin:
  metrics:
    enabled: false
```

### Disable REST Endpoint

```yaml
kyvernoPlugin:
  rest:
    enabled: false
```

### Enable NetworkPolicy

If enabled the Helm Chart creates an NetworkPolicy to allow Policy Reporter Kyverno Plugin egress traffic to the Kubernetes API (defaults to port `6443`) and ingress traffic to the REST API from Policy Reporter UI. Ingress- and Egress rules can be extended as needed.

```yaml
kyvernoPlugin:
  networkPolicy:
    enabled: true
    egress:
    - to:
      ports:
      - protocol: TCP
        port: 6443
    ingress: []
```

## Monitoring

<alert>

Each value is prefixed with `monitoring` to be clear that it has to be configured under this subchart.

</alert>
<alert>

See the complete <a href="https://raw.githubusercontent.com/kyverno/policy-reporter/main/charts/policy-reporter/charts/monitoring/values.yaml" target="_blank">values.yaml</a> for reference.

</alert>

The Monitoring Subchart integrates Policy Reporter into the <a href="https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-prometheus-stack" target="_blank">Prometheus Operator</a>.

This Subchart creates a ServiceMonitor for the available `/metrics` endpoint of Policy Reporter to be fetched from Prometheus and 3 preconfigured Dashboards as ConfigMaps for the provided Grafana.

Enable the Policy Reporter Monitoring subchart.

```yaml
monitoring:
  enabled: true
```

### Configure the ServiceMonitor

By default the ServiceMonitor is created in the HelmRelease namespace without additional labels. Depending on your installation you have to add labels to match the `serviceMonitorSelector` of your Prometheus CRD. If you use the Helm Chart of the Prometheus Community as linked above you need a label `release` with the release name of your Prometheus Operator installation. The `serviceMonitor.namespace` configuration is optional and can be changed as needed.

```yaml
monitoring:
  serviceMonitor:
    namespace:
    labels:
      release: prometheus-operator
```

### Enable the Kyverno Plugin integration

<alert>

Requires the `kyvernoPlugin` subchart to be enabled.

</alert>

If enabled a second ServiceMonitor will be created for the `/metrics` endpoint of the Kyerno Plugin. This ServiceMonitor uses also the configuration provided by `monitoring.serviceMonitor`

```yaml
monitoring:
  plugins:
    kyverno: true
```

### Configure Grafana Dashboards

To ensure that the ConfigMaps are identified as Dashboards for Grafana they need a special label `grafana_dashboard` and have to be in the namespace of the Grafana installation.

The required label is preconfigured but can be changed if needed. The namespace of has to be set. It is also possible to disable the ConfigMaps entirely.

```yaml
monitoring:
  grafana:
    # required: namespace of your Grafana installation
    namespace:
    dashboards:
      # Enable the deployment of grafana dashboards
      enabled: true 
      # Label to find dashboards using the k8s sidecar
      label: grafana_dashboard
    # works only if it is supported by your Grafana installation
    folder:
      # Annotation to enable folder storage using the k8s sidecar
      annotation: grafana_folder
      # Grafana folder in which to store the dashboards
      name: Policy Reporter
```

### Configure the Dashboard Views

Depending on your Policy Reports some Result Types like `skipped` or `error` are not supported. Also depending on the size of your cluster and the amount of namespaces some parts are to small to be clearly arranged. To be flexible the Subchart allows you to hide parts of the Dashboard and change the hight of the different components.

```yaml
monitoring:
  # PolicyReport Details Dashboard
  policyReportDetails:
    # high of pass and fail Bar Charts
    firstStatusRow:
      height: 8
    # high of warn and error Bar Charts
    secondStatusRow:
      enabled: true
      height: 2
    # change the high of the Timeline or disable it
    statusTimeline:
      enabled: true
      height: 8
    # change the high of the different status tables or disable them as you like
    passTable:
      enabled: true
      height: 8
    failTable:
      enabled: true
      height: 8
    warningTable:
      enabled: true
      height: 4
    errorTable:
      enabled: true
      height: 4

  # ClusterPolicyReport Details Dashboard
  clusterPolicyReportDetails:
    # high of the status counter boxes
    statusRow:
      height: 6
    # change the high of the Timeline or disable it
    statusTimeline:
      enabled: true
      height: 8
    # change the high of the different status tables or disable them as you like
    passTable:
      enabled: true
      height: 8
    failTable:
      enabled: true
      height: 8
    warningTable:
      enabled: true
      height: 4
    errorTable:
      enabled: true
      height: 4

  # PolicyReports Dashboard
  policyReportOverview:
    # high of the fail bar chart and counter box
    failingSummaryRow:
      height: 8
    # high of the timeline
    failingTimeline:
      height: 10
    # high of the PolicyReport fail table
    failingPolicyRuleTable:
      height: 10
    # high of the ClusterPolicyReport fail table
    failingClusterPolicyRuleTable:
      height: 10
```