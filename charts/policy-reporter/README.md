# policy-reporter-preview

Policy Reporter watches for PolicyReport Resources.
It creates Prometheus Metrics and can send rule validation events to different targets like Loki, Elasticsearch, Slack or Discord

![Version: 3.0.0-alpha](https://img.shields.io/badge/Version-3.0.0--alpha-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 3.0.0-alpha](https://img.shields.io/badge/AppVersion-3.0.0--alpha-informational?style=flat-square)

## Documentation

You can find detailed Information and Screens about Features and Configurations in the [Documentation](https://kyverno.github.io/policy-reporter/guide/02-getting-started#core--policy-reporter-ui).

## Installation with Helm v3

Installation via Helm Repository

### Add the Helm repository
```bash
helm repo add policy-reporter https://kyverno.github.io/policy-reporter
helm repo update
```

### Basic Installation

The basic installation provides an Prometheus Metrics Endpoint and different REST APIs, for more details have a look at the [Documentation](https://kyverno.github.io/policy-reporter/guide/02-getting-started).

```bash
helm install policy-reporter policy-reporter/policy-reporter -n policy-reporter --create-namespace
```

## Policy Reporter UI

You can use the Policy Reporter as standalone Application along with the optional UI SubChart.

### Installation with Policy Reporter UI and Kyverno Plugin enabled
```bash
helm install policy-reporter policy-reporter/policy-reporter --set kyvernoPlugin.enabled=true --set ui.enabled=true --set ui.plugins.kyverno=true -n policy-reporter --create-namespace
kubectl port-forward service/policy-reporter-ui 8082:8080 -n policy-reporter
```
Open `http://localhost:8082/` in your browser.

Check the [Documentation](https://kyverno.github.io/policy-reporter/guide/02-getting-started#core--policy-reporter-ui) for Screens and additional Information

## Resources

* [[Video] 37. #EveryoneCanContribute cafe: Policy reporter for Kyverno](https://youtu.be/1mKywg9f5Fw)
* [[Video] Rawkode Live: Hands on Policy Reporter](https://www.youtube.com/watch?v=ZrOtTELNLyg)
* [[Blog] Monitor Security and Best Practices with Kyverno and Policy Reporter](https://blog.webdev-jogeleit.de/blog/monitor-security-with-kyverno-and-policy-reporter/)

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| nameOverride | string | `""` |  |
| image.registry | string | `"ghcr.io"` |  |
| image.repository | string | `"kyverno/policy-reporter"` |  |
| image.pullPolicy | string | `"IfNotPresent"` |  |
| image.tag | string | `"2.17.2"` |  |
| imagePullSecrets | list | `[]` |  |
| priorityClassName | string | `""` |  |
| replicaCount | int | `1` |  |
| revisionHistoryLimit | int | `10` |  |
| deploymentStrategy | object | `{}` |  |
| port.name | string | `"http"` |  |
| port.number | int | `8080` |  |
| annotations | object | `{}` |  |
| rbac.enabled | bool | `true` |  |
| serviceAccount.create | bool | `true` |  |
| serviceAccount.annotations | object | `{}` |  |
| serviceAccount.name | string | `""` |  |
| service.enabled | bool | `true` |  |
| service.annotations | object | `{}` |  |
| service.labels | object | `{}` |  |
| service.type | string | `"ClusterIP"` |  |
| service.port | int | `8080` |  |
| podSecurityContext.fsGroup | int | `1234` |  |
| securityContext.runAsUser | int | `1234` |  |
| securityContext.runAsNonRoot | bool | `true` |  |
| securityContext.privileged | bool | `false` |  |
| securityContext.allowPrivilegeEscalation | bool | `false` |  |
| securityContext.readOnlyRootFilesystem | bool | `true` |  |
| securityContext.capabilities.drop[0] | string | `"ALL"` |  |
| securityContext.seccompProfile.type | string | `"RuntimeDefault"` |  |
| podAnnotations | object | `{}` |  |
| podLabels | object | `{}` |  |
| resources | object | `{}` |  |
| networkPolicy.enabled | bool | `false` |  |
| networkPolicy.egress[0].to | string | `nil` |  |
| networkPolicy.egress[0].ports[0].protocol | string | `"TCP"` |  |
| networkPolicy.egress[0].ports[0].port | int | `6443` |  |
| networkPolicy.ingress | list | `[]` |  |
| ingress.enabled | bool | `false` |  |
| ingress.className | string | `""` |  |
| ingress.labels | object | `{}` |  |
| ingress.annotations | object | `{}` |  |
| ingress.hosts[0].host | string | `"chart-example.local"` |  |
| ingress.hosts[0].paths | list | `[]` |  |
| ingress.tls | list | `[]` |  |
| logging.encoding | string | `"console"` |  |
| logging.logLevel | int | `0` |  |
| logging.development | bool | `false` |  |
| api.logging | bool | `false` |  |
| rest.enabled | bool | `false` |  |
| metrics.enabled | bool | `false` |  |
| metrics.mode | string | `"detailed"` |  |
| metrics.customLabels | list | `[]` |  |
| profiling.enabled | bool | `false` |  |
| worker | int | `5` |  |
| reportFilter.namespaces.include | list | `[]` |  |
| reportFilter.namespaces.exclude | list | `[]` |  |
| reportFilter.clusterReports.disabled | bool | `false` |  |
| ui.enabled | bool | `false` |  |
| kyvernoPlugin.enabled | bool | `false` |  |
| monitoring.enabled | bool | `false` |  |
| database.type | string | `""` |  |
| database.database | string | `""` |  |
| database.username | string | `""` |  |
| database.password | string | `""` |  |
| database.host | string | `""` |  |
| database.enableSSL | bool | `false` |  |
| database.dsn | string | `""` |  |
| database.secretRef | string | `""` |  |
| database.mountedSecret | string | `""` |  |
| global.plugins.kyverno | bool | `false` |  |
| global.backend | string | `""` |  |
| global.fullnameOverride | string | `""` |  |
| global.namespace | string | `""` |  |
| global.labels | object | `{}` |  |
| global.basicAuth.username | string | `""` |  |
| global.basicAuth.password | string | `""` |  |
| global.basicAuth.secretRef | string | `""` |  |
| policyPriorities | object | `{}` |  |
| emailReports.clusterName | string | `""` |  |
| emailReports.titlePrefix | string | `"Report"` |  |
| emailReports.smtp.secret | string | `""` |  |
| emailReports.smtp.host | string | `""` |  |
| emailReports.smtp.port | int | `465` |  |
| emailReports.smtp.username | string | `""` |  |
| emailReports.smtp.password | string | `""` |  |
| emailReports.smtp.from | string | `""` |  |
| emailReports.smtp.encryption | string | `""` |  |
| emailReports.summary.enabled | bool | `false` |  |
| emailReports.summary.schedule | string | `"0 8 * * *"` |  |
| emailReports.summary.activeDeadlineSeconds | int | `300` |  |
| emailReports.summary.backoffLimit | int | `3` |  |
| emailReports.summary.ttlSecondsAfterFinished | int | `0` |  |
| emailReports.summary.restartPolicy | string | `"Never"` |  |
| emailReports.summary.to | list | `[]` |  |
| emailReports.summary.filter | object | `{}` |  |
| emailReports.summary.channels | list | `[]` |  |
| emailReports.violations.enabled | bool | `false` |  |
| emailReports.violations.schedule | string | `"0 8 * * *"` |  |
| emailReports.violations.activeDeadlineSeconds | int | `300` |  |
| emailReports.violations.backoffLimit | int | `3` |  |
| emailReports.violations.ttlSecondsAfterFinished | int | `0` |  |
| emailReports.violations.restartPolicy | string | `"Never"` |  |
| emailReports.violations.to | list | `[]` |  |
| emailReports.violations.filter | object | `{}` |  |
| emailReports.violations.channels | list | `[]` |  |
| existingTargetConfig.enabled | bool | `false` |  |
| existingTargetConfig.name | string | `""` |  |
| existingTargetConfig.subPath | string | `""` |  |
| target.loki.host | string | `""` |  |
| target.loki.certificate | string | `""` |  |
| target.loki.skipTLS | bool | `false` |  |
| target.loki.secretRef | string | `""` |  |
| target.loki.mountedSecret | string | `""` |  |
| target.loki.path | string | `""` |  |
| target.loki.minimumPriority | string | `""` |  |
| target.loki.sources | list | `[]` |  |
| target.loki.skipExistingOnStartup | bool | `true` |  |
| target.loki.customLabels | object | `{}` |  |
| target.loki.filter | object | `{}` |  |
| target.loki.channels | list | `[]` |  |
| target.elasticsearch.host | string | `""` |  |
| target.elasticsearch.certificate | string | `""` |  |
| target.elasticsearch.skipTLS | bool | `false` |  |
| target.elasticsearch.index | string | `""` |  |
| target.elasticsearch.username | string | `""` |  |
| target.elasticsearch.password | string | `""` |  |
| target.elasticsearch.secretRef | string | `""` |  |
| target.elasticsearch.mountedSecret | string | `""` |  |
| target.elasticsearch.rotation | string | `""` |  |
| target.elasticsearch.minimumPriority | string | `""` |  |
| target.elasticsearch.sources | list | `[]` |  |
| target.elasticsearch.skipExistingOnStartup | bool | `true` |  |
| target.elasticsearch.customFields | object | `{}` |  |
| target.elasticsearch.filter | object | `{}` |  |
| target.elasticsearch.channels | list | `[]` |  |
| target.slack.webhook | string | `""` |  |
| target.slack.channel | string | `""` |  |
| target.slack.secretRef | string | `""` |  |
| target.slack.mountedSecret | string | `""` |  |
| target.slack.minimumPriority | string | `""` |  |
| target.slack.sources | list | `[]` |  |
| target.slack.skipExistingOnStartup | bool | `true` |  |
| target.slack.customFields | object | `{}` |  |
| target.slack.filter | object | `{}` |  |
| target.slack.channels | list | `[]` |  |
| target.discord.webhook | string | `""` |  |
| target.discord.secretRef | string | `""` |  |
| target.discord.mountedSecret | string | `""` |  |
| target.discord.minimumPriority | string | `""` |  |
| target.discord.sources | list | `[]` |  |
| target.discord.skipExistingOnStartup | bool | `true` |  |
| target.discord.filter | object | `{}` |  |
| target.discord.channels | list | `[]` |  |
| target.teams.webhook | string | `""` |  |
| target.teams.secretRef | string | `""` |  |
| target.teams.mountedSecret | string | `""` |  |
| target.teams.certificate | string | `""` |  |
| target.teams.skipTLS | bool | `false` |  |
| target.teams.minimumPriority | string | `""` |  |
| target.teams.sources | list | `[]` |  |
| target.teams.skipExistingOnStartup | bool | `true` |  |
| target.teams.filter | object | `{}` |  |
| target.teams.channels | list | `[]` |  |
| target.ui.host | string | `""` |  |
| target.ui.certificate | string | `""` |  |
| target.ui.skipTLS | bool | `false` |  |
| target.ui.minimumPriority | string | `"warning"` |  |
| target.ui.sources | list | `[]` |  |
| target.ui.skipExistingOnStartup | bool | `true` |  |
| target.webhook.host | string | `""` |  |
| target.webhook.certificate | string | `""` |  |
| target.webhook.skipTLS | bool | `false` |  |
| target.webhook.secretRef | string | `""` |  |
| target.webhook.mountedSecret | string | `""` |  |
| target.webhook.headers | object | `{}` |  |
| target.webhook.minimumPriority | string | `""` |  |
| target.webhook.sources | list | `[]` |  |
| target.webhook.skipExistingOnStartup | bool | `true` |  |
| target.webhook.customFields | object | `{}` |  |
| target.webhook.filter | object | `{}` |  |
| target.webhook.channels | list | `[]` |  |
| target.telegram.token | string | `""` |  |
| target.telegram.chatID | string | `""` |  |
| target.telegram.host | string | `""` |  |
| target.telegram.certificate | string | `""` |  |
| target.telegram.skipTLS | bool | `false` |  |
| target.telegram.secretRef | string | `""` |  |
| target.telegram.mountedSecret | string | `""` |  |
| target.telegram.headers | object | `{}` |  |
| target.telegram.minimumPriority | string | `""` |  |
| target.telegram.sources | list | `[]` |  |
| target.telegram.skipExistingOnStartup | bool | `true` |  |
| target.telegram.customFields | object | `{}` |  |
| target.telegram.filter | object | `{}` |  |
| target.telegram.channels | list | `[]` |  |
| target.googleChat.webhook | string | `""` |  |
| target.googleChat.certificate | string | `""` |  |
| target.googleChat.skipTLS | bool | `false` |  |
| target.googleChat.secretRef | string | `""` |  |
| target.googleChat.mountedSecret | string | `""` |  |
| target.googleChat.headers | object | `{}` |  |
| target.googleChat.minimumPriority | string | `""` |  |
| target.googleChat.sources | list | `[]` |  |
| target.googleChat.skipExistingOnStartup | bool | `true` |  |
| target.googleChat.customFields | object | `{}` |  |
| target.googleChat.filter | object | `{}` |  |
| target.googleChat.channels | list | `[]` |  |
| target.s3.accessKeyID | string | `""` |  |
| target.s3.secretAccessKey | string | `""` |  |
| target.s3.secretRef | string | `""` |  |
| target.s3.mountedSecret | string | `""` |  |
| target.s3.region | string | `""` |  |
| target.s3.endpoint | string | `""` |  |
| target.s3.bucket | string | `""` |  |
| target.s3.bucketKeyEnabled | bool | `false` |  |
| target.s3.kmsKeyId | string | `""` |  |
| target.s3.serverSideEncryption | string | `""` |  |
| target.s3.pathStyle | bool | `false` |  |
| target.s3.prefix | string | `""` |  |
| target.s3.minimumPriority | string | `""` |  |
| target.s3.sources | list | `[]` |  |
| target.s3.skipExistingOnStartup | bool | `true` |  |
| target.s3.customFields | object | `{}` |  |
| target.s3.filter | object | `{}` |  |
| target.s3.channels | list | `[]` |  |
| target.kinesis.accessKeyID | string | `""` |  |
| target.kinesis.secretAccessKey | string | `""` |  |
| target.kinesis.secretRef | string | `""` |  |
| target.kinesis.mountedSecret | string | `""` |  |
| target.kinesis.region | string | `""` |  |
| target.kinesis.endpoint | string | `""` |  |
| target.kinesis.streamName | string | `""` |  |
| target.kinesis.minimumPriority | string | `""` |  |
| target.kinesis.sources | list | `[]` |  |
| target.kinesis.skipExistingOnStartup | bool | `true` |  |
| target.kinesis.customFields | object | `{}` |  |
| target.kinesis.filter | object | `{}` |  |
| target.kinesis.channels | list | `[]` |  |
| target.securityHub.accessKeyID | string | `""` |  |
| target.securityHub.secretAccessKey | string | `""` |  |
| target.securityHub.secretRef | string | `""` |  |
| target.securityHub.mountedSecret | string | `""` |  |
| target.securityHub.region | string | `""` |  |
| target.securityHub.endpoint | string | `""` |  |
| target.securityHub.accountID | string | `""` |  |
| target.securityHub.minimumPriority | string | `""` |  |
| target.securityHub.sources | list | `[]` |  |
| target.securityHub.skipExistingOnStartup | bool | `true` |  |
| target.securityHub.customFields | object | `{}` |  |
| target.securityHub.filter | object | `{}` |  |
| target.securityHub.channels | list | `[]` |  |
| target.gcs.credentials | string | `""` |  |
| target.gcs.secretRef | string | `""` |  |
| target.gcs.mountedSecret | string | `""` |  |
| target.gcs.bucket | string | `""` |  |
| target.gcs.minimumPriority | string | `""` |  |
| target.gcs.sources | list | `[]` |  |
| target.gcs.skipExistingOnStartup | bool | `true` |  |
| target.gcs.customFields | object | `{}` |  |
| target.gcs.filter | object | `{}` |  |
| target.gcs.channels | list | `[]` |  |
| leaderElection.enabled | bool | `false` |  |
| leaderElection.releaseOnCancel | bool | `true` |  |
| leaderElection.leaseDuration | int | `15` |  |
| leaderElection.renewDeadline | int | `10` |  |
| leaderElection.retryPeriod | int | `2` |  |
| redis.enabled | bool | `false` |  |
| redis.address | string | `""` |  |
| redis.database | int | `0` |  |
| redis.prefix | string | `"policy-reporter"` |  |
| redis.username | string | `""` |  |
| redis.password | string | `""` |  |
| podDisruptionBudget.minAvailable | int | `1` | Configures the minimum available pods for policy-reporter disruptions. Cannot be used if `maxUnavailable` is set. |
| podDisruptionBudget.maxUnavailable | string | `nil` | Configures the maximum unavailable pods for policy-reporter disruptions. Cannot be used if `minAvailable` is set. |
| nodeSelector | object | `{}` |  |
| tolerations | list | `[]` |  |
| affinity | object | `{}` |  |
| topologySpreadConstraints | list | `[]` |  |
| livenessProbe.httpGet.path | string | `"/ready"` |  |
| livenessProbe.httpGet.port | string | `"http"` |  |
| readinessProbe.httpGet.path | string | `"/healthz"` |  |
| readinessProbe.httpGet.port | string | `"http"` |  |
| extraVolumes.volumeMounts | list | `[]` |  |
| extraVolumes.volumes | list | `[]` |  |
| sqliteVolume | object | `{}` |  |

## Source Code

* <https://github.com/kyverno/policy-reporter>

## Requirements

| Repository | Name | Version |
|------------|------|---------|
|  | kyvernoPlugin | 1.6.2 |
|  | monitoring | 2.8.1 |
|  | ui | 2.10.2 |

## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| Frank Jogeleit |  |  |

----------------------------------------------
Autogenerated from chart metadata using [helm-docs v1.11.0](https://github.com/norwoodj/helm-docs/releases/v1.11.0)
