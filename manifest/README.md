# Installation Manifests for Policy Reporter

You can use this manifests to install Policy Reporter without additional tools like Helm or Kustomize. The manifests are structured into five installations.

The installation requires a `policy-reporter` namespace. Because the installation includes RBAC resources which requires a serviceAccountName and a namespace configuration. The default namespace is `policy-reporter`. This namespace will be created if it does not exist.

## Policy Reporter

The `policy-reporter` folder is the basic installation for Policy Reporter without the UI or other components. Includes a basic Configuration Secret `policy-reporter-targets`, empty by default and the `http://policy-reporter:8080/metrics` Endpoint.

### Installation

```bash
kubectl apply -f https://raw.githubusercontent.com/kyverno/policy-reporter/main/manifest/policy-reporter/namespace.yaml
kubectl apply -f https://raw.githubusercontent.com/kyverno/policy-reporter/main/manifest/policy-reporter/install.yaml
```

## Policy Reporter + UI

The `policy-reporter-ui` contains manifests for Policy Reporter and the Policy Reporter UI. 

Enables: 
* Policy Reporter REST API (`http://policy-reporter:8080`) 
* Policy Reporter UI Endpoint (`http://policy-reporter-ui:8080`). 

Configures Policy Reporter UI as Target for Policy Reporter.

### Installation

```bash
kubectl apply -f https://raw.githubusercontent.com/kyverno/policy-reporter/main/manifest/policy-reporter-ui/namespace.yaml
kubectl apply -f https://raw.githubusercontent.com/kyverno/policy-reporter/main/manifest/policy-reporter-ui/target-secret.yaml
kubectl apply -f https://raw.githubusercontent.com/kyverno/policy-reporter/main/manifest/policy-reporter-ui/install.yaml
```

## Policy Reporter + KyvernoPlugin + UI

The `policy-reporter-kyverno-ui` contains manifests for Policy Reporter, Policy Reporter Kyverno Plugin and Policy Reporter UI. 

Enables:
* Policy Reporter REST API (`http://policy-reporter:8080`)
* Policy Reporter Metrics API (`http://policy-reporter:8080/metrics`)
* Kyverno Plugin Rest API (`http://policy-reporter-kyverno-plugin:8080/policies`)
* Kyverno Plugin Metrics API (`http://policy-reporter-kyverno-plugin:8080/metrics`) 
* Policy Reporter UI Endpoint (`http://policy-reporter-ui:8080`). 

Configures Policy Reporter UI as Target for Policy Reporter and enables the Kyverno Dashboard.

### Installation

```bash
kubectl apply -f https://raw.githubusercontent.com/kyverno/policy-reporter/main/manifest/policy-reporter-kyverno-ui/namespace.yaml
kubectl apply -f https://raw.githubusercontent.com/kyverno/policy-reporter/main/manifest/policy-reporter-kyverno-ui/target-secret.yaml
kubectl apply -f https://raw.githubusercontent.com/kyverno/policy-reporter/main/manifest/policy-reporter-kyverno-ui/install.yaml
```

## High Available Policy Reporter + KyvernoPlugin + UI

The `policy-reporter-kyverno-ui-ha` contains a high available setup for Policy Reporter, Policy Reporter Kyverno Plugin and Policy Reporter UI, it enabled leaderelection and uses redis as a external and central storage for shared caches and Logs (UI)

Enables:
* Policy Reporter REST API (`http://policy-reporter:8080`)
* Policy Reporter Metrics API (`http://policy-reporter:8080/metrics`)
* Kyverno Plugin Rest API (`http://policy-reporter-kyverno-plugin:8080/policies`)
* Kyverno Plugin Metrics API (`http://policy-reporter-kyverno-plugin:8080/metrics`) 
* Kyverno Plugin PolicyReport creation for blocked resources (by __Kyverno__ enforce policies)
* Policy Reporter UI Endpoint (`http://policy-reporter-ui:8080`).

Additional resources:
* `PodDisruptionBudget` for each component
* `Role` and `RoleBinding` for Policy Reporter and the KyvernoPlugin to manage Lease resources for leaderelection
* Basic `Redis`, used as central and external cache for Policy Reporter and as central Log storage for Policy Reporter UI

### Installation

