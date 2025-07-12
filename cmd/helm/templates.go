package helm

// getHelpersTemplate returns the _helpers.tpl template
func getHelpersTemplate() string {
	return `{{/*
Expand the name of the chart.
*/}}
{{- define "{{.ChartName}}.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "{{.ChartName}}.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "{{.ChartName}}.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "{{.ChartName}}.labels" -}}
helm.sh/chart: {{ include "{{.ChartName}}.chart" . }}
{{ include "{{.ChartName}}.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "{{.ChartName}}.selectorLabels" -}}
app.kubernetes.io/name: {{ include "{{.ChartName}}.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "{{.ChartName}}.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "{{.ChartName}}.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Database URL for external services
*/}}
{{- define "{{.ChartName}}.databaseUrl" -}}
{{- if .Values.postgresql.enabled }}
{{- printf "postgresql://%s:%s@%s-postgresql:5432/%s" .Values.postgresql.auth.username .Values.postgresql.auth.password .Release.Name .Values.postgresql.auth.database }}
{{- else }}
{{- .Values.externalDatabase.url }}
{{- end }}
{{- end }}

{{/*
Redis URL for external services
*/}}
{{- define "{{.ChartName}}.redisUrl" -}}
{{- if .Values.redis.enabled }}
{{- if .Values.redis.auth.enabled }}
{{- printf "redis://:%s@%s-redis-master:6379" .Values.redis.auth.password .Release.Name }}
{{- else }}
{{- printf "redis://%s-redis-master:6379" .Release.Name }}
{{- end }}
{{- else }}
{{- .Values.externalRedis.url }}
{{- end }}
{{- end }}`
}

// getDeploymentTemplate returns the deployment.yaml template
func getDeploymentTemplate() string {
	return `apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ include "{{.ChartName}}.fullname" . }}
  labels:
    {{- include "{{.ChartName}}.labels" . | nindent 4 }}
spec:
  {{- if not .Values.autoscaling.enabled }}
  replicas: {{ .Values.replicaCount }}
  {{- end }}
  selector:
    matchLabels:
      {{- include "{{.ChartName}}.selectorLabels" . | nindent 6 }}
  template:
    metadata:
      annotations:
        checksum/config: {{ include (print $.Template.BasePath "/configmap.yaml") . | sha256sum }}
        checksum/secret: {{ include (print $.Template.BasePath "/secret.yaml") . | sha256sum }}
        {{- with .Values.podAnnotations }}
        {{- toYaml . | nindent 8 }}
        {{- end }}
      labels:
        {{- include "{{.ChartName}}.selectorLabels" . | nindent 8 }}
    spec:
      {{- with .Values.imagePullSecrets }}
      imagePullSecrets:
        {{- toYaml . | nindent 8 }}
      {{- end }}
      serviceAccountName: {{ include "{{.ChartName}}.serviceAccountName" . }}
      securityContext:
        {{- toYaml .Values.podSecurityContext | nindent 8 }}
      containers:
        - name: {{ .Chart.Name }}
          securityContext:
            {{- toYaml .Values.securityContext | nindent 12 }}
          image: "{{ .Values.image.repository }}:{{ .Values.image.tag | default .Chart.AppVersion }}"
          imagePullPolicy: {{ .Values.image.pullPolicy }}
          ports:
            - name: http
              containerPort: {{.Project.Port}}
              protocol: TCP
          {{- if .Values.livenessProbe }}
          livenessProbe:
            {{- toYaml .Values.livenessProbe | nindent 12 }}
          {{- end }}
          {{- if .Values.readinessProbe }}
          readinessProbe:
            {{- toYaml .Values.readinessProbe | nindent 12 }}
          {{- end }}
          {{- if .Values.startupProbe }}
          startupProbe:
            {{- toYaml .Values.startupProbe | nindent 12 }}
          {{- end }}
          resources:
            {{- toYaml .Values.resources | nindent 12 }}
          env:
            - name: PORT
              value: "{{.Project.Port}}"
            {{- if .Values.postgresql.enabled }}
            - name: DATABASE_URL
              value: {{ include "{{.ChartName}}.databaseUrl" . }}
            {{- end }}
            {{- if .Values.redis.enabled }}
            - name: REDIS_URL
              value: {{ include "{{.ChartName}}.redisUrl" . }}
            {{- end }}
            {{- range .Values.env }}
            - {{- toYaml . | nindent 14 }}
            {{- end }}
          {{- if .Values.envFrom }}
          envFrom:
            {{- toYaml .Values.envFrom | nindent 12 }}
          {{- end }}
          {{- if .Values.volumeMounts }}
          volumeMounts:
            {{- toYaml .Values.volumeMounts | nindent 12 }}
          {{- end }}
      {{- if .Values.volumes }}
      volumes:
        {{- toYaml .Values.volumes | nindent 8 }}
      {{- end }}
      {{- with .Values.nodeSelector }}
      nodeSelector:
        {{- toYaml . | nindent 8 }}
      {{- end }}
      {{- with .Values.affinity }}
      affinity:
        {{- toYaml . | nindent 8 }}
      {{- end }}
      {{- with .Values.tolerations }}
      tolerations:
        {{- toYaml . | nindent 8 }}
      {{- end }}`
}

