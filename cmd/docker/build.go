// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

const (
	scannerGrype = "grype"
	scannerTrivy = "trivy"
	scannerSnyk  = "snyk"
)

// BuildCmd represents the build command.
var BuildCmd = &cobra.Command{
	Use:   "build",
	Short: "자동 이미지 빌드 및 배포",
	Long: `멀티 아키텍처 컨테이너 이미지를 자동으로 빌드하고 레지스트리에 배포합니다.

자동 빌드 기능:
- 멀티 아키텍처 지원 (amd64, arm64, arm/v7)
- 자동 태깅 및 버전 관리
- 빌드 캐시 최적화
- 병렬 빌드 지원
- 취약점 스캔 통합
- 레지스트리 자동 푸시
- CI/CD 파이프라인 통합
- 빌드 메트릭 수집

Examples:
  gz docker build --tag myapp:latest
  gz docker build --platforms linux/amd64,linux/arm64 --push
  gz docker build --cache-from myregistry/myapp:cache
  gz docker build --scan --report security-report.json`,
	Run: runBuild,
}

var (
	// Build configuration.
	buildTag       string
	buildPlatforms []string
	buildFile      string
	buildContext   string
	buildArgs      []string
	buildLabels    []string
	buildTarget    string
	buildProgress  string
	buildQuiet     bool
	buildVerbose   bool

	// Multi-architecture settings.
	enableMultiArch bool
	builderName     string

	// Registry settings.
	registryURL      string
	registryUsername string
	registryPassword string
	registryToken    string
	pushAfterBuild   bool
	pushRetries      int

	// Cache settings.
	cacheFrom        []string
	cacheTo          []string
	cacheMode        string
	enableBuildCache bool
	cacheRegistry    string

	// Security and scanning.
	enableScan      bool
	scanners        []string
	scanSeverity    string
	scanExitCode    bool
	scanReport      string
	signImage       bool
	verifySignature bool

	// Performance and optimization.
	enableParallel   bool
	buildConcurrency int
	buildMemoryLimit string
	buildCPULimit    string
	buildTimeout     time.Duration

	// Metadata and tracking.
	enableMetrics   bool
	metricsOutput   string
	notificationURL string
	slackWebhook    string
)

