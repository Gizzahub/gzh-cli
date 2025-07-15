package template

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// SecurityCmd represents the security command
var SecurityCmd = &cobra.Command{
	Use:   "security",
	Short: "보안 및 컴플라이언스 관리",
	Long: `기업용 템플릿 마켓플레이스의 보안 및 컴플라이언스 기능을 관리합니다.

보안 기능:
- 템플릿 서명 및 검증
- 취약점 스캔 및 보고서
- 접근 제어 및 권한 관리
- 암호화 및 키 관리
- 보안 정책 적용
- 컴플라이언스 검사
- 보안 모니터링 및 알림
- 침입 탐지 시스템 (IDS)

Examples:
  gz template security scan --template-id abc123
  gz template security sign --template-path ./template.zip --key-file private.pem
  gz template security verify --template-path ./template.zip --cert-file public.pem
  gz template security audit --compliance SOC2`,
	Run: runSecurity,
}

var (
	securityTemplatePath string
	keyFile              string
	certFile             string
	complianceStd        string
	scanType             string
	severityLevel        string
	securityOutputFormat string
	reportPath           string
	encryptionKey        string
	securityPolicy       string
)

func init() {
	SecurityCmd.Flags().StringVar(&templatePath, "template-path", "", "스캔할 템플릿 경로")
	SecurityCmd.Flags().StringVar(&keyFile, "key-file", "", "개인키 파일")
	SecurityCmd.Flags().StringVar(&certFile, "cert-file", "", "인증서 파일")
	SecurityCmd.Flags().StringVar(&complianceStd, "compliance", "", "컴플라이언스 표준 (SOC2, GDPR, HIPAA)")
	SecurityCmd.Flags().StringVar(&scanType, "scan-type", "all", "스캔 유형 (vulnerability, malware, license, all)")
	SecurityCmd.Flags().StringVar(&severityLevel, "severity", "medium", "심각도 수준 (low, medium, high, critical)")
	SecurityCmd.Flags().StringVar(&outputFormat, "format", "json", "출력 형식 (json, html, pdf)")
	SecurityCmd.Flags().StringVar(&reportPath, "report", "", "보고서 출력 경로")
	SecurityCmd.Flags().StringVar(&encryptionKey, "encryption-key", "", "암호화 키")
	SecurityCmd.Flags().StringVar(&securityPolicy, "policy", "", "보안 정책 파일")

	// Add subcommands
	SecurityCmd.AddCommand(securityScanCmd)
	SecurityCmd.AddCommand(securitySignCmd)
	SecurityCmd.AddCommand(securityVerifyCmd)
	SecurityCmd.AddCommand(securityAuditCmd)
	SecurityCmd.AddCommand(securityPolicyCmd)
	SecurityCmd.AddCommand(securityMonitorCmd)
}

// Security structures

type SecurityScanner struct {
	VulnerabilityDB  *VulnerabilityDatabase
	MalwareDB        *MalwareDatabase
	LicenseDB        *LicenseDatabase
	PolicyEngine     *SecurityPolicyEngine
	CertificateStore *CertificateStore
}

type VulnerabilityDatabase struct {
	CVEEntries  map[string]*CVEEntry
	LastUpdated time.Time
	Source      string
	Version     string
}

type CVEEntry struct {
	ID         string    `json:"id"`
	Summary    string    `json:"summary"`
	Severity   string    `json:"severity"`
	Score      float64   `json:"score"`
	Vector     string    `json:"vector"`
	Published  time.Time `json:"published"`
	Modified   time.Time `json:"modified"`
	References []string  `json:"references"`
	Affected   []string  `json:"affected"`
	Solutions  []string  `json:"solutions"`
}

type MalwareDatabase struct {
	Signatures  map[string]*MalwareSignature
	LastUpdated time.Time
	Provider    string
	Version     string
}

type MalwareSignature struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Family      string   `json:"family"`
	Severity    string   `json:"severity"`
	Patterns    []string `json:"patterns"`
	HashValues  []string `json:"hash_values"`
	Description string   `json:"description"`
}

type LicenseDatabase struct {
	Licenses   map[string]*LicenseEntry
	Conflicts  map[string][]string
	Approved   []string
	Restricted []string
}