// getServiceTemplate returns the service.yaml template
func getServiceTemplate() string {
	return `{{- if .Values.service }}
apiVersion: v1
kind: Service
metadata:
  name: {{ include "{{.ChartName}}.fullname" . }}
  labels:
    {{- include "{{.ChartName}}.labels" . | nindent 4 }}
spec:
  type: {{ .Values.service.type }}
  ports:
    - port: {{ .Values.service.port }}
      targetPort: http
      protocol: TCP
      name: http
  selector:
    {{- include "{{.ChartName}}.selectorLabels" . | nindent 4 }}
{{- end }}`
}

// getServiceAccountTemplate returns the serviceaccount.yaml template
func getServiceAccountTemplate() string {
	return `{{- if .Values.serviceAccount.create -}}
apiVersion: v1
kind: ServiceAccount
metadata:
  name: {{ include "{{.ChartName}}.serviceAccountName" . }}
  labels:
    {{- include "{{.ChartName}}.labels" . | nindent 4 }}
  {{- with .Values.serviceAccount.annotations }}
  annotations:
    {{- toYaml . | nindent 4 }}
  {{- end }}
{{- end }}`
}

// getIngressTemplate returns the ingress.yaml template
func getIngressTemplate() string {
	return `{{- if .Values.ingress.enabled -}}
{{- $fullName := include "{{.ChartName}}.fullname" . -}}
{{- $svcPort := .Values.service.port -}}
{{- if and .Values.ingress.className (not (hasKey .Values.ingress.annotations "kubernetes.io/ingress.class")) }}
  {{- $_ := set .Values.ingress.annotations "kubernetes.io/ingress.class" .Values.ingress.className}}
{{- end }}
{{- if semverCompare ">=1.19-0" .Capabilities.KubeVersion.GitVersion -}}
apiVersion: networking.k8s.io/v1
{{- else if semverCompare ">=1.14-0" .Capabilities.KubeVersion.GitVersion -}}
apiVersion: networking.k8s.io/v1beta1
{{- else -}}
apiVersion: extensions/v1beta1
{{- end }}
kind: Ingress
metadata:
  name: {{ $fullName }}
  labels:
    {{- include "{{.ChartName}}.labels" . | nindent 4 }}
  {{- with .Values.ingress.annotations }}
  annotations:
    {{- toYaml . | nindent 4 }}
  {{- end }}
spec:
  {{- if and .Values.ingress.className (semverCompare ">=1.18-0" .Capabilities.KubeVersion.GitVersion) }}
  ingressClassName: {{ .Values.ingress.className }}
  {{- end }}
  {{- if .Values.ingress.tls }}
  tls:
    {{- range .Values.ingress.tls }}
    - hosts:
        {{- range .hosts }}
        - {{ . | quote }}
        {{- end }}
      secretName: {{ .secretName }}
    {{- end }}
  {{- end }}
  rules:
    {{- range .Values.ingress.hosts }}
    - host: {{ .host | quote }}
      http:
        paths:
          {{- range .paths }}
          - path: {{ .path }}
            {{- if and .pathType (semverCompare ">=1.18-0" $.Capabilities.KubeVersion.GitVersion) }}
            pathType: {{ .pathType }}
            {{- end }}
            backend:
              {{- if semverCompare ">=1.19-0" $.Capabilities.KubeVersion.GitVersion }}
              service:
                name: {{ $fullName }}
                port:
                  number: {{ $svcPort }}
              {{- else }}
              serviceName: {{ $fullName }}
              servicePort: {{ $svcPort }}
              {{- end }}
          {{- end }}
    {{- end }}
{{- end }}`
}