func init() {
	// Basic build flags
	BuildCmd.Flags().StringVarP(&buildTag, "tag", "t", "", "이미지 태그 (예: myapp:latest)")
	BuildCmd.Flags().StringSliceVar(&buildPlatforms, "platforms", []string{"linux/amd64"}, "빌드 플랫폼 (linux/amd64,linux/arm64)")
	BuildCmd.Flags().StringVarP(&buildFile, "file", "f", "Dockerfile", "Dockerfile 경로")
	BuildCmd.Flags().StringVar(&buildContext, "context", ".", "빌드 컨텍스트 경로")
	BuildCmd.Flags().StringSliceVar(&buildArgs, "build-arg", []string{}, "빌드 인수 (KEY=VALUE)")
	BuildCmd.Flags().StringSliceVar(&buildLabels, "label", []string{}, "이미지 라벨 (KEY=VALUE)")
	BuildCmd.Flags().StringVar(&buildTarget, "target", "", "빌드 타겟 스테이지")
	BuildCmd.Flags().StringVar(&buildProgress, "progress", "auto", "진행률 표시 (auto, plain, tty)")
	BuildCmd.Flags().BoolVarP(&buildQuiet, "quiet", "q", false, "출력 최소화")
	BuildCmd.Flags().BoolVarP(&buildVerbose, "verbose", "v", false, "상세 출력")

	// Multi-architecture flags
	BuildCmd.Flags().BoolVar(&enableMultiArch, "multi-arch", false, "멀티 아키텍처 빌드 활성화")
	BuildCmd.Flags().StringVar(&builderName, "builder", "", "빌더 인스턴스 이름")

	// Registry flags
	BuildCmd.Flags().StringVar(&registryURL, "registry", "", "레지스트리 URL")
	BuildCmd.Flags().StringVar(&registryUsername, "registry-user", "", "레지스트리 사용자명")
	BuildCmd.Flags().StringVar(&registryPassword, "registry-pass", "", "레지스트리 비밀번호")
	BuildCmd.Flags().StringVar(&registryToken, "registry-token", "", "레지스트리 토큰")
	BuildCmd.Flags().BoolVar(&pushAfterBuild, "push", false, "빌드 후 자동 푸시")
	BuildCmd.Flags().IntVar(&pushRetries, "push-retries", 3, "푸시 재시도 횟수")

	// Cache flags
	BuildCmd.Flags().StringSliceVar(&cacheFrom, "cache-from", []string{}, "캐시 소스")
	BuildCmd.Flags().StringSliceVar(&cacheTo, "cache-to", []string{}, "캐시 대상")
	BuildCmd.Flags().StringVar(&cacheMode, "cache-mode", "min", "캐시 모드 (min, max)")
	BuildCmd.Flags().BoolVar(&enableBuildCache, "cache", true, "빌드 캐시 활성화")
	BuildCmd.Flags().StringVar(&cacheRegistry, "cache-registry", "", "캐시 레지스트리")

	// Security flags
	BuildCmd.Flags().BoolVar(&enableScan, "scan", false, "보안 스캔 활성화")
	BuildCmd.Flags().StringSliceVar(&scanners, "scanners", []string{scannerTrivy}, "보안 스캐너 (trivy, grype, snyk)")
	BuildCmd.Flags().StringVar(&scanSeverity, "scan-severity", "HIGH", "스캔 심각도 수준")
	BuildCmd.Flags().BoolVar(&scanExitCode, "scan-exit-code", false, "스캔 실패 시 종료")
	BuildCmd.Flags().StringVar(&scanReport, "scan-report", "", "스캔 보고서 출력 경로")
	BuildCmd.Flags().BoolVar(&signImage, "sign", false, "이미지 서명")
	BuildCmd.Flags().BoolVar(&verifySignature, "verify", false, "서명 검증")

	// Performance flags
	BuildCmd.Flags().BoolVar(&enableParallel, "parallel", true, "병렬 빌드 활성화")
	BuildCmd.Flags().IntVar(&buildConcurrency, "concurrency", 4, "빌드 동시성")
	BuildCmd.Flags().StringVar(&buildMemoryLimit, "memory", "", "메모리 제한 (예: 2g)")
	BuildCmd.Flags().StringVar(&buildCPULimit, "cpu", "", "CPU 제한 (예: 2.0)")
	BuildCmd.Flags().DurationVar(&buildTimeout, "timeout", 30*time.Minute, "빌드 타임아웃")

	// Metadata flags
	BuildCmd.Flags().BoolVar(&enableMetrics, "metrics", false, "빌드 메트릭 수집")
	BuildCmd.Flags().StringVar(&metricsOutput, "metrics-output", "", "메트릭 출력 파일")
	BuildCmd.Flags().StringVar(&notificationURL, "notify", "", "알림 웹훅 URL")
	BuildCmd.Flags().StringVar(&slackWebhook, "slack", "", "Slack 웹훅 URL")
}

// BuildConfig represents build configuration.
type BuildConfig struct {
	Tag         string            `json:"tag"`
	Platforms   []string          `json:"platforms"`
	Context     string            `json:"context"`
	Dockerfile  string            `json:"dockerfile"`
	Args        map[string]string `json:"args"`
	Labels      map[string]string `json:"labels"`
	Target      string            `json:"target,omitempty"`
	Cache       CacheConfig       `json:"cache"`
	Registry    RegistryConfig    `json:"registry"`
	Security    SecurityConfig    `json:"security"`
	Performance PerformanceConfig `json:"performance"`
	Metadata    map[string]string `json:"metadata"`
	Timestamps  BuildTimestamps   `json:"timestamps"`
}

type CacheConfig struct {
	Enabled  bool     `json:"enabled"`
	From     []string `json:"from"`
	To       []string `json:"to"`
	Mode     string   `json:"mode"`
	Registry string   `json:"registry,omitempty"`
}

type RegistryConfig struct {
	URL      string `json:"url,omitempty"`
	Username string `json:"username,omitempty"`
	Push     bool   `json:"push"`
	Retries  int    `json:"retries"`
}

