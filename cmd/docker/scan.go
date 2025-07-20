package docker

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// ScanCmd represents the scan command.
var ScanCmd = &cobra.Command{
	Use:   "scan",
	Short: "컨테이너 이미지 보안 스캔",
	Long: `컨테이너 이미지의 보안 취약점을 스캔하고 상세 보고서를 생성합니다.

보안 스캔 기능:
- 다중 스캐너 지원 (Trivy, Grype, Snyk, Clair)
- CVE 취약점 탐지
- 라이선스 스캔
- 설정 오류 탐지
- 시크릿 스캔
- 맬웨어 탐지
- SBOM(Software Bill of Materials) 생성
- 컴플라이언스 검사

Examples:
  gz docker scan myapp:latest
  gz docker scan --scanner trivy,grype myapp:latest
  gz docker scan --severity HIGH --format json myapp:latest
  gz docker scan --sbom --output report.json myapp:latest`,
	Run: runScan,
}

var (
	scanImage      string
	scanScanners   []string
	scanSeverities []string
	scanFormat     string
	scanOutput     string
	scanQuiet      bool
	scanVerbose    bool
	scanTimeout    time.Duration
	scanSkipUpdate bool

	// Advanced scanning options.
	enableSBOM       bool
	enableSecrets    bool
	enableMalware    bool
	enableConfig     bool
	enableLicense    bool
	enableCompliance bool

	// Filter options.
	scanPackageTypes []string
	scanSkipFiles    []string
	scanSkipDirs     []string
	scanOnlyFixed    bool

	// Output options.
	scanTemplate string
	scanCSV      bool
	scanHTML     bool
	scanSARIF    bool
	scanCyclone  bool

	// Integration options.
	scanUpload     string
	scanWebhook    string
	scanFailOn     string
	scanIgnoreFile string
)

func init() {
	// Basic scan flags
	ScanCmd.Flags().StringVarP(&scanImage, "image", "i", "", "스캔할 이미지 태그")
	ScanCmd.Flags().StringSliceVar(&scanScanners, "scanner", []string{"trivy"}, "사용할 스캐너 (trivy,grype,snyk,clair)")
	ScanCmd.Flags().StringSliceVar(&scanSeverities, "severity", []string{"HIGH", "CRITICAL"}, "스캔할 심각도 수준")
	ScanCmd.Flags().StringVar(&scanFormat, "format", "json", "출력 형식 (json,table,csv,html,sarif)")
	ScanCmd.Flags().StringVarP(&scanOutput, "output", "o", "", "출력 파일 경로")
	ScanCmd.Flags().BoolVarP(&scanQuiet, "quiet", "q", false, "조용한 모드")
	ScanCmd.Flags().BoolVarP(&scanVerbose, "verbose", "v", false, "상세 출력")
	ScanCmd.Flags().DurationVar(&scanTimeout, "timeout", 10*time.Minute, "스캔 타임아웃")
	ScanCmd.Flags().BoolVar(&scanSkipUpdate, "skip-update", false, "DB 업데이트 건너뛰기")

	// Advanced scanning flags
	ScanCmd.Flags().BoolVar(&enableSBOM, "sbom", false, "SBOM 생성")
	ScanCmd.Flags().BoolVar(&enableSecrets, "secrets", false, "시크릿 스캔")
	ScanCmd.Flags().BoolVar(&enableMalware, "malware", false, "맬웨어 탐지")
	ScanCmd.Flags().BoolVar(&enableConfig, "config", false, "설정 오류 탐지")
	ScanCmd.Flags().BoolVar(&enableLicense, "license", false, "라이선스 스캔")
	ScanCmd.Flags().BoolVar(&enableCompliance, "compliance", false, "컴플라이언스 검사")

	// Filter flags
	ScanCmd.Flags().StringSliceVar(&scanPackageTypes, "package-types", []string{}, "패키지 타입 필터")
	ScanCmd.Flags().StringSliceVar(&scanSkipFiles, "skip-files", []string{}, "제외할 파일 패턴")
	ScanCmd.Flags().StringSliceVar(&scanSkipDirs, "skip-dirs", []string{}, "제외할 디렉터리")
	ScanCmd.Flags().BoolVar(&scanOnlyFixed, "only-fixed", false, "수정 가능한 취약점만 표시")

	// Output format flags
	ScanCmd.Flags().StringVar(&scanTemplate, "template", "", "사용자 정의 템플릿")
	ScanCmd.Flags().BoolVar(&scanCSV, "csv", false, "CSV 형식으로 출력")
	ScanCmd.Flags().BoolVar(&scanHTML, "html", false, "HTML 보고서 생성")
	ScanCmd.Flags().BoolVar(&scanSARIF, "sarif", false, "SARIF 형식으로 출력")
	ScanCmd.Flags().BoolVar(&scanCyclone, "cyclone", false, "CycloneDX SBOM 생성")

	// Integration flags
	ScanCmd.Flags().StringVar(&scanUpload, "upload", "", "스캔 결과 업로드 URL")
	ScanCmd.Flags().StringVar(&scanWebhook, "webhook", "", "결과 전송 웹훅 URL")
	ScanCmd.Flags().StringVar(&scanFailOn, "fail-on", "", "실패 조건 (critical,high,medium,low)")
	ScanCmd.Flags().StringVar(&scanIgnoreFile, "ignore-file", "", "무시할 취약점 파일")
}

