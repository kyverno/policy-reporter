---
title: Config Reference
description: ''
position: 10
category: 'Policy Reporter'
---

Configuration File Reference with all possible options.

```yaml
kubeconfig: '~/.kube/config' 

api:
  port: 8080

rest:
  enabled: false

metrics:
  enabled: false

dbfile: "sqlite-database.db"

priorityMap: {}

loki:
  host: ""
  minimumPriority: "warning"
  skipExistingOnStartup: true
  sources: []

elasticsearch:
  host: ""
  index: "policy-reporter"
  rotation: "dayli"
  minimumPriority: ""
  skipExistingOnStartup: true
  sources: []

slack:
  webhook: ""
  minimumPriority: ""
  skipExistingOnStartup: true
  sources: []

discord:
  webhook: ""
  minimumPriority: ""
  skipExistingOnStartup: true
  sources: []

teams:
  webhook: ""
  minimumPriority: ""
  skipExistingOnStartup: true
  sources: []

ui:
  host: ""
  minimumPriority: "warning"
  skipExistingOnStartup: true
  sources: []

s3:
  endpoint: ""
  region: ""
  bucket: ""
  secretAccessKey: ""
  accessKeyID: ""
  minimumPriority: "warning"
  skipExistingOnStartup: true
  sources: []
```