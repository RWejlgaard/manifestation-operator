{{/* Base name used for resources, labels and selectors. Hardcoded "manifestation-operator"
     so the published artifact can be manifestation-operator-chart on Docker Hub without
     leaking that into in-cluster object names. */}}
{{- define "manifestation-operator.name" -}}
{{- default "manifestation-operator" .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "manifestation-operator.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default "manifestation-operator" .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{- define "manifestation-operator.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "manifestation-operator.labels" -}}
helm.sh/chart: {{ include "manifestation-operator.chart" . }}
{{ include "manifestation-operator.selectorLabels" . }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: manifestation-operator
{{- end -}}

{{- define "manifestation-operator.selectorLabels" -}}
app.kubernetes.io/name: {{ include "manifestation-operator.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
control-plane: controller-manager
{{- end -}}

{{- define "manifestation-operator.serviceAccountName" -}}
{{- if .Values.serviceAccount.create -}}
{{- default (include "manifestation-operator.fullname" .) .Values.serviceAccount.name -}}
{{- else -}}
{{- default "default" .Values.serviceAccount.name -}}
{{- end -}}
{{- end -}}

{{- define "manifestation-operator.image" -}}
{{- $tag := default .Chart.AppVersion .Values.image.tag -}}
{{- printf "%s:%s" .Values.image.repository $tag -}}
{{- end -}}
