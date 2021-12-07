---
title: Local Development
description: ''
position: 15
category: 'Kyverno Plugin'
---

## Requirements

* Go >= v1.17
* (optional) Kubernetes Cluster with <a href="https://kyverno.io">Kyverno</a> installed

## Getting started

Fork and/or checkout <a href="https://github.com/kyverno/policy-reporter-kyverno-plugin" target="_blank">Policy Reporter Kyverno Plugin on GitHub</a>

### Install dependencies

```bash
go get ./...
```

### Unit Tests

Run Unit Tests locally with

```bash
make test
```

or 

```bash
go test -v ./... -timeout=10s
```

## Running Kyverno Plugin

To run the Kyverno plugin locally, you must connect it to a Kubernetes cluster. This is required because it connects to the Kubernetes API to watch for Kyverno policies and ClusterPolicies. The configuration can be done via CLI arguments, <a href="/kyverno-plugin/16-config-reference" target="_blank">config.yaml</a> or a mix of both.

```bash
go run main.go run -k ~/.kube/config
```

### Argument Referece

| Argument            | Short   | Discription                                                           |Default              |
|---------------------|:-------:|-----------------------------------------------------------------------|---------------------|
| `--kubeconfig`      | `-k`    | path to the kubeconfig file,<br>used to connect to the Kubernetes API |                     |
| `--config`          | `-c`    | path to the Policy Reporter config file                               |`config.yaml`        |
| `--metrics-enabled` | `-m`    | enables the Metrics API Endpoints                                     |`false`              |
| `--rest-enabled`    | `-r`    | enables the REST API Endpoints                                        |`false`              |
| `--port`            | `-p`    | used port for the HTTP Server                                         |`8080`               |

### Compile and run Policy Reporter

```bash
make build

./build/policyreporter run -k ~/.kube/config
```