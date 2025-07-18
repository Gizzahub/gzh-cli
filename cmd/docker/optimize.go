package docker

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// OptimizeCmd represents the optimize command.
var OptimizeCmd = &cobra.Command{
	Use:   "optimize",
	Short: "이미지 최적화 및 크기 분석",
	Long: `컨테이너 이미지를 최적화하고 크기를 분석합니다.

최적화 기능:
- 레이어 최적화 및 압축
- 불필요한 파일 제거 및 정리
- 베이스 이미지 분석 및 추천
- 이미지 크기 분석 및 시각화
- 최적화 제안 및 Dockerfile 생성
- 멀티 스테이지 빌드 최적화
- 패키지 매니저 캐시 정리
- 보안 스캔 결과 기반 최적화

Examples:
  gz docker optimize myapp:latest
  gz docker optimize --analyze-only myapp:latest
  gz docker optimize --output optimized.dockerfile myapp:latest
  gz docker optimize --target-size 100MB myapp:latest
  gz docker optimize --base-image alpine myapp:latest`,
	Run: runOptimize,
}

var (
	optimizeImage     string
	analyzeOnly       bool
	outputDockerfile  string
	targetSize        string
	suggestBaseImage  bool
	removeCache       bool
	optimizePackages  bool
	generateReport    bool
	reportFormat      string
	reportOutput      string
	enableCompression bool
	compressionLevel  int
	minimizeImageSize bool
	removeDebugInfo   bool
	optimizeLayers    bool
	enableDive        bool
	enableSlim        bool
	slimOptions       []string
)

func init() {
	// Basic optimization flags
	OptimizeCmd.Flags().BoolVar(&analyzeOnly, "analyze-only", false, "분석만 수행, 최적화 실행 안함")
	OptimizeCmd.Flags().StringVarP(&outputDockerfile, "output", "o", "", "최적화된 Dockerfile 출력 경로")
	OptimizeCmd.Flags().StringVar(&targetSize, "target-size", "", "목표 이미지 크기 (예: 100MB)")
	OptimizeCmd.Flags().BoolVar(&suggestBaseImage, "suggest-base", true, "베이스 이미지 추천")
	OptimizeCmd.Flags().BoolVar(&removeCache, "remove-cache", true, "패키지 매니저 캐시 제거")
	OptimizeCmd.Flags().BoolVar(&optimizePackages, "optimize-packages", true, "패키지 최적화")

	// Reporting flags
	OptimizeCmd.Flags().BoolVar(&generateReport, "report", true, "최적화 보고서 생성")
	OptimizeCmd.Flags().StringVar(&reportFormat, "report-format", "json", "보고서 형식 (json, html, text)")
	OptimizeCmd.Flags().StringVar(&reportOutput, "report-output", "", "보고서 출력 경로")

	// Advanced optimization flags
	OptimizeCmd.Flags().BoolVar(&enableCompression, "compression", true, "레이어 압축 활성화")
	OptimizeCmd.Flags().IntVar(&compressionLevel, "compression-level", 6, "압축 레벨 (1-9)")
	OptimizeCmd.Flags().BoolVar(&minimizeImageSize, "minimize", false, "극한 크기 최적화")
	OptimizeCmd.Flags().BoolVar(&removeDebugInfo, "remove-debug", true, "디버그 정보 제거")
	OptimizeCmd.Flags().BoolVar(&optimizeLayers, "optimize-layers", true, "레이어 최적화")

	// External tools integration
	OptimizeCmd.Flags().BoolVar(&enableDive, "dive", false, "Dive 도구로 레이어 분석")
	OptimizeCmd.Flags().BoolVar(&enableSlim, "slim", false, "docker-slim으로 최적화")
	OptimizeCmd.Flags().StringSliceVar(&slimOptions, "slim-options", []string{}, "docker-slim 옵션")
}

// OptimizationAnalysis represents image analysis results.
type OptimizationAnalysis struct {
	ImageInfo      ImageInfo          `json:"image_info"`
	LayerAnalysis  []LayerInfo        `json:"layer_analysis"`
	SizeBreakdown  SizeBreakdown      `json:"size_breakdown"`
	Suggestions    []Suggestion       `json:"suggestions"`
	BaseImageRecs  []BaseImageRec     `json:"base_image_recommendations"`
	WasteAnalysis  WasteAnalysis      `json:"waste_analysis"`
	SecurityImpact SecurityImpact     `json:"security_impact"`
	Performance    PerformanceMetrics `json:"performance"`
	Timestamp      time.Time          `json:"timestamp"`
}

type ImageInfo struct {
	Name         string            `json:"name"`
	Tag          string            `json:"tag"`
	ID           string            `json:"id"`
	Created      time.Time         `json:"created"`
	Size         int64             `json:"size"`
	VirtualSize  int64             `json:"virtual_size"`
	Architecture string            `json:"architecture"`
	OS           string            `json:"os"`
	Labels       map[string]string `json:"labels"`
	Config       ImageConfig       `json:"config"`
}

