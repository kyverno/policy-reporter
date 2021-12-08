---
title: Getting started
description: ''
position: 2
category: Guide
---

Policy Reporter can easily installed with Helm 3 or with the provided static [manifest files](https://github.com/kyverno/policy-reporter/tree/main/manifest). It consists of 4 Parts and can be installed and configured as needed.

## Installation

### Helm Repository

  ```bash
  helm repo add policy-reporter https://kyverno.github.io/policy-reporter
  helm repo update
  ```

### Core Installation

Install only the core application to get REST APIs and/or an metrics endpoint, both are optional and disabled by default.

<code-group>
  <code-block label="Helm 3" active>

  ```bash
  helm upgrade --install policy-reporter policy-reporter/policy-reporter --create-namespace -n policy-reporter --set metrics.enabled=true --set api.enabled=true
  ```

  </code-block>
  <code-block label="Static Manifests">

  ```bash
  kubectl apply -f https://raw.githubusercontent.com/kyverno/policy-reporter/main/manifest/core/install.yaml
  ```
  </code-block>
</code-group>

Access your metrics endpoint on <a href="http://localhost:8080/metrics" target="_blank">http://localhost:8080/metrics</a> via Port Forward:

```bash
kubectl port-forward service/policy-reporter 8080:8080 -n policy-reporter
```

Access your REST API endpoints on <a href="http://localhost:8080/v1/targets" target="_blank">http://localhost:8080/v1/targets</a> via Port Forward:

```bash
kubectl port-forward service/policy-reporter 8080:8080 -n policy-reporter
```

See [API Reference](/core/07-api-reference) for all available endpoints.

### Core + Policy Reporter UI

Install the Policy Reporter core application and the Policy Reporter UI. 
This installation also sets Policy Reporter UI as alert target for new violations.

<code-group>
  <code-block label="Helm 3" active>

  ```bash
  helm upgrade --install policy-reporter policy-reporter/policy-reporter --create-namespace -n policy-reporter --set ui.enabled=true
  ```

  </code-block>
  <code-block label="Static Manifests">

  ```bash
  kubectl apply -f https://raw.githubusercontent.com/kyverno/policy-reporter/main/manifest/default-policy-reporter-ui/namespace.yaml
  kubectl apply -f https://raw.githubusercontent.com/kyverno/policy-reporter/main/manifest/default-policy-reporter-ui/config-secret.yaml
  kubectl apply -f https://raw.githubusercontent.com/kyverno/policy-reporter/main/manifest/default-policy-reporter-ui/install.yaml
  ```
  </code-block>
</code-group>

Access Policy Reporter on <a href="http://localhost:8081" target="_blank">http://localhost:8081</a> via Port Forward:

```bash
kubectl port-forward service/policy-reporter-ui 8081:8080 -n policy-reporter
```

<img src="/images/screenshots/basic-ui-light.png" style="border: 1px solid #ccc" class="light-img" alt="Dashboard light" />
<img src="/images/screenshots/basic-ui-dark.png" style="border: 1px solid #555" class="dark-img" alt="Dashboard dark" />

### Core + Policy Reporter UI + Kyverno Plugin

Install the Policy Reporter core application, Policy Reporter Kyverno Plugin and the Policy Reporter UI with the Kyverno views enabled. 
This installation also sets Policy Reporter UI as alert target for new violations.

<code-group>
  <code-block label="Helm 3" active>

  ```bash
  helm upgrade --install policy-reporter policy-reporter/policy-reporter --create-namespace -n policy-reporter --set kyvernoPlugin.enabled=true --set ui.enabled=true --set ui.plugins.kyverno=true
  ```

  </code-block>
  <code-block label="Static Manifests">

  ```bash
  kubectl apply -f https://raw.githubusercontent.com/kyverno/policy-reporter/main/manifest/kyverno-policy-reporter-ui/namespace.yaml
  kubectl apply -f https://raw.githubusercontent.com/kyverno/policy-reporter/main/manifest/kyverno-policy-reporter-ui/config-secret.yaml
  kubectl apply -f https://raw.githubusercontent.com/kyverno/policy-reporter/main/manifest/kyverno-policy-reporter-ui/install.yaml
  ```
  </code-block>
</code-group>

Access Policy Reporter on <a href="http://localhost:8081" target="_blank">http://localhost:8081</a> via Port Forward:

```bash
kubectl port-forward service/policy-reporter-ui 8081:8080 -n policy-reporter
```

<img src="/images/screenshots/kyverno-dashboard-light.png" style="border: 1px solid #ccc" class="light-img" alt="Kyverno Policy Dashboard light" />
<img src="/images/screenshots/kyverno-dashboard-dark.png" style="border: 1px solid #555" class="dark-img" alt="Kyverno Policy Dashboard dark" />

### Policy Reporter + Prometheus Operator

Install Policy Reporter core application with metrics enabled and the monitoring subchart to install a ServiceMonitor and 3 preconfigured Grafana Dashboards. Change the `monitoring.grafana.namespace` as needed as well `monitoring.serviceMonitor.labels` to match the `serviceMonitorSelector` of your Prometheus CRD.

See <a href="/guide/04-helm-chart-core#configure-the-servicemonitor" target="_blank">Helm Chart - Monitoring</a> for details.

```bash
helm upgrade --install policy-reporter policy-reporter/policy-reporter --set monitoring.enabled=true --set monitoring.grafana.namespace=monitoring --set monitoring.serviceMonitor.labels.release=monitoring -n policy-reporter --create-namespace
```

<img src="/images/screenshots/grafana-policy-reports-dashboard.png" style="border: 1px solid #555" alt="Grafana Policy Reports Dashboard" />
<img src="/images/screenshots/grafana-policy-reports-details.png" style="border: 1px solid #555" alt="Grafana Policy Reports Dashboard" />
<img src="/images/screenshots/grafana-cluster-policy-reports-details.png" style="border: 1px solid #555" alt="Grafana Policy Reports Dashboard" />