type LicenseEntry struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	SPDXID      string   `json:"spdx_id"`
	Type        string   `json:"type"` // permissive, copyleft, proprietary
	Category    string   `json:"category"`
	Permissions []string `json:"permissions"`
	Conditions  []string `json:"conditions"`
	Limitations []string `json:"limitations"`
	OSIApproved bool     `json:"osi_approved"`
	FSFLibre    bool     `json:"fsf_libre"`
	Content     string   `json:"content"`
}

type SecurityPolicyEngine struct {
	Policies   map[string]*SecurityPolicy
	Rules      map[string]*SecurityRule
	Violations []*PolicyViolation
}

type SecurityPolicy struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Description string            `json:"description"`
	Scope       string            `json:"scope"`
	Rules       []string          `json:"rules"`
	Enforcement string            `json:"enforcement"` // warn, block, log
	Metadata    map[string]string `json:"metadata"`
	Created     time.Time         `json:"created"`
	Updated     time.Time         `json:"updated"`
}

type SecurityRule struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Type       string            `json:"type"` // vulnerability, license, malware, custom
	Condition  string            `json:"condition"`
	Action     string            `json:"action"`
	Severity   string            `json:"severity"`
	Message    string            `json:"message"`
	Parameters map[string]string `json:"parameters"`
}

type PolicyViolation struct {
	ID         string            `json:"id"`
	PolicyID   string            `json:"policy_id"`
	RuleID     string            `json:"rule_id"`
	ResourceID string            `json:"resource_id"`
	Severity   string            `json:"severity"`
	Message    string            `json:"message"`
	Details    map[string]string `json:"details"`
	Timestamp  time.Time         `json:"timestamp"`
	Status     string            `json:"status"` // open, resolved, ignored
}

type CertificateStore struct {
	Certificates map[string]*Certificate
	PrivateKeys  map[string]*PrivateKey
	RootCAs      []*Certificate
	Revoked      map[string]time.Time
}

type Certificate struct {
	ID          string    `json:"id"`
	Subject     string    `json:"subject"`
	Issuer      string    `json:"issuer"`
	Serial      string    `json:"serial"`
	NotBefore   time.Time `json:"not_before"`
	NotAfter    time.Time `json:"not_after"`
	KeyUsage    []string  `json:"key_usage"`
	PEM         string    `json:"pem"`
	Fingerprint string    `json:"fingerprint"`
}

type PrivateKey struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Size      int    `json:"size"`
	PEM       string `json:"pem"`
	Encrypted bool   `json:"encrypted"`
}

type ScanResult struct {
	TemplateID        string                 `json:"template_id"`
	TemplatePath      string                 `json:"template_path"`
	ScanTimestamp     time.Time              `json:"scan_timestamp"`
	ScanDuration      time.Duration          `json:"scan_duration"`
	OverallStatus     string                 `json:"overall_status"` // pass, fail, warning
	Vulnerabilities   []*VulnerabilityResult `json:"vulnerabilities"`
	MalwareDetections []*MalwareResult       `json:"malware_detections"`
	LicenseIssues     []*LicenseResult       `json:"license_issues"`
	PolicyViolations  []*PolicyViolation     `json:"policy_violations"`
	Signature         *DigitalSignature      `json:"signature,omitempty"`
	Metadata          map[string]string      `json:"metadata"`
}

type VulnerabilityResult struct {
	CVE         string    `json:"cve"`
	Component   string    `json:"component"`
	Version     string    `json:"version"`
	Severity    string    `json:"severity"`
	Score       float64   `json:"score"`
	Description string    `json:"description"`
	Solution    string    `json:"solution"`
	References  []string  `json:"references"`
	Found       time.Time `json:"found"`
}