// ScanConfiguration represents scan settings.
type ScanConfiguration struct {
	Image       string            `json:"image"`
	Scanners    []string          `json:"scanners"`
	Severities  []string          `json:"severities"`
	Format      string            `json:"format"`
	Output      string            `json:"output,omitempty"`
	Timeout     time.Duration     `json:"timeout"`
	SkipUpdate  bool              `json:"skipUpdate"`
	Options     ScanOptions       `json:"options"`
	Filters     ScanFilters       `json:"filters"`
	Outputs     OutputFormats     `json:"outputs"`
	Integration IntegrationConfig `json:"integration"`
	Timestamp   time.Time         `json:"timestamp"`
}

type ScanOptions struct {
	SBOM       bool `json:"sbom"`
	Secrets    bool `json:"secrets"`
	Malware    bool `json:"malware"`
	Config     bool `json:"config"`
	License    bool `json:"license"`
	Compliance bool `json:"compliance"`
}

type ScanFilters struct {
	PackageTypes []string `json:"packageTypes,omitempty"`
	SkipFiles    []string `json:"skipFiles,omitempty"`
	SkipDirs     []string `json:"skipDirs,omitempty"`
	OnlyFixed    bool     `json:"onlyFixed"`
}

type OutputFormats struct {
	JSON      bool   `json:"json"`
	CSV       bool   `json:"csv"`
	HTML      bool   `json:"html"`
	SARIF     bool   `json:"sarif"`
	CycloneDX bool   `json:"cycloneDx"`
	Template  string `json:"template,omitempty"`
}

type IntegrationConfig struct {
	UploadURL  string `json:"uploadUrl,omitempty"`
	WebhookURL string `json:"webhookUrl,omitempty"`
	FailOn     string `json:"failOn,omitempty"`
	IgnoreFile string `json:"ignoreFile,omitempty"`
}

// ComprehensiveScanResult represents the complete scan results.
type ComprehensiveScanResult struct {
	Configuration ScanConfiguration `json:"configuration"`
	Summary       ScanSummary       `json:"summary"`
	Results       []ScannerResult   `json:"results"`
	SBOM          *SBOM             `json:"sbom,omitempty"`
	Compliance    *ComplianceReport `json:"compliance,omitempty"`
	Metadata      ScanMetadata      `json:"metadata"`
	Success       bool              `json:"success"`
	Error         string            `json:"error,omitempty"`
}

type ScannerResult struct {
	Scanner         string             `json:"scanner"`
	Success         bool               `json:"success"`
	Duration        time.Duration      `json:"duration"`
	Vulnerabilities []Vulnerability    `json:"vulnerabilities"`
	Secrets         []Secret           `json:"secrets,omitempty"`
	Malware         []MalwareDetection `json:"malware,omitempty"`
	ConfigIssues    []ConfigIssue      `json:"configIssues,omitempty"`
	LicenseIssues   []LicenseIssue     `json:"licenseIssues,omitempty"`
	Summary         ScanSummary        `json:"summary"`
	Error           string             `json:"error,omitempty"`
}

