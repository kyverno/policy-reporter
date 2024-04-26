# Configure Authentication for Policy Reporter UI

With Policy Reporter UI v2 it is possible to use either OAuth2 or OpenIDConnect as authentication mechanism.

Its not possible to reduce or configure view permission based on roles or any other information yet. 
Authentication ensures that no unauthorized person is able to open the UI at all.

## OAuth2

Policy Reporter UI v2 supports a fixed set of oauth2 providers. If the provider of your choice is not yet supported, you can submit a feature request for it.

### Supported OAuth Provider

* amazon
* gitlab 
* github 
* apple
* google
* yandex
* azuread

### Example Configuration (GitHub Provider)

Since the callback URL depends on your setup, you must explicitly configure it.

```yaml
ui:
  oauth:
    enabled: true
    clientId: c79c02881aa1...
    clientSecret: fb2035255d0bd182c9...
    provider: github
    callback: http://localhost:8082/callback
    scopes: []
```

### Example SecretRef

Instead of providing the information directly in the values, you can also fetch the information from an existing secret.

#### Values

```yaml
ui:
  oauth:
    enabled: true
    callback: http://localhost:8082/callback
    scopes: []
    secretRef: 'github-provider'
```
#### Secret

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: github-provider
data:
  clientId: Yzc5YzAyODgxYWEx
  clientSecret: ZmIyMDM1MjU1ZDBiZDE4MmM5
  provider: Z2l0aHVi
```

## OpenIDConnect

This authentication mechanism supports all compatible services and systems.

### Example Configuration (Keycloak)

```yaml
ui:
  openIDConnect:
    enabled: true
    clientId: policy-reporter
    clientSecret: c11cYF9tNtL94w....
    callbackUrl: http://localhost:8082/callback
    discoveryUrl: 'https://keycloak.instance.de/realms/timetracker'
```

### Example SecretRef

Instead of providing the information directly in the values, you can also fetch the information from an existing secret.

#### Values

```yaml
ui:
  openIDConnect:
    enabled: true
    callback: http://localhost:8082/callback
    secretRef: 'keycloak-provider'
```
#### Secret

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: keycloak-provider
data:
  clientId: Yzc5YzAyODgxYWEx
  clientSecret: ZmIyMDM1MjU1ZDBiZDE4MmM5
  discoveryUrl: aHR0cHM6Ly9rZXljbG9hay5pbnN0YW5jZS5kZS9yZWFsbXMvdGltZXRyYWNrZXI=
```