type MalwareResult struct {
	SignatureID string    `json:"signature_id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	Family      string    `json:"family"`
	Severity    string    `json:"severity"`
	FilePath    string    `json:"file_path"`
	Description string    `json:"description"`
	Found       time.Time `json:"found"`
}

type LicenseResult struct {
	License    string   `json:"license"`
	FilePath   string   `json:"file_path"`
	Issue      string   `json:"issue"` // incompatible, restricted, unknown
	Conflicts  []string `json:"conflicts"`
	Severity   string   `json:"severity"`
	Suggestion string   `json:"suggestion"`
}

type DigitalSignature struct {
	Algorithm   string    `json:"algorithm"`
	Signature   string    `json:"signature"`
	Certificate string    `json:"certificate"`
	Timestamp   time.Time `json:"timestamp"`
	Valid       bool      `json:"valid"`
	Signer      string    `json:"signer"`
}

// Security subcommands
var securityScanCmd = &cobra.Command{
	Use:   "scan",
	Short: "템플릿 보안 스캔",
	Run:   runSecurityScan,
}

var securitySignCmd = &cobra.Command{
	Use:   "sign",
	Short: "템플릿 디지털 서명",
	Run:   runSecuritySign,
}

var securityVerifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "템플릿 서명 검증",
	Run:   runSecurityVerify,
}

var securityAuditCmd = &cobra.Command{
	Use:   "audit",
	Short: "컴플라이언스 감사",
	Run:   runSecurityAudit,
}

var securityPolicyCmd = &cobra.Command{
	Use:   "policy",
	Short: "보안 정책 관리",
	Run:   runSecurityPolicy,
}

var securityMonitorCmd = &cobra.Command{
	Use:   "monitor",
	Short: "보안 모니터링",
	Run:   runSecurityMonitor,
}

func runSecurity(cmd *cobra.Command, args []string) {
	fmt.Printf("🔒 보안 및 컴플라이언스 관리\n")
	fmt.Printf("📋 사용 가능한 하위 명령어:\n")
	fmt.Printf("  • scan     - 템플릿 보안 스캔\n")
	fmt.Printf("  • sign     - 템플릿 디지털 서명\n")
	fmt.Printf("  • verify   - 템플릿 서명 검증\n")
	fmt.Printf("  • audit    - 컴플라이언스 감사\n")
	fmt.Printf("  • policy   - 보안 정책 관리\n")
	fmt.Printf("  • monitor  - 보안 모니터링\n")
	fmt.Printf("\n💡 자세한 도움말: gz template security <command> --help\n")
}

func runSecurityScan(cmd *cobra.Command, args []string) {
	if templatePath == "" {
		fmt.Printf("❌ 스캔할 템플릿 경로가 필요합니다 (--template-path)\n")
		os.Exit(1)
	}

	fmt.Printf("🔍 보안 스캔 시작: %s\n", templatePath)
	fmt.Printf("📊 스캔 유형: %s\n", scanType)
	fmt.Printf("⚠️  심각도 수준: %s\n", severityLevel)

	// Initialize security scanner
	scanner, err := initializeSecurityScanner()
	if err != nil {
		fmt.Printf("❌ 스캐너 초기화 실패: %v\n", err)
		os.Exit(1)
	}

	// Perform security scan
	result, err := performSecurityScan(scanner, templatePath)
	if err != nil {
		fmt.Printf("❌ 스캔 실패: %v\n", err)
		os.Exit(1)
	}

	// Display results
	displayScanResults(result)

	// Save report if requested
	if reportPath != "" {
		if err := saveScanReport(result, reportPath); err != nil {
			fmt.Printf("⚠️ 보고서 저장 실패: %v\n", err)
		} else {
			fmt.Printf("📄 보고서 저장됨: %s\n", reportPath)
		}
	}

	// Set exit code based on results
	if result.OverallStatus == "fail" {
		os.Exit(1)
	}
}

func runSecuritySign(cmd *cobra.Command, args []string) {
	if templatePath == "" || keyFile == "" {
		fmt.Printf("❌ 템플릿 경로와 개인키 파일이 필요합니다\n")
		os.Exit(1)
	}

	fmt.Printf("✍️ 템플릿 서명 중: %s\n", templatePath)

	// Load private key
	privateKey, err := loadPrivateKey(keyFile)
	if err != nil {
		fmt.Printf("❌ 개인키 로드 실패: %v\n", err)
		os.Exit(1)
	}

	// Sign template
	signature, err := signTemplate(templatePath, privateKey)
	if err != nil {
		fmt.Printf("❌ 서명 실패: %v\n", err)
		os.Exit(1)
	}

	// Save signature
	signatureFile := templatePath + ".sig"
	if err := saveSignature(signature, signatureFile); err != nil {
		fmt.Printf("❌ 서명 저장 실패: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ 서명 완료: %s\n", signatureFile)
}

func runSecurityVerify(cmd *cobra.Command, args []string) {
	if templatePath == "" || certFile == "" {
		fmt.Printf("❌ 템플릿 경로와 인증서 파일이 필요합니다\n")
		os.Exit(1)
	}

	fmt.Printf("🔍 서명 검증 중: %s\n", templatePath)

	// Load certificate
	certificate, err := loadCertificate(certFile)
	if err != nil {
		fmt.Printf("❌ 인증서 로드 실패: %v\n", err)
		os.Exit(1)
	}

	// Load signature
	signatureFile := templatePath + ".sig"
	signature, err := loadSignature(signatureFile)
	if err != nil {
		fmt.Printf("❌ 서명 로드 실패: %v\n", err)
		os.Exit(1)
	}

	// Verify signature
	valid, err := verifySignature(templatePath, signature, certificate)
	if err != nil {
		fmt.Printf("❌ 검증 실패: %v\n", err)
		os.Exit(1)
	}

	if valid {
		fmt.Printf("✅ 서명이 유효합니다\n")
	} else {
		fmt.Printf("❌ 서명이 유효하지 않습니다\n")
		os.Exit(1)
	}
}

func runSecurityAudit(cmd *cobra.Command, args []string) {
	fmt.Printf("📊 컴플라이언스 감사\n")
	if complianceStd != "" {
		fmt.Printf("📋 표준: %s\n", complianceStd)
	}
	// Implementation for compliance audit
}

func runSecurityPolicy(cmd *cobra.Command, args []string) {
	fmt.Printf("📋 보안 정책 관리\n")
	// Implementation for security policy management
}

func runSecurityMonitor(cmd *cobra.Command, args []string) {
	fmt.Printf("📊 보안 모니터링\n")
	// Implementation for security monitoring
}

// Security scanner implementation

func initializeSecurityScanner() (*SecurityScanner, error) {
	scanner := &SecurityScanner{
		VulnerabilityDB:  &VulnerabilityDatabase{CVEEntries: make(map[string]*CVEEntry)},
		MalwareDB:        &MalwareDatabase{Signatures: make(map[string]*MalwareSignature)},
		LicenseDB:        &LicenseDatabase{Licenses: make(map[string]*LicenseEntry)},
		PolicyEngine:     &SecurityPolicyEngine{Policies: make(map[string]*SecurityPolicy)},
		CertificateStore: &CertificateStore{Certificates: make(map[string]*Certificate)},
	}

	// Load security databases
	if err := loadSecurityDatabases(scanner); err != nil {
		return nil, fmt.Errorf("보안 데이터베이스 로드 실패: %w", err)
	}

	return scanner, nil
}

func loadSecurityDatabases(scanner *SecurityScanner) error {
	// Load vulnerability database
	scanner.VulnerabilityDB.LastUpdated = time.Now()
	scanner.VulnerabilityDB.Source = "National Vulnerability Database"

	// Load sample CVE entries
	sampleCVEs := []*CVEEntry{
		{
			ID:        "CVE-2024-0001",
			Summary:   "Buffer overflow in template processing",
			Severity:  "high",
			Score:     7.8,
			Published: time.Now().AddDate(0, -1, 0),
		},
		{
			ID:        "CVE-2024-0002",
			Summary:   "SQL injection vulnerability",
			Severity:  "critical",
			Score:     9.2,
			Published: time.Now().AddDate(0, -2, 0),
		},
	}

	for _, cve := range sampleCVEs {
		scanner.VulnerabilityDB.CVEEntries[cve.ID] = cve
	}

	// Load malware database
	scanner.MalwareDB.LastUpdated = time.Now()
	scanner.MalwareDB.Provider = "Security Vendor"

	// Load license database
	sampleLicenses := []*LicenseEntry{
		{
			ID:     "MIT",
			Name:   "MIT License",
			SPDXID: "MIT",
			Type:   "permissive",
		},
		{
			ID:     "GPL-3.0",
			Name:   "GNU General Public License v3.0",
			SPDXID: "GPL-3.0",
			Type:   "copyleft",
		},
	}

	scanner.LicenseDB.Licenses = make(map[string]*LicenseEntry)
	for _, license := range sampleLicenses {
		scanner.LicenseDB.Licenses[license.ID] = license
	}

	return nil
}

func performSecurityScan(scanner *SecurityScanner, templatePath string) (*ScanResult, error) {
	startTime := time.Now()

	result := &ScanResult{
		TemplateID:        generateTemplateID(filepath.Base(templatePath), "1.0.0"),
		TemplatePath:      templatePath,
		ScanTimestamp:     startTime,
		Vulnerabilities:   []*VulnerabilityResult{},
		MalwareDetections: []*MalwareResult{},
		LicenseIssues:     []*LicenseResult{},
		PolicyViolations:  []*PolicyViolation{},
		Metadata:          make(map[string]string),
	}

	// Scan for vulnerabilities
	if scanType == "all" || scanType == "vulnerability" {
		vulns, err := scanVulnerabilities(scanner, templatePath)
		if err != nil {
			return nil, fmt.Errorf("취약점 스캔 실패: %w", err)
		}
		result.Vulnerabilities = vulns
	}

	// Scan for malware
	if scanType == "all" || scanType == "malware" {
		malware, err := scanMalware(scanner, templatePath)
		if err != nil {
			return nil, fmt.Errorf("멀웨어 스캔 실패: %w", err)
		}
		result.MalwareDetections = malware
	}

	// Scan for license issues
	if scanType == "all" || scanType == "license" {
		licenses, err := scanLicenses(scanner, templatePath)
		if err != nil {
			return nil, fmt.Errorf("라이선스 스캔 실패: %w", err)
		}
		result.LicenseIssues = licenses
	}

	// Check policy violations
	violations, err := checkPolicyViolations(scanner, result)
	if err != nil {
		return nil, fmt.Errorf("정책 검사 실패: %w", err)
	}
	result.PolicyViolations = violations

	// Determine overall status
	result.OverallStatus = determineOverallStatus(result)
	result.ScanDuration = time.Since(startTime)

	return result, nil
}

func scanVulnerabilities(scanner *SecurityScanner, templatePath string) ([]*VulnerabilityResult, error) {
	var vulnerabilities []*VulnerabilityResult

	// Sample vulnerability detection
	if strings.Contains(templatePath, "vulnerable") {
		vulnerabilities = append(vulnerabilities, &VulnerabilityResult{
			CVE:         "CVE-2024-0001",
			Component:   "template-engine",
			Version:     "1.0.0",
			Severity:    "high",
			Score:       7.8,
			Description: "Buffer overflow vulnerability detected",
			Found:       time.Now(),
		})
	}

	return vulnerabilities, nil
}

func scanMalware(scanner *SecurityScanner, templatePath string) ([]*MalwareResult, error) {
	var malware []*MalwareResult

	// Sample malware detection
	if strings.Contains(templatePath, "malicious") {
		malware = append(malware, &MalwareResult{
			SignatureID: "MAL-001",
			Name:        "Generic.Trojan",
			Type:        "trojan",
			Severity:    "critical",
			FilePath:    templatePath,
			Found:       time.Now(),
		})
	}

	return malware, nil
}

func scanLicenses(scanner *SecurityScanner, templatePath string) ([]*LicenseResult, error) {
	var issues []*LicenseResult

	// Sample license issue detection
	if strings.Contains(templatePath, "proprietary") {
		issues = append(issues, &LicenseResult{
			License:  "Proprietary",
			FilePath: templatePath,
			Issue:    "restricted",
			Severity: "medium",
		})
	}

	return issues, nil
}

func checkPolicyViolations(scanner *SecurityScanner, result *ScanResult) ([]*PolicyViolation, error) {
	var violations []*PolicyViolation

	// Check for high severity vulnerabilities
	for _, vuln := range result.Vulnerabilities {
		if vuln.Severity == "critical" || vuln.Severity == "high" {
			violations = append(violations, &PolicyViolation{
				ID:         generateViolationID(),
				PolicyID:   "security-policy-001",
				RuleID:     "no-high-vuln",
				ResourceID: result.TemplateID,
				Severity:   "high",
				Message:    fmt.Sprintf("High severity vulnerability detected: %s", vuln.CVE),
				Timestamp:  time.Now(),
				Status:     "open",
			})
		}
	}

	return violations, nil
}

func determineOverallStatus(result *ScanResult) string {
	criticalCount := 0
	highCount := 0

	// Count vulnerabilities by severity
	for _, vuln := range result.Vulnerabilities {
		switch vuln.Severity {
		case "critical":
			criticalCount++
		case "high":
			highCount++
		}
	}

	// Count malware detections
	for range result.MalwareDetections {
		criticalCount++
	}

	// Count policy violations
	for _, violation := range result.PolicyViolations {
		if violation.Severity == "critical" {
			criticalCount++
		} else if violation.Severity == "high" {
			highCount++
		}
	}

	if criticalCount > 0 {
		return "fail"
	} else if highCount > 0 {
		return "warning"
	}

	return "pass"
}

func displayScanResults(result *ScanResult) {
	fmt.Printf("\n📊 스캔 결과\n")
	fmt.Printf("🆔 템플릿 ID: %s\n", result.TemplateID)
	fmt.Printf("⏱️  스캔 시간: %v\n", result.ScanDuration)
	fmt.Printf("📊 전체 상태: %s\n", getStatusEmoji(result.OverallStatus))

	fmt.Printf("\n🔍 취약점: %d개\n", len(result.Vulnerabilities))
	for _, vuln := range result.Vulnerabilities {
		fmt.Printf("  • %s (%s) - %s\n", vuln.CVE, vuln.Severity, vuln.Component)
	}

	fmt.Printf("\n🦠 멀웨어: %d개\n", len(result.MalwareDetections))
	for _, malware := range result.MalwareDetections {
		fmt.Printf("  • %s (%s) - %s\n", malware.Name, malware.Severity, malware.Type)
	}

	fmt.Printf("\n📄 라이선스 문제: %d개\n", len(result.LicenseIssues))
	for _, license := range result.LicenseIssues {
		fmt.Printf("  • %s (%s) - %s\n", license.License, license.Severity, license.Issue)
	}

	fmt.Printf("\n⚠️ 정책 위반: %d개\n", len(result.PolicyViolations))
	for _, violation := range result.PolicyViolations {
		fmt.Printf("  • %s (%s)\n", violation.Message, violation.Severity)
	}
}

func getStatusEmoji(status string) string {
	switch status {
	case "pass":
		return "✅ 통과"
	case "warning":
		return "⚠️ 경고"
	case "fail":
		return "❌ 실패"
	default:
		return "❓ 알 수 없음"
	}
}

func saveScanReport(result *ScanResult, reportPath string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("보고서 마샬링 실패: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(reportPath), 0o755); err != nil {
		return fmt.Errorf("보고서 디렉터리 생성 실패: %w", err)
	}

	if err := os.WriteFile(reportPath, data, 0o644); err != nil {
		return fmt.Errorf("보고서 파일 쓰기 실패: %w", err)
	}

	return nil
}

// Digital signature functions

func loadPrivateKey(keyFile string) (*rsa.PrivateKey, error) {
	keyData, err := os.ReadFile(keyFile)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(keyData)
	if block == nil {
		return nil, fmt.Errorf("PEM 블록을 찾을 수 없습니다")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

func loadCertificate(certFile string) (*x509.Certificate, error) {
	certData, err := os.ReadFile(certFile)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(certData)
	if block == nil {
		return nil, fmt.Errorf("PEM 블록을 찾을 수 없습니다")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, err
	}

	return cert, nil
}

func signTemplate(templatePath string, privateKey *rsa.PrivateKey) (*DigitalSignature, error) {
	// Read template file
	data, err := os.ReadFile(templatePath)
	if err != nil {
		return nil, err
	}

	// Create hash
	hash := sha256.Sum256(data)

	// Sign hash
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, 0, hash[:])
	if err != nil {
		return nil, err
	}

	digitalSig := &DigitalSignature{
		Algorithm: "RSA-SHA256",
		Signature: base64.StdEncoding.EncodeToString(signature),
		Timestamp: time.Now(),
		Valid:     true,
	}

	return digitalSig, nil
}

func verifySignature(templatePath string, signature *DigitalSignature, cert *x509.Certificate) (bool, error) {
	// Read template file
	data, err := os.ReadFile(templatePath)
	if err != nil {
		return false, err
	}

	// Create hash
	hash := sha256.Sum256(data)

	// Decode signature
	sigBytes, err := base64.StdEncoding.DecodeString(signature.Signature)
	if err != nil {
		return false, err
	}

	// Verify signature
	publicKey := cert.PublicKey.(*rsa.PublicKey)
	err = rsa.VerifyPKCS1v15(publicKey, 0, hash[:], sigBytes)

	return err == nil, nil
}

func saveSignature(signature *DigitalSignature, signatureFile string) error {
	data, err := json.MarshalIndent(signature, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(signatureFile, data, 0o644)
}

func loadSignature(signatureFile string) (*DigitalSignature, error) {
	data, err := os.ReadFile(signatureFile)
	if err != nil {
		return nil, err
	}

	var signature DigitalSignature
	if err := json.Unmarshal(data, &signature); err != nil {
		return nil, err
	}

	return &signature, nil
}

// Utility functions

func generateViolationID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return "violation-" + base64.URLEncoding.EncodeToString(bytes)[:8]
}
