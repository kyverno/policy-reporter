# Changelog

# 2.18.1

* Signed Helm Chart
* Policy Reporter
    * New `channel` property for Slack targets to define the Slack channel to send the results too
    * New `mountedSecret` property to read target configs from a mounted secret [[#282](https://github.com/kyverno/policy-reporter/pull/282) by [rromic](https://github.com/rromic)]
    * AWS KMS support for S3 target with new properties `bucketKeyEnabled`, `kmsKeyId` and `serverSideEncryption` [[#282](https://github.com/kyverno/policy-reporter/pull/282) by [rromic](https://github.com/rromic)]
        * Mountet secret needs to be in json format with keys defined in kubernetes/secrets Values struct.

* Monitoring
    * Add `namespaceSelector` to `serviceMonitor` values

# 2.18.0
* Policy Reporter
    * Improved logging configuration
        * Support JSON logging
        * Support log level
    * optional API access logging with `api.logging` set to `true`
    * New aggregation table for API performance improvements
    * Helm Ingress template
    * New Google Cloud Storage Target
        * Requires `credentials` as JSON String and the `bucket` name
        * Added in the helm valus under `target.gcs`
* Policy Reporter KyvernoPlugin
    * Helm Ingress template
    * Improved logging configuration
        * Support JSON logging
        * Support log level
* Policy Reporter UI
    * Improved logging configuration
        * Support JSON logging
        * Support log level
        * Proxy Logging

# 2.17.0
* Policy Reporter
    * Use metaclient to reduce informer memory usage
    * Use workerqueue to control concurrent processing of PolicyReports
    * Remove internal PolicyReport structures
    * Make sqlite volume configurable [[#255](https://github.com/kyverno/policy-reporter/pull/255) by [monotek](https://github.com/monotek)]
    * use defer to unlock when possible [[#259](https://github.com/kyverno/policy-reporter/pull/259) by [eddycharly](https://github.com/eddycharly)]
* Policy Reporter UI
    * New SSL configs for external clusters
        * `skipTLS` to disable SSL verification
        * `certificate` to configure a path to a custom CA for self signed URLs
    * New Helm values `ui.volumes` and `ui.volumeMounts` to add your custom CAs as mounts to the UI deployment.

# 2.16.0
* Add `nameOverride` to all charts [[#254](https://github.com/kyverno/policy-reporter/pull/254) by [mjnagel](https://github.com/mjnagel)]

# 2.15.0

* Add values to configure `topologySpreadConstraints` for all components [[#241](https://github.com/kyverno/policy-reporter/pull/241) by [Kostavro](https://github.com/Kostavro)]
* Fixing comment formats and deprecations [[#250](https://github.com/kyverno/policy-reporter/pull/250) by [fengshunli](https://github.com/fengshunli)]
* Add new APIs for PolicyReport and ClusterPolicyReport metadata (`/v1/policy-reports`, `/v1/cluster-policy-reports`) [[#251](https://github.com/kyverno/policy-reporter/pull/251)
* `search` filter also checks the resource kind
* Use correct probes in core deployment [[#236](https://github.com/kyverno/policy-reporter/pull/236) by [rgarcia89](https://github.com/rgarcia89)]
* Add source to PolicyReport Table and improve report-label API [[#252](https://github.com/kyverno/policy-reporter/pull/252)

# 2.14.1

* Policy Reporter
    * Fix generate multiple custom metrics

# 2.14.0
* Policy Reporter
    * Persist also PolicyReport labels
    * API
        * New API to get available labels for PolicyReports: `/v1/namespaced-resources/report-labels`
        * New API to get available labels for ClusterPolicyReports: `/v1/cluster-resources/report-labels`
    * Metrics
        * special syntax to add report labels to metric labels: `label:report-label-name`, special characters like `-`, `/`, `.`, `:` will be transformed to `_` in metrics
    * New Target Filter `reportLabel` to, filter results based on labels of the related (Cluster)PolicyReport
* Monitoring
    * New values to disable dedicated Grafana Dashboards:
        - `grafana.dashboards.enable.overview`, default `true`
        - `grafana.dashboards.enable.policyReportDetails`, default `true`
        - `grafana.dashboards.enable.clusterPolicyReportDetails`, default `true`
    * New values to configure the Grafana Dashboard datasource label, pluginName, pluginId
        - `grafana.datasource.label`, default `Prometheus`
        - `grafana.datasource.pluginName`, default `Prometheus`
        - `grafana.datasource.pluginId`, default `prometheus`
    * New value `grafana.dashboards.labelFilter` to add custom report labels as dashboard filter, default `[]`. Label has to be a valid prometheus label, e.g. `created-by` => `created_by`.
    * New values `grafana.dashboards.multicluster.enabled` and `grafana.dashboards.multicluster.label` to add an optional `cluster` label.
* Kyverno Plugin
    * New HTML Compliance Reports
        * Grouped by Policy with Details per Namespace and Resource
        * Grouped by Namespace with Details per Policy and Resource
    * Go update to 1.19
* UI
    * Integrate new Compliance Reports
    * New PolicyReport label based filter, use `ui.labelFilter` to define a list of labels to add
    * Go update to 1.19

# 2.13.5
* Add configuration `target.s3.pathStyle` for the `S3` output

# 2.13.4
* Fix `customFields` mapping in TargetFactory

# 2.13.3
* Fix `customFields` property in values.yaml
* Fix PolicyReporter `image.tag` version

# 2.13.2
* Policy Reporter
    * Add `customFields` property to missing targets: `Elasticsearch`, `S3`, `Webhook`, `Kinesis`
* Policy Reporter UI
    * Create Links out of URL property values
* Monitoring
    * New `monitoring.serviceMonitor.honorLabels` and `monitoring.kyverno.serviceMonitor.honorLabels` value: chooses the metrics labels on collisions with target labels [[#216](https://github.com/kyverno/policy-reporter/pull/216) by [monotek](https://github.com/monotek)]

# 2.13.1
* Policy Reporter
    * Fix persist error for duplicated IDs
    * Disable UI SA automount

# 2.13.0
* Policy Reporter
    * New `certificate` config for `loki`, `elasticsearch`, `teams`, `webhook` and `ui`, to set the path to your custom certificate for the related client.
    * New `skipTLS` config for `loki`, `elasticsearch`, `teams`, `webhook` and `ui`, to skip tls if needed for the given target.
    * New `secretRef` for targets to reference a secret with the related `username`, `password`, `webhook`, `host`, `accessKeyID`, `secretAccessKey` information of the given target, instead of configure your credentials directly.
* Policy Reporter UI
    * New value `refreshInterval` to configure the default refresh interval for API polling. Set `0` to disable polling.
* Policy Reporter Kyverno Plugin
    * Fix the creation of duplicated results for PolicyReportResults.


# 2.12.0
* Policy Reporter
    * New Helm Chart value to add extra volumes to PolicyReporter deployment [[#186](https://github.com/kyverno/policy-reporter/pull/186) by [preved911](https://github.com/preved911)]
    * HTTP Basic authentication for Elasticsearch targets with `username` and `password` configuration fields
    * `target.slack.customFields` map property for Slack pushes to add additional metadata to notifications like clustername
    * Add timestamp to Result REST APIs
    * Overwrite the installation target namespace via the new `global.namespace` value.

# 2.11.3
* Policy Reporter
    * New `emailReports.smtp.secret` configuration to use an existing external secret to configure your SMTP connection
        * You can set all or a subset of the available keys in your secret: `host`, `port`, `username`, `password`, `from`, `encryption`
        * Keys available in your secret have a higher priority as your Helm release values.

# 2.11.2
* Policy Reporter
    * Add new Severity values `info` and `critical`
    * Update PolicyReport ID generierung
* Policy Reporter UI
    * Fix Grouping by Policy and Categories
    * Fix ReverseProxy RequestHost
    * New configuration `ui.clusterName` which is used in the ClusterSelect, if you configure additional Clusters
* Policy Reporter Kyverno Plugin
    * Add `time` property to PolicyReportResults

# 2.11.1
* Policy Reporter
    * Fix `CronJob` Resources by [[#157](https://github.com/kyverno/policy-reporter/pull/178) by [MaxRink](https://github.com/MaxRink)]
* Policy Reporter UI
    * Fix API Proxy for APIs behind ReverseProxy (like NGINX Ingress)

# 2.11.0
* Policy Reporter
    * High Availability support with leaderelection for necessary features like target pushes, to avoid duplicated pushes by multiple instances
        * Add new `role` and `rolebinding` to manage lease objects if leaderelection is enabled
    * Add redis configuration to the Helm Chart for external cache storage
    * Add PodDisruptionBudget for HA Deployments (replicaCount > 1)
    * Add `skipTLS` configuration for MS Teams Webhook

* Policy Reporter KyvernoPlugin
    * High Availability support with leaderelection for necessary features like PolicyReport management for blocked resources
        * Add new `role` and `rolebinding` to manage lease objects if leaderelection is enabled
    * Add PodDisruptionBudget for HA Deployments (replicaCount > 1)
    * Internal refactoring for better CRD management

* Policy Reporter UI
    * Add redis as possible log storage to support high availability deployments
    * Add PodDisruptionBudget for HA Deployments (replicaCount > 1)

# 2.10.3
* Policy Reporter
    * Add new config `target.loki.path` to overwrite the deprected prom push API

# 2.10.2
* Policy Reporter UI
    * New option `ui.clusters` makes it possible to configure additional external Policy Reporter APIs (<a href="https://kyverno.github.io/policy-reporter/guide/helm-chart-core#external-clusters" target="_blank">details</a>)
    * General UI improvements for loading state and error handling

# 2.10.1
* Monitoring
    * Fix Datasource for Metrics and Filters in the preconfigured Dashboards
    * Add Datasource as additional Select to the preconfigured Dashboards

# 2.10.0
* Policy Reporter
    * Email Reports
        * Send Summary Reports over SMTP to different E-Mails
        * Supports channels and filters to send different subsets of Namespaces or Sources to dedicated E-Mails
        * Reports are generated and send over dedicated CronJobs, this makes it easy to send the reports as often as needed
        * Currently a basic summary and a more detailed violation report is available and can be separatly enabled and configured
    * Metrics
        * Add `metrics.mode` for less or custom metric values, to reduce cardinality
    * Monitoring
        * Fix Source Column for result tables
        * Fix Warn counter for ClusterPolicyReport Details

# 2.9.5
* Fix Policy Reporter Version in the Helm Chart values.yaml

# 2.9.4
* Policy Reporter
    * Add [AWS Kinesis](https://aws.amazon.com/kinesis) compatible target
    * Add new Helm value `profiling.enabled` to enable pprof profiling, disabled by default
    * Improved Informer handling

# 2.9.3
* Policy Reporter
    * Fix `grafana.dashboards.value` type conversion [[fix #158](https://github.com/kyverno/policy-reporter/issues/158)]

# 2.9.2
* Policy Reporter
    * Add `grafana.dashboards.value` value to configure the ConfigMap label value for the Prometheus Operator by [[#157](https://github.com/kyverno/policy-reporter/pull/157) by [stone-z](https://github.com/stone-z)]

# 2.9.1
* Policy Reporter
    * Name Configuration for Target (Channels) to customize UI Labels
* Policy Reporter UI
    * Fix table on chip selection
    * Order labels
    * Return 404 Status Code for non existing URL paths

# 2.9.0
* Policy Reporter
    * New configuration to use Redis as external result caching store
    * SQLite Improvement: Use batch insertion for PolicyReportResults
    * PolicyReport Informer Update: Use typed informer to improve performance and memory usage
    * Drop support for `v1alpha1` of the PolicyReport CRD
    * Serverside Pagination for better Dashboard performance
    * Concurrent PolicyReport processing
* Policy Reporter UI
    * Serverside Pagination support
    * Dynamic Chart sizes
* Policy Reporter Kyverno Plugin
    * Generate Policy Reports for enforcement violations

# 2.8.0
* Policy Reporter
    * New target filter and channels to define multiple configurations of the same target
        * Filter target results by exclude and include rules for namesapces, priorities and policies
        * Support wildcards for policies and namespaces
    * New __webhook__ target
        * this target is a simple way to send notifications to custom tools and APIs
        * results are send as POST requests with a JSON representation of the result
        * the _headers_ properties allows you to send custom header with the request to allow for example authentication

# 2.7.1
* Policy Reporter
    * Add Resource APIVersion to the Results REST APIs

# 2.7.0
* Policy Reporter
    * PolicyReport Filter:
        * PolicyReporter CRD Filter by Namespaces
        * Disable ClusterPolicyReport CRD processing

# 2.6.3
* Policy Reporter
    * Fix Debouncer has wrong reference to OldPolicyReport when a result was cached.

# 2.6.2
* Policy Reporter
    * Update Go  to 1.17.8
    * Add `serviceMonitor.relabelings` and `serviceMonitor.metricRelabelings` for ServiceMonitor configuration in the `monitoring` Subchart.
    * Add `kyverno.serviceMonitor.relabelings` and `kyverno.serviceMonitor.metricRelabelings` for the KyvernoPlugin ServiceMonitor configuration in the `monitoring` Subchart.

* Policy Reporter UI
    * Update Go  to 1.17.8

* Policy Reporter KyvernoPlugin
    * Update Go  to 1.17.8

# 2.6.1
* Update Policy Reporter UI to v1.3.2
    * Support access over Subpaths, e.g. Rancher Reverse Proxy
* Update Policy Reporter Monitoring to v2.1.0
    * Fix Failing ClusterPolicyRules Columns of the PolicyReports Dashboard
    * Add Filter to the PolicyReports Dashboard

# 2.6.0
* Add seccomp profile support [[#120](https://github.com/kyverno/policy-reporter/pull/120) by [eddycharly](https://github.com/eddycharly)]

# 2.5.0
* New Policy Reporter API to get a list of available resources
* New Filter for Policies, Kinds, Categories and Results APIs

# 2.4.0
* Policy Reporter
    * Add Support for custom Loki labels

# 2.3.0
* Policy Reporter
    * Add Support for linux/s390x [[#115](https://github.com/kyverno/policy-reporter/pull/115) by [skuethe](https://github.com/skuethe)]

* Policy Reporter UI
    * Add Support for linux/s390x [[#98](https://github.com/kyverno/policy-reporter-ui/pull/98) by [skuethe](https://github.com/skuethe)]

* Policy Reporter KyvernoPlugin
    * Add Support for linux/s390x [[#13](https://github.com/kyverno/policy-reporter-kyverno-plugin/pull/13) by [skuethe](https://github.com/skuethe)]

# 2.2.6
* Use upper case on drop capabilities [[#113](https://github.com/kyverno/policy-reporter/pull/113) by [skuethe](https://github.com/skuethe)]

# 2.2.5
* Policy Reporter
    * Update Go  to 1.17.6 [[#110](https://github.com/kyverno/policy-reporter/pull/110) by [realshuting](https://github.com/realshuting)]
    * Update Helm Chart with new component versions
    * Update dependencies

* Policy Reporter UI
    * Update Go  to 1.17.6 [[#93](https://github.com/kyverno/policy-reporter-ui/pull/93) by [realshuting](https://github.com/realshuting)]
    * Update dependencies

* Policy Reporter KyvernoPlugin
    * Update Go  to 1.17.6 [[#12](https://github.com/kyverno/policy-reporter-kyverno-plugin/pull/12) by [realshuting](https://github.com/realshuting)]

# 2.2.4
* Fix PolicyReport Napper - string casting

# 2.2.3
* Fix Helm Chart uihost template function.

# 2.2.2
* Fix Helm Chart `values.yaml`. Cleanup unused default configurations. [[#103](https://github.com/kyverno/policy-reporter/pull/103) by [AndersBennedsgaard](https://github.com/AndersBennedsgaard)]

# 2.2.1
* Fix Typo in values.yaml [[#102](https://github.com/kyverno/policy-reporter-ui/pull/102) by [christophefromparis](https://github.com/christophefromparis)]

# 2.2.0
* Policy Reporter UI v1.2.0
    * New configurations to customize the dashboard by disable PolicyReport- or ClusterPolicyReport information

# 2.1.1
* Fix KyvernoPlugin Metrics ServiceMonitor Port [[#96](https://github.com/kyverno/policy-reporter-ui/pull/96) by [z0rc](https://github.com/z0rc)]
* Remove unused Port from KyvernoPlugin Deployment and Service

# 2.1.0
* KyvernoPlugin v1.1.0
    * New KyvernoPlugin API - VerifyImages Rules (<a href="https://kyverno.github.io/policy-reporter/guide/06-troubleshooting#readinessprobe-fails" target="_blank">details</a>)
* Policy Reporter UI v1.1.0
    * New Kyverno VerifyImages view in Policy Reporter UI
    * New configurations to disable views (<a href="https://kyverno.github.io/policy-reporter/guide/04-helm-chart-core#configure-views" target="_blank">details</a>)

# 2.0.1
* Remove NetworkPolicy ingress rule for UI if not enabled
* Update Policy Reporter UI
    * Fix: Show PolicyReportResult Properties in Tables

# 2.0.0

## Chart
* Removed deprecated values `crdVersion`, `cleanupDebounceTime`
* Simplify `policyPriorities`, `policyPriorities.enabled` was removed along with the watch feature
    * Priority determined mainly over severity
* Add `sources` filter to target configurations
* Improved `NetworkPolicy` configuration for all components
* Metrics now an optional feature
* Each component expose a single Port `8080`

See [Migration Docs](https://kyverno.github.io/policy-reporter/guide/05-migration) for details

## Policy Reporter
* modular functions for separate activation/deactivation
    * REST API
    * Metrics API
    * Target pushes
* PolicyReports are now stored in an internal SQLite
* extended REST API based on the new SQLite DB for filters and grouping of data
* metrics API is now optional
* metrics and REST API using the same HTTP Server (were separated before)
* improved CRD watch logic with Kubernetes client informer
* `Yandex` changed to a general `S3` target.

## Policy Reporter UI
* Rewrite with NuxtJS
* Simplified Proxy
* Improved SPA file handling

## Policy Reporter Kyverno Plugin
* modular functions for separate activation/deactivation
    * REST API
    * Metrics API
* metrics and REST API using the same HTTP Server (were separated before)
* improved CRD watch logic with Kubernetes client informer

# 1.12.6
* Update Go Base Image for all Components
    * Policy Reporter [[#90](https://github.com/kyverno/policy-reporter-ui/pull/90) by [fjogeleit](https://github.com/fjogeleit)]
    * Policy Reporter UI [[#11](https://github.com/kyverno/policy-reporter-ui/pull/11) by [realshuting](https://github.com/realshuting)]
    * Policy Reporter Kyverno Plugin [[#9](https://github.com/kyverno/policy-reporter-ui/pull/9) by [realshuting](https://github.com/realshuting)]

# 1.12.5
* Dependency Update

# 1.12.4
* Fix policy-reporter-ui ServiceName function [[#87](https://github.com/kyverno/policy-reporter/pull/87) by [m-yosefpor](https://github.com/m-yosefpor)]

# 1.12.3
* Fix policy-reporter-ui backend name [[#85](https://github.com/kyverno/policy-reporter/pull/85) by [m-yosefpor](https://github.com/m-yosefpor)]

# 1.12.2
* Fix CRD registration for PolicyReport and ClusterPolicyReport

# 1.12.0
* Add Yandex as new Target for Policy Reporter

# 1.11.0
* Add Yandex as new Target for Policy Reporter

# 1.10.0
* Update Policy Reporter UI to v0.15.0
    * Add Filters as Query Parameters, make them shareable over links
* Hosting all new Images on the GitHub Container Registry instead of DockerHub
* Go Version updates to Go 1.17 of all components

# 1.9.4
* Make the Image Registry configurable with `image.registry` [[#74](https://github.com/kyverno/policy-reporter/pull/74) by [stone-z](https://github.com/stone-z)]

# 1.9.3
* Fix loki target messages for labels with dots

# 1.9.2
* Add additional egress rules to kyvernoPlugin and UI subchart with `networkPolicy.egress`

# 1.9.1
* Configure the Kubernetes API Port for NetworkPolicy with `networkPolicy.kubernetesApiPort`

# 1.9.0
* Implement NetworkPolicy for Policy Reporter and related Components [[#68](https://github.com/kyverno/policy-reporter/pull/68) by [windowsrefund](https://github.com/windowsrefund)]
* Customize liveness- and readinessProbe for Policy Reporter [[#67](https://github.com/kyverno/policy-reporter/pull/67) by [windowsrefund](https://github.com/windowsrefund)]

# 1.8.10
* Fix ServiceMonitor Namespace overwrite with `monitoring.serviceMonitor.namespace` instead of `monitoring.namespace`

# 1.8.9
* Ensure Backward Compatibility for `monitoring.namespace` configuration

# 1.8.8
* Optional Namespace Configuration for Monitoring ServiceMonitor
* Separat Namespace Configuration for Monitoring ConfigMaps with `monitoring.grafana.namespace`

# 1.8.7
* Update Policy Reporter UI to 0.14.0
    * Colored Diagrams
    * Suppport SubPath Configuration
* Restart CRD Watches when no CRDs are found
* Fix Ingress Resource in the UI Subchart
* Allow to override namespace for serviceMonitor [[#57](https://github.com/kyverno/policy-reporter/pull/57) by [Issif](https://github.com/Issif)]

# 1.8.6
* Update Policy Reporter UI to 0.13.1
    * Hide Rule Chips if rule name is empty
* Update Policy Reporter Kyvern Plugin to 0.3.2
    * Improved LivenessProbe, checks now if Kyverno CRDs are available
* Update Policy Reporter to 1.8.4
    * Improved LivenessProbe, checks now if any PolicyReport CRD is available

# 1.8.5
* Added trusted root CA's [[#52](https://github.com/kyverno/policy-reporter/pull/52) by [frezbo](https://github.com/frezbo)]

# 1.8.4
* Changed Organization

# 1.8.3
* Update Policy Reporter UI to 0.13.0
    * Change Result Grouping between by Status and by Category
    * Add source filter to ClusterPolicyReports

# 1.8.2
* Fix `scored` mapping for `v1alpha2/policyreports`
* Disable KyvernPlugin as default as expected
* Support `source` and `properties` for `policyreports/v1alpha2` in Policy Reporter UI
    * Update Policy Reporter UI to `0.12.0`

# 1.8.1
* Customize label and annotation for Grafana dashboards [[#43](https://github.com/kyverno/policy-reporter/pull/43) by [nlamirault](https://github.com/nlamirault)]
* ARM64 Support for all Components

# 1.7.3
* Update Policy Reporter - Kyverno Plugin to 0.2.0
    * New APIs for Liveness and Readiness Probes

# 1.7.2
* Update Policy Reporter - Kyverno Plugin to 0.1.2
    * Fix Handling of Validations with empty messages

# 1.7.1
* Fix HelmChart - Deployment Probes for Policy Reporter

# 1.7.0
* Enable REST API by default
    * Add `/healthz` and `/ready` APIs as new endpoints for readinessProbe and livenessProbe
* Helm Chart Updates
    * Add `global.labels` to add `labels` on every resource created
    * Add default labels on every resource

# 1.6.2
* Increase Result Caching Time to handle Kyverno issues with Policy reconcilation [Issue](https://github.com/kyverno/kyverno/issues/1921)
* Fix golint errors

# 1.6.1
* Add .global.fullnameOverride as new configuration for Policy Reporter Helm Chart
* Add static manifests to install Policy Reporter without Helm or Kustomize

# 1.6.0
* Internal refactoring
    * Unification of PolicyReports and ClusterPolicyReports processing, APIs still stable
    * DEPRECETED `crdVersion`, Policy Reporter handels now both versions by default
    * DEPRECETED `cleanupDebounceTime`, new internal caching replaced the debounce mechanism, debounce still exist with a fixed period to improve stable metric values.

# 1.5.0
* Support multiple Resources for a single Result
    * Mapping Result with multiple Resources in multiple Results with a single Resource
    * Upate UI handling with Results without Resources

# 1.4.1
* Update Kyverno Plugin
    * Fix Rule Type mapping
* Update Policy Reporter UI
    * Fix Chart rerender when values are the same

# 1.4.0
* Add Kyverno Plugins to the Helm Chart

## 1.3.4

* Configure Debounce Time in seconds for Cleanup Events over Helm Chart
    * Helm Value `cleanupDebounceTime` - default: 20
* Improved securityContext defaults

## 1.3.3

* Update Policy Reporter UI to v0.9.0
    * expand Tables with Validation Message
* Reduce log messages

## 1.3.2

* Compress REST API with GZIP
* Update Policy Reporter UI to 0.8.0
    * Support for GZIP Responses

## 1.3.1

* Debounce reconcile modification events for 10s to prevent resending violations

## 1.3.0

* New Helm Configuration
    * `crdVersion` changes the version of the PolicyReporter CRD - v1alpha1 is the current default

## 1.2.3

* Fix resend violations after KubeAPI reconnect

## 1.2.2

* Fix PolicyReportResult.timestamp parsing

## 1.2.1

* Support PolicyReportResult.status as well as PolicyReportResult.result for newer CRD versions

## 1.2.0

* Support for (Cluster)PolicyReport CRD Properties in Target Output
* Support for (Cluster)PolicyReport CRD Timestamp in Target Output
* Fix resend violations after Kyverno Cleanup with ResultHashes

## 1.1.0

* Added PolicyReport Category to Metrics
* New (Cluster)PolicyReport filter for Grafana Dashboards
    * Add __All__ Selection for Policy Filter
    * Category Filter
    * Severity Filter
    * Kind Filter
    * Namespacefilter (PolicyReports only)
* New (Cluster)PolicyReport filter for Policy Reporter UI
    * Category Filter
    * Severity Filter
    * Kind Filter

## 1.0.0

* Support Priority by Severity
    * high -> critical
    * medium -> warning
    * low -> information
* Severity is added as label to result metrics
* Severity is added in Policy Reporter UI tables
* Add "Critical" as new Priority to differ between Errored Policies and Failed priorities with High Severity
* Use "Warning" as new default Priority instead of Error which should now used for Policies in Error Status

## 0.22.0

* New Target Policy Reporter UI
  * New Log View in the Policy Reporter UI to see the latest log entries
  * Default: latest 200 logs with priority >= warning

## 0.21.0

* New Target MS Teams

## 0.20.2

* Policy Reporter UI update
  * Select All option for Policy Filter
  * New Namespace Filter for PolicyReport View

## 0.20.0

* [Breaking Change] rename policy-reporter-ui Subchart to ui
    * Simplify the customization by configure all PolicyReporter UI values under `ui`

## 0.19.0

* PolicyResult Priority mapping is now configurable over the Helm Chart

## 0.18.0

* Helm Chart updates [#16](https://github.com/kyverno/policy-reporter/pull/16) fixes [#14](https://github.com/kyverno/policy-reporter/issues/14)
    * Target Configuration are now configured under `target` in the HelmChart `values.yaml`
    * config.yaml are now deployed as Secret with encoded data body (plain stringData before)

## 0.17.0

* New Helm Linting Workflow by kolikons [#15](https://github.com/kyverno/policy-reporter/pull/15)
* Improved Helm Chart by kolikons [#13](https://github.com/kyverno/policy-reporter/pull/13)
    * More configuration possibilities like UI Ingress, ReplicaCount
    * Role and RoleBindings for ConfigMaps are now optional (required for Priority configuration)
## 0.16.0

* New Optional REST API
* New Optional Policy Reporter UI Helm SubChart

## 0.15.1

* Add a checksum for the target configuration secret to the deployment. This enforces a pod recreation when the configuration changed by a Helm upgrade.

## 0.15.0

* Customizable Dashboards via new Helm values for the Monitoring Subchart.
## 0.14.0

* Internal refactoring
    * Improved test coverage
    * Removed duplicated caching
* Updated Dashboard
    * Filter zero values from Policy Report Detail after Policies / Resources are deleted

## 0.13.0

* Split the Monitoring out in a Sub Helm chart
    * Changed naming from `metrics` to `monitoring`
* Make Annotations for the Deployment configurable
* Add two new Grafana Dashboard (PolicyReport Details, ClusterPolicyReport Details)

## 0.12.0

* Add support for a special `default` key in the Policy Priority. The `default` key can be used to configure a global default priority instead of `error`

## 0.11.1

* Use a Secret instead of ConfigMap to persist target configurations

## 0.11.0

* Helm Chart Value `metrics.serviceMonitor` changed to `metrics.serviceMonitor.enabled`
* New Helm Chart Value `metrics.serviceMonitor.labels` can be used to add additional `labels` to the `SeriveMonitor`. This helps to fullfil the `serviceMonitorSelector` of the `Prometheus` Resource in the MonitoringStack.

## 0.10.0

* Implement Discord as Target for PolicyReportResults

## 0.9.0

* Implement Slack as Target for PolicyReportResults

## 0.8.0

* Implement Elasticsearch as Target for PolicyReportResults
* Replace CLI flags with a single `config.yaml` to manage target-configurations as separate `ConfigMap`
* Set `loki.skipExistingOnStartup` default value to `true`
