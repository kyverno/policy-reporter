# Changelog

# 1.8.2
* Fix `scored` mapping for `v1alpha2/policyreports`
* Disable KyvernPlugin as default as expected
* Support `source` and `properties` for `policyreports/v1alpha2` in Policy Reporter UI
    * Update Policy Reporter UI to `0.12.0`

# 1.8.1
* Customize label and annotation for Grafana dashboards [#43](https://github.com/fjogeleit/policy-reporter/pull/43) by [nlamirault](https://github.com/nlamirault)
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

* Helm Chart updates [#16](https://github.com/fjogeleit/policy-reporter/pull/16) fixes [#14](https://github.com/fjogeleit/policy-reporter/issues/14)
    * Target Configuration are now configured under `target` in the HelmChart `values.yaml`
    * config.yaml are now deployed as Secret with encoded data body (plain stringData before)

## 0.17.0

* New Helm Linting Workflow by kolikons [#15](https://github.com/fjogeleit/policy-reporter/pull/15)
* Improved Helm Chart by kolikons [#13](https://github.com/fjogeleit/policy-reporter/pull/13)
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
