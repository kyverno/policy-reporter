# Demo Instructions

## Kind Cluster

```bash
make kind-create-cluster
```

## Kyverno

### Add Repository

```bash
helm repo add kyverno https://kyverno.github.io/kyverno
```

### Install

```bash
helm upgrade --install kyverno kyverno/kyverno -n kyverno --create-namespace
helm upgrade --install kyverno-policies kyverno/kyverno-policies --set podSecurityStandard=restricted
```

## Falco

### Add Repository

```bash
helm repo add falcosecurity https://falcosecurity.github.io/charts
```

### Install

```bash
helm upgrade --install falco falcosecurity/falco --set falcosidekick.enabled=true --set falcosidekick.config.policyreport.enabled=true --set falcosidekick.image.tag=latest  --namespace falco --create-namespace
```

## Trivy Operator

### Add Repository

```bash
helm repo add aqua https://aquasecurity.github.io/helm-charts/
helm repo add trivy-operator-polr-adapter https://fjogeleit.github.io/trivy-operator-polr-adapter
```

### Install

```bash
helm upgrade --install trivy-operator aqua/trivy-operator -n trivy-system --create-namespace --set="trivy.ignoreUnfixed=true"
helm upgrade --install trivy-operator-polr-adapter trivy-operator-polr-adapter/trivy-operator-polr-adapter -n trivy-system
```

## Policy Reporter

### Add Repository

```bash
helm repo add policy-reporter https://kyverno.github.io/policy-reporter
```

### Install

#### Slack Secret

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: webhook-secret
  namespace: policy-reporter
type: Opaque
data:
  webhook: aHR0cHM6Ly9ob29rcy5z...
```

#### Values

```yaml
plugin:
  kyverno:
    enabled: true

  trivy:
    enabled: true

ui:
  enabled: true

  ingress:
    enabled: true
    annotations:
      nginx.ingress.kubernetes.io/rewrite-target: /$1
    className: nginx
    hosts:
      - host: localhost
        paths:
        - path: "/ui/(.*)"
          pathType: ImplementationSpecific

  sources:
    - name: Trivy ConfigAudit
      type: severity
      excludes:
        results:
        - pass
        - error

    - name: Trivy Vulnerability
      type: severity
      excludes:
        results:
        - pass
        - error

    - name: Falco
      excludes:
        results:
        - pass
        - skip

target:
  slack:
    name: Kyverno Channel
    channel: kyverno
    secretRef: webhook-secret
    minimumPriority: warning
    skipExistingOnStartup: true
    sources: [kyverno]
    filter:
      namespaces:
        exclude: ['trivy-system']
    channels:
      - name: Trivy Operator
        channel: trivy-operator
        sources: [Trivy Vulnerability]
        filter:
          namespaces:
            exclude: ['trivy-system']
```

```bash
helm upgrade --install policy-reporter policy-reporter/policy-reporter-preview --create-namespace -n policy-reporter -f values.yaml --devel
```