// getHPATemplate returns the hpa.yaml template
func getHPATemplate() string {
	return `{{- if .Values.autoscaling.enabled }}
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: {{ include "{{.ChartName}}.fullname" . }}
  labels:
    {{- include "{{.ChartName}}.labels" . | nindent 4 }}
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: {{ include "{{.ChartName}}.fullname" . }}
  minReplicas: {{ .Values.autoscaling.minReplicas }}
  maxReplicas: {{ .Values.autoscaling.maxReplicas }}
  metrics:
    {{- if .Values.autoscaling.targetCPUUtilizationPercentage }}
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: {{ .Values.autoscaling.targetCPUUtilizationPercentage }}
    {{- end }}
    {{- if .Values.autoscaling.targetMemoryUtilizationPercentage }}
    - type: Resource
      resource:
        name: memory
        target:
          type: Utilization
          averageUtilization: {{ .Values.autoscaling.targetMemoryUtilizationPercentage }}
    {{- end }}
{{- end }}`
}

// getPDBTemplate returns the pdb.yaml template
func getPDBTemplate() string {
	return `{{- if .Values.podDisruptionBudget.enabled }}
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: {{ include "{{.ChartName}}.fullname" . }}
  labels:
    {{- include "{{.ChartName}}.labels" . | nindent 4 }}
spec:
  {{- if .Values.podDisruptionBudget.minAvailable }}
  minAvailable: {{ .Values.podDisruptionBudget.minAvailable }}
  {{- end }}
  {{- if .Values.podDisruptionBudget.maxUnavailable }}
  maxUnavailable: {{ .Values.podDisruptionBudget.maxUnavailable }}
  {{- end }}
  selector:
    matchLabels:
      {{- include "{{.ChartName}}.selectorLabels" . | nindent 6 }}
{{- end }}`
}

// getConfigMapTemplate returns the configmap.yaml template
func getConfigMapTemplate() string {
	return `apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ include "{{.ChartName}}.fullname" . }}
  labels:
    {{- include "{{.ChartName}}.labels" . | nindent 4 }}
data:
  # Add your configuration data here
  app.properties: |
    # Application configuration
    server.port={{.Project.Port}}
    {{- if eq .Project.Language "go" }}
    # Go application configuration
    go.env=production
    {{- else if eq .Project.Language "node" }}
    # Node.js application configuration
    NODE_ENV=production
    {{- else if eq .Project.Language "python" }}
    # Python application configuration
    PYTHONUNBUFFERED=1
    {{- else if eq .Project.Language "java" }}
    # Java application configuration
    JAVA_OPTS=-Xmx512m -Xms256m
    {{- end }}`
}

// getSecretTemplate returns the secret.yaml template
func getSecretTemplate() string {
	return `apiVersion: v1
kind: Secret
metadata:
  name: {{ include "{{.ChartName}}.fullname" . }}
  labels:
    {{- include "{{.ChartName}}.labels" . | nindent 4 }}
type: Opaque
data:
  # Add your secret data here (base64 encoded)
  # Example:
  # api-key: {{ .Values.apiKey | b64enc | quote }}
  {{- if .Values.secrets }}
  {{- range $key, $value := .Values.secrets }}
  {{ $key }}: {{ $value | b64enc | quote }}
  {{- end }}
  {{- end }}`
}

// getTestTemplate returns the test-connection.yaml template
func getTestTemplate() string {
	return `apiVersion: v1
kind: Pod
metadata:
  name: "{{ include "{{.ChartName}}.fullname" . }}-test-connection"
  labels:
    {{- include "{{.ChartName}}.labels" . | nindent 4 }}
  annotations:
    "helm.sh/hook": test
spec:
  restartPolicy: Never
  containers:
    - name: wget
      image: busybox
      command: ['wget']
      args: ['{{ include "{{.ChartName}}.fullname" . }}:{{ .Values.service.port }}']`
}
