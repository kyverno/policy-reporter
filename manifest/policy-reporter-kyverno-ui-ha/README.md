# Policy Reporter Configuration

Encoded configuraion from `config-core.yaml`

Seet the [Documentation](https://kyverno.github.io/policy-reporter/core/config-reference) for a full reference of all possible configurations

```yaml
# send pushes to the Policy Reporter UI
ui:
  host: http://policy-reporter-ui:8082
  minimumSeverity: "medium"
  skipExistingOnStartup: true

# (optional) cache results in an central, external redis
redis:
  enabled: true
  address: "redis:6379"
  database: 1
  prefix: "policy-reporter"
```

# Policy Reporter UI Configuration

Encoded configuraion from `config-ui.yaml`

Seet the [Documentation](https://kyverno.github.io/policy-reporter/ui/config-reference) for a full reference of all possible configurations

```yaml
# central and external Log storage
# only needed if you use the "Logs" page and want to have all logs on each instance available
redis:
  enabled: true
  address: "redis:6379"
  database: 1
  prefix: "policy-reporter-ui"
```

# Policy Reporter Kyverno Plugin Configuration

Encoded configuraion from `config-kyverno-plugin.yaml`

Seet the [Documentation](https://kyverno.github.io/policy-reporter/kyverno-plugin/config-reference) for a full reference of all possible configurations

```yaml
# enables the creation of PolicyReports for blocked resources by enforce policies
blockReports:
  enabled: true
```
