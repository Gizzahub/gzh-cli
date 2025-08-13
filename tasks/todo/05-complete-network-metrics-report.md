# TODO: Network Metrics 리포트 기능 완성

- status: [ ]
- priority: medium (P2)
- category: network-management
- estimated_effort: 2-3시간
- depends_on: []
- spec_reference: `cmd/net-env/network_metrics_cmd.go` (다수 TODO)

## 📋 작업 개요

현재 부분적으로 구현된 Network Metrics 기능의 리포트 생성 부분을 완성합니다. TODO로 표시된 여러 기능들을 구현하여 완전한 네트워크 모니터링 및 분석 도구로 만듭니다.

## 🎯 구현해야 할 TODO 항목들

### 1. **포괄적인 리포트 생성** 
- [ ] **위치**: `cmd/net-env/network_metrics_cmd.go:532`
- [ ] **TODO**: `TODO: Implement comprehensive report generation`
- [ ] **내용**: 전체 네트워크 상태에 대한 종합적인 분석 리포트 생성

### 2. **대역폭 속도 계산**
- [ ] **위치**: `cmd/net-env/network_metrics_cmd.go:712` 
- [ ] **TODO**: `TODO: Calculate actual bandwidth rates by comparing with previous measurements`
- [ ] **내용**: 이전 측정값과 비교하여 실제 대역폭 사용률 계산

### 3. **인터페이스 속도 및 사용률 계산**
- [ ] **위치**: `cmd/net-env/network_metrics_cmd.go:801`
- [ ] **TODO**: `TODO: Get interface speed and calculate actual utilization`
- [ ] **내용**: 네트워크 인터페이스 최대 속도 확인 및 실제 사용률 계산

### 4. **HTML 리포트 생성**
- [ ] **위치**: `cmd/net-env/network_metrics_cmd.go:1019`
- [ ] **TODO**: `TODO: Implement HTML report generation`
- [ ] **내용**: 웹 브라우저에서 볼 수 있는 HTML 형식의 리포트

### 5. **텍스트 리포트 생성**
- [ ] **위치**: `cmd/net-env/network_metrics_cmd.go:1023`
- [ ] **TODO**: `TODO: Implement text report generation`  
- [ ] **내용**: 콘솔이나 파일로 출력 가능한 텍스트 형식 리포트

## 🔧 기술적 구현

### 1. 포괄적인 리포트 생성
```go
type ComprehensiveNetworkReport struct {
    Summary         NetworkSummary       `json:"summary"`
    Interfaces      []InterfaceReport    `json:"interfaces"`
    BandwidthUsage  BandwidthReport      `json:"bandwidth_usage"`
    LatencyMetrics  LatencyReport        `json:"latency_metrics"`
    Recommendations []string             `json:"recommendations"`
    Timestamp       time.Time            `json:"timestamp"`
    Duration        time.Duration        `json:"duration"`
}

type NetworkSummary struct {
    TotalInterfaces   int     `json:"total_interfaces"`
    ActiveInterfaces  int     `json:"active_interfaces"`
    TotalBandwidth    uint64  `json:"total_bandwidth_bps"`
    UsedBandwidth     uint64  `json:"used_bandwidth_bps"`
    UtilizationPercent float64 `json:"utilization_percent"`
    AverageLatency    float64 `json:"average_latency_ms"`
    PacketLossPercent float64 `json:"packet_loss_percent"`
}

func (cmd *networkMetricsCmd) generateComprehensiveReport() (*ComprehensiveNetworkReport, error) {
    interfaces, err := cmd.collectInterfaceMetrics()
    if err != nil {
        return nil, err
    }
    
    bandwidth, err := cmd.calculateBandwidthUsage(interfaces)
    if err != nil {
        return nil, err
    }
    
    latency, err := cmd.measureLatencyMetrics()
    if err != nil {
        return nil, err
    }
    
    summary := cmd.generateSummary(interfaces, bandwidth, latency)
    recommendations := cmd.generateRecommendations(summary, interfaces)
    
    return &ComprehensiveNetworkReport{
        Summary:         summary,
        Interfaces:      interfaces,
        BandwidthUsage:  bandwidth, 
        LatencyMetrics:  latency,
        Recommendations: recommendations,
        Timestamp:       time.Now(),
        Duration:        cmd.measurementDuration,
    }, nil
}
```

