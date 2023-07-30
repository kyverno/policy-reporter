{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "ui.fullname" -}}
{{- $name := "ui" }}
{{- if .Values.global.fullnameOverride }}
{{- printf "%s-%s" .Values.global.fullnameOverride $name | trunc 63 | trimSuffix "-" }}
{{- else if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}

{{- define "ui.name" -}}
{{- "ui" }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "ui.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
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
app.kubernetes.io/component: ui
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: policy-reporter
{{- with .Values.global.labels }}
{{ toYaml . }}
{{- end -}}
{{- with .Values.ingress.labels }}
{{ toYaml . }}
{{- end -}}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "ui.selectorLabels" -}}
app.kubernetes.io/name: {{ include "ui.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Pod labels
*/}}
{{- define "ui.podLabels" -}}
helm.sh/chart: {{ include "ui.chart" . }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
app.kubernetes.io/part-of: policy-reporter
{{- end }}

{{/*
Policy Reporter Selector labels
*/}}
{{- define "policyreporter.selectorLabels" -}}
app.kubernetes.io/name: policy-reporter
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Kyverno Plugin Selector labels
*/}}
{{- define "kyvernoplugin.selectorLabels" -}}
app.kubernetes.io/name: kyverno-plugin
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

{{- define "ui.kyvernoPluginServiceName" -}}
{{- $name := "kyverno-plugin" }}
{{- if .Values.global.fullnameOverride }}
{{- printf "%s-%s" .Values.global.fullnameOverride $name | trunc 63 | trimSuffix "-" }}
{{- else if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}

{{- define "ui.policyReportServiceName" -}}
{{- $name := "policy-reporter" }}
{{- if .Values.global.backend }}
{{- .Values.global.backend }}
{{- else if .Values.global.fullnameOverride }}
{{- .Values.global.fullnameOverride }}
{{- else if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}

{{- define "ui.securityContext" -}}
{{- if semverCompare "<1.19" .Capabilities.KubeVersion.Version }}
{{ toYaml (omit .Values.securityContext "seccompProfile") }}
{{- else }}
{{ toYaml .Values.securityContext }}
{{- end }}
{{- end }}

{{/* Get the namespace name. */}}
{{- define "ui.namespace" -}}
{{- if .Values.global.namespace -}}
    {{- .Values.global.namespace -}}
{{- else -}}
    {{- .Release.Namespace -}}
{{- end -}}
{{- end -}}

{{/* Get the namespace name. */}}
{{- define "ui.logLevel" -}}
{{- if .Values.api.logging -}}
-1
{{- else -}}
{{- .Values.logging.logLevel -}}
{{- end -}}
{{- end -}}
