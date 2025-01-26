# Policy Reporter 3.x
[![CI](https://github.com/kyverno/policy-reporter/actions/workflows/ci.yaml/badge.svg)](https://github.com/kyverno/policy-reporter/actions/workflows/ci.yaml) [![Go Report Card](https://goreportcard.com/badge/github.com/kyverno/policy-reporter)](https://goreportcard.com/report/github.com/kyverno/policy-reporter) [![Coverage Status](https://coveralls.io/repos/github/kyverno/policy-reporter/badge.svg?branch=main)](https://coveralls.io/github/kyverno/policy-reporter?branch=main)


![Screenshot Policy Reporter UI v2](https://github.com/kyverno/policy-reporter/blob/main/docs/images/screen.png)

## Documentation

The documentation for Policy Reporter v3 can be found here: [https://kyverno.github.io/policy-reporter-docs/](https://kyverno.github.io/policy-reporter-docs/)

## Getting Started

## Installation with Helm v3

Installation via Helm Repository

### Add the Helm repository
```bash
helm repo add policy-reporter https://kyverno.github.io/policy-reporter
helm repo update
```

### Installation with Policy Reporter UI and Kyverno Plugin enabled
```bash
helm install policy-reporter policy-reporter/policy-reporter --create-namespace -n policy-reporter --set ui.enabled=true --set kyverno-plugin.enabled=true
kubectl port-forward service/policy-reporter-ui 8082:8080 -n policy-reporter
```

Open `http://localhost:8082/` in your browser.