```bash
kubectl apply -f https://raw.githubusercontent.com/kyverno/policy-reporter/main/manifest/policy-reporter-kyverno-ui-ha/namespace.yaml
kubectl apply -f https://raw.githubusercontent.com/kyverno/policy-reporter/main/manifest/policy-reporter-kyverno-ui-ha/config-core.yaml
kubectl apply -f https://raw.githubusercontent.com/kyverno/policy-reporter/main/manifest/policy-reporter-kyverno-ui-ha/config-ui.yaml
kubectl apply -f https://raw.githubusercontent.com/kyverno/policy-reporter/main/manifest/policy-reporter-kyverno-ui-ha/config-kyverno-plugin.yaml
kubectl apply -f https://raw.githubusercontent.com/kyverno/policy-reporter/main/manifest/policy-reporter-kyverno-ui-ha/redis.yaml
kubectl apply -f https://raw.githubusercontent.com/kyverno/policy-reporter/main/manifest/policy-reporter-kyverno-ui-ha/install.yaml
```

See `complete-ha/README.md` for details about the used configuration values.

## Policy Reporter Configuration

To configure policy-reporter, for example your notification targets, create a secret called `policy-reporter-targets` in the `policy-reporter` namespace with an key `config.yaml` as key and the following structure as value:

```yaml
priorityMap: {}

loki:
  host: ""
  minimumPriority: ""
  skipExistingOnStartup: true
  customLabels: {}
  sources: []
  channels: []

elasticsearch:
  host: ""
  index: "policy-reporter"
  rotation: "dayli"
  minimumPriority: ""
  skipExistingOnStartup: true
  sources: []
  channels: []

slack:
  webhook: ""
  minimumPriority: ""
  skipExistingOnStartup: true
  sources: []
  channels: []

discord:
  webhook: ""
  minimumPriority: ""
  skipExistingOnStartup: true
  sources: []
  channels: []

teams:
  webhook: ""
  minimumPriority: ""
  skipExistingOnStartup: true
  sources: []
  channels: []

ui:
  host: ""
  minimumPriority: ""
  skipExistingOnStartup: true
  sources: []

webhook:
  host: ""
  headers: {}
  minimumPriority: ""
  skipExistingOnStartup: true
  sources: []
  channels: []

s3:
  endpoint: ""
  region: ""
  bucket: ""
  secretAccessKey: ""
  accessKeyID: ""
  minimumPriority: "warning"
  skipExistingOnStartup: true
  sources: []
  channels: []

reportFilter:
  namespaces:
    include: []
    exclucde: []
  clusterReports:
    disabled: false

# optional external result caching
redis:
  enabled: false
  address: ""
  database: 0
  prefix: "policy-reporter"
  username: ""
  password: ""

leaderElection:
  enabled: false
  releaseOnCancel: true
  leaseDuration: 15
  renewDeadline: 10
  retryPeriod: 2
```

The `kyverno-policy-reporter-ui` and `default-policy-reporter-ui` installation has an optional preconfigured `target-security.yaml` to apply. This secret configures the Policy Reporter UI as target for Policy Reporter.

When you change the secret while Policy Reporter is already running, you have to delete the current `policy-reporter` Pod.

## Policy Reporter Summary Email Report

The `violations-email-report` folder can be used to install Policy Reporter only for the matter of sending E-Mail Summary Reports. You can install the Email Summary Report without the requirement of the Policy Reporter core application. If you already have Policy Reporter installed, you can just apply `config-secret.yaml` and `cronjob.yaml` to add the email report feature. It will reuse the existing `ServiceAccount` and `Namespace`.

### Installation

```bash
kubectl apply -f https://raw.githubusercontent.com/kyverno/policy-reporter/main/manifest/violations-email-report/namespace.yaml
kubectl apply -f https://raw.githubusercontent.com/kyverno/policy-reporter/main/manifest/violations-email-report/config-secret.yaml
kubectl apply -f https://raw.githubusercontent.com/kyverno/policy-reporter/main/manifest/violations-email-report/serviceaccount.yaml
kubectl apply -f https://raw.githubusercontent.com/kyverno/policy-reporter/main/manifest/violations-email-report/cronjob.yaml
```

### Configuration

To configure your SMTP server and receiver emails use the following configuration template and replace the `config.yaml` value of `config-secret.yaml` with your base64 encoded configuration.

```yaml
emailReports:
  clusterName: '' # optional clustername shown in the Report
  smtp:
    host: ''
    port: 465
    username: ''
    password: ''
    from: '' # from E-Mail address
    encryption: '' # default is none, supports ssl/tls and starttls
  violations:
    to: []
    filter:
      disableClusterReports: false # remove ClusterPolicyResults from Reports
      namespaces:
        include: []
        exclude: []
      sources:
        include: []
        exclude: []
    channels: []
```
