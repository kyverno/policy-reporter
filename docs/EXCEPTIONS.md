# Policy Reporter UI - Generate Kyverno PolicyExceptions

The Policy Reporter UI provides a visual overview of the policy status in your cluster, but no action you can take to change the status by default.

In the case of Kyverno, you have two options for dealing with policy failure. You can either fix it or create an exception for it. While the first option is difficult to automate and not always possible, creating an exception is relatively easy and can help exclude resources from validation that you are not able to fix immediately.

To support this process, the new Policy Reporter plugin system provides an Exception API that can be used to implement source-specific logic for PolicyException creation. The new Policy Reporter Kyverno plugin utilizes this API to provide an automated method for generating Kyverno PolicyException CRD resources that excludes a single or all failed policies depending on the context in the UI.

## Configuration

Because the Exception API is part of the Policy Reporter Kyverno Plugin, its required to install this plugin to use it and enable the exception feature.

### Helm 3 Configuration

```yaml
plugin:
  kyverno:
    enabled: true

ui:
  enabled: true
  sources:
    - name: kyverno
      exceptions: true
      excludes:
        namespaceKinds:
        - Pod
        - Job
        - ReplicaSet
        results:
        - warn
        - error
```

### Alternative manual UI Configuration

```yaml
# Configure the Kyverno Plugin the Cluster config
clusters:
- name: Default
  host: http://policy-reporter:8080
  plugins:
  - name: kyverno
    host: http://policy-reporter-kyverno-plugin:8080/api

# Enable `exceptions` in the kyverno source configuration
sources:
  - name: kyverno
    exceptions: true
    excludes:
      namespaceKinds:
      - Pod
      - Job
      - ReplicaSet
      results:
      - warn
      - error
```

### Examples

![Exception Resource List](https://github.com/kyverno/policy-reporter/blob/3.x/docs/images/exceptions/resource-list.png)

![Exception Dialog](https://github.com/kyverno/policy-reporter/blob/3.x/docs/images/exceptions/exception-dialog.png)