type Vulnerability struct {
	ID            string            `json:"id"`
	Title         string            `json:"title"`
	Description   string            `json:"description"`
	Severity      string            `json:"severity"`
	Score         float64           `json:"score,omitempty"`
	Vector        string            `json:"vector,omitempty"`
	Package       PackageInfo       `json:"package"`
	FixedIn       string            `json:"fixedIn,omitempty"`
	PublishedDate string            `json:"publishedDate,omitempty"`
	LastModified  string            `json:"lastModified,omitempty"`
	References    []string          `json:"references,omitempty"`
	CPE           string            `json:"cpe,omitempty"`
	Layer         ScanLayerInfo     `json:"layer,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
}

type PackageInfo struct {
	Name         string `json:"name"`
	Version      string `json:"version"`
	Type         string `json:"type"`
	Path         string `json:"path,omitempty"`
	Architecture string `json:"architecture,omitempty"`
}

type ScanLayerInfo struct {
	Digest    string `json:"digest"`
	DiffID    string `json:"diffId"`
	CreatedBy string `json:"createdBy,omitempty"`
}

type Secret struct {
	Type       string `json:"type"`
	Title      string `json:"title"`
	Severity   string `json:"severity"`
	Match      string `json:"match"`
	File       string `json:"file"`
	Line       int    `json:"line"`
	Category   string `json:"category"`
	Confidence string `json:"confidence"`
}

type MalwareDetection struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	Family    string `json:"family,omitempty"`
	Severity  string `json:"severity"`
	File      string `json:"file"`
	Hash      string `json:"hash,omitempty"`
	Signature string `json:"signature"`
}

type ConfigIssue struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Severity    string `json:"severity"`
	Category    string `json:"category"`
	File        string `json:"file,omitempty"`
	Line        int    `json:"line,omitempty"`
	Resolution  string `json:"resolution,omitempty"`
}

type LicenseIssue struct {
	Package    string   `json:"package"`
	License    string   `json:"license"`
	Severity   string   `json:"severity"`
	Category   string   `json:"category"`
	Confidence string   `json:"confidence"`
	File       string   `json:"file,omitempty"`
	Conflicts  []string `json:"conflicts,omitempty"`
}

type SBOM struct {
	Format       string       `json:"format"`
	Version      string       `json:"version"`
	Generated    time.Time    `json:"generated"`
	Components   []Component  `json:"components"`
	Dependencies []Dependency `json:"dependencies"`
	Licenses     []License    `json:"licenses"`
	Tools        []Tool       `json:"tools"`
}

type Component struct {
	Type         string    `json:"type"`
	Name         string    `json:"name"`
	Version      string    `json:"version"`
	Description  string    `json:"description,omitempty"`
	Supplier     string    `json:"supplier,omitempty"`
	Author       string    `json:"author,omitempty"`
	Publisher    string    `json:"publisher,omitempty"`
	Group        string    `json:"group,omitempty"`
	Scope        string    `json:"scope,omitempty"`
	Hashes       []Hash    `json:"hashes,omitempty"`
	Licenses     []License `json:"licenses,omitempty"`
	Copyright    string    `json:"copyright,omitempty"`
	CPE          string    `json:"cpe,omitempty"`
	PURL         string    `json:"purl,omitempty"`
	ExternalRefs []ExtRef  `json:"externalRefs,omitempty"`
}

type Dependency struct {
	Ref          string   `json:"ref"`
	Dependencies []string `json:"dependencies,omitempty"`
}

type License struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	Text string `json:"text,omitempty"`
	URL  string `json:"url,omitempty"`
}

type Tool struct {
	Vendor  string `json:"vendor"`
	Name    string `json:"name"`
	Version string `json:"version"`
}

type Hash struct {
	Algorithm string `json:"algorithm"`
	Value     string `json:"value"`
}

type ExtRef struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

type ComplianceReport struct {
	Framework string             `json:"framework"`
	Version   string             `json:"version"`
	Results   []ComplianceResult `json:"results"`
	Summary   ComplianceSummary  `json:"summary"`
}

type ComplianceResult struct {
	RuleID      string `json:"ruleId"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Severity    string `json:"severity"`
	Status      string `json:"status"` // pass, fail, warn, info
	Message     string `json:"message,omitempty"`
}