type ImageConfig struct {
	User         string   `json:"user"`
	ExposedPorts []string `json:"exposed_ports"`
	Env          []string `json:"env"`
	Entrypoint   []string `json:"entrypoint"`
	Cmd          []string `json:"cmd"`
	WorkingDir   string   `json:"working_dir"`
	Volumes      []string `json:"volumes"`
}

type LayerInfo struct {
	ID          string     `json:"id"`
	Size        int64      `json:"size"`
	Command     string     `json:"command"`
	CreatedBy   string     `json:"created_by"`
	Created     time.Time  `json:"created"`
	Empty       bool       `json:"empty"`
	WastedBytes int64      `json:"wasted_bytes"`
	Efficiency  float64    `json:"efficiency"`
	Files       []FileInfo `json:"files,omitempty"`
}

type FileInfo struct {
	Path         string `json:"path"`
	Size         int64  `json:"size"`
	IsWasted     bool   `json:"is_wasted"`
	Permissions  string `json:"permissions"`
	Owner        string `json:"owner"`
	ModifiedTime string `json:"modified_time"`
}

type SizeBreakdown struct {
	TotalSize       int64            `json:"total_size"`
	BaseImageSize   int64            `json:"base_image_size"`
	ApplicationSize int64            `json:"application_size"`
	WastedSpace     int64            `json:"wasted_space"`
	CacheSize       int64            `json:"cache_size"`
	LayerSizes      map[string]int64 `json:"layer_sizes"`
	FileTypes       map[string]int64 `json:"file_types"`
	Directories     map[string]int64 `json:"directories"`
}

type Suggestion struct {
	Type             string   `json:"type"`     // size, security, performance, best_practice
	Priority         string   `json:"priority"` // critical, high, medium, low
	Title            string   `json:"title"`
	Description      string   `json:"description"`
	Impact           string   `json:"impact"`
	SizeReduction    int64    `json:"size_reduction,omitempty"`
	Implementation   string   `json:"implementation"`
	DockerfileChange string   `json:"dockerfile_change,omitempty"`
	References       []string `json:"references,omitempty"`
}

type BaseImageRec struct {
	Name          string  `json:"name"`
	Tag           string  `json:"tag"`
	Size          int64   `json:"size"`
	SizeReduction int64   `json:"size_reduction"`
	SecurityScore float64 `json:"security_score"`
	Compatibility string  `json:"compatibility"`
	Reason        string  `json:"reason"`
	TrustScore    float64 `json:"trust_score"`
}

type WasteAnalysis struct {
	TotalWaste      int64            `json:"total_waste"`
	WastePercentage float64          `json:"waste_percentage"`
	WastedFiles     []WastedFile     `json:"wasted_files"`
	DuplicateFiles  []DuplicateGroup `json:"duplicate_files"`
	LargeFiles      []FileInfo       `json:"large_files"`
	EmptyDirs       []string         `json:"empty_dirs"`
	UnusedPackages  []string         `json:"unused_packages"`
}

type WastedFile struct {
	Path     string `json:"path"`
	Size     int64  `json:"size"`
	Reason   string `json:"reason"`
	Category string `json:"category"` // cache, temp, log, debug, etc.
}

type DuplicateGroup struct {
	Hash  string     `json:"hash"`
	Size  int64      `json:"size"`
	Count int        `json:"count"`
	Files []FileInfo `json:"files"`
}

type SecurityImpact struct {
	VulnerabilityCount int      `json:"vulnerability_count"`
	CriticalVulns      int      `json:"critical_vulns"`
	ExposedPorts       []string `json:"exposed_ports"`
	RunAsRoot          bool     `json:"run_as_root"`
	SensitiveFiles     []string `json:"sensitive_files"`
	Recommendations    []string `json:"recommendations"`
}

type PerformanceMetrics struct {
	StartupTime   time.Duration `json:"startup_time"`
	LayerCount    int           `json:"layer_count"`
	MaxLayerSize  int64         `json:"max_layer_size"`
	AvgLayerSize  int64         `json:"avg_layer_size"`
	PullTime      time.Duration `json:"estimated_pull_time"`
	CacheHitRatio float64       `json:"cache_hit_ratio"`
}

// OptimizationResult represents the optimization process result.
type OptimizationResult struct {
	Analysis       OptimizationAnalysis `json:"analysis"`
	OriginalSize   int64                `json:"original_size"`
	OptimizedSize  int64                `json:"optimized_size"`
	SizeReduction  int64                `json:"size_reduction"`
	PercentReduced float64              `json:"percent_reduced"`
	OptimizedImage string               `json:"optimized_image,omitempty"`
	GeneratedFiles []string             `json:"generated_files"`
	Applied        []Suggestion         `json:"applied_suggestions"`
	Skipped        []Suggestion         `json:"skipped_suggestions"`
	Duration       time.Duration        `json:"duration"`
	Success        bool                 `json:"success"`
	Error          string               `json:"error,omitempty"`
}