### 2. 대역폭 속도 계산
```go
type BandwidthCalculator struct {
    previousMeasurements map[string]*InterfaceMetrics
    measurementInterval  time.Duration
}

type BandwidthRate struct {
    Interface     string  `json:"interface"`
    RxBytesPerSec uint64  `json:"rx_bytes_per_sec"`
    TxBytesPerSec uint64  `json:"tx_bytes_per_sec"`
    RxPacketsPerSec uint64 `json:"rx_packets_per_sec"`
    TxPacketsPerSec uint64 `json:"tx_packets_per_sec"`
    Utilization   float64 `json:"utilization_percent"`
}

func (bc *BandwidthCalculator) CalculateRates(current map[string]*InterfaceMetrics) []BandwidthRate {
    var rates []BandwidthRate
    
    for interfaceName, currentMetrics := range current {
        if prev, exists := bc.previousMeasurements[interfaceName]; exists {
            timeDiff := currentMetrics.Timestamp.Sub(prev.Timestamp).Seconds()
            
            if timeDiff > 0 {
                rxRate := uint64(float64(currentMetrics.RxBytes-prev.RxBytes) / timeDiff)
                txRate := uint64(float64(currentMetrics.TxBytes-prev.TxBytes) / timeDiff)
                
                // 인터페이스 최대 속도 대비 사용률 계산
                maxSpeed := bc.getInterfaceMaxSpeed(interfaceName)
                utilization := float64(rxRate+txRate) / float64(maxSpeed) * 100
                
                rates = append(rates, BandwidthRate{
                    Interface:       interfaceName,
                    RxBytesPerSec:   rxRate,
                    TxBytesPerSec:   txRate,
                    RxPacketsPerSec: uint64(float64(currentMetrics.RxPackets-prev.RxPackets) / timeDiff),
                    TxPacketsPerSec: uint64(float64(currentMetrics.TxPackets-prev.TxPackets) / timeDiff),
                    Utilization:     utilization,
                })
            }
        }
    }
    
    // 현재 측정값을 다음 계산을 위해 저장
    bc.previousMeasurements = current
    
    return rates
}
```

### 3. 인터페이스 속도 및 사용률 계산
```go
type InterfaceSpeedDetector struct {
    cache map[string]uint64 // 인터페이스별 최대 속도 캐시
}

func (isd *InterfaceSpeedDetector) GetInterfaceMaxSpeed(interfaceName string) (uint64, error) {
    // 캐시에서 먼저 확인
    if speed, exists := isd.cache[interfaceName]; exists {
        return speed, nil
    }
    
    // Linux: /sys/class/net/{interface}/speed 파일에서 읽기
    speedFile := fmt.Sprintf("/sys/class/net/%s/speed", interfaceName)
    if data, err := os.ReadFile(speedFile); err == nil {
        if speed, err := strconv.ParseUint(strings.TrimSpace(string(data)), 10, 64); err == nil {
            // Mbps를 bps로 변환
            speedBps := speed * 1000000
            isd.cache[interfaceName] = speedBps
            return speedBps, nil
        }
    }
    
    // ethtool을 사용한 속도 감지
    cmd := exec.Command("ethtool", interfaceName)
    output, err := cmd.Output()
    if err == nil {
        return isd.parseEthtoolOutput(string(output))
    }
    
    // 기본값으로 인터페이스 타입에 따른 추정
    return isd.estimateSpeedByType(interfaceName), nil
}

func (isd *InterfaceSpeedDetector) CalculateUtilization(interfaceName string, rxRate, txRate uint64) (float64, error) {
    maxSpeed, err := isd.GetInterfaceMaxSpeed(interfaceName)
    if err != nil {
        return 0, err
    }
    
    totalRate := rxRate + txRate
    utilization := float64(totalRate) / float64(maxSpeed) * 100
    
    // 100% 초과 방지 (측정 오차 고려)
    if utilization > 100 {
        utilization = 100
    }
    
    return utilization, nil
}
```

