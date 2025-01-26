# Setup Kyverno and Policy Reporter

## Install Kyverno + Kyverno PSS Policies

Add Helm Repo

```bash
helm repo add kyverno https://kyverno.github.io/kyverno/
helm repo update
```

Install Kyverno + PSS Policies

```bash
helm upgrade --install kyverno kyverno/kyverno -n kyverno --create-namespace
helm upgrade --install kyverno-policies kyverno/kyverno-policies -n kyverno --create-namespace --set podSecurityStandard=restricted
```

## Installing Policy Reporter v3 Preview and Policy Reporter UI v2 + Kyverno Plugin

```bash
helm repo add policy-reporter https://kyverno.github.io/policy-reporter
helm repo update
```

Install the Policy Reporter Preview

```bash
helm upgrade --install policy-reporter policy-reporter/policy-reporter --create-namespace -n policy-reporter --set ui.enabled=true --set plugin.kyverno.enabled=true
```