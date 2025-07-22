# 관찰성 플랫폼 기능

## 개요

종합적인 모니터링, 로깅, 추적 및 메트릭 수집 플랫폼

## 제거된 기능

### 1. 메트릭 수집 및 모니터링

- **명령어**: `gz serve metrics`, `gz monitoring start`
- **기능**: Prometheus 메트릭 수집 및 Grafana 대시보드
- **특징**:
  - 실시간 시스템 메트릭
  - 커스텀 메트릭 정의
  - 알람 및 알림 규칙
  - 다중 데이터 소스 지원

### 2. 로그 집계 및 분석

- **명령어**: `gz logs collect`, `gz logs analyze`
- **기능**: 중앙집중식 로그 관리 및 분석
- **특징**:
  - 구조화된 로그 수집
  - 로그 파싱 및 인덱싱
  - 실시간 로그 스트리밍
  - 로그 보관 정책

### 3. 분산 추적

- **명령어**: `gz tracing setup`, `gz tracing analyze`
- **기능**: 마이크로서비스 간 요청 추적
- **특징**:
  - OpenTelemetry 통합
  - 서비스 맵 생성
  - 성능 병목 지점 식별
  - 에러 전파 추적

### 4. 성능 모니터링

- **명령어**: `gz performance monitor`, `gz performance analyze`
- **기능**: 애플리케이션 성능 모니터링 (APM)
- **특징**:
  - 응답 시간 측정
  - 처리량 분석
  - 에러율 추적
  - 사용자 경험 모니터링

## 사용 예시 (제거 전)

```bash
# 모니터링 스택 시작
gz serve metrics --port 9090 --grafana-port 3000

# 메트릭 수집 설정
gz monitoring setup --targets "localhost:8080,localhost:8081" \
  --scrape-interval 15s

# 로그 수집기 시작
gz logs collect --input "/var/log/*.log" \
  --output elasticsearch://localhost:9200

# 추적 설정
gz tracing setup --jaeger-endpoint http://localhost:14268 \
  --service-name myapp
```

## 설정 파일 형식

```yaml
observability:
  metrics:
    prometheus:
      listen_address: "0.0.0.0:9090"
      scrape_interval: 15s
      evaluation_interval: 15s

      scrape_configs:
        - job_name: "myapp"
          static_configs:
            - targets: ["localhost:8080"]
          metrics_path: /metrics
          scrape_interval: 5s

        - job_name: "node-exporter"
          static_configs:
            - targets: ["localhost:9100"]

    grafana:
      listen_address: "0.0.0.0:3000"
      admin_user: admin
      admin_password: ${GRAFANA_PASSWORD}

      datasources:
        - name: Prometheus
          type: prometheus
          url: http://localhost:9090
          access: proxy
          is_default: true

      dashboards:
        - name: Application Metrics
          file: dashboards/app-metrics.json
        - name: Infrastructure Metrics
          file: dashboards/infra-metrics.json

  logging:
    elasticsearch:
      hosts: ["localhost:9200"]
      index_template: "logs-%{+YYYY.MM.dd}"

    logstash:
      host: "localhost:5044"
      input:
        beats:
          port: 5044
      filter:
        grok:
          patterns_dir: "/etc/logstash/patterns"
      output:
        elasticsearch:
          hosts: ["localhost:9200"]

    filebeat:
      inputs:
        - type: log
          paths:
            - /var/log/myapp/*.log
          fields:
            service: myapp
            environment: production

  tracing:
    jaeger:
      collector_endpoint: "http://localhost:14268/api/traces"
      agent_endpoint: "localhost:6831"

    opentelemetry:
      receivers:
        otlp:
          protocols:
            grpc:
              endpoint: 0.0.0.0:4317
            http:
              endpoint: 0.0.0.0:4318

      processors:
        batch:
          timeout: 1s
          send_batch_size: 1024

      exporters:
        jaeger:
          endpoint: http://localhost:14250

  alerting:
    rules:
      - name: high_cpu_usage
        expr: cpu_usage_percent > 80
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High CPU usage detected"

      - name: high_error_rate
        expr: error_rate > 0.05
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "High error rate detected"

    notification_channels:
      - name: slack
        type: slack
        webhook_url: https://hooks.slack.com/...
        channel: "#alerts"

      - name: email
        type: email
        addresses: [devops@company.com]

  retention:
    metrics: 30d
    logs: 7d
    traces: 3d
```

## 고급 기능

### 1. 서비스 디스커버리

- Kubernetes 서비스 자동 발견
- Consul 연동
- DNS 기반 발견
- 동적 타겟 관리