### 4. HTML 리포트 생성
```go
type HTMLReportGenerator struct {
    templateEngine *template.Template
    outputDir      string
}

const htmlReportTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>Network Metrics Report</title>
    <meta charset="utf-8">
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .header { background: #f5f5f5; padding: 20px; border-radius: 5px; }
        .summary { display: grid; grid-template-columns: repeat(3, 1fr); gap: 20px; margin: 20px 0; }
        .metric-card { background: white; border: 1px solid #ddd; padding: 15px; border-radius: 5px; }
        .interface-table { width: 100%; border-collapse: collapse; margin: 20px 0; }
        .interface-table th, .interface-table td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        .chart { margin: 20px 0; height: 300px; }
        .recommendations { background: #e8f4f8; padding: 15px; border-radius: 5px; }
    </style>
    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
</head>
<body>
    <div class="header">
        <h1>🌐 Network Metrics Report</h1>
        <p>Generated: {{.Timestamp.Format "2006-01-02 15:04:05"}}</p>
        <p>Duration: {{.Duration}}</p>
    </div>
    
    <div class="summary">
        <div class="metric-card">
            <h3>📊 Interface Summary</h3>
            <p>Total: {{.Summary.TotalInterfaces}}</p>
            <p>Active: {{.Summary.ActiveInterfaces}}</p>
        </div>
        <div class="metric-card">
            <h3>📈 Bandwidth Usage</h3>
            <p>Total: {{.Summary.TotalBandwidth | formatBytes}}/s</p>
            <p>Used: {{.Summary.UsedBandwidth | formatBytes}}/s</p>
            <p>Utilization: {{.Summary.UtilizationPercent | printf "%.1f"}}%</p>
        </div>
        <div class="metric-card">
            <h3>⚡ Latency</h3>
            <p>Average: {{.Summary.AverageLatency | printf "%.2f"}} ms</p>
            <p>Packet Loss: {{.Summary.PacketLossPercent | printf "%.2f"}}%</p>
        </div>
    </div>
    
    <div class="chart">
        <canvas id="bandwidthChart"></canvas>
    </div>
    
    <table class="interface-table">
        <thead>
            <tr>
                <th>Interface</th>
                <th>Status</th>
                <th>Speed</th>
                <th>RX Rate</th>
                <th>TX Rate</th>
                <th>Utilization</th>
            </tr>
        </thead>
        <tbody>
            {{range .Interfaces}}
            <tr>
                <td>{{.Name}}</td>
                <td>{{.Status}}</td>
                <td>{{.MaxSpeed | formatBytes}}/s</td>
                <td>{{.RxRate | formatBytes}}/s</td>
                <td>{{.TxRate | formatBytes}}/s</td>
                <td>{{.Utilization | printf "%.1f"}}%</td>
            </tr>
            {{end}}
        </tbody>
    </table>
    
    {{if .Recommendations}}
    <div class="recommendations">
        <h3>💡 Recommendations</h3>
        <ul>
        {{range .Recommendations}}
            <li>{{.}}</li>
        {{end}}
        </ul>
    </div>
    {{end}}
    
    <script>
        // 대역폭 차트 생성
        const ctx = document.getElementById('bandwidthChart').getContext('2d');
        const chart = new Chart(ctx, {
            type: 'line',
            data: {
                labels: [{{range .BandwidthHistory}}'{{.Timestamp.Format "15:04:05"}}',{{end}}],
                datasets: [{
                    label: 'RX Rate',
                    data: [{{range .BandwidthHistory}}{{.RxRate}},{{end}}],
                    borderColor: 'rgb(75, 192, 192)',
                    tension: 0.1
                }, {
                    label: 'TX Rate', 
                    data: [{{range .BandwidthHistory}}{{.TxRate}},{{end}}],
                    borderColor: 'rgb(255, 99, 132)',
                    tension: 0.1
                }]
            },
            options: {
                responsive: true,
                scales: {
                    y: {
                        beginAtZero: true,
                        title: {
                            display: true,
                            text: 'Bytes/sec'
                        }
                    }
                }
            }
        });
    </script>
</body>
</html>
`

