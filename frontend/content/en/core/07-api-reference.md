---
title: API Reference
description: ''
position: 7
category: 'Policy Reporter'
---

Reference of all available HTTP Endpoints provided by Policy Reporter

## Core APIs

### Healthz API

| Method | API        | Description                                                    | Codes |
|--------|------------|----------------------------------------------------------------|----------------|
| `GET`  | `/healthz` | Returns if the App is healthy and required CRDs are installed  | `200`, `503`   |

#### Example

```bash
curl -X GET "http://localhost:8080/healthz"
```

* Response `200`

```json
{}
```

* Response `503` 

```json
{ "error": "No PolicyReport CRDs found" }
```

### Readiness API

| Method | API      | Description                           | Codes |
|--------|----------|---------------------------------------|----------------|
| `GET`  | `/ready` | Returns if the App is up and running  | `200`          |

#### Example

```bash
curl -X GET "http://localhost:8080/ready"
```

* Response `200`

```json
{}
```

## V1 General APIs

### Targets API

| Method | API          | Description                 | Codes |
|--------|--------------|-----------------------------|----------------|
| `GET`  | `/v1/targets` | List of configured targets  | `200`, `500`   |

#### Example

```bash
curl -X GET "http://localhost:8080/v1/targets"
```

* Response `200`

```json
[
   {
      "name":"UI",
      "minimumPriority":"warning",
      "sources":[
         "Kube Bench",
         "Kyverno"
      ],
      "skipExistingOnStartup":true
   },
   {
      "name":"S3",
      "minimumPriority":"warning",
      "skipExistingOnStartup":true
   }
]
```

* Response `500`

```json
{ "message": "Error Message" }
```

### Categories API

| Method | API              | Description                                                          | Codes |
|--------|------------------|----------------------------------------------------------------------|----------------|
| `GET`  | `/v1/categories` | List of all defined PolicyReport and ClusterPolicyReport Categories  | `200`, `500`   |

#### Query Filter Parameters

| Filter      | Type          | Description                 |
|-------------|---------------|-----------------------------|
| __source__  | `string[]`    | Filter by a list of sources |

#### Example

```bash
curl -X GET "http://localhost:8080/v1/categories?source=kyverno"
```

* Response `200`

```json
[
  "Pod Security Standards (Default)",
  "Pod Security Standards (Restricted)"
]
```

* Response `500`

```json
{ "message": "Error Message" }
```

### Namespaces API

| Method | API              | Description                                      | Codes |
|--------|------------------|--------------------------------------------------|----------------|
| `GET`  | `/v1/namespaces` | List of all Namespaces with PolicyReportResults  | `200`, `500`   |

#### Query Filter Parameters

| Filter      | Type          | Description                 |
|-------------|---------------|-----------------------------|
| __source__  | `string[]`    | Filter by a list of sources |

#### Example

```bash
curl -X GET "http://localhost:8080/v1/namespaces?source=kyverno&sorce=falco"
```

* Response `200`

```json
[
  "policy-reporter",
  "blog",
  "test"
]
```

* Response `500`

```json
{ "message": "Error Message" }
```

### Rule Status Count API

| Method | API                    | Description                                                | Codes |
|--------|------------------------|------------------------------------------------------------|----------------|
| `GET`  | `/v1/rule-status-count`| List of counts per result<br>of the selected policy and rule  | `200`, `500`   |

#### Query Filter Parameters

| Filter      | Type          | Description                                  | Required   |
|-------------|---------------|----------------------------------------------|------------|
| __rule__    | `string`      | Select the __Rule__ for the requested counts |__required__|
| __policy__  | `string`      | Select the __Policy__ of selected __Rule__   |__required__|

#### Example

```bash
curl -X GET "http://localhost:8080/v1/rule-status-count?policy=require-non-root-groups&rule=autogen-check-fsGroup"
```

* Response `200`

```json
[
   {
      "status":"pass",
      "count":25
   },
   {
      "status":"fail",
      "count":0
   },
   {
      "status":"warn",
      "count":0
   },
   {
      "status":"error",
      "count":0
   },
   {
      "status":"skip",
      "count":0
   }
]
```

* Response `500`

```json
{ "message": "Error Message" }
```

## V1 PolicyReport APIs

### Policies API

| Method | API                                 | Description                                           | Codes |
|--------|-------------------------------------|-------------------------------------------------------|----------------|
| `GET`  | `/v1/namespaced-resources/policies` | List of all Policies<br>with namespace scoped results | `200`, `500`   |

#### Query Filter Parameters

| Filter      | Type       | Description      |
|-------------|------------|------------------|
| __source__  | `string`   | Filter by source |

#### Example

```bash
curl -X GET "http://localhost:8080/v1/namespaced-resources/policies?source=kyverno"
```

