---
title: Local Development
description: ''
position: 12
category: 'Policy Reporter UI'
---

## Go Backend

The Go Backend acts as:
* Backend Store and API for the Policy Report pushes
* FileServer for the NuxtJS Single Plage Application (the actual UI)
* HTTP Proxy for the Policy Reporter REST API
* HTTP Proxy for the Kyverno Plugin REST API (if enabled)

### Requirements

* Go >= v1.17

### Getting started

Fork and/or checkout <a href="https://github.com/kyverno/policy-reporter-ui" target="_blank">Policy Reporter UI on GitHub</a>

The Go Backend is located in the `./server` directory

### Install dependencies

```bash
cd server

go get ./...
```

## Running Policy Reporter UI

```bash
go run main.go -no-ui -dev -port=8082
```

### Argument Referece

| Argument            | Discription                                                                                  |Default       |
|---------------------|----------------------------------------------------------------------------------------------|------------- |
| `-config`           | path to the Policy Reporter UI config file                                                   |`config.yaml` |
| `-dev`              | add the __Access-Control-Allow-Origin__ HTTP Header<br>to all APIs to avaid CORS errors      |`false`       |
| `-no-ui`            | disable the SPA Handler to start the backend without the UI,<br>only for development purpose |`false`       |
| `-policy-reporter`  | Host URL to Policy Reporter,<br>used to proxy API requests to from the UI                    |              |
| `-kyverno-plugin`   | Host URL to Policy Reporter Kyverno Plugin,<br>used to proxy API requests to from the UI     |              |
| `-port`             | used port for the HTTP Server                                                                |`8080`        |

### Compile and run Policy Reporter UI

```bash
make build

./build/policyreporter-ui -no-ui -dev -port=8082
```

## NuxtJS Frontend

The actual frontend is a single page application based on <a href="https://nuxtjs.org/" target="_blank">NuxtJS</a> and <a href="https://www.typescriptlang.org/" target="_blank">TypeScript</a>.

### Requirements

* NodeJS >= v16
* Local running Policy Reporter UI Backend
* Accessable Policy Reporter REST API
* Accessable Kyverno Plugin REST API (optional)

### Preparation

Access Policy Reporter via Port Forward: 

```bash
kubectl port-forward service/policy-reporter 8080:8080 -n policy-reporter
```

Access Policy Reporter Kyverno Plugin via Port Forward: 

```bash
kubectl port-forward service/policy-reporter-kyverno-plugin 8083:8080 -n policy-reporter
```

Start the Policy Reporter UI Server in development mode without the UI. The server has to be started in the `server` directory of the Policy Reporter UI project.

```bash
go run main.go -no-ui -dev -port=8082 -policy-reporter http://localhost:8080 -kyverno-plugin http://localhost:8083
```

### Install Dependencies

Dependencies are managed with NPM.

```bash
npm install
```

### Running Policy Reporter UI

Create a .env File to configure the Policy Reporter UI - Backend URL. With this setup you can just copy the prepared `.env.example`.

```bash
cp .env.example .env
```

Start the NuxtJS development server

```bash
npm run dev
```

Open <a href="http://localhost:3000" target="_blank">http://localhost:3000</a>.

Check the output of the `npm run dev` command if this port is not working.