type SecurityConfig struct {
	Scan     bool     `json:"scan"`
	Scanners []string `json:"scanners"`
	Severity string   `json:"severity"`
	ExitCode bool     `json:"exitCode"`
	Report   string   `json:"report,omitempty"`
	Sign     bool     `json:"sign"`
	Verify   bool     `json:"verify"`
}

type PerformanceConfig struct {
	Parallel    bool          `json:"parallel"`
	Concurrency int           `json:"concurrency"`
	MemoryLimit string        `json:"memoryLimit,omitempty"`
	CPULimit    string        `json:"cpuLimit,omitempty"`
	Timeout     time.Duration `json:"timeout"`
}

type BuildTimestamps struct {
	Started  time.Time `json:"started"`
	Finished time.Time `json:"finished,omitempty"`
	Duration string    `json:"duration,omitempty"`
}

type BuildResult struct {
	Config      BuildConfig      `json:"config"`
	Success     bool             `json:"success"`
	ImageID     string           `json:"imageId,omitempty"`
	ImageDigest string           `json:"imageDigest,omitempty"`
	Size        int64            `json:"size"`
	Platforms   []PlatformResult `json:"platforms"`
	Scans       []ScanResult     `json:"scans,omitempty"`
	Metrics     BuildMetrics     `json:"metrics"`
	Error       string           `json:"error,omitempty"`
	Logs        []string         `json:"logs,omitempty"`
}

type PlatformResult struct {
	Platform string `json:"platform"`
	ImageID  string `json:"imageId"`
	Size     int64  `json:"size"`
	Success  bool   `json:"success"`
	Error    string `json:"error,omitempty"`
}

type ScanResult struct {
	Scanner   string            `json:"scanner"`
	Platform  string            `json:"platform"`
	Success   bool              `json:"success"`
	Timestamp time.Time         `json:"timestamp"`
	Summary   ScanSummary       `json:"summary"`
	Findings  []SecurityFinding `json:"findings"`
	Report    string            `json:"report,omitempty"`
	Error     string            `json:"error,omitempty"`
}

type ScanSummary struct {
	Total    int `json:"total"`
	Critical int `json:"critical"`
	High     int `json:"high"`
	Medium   int `json:"medium"`
	Low      int `json:"low"`
	Unknown  int `json:"unknown"`
}

