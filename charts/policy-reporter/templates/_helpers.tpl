{{- define "policyreporter.name" -}}
{{- "policy-reporter" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "policyreporter.fullname" -}}
{{- $name := default .Chart.Name .Values.nameOverride }}
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
app.kubernetes.io/component: reporting
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: policy-reporter
{{- with .Values.global.labels }}
{{ toYaml . }}
{{- end -}}
{{- end }}

{{/*
Pod labels
*/}}
{{- define "policyreporter.podLabels" -}}
helm.sh/chart: {{ include "policyreporter.chart" . }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
app.kubernetes.io/part-of: policy-reporter
{{- end }}

{{/*
Selector labels
*/}}
{{- define "policyreporter.selectorLabels" -}}
app.kubernetes.io/name: policy-reporter
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
{{- .Values.target.ui.host }}
{{- else if not .Values.ui.enabled }}
{{- "" }}
{{- else if and .Values.ui.enabled (and .Values.ui.views.logs .Values.ui.service.enabled) }}
{{- printf "http://%s:%s" (include "ui.fullname" .) (.Values.ui.service.port | toString) }}
{{- else }}
{{- "" }}
{{- end }}
{{- end }}

{{- define "policyreporter.securityContext" -}}
{{- if semverCompare "<1.19" .Capabilities.KubeVersion.Version }}
{{- toYaml (omit .Values.securityContext "seccompProfile") }}
{{- else }}
{{- toYaml .Values.securityContext }}
{{- end }}
{{- end }}

{{- define "policyreporter.podDisruptionBudget" -}}
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

{{/* Get the namespace name. */}}
{{- define "policyreporter.namespace" -}}
{{- if .Values.global.namespace -}}
    {{- .Values.global.namespace -}}
{{- else -}}
    {{- .Release.Namespace -}}
{{- end -}}
{{- end -}}

{{/* Get the namespace name. */}}
{{- define "policyreporter.logLevel" -}}
{{- if .Values.api.logging -}}
-1
{{- else -}}
{{- .Values.logging.logLevel -}}
{{- end -}}
{{- end -}}
