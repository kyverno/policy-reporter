{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "monitoring.fullname" -}}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if .Values.global.fullnameOverride }}
{{- printf "%s-%s" .Values.global.fullnameOverride $name | trunc 63 | trimSuffix "-" }}
{{- else if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "monitoring.labels" -}}
helm.sh/chart: {{ include "policyreporter.chart" . }}
{{ include "monitoring.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/component: monitoring
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: kyverno
{{- with .Values.global.labels }}
{{ toYaml . }}
{{- end -}}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "monitoring.selectorLabels" -}}
app.kubernetes.io/name: {{ include "monitoring.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{- define "monitoring.name" -}}
{{- "monitoring" }}
{{- end }}

{{- define "monitoring.namespace" -}}
{{-  if .Values.grafana.namespace -}}
{{- .Values.grafana.namespace -}}
{{- else if .Values.global.namespace -}}
    {{- .Values.global.namespace -}}
{{- else -}}
{{- .Release.Namespace -}}
{{- end }}
{{- end }}

{{- define "kyvernoplugin.selectorLabels" -}}
app.kubernetes.io/name: kyverno-plugin
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/* Get the namespace name. */}}
{{- define "monitoring.smNamespace" -}}
{{-  if .Values.serviceMonitor.namespace -}}
{{- .Values.serviceMonitor.namespace -}}
{{- else if .Values.global.namespace -}}
    {{- .Values.global.namespace -}}
{{- else -}}
{{- .Release.Namespace -}}
{{- end }}
{{- end }}