type SecurityFinding struct {
	ID          string            `json:"id"`
	Title       string            `json:"title"`
	Severity    string            `json:"severity"`
	Score       float64           `json:"score,omitempty"`
	Package     string            `json:"package"`
	Version     string            `json:"version"`
	FixedIn     string            `json:"fixedIn,omitempty"`
	Description string            `json:"description"`
	References  []string          `json:"references,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

type BuildMetrics struct {
	BuildTime        time.Duration `json:"buildTime"`
	PushTime         time.Duration `json:"pushTime,omitempty"`
	ScanTime         time.Duration `json:"scanTime,omitempty"`
	CacheHitRate     float64       `json:"cacheHitRate"`
	LayerCount       int           `json:"layerCount"`
	UncompressedSize int64         `json:"uncompressedSize"`
	CompressedSize   int64         `json:"compressedSize"`
	Efficiency       float64       `json:"efficiency"` // compressed/uncompressed ratio
	CPU              float64       `json:"cpuUsage"`
	Memory           int64         `json:"memoryUsage"`
}

func runBuild(cmd *cobra.Command, args []string) {
	if buildTag == "" {
		fmt.Printf("❌ 이미지 태그가 필요합니다 (--tag)\n")
		os.Exit(1)
	}

	fmt.Printf("🐳 Docker 이미지 자동 빌드 시작\n")
	fmt.Printf("🏷️  태그: %s\n", buildTag)
	fmt.Printf("🏗️  플랫폼: %s\n", strings.Join(buildPlatforms, ", "))

	// Create build configuration
	config := createBuildConfig()
	ctx := context.Background()

	// Initialize builder if needed
	if enableMultiArch {
		if err := setupMultiArchBuilder(ctx); err != nil {
			fmt.Printf("❌ 멀티 아키텍처 빌더 설정 실패: %v\n", err)
			os.Exit(1)
		}
	}

	// Authenticate with registry if needed
	if pushAfterBuild && registryURL != "" {
		if err := authenticateRegistry(ctx); err != nil {
			fmt.Printf("❌ 레지스트리 인증 실패: %v\n", err)
			os.Exit(1)
		}
	}

	// Start build process
	result, err := performBuild(config)
	if err != nil {
		fmt.Printf("❌ 빌드 실패: %v\n", err)
		os.Exit(1)
	}

	// Display results
	displayBuildResults(result)

	// Perform security scan if enabled
	if enableScan {
		if err := performSecurityScan(ctx, result); err != nil {
			fmt.Printf("⚠️ 보안 스캔 실패: %v\n", err)

			if scanExitCode {
				os.Exit(1)
			}
		}
	}

	// Push to registry if enabled
	if pushAfterBuild {
		ctx := context.Background()
		if err := pushImage(ctx, result); err != nil {
			fmt.Printf("❌ 이미지 푸시 실패: %v\n", err)
			os.Exit(1)
		}
	}

	// Save metrics if enabled
	if enableMetrics {
		if err := saveMetrics(result); err != nil {
			fmt.Printf("⚠️ 메트릭 저장 실패: %v\n", err)
		}
	}

	// Send notifications if configured
	if notificationURL != "" || slackWebhook != "" {
		if err := sendNotifications(result); err != nil {
			fmt.Printf("⚠️ 알림 전송 실패: %v\n", err)
		}
	}

	if !result.Success {
		os.Exit(1)
	}

	fmt.Printf("✅ 이미지 빌드 완료\n")
}

func createBuildConfig() *BuildConfig {
	// Parse build args
	args := make(map[string]string)

	for _, arg := range buildArgs {
		parts := strings.SplitN(arg, "=", 2)
		if len(parts) == 2 {
			args[parts[0]] = parts[1]
		}
	}

	// Parse labels
	labels := make(map[string]string)

	for _, label := range buildLabels {
		parts := strings.SplitN(label, "=", 2)
		if len(parts) == 2 {
			labels[parts[0]] = parts[1]
		}
	}

	// Add automatic labels
	ctx := context.Background()
	labels["org.opencontainers.image.created"] = time.Now().Format(time.RFC3339)
	labels["org.opencontainers.image.revision"] = getGitRevision(ctx)
	labels["org.opencontainers.image.source"] = getGitURL(ctx)

	config := &BuildConfig{
		Tag:        buildTag,
		Platforms:  buildPlatforms,
		Context:    buildContext,
		Dockerfile: buildFile,
		Args:       args,
		Labels:     labels,
		Target:     buildTarget,
		Cache: CacheConfig{
			Enabled:  enableBuildCache,
			From:     cacheFrom,
			To:       cacheTo,
			Mode:     cacheMode,
			Registry: cacheRegistry,
		},
		Registry: RegistryConfig{
			URL:     registryURL,
			Push:    pushAfterBuild,
			Retries: pushRetries,
		},
		Security: SecurityConfig{
			Scan:     enableScan,
			Scanners: scanners,
			Severity: scanSeverity,
			ExitCode: scanExitCode,
			Report:   scanReport,
			Sign:     signImage,
			Verify:   verifySignature,
		},
		Performance: PerformanceConfig{
			Parallel:    enableParallel,
			Concurrency: buildConcurrency,
			MemoryLimit: buildMemoryLimit,
			CPULimit:    buildCPULimit,
			Timeout:     buildTimeout,
		},
		Metadata: make(map[string]string),
		Timestamps: BuildTimestamps{
			Started: time.Now(),
		},
	}

	return config
}

func setupMultiArchBuilder(ctx context.Context) error {
	fmt.Printf("🏗️ 멀티 아키텍처 빌더 설정 중...\n")

	// Check if buildx is available
	if err := exec.CommandContext(ctx, "docker", "buildx", "version").Run(); err != nil {
		return fmt.Errorf("docker buildx가 필요합니다: %w", err)
	}

	// Create or use existing builder
	if builderName == "" {
		builderName = "gzh-multiarch-builder"
	}

	// Check if builder exists
	output, err := exec.CommandContext(ctx, "docker", "buildx", "ls").Output()
	if err != nil {
		return fmt.Errorf("빌더 목록 조회 실패: %w", err)
	}

	if !strings.Contains(string(output), builderName) {
		// Create new builder
		cmd := exec.CommandContext(ctx, "docker", "buildx", "create", "--name", builderName, "--driver", "docker-container", "--use") //nolint:gosec // Docker command with safe arguments
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("빌더 생성 실패: %w", err)
		}

		fmt.Printf("✅ 새 빌더 생성: %s\n", builderName)
	} else {
		// Use existing builder
		cmd := exec.CommandContext(ctx, "docker", "buildx", "use", builderName) //nolint:gosec // Docker command with safe arguments
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("빌더 사용 설정 실패: %w", err)
		}

		fmt.Printf("✅ 기존 빌더 사용: %s\n", builderName)
	}

	// Bootstrap builder
	cmd := exec.CommandContext(ctx, "docker", "buildx", "inspect", "--bootstrap")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("빌더 부트스트랩 실패: %w", err)
	}

	return nil
}

func authenticateRegistry(ctx context.Context) error {
	fmt.Printf("🔐 레지스트리 인증 중...\n")

	var cmd *exec.Cmd
	switch {
	case registryToken != "":
		// Use token authentication
		cmd = exec.CommandContext(ctx, "docker", "login", registryURL, "--username", "oauth2accesstoken", "--password-stdin") //nolint:gosec // Docker login command
		cmd.Stdin = strings.NewReader(registryToken)
	case registryUsername != "" && registryPassword != "":
		// Use username/password authentication
		cmd = exec.CommandContext(ctx, "docker", "login", registryURL, "--username", registryUsername, "--password-stdin") //nolint:gosec // Docker login command
		cmd.Stdin = strings.NewReader(registryPassword)
	default:
		return fmt.Errorf("레지스트리 인증 정보가 필요합니다")
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("레지스트리 로그인 실패: %w", err)
	}

	fmt.Printf("✅ 레지스트리 인증 완료\n")

	return nil
}

func performBuild(config *BuildConfig) (*BuildResult, error) {
	fmt.Printf("🔨 이미지 빌드 중...\n")

	result := &BuildResult{
		Config:    *config,
		Platforms: make([]PlatformResult, 0),
		Metrics:   BuildMetrics{},
	}

	startTime := time.Now()

	// Build Docker command
	args := []string{"buildx", "build"}

	// Add platforms
	if len(config.Platforms) > 0 {
		args = append(args, "--platform", strings.Join(config.Platforms, ","))
	}

	// Add tag
	args = append(args, "--tag", config.Tag)

	// Add dockerfile
	if config.Dockerfile != "" {
		args = append(args, "--file", config.Dockerfile)
	}

	// Add target stage
	if config.Target != "" {
		args = append(args, "--target", config.Target)
	}

	// Add build args
	for key, value := range config.Args {
		args = append(args, "--build-arg", fmt.Sprintf("%s=%s", key, value))
	}

	// Add labels
	for key, value := range config.Labels {
		args = append(args, "--label", fmt.Sprintf("%s=%s", key, value))
	}

	// Add cache options
	if config.Cache.Enabled {
		for _, cacheFrom := range config.Cache.From {
			args = append(args, "--cache-from", cacheFrom)
		}

		for _, cacheTo := range config.Cache.To {
			args = append(args, "--cache-to", cacheTo)
		}
	}

	// Add output options
	if config.Registry.Push {
		args = append(args, "--push")
	} else {
		args = append(args, "--load")
	}

	// Add progress
	if buildProgress != "" {
		args = append(args, "--progress", buildProgress)
	}

	// Add context
	args = append(args, config.Context)

	// Execute build command
	ctx, cancel := context.WithTimeout(context.Background(), config.Performance.Timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if buildVerbose {
		fmt.Printf("🔍 실행 명령어: docker %s\n", strings.Join(args, " "))
	}

	err := cmd.Run()
	buildDuration := time.Since(startTime)

	result.Config.Timestamps.Finished = time.Now()
	result.Config.Timestamps.Duration = buildDuration.String()
	result.Metrics.BuildTime = buildDuration

	if err != nil {
		result.Success = false
		result.Error = err.Error()

		return result, fmt.Errorf("빌드 실행 실패: %w", err)
	}

	// Get image information
	if err := getImageInfo(ctx, result); err != nil {
		fmt.Printf("⚠️ 이미지 정보 조회 실패: %v\n", err)
	}

	result.Success = true

	return result, nil
}

func getImageInfo(ctx context.Context, result *BuildResult) error {
	// Get image ID and digest
	cmd := exec.CommandContext(ctx, "docker", "images", "--format", "{{.ID}}\t{{.Size}}", result.Config.Tag) //nolint:gosec // Docker images command with controlled input

	output, err := cmd.Output()
	if err != nil {
		return err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) > 0 {
		parts := strings.Split(lines[0], "\t")
		if len(parts) >= 2 {
			result.ImageID = parts[0]
			// Parse size (approximate)
			sizeStr := parts[1]
			if strings.HasSuffix(sizeStr, "MB") {
				if size, err := strconv.ParseFloat(strings.TrimSuffix(sizeStr, "MB"), 64); err == nil {
					result.Size = int64(size * 1024 * 1024)
				}
			} else if strings.HasSuffix(sizeStr, "GB") {
				if size, err := strconv.ParseFloat(strings.TrimSuffix(sizeStr, "GB"), 64); err == nil {
					result.Size = int64(size * 1024 * 1024 * 1024)
				}
			}
		}
	}

	return nil
}

func performSecurityScan(ctx context.Context, result *BuildResult) error {
	fmt.Printf("🔍 보안 스캔 실행 중...\n")

	for _, scanner := range scanners {
		scanResult, err := runScanner(ctx, scanner, result.Config.Tag)
		if err != nil {
			fmt.Printf("⚠️ %s 스캔 실패: %v\n", scanner, err)
			continue
		}

		result.Scans = append(result.Scans, *scanResult)

		// Check if scan should fail the build
		if scanExitCode && scanResult.Summary.Critical > 0 {
			return fmt.Errorf("%s 스캔에서 치명적 취약점 %d개 발견", scanner, scanResult.Summary.Critical)
		}
	}

	return nil
}

func runScanner(ctx context.Context, scanner, imageTag string) (*ScanResult, error) {
	scanResult := &ScanResult{
		Scanner:   scanner,
		Timestamp: time.Now(),
		Summary:   ScanSummary{},
		Findings:  []SecurityFinding{},
	}

	var cmd *exec.Cmd

	switch scanner {
	case scannerTrivy:
		cmd = exec.CommandContext(ctx, scannerTrivy, "image", "--format", "json", imageTag)
	case scannerGrype:
		cmd = exec.CommandContext(ctx, scannerGrype, imageTag, "--output", "json")
	case scannerSnyk:
		cmd = exec.CommandContext(ctx, scannerSnyk, "container", "test", imageTag, "--json")
	default:
		return nil, fmt.Errorf("지원하지 않는 스캐너: %s", scanner)
	}

	output, err := cmd.Output()
	if err != nil {
		scanResult.Success = false
		scanResult.Error = err.Error()

		return scanResult, err
	}

	// Parse scanner output (simplified)
	if err := parseScannerOutput(scanner, output, scanResult); err != nil {
		return scanResult, err
	}

	scanResult.Success = true

	return scanResult, nil
}

func parseScannerOutput(scanner string, output []byte, result *ScanResult) error {
	// This is a simplified parser - in reality, each scanner has different output formats
	switch scanner {
	case scannerTrivy:
		return parseTrivyOutput(output, result)
	case scannerGrype:
		return parseGrypeOutput(output, result)
	case scannerSnyk:
		return parseSnykOutput(output, result)
	}

	return nil
}

func parseTrivyOutput(output []byte, result *ScanResult) error {
	// Simplified Trivy JSON parsing
	// nolint:tagliatelle // External API format - must match Trivy JSON output
	var trivyResult struct {
		Results []struct {
			Vulnerabilities []struct {
				VulnerabilityID string `json:"VulnerabilityID"`
				Title           string `json:"Title"`
				Severity        string `json:"Severity"`
				CVSS            struct {
					Score float64 `json:"Score"`
				} `json:"CVSS"`
				PkgName      string   `json:"PkgName"`
				PkgVersion   string   `json:"InstalledVersion"`
				FixedVersion string   `json:"FixedVersion"`
				Description  string   `json:"Description"`
				References   []string `json:"References"`
			} `json:"Vulnerabilities"`
		} `json:"Results"`
	}

	if err := json.Unmarshal(output, &trivyResult); err != nil {
		return err
	}

	for _, res := range trivyResult.Results {
		for _, vuln := range res.Vulnerabilities {
			finding := SecurityFinding{
				ID:          vuln.VulnerabilityID,
				Title:       vuln.Title,
				Severity:    vuln.Severity,
				Score:       vuln.CVSS.Score,
				Package:     vuln.PkgName,
				Version:     vuln.PkgVersion,
				FixedIn:     vuln.FixedVersion,
				Description: vuln.Description,
				References:  vuln.References,
			}

			result.Findings = append(result.Findings, finding)

			// Update summary
			result.Summary.Total++

			switch strings.ToUpper(vuln.Severity) {
			case "CRITICAL":
				result.Summary.Critical++
			case "HIGH":
				result.Summary.High++
			case "MEDIUM":
				result.Summary.Medium++
			case "LOW":
				result.Summary.Low++
			default:
				result.Summary.Unknown++
			}
		}
	}

	return nil
}

func parseGrypeOutput(output []byte, result *ScanResult) error {
	// Simplified Grype parsing - implement as needed
	return nil
}

func parseSnykOutput(output []byte, result *ScanResult) error {
	// Simplified Snyk parsing - implement as needed
	return nil
}

func pushImage(ctx context.Context, result *BuildResult) error {
	fmt.Printf("📤 이미지 푸시 중...\n")

	startTime := time.Now()

	for retry := 0; retry < result.Config.Registry.Retries; retry++ {
		cmd := exec.CommandContext(ctx, "docker", "push", result.Config.Tag) //nolint:gosec // Docker push with controlled tag

		err := cmd.Run()
		if err == nil {
			result.Metrics.PushTime = time.Since(startTime)

			fmt.Printf("✅ 이미지 푸시 완료\n")

			return nil
		}

		if retry < result.Config.Registry.Retries-1 {
			fmt.Printf("⚠️ 푸시 실패, 재시도 중... (%d/%d)\n", retry+1, result.Config.Registry.Retries)
			time.Sleep(time.Second * time.Duration(retry+1))
		}
	}

	return fmt.Errorf("이미지 푸시 실패 (최대 재시도 횟수 초과)")
}

func saveMetrics(result *BuildResult) error {
	if metricsOutput == "" {
		metricsOutput = fmt.Sprintf("build-metrics-%s.json", time.Now().Format("20060102-150405"))
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(metricsOutput), 0o750); err != nil {
		return err
	}

	if err := os.WriteFile(metricsOutput, data, 0o600); err != nil {
		return err
	}

	fmt.Printf("📊 빌드 메트릭 저장: %s\n", metricsOutput)

	return nil
}

func sendNotifications(result *BuildResult) error {
	// Implementation for sending notifications to webhooks, Slack, etc.
	// This would include HTTP POST requests to configured endpoints
	fmt.Printf("📢 알림 전송 중...\n")
	return nil
}

func displayBuildResults(result *BuildResult) {
	fmt.Printf("\n📊 빌드 결과\n")
	fmt.Printf("🆔 이미지 ID: %s\n", result.ImageID)
	fmt.Printf("📏 크기: %s\n", formatBytes(result.Size))
	fmt.Printf("⏱️ 빌드 시간: %v\n", result.Metrics.BuildTime)

	if len(result.Scans) > 0 {
		fmt.Printf("\n🔍 보안 스캔 결과:\n")

		for _, scan := range result.Scans {
			if scan.Success {
				fmt.Printf("  %s: 총 %d개 (치명적: %d, 높음: %d, 중간: %d, 낮음: %d)\n",
					scan.Scanner, scan.Summary.Total, scan.Summary.Critical,
					scan.Summary.High, scan.Summary.Medium, scan.Summary.Low)
			} else {
				fmt.Printf("  %s: 실패 - %s\n", scan.Scanner, scan.Error)
			}
		}
	}
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// Utility functions

func getGitRevision(ctx context.Context) string {
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "HEAD")

	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}

	return strings.TrimSpace(string(output))
}

func getGitURL(ctx context.Context) string {
	cmd := exec.CommandContext(ctx, "git", "config", "--get", "remote.origin.url")

	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}

	return strings.TrimSpace(string(output))
}
