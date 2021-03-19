# Changelog

## 0.20.1

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
