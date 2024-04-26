# Policy Reporter UI - Custom Boards

CustomBoards allows you to configure additional dashboards with a custom subset of sources and namespaces, selected via a list and/or label selector

## Example CustomBoard Config

![Custom Boards](https://github.com/kyverno/policy-reporter/blob/3.x/docs/images/custom-boards/list.png)

```yaml
ui:
  enabled: true

  customBoards:
  - name: System
    namespaces:
      list:
        - kube-system
        - kyverno
        - policy-reporter
```

### CustomBoard with NamespaceSelector

![Custom Boards](https://github.com/kyverno/policy-reporter/blob/3.x/docs/images/custom-boards/selector.png)

```yaml
ui:
  enabled: true

  customBoards:
  - name: System
    namespaces:
      selector:
        group: system
```

### CustomBoard with ClusterResources
![Custom Boards](https://github.com/kyverno/policy-reporter/blob/3.x/docs/images/custom-boards/cluster.png)

```yaml
ui:
  enabled: true

  customBoards:
  - name: System
    clusterScope:
      enabled: true
    namespaces:
      selector:
        group: system
```

### CustomBoard with Source List

![Custom Boards](https://github.com/kyverno/policy-reporter/blob/3.x/docs/images/custom-boards/source.png)

```yaml
ui:
  enabled: true

  customBoards:
  - name: System
    clusterScope:
      enabled: true
    namespaces:
      selector:
        group: system
    sources:
        list: [kyverno]
```
