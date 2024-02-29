# Policy Reporter 3.x Preview
[![CI](https://github.com/kyverno/policy-reporter/actions/workflows/ci.yaml/badge.svg)](https://github.com/kyverno/policy-reporter/actions/workflows/ci.yaml) [![Go Report Card](https://goreportcard.com/badge/github.com/kyverno/policy-reporter)](https://goreportcard.com/report/github.com/kyverno/policy-reporter) [![Coverage Status](https://coveralls.io/repos/github/kyverno/policy-reporter/badge.svg?branch=main)](https://coveralls.io/github/kyverno/policy-reporter?branch=main)

## Preview Feature Docs

Documentation for upcoming features and changes for the new Policy Reporter UI v2 are located in [Docs](https://github.com/kyverno/policy-reporter/tree/3.x/docs)

* [Installation](https://github.com/kyverno/policy-reporter/blob/3.x/docs/SETUP.md)
* [OAUth2 / OpenIDConnect](https://github.com/kyverno/policy-reporter/blob/3.x/docs/UI_AUTH.md)
* [UI CustomBoards](https://github.com/kyverno/policy-reporter/blob/3.x/docs/CUSTOM_BOARDS.md)

## Getting Started

## Installation with Helm v3

Installation via Helm Repository

### Add the Helm repository
```bash
helm repo add policy-reporter https://kyverno.github.io/policy-reporter
helm repo update
```

### Installation

### Installation with Policy Reporter UI and Kyverno Plugin enabled
```bash
helm install policy-reporter policy-reporter/policy-reporter-preview --create-namespace -n policy-reporter --devel --set ui.enabled=true --set kyverno-plugin.enabled=true
kubectl port-forward service/policy-reporter-ui 8082:8080 -n policy-reporter
```

Open `http://localhost:8082/` in your browser.