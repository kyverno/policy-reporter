{{- define "policyreporter.name" -}}
{{- "policy-reporter" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "policyreporter.fullname" -}}
{{- $name := .Chart.Name }}
{{- if .Values.global.fullnameOverride }}
{{- .Values.global.fullnameOverride }}
{{- else if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "policyreporter.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "policyreporter.labels" -}}
helm.sh/chart: {{ include "policyreporter.chart" . }}
{{ include "policyreporter.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- with .Values.global.labels }}
{{ toYaml . }}
{{- end -}}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "policyreporter.selectorLabels" -}}
app.kubernetes.io/name: {{ include "policyreporter.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "policyreporter.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "policyreporter.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Create UI target host based on configuration
*/}}
{{- define "policyreporter.uihost" -}}
{{ if .Values.target.ui.host }}
{{- else if .Values.ui.enabled }}
{{- printf "http://%s-ui:%s" .Release.Name (.Values.ui.service.port | toString) }}
{{- else }}
{{- "" }}
{{- end }}
{{- end }}