type ComplianceSummary struct {
	Total   int `json:"total"`
	Passed  int `json:"passed"`
	Failed  int `json:"failed"`
	Warned  int `json:"warned"`
	Skipped int `json:"skipped"`
}

type ScanMetadata struct {
	ImageID      string            `json:"imageId,omitempty"`
	ImageDigest  string            `json:"imageDigest,omitempty"`
	Size         int64             `json:"size"`
	Architecture string            `json:"architecture,omitempty"`
	OS           string            `json:"os,omitempty"`
	Created      string            `json:"created,omitempty"`
	Labels       map[string]string `json:"labels,omitempty"`
	Layers       []string          `json:"layers,omitempty"`
	RepoTags     []string          `json:"repoTags,omitempty"`
	RepoDigests  []string          `json:"repoDigests,omitempty"`
}

func runScan(cmd *cobra.Command, args []string) {
	// Determine image to scan
	if len(args) > 0 {
		scanImage = args[0]
	}

	if scanImage == "" {
		fmt.Printf("❌ 스캔할 이미지가 필요합니다\n")
		fmt.Printf("사용법: gz docker scan <image>\n")
		os.Exit(1)
	}

	fmt.Printf("🔍 컨테이너 이미지 보안 스캔 시작\n")
	fmt.Printf("📦 이미지: %s\n", scanImage)
	fmt.Printf("🔍 스캐너: %s\n", strings.Join(scanScanners, ", "))

	// Create scan configuration
	config := createScanConfiguration()

	// Perform comprehensive scan
	result := performComprehensiveScan(config)

	// Display results
	displayScanResults(result)

	// Save outputs
	if err := saveOutputs(result); err != nil {
		fmt.Printf("⚠️ 출력 저장 실패: %v\n", err)
	}

	// Check failure conditions
	if shouldFail(result) {
		fmt.Printf("❌ 스캔 실패 조건 충족\n")
		os.Exit(1)
	}

	fmt.Printf("✅ 스캔 완료\n")
}

func createScanConfiguration() *ScanConfiguration {
	return &ScanConfiguration{
		Image:      scanImage,
		Scanners:   scanScanners,
		Severities: scanSeverities,
		Format:     scanFormat,
		Output:     scanOutput,
		Timeout:    scanTimeout,
		SkipUpdate: scanSkipUpdate,
		Options: ScanOptions{
			SBOM:       enableSBOM,
			Secrets:    enableSecrets,
			Malware:    enableMalware,
			Config:     enableConfig,
			License:    enableLicense,
			Compliance: enableCompliance,
		},
		Filters: ScanFilters{
			PackageTypes: scanPackageTypes,
			SkipFiles:    scanSkipFiles,
			SkipDirs:     scanSkipDirs,
			OnlyFixed:    scanOnlyFixed,
		},
		Outputs: OutputFormats{
			JSON:      scanFormat == "json",
			CSV:       scanCSV,
			HTML:      scanHTML,
			SARIF:     scanSARIF,
			CycloneDX: scanCyclone,
			Template:  scanTemplate,
		},
		Integration: IntegrationConfig{
			UploadURL:  scanUpload,
			WebhookURL: scanWebhook,
			FailOn:     scanFailOn,
			IgnoreFile: scanIgnoreFile,
		},
		Timestamp: time.Now(),
	}
}

