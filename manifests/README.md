# Installation Manifests for Policy Reporter

You can use this manifests to install Policy Reporter without additional tools like Helm or Kustomize. The manifests are structured into five installations.

The installation requires to be in the `policy-reporter` namespace. As its the configured namespaces for RBAC resources.

## Policy Reporter

The `policy-reporter` folder is a basic installation for Policy Reporter without the UI or other components. It runs with the REST API and Metrics Endpoint enabled.

```bash
kubectl apply -f https://raw.githubusercontent.com/kyverno/policy-reporter/main/manifests/policy-reporter/install.yaml
```

## Policy Reporter UI

The `policy-reporter-ui` folder installs Policy Reporter together with the Policy Reporter UI components and Metrics enabled.

```bash
kubectl apply -f https://raw.githubusercontent.com/kyverno/policy-reporter/main/manifests/policy-reporter-ui/install.yaml
```

## Policy Reporter UI + Kyverno Plugin

The `policy-reporter-kyverno-ui` folder installs Policy Reporter together with the Policy Reporter UI, Kyverno Plugin components and Metrics enabled.

```bash
kubectl apply -f https://raw.githubusercontent.com/kyverno/policy-reporter/main/manifests/policy-reporter-kyverno-ui/install.yaml
```

## Policy Reporter UI + Kyverno Plugin in HA Mode

The `policy-reporter-kyverno-ui-ha` installs the same compoments as `policy-reporter-kyverno-ui` but runs all components in HA mode (2 replicas) and creates additional resources for leader elections.

```bash
kubectl apply -f https://raw.githubusercontent.com/kyverno/policy-reporter/main/manifests/policy-reporter-kyverno-ui-ha/install.yaml
```