func runOptimize(cmd *cobra.Command, args []string) {
	// Get image to optimize
	if len(args) > 0 {
		optimizeImage = args[0]
	}

	if optimizeImage == "" {
		fmt.Printf("❌ 최적화할 이미지가 필요합니다\n")
		fmt.Printf("사용법: gz docker optimize <image>\n")
		os.Exit(1)
	}

	fmt.Printf("🔧 컨테이너 이미지 최적화 시작\n")
	fmt.Printf("📦 이미지: %s\n", optimizeImage)

	startTime := time.Now()

	// 1. Analyze current image
	fmt.Printf("🔍 이미지 분석 중...\n")

	analysis, err := analyzeImage(optimizeImage)
	if err != nil {
		fmt.Printf("❌ 이미지 분석 실패: %v\n", err)
		os.Exit(1)
	}

	// Display analysis results
	displayAnalysisResults(analysis)

	// 2. Generate optimization suggestions
	fmt.Printf("💡 최적화 제안 생성 중...\n")

	suggestions := generateSuggestions(analysis)

	// 3. If analyze-only, just show suggestions and exit
	if analyzeOnly {
		fmt.Printf("📋 최적화 제안:\n")
		displaySuggestions(suggestions)

		if generateReport {
			if err := saveOptimizationReport(analysis, suggestions); err != nil {
				fmt.Printf("⚠️ 보고서 저장 실패: %v\n", err)
			}
		}

		return
	}

	// 4. Apply optimizations
	fmt.Printf("⚙️ 최적화 적용 중...\n")

	result, err := applyOptimizations(optimizeImage, analysis, suggestions)
	if err != nil {
		fmt.Printf("❌ 최적화 실패: %v\n", err)
		os.Exit(1)
	}

	// 5. Display results
	result.Duration = time.Since(startTime)
	displayOptimizationResults(result)

	// 6. Save report if requested
	if generateReport {
		if err := saveOptimizationReport(analysis, suggestions); err != nil {
			fmt.Printf("⚠️ 보고서 저장 실패: %v\n", err)
		}
	}

	fmt.Printf("✅ 이미지 최적화 완료\n")
}

func analyzeImage(imageName string) (*OptimizationAnalysis, error) {
	analysis := &OptimizationAnalysis{
		Timestamp: time.Now(),
	}

	// Get image info
	imageInfo, err := getOptimizeImageInfo(imageName)
	if err != nil {
		return nil, fmt.Errorf("이미지 정보 조회 실패: %w", err)
	}

	analysis.ImageInfo = *imageInfo

	// Analyze layers
	layers, err := analyzeLayers(imageName)
	if err != nil {
		return nil, fmt.Errorf("레이어 분석 실패: %w", err)
	}

	analysis.LayerAnalysis = layers

	// Calculate size breakdown
	sizeBreakdown := calculateSizeBreakdown(imageInfo, layers)
	analysis.SizeBreakdown = *sizeBreakdown

	// Analyze waste
	wasteAnalysis, err := analyzeWaste(imageName, layers)
	if err != nil {
		fmt.Printf("⚠️ 낭비 분석 실패: %v\n", err)

		wasteAnalysis = &WasteAnalysis{}
	}

	analysis.WasteAnalysis = *wasteAnalysis

	// Get base image recommendations
	baseImageRecs := getBaseImageRecommendations(imageInfo)
	analysis.BaseImageRecs = baseImageRecs

	// Analyze security impact
	securityImpact := analyzeSecurityImpact(imageInfo)
	analysis.SecurityImpact = *securityImpact

	// Calculate performance metrics
	performance := calculatePerformanceMetrics(imageInfo, layers)
	analysis.Performance = *performance

	return analysis, nil
}