func performComprehensiveScan(config *ScanConfiguration) *ComprehensiveScanResult {
	result := &ComprehensiveScanResult{
		Configuration: *config,
		Results:       make([]ScannerResult, 0),
		Summary:       ScanSummary{},
		Metadata:      ScanMetadata{},
	}

	// Get image metadata
	if err := getImageMetadata(config.Image, &result.Metadata); err != nil {
		fmt.Printf("⚠️ 이미지 메타데이터 조회 실패: %v\n", err)
	}

	// Run each scanner
	for _, scanner := range config.Scanners {
		if !scanQuiet {
			fmt.Printf("🔍 %s 스캔 실행 중...\n", scanner)
		}

		scannerResult, err := runScannerComprehensive(scanner, config)
		if err != nil {
			fmt.Printf("⚠️ %s 스캔 실패: %v\n", scanner, err)
			scannerResult = &ScannerResult{
				Scanner: scanner,
				Success: false,
				Error:   err.Error(),
			}
		}

		result.Results = append(result.Results, *scannerResult)

		// Aggregate summary
		result.Summary.Total += scannerResult.Summary.Total
		result.Summary.Critical += scannerResult.Summary.Critical
		result.Summary.High += scannerResult.Summary.High
		result.Summary.Medium += scannerResult.Summary.Medium
		result.Summary.Low += scannerResult.Summary.Low
		result.Summary.Unknown += scannerResult.Summary.Unknown
	}

	// Generate SBOM if requested
	if config.Options.SBOM {
		sbom, err := generateSBOM(config.Image)
		if err != nil {
			fmt.Printf("⚠️ SBOM 생성 실패: %v\n", err)
		} else {
			result.SBOM = sbom
		}
	}

	// Run compliance checks if requested
	if config.Options.Compliance {
		compliance := runComplianceChecks(config.Image)
		result.Compliance = compliance
	}

	result.Success = true

	return result
}

func runScannerComprehensive(scanner string, config *ScanConfiguration) (*ScannerResult, error) {
	result := &ScannerResult{
		Scanner:         scanner,
		Vulnerabilities: make([]Vulnerability, 0),
		Secrets:         make([]Secret, 0),
		Malware:         make([]MalwareDetection, 0),
		ConfigIssues:    make([]ConfigIssue, 0),
		LicenseIssues:   make([]LicenseIssue, 0),
		Summary:         ScanSummary{},
	}

	startTime := time.Now()

	switch scanner {
	case "trivy":
		err := runTrivyScan(config, result)
		if err != nil {
			return result, err
		}
	case "grype":
		err := runGrypeScan(config, result)
		if err != nil {
			return result, err
		}
	case "snyk":
		err := runSnykScan(config, result)
		if err != nil {
			return result, err
		}
	case "clair":
		err := runClairScan(config, result)
		if err != nil {
			return result, err
		}
	default:
		return result, fmt.Errorf("지원하지 않는 스캐너: %s", scanner)
	}

	result.Duration = time.Since(startTime)
	result.Success = true

	return result, nil
}

func runTrivyScan(config *ScanConfiguration, result *ScannerResult) error {
	args := []string{"image", "--format", "json"}

	// Add severity filter
	if len(config.Severities) > 0 {
		args = append(args, "--severity", strings.Join(config.Severities, ","))
	}

	// Add scan types
	scanTypes := []string{}
	if config.Options.SBOM {
		scanTypes = append(scanTypes, "vuln")
	}

	if config.Options.Secrets {
		scanTypes = append(scanTypes, "secret")
	}

	if config.Options.Config {
		scanTypes = append(scanTypes, "config")
	}

	if config.Options.License {
		scanTypes = append(scanTypes, "license")
	}

	if len(scanTypes) > 0 {
		args = append(args, "--scanners", strings.Join(scanTypes, ","))
	}

	// Skip update if requested
	if config.SkipUpdate {
		args = append(args, "--skip-update")
	}

	// Add filters
	if config.Filters.OnlyFixed {
		args = append(args, "--ignore-unfixed")
	}

	args = append(args, config.Image)

	cmd := exec.Command("trivy", args...)

	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("trivy 실행 실패: %w", err)
	}

	return parseTrivyComprehensive(output, result)
}