func (hrg *HTMLReportGenerator) GenerateReport(report *ComprehensiveNetworkReport, filename string) error {
    tmpl, err := template.New("report").Funcs(template.FuncMap{
        "formatBytes": hrg.formatBytes,
        "printf":      fmt.Sprintf,
    }).Parse(htmlReportTemplate)
    
    if err != nil {
        return fmt.Errorf("failed to parse template: %w", err)
    }
    
    outputPath := filepath.Join(hrg.outputDir, filename)
    file, err := os.Create(outputPath)
    if err != nil {
        return fmt.Errorf("failed to create report file: %w", err)
    }
    defer file.Close()
    
    return tmpl.Execute(file, report)
}
```

### 5. 텍스트 리포트 생성
```go
type TextReportGenerator struct {
    outputDir string
}

func (trg *TextReportGenerator) GenerateReport(report *ComprehensiveNetworkReport, filename string) error {
    var buf strings.Builder
    
    // 헤더
    buf.WriteString("📊 NETWORK METRICS REPORT\n")
    buf.WriteString(strings.Repeat("=", 50) + "\n\n")
    buf.WriteString(fmt.Sprintf("Generated: %s\n", report.Timestamp.Format("2006-01-02 15:04:05")))
    buf.WriteString(fmt.Sprintf("Duration: %v\n\n", report.Duration))
    
    // 요약
    buf.WriteString("📈 SUMMARY\n")
    buf.WriteString(strings.Repeat("-", 20) + "\n")
    buf.WriteString(fmt.Sprintf("Interfaces: %d total, %d active\n", 
        report.Summary.TotalInterfaces, report.Summary.ActiveInterfaces))
    buf.WriteString(fmt.Sprintf("Bandwidth: %s total, %s used (%.1f%% utilization)\n",
        trg.formatBytes(report.Summary.TotalBandwidth),
        trg.formatBytes(report.Summary.UsedBandwidth),
        report.Summary.UtilizationPercent))
    buf.WriteString(fmt.Sprintf("Latency: %.2f ms average, %.2f%% packet loss\n\n",
        report.Summary.AverageLatency, report.Summary.PacketLossPercent))
    
    // 인터페이스 상세 정보
    buf.WriteString("🔌 INTERFACE DETAILS\n")
    buf.WriteString(strings.Repeat("-", 20) + "\n")
    buf.WriteString(fmt.Sprintf("%-12s %-8s %-12s %-12s %-12s %-8s\n",
        "Interface", "Status", "Speed", "RX Rate", "TX Rate", "Usage%"))
    buf.WriteString(strings.Repeat("-", 70) + "\n")
    
    for _, iface := range report.Interfaces {
        buf.WriteString(fmt.Sprintf("%-12s %-8s %-12s %-12s %-12s %-8.1f\n",
            iface.Name,
            iface.Status,
            trg.formatBytes(iface.MaxSpeed)+"/s",
            trg.formatBytes(iface.RxRate)+"/s", 
            trg.formatBytes(iface.TxRate)+"/s",
            iface.Utilization))
    }
    buf.WriteString("\n")
    
    // 권장사항
    if len(report.Recommendations) > 0 {
        buf.WriteString("💡 RECOMMENDATIONS\n")
        buf.WriteString(strings.Repeat("-", 20) + "\n")
        for i, rec := range report.Recommendations {
            buf.WriteString(fmt.Sprintf("%d. %s\n", i+1, rec))
        }
        buf.WriteString("\n")
    }
    
    // 파일 저장
    outputPath := filepath.Join(trg.outputDir, filename)
    return os.WriteFile(outputPath, []byte(buf.String()), 0644)
}
```

## 📁 관련 파일들

### 수정할 파일
- `cmd/net-env/network_metrics_cmd.go` - 모든 TODO 구현

### 새로 생성할 파일
- `internal/netenv/reports/comprehensive.go` - 종합 리포트 생성
- `internal/netenv/reports/bandwidth.go` - 대역폭 계산 로직
- `internal/netenv/reports/interface_speed.go` - 인터페이스 속도 감지
- `internal/netenv/reports/html_generator.go` - HTML 리포트 생성
- `internal/netenv/reports/text_generator.go` - 텍스트 리포트 생성
- `internal/netenv/reports/templates/` - 리포트 템플릿 파일들

## 🎯 명령어 확장

```bash
# 기존 명령어에 리포트 생성 옵션 추가
gz net-env metrics --report-format html --output network-report.html
gz net-env metrics --report-format text --output network-report.txt
gz net-env metrics --report-format json --output network-report.json

