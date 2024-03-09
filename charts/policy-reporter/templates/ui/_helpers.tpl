{{/*
Expand the name of the chart.
*/}}
{{- define "ui.name" -}}
{{ template "policyreporter.name" . }}-ui
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "ui.fullname" -}}
{{ template "policyreporter.fullname" . }}-ui
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "ui.chart" -}}
{{ template "policyreporter.chart" . }}-ui
{{- end }}

{{/*
Common labels
*/}}
{{- define "ui.labels" -}}
helm.sh/chart: {{ include "ui.chart" . }}
{{ include "ui.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "ui.selectorLabels" -}}
app.kubernetes.io/name: {{ include "ui.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "ui.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "ui.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{- define "ui.podDisruptionBudget" -}}
{{- if and .Values.podDisruptionBudget.minAvailable .Values.podDisruptionBudget.maxUnavailable }}
{{- fail "Cannot set both .Values.podDisruptionBudget.minAvailable and .Values.podDisruptionBudget.maxUnavailable" -}}
{{- end }}
{{- if not .Values.podDisruptionBudget.maxUnavailable }}
minAvailable: {{ default 1 .Values.podDisruptionBudget.minAvailable }}
{{- end }}
{{- if .Values.podDisruptionBudget.maxUnavailable }}
maxUnavailable: {{ .Values.podDisruptionBudget.maxUnavailable }}
{{- end }}
{{- end }}