* Response `200`

```json
[
   "deny-privilege-escalation",
   "disallow-add-capabilities",
   "disallow-host-namespaces",
   "disallow-host-path",
   "disallow-host-ports",
   "disallow-privileged-containers",
   "disallow-selinux",
   "require-default-proc-mount",
   "require-non-root-groups",
   "require-run-as-non-root",
   "restrict-apparmor-profiles",
   "restrict-seccomp",
   "restrict-sysctls",
   "restrict-volume-types"
]
```

* Response `500`

```json
{ "message": "Error Message" }
```

### Kinds API

| Method | API                                 | Description                                        | Codes |
|--------|-------------------------------------|----------------------------------------------------|----------------|
| `GET`  | `/v1/namespaced-resources/kinds`    | List of all Kinds<br>with namespace scoped results | `200`, `500`   |

#### Query Filter Parameters

| Filter      | Type       | Description      |
|-------------|------------|------------------|
| __source__  | `string`   | Filter by source |

#### Example

```bash
curl -X GET "http://localhost:8080/v1/namespaced-resources/kinds?source=kyverno"
```

* Response `200`

```json
[
   "CronJob",
   "Deployment",
   "Pod",
   "StatefulSet"
]
```

* Response `500`

```json
{ "message": "Error Message" }
```

### Sources API

| Method | API                                 | Description                                          | Codes |
|--------|-------------------------------------|------------------------------------------------------|----------------|
| `GET`  | `/v1/namespaced-resources/sources`  | List of all Sources<br>with namespace scoped results | `200`, `500`   |

#### Example

```bash
curl -X GET "http://localhost:8080/v1/namespaced-resources/sources"
```

* Response `200`

```json
[
   "Kyverno",
   "Kube Benchh"
]
```

* Response `500`

```json
{ "message": "Error Message" }
```

### Status Counts API

| Method | API                                         | Description                                        | Codes |
|--------|---------------------------------------------|----------------------------------------------------|----------------|
| `GET`  | `/v1/namespaced-resources/status-counts`    | Count of result status<br>per status and namespace | `200`, `500`   |

#### Query Filter Parameters

| Filter          | Type          | Description                    | Enum                                    |
|-----------------|---------------|--------------------------------|-----------------------------------------|
| __source__      | `string[]`    | Filter by a list of sources    |                                         |
| __namespaces__  | `string[]`    | Filter by a list of namespaces |                                         |
| __kinds__       | `string[]`    | Filter by a list of kinds      |                                         |
| __categories__  | `string[]`    | Filter by a list of categories |                                         |
| __policies__    | `string[]`    | Filter by a list of policies   |                                         |
| __status__      | `string[]`    | Filter by a list of status     | `fail`, `pass`, `warn`, `error`, `skip` |
| __severities__  | `string[]`    | Filter by a list of severities | `low`, `medium`, `high`                 |


#### Example

```bash
curl -X GET "http://localhost:8080/v1/namespaced-resources/status-counts?source=kyverno&status=pass&status=fail"
```

* Response `200`

```json
[
   {
      "status":"pass",
      "items":[
         {
            "namespace":"argo-cd",
            "count":206
         },
         {
            "namespace":"blog",
            "count":34
         },
         {
            "namespace":"policy-reporter",
            "count":105
         },
         {
            "namespace":"test",
            "count":34
         }
      ]
   },
   {
      "status":"fail",
      "items":[
         {
            "namespace":"argo-cd",
            "count":4
         },
         {
            "namespace":"blog",
            "count":1
         },
         {
            "namespace":"test",
            "count":1
         }
      ]
   }
]
```

* Response `500`

```json
{ "message": "Error Message" }
```

### Results API

| Method | API                                   | Description                      | Codes |
|--------|---------------------------------------|----------------------------------|----------------|
| `GET`  | `/v1/namespaced-resources/results`    | List of namespace scoped results | `200`, `500`   |

#### Query Filter Parameters

| Filter          | Type          | Description                    | Enum                                    |
|-----------------|---------------|--------------------------------|-----------------------------------------|
| __source__      | `string[]`    | Filter by a list of sources    |                                         |
| __namespaces__  | `string[]`    | Filter by a list of namespaces |                                         |
| __kinds__       | `string[]`    | Filter by a list of kinds      |                                         |
| __categories__  | `string[]`    | Filter by a list of categories |                                         |
| __policies__    | `string[]`    | Filter by a list of policies   |                                         |
| __status__      | `string[]`    | Filter by a list of status     | `fail`, `pass`, `warn`, `error`, `skip` |
| __severities__  | `string[]`    | Filter by a list of severities | `low`, `medium`, `high`                 |


#### Example

