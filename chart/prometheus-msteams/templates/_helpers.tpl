{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "app.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app.name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "app.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "app.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Common labels
*/}}
{{- define "app.labels" -}}
helm.sh/chart: {{ include "app.chart" . }}
{{ include "app.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "app.selectorLabels" -}}
app.kubernetes.io/name: {{ include "app.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "app.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "app.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Create the card template to use.
*/}}
{{- define "app.cardTemplate" -}}
{{- if .Values.customCardTemplateFile }}
{{ .Files.Get .Values.customCardTemplateFile }}
{{- else if .Values.customCardTemplate }}
{{ .Values.customCardTemplate }}
{{- else if .Values.workflowWebhook}}
{{ .Files.Get "data/cardWorkflow.tmpl" }}
{{- else }}
{{ .Files.Get "data/card.tmpl" }}
{{- end }}
{{- end }}

{{- define "app.customTemplates.config" -}}
{{- $result := list -}}
{{- range $i, $c := . -}}
{{- $d := (dict 
  "request_path" $c.request_path 
  "template_file" (printf "/etc/template/custom_card_%d.tmpl" $i) 
  "webhook_url" $c.webhook_url "escape_underscores"
  (hasKey $c "escape_underscores" | ternary $c.escape_underscores nil)
  ) -}}
{{- $result = append $result $d -}}
{{- end }}
{{- $result | toYaml -}}
{{- end }}

{{- define "app.customTemplates.tmpl" -}}
{{- $result := list -}}
{{- range $i, $c := . -}}
{{- if hasKey $c "template_file" }}
{{- $key := printf "custom_card_%d.tmpl" $i -}}
{{- $d := (dict $key ($c.template_file | b64enc)) -}}
{{- $result = append $result $d -}}
{{- end }}
{{- end }}
{{- $result | toYaml -}}
{{- end }}