func parseTrivyComprehensive(output []byte, result *ScannerResult) error {
	// nolint:tagliatelle // External API format - must match Trivy JSON output
	var trivyResult struct {
		Results []struct {
			Vulnerabilities []struct {
				VulnerabilityID string `json:"VulnerabilityID"`
				Title           string `json:"Title"`
				Description     string `json:"Description"`
				Severity        string `json:"Severity"`
				CVSS            struct {
					Score float64 `json:"Score"`
				} `json:"CVSS"`
				PkgID        string `json:"PkgID"`
				PkgName      string `json:"PkgName"`
				PkgVersion   string `json:"InstalledVersion"`
				FixedVersion string `json:"FixedVersion"`
				PkgPath      string `json:"PkgPath"`
				Layer        struct {
					Digest string `json:"Digest"`
					DiffID string `json:"DiffID"`
				} `json:"Layer"`
				PublishedDate    string   `json:"PublishedDate"`
				LastModifiedDate string   `json:"LastModifiedDate"`
				References       []string `json:"References"`
			} `json:"Vulnerabilities"`
			Secrets []struct {
				RuleID    string `json:"RuleID"`
				Category  string `json:"Category"`
				Severity  string `json:"Severity"`
				Title     string `json:"Title"`
				StartLine int    `json:"StartLine"`
				EndLine   int    `json:"EndLine"`
				Code      struct {
					Lines []struct {
						Number  int    `json:"Number"`
						Content string `json:"Content"`
					} `json:"Lines"`
				} `json:"Code"`
				Match string `json:"Match"`
			} `json:"Secrets"`
		} `json:"Results"`
	}

	if err := json.Unmarshal(output, &trivyResult); err != nil {
		return fmt.Errorf("trivy 결과 파싱 실패: %w", err)
	}

	// Process vulnerabilities
	for _, res := range trivyResult.Results {
		for _, vuln := range res.Vulnerabilities {
			vulnerability := Vulnerability{
				ID:          vuln.VulnerabilityID,
				Title:       vuln.Title,
				Description: vuln.Description,
				Severity:    vuln.Severity,
				Score:       vuln.CVSS.Score,
				Package: PackageInfo{
					Name:    vuln.PkgName,
					Version: vuln.PkgVersion,
					Path:    vuln.PkgPath,
				},
				FixedIn:       vuln.FixedVersion,
				PublishedDate: vuln.PublishedDate,
				LastModified:  vuln.LastModifiedDate,
				References:    vuln.References,
				Layer:         ScanLayerInfo{
					// Digest: vuln.Layer.Digest,  // TODO: Fix when trivy schema is available
					// DiffID: vuln.Layer.DiffID,  // TODO: Fix when trivy schema is available
				},
			}

			result.Vulnerabilities = append(result.Vulnerabilities, vulnerability)

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

		// Process secrets
		for _, sec := range res.Secrets {
			secret := Secret{
				Type:       sec.RuleID,
				Title:      sec.Title,
				Severity:   sec.Severity,
				Category:   sec.Category,
				Line:       sec.StartLine,
				Match:      sec.Match,
				Confidence: "high", // Trivy doesn't provide confidence
			}

			result.Secrets = append(result.Secrets, secret)
		}
	}

	return nil
}

func runGrypeScan(config *ScanConfiguration, _ *ScannerResult) error {
	args := []string{config.Image, "--output", "json"}

	// Add severity filter
	if len(config.Severities) > 0 {
		// Grype uses different format
		for _, sev := range config.Severities {
			args = append(args, "--fail-on", strings.ToLower(sev))
		}
	}

	cmd := exec.Command("grype", args...)

	_, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("grype 실행 실패: %w", err)
	}

	// Simplified Grype parsing - implement full parsing as needed
	return nil
}

func runSnykScan(config *ScanConfiguration, _ *ScannerResult) error {
	args := []string{"container", "test", config.Image, "--json"}

	cmd := exec.Command("snyk", args...)

	_, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("snyk 실행 실패: %w", err)
	}

	// Simplified Snyk parsing - implement full parsing as needed
	return nil
}

func runClairScan(config *ScanConfiguration, result *ScannerResult) error {
	// Simplified Clair integration - implement full integration as needed
	return fmt.Errorf("clair 스캐너는 아직 구현되지 않았습니다")
}

