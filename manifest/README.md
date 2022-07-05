# Installation Manifests for Policy Reporter

You can use this manifests to install Policy Reporter without additional tools like Helm or Kustomize. The manifests are structured into three installations.

The installation requires a `policy-reporter` namespace. Because the installation includes RBAC resources which requires a serviceAccountName and a namespace configuration. The default namespace is `policy-reporter`. If this namespace will be created if it does not exist.

## Policy Reporter

The `policy-reporter` folder is the basic installation for Policy Reporter without the UI. Includes a basic Configuration Secret `policy-reporter-targets`, empty by default and the `http://policy-reporter:8080/metrics` Endpoint.

### Installation

```bash
kubectl apply -f https://raw.githubusercontent.com/kyverno/policy-reporter/main/manifest/policy-reporter/install.yaml
```

## Default Policy Reporter UI

The `default-policy-reporter-ui` folder is the extended Policy Reporter and the default Policy Reporter UI installation. 

Enables: 
* Policy Reporter REST API (`http://policy-reporter:8080`) 
* Policy Reporter UI Endpoint (`http://policy-reporter-ui:8080`). 

Configures Policy Reporter UI as Target for Policy Reporter.

### Installation

```bash
kubectl apply -f https://raw.githubusercontent.com/kyverno/policy-reporter/main/manifest/default-policy-reporter-ui/namespace.yaml
kubectl apply -f https://raw.githubusercontent.com/kyverno/policy-reporter/main/manifest/default-policy-reporter-ui/target-secret.yaml
kubectl apply -f https://raw.githubusercontent.com/kyverno/policy-reporter/main/manifest/default-policy-reporter-ui/install.yaml
```

## Kyverno Policy Reporter UI

The `kyverno-policy-reporter-ui` folder is the extended Policy Reporter, Policy Reporter Kyverno Plugin and the extended Policy Reporter UI installation. 

Enables:
* Policy Reporter REST API (`http://policy-reporter:8080`)
* Policy Reporter Metrics API (`http://policy-reporter:8080/metrics`)
* Kyverno Plugin Rest API (`http://policy-reporter-kyverno-plugin:8080/policies`)
* Kyverno Plugin Metrics API (`http://policy-reporter-kyverno-plugin:8080/metrics`) 
* Policy Reporter UI Endpoint (`http://policy-reporter-ui:8080`). 

Configures Policy Reporter UI as Target for Policy Reporter and enables the Kyverno Dashboard.

### Installation

```bash
kubectl apply -f https://raw.githubusercontent.com/kyverno/policy-reporter/main/manifest/kyverno-policy-reporter-ui/namespace.yaml
kubectl apply -f https://raw.githubusercontent.com/kyverno/policy-reporter/main/manifest/kyverno-policy-reporter-ui/target-secret.yaml
kubectl apply -f https://raw.githubusercontent.com/kyverno/policy-reporter/main/manifest/kyverno-policy-reporter-ui/install.yaml
```

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
