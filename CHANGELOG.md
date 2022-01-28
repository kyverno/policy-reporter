# Changelog

# 2.2.3
* Fix Helm Chart uihost template function.

# 2.2.2
* Fix Helm Chart `values.yaml`. Cleanup unused default configurations. [[#103](https://github.com/kyverno/policy-reporter-ui/pull/103) by [AndersBennedsgaard](https://github.com/AndersBennedsgaard)]

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