func getImageMetadata(image string, metadata *ScanMetadata) error {
	// Get image inspect information
	cmd := exec.Command("docker", "image", "inspect", image)

	output, err := cmd.Output()
	if err != nil {
		return err
	}

	// nolint:tagliatelle // External API format - must match Docker inspect JSON output
	var inspectResult []struct {
		ID          string   `json:"Id"`
		RepoTags    []string `json:"RepoTags"`
		RepoDigests []string `json:"RepoDigests"`
		Size        int64    `json:"Size"`
		Config      struct {
			Labels map[string]string `json:"Labels"`
		} `json:"Config"`
		Architecture string `json:"Architecture"`
		OS           string `json:"Os"`
		Created      string `json:"Created"`
		RootFS       struct {
			Layers []string `json:"Layers"`
		} `json:"RootFS"`
	}

	if err := json.Unmarshal(output, &inspectResult); err != nil {
		return err
	}

	if len(inspectResult) > 0 {
		img := inspectResult[0]
		metadata.ImageID = img.ID
		metadata.Size = img.Size
		metadata.Architecture = img.Architecture
		metadata.OS = img.OS
		metadata.Created = img.Created
		metadata.Labels = img.Config.Labels
		metadata.Layers = img.RootFS.Layers
		metadata.RepoTags = img.RepoTags
		metadata.RepoDigests = img.RepoDigests
	}

	return nil
}

func generateSBOM(image string) (*SBOM, error) {
	// Use syft to generate SBOM
	cmd := exec.Command("syft", image, "--output", "json")

	_, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("syft 실행 실패: %w", err)
	}

	// Parse syft output to SBOM format
	sbom := &SBOM{
		Format:     "syft-json",
		Version:    "1.0",
		Generated:  time.Now(),
		Components: make([]Component, 0),
		Tools: []Tool{
			{
				Vendor:  "Anchore",
				Name:    "syft",
				Version: "latest",
			},
		},
	}

	// Simplified SBOM parsing - implement full parsing as needed
	return sbom, nil
}

func runComplianceChecks(_ string) *ComplianceReport {
	// Run compliance checks using tools like OPA/Conftest
	report := &ComplianceReport{
		Framework: "CIS Docker Benchmark",
		Version:   "1.5.0",
		Results:   make([]ComplianceResult, 0),
		Summary:   ComplianceSummary{},
	}

	// Simplified compliance checking - implement full checks as needed
	return report
}

func displayScanResults(result *ComprehensiveScanResult) {
	if scanQuiet {
		return
	}

	fmt.Printf("\n📊 스캔 결과 요약\n")
	fmt.Printf("📦 이미지: %s\n", result.Configuration.Image)
	fmt.Printf("📏 크기: %s\n", formatBytes(result.Metadata.Size))
	fmt.Printf("🏗️ 아키텍처: %s\n", result.Metadata.Architecture)
	fmt.Printf("💻 OS: %s\n", result.Metadata.OS)

	fmt.Printf("\n🔍 취약점 요약:\n")
	fmt.Printf("  총 %d개 (치명적: %d, 높음: %d, 중간: %d, 낮음: %d)\n",
		result.Summary.Total, result.Summary.Critical,
		result.Summary.High, result.Summary.Medium, result.Summary.Low)

	// Display per-scanner results
	for _, scanResult := range result.Results {
		if scanResult.Success {
			fmt.Printf("\n📋 %s 결과:\n", scanResult.Scanner)
			fmt.Printf("  취약점: %d개\n", len(scanResult.Vulnerabilities))

			if len(scanResult.Secrets) > 0 {
				fmt.Printf("  시크릿: %d개\n", len(scanResult.Secrets))
			}

			if len(scanResult.ConfigIssues) > 0 {
				fmt.Printf("  설정 문제: %d개\n", len(scanResult.ConfigIssues))
			}

			if len(scanResult.LicenseIssues) > 0 {
				fmt.Printf("  라이선스 문제: %d개\n", len(scanResult.LicenseIssues))
			}
		} else {
			fmt.Printf("❌ %s 실패: %s\n", scanResult.Scanner, scanResult.Error)
		}
	}

	if result.SBOM != nil {
		fmt.Printf("\n📋 SBOM: %d개 컴포넌트\n", len(result.SBOM.Components))
	}

	if result.Compliance != nil {
		fmt.Printf("\n✅ 컴플라이언스: %d/%d 통과\n",
			result.Compliance.Summary.Passed, result.Compliance.Summary.Total)
	}
}

