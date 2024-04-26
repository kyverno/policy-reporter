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
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride }}
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
{{- fail "Cannot set both minAvailable and maxUnavailable" -}}
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
{{- if .Values.namespaceOverride -}}
    {{- .Values.namespaceOverride -}}
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

{{- define "target" -}}
name: {{ .name | quote }}
path: {{ .path | quote }}
secretRef: {{ .secretRef | quote }}
mountedSecret: {{ .mountedSecret | quote }}
minimumPriority: {{ .minimumPriority | quote }}
skipExistingOnStartup: {{ .skipExistingOnStartup }}
{{- with .customFields }}
customFields:
{{- toYaml . | nindent 2 }}
{{- end }}
{{- with .sources }}
sources:
{{- toYaml . | nindent 2 }}
{{- end }}
{{- with .filter }}
filter:
{{- toYaml . | nindent 2 }}
{{- end }}
{{- end }}

{{- define "target.loki" -}}
config:
  host: {{ .host | quote }}
  certificate: {{ .certificate | quote }}
  skipTLS: {{ .skipTLS }}
  path: {{ .path | quote }}
{{ include "target" . }}
{{- end }}

{{- define "target.elasticsearch" -}}
config:
  host: {{ .host | quote }}
  certificate: {{ .certificate | quote }}
  skipTLS: {{ .skipTLS }}
  username: {{ .username | quote }}
  password: {{ .password | quote }}
  apiKey: {{ .apiKey | quote }}
  index: {{ .index| quote }}
  rotation: {{ .rotation | quote }}
{{ include "target" . }}
{{- end }}

{{- define "target.slack" -}}
config:
  webhook: {{ .webhook | quote }}
  channel: {{ .channel | quote }}
  certificate: {{ .certificate | quote }}
  skipTLS: {{ .skipTLS }}
  {{- with .headers }}
  headers:
  {{- toYaml . | nindent 4 }}
  {{- end }}
{{ include "target" . }}
{{- end }}

{{- define "target.webhook" -}}
config:
  webhook: {{ .webhook | quote }}
  certificate: {{ .certificate | quote }}
  skipTLS: {{ .skipTLS }}
  {{- with .headers }}
  headers:
  {{- toYaml . | nindent 4 }}
  {{- end }}
{{ include "target" . }}
{{- end }}

{{- define "target.telegram" -}}
config:
  chatID: {{ .chatID | quote }}
  token: {{ .token | quote }}
  webhook: {{ .webhook | quote }}
  certificate: {{ .certificate | quote }}
  skipTLS: {{ .skipTLS }}
  {{- with .headers }}
  headers:
  {{- toYaml . | nindent 4 }}
  {{- end }}
{{ include "target" . }}
{{- end }}

{{- define "target.s3" -}}
config:
  accessKeyID: {{ .accessKeyID }}
  secretAccessKey:  {{ .secretAccessKey }}
  region: {{ .region }}
  endpoint: {{ .endpoint }}
  bucket: {{ .bucket }}
  bucketKeyEnabled: {{ .bucketKeyEnabled }}
  kmsKeyId: {{ .kmsKeyId }}
  serverSideEncryption: {{ .serverSideEncryption }}
  pathStyle: {{ .pathStyle }}
  prefix: {{ .prefix }}
{{ include "target" . }}
{{- end }}

{{- define "target.kinesis" -}}
config:
  accessKeyID: {{ .accessKeyID }}
  secretAccessKey:  {{ .secretAccessKey }}
  region: {{ .region }}
  endpoint: {{ .endpoint }}
  streamName: {{ .streamName }}
{{ include "target" . }}
{{- end }}

{{- define "target.securityhub" -}}
config:
  accessKeyID: {{ .accessKeyID }}
  secretAccessKey:  {{ .secretAccessKey }}
  region: {{ .region }}
  endpoint: {{ .endpoint }}
  accountID: {{ .accountID }}
{{ include "target" . }}
{{- end }}

{{- define "target.gcs" -}}
config:
  credentials: {{ .credentials }}
  bucket: {{ .bucket }}
  prefix: {{ .prefix }}
{{ include "target" . }}
{{- end }}