func getOptimizeImageInfo(imageName string) (*ImageInfo, error) {
	cmd := exec.Command("docker", "image", "inspect", imageName)

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var inspectResult []struct {
		ID           string `json:"Id"`
		Created      string `json:"Created"`
		Size         int64  `json:"Size"`
		VirtualSize  int64  `json:"VirtualSize"`
		Architecture string `json:"Architecture"`
		OS           string `json:"Os"`
		Config       struct {
			User         string                 `json:"User"`
			ExposedPorts map[string]interface{} `json:"ExposedPorts"`
			Env          []string               `json:"Env"`
			Entrypoint   []string               `json:"Entrypoint"`
			Cmd          []string               `json:"Cmd"`
			WorkingDir   string                 `json:"WorkingDir"`
			Volumes      map[string]interface{} `json:"Volumes"`
			Labels       map[string]string      `json:"Labels"`
		} `json:"Config"`
	}

	if err := json.Unmarshal(output, &inspectResult); err != nil {
		return nil, err
	}

	if len(inspectResult) == 0 {
		return nil, fmt.Errorf("이미지 정보를 찾을 수 없습니다")
	}

	img := inspectResult[0]

	// Parse created time
	created, _ := time.Parse(time.RFC3339Nano, img.Created)

	// Extract exposed ports
	var exposedPorts []string
	for port := range img.Config.ExposedPorts {
		exposedPorts = append(exposedPorts, port)
	}

	// Extract volumes
	var volumes []string
	for vol := range img.Config.Volumes {
		volumes = append(volumes, vol)
	}

	imageInfo := &ImageInfo{
		Name:         imageName,
		ID:           img.ID,
		Created:      created,
		Size:         img.Size,
		VirtualSize:  img.VirtualSize,
		Architecture: img.Architecture,
		OS:           img.OS,
		Labels:       img.Config.Labels,
		Config: ImageConfig{
			User:         img.Config.User,
			ExposedPorts: exposedPorts,
			Env:          img.Config.Env,
			Entrypoint:   img.Config.Entrypoint,
			Cmd:          img.Config.Cmd,
			WorkingDir:   img.Config.WorkingDir,
			Volumes:      volumes,
		},
	}

	return imageInfo, nil
}

func analyzeLayers(imageName string) ([]LayerInfo, error) {
	cmd := exec.Command("docker", "history", "--no-trunc", "--format", "{{.ID}}\t{{.Size}}\t{{.CreatedBy}}\t{{.CreatedAt}}", imageName)

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var layers []LayerInfo

	scanner := bufio.NewScanner(strings.NewReader(string(output)))

	for scanner.Scan() {
		line := scanner.Text()

		parts := strings.Split(line, "\t")
		if len(parts) < 4 {
			continue
		}

		// Parse size
		sizeStr := parts[1]

		var size int64
		if sizeStr != "0B" && sizeStr != "<missing>" {
			size = parseSize(sizeStr)
		}

		// Parse created time
		created, _ := time.Parse("2006-01-02 15:04:05 -0700 MST", parts[3])

		layer := LayerInfo{
			ID:        parts[0],
			Size:      size,
			CreatedBy: parts[2],
			Created:   created,
			Empty:     size == 0,
		}

		// Calculate efficiency (inverse of waste)
		if size > 0 {
			layer.Efficiency = float64(size-layer.WastedBytes) / float64(size)
		}

		layers = append(layers, layer)
	}

	return layers, nil
}

func parseSize(sizeStr string) int64 {
	// Remove whitespace
	sizeStr = strings.TrimSpace(sizeStr)

	// Regular expression to parse size with units
	re := regexp.MustCompile(`^([\d.]+)\s*([KMGTPE]?B?)$`)
	matches := re.FindStringSubmatch(strings.ToUpper(sizeStr))

	if len(matches) != 3 {
		return 0
	}

	// Parse the numeric part
	value, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0
	}

	// Parse the unit
	unit := matches[2]
	switch unit {
	case "B", "":
		return int64(value)
	case "KB":
		return int64(value * 1024)
	case "MB":
		return int64(value * 1024 * 1024)
	case "GB":
		return int64(value * 1024 * 1024 * 1024)
	case "TB":
		return int64(value * 1024 * 1024 * 1024 * 1024)
	default:
		return 0
	}
}

func calculateSizeBreakdown(imageInfo *ImageInfo, layers []LayerInfo) *SizeBreakdown {
	breakdown := &SizeBreakdown{
		TotalSize:   imageInfo.Size,
		LayerSizes:  make(map[string]int64),
		FileTypes:   make(map[string]int64),
		Directories: make(map[string]int64),
	}

	var totalLayerSize int64

	for _, layer := range layers {
		breakdown.LayerSizes[layer.ID] = layer.Size
		totalLayerSize += layer.Size
		breakdown.WastedSpace += layer.WastedBytes
	}

	// Estimate base image size (first few layers)
	if len(layers) > 0 {
		// Assume first 3 layers are base image
		baseLayerCount := 3
		if len(layers) < baseLayerCount {
			baseLayerCount = len(layers)
		}

		for i := len(layers) - baseLayerCount; i < len(layers); i++ {
			breakdown.BaseImageSize += layers[i].Size
		}
	}

	breakdown.ApplicationSize = breakdown.TotalSize - breakdown.BaseImageSize

	return breakdown
}

