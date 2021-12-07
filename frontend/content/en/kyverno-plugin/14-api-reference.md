---
title: API Reference
description: ''
position: 14
category: 'Kyverno Plugin'
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
{ "error": "No Kyverno CRDs found" }
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

## CRD APIs

### Policies API

| Method | API         | Description                                        | Codes |
|--------|-------------|----------------------------------------------------|----------------|
| `GET`  | `/policies` | List of all avilable Policies and ClusterPolicies  | `200`, `500`   |

#### Example

```bash
curl -X GET "http://localhost:8080/policies"
```

* Response `200`

```json
[
   {
      "kind":"ClusterPolicy",
      "name":"deny-privilege-escalation",
      "autogenControllers":[
         "DaemonSet",
         "Deployment",
         "Job",
         "StatefulSet",
         "CronJob"
      ],
      "validationFailureAction":"audit",
      "background":true,
      "rules":[
         {
            "message":"Privilege escalation is disallowed. The fields spec.containers[*].securityContext.allowPrivilegeEscalation, and spec.initContainers[*].securityContext.allowPrivilegeEscalation must be undefined or set to `false`.",
            "name":"deny-privilege-escalation",
            "type":"validation"
         },
         {
            "message":"Privilege escalation is disallowed. The fields spec.containers[*].securityContext.allowPrivilegeEscalation, and spec.initContainers[*].securityContext.allowPrivilegeEscalation must be undefined or set to `false`.",
            "name":"autogen-deny-privilege-escalation",
            "type":"validation"
         },
         {
            "message":"Privilege escalation is disallowed. The fields spec.containers[*].securityContext.allowPrivilegeEscalation, and spec.initContainers[*].securityContext.allowPrivilegeEscalation must be undefined or set to `false`.",
            "name":"autogen-cronjob-deny-privilege-escalation",
            "type":"validation"
         }
      ],
      "category":"Pod Security Standards (Restricted)",
      "description":"Privilege escalation, such as via set-user-ID or set-group-ID file mode, should not be allowed.",
      "severity":"medium",
      "creationTimestamp":"2021-11-07T18:32:40Z",
      "uid":"7cabc2f3-0e9b-4d1e-a434-a19275a54d29",
      "content":"apiVersion: kyverno.io/v1\nkind: ClusterPolicy\nmetadata:\n  annotations:\n    pod-policies.kyverno.io/autogen-controllers: DaemonSet,Deployment,Job,StatefulSet,CronJob\n    policies.kyverno.io/category: Pod Security Standards (Restricted)\n    policies.kyverno.io/description: Privilege escalation, such as via set-user-ID\n      or set-group-ID file mode, should not be allowed.\n    policies.kyverno.io/severity: medium\n  creationTimestamp: \"2021-11-07T18:32:40Z\"\n  generation: 16\n  labels:\n    app: kyverno\n    app.kubernetes.io/component: kyverno\n    app.kubernetes.io/instance: kyverno-policies\n    app.kubernetes.io/managed-by: Helm\n    app.kubernetes.io/name: kyverno-policies\n    app.kubernetes.io/part-of: kyverno-policies\n    app.kubernetes.io/version: v2.1.3\n    argocd.argoproj.io/instance: kyverno-policies\n    helm.sh/chart: kyverno-policies-v2.1.3\n  name: deny-privilege-escalation\n  resourceVersion: \"1742766\"\n  uid: 7cabc2f3-0e9b-4d1e-a434-a19275a54d29\nspec:\n  background: true\n  failurePolicy: Fail\n  rules:\n  - exclude:\n      resources: {}\n    generate:\n      clone: {}\n    match:\n      resources:\n        kinds:\n        - Pod\n    mutate: {}\n    name: deny-privilege-escalation\n    validate:\n      message: Privilege escalation is disallowed. The fields spec.containers[*].securityContext.allowPrivilegeEscalation,\n        and spec.initContainers[*].securityContext.allowPrivilegeEscalation must be\n        undefined or set to `false`.\n      pattern:\n        spec:\n          =(initContainers):\n          - =(securityContext):\n              =(allowPrivilegeEscalation): \"false\"\n          containers:\n          - =(securityContext):\n              =(allowPrivilegeEscalation): \"false\"\n  - exclude:\n      resources: {}\n    generate:\n      clone: {}\n    match:\n      resources:\n        kinds:\n        - DaemonSet\n        - Deployment\n        - Job\n        - StatefulSet\n    mutate: {}\n    name: autogen-deny-privilege-escalation\n    validate:\n      message: Privilege escalation is disallowed. The fields spec.containers[*].securityContext.allowPrivilegeEscalation,\n        and spec.initContainers[*].securityContext.allowPrivilegeEscalation must be\n        undefined or set to `false`.\n      pattern:\n        spec:\n          template:\n            spec:\n              =(initContainers):\n              - =(securityContext):\n                  =(allowPrivilegeEscalation): \"false\"\n              containers:\n              - =(securityContext):\n                  =(allowPrivilegeEscalation): \"false\"\n  - exclude:\n      resources: {}\n    generate:\n      clone: {}\n    match:\n      resources:\n        kinds:\n        - CronJob\n    mutate: {}\n    name: autogen-cronjob-deny-privilege-escalation\n    validate:\n      message: Privilege escalation is disallowed. The fields spec.containers[*].securityContext.allowPrivilegeEscalation,\n        and spec.initContainers[*].securityContext.allowPrivilegeEscalation must be\n        undefined or set to `false`.\n      pattern:\n        spec:\n          jobTemplate:\n            spec:\n              template:\n                spec:\n                  =(initContainers):\n                  - =(securityContext):\n                      =(allowPrivilegeEscalation): \"false\"\n                  containers:\n                  - =(securityContext):\n                      =(allowPrivilegeEscalation): \"false\"\n  validationFailureAction: audit\n"
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

### kyverno_policy

Gauge: One Entry represent one __Rule__ of a Policy or ClusterPolicy. Deleted Policies and Rules will also be removed from the Metrics

| Label                     | Description                                          |
|---------------------------|------------------------------------------------------|
| `background`              | Background scan enabled or disabled                  |
| `category`                | Category of the policy                               |
| `kind`                    | Policy or ClusterPolicy                              |
| `namespace`               | Namespace of the policy                              |
| `policy`                  | Name of the policy                                   |
| `rule`                    | Name of the rule within the policy                   |
| `rule`                    | Rule of the result                                   |
| `severity`                | Severity of the policy                               |
| `type`                    | Type of the rule: validation / mutation / generation |
| `validationFailureAction` | validationFailureAction of the rule: audit / enforce |

### Example

```bash
curl -X GET "http://localhost:8080/metrics"
```

* Response `200`

```text
# HELP policy_report_kyverno_policy List of all Policies
# TYPE policy_report_kyverno_policy gauge
kyverno_policy{background="true",category="",kind="ClusterPolicy",namespace="",policy="require-ns-labels",rule="check-for-labels-on-namespace",severity="",type="validation",validationFailureAction="audit"} 1
kyverno_policy{background="true",category="Pod Security Standards (Default)",kind="ClusterPolicy",namespace="",policy="disallow-add-capabilities",rule="autogen-capabilities",severity="medium",type="validation",validationFailureAction="audit"} 1
kyverno_policy{background="true",category="Pod Security Standards (Default)",kind="ClusterPolicy",namespace="",policy="disallow-add-capabilities",rule="autogen-cronjob-capabilities",severity="medium",type="validation",validationFailureAction="audit"} 1
```
