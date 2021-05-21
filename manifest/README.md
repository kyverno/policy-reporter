# Installation Manifests for Policy Reporter

You can use this manifests to install Policy Reporter without additional tools like Helm or Kustomize. The manifests are structured into three installations.

The installation requires a `policy-reporter` namespace. Because the installation includes RBAC resources which requires a serviceAccountName and a namespace configuration. The default namespace is `policy-reporter`. If this namespace will be created if it does not exist.

## Policy Reporter

The `policy-reporter` folder is the basic installation for Policy Reporter without the UI. Includes a basic Configuration Secret `policy-reporter-targets`, empty by default and the `http://policy-reporter:2112/metrics` Endpoint.

### Installation

```bash
kubectl apply -f ./manifest/policy-reporter/install.yaml
```

## Default Policy Reporter UI

The `default-policy-reporter-ui` folder is the extended Policy Reporter and the default Policy Reporter UI installation. 

Enables: 
* Policy Reporter REST API (`http://policy-reporter:8080`) 
* Policy Reporter UI Endpoint (`http://policy-reporter-ui:8080`). 

Configures Policy Reporter UI as Target for Policy Reporter.

### Installation

```bash
kubectl apply -f ./manifest/default-policy-reporter-ui/install.yaml
```

## Kyverno Policy Reporter UI

The `default-policy-reporter-ui` folder is the extended Policy Reporter, Policy Reporter Kyverno Plugin and the extended Policy Reporter UI installation. 

Enables:
* Policy Reporter REST API (`http://policy-reporter:8080`)
* Policy Reporter Metrics API (`http://policy-reporter:2112/metrics`)
* Kyverno Plugin Rest API (`http://policy-reporter-kyverno-plugin:2112/policies`)
* Kyverno Plugin Metrics API (`http://policy-reporter-kyverno-plugin:2113/metrics`) 
* Policy Reporter UI Endpoint (`http://policy-reporter-ui:8080`). 

Configures Policy Reporter UI as Target for Policy Reporter and enables the Kyverno Dashboard.

### Installation

```bash
kubectl apply -f ./manifest/kyverno-policy-reporter-ui/install.yaml
```