# 리포트에 포함할 메트릭 선택
gz net-env metrics --report-format html --include bandwidth,latency,interfaces

# 측정 시간 지정
gz net-env metrics --duration 5m --report-format html

# 자동으로 브라우저에서 열기
gz net-env metrics --report-format html --open
```

## 🧪 테스트 요구사항

### 1. 단위 테스트
```go
func TestBandwidthCalculator_CalculateRates(t *testing.T) {
    // 대역폭 계산 정확성 테스트
}

func TestInterfaceSpeedDetector_GetMaxSpeed(t *testing.T) {
    // 인터페이스 속도 감지 테스트  
}

func TestHTMLReportGenerator_GenerateReport(t *testing.T) {
    // HTML 리포트 생성 테스트
}
```

### 2. 통합 테스트
```bash
# 실제 네트워크 환경에서 리포트 생성 테스트
go test ./internal/netenv/reports -tags=integration
```

## ✅ 완료 기준

### 기능 완성도
- [ ] 5개 TODO 항목 모두 구현
- [ ] HTML, 텍스트, JSON 리포트 형식 지원
- [ ] 실시간 대역폭 계산 정확성
- [ ] 인터페이스 속도 자동 감지

### 사용자 경험
- [ ] 직관적인 리포트 레이아웃
- [ ] 차트와 그래프로 시각화
- [ ] 실행 가능한 권장사항 제공
- [ ] 다양한 출력 형식 지원

## 🚀 커밋 메시지 가이드

```
feat(claude-opus): Network Metrics 리포트 기능 완성

- 포괄적인 네트워크 상태 분석 리포트 구현
- 실시간 대역폭 사용률 계산 및 히스토리 추적
- 인터페이스별 최대 속도 자동 감지 및 사용률 계산
- HTML/텍스트/JSON 형식 리포트 생성 지원
- Chart.js 기반 시각화 및 권장사항 시스템

Resolves: cmd/net-env/network_metrics_cmd.go TODO items
- L532: comprehensive report generation
- L712: bandwidth rate calculations  
- L801: interface speed detection
- L1019: HTML report generation
- L1023: text report generation

🤖 Generated with [Claude Code](https://claude.ai/code)

Co-Authored-By: Claude <noreply@anthropic.com>
```

## 💡 구현 힌트

1. **점진적 구현**: 텍스트 리포트부터 구현 후 HTML 추가
2. **성능 최적화**: 대역폭 계산은 별도 고루틴에서 처리
3. **크로스 플랫폼**: Linux, macOS, Windows 각각의 네트워크 정보 수집 방법
4. **차트 라이브러리**: Chart.js 등 경량 라이브러리 활용

## ⚠️ 주의사항

- 네트워크 인터페이스 정보 수집 시 권한 필요할 수 있음
- 대역폭 계산은 측정 간격에 따라 정확도 차이 발생
- HTML 리포트는 외부 CDN 의존성 있음 (Chart.js)
- 일부 가상 네트워크 인터페이스는 속도 정보 없을 수 있음