```bash
curl -X GET "http://localhost:8080/v1/namespaced-resources/results?source=kyverno&status=fail&namespaces=test"
```

* Response `200`

```json
[
   {
      "id":"e8b7f35799c2d3cf9a50b492a8566e66dad465d9",
      "namespace":"test",
      "kind":"Pod",
      "name":"nginx",
      "message":"validation error: Running as root is not allowed. The fields spec.securityContext.runAsNonRoot, spec.containers[*].securityContext.runAsNonRoot, and spec.initContainers[*].securityContext.runAsNonRoot must be `true`. Rule check-containers[0] failed at path /spec/securityContext/runAsNonRoot/. Rule check-containers[1] failed at path /spec/containers/0/securityContext/.",
      "policy":"require-run-as-non-root",
      "rule":"check-containers",
      "status":"fail"
   }
]
```

* Response `500`

```json
{ "message": "Error Message" }
```

## V1 ClusterPolicyReport APIs

### Policies API

| Method | API                              | Description                                           | Codes |
|--------|----------------------------------|-------------------------------------------------------|----------------|
| `GET`  | `/v1/cluster-resources/policies` | List of all Policies<br>with cluster scoped results | `200`, `500`   |

#### Query Filter Parameters

| Filter      | Type       | Description      |
|-------------|------------|------------------|
| __source__  | `string`   | Filter by source |

#### Example

```bash
curl -X GET "http://localhost:8080/v1/cluster-resources/policies?source=kyverno"
```

* Response `200`

```json
[
   "require-ns-labels"
]
```

* Response `500`

```json
{ "message": "Error Message" }
```

### Kinds API

| Method | API                              | Description                                      | Codes |
|--------|----------------------------------|--------------------------------------------------|----------------|
| `GET`  | `/v1/cluster-resources/kinds`    | List of all Kinds<br>with cluster scoped results | `200`, `500`   |

#### Query Filter Parameters

| Filter      | Type       | Description      |
|-------------|------------|------------------|
| __source__  | `string`   | Filter by source |

#### Example

```bash
curl -X GET "http://localhost:8080/v1/cluster-resources/kinds?source=kyverno"
```

* Response `200`

```json
[
   "Namespace"
]
```

* Response `500`

```json
{ "message": "Error Message" }
```

### Sources API

| Method | API                              | Description                                        | Codes |
|--------|----------------------------------|----------------------------------------------------|----------------|
| `GET`  | `/v1/cluster-resources/sources`  | List of all Sources<br>with cluster scoped results | `200`, `500`   |

#### Example

```bash
curl -X GET "http://localhost:8080/v1/cluster-resources/sources"
```

* Response `200`

```json
[
   "Kyverno",
   "Kube Benchh"
]
```

* Response `500`

```json
{ "message": "Error Message" }
```

### Status Counts API

| Method | API                                      | Description                                        | Codes |
|--------|------------------------------------------|----------------------------------------------------|----------------|
| `GET`  | `/v1/cluster-resources/status-counts`    | Count of result status<br>per status and namespace | `200`, `500`   |

#### Query Filter Parameters

| Filter          | Type          | Description                    | Enum                                    |
|-----------------|---------------|--------------------------------|-----------------------------------------|
| __source__      | `string[]`    | Filter by a list of sources    |                                         |
| __kinds__       | `string[]`    | Filter by a list of kinds      |                                         |
| __categories__  | `string[]`    | Filter by a list of categories |                                         |
| __policies__    | `string[]`    | Filter by a list of policies   |                                         |
| __status__      | `string[]`    | Filter by a list of status     | `fail`, `pass`, `warn`, `error`, `skip` |
| __severities__  | `string[]`    | Filter by a list of severities | `low`, `medium`, `high`                 |


#### Example

```bash
curl -X GET "http://localhost:8080/v1/cluster-resources/status-counts?source=kyverno&status=pass&status=fail"
```

* Response `200`

```json
[
   {
      "status":"pass",
      "count":0
   },
   {
      "status":"fail",
      "count":26
   }
]
```

* Response `500`

```json
{ "message": "Error Message" }
```

### Results API

| Method | API                                | Description                    | Codes |
|--------|------------------------------------|--------------------------------|----------------|
| `GET`  | `/v1/cluster-resources/results`    | List of cluster scoped results | `200`, `500`   |

#### Query Filter Parameters

| Filter          | Type          | Description                    | Enum                                    |
|-----------------|---------------|--------------------------------|-----------------------------------------|
| __source__      | `string[]`    | Filter by a list of sources    |                                         |
| __kinds__       | `string[]`    | Filter by a list of kinds      |                                         |
| __categories__  | `string[]`    | Filter by a list of categories |                                         |
| __policies__    | `string[]`    | Filter by a list of policies   |                                         |
| __status__      | `string[]`    | Filter by a list of status     | `fail`, `pass`, `warn`, `error`, `skip` |
| __severities__  | `string[]`    | Filter by a list of severities | `low`, `medium`, `high`                 |


