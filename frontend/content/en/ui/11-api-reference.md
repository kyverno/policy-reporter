---
title: API Reference
description: ''
position: 11
category: 'Policy Reporter UI'
---

Reference of all available HTTP Endpoints provided by Policy Reporter UI

## Core APIs

### Push API

| Method | API          | Description                                                   | Codes |
|--------|--------------|---------------------------------------------------------------|----------------|
| `POST`  | `/api/push` | Receive a single PolicyReport Result and stores it in memory  | `200`, `500`   |

#### Example

```bash
curl -X POST -H "Content-type: application/json" -d '{"Message":"validation error: Running as root is not allowed. The fields spec.securityContext.runAsNonRoot, spec.containers[*].securityContext.runAsNonRoot, and spec.initContainers[*].securityContext.runAsNonRoot must be `true`. Rule check-containers[0] failed at path /spec/securityContext/runAsNonRoot/. Rule check-containers[1] failed at path /spec/containers/0/securityContext/.","Policy":"require-run-as-non-root","Rule":"check-containers","Priority":"warning","Status":"fail","Category":"Pod Security Standards (Restricted)","Source":"Kyverno","Scored":true,"CreationTimestamp":"2021-12-04T10:13:02Z","Resource":{"APIVersion":"v1","Kind":"Pod","Name":"nginx2","Namespace":"test","UID":"ac4d11f3-0aa8-43f0-8056-98f4eae0d956"}}' "http://localhost:8080/api/push"
```

Body JSON

```json
{
   "Message":"validation error: Running as root is not allowed. The fields spec.securityContext.runAsNonRoot, spec.containers[*].securityContext.runAsNonRoot, and spec.initContainers[*].securityContext.runAsNonRoot must be `true`. Rule check-containers[0] failed at path /spec/securityContext/runAsNonRoot/. Rule check-containers[1] failed at path /spec/containers/0/securityContext/.",
   "Policy":"require-run-as-non-root",
   "Rule":"check-containers",
   "Priority":"warning",
   "Status":"fail",
   "Category":"Pod Security Standards (Restricted)",
   "Source":"Kyverno",
   "Scored":true,
   "CreationTimestamp":"2021-12-04T10:13:02Z",
   "Resource":{
      "APIVersion":"v1",
      "Kind":"Pod",
      "Name":"nginx2",
      "Namespace":"test",
      "UID":"ac4d11f3-0aa8-43f0-8056-98f4eae0d956"
   }
}
```

* Response `200`

```json
{}
```

* Response `500` 

```json
{ "message": "Error Message" }
```

### Results API

| Method | API               | Description                           | Codes |
|--------|-------------------|---------------------------------------|----------------|
| `GET`  | `/api/result-log` | Returns the logs of received Results  | `200`, `500`   |

#### Example

```bash
curl -X GET "http://localhost:8080/api/result-log"
```

* Response `200`

```json
[
   {
      "message":"validation error: Running as root is not allowed. The fields spec.securityContext.runAsNonRoot, spec.containers[*].securityContext.runAsNonRoot, and spec.initContainers[*].securityContext.runAsNonRoot must be `true`. Rule check-containers[0] failed at path /spec/securityContext/runAsNonRoot/. Rule check-containers[1] failed at path /spec/containers/0/securityContext/.",
      "policy":"require-run-as-non-root",
      "rule":"check-containers",
      "priority":"warning",
      "status":"fail",
      "category":"Pod Security Standards (Restricted)",
      "scored":true,
      "resource":{
         "apiVersion":"v1",
         "kind":"Pod",
         "name":"nginx2",
         "namespace":"test",
         "uid":"ac4d11f3-0aa8-43f0-8056-98f4eae0d956"
      },
      "creationTimestamp":"2021-12-04T10:13:02Z"
   }
]
```

* Response `500` 

```json
{ "message": "Error Message" }
```

### Config API

| Method | API           | Description                                            | Codes |
|--------|---------------|--------------------------------------------------------|----------------|
| `GET`  | `/api/config` | Returns configured Plugins and the default DisplayMode | `200`, `500`   |

#### Example

```bash
curl -X GET "http://localhost:8080/api/config"
```

* Response `200`

```json
{
   "displayMode":"",
   "plugins":[
      "kyverno"
   ]
}
```

* Response `500` 

```json
{ "message": "Error Message" }
```

## Proxy APIs

### Policy Reporter

| Method | API         | Description                                       |
|--------|-------------|---------------------------------------------------|
| `GET`  | `/api/v1/*` | Proxy to the configured Policy Reporter Host URL  |

See [Policy Reporter - API Reference](/core/07-api-reference#v1-general-apis) for all available endpoints.

### Kyverno Plugin

| Method | API         | Description                                       |
|--------|-------------|---------------------------------------------------|
| `GET`  | `/api/kyverno/*` | Proxy to the configured Policy Reporter Kyverno Plugin Host URL  |

See [Kyverno Plugin - API Reference](/kyverno-plugin/14-api-reference#v1-general-apis) for all available endpoints.