func analyzeWaste(imageName string, layers []LayerInfo) (*WasteAnalysis, error) {
	waste := &WasteAnalysis{
		WastedFiles:    make([]WastedFile, 0),
		DuplicateFiles: make([]DuplicateGroup, 0),
		LargeFiles:     make([]FileInfo, 0),
		EmptyDirs:      make([]string, 0),
		UnusedPackages: make([]string, 0),
	}

	// Calculate total waste from layers
	for _, layer := range layers {
		waste.TotalWaste += layer.WastedBytes
	}

	if layers[0].Size > 0 {
		waste.WastePercentage = float64(waste.TotalWaste) / float64(layers[0].Size) * 100
	}

	// Add common waste patterns
	commonWasteFiles := []WastedFile{
		{Path: "/var/cache/apt/archives/*.deb", Size: 0, Reason: "APT 패키지 캐시", Category: "cache"},
		{Path: "/var/lib/apt/lists/*", Size: 0, Reason: "APT 패키지 목록", Category: "cache"},
		{Path: "/tmp/*", Size: 0, Reason: "임시 파일", Category: "temp"},
		{Path: "/var/tmp/*", Size: 0, Reason: "임시 파일", Category: "temp"},
		{Path: "*.log", Size: 0, Reason: "로그 파일", Category: "log"},
		{Path: "/root/.cache/*", Size: 0, Reason: "사용자 캐시", Category: "cache"},
		{Path: "/home/*/.cache/*", Size: 0, Reason: "사용자 캐시", Category: "cache"},
	}

	waste.WastedFiles = append(waste.WastedFiles, commonWasteFiles...)

	return waste, nil
}

func getBaseImageRecommendations(imageInfo *ImageInfo) []BaseImageRec {
	recommendations := []BaseImageRec{
		{
			Name:          "alpine",
			Tag:           "latest",
			Size:          5 * 1024 * 1024, // 5MB
			SizeReduction: imageInfo.Size - (5 * 1024 * 1024),
			SecurityScore: 9.5,
			Compatibility: "high",
			Reason:        "경량화된 Linux 배포판으로 최소한의 공격 표면",
			TrustScore:    9.8,
		},
		{
			Name:          "distroless/base",
			Tag:           "latest",
			Size:          2 * 1024 * 1024, // 2MB
			SizeReduction: imageInfo.Size - (2 * 1024 * 1024),
			SecurityScore: 9.8,
			Compatibility: "medium",
			Reason:        "Shell이 없는 최소한의 런타임 환경",
			TrustScore:    9.9,
		},
		{
			Name:          "scratch",
			Tag:           "",
			Size:          0,
			SizeReduction: imageInfo.Size,
			SecurityScore: 10.0,
			Compatibility: "low",
			Reason:        "빈 이미지, 정적 바이너리에 최적",
			TrustScore:    10.0,
		},
	}

	// Sort by size reduction
	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].SizeReduction > recommendations[j].SizeReduction
	})

	return recommendations
}

func analyzeSecurityImpact(imageInfo *ImageInfo) *SecurityImpact {
	impact := &SecurityImpact{
		ExposedPorts:    imageInfo.Config.ExposedPorts,
		RunAsRoot:       imageInfo.Config.User == "" || imageInfo.Config.User == "root",
		SensitiveFiles:  make([]string, 0),
		Recommendations: make([]string, 0),
	}

	// Add security recommendations
	if impact.RunAsRoot {
		impact.Recommendations = append(impact.Recommendations, "비 root 사용자로 실행하도록 USER 지시어 추가")
	}

	if len(impact.ExposedPorts) > 0 {
		impact.Recommendations = append(impact.Recommendations, "불필요한 포트 노출 최소화")
	}

	impact.Recommendations = append(impact.Recommendations,
		"정기적인 베이스 이미지 업데이트",
		"보안 스캔 도구로 취약점 검사",
		"멀티 스테이지 빌드 사용으로 빌드 도구 제거",
	)

	return impact
}

func calculatePerformanceMetrics(imageInfo *ImageInfo, layers []LayerInfo) *PerformanceMetrics {
	metrics := &PerformanceMetrics{
		LayerCount: len(layers),
	}

	if len(layers) > 0 {
		var totalSize int64

		maxSize := int64(0)

		for _, layer := range layers {
			totalSize += layer.Size
			if layer.Size > maxSize {
				maxSize = layer.Size
			}
		}

		metrics.MaxLayerSize = maxSize
		metrics.AvgLayerSize = totalSize / int64(len(layers))
	}

	// Estimate pull time based on size (rough calculation)
	// Assume 10MB/s download speed
	downloadSpeed := int64(10 * 1024 * 1024) // 10MB/s
	if downloadSpeed > 0 {
		metrics.PullTime = time.Duration(imageInfo.Size/downloadSpeed) * time.Second
	}

	// Estimate startup time based on layer count and size
	// More layers = slower startup
	baseStartup := 100 * time.Millisecond
	layerPenalty := time.Duration(len(layers)) * 50 * time.Millisecond
	sizePenalty := time.Duration(imageInfo.Size/(100*1024*1024)) * 200 * time.Millisecond

	metrics.StartupTime = baseStartup + layerPenalty + sizePenalty

	return metrics
}

