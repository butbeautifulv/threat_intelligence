{{- define "veil.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "veil.fullname" -}}
{{- printf "%s-%s" .Release.Name (include "veil.name" .) | trunc 63 | trimSuffix "-" }}
{{- end }}