### 2. 멀티 테넌시

- 조직별 데이터 분리
- 사용자 권한 관리
- 리소스 할당량
- 독립적인 대시보드

### 3. 고가용성

- 클러스터 구성
- 데이터 복제
- 자동 장애조치
- 로드 밸런싱

### 4. 데이터 압축 및 샘플링

- 메트릭 다운샘플링
- 로그 압축
- 추적 샘플링
- 스토리지 최적화

## 대시보드 템플릿

### 1. 애플리케이션 메트릭

```json
{
  "dashboard": {
    "title": "Application Metrics",
    "panels": [
      {
        "title": "Request Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(http_requests_total[5m])",
            "legendFormat": "{{method}} {{status}}"
          }
        ]
      },
      {
        "title": "Response Time",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))",
            "legendFormat": "95th percentile"
          }
        ]
      },
      {
        "title": "Error Rate",
        "type": "singlestat",
        "targets": [
          {
            "expr": "rate(http_requests_total{status=~\"5..\"}[5m]) / rate(http_requests_total[5m])",
            "legendFormat": "Error Rate"
          }
        ]
      }
    ]
  }
}
```

### 2. 인프라 메트릭

```json
{
  "dashboard": {
    "title": "Infrastructure Metrics",
    "panels": [
      {
        "title": "CPU Usage",
        "type": "graph",
        "targets": [
          {
            "expr": "100 - (avg by (instance) (rate(node_cpu_seconds_total{mode=\"idle\"}[5m])) * 100)",
            "legendFormat": "{{instance}}"
          }
        ]
      },
      {
        "title": "Memory Usage",
        "type": "graph",
        "targets": [
          {
            "expr": "(1 - (node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes)) * 100",
            "legendFormat": "{{instance}}"
          }
        ]
      },
      {
        "title": "Disk I/O",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(node_disk_read_bytes_total[5m])",
            "legendFormat": "Read {{device}}"
          },
          {
            "expr": "rate(node_disk_written_bytes_total[5m])",
            "legendFormat": "Write {{device}}"
          }
        ]
      }
    ]
  }
}
```

## 알림 규칙

### 1. 시스템 알림

```yaml
groups:
  - name: system.rules
    rules:
      - alert: HighCPUUsage
        expr: 100 - (avg by (instance) (rate(node_cpu_seconds_total{mode="idle"}[5m])) * 100) > 80
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High CPU usage on {{ $labels.instance }}"

      - alert: HighMemoryUsage
        expr: (1 - (node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes)) * 100 > 90
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "High memory usage on {{ $labels.instance }}"

      - alert: DiskSpaceLow
        expr: (node_filesystem_avail_bytes / node_filesystem_size_bytes) * 100 < 10
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "Low disk space on {{ $labels.instance }}"
```

### 2. 애플리케이션 알림

```yaml
groups:
  - name: application.rules
    rules:
      - alert: HighErrorRate
        expr: rate(http_requests_total{status=~"5.."}[5m]) / rate(http_requests_total[5m]) > 0.05
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "High error rate detected"

      - alert: SlowResponseTime
        expr: histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m])) > 1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Slow response time detected"
```

## 통합 기능

### 1. 클라우드 모니터링

- AWS CloudWatch 연동
- GCP Monitoring 연동
- Azure Monitor 연동
- 클라우드 네이티브 메트릭

### 2. 컨테이너 모니터링

- Docker 컨테이너 메트릭
- Kubernetes 클러스터 모니터링
- Pod 및 서비스 추적
- 리소스 사용량 분석

### 3. 데이터베이스 모니터링

- MySQL, PostgreSQL 메트릭
- MongoDB, Redis 모니터링
- 쿼리 성능 분석
- 연결 풀 모니터링

## 권장 대안 도구

1. **Prometheus + Grafana**: 오픈소스 모니터링 스택
2. **ELK Stack**: Elasticsearch, Logstash, Kibana
3. **Jaeger**: 분산 추적 시스템
4. **DataDog**: 종합 모니터링 SaaS
5. **New Relic**: APM 및 인프라 모니터링
6. **Splunk**: 로그 분석 및 SIEM
7. **OpenTelemetry**: 관찰성 데이터 수집 표준

## 복원 시 고려사항

- 스토리지 용량 및 성능 요구사항
- 네트워크 대역폭 및 지연시간
- 보안 및 데이터 보호 정책
- 확장성 및 고가용성 설계
- 데이터 보존 및 백업 전략
- 사용자 권한 및 접근 제어
- 비용 최적화 및 리소스 관리