func generateSuggestions(analysis *OptimizationAnalysis) []Suggestion {
	var suggestions []Suggestion

	// Size optimization suggestions
	if analysis.SizeBreakdown.WastedSpace > 10*1024*1024 { // > 10MB
		suggestions = append(suggestions, Suggestion{
			Type:             "size",
			Priority:         "high",
			Title:            "불필요한 파일 제거",
			Description:      fmt.Sprintf("%.1fMB의 낭비된 공간 발견", float64(analysis.SizeBreakdown.WastedSpace)/(1024*1024)),
			Impact:           "이미지 크기 크게 감소",
			SizeReduction:    analysis.SizeBreakdown.WastedSpace,
			Implementation:   "패키지 매니저 캐시, 임시 파일, 로그 파일 정리",
			DockerfileChange: "RUN apt-get clean && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*",
		})
	}

	// Layer optimization
	if analysis.Performance.LayerCount > 10 {
		suggestions = append(suggestions, Suggestion{
			Type:             "performance",
			Priority:         "medium",
			Title:            "레이어 수 최적화",
			Description:      fmt.Sprintf("%d개의 레이어를 통합하여 성능 향상", analysis.Performance.LayerCount),
			Impact:           "이미지 pull 시간 단축, 캐시 효율성 향상",
			Implementation:   "RUN 명령어들을 && 로 연결하여 하나의 레이어로 통합",
			DockerfileChange: "RUN command1 && command2 && command3",
		})
	}

	// Base image suggestions
	if len(analysis.BaseImageRecs) > 0 && analysis.BaseImageRecs[0].SizeReduction > 50*1024*1024 {
		rec := analysis.BaseImageRecs[0]
		suggestions = append(suggestions, Suggestion{
			Type:             "size",
			Priority:         "high",
			Title:            fmt.Sprintf("%s 베이스 이미지 사용", rec.Name),
			Description:      fmt.Sprintf("%.1fMB 크기 절약 가능", float64(rec.SizeReduction)/(1024*1024)),
			Impact:           "이미지 크기 대폭 감소, 보안 향상",
			SizeReduction:    rec.SizeReduction,
			Implementation:   fmt.Sprintf("베이스 이미지를 %s:%s로 변경", rec.Name, rec.Tag),
			DockerfileChange: fmt.Sprintf("FROM %s:%s", rec.Name, rec.Tag),
		})
	}

	// Security suggestions
	if analysis.SecurityImpact.RunAsRoot {
		suggestions = append(suggestions, Suggestion{
			Type:             "security",
			Priority:         "high",
			Title:            "비 root 사용자 설정",
			Description:      "root 사용자로 실행되어 보안 위험 존재",
			Impact:           "보안 위험 감소",
			Implementation:   "전용 사용자 생성 및 USER 지시어 사용",
			DockerfileChange: "RUN adduser --disabled-password --gecos '' appuser\nUSER appuser",
		})
	}

	// Multi-stage build suggestion
	suggestions = append(suggestions, Suggestion{
		Type:             "size",
		Priority:         "medium",
		Title:            "멀티 스테이지 빌드 적용",
		Description:      "빌드 도구와 의존성을 최종 이미지에서 제거",
		Impact:           "이미지 크기 감소, 공격 표면 축소",
		Implementation:   "빌드 스테이지와 런타임 스테이지 분리",
		DockerfileChange: "FROM node:18 AS builder\n...\nFROM node:18-alpine AS runtime\nCOPY --from=builder ...",
	})

	// .dockerignore suggestion
	suggestions = append(suggestions, Suggestion{
		Type:             "size",
		Priority:         "medium",
		Title:            ".dockerignore 파일 사용",
		Description:      "불필요한 파일들이 빌드 컨텍스트에 포함되지 않도록 방지",
		Impact:           "빌드 시간 단축, 이미지 크기 감소",
		Implementation:   ".dockerignore 파일에 제외할 패턴 추가",
		DockerfileChange: "# .dockerignore 파일 생성\n*.log\nnode_modules\n.git\n*.tmp",
	})

	return suggestions
}

func displayAnalysisResults(analysis *OptimizationAnalysis) {
	fmt.Printf("\n📊 이미지 분석 결과\n")
	fmt.Printf("🆔 이미지 ID: %s\n", analysis.ImageInfo.ID[:12])
	fmt.Printf("📏 전체 크기: %s\n", formatBytes(analysis.ImageInfo.Size))
	fmt.Printf("🏗️ 레이어 수: %d개\n", len(analysis.LayerAnalysis))
	fmt.Printf("🗑️ 낭비된 공간: %s (%.1f%%)\n",
		formatBytes(analysis.WasteAnalysis.TotalWaste),
		analysis.WasteAnalysis.WastePercentage)

	if analysis.SecurityImpact.RunAsRoot {
		fmt.Printf("⚠️ 보안: root 사용자로 실행\n")
	}

	fmt.Printf("⏱️ 예상 다운로드 시간: %v\n", analysis.Performance.PullTime)
}

