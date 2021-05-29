{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "kyvernoplugin.fullname" -}}
{{- $name := "kyverno-plugin" }}
{{- if .Values.global.fullnameOverride }}
{{- printf "%s-%s" .Values.global.fullnameOverride $name | trunc 63 | trimSuffix "-" }}
{{- else if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}

{{- define "kyvernoplugin.name" -}}
{{- "kyverno-plugin" }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "kyvernoplugin.chart" -}}
{{- printf "kyverno-plugin-%s" .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "kyvernoplugin.labels" -}}
helm.sh/chart: {{ include "kyvernoplugin.chart" . }}
{{ include "kyvernoplugin.selectorLabels" . }}
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
{{- define "kyvernoplugin.selectorLabels" -}}
app.kubernetes.io/name: {{ include "kyvernoplugin.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "kyvernoplugin.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "kyvernoplugin.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}