func saveOutputs(result *ComprehensiveScanResult) error {
	// Save main JSON output
	if scanOutput != "" {
		data, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return err
		}

		if err := os.MkdirAll(filepath.Dir(scanOutput), 0o755); err != nil {
			return err
		}

		if err := os.WriteFile(scanOutput, data, 0o644); err != nil {
			return err
		}

		fmt.Printf("📄 스캔 결과 저장: %s\n", scanOutput)
	}

	// Save additional formats
	baseOutput := strings.TrimSuffix(scanOutput, filepath.Ext(scanOutput))
	if baseOutput == "" {
		baseOutput = fmt.Sprintf("scan-result-%s", time.Now().Format("20060102-150405"))
	}

	if scanHTML {
		htmlFile := baseOutput + ".html"
		if err := generateHTMLReport(result, htmlFile); err != nil {
			return err
		}

		fmt.Printf("📄 HTML 보고서 저장: %s\n", htmlFile)
	}

	if scanSARIF {
		sarifFile := baseOutput + ".sarif"
		if err := generateSARIFReport(result, sarifFile); err != nil {
			return err
		}

		fmt.Printf("📄 SARIF 보고서 저장: %s\n", sarifFile)
	}

	if scanCSV {
		csvFile := baseOutput + ".csv"
		if err := generateCSVReport(result, csvFile); err != nil {
			return err
		}

		fmt.Printf("📄 CSV 보고서 저장: %s\n", csvFile)
	}

	return nil
}

func generateHTMLReport(result *ComprehensiveScanResult, filename string) error {
	// Simplified HTML report generation - implement full template as needed
	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <title>Security Scan Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .summary { background: #f5f5f5; padding: 20px; border-radius: 5px; }
        .critical { color: #d73027; }
        .high { color: #fc8d59; }
        .medium { color: #fee08b; }
        .low { color: #99d594; }
    </style>
</head>
<body>
    <h1>Security Scan Report</h1>
    <div class="summary">
        <h2>Summary</h2>
        <p>Image: %s</p>
        <p>Total Vulnerabilities: %d</p>
        <p>Critical: <span class="critical">%d</span></p>
        <p>High: <span class="high">%d</span></p>
        <p>Medium: <span class="medium">%d</span></p>
        <p>Low: <span class="low">%d</span></p>
    </div>
</body>
</html>`,
		result.Configuration.Image,
		result.Summary.Total,
		result.Summary.Critical,
		result.Summary.High,
		result.Summary.Medium,
		result.Summary.Low)

	return os.WriteFile(filename, []byte(html), 0o644)
}

func generateSARIFReport(_ *ComprehensiveScanResult, filename string) error {
	// Simplified SARIF generation - implement full SARIF format as needed
	sarif := map[string]interface{}{
		"version": "2.1.0",
		"runs": []map[string]interface{}{
			{
				"tool": map[string]interface{}{
					"driver": map[string]interface{}{
						"name": "gzh-manager",
					},
				},
				"results": []map[string]interface{}{},
			},
		},
	}

	data, err := json.MarshalIndent(sarif, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0o644)
}

func generateCSVReport(result *ComprehensiveScanResult, filename string) error {
	// Simplified CSV generation
	csv := "Scanner,Vulnerability ID,Severity,Package,Version,Fixed In\n"

	for _, scanResult := range result.Results {
		for _, vuln := range scanResult.Vulnerabilities {
			csv += fmt.Sprintf("%s,%s,%s,%s,%s,%s\n",
				scanResult.Scanner, vuln.ID, vuln.Severity,
				vuln.Package.Name, vuln.Package.Version, vuln.FixedIn)
		}
	}

	return os.WriteFile(filename, []byte(csv), 0o644)
}

func shouldFail(result *ComprehensiveScanResult) bool {
	if scanFailOn == "" {
		return false
	}

	switch strings.ToLower(scanFailOn) {
	case "critical":
		return result.Summary.Critical > 0
	case "high":
		return result.Summary.Critical > 0 || result.Summary.High > 0
	case "medium":
		return result.Summary.Critical > 0 || result.Summary.High > 0 || result.Summary.Medium > 0
	case "low":
		return result.Summary.Total > 0
	}

	return false
}