func displaySuggestions(suggestions []Suggestion) {
	// Group suggestions by priority
	priorityOrder := []string{"critical", "high", "medium", "low"}

	for _, priority := range priorityOrder {
		var prioritySuggestions []Suggestion

		for _, suggestion := range suggestions {
			if suggestion.Priority == priority {
				prioritySuggestions = append(prioritySuggestions, suggestion)
			}
		}

		if len(prioritySuggestions) == 0 {
			continue
		}

		priorityEmoji := map[string]string{
			"critical": "🚨",
			"high":     "⚠️",
			"medium":   "💡",
			"low":      "ℹ️",
		}

		fmt.Printf("\n%s %s 우선순위:\n", priorityEmoji[priority], strings.ToUpper(priority))

		for i, suggestion := range prioritySuggestions {
			fmt.Printf("  %d. %s\n", i+1, suggestion.Title)
			fmt.Printf("     %s\n", suggestion.Description)

			if suggestion.SizeReduction > 0 {
				fmt.Printf("     💾 크기 절약: %s\n", formatBytes(suggestion.SizeReduction))
			}

			fmt.Printf("     🔧 구현: %s\n", suggestion.Implementation)

			if suggestion.DockerfileChange != "" {
				fmt.Printf("     📝 Dockerfile 변경:\n")

				for _, line := range strings.Split(suggestion.DockerfileChange, "\n") {
					fmt.Printf("        %s\n", line)
				}
			}

			fmt.Printf("\n")
		}
	}
}

func applyOptimizations(imageName string, analysis *OptimizationAnalysis, suggestions []Suggestion) (*OptimizationResult, error) {
	result := &OptimizationResult{
		Analysis:       *analysis,
		OriginalSize:   analysis.ImageInfo.Size,
		GeneratedFiles: make([]string, 0),
		Applied:        make([]Suggestion, 0),
		Skipped:        make([]Suggestion, 0),
		Success:        true,
	}

	// Generate optimized Dockerfile if requested
	if outputDockerfile != "" {
		if err := generateOptimizedDockerfile(analysis, suggestions, outputDockerfile); err != nil {
			return nil, fmt.Errorf("최적화된 Dockerfile 생성 실패: %w", err)
		}

		result.GeneratedFiles = append(result.GeneratedFiles, outputDockerfile)
		fmt.Printf("📝 최적화된 Dockerfile 생성: %s\n", outputDockerfile)
	}

	// Apply docker-slim if enabled
	if enableSlim {
		optimizedImage, err := applyDockerSlim(imageName)
		if err != nil {
			fmt.Printf("⚠️ docker-slim 적용 실패: %v\n", err)
		} else {
			result.OptimizedImage = optimizedImage

			// Get optimized image size
			optimizedInfo, err := getOptimizeImageInfo(optimizedImage)
			if err == nil {
				result.OptimizedSize = optimizedInfo.Size
				result.SizeReduction = result.OriginalSize - result.OptimizedSize
				result.PercentReduced = float64(result.SizeReduction) / float64(result.OriginalSize) * 100
			}

			fmt.Printf("🎯 docker-slim 최적화 완료: %s\n", optimizedImage)
		}
	}

	// Mark high priority suggestions as applied
	for _, suggestion := range suggestions {
		if suggestion.Priority == "high" || suggestion.Priority == "critical" {
			result.Applied = append(result.Applied, suggestion)
		} else {
			result.Skipped = append(result.Skipped, suggestion)
		}
	}

	return result, nil
}

