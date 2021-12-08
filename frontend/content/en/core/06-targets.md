---
title: Targets
description: ''
position: 6
category: 'Policy Reporter'
---

Policy Reporter supports different targets to send new PolicyReport Results too. This makes it possible to create a log or get notified as soon as a new validation Result was detected. The set of supported tools are based on personal needs and user requests. Feel free to create an issue or pull request if you want support for a new target.

## Target Configurations

Each Target has similar configuration values. Required is always a valid and accessable __host__ or __webhook__ configuration to be able to send the events.

Possible filters for events currently are:
* __minimumPriority__: Only events with the given priority or higher are sent, by default each priority is sent. See [priority mapping](/core/08-priority-mapping) for details.
* __sources__: Sent only Results of the configured sources, by default Results from all sources are sent.
* __skipExistingOnStartup__: On startup Policy Reporter registers all existing Results in the cluster. By default this Results are ignored. If you also want to sent them to your target you can set this option to *false*.

## Grafana Loki

Policy Reporter can send Results directly to Grafana Loki without the need of Promtail. Each Result includes all available information as labels as well as a *source* label with *policy-reporter* as value. To query all messages from Policy Reporter just use `{source="policy-reporter"}` as Query.

### Example

The minimal configuration for Grafana Loki requires a valid and accessable host.

```yaml
loki:
  host: "http://loki.loki-stack:3100"
  minimumPriority: "warning"
  skipExistingOnStartup: true
  sources:
  - kyverno
```

### Screenshot
<a href="/images/targets/grafana-loki.png" target="_blank">
    <img src="/images/targets/grafana-loki.png" style="border: 1px solid #555" alt="Grafana Loki Screenshot with PolicyReportResults" />
</a>

## Elasticsearch

Policy Reporter sends Results in a JSON representation and all available information to a customizable index. This index can be expanded by selecting one of the different rotations or *none* to disable this function.. By default Policy Reporter creates a new index on daily bases.

### Additional configuration
* __index__ is used as index name, uses `policy-reporter` if not configured
* __rotation__ is used to create rotating indexes by adding the rotation date as suffix to the index name. Possible values are `daily`, `monthly`, `annually` and `none`, uses `daily` if not configured

### Example

The minimal configuration for Elasticsearch requires a valid and accessable host.

```yaml
elasticsearch:
  host: "http://loki.loki-stack:3100"
  index: "policy-reporter"
  rotation: "daily"
  minimumPriority: "warning"
  skipExistingOnStartup: true
  sources:
  - kyverno
```

### Screenshot

<a href="/images/targets/elasticsearch.png" target="_blank">
    <img src="/images/targets/elasticsearch.png" style="border: 1px solid #555" alt="Elasticvue Screenshot with PolicyReportResults" />
</a>

## Microsoft Teams

Send new PolicyReportResults with all available information over the Webhook API to Microsoft Teams.

### Example

The minimal configuration for Microsoft Teams requires a valid and accessable webhook URL.

```yaml
teams:
  webhook: "https://m365x682156.webhook.office.com"
  minimumPriority: "critical"
  skipExistingOnStartup: true
  sources:
  - kyverno
```

### Screenshot
<a href="/images/targets/ms-teams.png" target="_blank">
    <img src="/images/targets/ms-teams.png" style="border: 1px solid #555" alt="MS Teams Notification for a PolicyReportResult" />
</a>

## Slack

Send new PolicyReportResults with all available information over the Webhook API to Slack.

### Example

The minimal configuration for Slack requires a valid and accessable webhook URL.

```yaml
slack:
  webhook: "https://hooks.slack.com/services/..."
  minimumPriority: "critical"
  skipExistingOnStartup: true
  sources:
  - kyverno
```

### Screenshot
<a href="/images/targets/slack.png" target="_blank">
    <img src="/images/targets/slack.png" style="border: 1px solid #555" alt="Slack Notification for a PolicyReportResult" />
</a>

## Discord

Send new PolicyReportResults with all available information over the Webhook API to Discord.

### Example

The minimal configuration for Discord requires a valid and accessable webhook URL.

```yaml
discord:
  webhook: "https://discordapp.com/api/webhooks/..."
  minimumPriority: "critical"
  skipExistingOnStartup: true
  sources:
  - kyverno
```

### Screenshot
<a href="/images/targets/discord.png" target="_blank">
    <img src="/images/targets/discord.png" style="border: 1px solid #555" alt="Discord Notification for a PolicyReportResult" />
</a>

## Policy Reporter UI

Send new PolicyReportResults with all available information as JSON to the REST API of Policy Reporter UI. You can find the received Results in the __Logs__ view of the UI. The results are stored in memory, the max number of stored results can be configured by in Policy Reporter UI.

If Policy Reporter UI is installed via the Helm Charts it is configured as Target by default with a minimumPriority of *warning* and a maximal logSize of *200*. The LogSize can be coniured in the [Policy Reporter UI Subchart](/guide/04-helm-chart-core#log-size).

### Example

The minimal configuration for Discord requires a valid and accessable webhook URL.

```yaml
ui:
  host: "http://policy-reporter-ui:8080"
  minimumPriority: "warning"
  skipExistingOnStartup: true
  sources:
  - kyverno
```

### Screenshot
<a href="/images/targets/policy-reporter-log-dark.png" target="_blank" class="dark-img">
    <img src="/images/targets/policy-reporter-log-dark.png" style="border: 1px solid #555" alt="Policy Reporter UI - Logs View dark mode" />
</a>

<a href="/images/targets/policy-reporter-log-light.png" target="_blank" class="light-img">
    <img src="/images/targets/policy-reporter-log-light.png" style="border: 1px solid #555" alt="Policy Reporter UI - Logs View light mode" />
</a>

## S3 compatible Storage

Policy Reporter can also send Results to S3 compatible Services like __MinIO__, __Yandex__ or __AWS S3__.

It persists each Result als JSON in the following structure: `s3://<bucket>/<prefix>/YYYY-MM-DD/<policy-name>-<result-id>-YYYY-MM-DDTHH:mm:ss.json`

### Additional Configure

* __endpoint__ to the S3 API
* __accessKeyID__ and __secretAccessKey__ for authentication with the required write permissions for the selected bucket
* __bucket__ in which the results are persisted
* __region__ of the bucket
* __prefix__ of the file path, uses "policy-reporter" as default

### Example

```yaml
s3:
  endpoint: https://storage.yandexcloud.net
  region: ru-central1
  bucket: dev-cluster
  secretAccessKey: secretAccessKey
  accessKeyID: accessKeyID
  minimumPriority: "warning"
  skipExistingOnStartup: true
  sources:
  - kyverno
```

### Content Example

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
   "Timestamp":"2021-12-04T10:13:02Z",
   "Resource":{
      "APIVersion":"v1",
      "Kind":"Pod",
      "Name":"nginx2",
      "Namespace":"test",
      "UID":"ac4d11f3-0aa8-43f0-8056-98f4eae0d956"
   }
}
```

## Configuration Reference

<code-group>
  <code-block label="Helm 3" active>

  ```yaml
  # values.yaml
  target:
    loki:
      host: ""
      minimumPriority: ""
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
      host: http://policy-reporter-ui:8080
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

  </code-block>
  <code-block label="config.yaml">

  ```yaml
  loki:
    host: ""
    minimumPriority: ""
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
    host: http://policy-reporter-ui:8080
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
  </code-block>
</code-group>
