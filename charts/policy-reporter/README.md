# policy-reporter-preview

Policy Reporter watches for PolicyReport Resources.
It creates Prometheus Metrics and can send rule validation events to different targets like Loki, Elasticsearch, Slack or Discord

![Version: 3.0.0-beta.16](https://img.shields.io/badge/Version-3.0.0--beta.16-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 3.0.0-beta](https://img.shields.io/badge/AppVersion-3.0.0--beta-informational?style=flat-square)

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
| fullnameOverride | string | `"policy-reporter"` |  |
| namespaceOverride | string | `""` |  |
| image.registry | string | `"ghcr.io"` |  |
| image.repository | string | `"kyverno/policy-reporter"` |  |
| image.pullPolicy | string | `"IfNotPresent"` |  |
| image.tag | string | `"ac016ac"` |  |
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
| reportFilter.disableClusterReports | bool | `false` |  |
| sourceConfig | list | `[]` | Customize source specific logic like result ID generation |
| sourceFilters | list | `[{"disableClusterReports":false,"kinds":{"exclude":["ReplicaSet"]},"selector":{"source":"kyverno"},"uncontrolledOnly":true}]` | Source based PolicyReport filter |
| sourceFilters[0] | object | `{"disableClusterReports":false,"kinds":{"exclude":["ReplicaSet"]},"selector":{"source":"kyverno"},"uncontrolledOnly":true}` | PolicyReport selector. |
| sourceFilters[0].selector.source | string | `"kyverno"` | select PolicyReport by source |
| sourceFilters[0].uncontrolledOnly | bool | `true` | Filter out PolicyReports of controlled Pods and Jobs, only works for PolicyReport with scope resource |
| sourceFilters[0].disableClusterReports | bool | `false` | Filter out ClusterPolicyReports |
| sourceFilters[0].kinds | object | `{"exclude":["ReplicaSet"]}` | Filter out PolicyReports based on the scope resource kind |
| global.labels | object | `{}` |  |
| basicAuth.username | string | `""` |  |
| basicAuth.password | string | `""` |  |
| basicAuth.secretRef | string | `""` |  |
| emailReports.clusterName | string | `""` |  |
| emailReports.titlePrefix | string | `"Report"` |  |
| emailReports.resources | object | `{}` |  |
| emailReports.smtp.secret | string | `""` |  |
| emailReports.smtp.host | string | `""` |  |
| emailReports.smtp.port | int | `465` |  |
| emailReports.smtp.username | string | `""` |  |
| emailReports.smtp.password | string | `""` |  |
| emailReports.smtp.from | string | `""` |  |
| emailReports.smtp.encryption | string | `""` |  |
| emailReports.smtp.skipTLS | bool | `false` |  |
| emailReports.smtp.certificate | string | `""` |  |
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
| target.loki.minimumSeverity | string | `""` |  |
| target.loki.sources | list | `[]` |  |
| target.loki.skipExistingOnStartup | bool | `true` |  |
| target.loki.customFields | object | `{}` |  |
| target.loki.headers | object | `{}` |  |
| target.loki.username | string | `""` |  |
| target.loki.password | string | `""` |  |
| target.loki.filter | object | `{}` |  |
| target.loki.channels | list | `[]` |  |
| target.elasticsearch.host | string | `""` |  |
| target.elasticsearch.certificate | string | `""` |  |
| target.elasticsearch.skipTLS | bool | `false` |  |
| target.elasticsearch.index | string | `"policy-reporter"` |  |
| target.elasticsearch.username | string | `""` |  |
| target.elasticsearch.password | string | `""` |  |
| target.elasticsearch.apiKey | string | `""` |  |
| target.elasticsearch.secretRef | string | `""` |  |
| target.elasticsearch.mountedSecret | string | `""` |  |
| target.elasticsearch.rotation | string | `"daily"` |  |
| target.elasticsearch.minimumSeverity | string | `""` |  |
| target.elasticsearch.sources | list | `[]` |  |
| target.elasticsearch.skipExistingOnStartup | bool | `true` |  |
| target.elasticsearch.typelessApi | bool | `false` |  |
| target.elasticsearch.customFields | object | `{}` |  |
| target.elasticsearch.filter | object | `{}` |  |
| target.elasticsearch.channels | list | `[]` |  |
| target.slack.webhook | string | `""` |  |
| target.slack.channel | string | `""` |  |
| target.slack.secretRef | string | `""` |  |
| target.slack.mountedSecret | string | `""` |  |
| target.slack.minimumSeverity | string | `""` |  |
| target.slack.sources | list | `[]` |  |
| target.slack.skipExistingOnStartup | bool | `true` |  |
| target.slack.customFields | object | `{}` |  |
| target.slack.filter | object | `{}` |  |
| target.slack.channels | list | `[]` |  |
| target.discord.webhook | string | `""` |  |
| target.discord.secretRef | string | `""` |  |
| target.discord.mountedSecret | string | `""` |  |
| target.discord.minimumSeverity | string | `""` |  |
| target.discord.sources | list | `[]` |  |
| target.discord.skipExistingOnStartup | bool | `true` |  |
| target.discord.filter | object | `{}` |  |
| target.discord.channels | list | `[]` |  |
| target.teams.webhook | string | `""` |  |
| target.teams.secretRef | string | `""` |  |
| target.teams.mountedSecret | string | `""` |  |
| target.teams.certificate | string | `""` |  |
| target.teams.skipTLS | bool | `false` |  |
| target.teams.minimumSeverity | string | `""` |  |
| target.teams.sources | list | `[]` |  |
| target.teams.skipExistingOnStartup | bool | `true` |  |
| target.teams.filter | object | `{}` |  |
| target.teams.channels | list | `[]` |  |
| target.webhook.host | string | `""` |  |
| target.webhook.certificate | string | `""` |  |
| target.webhook.skipTLS | bool | `false` |  |
| target.webhook.secretRef | string | `""` |  |
| target.webhook.mountedSecret | string | `""` |  |
| target.webhook.headers | object | `{}` |  |
| target.webhook.minimumSeverity | string | `""` |  |
| target.webhook.sources | list | `[]` |  |
| target.webhook.skipExistingOnStartup | bool | `true` |  |
| target.webhook.customFields | object | `{}` |  |
| target.webhook.filter | object | `{}` |  |
| target.webhook.channels | list | `[]` |  |
| target.telegram.token | string | `""` |  |
| target.telegram.chatId | string | `""` |  |
| target.telegram.host | string | `""` |  |
| target.telegram.certificate | string | `""` |  |
| target.telegram.skipTLS | bool | `false` |  |
| target.telegram.secretRef | string | `""` |  |
| target.telegram.mountedSecret | string | `""` |  |
| target.telegram.headers | object | `{}` |  |
| target.telegram.minimumSeverity | string | `""` |  |
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
| target.googleChat.minimumSeverity | string | `""` |  |
| target.googleChat.sources | list | `[]` |  |
| target.googleChat.skipExistingOnStartup | bool | `true` |  |
| target.googleChat.customFields | object | `{}` |  |
| target.googleChat.filter | object | `{}` |  |
| target.googleChat.channels | list | `[]` |  |
| target.s3.accessKeyId | string | `""` |  |
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
| target.s3.minimumSeverity | string | `""` |  |
| target.s3.sources | list | `[]` |  |
| target.s3.skipExistingOnStartup | bool | `true` |  |
| target.s3.customFields | object | `{}` |  |
| target.s3.filter | object | `{}` |  |
| target.s3.channels | list | `[]` |  |
| target.kinesis.accessKeyId | string | `""` |  |
| target.kinesis.secretAccessKey | string | `""` |  |
| target.kinesis.secretRef | string | `""` |  |
| target.kinesis.mountedSecret | string | `""` |  |
| target.kinesis.region | string | `""` |  |
| target.kinesis.endpoint | string | `""` |  |
| target.kinesis.streamName | string | `""` |  |
| target.kinesis.minimumSeverity | string | `""` |  |
| target.kinesis.sources | list | `[]` |  |
| target.kinesis.skipExistingOnStartup | bool | `true` |  |
| target.kinesis.customFields | object | `{}` |  |
| target.kinesis.filter | object | `{}` |  |
| target.kinesis.channels | list | `[]` |  |
| target.securityHub.accessKeyId | string | `""` |  |
| target.securityHub.secretAccessKey | string | `""` |  |
| target.securityHub.secretRef | string | `""` |  |
| target.securityHub.mountedSecret | string | `""` |  |
| target.securityHub.region | string | `""` |  |
| target.securityHub.endpoint | string | `""` |  |
| target.securityHub.accountId | string | `""` |  |
| target.securityHub.productName | string | `""` |  |
| target.securityHub.companyName | string | `""` |  |
| target.securityHub.minimumSeverity | string | `""` |  |
| target.securityHub.sources | list | `[]` |  |
| target.securityHub.skipExistingOnStartup | bool | `false` |  |
| target.securityHub.synchronize | bool | `true` |  |
| target.securityHub.delayInSeconds | int | `2` |  |
| target.securityHub.customFields | object | `{}` |  |
| target.securityHub.filter | object | `{}` |  |
| target.securityHub.channels | list | `[]` |  |
| target.gcs.credentials | string | `""` |  |
| target.gcs.secretRef | string | `""` |  |
| target.gcs.mountedSecret | string | `""` |  |
| target.gcs.bucket | string | `""` |  |
| target.gcs.minimumSeverity | string | `""` |  |
| target.gcs.sources | list | `[]` |  |
| target.gcs.skipExistingOnStartup | bool | `true` |  |
| target.gcs.customFields | object | `{}` |  |
| target.gcs.filter | object | `{}` |  |
| target.gcs.channels | list | `[]` |  |
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
| database.type | string | `""` |  |
| database.database | string | `""` |  |
| database.username | string | `""` |  |
| database.password | string | `""` |  |
| database.host | string | `""` |  |
| database.enableSSL | bool | `false` |  |
| database.dsn | string | `""` |  |
| database.secretRef | string | `""` |  |
| database.mountedSecret | string | `""` |  |
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
| envVars | list | `[]` | Allow additional env variables to be added |
| tmpVolume | object | `{}` | Allow custom configuration of the /tmp volume |
| ui.enabled | bool | `false` |  |
| ui.image.registry | string | `"ghcr.io"` | Image registry |
| ui.image.repository | string | `"kyverno/policy-reporter-ui"` | Image repository |
| ui.image.pullPolicy | string | `"IfNotPresent"` | Image PullPolicy |
| ui.image.tag | string | `"2.0.0-beta.14"` | Image tag Defaults to `Chart.AppVersion` if omitted |
| ui.replicaCount | int | `1` | Deployment replica count |
| ui.tempDir | string | `"/tmp"` | Temporary Directory to persist session data for authentication |
| ui.logging.encoding | string | `"console"` | log encoding possible encodings are console and json |
| ui.logging.logLevel | int | `0` | log level default info |
| ui.server.port | int | `8080` | Application port |
| ui.server.logging | bool | `false` | Enables Access logging |
| ui.server.overwriteHost | bool | `true` | Overwrites Request Host with Proxy Host and adds `X-Forwarded-Host` and `X-Origin-Host` headers |
| ui.openIDConnect.enabled | bool | `false` | Enable openID Connect authentication |
| ui.openIDConnect.discoveryUrl | string | `""` | OpenID Connect Discovery URL |
| ui.openIDConnect.callbackUrl | string | `""` | OpenID Connect Callback URL |
| ui.openIDConnect.clientId | string | `""` | OpenID Connect ClientID |
| ui.openIDConnect.clientSecret | string | `""` | OpenID Connect ClientSecret |
| ui.openIDConnect.scopes | list | `[]` | OpenID Connect allowed Scopes |
| ui.openIDConnect.secretRef | string | `""` | Provide OpenID Connect configuration via Secret supported keys: `discoveryUrl`, `clientId`, `clientSecret` |
| ui.oauth.enabled | bool | `false` | Enable openID Connect authentication |
| ui.oauth.provider | string | `""` | OAuth2 Provider supported: amazon, gitlab, github, apple, google, yandex, azuread |
| ui.oauth.callbackUrl | string | `""` | OpenID Connect Callback URL |
| ui.oauth.clientId | string | `""` | OpenID Connect ClientID |
| ui.oauth.clientSecret | string | `""` | OpenID Connect ClientSecret |
| ui.oauth.scopes | list | `[]` | OpenID Connect allowed Scopes |
| ui.oauth.secretRef | string | `""` | Provide OpenID Connect configuration via Secret supported keys: `provider`, `clientId`, `clientSecret` |
| ui.banner | string | `""` | optional banner text |
| ui.displayMode | string | `""` | DisplayMode dark/light/colorblind/colorblinddark uses the OS configured prefered color scheme as default |
| ui.customBoards | list | `[]` | Additional customizable dashboards |
| ui.sources | list | `[]` | source specific configurations |
| ui.name | string | `"Default"` |  |
| ui.clusters | list | `[]` | Connected Policy Reporter APIs |
| ui.imagePullSecrets | list | `[]` | Image pull secrets for image verification policies, this will define the `--imagePullSecrets` argument |
| ui.serviceAccount.create | bool | `true` | Create ServiceAccount |
| ui.serviceAccount.automount | bool | `true` | Enable ServiceAccount automaount |
| ui.serviceAccount.annotations | object | `{}` | Annotations for the ServiceAccount |
| ui.serviceAccount.name | string | `""` | The ServiceAccount name |
| ui.extraManifests | list | `[]` | list of extra manifests |
| ui.sidecarContainers | object | `{}` | Add sidecar containers to the UI deployment  sidecarContainers:    oauth-proxy:      image: quay.io/oauth2-proxy/oauth2-proxy:v7.6.0      args:      - --upstream=http://127.0.0.1:8080      - --http-address=0.0.0.0:8081      - ...      ports:      - containerPort: 8081        name: oauth-proxy        protocol: TCP      resources: {} |
| ui.podAnnotations | object | `{}` | Additional annotations to add to each pod |
| ui.podLabels | object | `{}` | Additional labels to add to each pod |
| ui.updateStrategy | object | `{}` | Deployment update strategy. Ref: https://kubernetes.io/docs/concepts/workloads/controllers/deployment/#strategy |
| ui.revisionHistoryLimit | int | `10` | The number of revisions to keep |
| ui.podSecurityContext | object | `{"runAsGroup":1234,"runAsUser":1234}` | Security context for the pod |
| ui.envVars | list | `[]` | Allow additional env variables to be added |
| ui.rbac.enabled | bool | `true` | Create RBAC resources |
| ui.securityContext | object | `{"allowPrivilegeEscalation":false,"capabilities":{"drop":["ALL"]},"privileged":false,"readOnlyRootFilesystem":true,"runAsNonRoot":true,"runAsUser":1234,"seccompProfile":{"type":"RuntimeDefault"}}` | Container security context |
| ui.service.type | string | `"ClusterIP"` | Service type. |
| ui.service.port | int | `8080` | Service port. |
| ui.service.annotations | object | `{}` | Service annotations. |
| ui.service.labels | object | `{}` | Service labels. |
| ui.service.additionalPorts | list | `[]` | Additional service ports for e.g. Sidecars  # - name: authenticated additionalPorts: - name: authenticated   port: 8081   targetPort: 8081 |
| ui.ingress.enabled | bool | `false` | Create ingress resource. |
| ui.ingress.port | string | `nil` | Redirect ingress to an additional defined port on the service |
| ui.ingress.className | string | `""` | Ingress class name. |
| ui.ingress.labels | object | `{}` | Ingress labels. |
| ui.ingress.annotations | object | `{}` | Ingress annotations. |
| ui.ingress.hosts | list | `[]` | List of ingress host configurations. |
| ui.ingress.tls | list | `[]` | List of ingress TLS configurations. |
| ui.networkPolicy.enabled | bool | `false` | When true, use a NetworkPolicy to allow ingress to the webhook This is useful on clusters using Calico and/or native k8s network policies in a default-deny setup. |
| ui.networkPolicy.egress | list | `[{"ports":[{"port":6443,"protocol":"TCP"}]}]` | A list of valid from selectors according to https://kubernetes.io/docs/concepts/services-networking/network-policies. Enables Kubernetes API Server by default |
| ui.networkPolicy.ingress | list | `[]` | A list of valid from selectors according to https://kubernetes.io/docs/concepts/services-networking/network-policies. |
| ui.resources | object | `{}` |  |
| ui.podDisruptionBudget.minAvailable | int | `1` | Configures the minimum available pods for kyvernoPlugin disruptions. Cannot be used if `maxUnavailable` is set. |
| ui.podDisruptionBudget.maxUnavailable | string | `nil` | Configures the maximum unavailable pods for kyvernoPlugin disruptions. Cannot be used if `minAvailable` is set. |
| ui.nodeSelector | object | `{}` | Node labels for pod assignment |
| ui.tolerations | list | `[]` | List of node taints to tolerate |
| ui.affinity | object | `{}` | Affinity constraints. |
| plugin.kyverno.enabled | bool | `false` | Enable Kyverno Plugin |
| plugin.kyverno.image.registry | string | `"ghcr.io"` | Image registry |
| plugin.kyverno.image.repository | string | `"kyverno/policy-reporter/kyverno-plugin"` | Image repository |
| plugin.kyverno.image.pullPolicy | string | `"IfNotPresent"` | Image PullPolicy |
| plugin.kyverno.image.tag | string | `"0.2.0"` | Image tag Defaults to `Chart.AppVersion` if omitted |
| plugin.kyverno.replicaCount | int | `1` | Deployment replica count |
| plugin.kyverno.logging.encoding | string | `"console"` | log encoding possible encodings are console and json |
| plugin.kyverno.logging.logLevel | int | `0` | log level default info |
| plugin.kyverno.server.port | int | `8080` | Application port |
| plugin.kyverno.server.logging | bool | `false` | Enables Access logging |
| plugin.kyverno.blockReports.enabled | bool | `false` | Enables he BlockReport feature |
| plugin.kyverno.blockReports.eventNamespace | string | `"default"` | Watches for Kyverno Events in the configured namespace leave blank to watch in all namespaces |
| plugin.kyverno.blockReports.results.maxPerReport | int | `200` | Max items per PolicyReport resource |
| plugin.kyverno.blockReports.results.keepOnlyLatest | bool | `false` | Keep only the latest of duplicated events |
| plugin.kyverno.imagePullSecrets | list | `[]` | Image pull secrets for image verification policies, this will define the `--imagePullSecrets` argument |
| plugin.kyverno.serviceAccount.create | bool | `true` | Create ServiceAccount |
| plugin.kyverno.serviceAccount.automount | bool | `true` | Enable ServiceAccount automaount |
| plugin.kyverno.serviceAccount.annotations | object | `{}` | Annotations for the ServiceAccount |
| plugin.kyverno.serviceAccount.name | string | `""` | The ServiceAccount name |
| plugin.kyverno.podAnnotations | object | `{}` | Additional annotations to add to each pod |
| plugin.kyverno.podLabels | object | `{}` | Additional labels to add to each pod |
| plugin.kyverno.updateStrategy | object | `{}` | Deployment update strategy. Ref: https://kubernetes.io/docs/concepts/workloads/controllers/deployment/#strategy |
| plugin.kyverno.revisionHistoryLimit | int | `10` | The number of revisions to keep |
| plugin.kyverno.podSecurityContext | object | `{"runAsGroup":1234,"runAsUser":1234}` | Security context for the pod |
| plugin.kyverno.envVars | list | `[]` | Allow additional env variables to be added |
| plugin.kyverno.rbac.enabled | bool | `true` | Create RBAC resources |
| plugin.kyverno.securityContext | object | `{"allowPrivilegeEscalation":false,"capabilities":{"drop":["ALL"]},"privileged":false,"readOnlyRootFilesystem":true,"runAsNonRoot":true,"runAsUser":1234,"seccompProfile":{"type":"RuntimeDefault"}}` | Container security context |
| plugin.kyverno.service.type | string | `"ClusterIP"` | Service type. |
| plugin.kyverno.service.port | int | `8080` | Service port. |
| plugin.kyverno.service.annotations | object | `{}` | Service annotations. |
| plugin.kyverno.service.labels | object | `{}` | Service labels. |
| plugin.kyverno.ingress.enabled | bool | `false` | Create ingress resource. |
| plugin.kyverno.ingress.className | string | `""` | Ingress class name. |
| plugin.kyverno.ingress.labels | object | `{}` | Ingress labels. |
| plugin.kyverno.ingress.annotations | object | `{}` | Ingress annotations. |
| plugin.kyverno.ingress.hosts | list | `[]` | List of ingress host configurations. |
| plugin.kyverno.ingress.tls | list | `[]` | List of ingress TLS configurations. |
| plugin.kyverno.networkPolicy.enabled | bool | `false` | When true, use a NetworkPolicy to allow ingress to the webhook This is useful on clusters using Calico and/or native k8s network policies in a default-deny setup. |
| plugin.kyverno.networkPolicy.egress | list | `[{"ports":[{"port":6443,"protocol":"TCP"}]}]` | A list of valid from selectors according to https://kubernetes.io/docs/concepts/services-networking/network-policies. Enables Kubernetes API Server by default |
| plugin.kyverno.networkPolicy.ingress | list | `[]` | A list of valid from selectors according to https://kubernetes.io/docs/concepts/services-networking/network-policies. |
| plugin.kyverno.resources | object | `{}` |  |
| plugin.kyverno.leaderElection.lockName | string | `"kyverno-plugin"` | Lock Name |
| plugin.kyverno.leaderElection.releaseOnCancel | bool | `true` | Released lock when the run context is cancelled. |
| plugin.kyverno.leaderElection.leaseDuration | int | `15` | LeaseDuration is the duration that non-leader candidates will wait to force acquire leadership. |
| plugin.kyverno.leaderElection.renewDeadline | int | `10` | RenewDeadline is the duration that the acting master will retry refreshing leadership before giving up. |
| plugin.kyverno.leaderElection.retryPeriod | int | `2` | RetryPeriod is the duration the LeaderElector clients should wait between tries of actions. |
| plugin.kyverno.podDisruptionBudget.minAvailable | int | `1` | Configures the minimum available pods for kyvernoPlugin disruptions. Cannot be used if `maxUnavailable` is set. |
| plugin.kyverno.podDisruptionBudget.maxUnavailable | string | `nil` | Configures the maximum unavailable pods for kyvernoPlugin disruptions. Cannot be used if `minAvailable` is set. |
| plugin.kyverno.nodeSelector | object | `{}` | Node labels for pod assignment |
| plugin.kyverno.tolerations | list | `[]` | List of node taints to tolerate |
| plugin.kyverno.affinity | object | `{}` | Affinity constraints. |
| plugin.trivy.enabled | bool | `false` | Enable Trivy Operator Plugin |
| plugin.trivy.image.registry | string | `"ghcr.io"` | Image registry |
| plugin.trivy.image.repository | string | `"kyverno/policy-reporter/trivy-plugin"` | Image repository |
| plugin.trivy.image.pullPolicy | string | `"IfNotPresent"` | Image PullPolicy |
| plugin.trivy.image.tag | string | `"0.0.3"` | Image tag Defaults to `Chart.AppVersion` if omitted |
| plugin.trivy.replicaCount | int | `1` | Deployment replica count |
| plugin.trivy.logging.encoding | string | `"console"` | log encoding possible encodings are console and json |
| plugin.trivy.logging.logLevel | int | `0` | log level default info |
| plugin.trivy.server.port | int | `8080` | Application port |
| plugin.trivy.server.logging | bool | `false` | Enables Access logging |
| plugin.trivy.policyReporter.skipTLS | bool | `false` | Skip TLS Verification |
| plugin.trivy.policyReporter.certificate | string | `""` | TLS Certificate |
| plugin.trivy.policyReporter.secretRef | string | `""` | Secret to read the API configuration from supports `host`, `certificate`, `skipTLS`, `username`, `password` key |
| plugin.trivy.imagePullSecrets | list | `[]` | Image pull secrets for image verification policies, this will define the `--imagePullSecrets` argument |
| plugin.trivy.serviceAccount.create | bool | `true` | Create ServiceAccount |
| plugin.trivy.serviceAccount.automount | bool | `true` | Enable ServiceAccount automaount |
| plugin.trivy.serviceAccount.annotations | object | `{}` | Annotations for the ServiceAccount |
| plugin.trivy.serviceAccount.name | string | `""` | The ServiceAccount name |
| plugin.trivy.podAnnotations | object | `{}` | Additional annotations to add to each pod |
| plugin.trivy.podLabels | object | `{}` | Additional labels to add to each pod |
| plugin.trivy.updateStrategy | object | `{}` | Deployment update strategy. Ref: https://kubernetes.io/docs/concepts/workloads/controllers/deployment/#strategy |
| plugin.trivy.revisionHistoryLimit | int | `10` | The number of revisions to keep |
| plugin.trivy.podSecurityContext | object | `{"runAsGroup":1234,"runAsUser":1234}` | Security context for the pod |
| plugin.trivy.envVars | list | `[]` | Allow additional env variables to be added |
| plugin.trivy.rbac.enabled | bool | `true` | Create RBAC resources |
| plugin.trivy.securityContext | object | `{"allowPrivilegeEscalation":false,"capabilities":{"drop":["ALL"]},"privileged":false,"readOnlyRootFilesystem":true,"runAsNonRoot":true,"runAsUser":1234,"seccompProfile":{"type":"RuntimeDefault"}}` | Container security context |
| plugin.trivy.service.type | string | `"ClusterIP"` | Service type. |
| plugin.trivy.service.port | int | `8080` | Service port. |
| plugin.trivy.service.annotations | object | `{}` | Service annotations. |
| plugin.trivy.service.labels | object | `{}` | Service labels. |
| plugin.trivy.ingress.enabled | bool | `false` | Create ingress resource. |
| plugin.trivy.ingress.className | string | `""` | Ingress class name. |
| plugin.trivy.ingress.labels | object | `{}` | Ingress labels. |
| plugin.trivy.ingress.annotations | object | `{}` | Ingress annotations. |
| plugin.trivy.ingress.hosts | list | `[]` | List of ingress host configurations. |
| plugin.trivy.ingress.tls | list | `[]` | List of ingress TLS configurations. |
| plugin.trivy.networkPolicy.enabled | bool | `false` | When true, use a NetworkPolicy to allow ingress to the webhook This is useful on clusters using Calico and/or native k8s network policies in a default-deny setup. |
| plugin.trivy.networkPolicy.egress | list | `[{"ports":[{"port":6443,"protocol":"TCP"}]}]` | A list of valid from selectors according to https://kubernetes.io/docs/concepts/services-networking/network-policies. Enables Kubernetes API Server by default |
| plugin.trivy.networkPolicy.ingress | list | `[]` | A list of valid from selectors according to https://kubernetes.io/docs/concepts/services-networking/network-policies. |
| plugin.trivy.resources | object | `{}` |  |
| plugin.trivy.podDisruptionBudget.minAvailable | int | `1` | Configures the minimum available pods for kyvernoPlugin disruptions. Cannot be used if `maxUnavailable` is set. |
| plugin.trivy.podDisruptionBudget.maxUnavailable | string | `nil` | Configures the maximum unavailable pods for kyvernoPlugin disruptions. Cannot be used if `minAvailable` is set. |
| plugin.trivy.nodeSelector | object | `{}` | Node labels for pod assignment |
| plugin.trivy.tolerations | list | `[]` | List of node taints to tolerate |
| plugin.trivy.affinity | object | `{}` | Affinity constraints. |
| monitoring.enabled | bool | `false` |  |
| monitoring.annotations | object | `{}` |  |
| monitoring.serviceMonitor.honorLabels | bool | `false` |  |
| monitoring.serviceMonitor.namespace | string | `nil` |  |
| monitoring.serviceMonitor.labels | object | `{}` |  |
| monitoring.serviceMonitor.relabelings | list | `[]` |  |
| monitoring.serviceMonitor.metricRelabelings | list | `[]` |  |
| monitoring.serviceMonitor.namespaceSelector | object | `{}` |  |
| monitoring.serviceMonitor.scrapeTimeout | string | `nil` |  |
| monitoring.serviceMonitor.interval | string | `nil` |  |
| monitoring.grafana.namespace | string | `nil` |  |
| monitoring.grafana.dashboards.enabled | bool | `true` |  |
| monitoring.grafana.dashboards.label | string | `"grafana_dashboard"` |  |
| monitoring.grafana.dashboards.value | string | `"1"` |  |
| monitoring.grafana.dashboards.labelFilter | list | `[]` |  |
| monitoring.grafana.dashboards.multicluster.enabled | bool | `false` |  |
| monitoring.grafana.dashboards.multicluster.label | string | `"cluster"` |  |
| monitoring.grafana.dashboards.enable.overview | bool | `true` |  |
| monitoring.grafana.dashboards.enable.policyReportDetails | bool | `true` |  |
| monitoring.grafana.dashboards.enable.clusterPolicyReportDetails | bool | `true` |  |
| monitoring.grafana.folder.annotation | string | `"grafana_folder"` |  |
| monitoring.grafana.folder.name | string | `"Policy Reporter"` |  |
| monitoring.grafana.datasource.label | string | `"Prometheus"` |  |
| monitoring.grafana.datasource.pluginId | string | `"prometheus"` |  |
| monitoring.grafana.datasource.pluginName | string | `"Prometheus"` |  |
| monitoring.grafana.grafanaDashboard | object | `{"allowCrossNamespaceImport":true,"enabled":false,"folder":"kyverno","matchLabels":{"dashboards":"grafana"}}` | create GrafanaDashboard custom resource referencing to the configMap. according to https://grafana-operator.github.io/grafana-operator/docs/examples/dashboard_from_configmap/readme/ |
| monitoring.policyReportDetails.firstStatusRow.height | int | `8` |  |
| monitoring.policyReportDetails.secondStatusRow.enabled | bool | `true` |  |
| monitoring.policyReportDetails.secondStatusRow.height | int | `2` |  |
| monitoring.policyReportDetails.statusTimeline.enabled | bool | `true` |  |
| monitoring.policyReportDetails.statusTimeline.height | int | `8` |  |
| monitoring.policyReportDetails.passTable.enabled | bool | `true` |  |
| monitoring.policyReportDetails.passTable.height | int | `8` |  |
| monitoring.policyReportDetails.failTable.enabled | bool | `true` |  |
| monitoring.policyReportDetails.failTable.height | int | `8` |  |
| monitoring.policyReportDetails.warningTable.enabled | bool | `true` |  |
| monitoring.policyReportDetails.warningTable.height | int | `4` |  |
| monitoring.policyReportDetails.errorTable.enabled | bool | `true` |  |
| monitoring.policyReportDetails.errorTable.height | int | `4` |  |
| monitoring.clusterPolicyReportDetails.statusRow.height | int | `6` |  |
| monitoring.clusterPolicyReportDetails.statusTimeline.enabled | bool | `true` |  |
| monitoring.clusterPolicyReportDetails.statusTimeline.height | int | `8` |  |
| monitoring.clusterPolicyReportDetails.passTable.enabled | bool | `true` |  |
| monitoring.clusterPolicyReportDetails.passTable.height | int | `8` |  |
| monitoring.clusterPolicyReportDetails.failTable.enabled | bool | `true` |  |
| monitoring.clusterPolicyReportDetails.failTable.height | int | `8` |  |
| monitoring.clusterPolicyReportDetails.warningTable.enabled | bool | `true` |  |
| monitoring.clusterPolicyReportDetails.warningTable.height | int | `4` |  |
| monitoring.clusterPolicyReportDetails.errorTable.enabled | bool | `true` |  |
| monitoring.clusterPolicyReportDetails.errorTable.height | int | `4` |  |
| monitoring.policyReportOverview.failingSummaryRow.height | int | `8` |  |
| monitoring.policyReportOverview.failingTimeline.height | int | `10` |  |
| monitoring.policyReportOverview.failingPolicyRuleTable.height | int | `10` |  |
| monitoring.policyReportOverview.failingClusterPolicyRuleTable.height | int | `10` |  |

## Source Code

* <https://github.com/kyverno/policy-reporter>

## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| Frank Jogeleit |  |  |

----------------------------------------------
Autogenerated from chart metadata using [helm-docs v1.11.0](https://github.com/norwoodj/helm-docs/releases/v1.11.0)