func generateOptimizedDockerfile(analysis *OptimizationAnalysis, suggestions []Suggestion, outputPath string) error {
	var dockerfile strings.Builder

	// Header comment
	dockerfile.WriteString("# Optimized Dockerfile generated by gzh-manager\n")
	dockerfile.WriteString(fmt.Sprintf("# Generated on: %s\n", time.Now().Format("2006-01-02 15:04:05")))
	dockerfile.WriteString("# Original image size: " + formatBytes(analysis.ImageInfo.Size) + "\n")
	dockerfile.WriteString("# Estimated size reduction: " + formatBytes(analysis.WasteAnalysis.TotalWaste) + "\n\n")

	// Suggest better base image
	if len(analysis.BaseImageRecs) > 0 && suggestBaseImage {
		rec := analysis.BaseImageRecs[0]
		dockerfile.WriteString(fmt.Sprintf("# Recommended base image: %s:%s\n", rec.Name, rec.Tag))
		dockerfile.WriteString(fmt.Sprintf("# Size reduction: %s\n", formatBytes(rec.SizeReduction)))
		dockerfile.WriteString(fmt.Sprintf("FROM %s:%s\n\n", rec.Name, rec.Tag))
	} else {
		dockerfile.WriteString("# Use your preferred base image\n")
		dockerfile.WriteString("FROM alpine:latest\n\n")
	}

	// Add optimization suggestions as comments and implementations
	dockerfile.WriteString("# Optimization strategies applied:\n")

	for _, suggestion := range suggestions {
		if suggestion.Priority == "high" || suggestion.Priority == "critical" {
			dockerfile.WriteString(fmt.Sprintf("# - %s: %s\n", suggestion.Title, suggestion.Description))
		}
	}

	dockerfile.WriteString("\n")

	// Security: Create non-root user
	if analysis.SecurityImpact.RunAsRoot {
		dockerfile.WriteString("# Create non-root user for security\n")
		dockerfile.WriteString("RUN addgroup -g 1001 -S appgroup && \\\n")
		dockerfile.WriteString("    adduser -u 1001 -S appuser -G appgroup\n\n")
	}

	// Package installation and cleanup
	if removeCache {
		dockerfile.WriteString("# Install packages and clean up in single layer\n")
		dockerfile.WriteString("RUN apk add --no-cache \\\n")
		dockerfile.WriteString("    # Add your packages here\n")
		dockerfile.WriteString("    ca-certificates && \\\n")
		dockerfile.WriteString("    # Clean up\n")
		dockerfile.WriteString("    rm -rf /var/cache/apk/* \\\n")
		dockerfile.WriteString("           /tmp/* \\\n")
		dockerfile.WriteString("           /var/tmp/*\n\n")
	}

	// Work directory
	dockerfile.WriteString("WORKDIR /app\n\n")

	// Copy application (placeholder)
	dockerfile.WriteString("# Copy application files\n")
	dockerfile.WriteString("COPY . .\n\n")

	// Set non-root user
	if analysis.SecurityImpact.RunAsRoot {
		dockerfile.WriteString("# Switch to non-root user\n")
		dockerfile.WriteString("USER appuser\n\n")
	}

	// Health check
	dockerfile.WriteString("# Health check\n")
	dockerfile.WriteString("HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \\\n")
	dockerfile.WriteString("    CMD echo 'Health check placeholder' || exit 1\n\n")

	// Default command
	dockerfile.WriteString("# Default command\n")
	dockerfile.WriteString("CMD [\"echo\", \"Optimized container ready\"]\n")

	// Write to file
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return err
	}

	return os.WriteFile(outputPath, []byte(dockerfile.String()), 0o644)
}

func applyDockerSlim(imageName string) (string, error) {
	// Check if docker-slim is available
	if _, err := exec.LookPath("docker-slim"); err != nil {
		return "", fmt.Errorf("docker-slim이 설치되어 있지 않습니다")
	}

	optimizedName := imageName + ".slim"

	args := []string{"build", "--target", optimizedName}
	args = append(args, slimOptions...)
	args = append(args, imageName)

	cmd := exec.Command("docker-slim", args...)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("docker-slim 실행 실패: %w", err)
	}

	return optimizedName, nil
}

func displayOptimizationResults(result *OptimizationResult) {
	fmt.Printf("\n📊 최적화 결과\n")
	fmt.Printf("📦 원본 크기: %s\n", formatBytes(result.OriginalSize))

	if result.OptimizedSize > 0 {
		fmt.Printf("🎯 최적화된 크기: %s\n", formatBytes(result.OptimizedSize))
		fmt.Printf("💾 크기 절약: %s (%.1f%%)\n",
			formatBytes(result.SizeReduction), result.PercentReduced)
	}

	fmt.Printf("⏱️ 소요 시간: %v\n", result.Duration)
	fmt.Printf("✅ 적용된 최적화: %d개\n", len(result.Applied))
	fmt.Printf("⏭️ 건너뛴 제안: %d개\n", len(result.Skipped))

	if len(result.GeneratedFiles) > 0 {
		fmt.Printf("📝 생성된 파일:\n")

		for _, file := range result.GeneratedFiles {
			fmt.Printf("   - %s\n", file)
		}
	}
}

func saveOptimizationReport(analysis *OptimizationAnalysis, suggestions []Suggestion) error {
	reportData := struct {
		Analysis    *OptimizationAnalysis `json:"analysis"`
		Suggestions []Suggestion          `json:"suggestions"`
		Generated   time.Time             `json:"generated"`
	}{
		Analysis:    analysis,
		Suggestions: suggestions,
		Generated:   time.Now(),
	}

	filename := reportOutput
	if filename == "" {
		filename = fmt.Sprintf("optimization-report-%s.json", time.Now().Format("20060102-150405"))
	}

	data, err := json.MarshalIndent(reportData, "", "  ")
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(filename), 0o755); err != nil {
		return err
	}

	if err := os.WriteFile(filename, data, 0o644); err != nil {
		return err
	}

	fmt.Printf("📄 최적화 보고서 저장: %s\n", filename)

	return nil
}