#### Example

```bash
curl -X GET "http://localhost:8080/v1/cluster-resources/results?source=kyverno&status=fail"
```

* Response `200`

```json
[
   {
      "id":"ca7c83998f8633b4e0da1de36e2996202e14e7a4",
      "kind":"Namespace",
      "name":"blog",
      "message":"validation error: The label `thisshouldntexist` is required. Rule check-for-labels-on-namespace failed at path /metadata/labels/thisshouldntexist/",
      "policy":"require-ns-labels",
      "rule":"check-for-labels-on-namespace",
      "status":"fail"
   }
]
```

* Response `500`

```json
{ "message": "Error Message" }
```

## Metrics

| Method | API        | Description             | Codes   |
|--------|------------|-------------------------|---------|
| `GET`  | `/metrics` | Prometheus Metrics API  | `200`   |

### cluster_policy_report_summary

Gauge: Summary count of each status per CluserPolicyReport

| Label       | Description                      |
|-------------|----------------------------------|
| `name`      | Name of the ClusterPolicyReport  |
| `status`    | Status of the Summary count      |

### cluster_policy_report_result

Gauge: One Entry represent one Result in a ClusterPolicyReport. Deleted Results will also be removed from the Metrics

| Label       | Description                                                 |
|-------------|-------------------------------------------------------------|
| `category`  | Category of the Result                                      |
| `kind`      | Kind of the result resource                                 |
| `name`      | Name of the result resource                                 |
| `policy`    | Policy of the result                                        |
| `report`    | Name of the ClusterPolicyReport where this result was found |
| `rule`      | Rule of the result                                          |
| `severity`  | Severity of the result                                      |
| `status`    | Status of the Result                                        |

### policy_report_summary

Gauge: Summary count of each status per PolicyReport

| Label       | Description                      |
|-------------|----------------------------------|
| `name`      | Name of the PolicyReport         |
| `status`    | Status of the Summary count      |
| `namespace` | Namespace of the PolicyReport    |

### policy_report_result

Gauge: One Entry represent one Result in a PolicyReport. Deleted Results will also be removed from the Metrics

| Label       | Description                                                 |
|-------------|-------------------------------------------------------------|
| `category`  | Category of the Result                                      |
| `kind`      | Kind of the result resource                                 |
| `name`      | Name of the result resource                                 |
| `namespace` | Namespace of the result resource                            |
| `policy`    | Policy of the result                                        |
| `report`    | Name of the ClusterPolicyReport where this result was found |
| `rule`      | Rule of the result                                          |
| `severity`  | Rule of the severity                                        |
| `status`    | Status of the Result                                        |

### Example

```bash
curl -X GET "http://localhost:8080/metrics"
```

* Response `200`

```text
# HELP cluster_policy_report_result List of all ClusterPolicyReport Results
# TYPE cluster_policy_report_result gauge
cluster_policy_report_result{category="",kind="Namespace",name="argo-cd",policy="require-ns-labels",report="clusterpolicyreport",rule="check-for-labels-on-namespace",severity="",status="fail"} 1

# HELP cluster_policy_report_summary Summary of all ClusterPolicyReports
# TYPE cluster_policy_report_summary gauge
cluster_policy_report_summary{name="clusterpolicyreport",status="Error"} 0
cluster_policy_report_summary{name="clusterpolicyreport",status="Fail"} 26
cluster_policy_report_summary{name="clusterpolicyreport",status="Pass"} 0
cluster_policy_report_summary{name="clusterpolicyreport",status="Skip"} 0
cluster_policy_report_summary{name="clusterpolicyreport",status="Warn"} 0

# HELP policy_report_result List of all PolicyReport Results
# TYPE policy_report_result gauge
policy_report_result{category="Pod Security Standards (Default)",kind="Pod",name="nginx",namespace="test",policy="disallow-add-capabilities",report="polr-ns-test",rule="capabilities",severity="medium",status="pass"} 1

# HELP policy_report_summary Summary of all PolicyReports
# TYPE policy_report_summary gauge
policy_report_summary{name="polr-ns-test",namespace="test",status="Error"} 0
policy_report_summary{name="polr-ns-test",namespace="test",status="Fail"} 1
policy_report_summary{name="polr-ns-test",namespace="test",status="Pass"} 34
policy_report_summary{name="polr-ns-test",namespace="test",status="Skip"} 0
policy_report_summary{name="polr-ns-test",namespace="test",status="Warn"} 0
```
