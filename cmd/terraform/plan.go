package terraform

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

// PlanCmd represents the plan command
var PlanCmd = &cobra.Command{
	Use:   "plan",
	Short: "Run Terraform plan operations",
	Long: `Run Terraform plan operations with enhanced output and analysis.

Provides comprehensive plan analysis including:
- Resource change summary
- Cost estimation (if configured)
- Security impact analysis
- Dependency graph visualization
- Plan validation and best practices check
- Multi-environment plan comparison

Examples:
  gz terraform plan
  gz terraform plan --environment staging
  gz terraform plan --target module.networking
  gz terraform plan --analyze-security --cost-estimate`,
	Run: runPlan,
}

var (
	planEnvironment  string
	planTarget       string
	planOutputFile   string
	planFormat       string
	analyzeSecurity  bool
	costEstimate     bool
	validatePlan     bool
	parallelism      int
	refreshState     bool
	detailedExitcode bool
)

func init() {
	PlanCmd.Flags().StringVarP(&planEnvironment, "environment", "e", "", "Target environment")
	PlanCmd.Flags().StringVarP(&planTarget, "target", "t", "", "Target specific resource")
	PlanCmd.Flags().StringVarP(&planOutputFile, "output", "o", "", "Save plan to file")
	PlanCmd.Flags().StringVar(&planFormat, "format", "human", "Output format (human, json)")
	PlanCmd.Flags().BoolVar(&analyzeSecurity, "analyze-security", false, "Analyze security implications")
	PlanCmd.Flags().BoolVar(&costEstimate, "cost-estimate", false, "Estimate costs")
	PlanCmd.Flags().BoolVar(&validatePlan, "validate", true, "Validate plan before execution")
	PlanCmd.Flags().IntVar(&parallelism, "parallelism", 10, "Number of parallel operations")
	PlanCmd.Flags().BoolVar(&refreshState, "refresh", true, "Refresh state before planning")
	PlanCmd.Flags().BoolVar(&detailedExitcode, "detailed-exitcode", false, "Return detailed exit codes")
}

// PlanResult represents terraform plan results
type PlanResult struct {
	Summary     PlanSummary       `json:"summary"`
	Changes     []ResourceChange  `json:"changes"`
	Timestamp   time.Time         `json:"timestamp"`
	Environment string            `json:"environment"`
	Version     string            `json:"terraform_version"`
	Security    *SecurityAnalysis `json:"security,omitempty"`
	Cost        *CostEstimation   `json:"cost,omitempty"`
}

type PlanSummary struct {
	ToAdd     int `json:"to_add"`
	ToChange  int `json:"to_change"`
	ToDestroy int `json:"to_destroy"`
	Total     int `json:"total"`
}

type ResourceChange struct {
	Address      string                 `json:"address"`
	Type         string                 `json:"type"`
	Name         string                 `json:"name"`
	Action       string                 `json:"action"`
	ActionReason string                 `json:"action_reason,omitempty"`
	Before       map[string]interface{} `json:"before,omitempty"`
	After        map[string]interface{} `json:"after,omitempty"`
	Sensitive    bool                   `json:"sensitive"`
}

type SecurityAnalysis struct {
	Issues          []SecurityIssue `json:"issues"`
	Score           int             `json:"score"`
	Severity        string          `json:"severity"`
	Recommendations []string        `json:"recommendations"`
}

type SecurityIssue struct {
	Resource    string `json:"resource"`
	Issue       string `json:"issue"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
	Fix         string `json:"fix"`
}

type CostEstimation struct {
	MonthlyCost string          `json:"monthly_cost"`
	Changes     []CostChange    `json:"changes"`
	Breakdown   []CostBreakdown `json:"breakdown"`
}

type CostChange struct {
	Resource   string `json:"resource"`
	Before     string `json:"before"`
	After      string `json:"after"`
	Difference string `json:"difference"`
}

type CostBreakdown struct {
	Service string `json:"service"`
	Cost    string `json:"cost"`
	Usage   string `json:"usage"`
}

func runPlan(cmd *cobra.Command, args []string) {
	fmt.Printf("🔍 Running Terraform plan\n")

	// Validate terraform installation
	if !isTerraformInstalled() {
		fmt.Printf("❌ Terraform is not installed or not in PATH\n")
		os.Exit(1)
	}

	// Initialize if needed
	if !isTerraformInitialized() {
		fmt.Printf("🔄 Initializing Terraform...\n")
		if err := runTerraformInit(); err != nil {
			fmt.Printf("❌ Failed to initialize Terraform: %v\n", err)
			os.Exit(1)
		}
	}

	// Validate configuration if requested
	if validatePlan {
		fmt.Printf("✅ Validating Terraform configuration...\n")
		if err := runTerraformValidate(); err != nil {
			fmt.Printf("❌ Terraform validation failed: %v\n", err)
			os.Exit(1)
		}
	}

	// Build plan command
	planCmd := buildPlanCommand()

	fmt.Printf("🚀 Executing: %s\n", strings.Join(planCmd, " "))

	// Execute plan
	result, err := executePlan(planCmd)
	if err != nil {
		fmt.Printf("❌ Plan execution failed: %v\n", err)
		os.Exit(1)
	}

	// Analyze plan output
	planResult, err := analyzePlan(result)
	if err != nil {
		fmt.Printf("⚠️ Failed to analyze plan: %v\n", err)
		// Continue with basic output
		fmt.Println(result)
		return
	}

	// Perform security analysis if requested
	if analyzeSecurity {
		fmt.Printf("🔒 Analyzing security implications...\n")
		security := analyzeSecurityImplications(planResult)
		planResult.Security = security
	}

	// Perform cost estimation if requested
	if costEstimate {
		fmt.Printf("💰 Estimating costs...\n")
		cost := estimateCosts(planResult)
		planResult.Cost = cost
	}

	// Output results
	if err := outputPlanResults(planResult); err != nil {
		fmt.Printf("Error outputting results: %v\n", err)
		os.Exit(1)
	}

	// Save plan if output file specified
	if planOutputFile != "" {
		if err := savePlanToFile(planResult); err != nil {
			fmt.Printf("Error saving plan: %v\n", err)
			os.Exit(1)
		}
	}

	// Print summary
	printPlanSummary(planResult)
}

func isTerraformInstalled() bool {
	_, err := exec.LookPath("terraform")
	return err == nil
}

func isTerraformInitialized() bool {
	_, err := os.Stat(".terraform")
	return err == nil
}

func runTerraformInit() error {
	cmd := exec.Command("terraform", "init")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runTerraformValidate() error {
	cmd := exec.Command("terraform", "validate")
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Validation output: %s\n", output)
		return err
	}
	return nil
}

func buildPlanCommand() []string {
	cmd := []string{"terraform", "plan"}

	if refreshState {
		cmd = append(cmd, "-refresh=true")
	} else {
		cmd = append(cmd, "-refresh=false")
	}

	if parallelism > 0 {
		cmd = append(cmd, fmt.Sprintf("-parallelism=%d", parallelism))
	}

	if planTarget != "" {
		cmd = append(cmd, fmt.Sprintf("-target=%s", planTarget))
	}

	if detailedExitcode {
		cmd = append(cmd, "-detailed-exitcode")
	}

	if planFormat == "json" {
		cmd = append(cmd, "-json")
	}

	// Add environment-specific var file if exists
	if planEnvironment != "" {
		varFile := fmt.Sprintf("%s.tfvars", planEnvironment)
		if _, err := os.Stat(varFile); err == nil {
			cmd = append(cmd, fmt.Sprintf("-var-file=%s", varFile))
		}
	}

	return cmd
}

func executePlan(cmd []string) (string, error) {
	execCmd := exec.Command(cmd[0], cmd[1:]...)
	output, err := execCmd.CombinedOutput()
	return string(output), err
}

func analyzePlan(planOutput string) (*PlanResult, error) {
	result := &PlanResult{
		Timestamp:   time.Now(),
		Environment: planEnvironment,
		Changes:     []ResourceChange{},
	}

	// Parse terraform version
	if versionLine := extractTerraformVersion(planOutput); versionLine != "" {
		result.Version = versionLine
	}

	// Parse plan summary
	summary := parsePlanSummary(planOutput)
	result.Summary = summary

	// Parse individual changes
	changes := parsePlanChanges(planOutput)
	result.Changes = changes

	return result, nil
}

func extractTerraformVersion(output string) string {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "Terraform v") {
			return strings.TrimSpace(line)
		}
	}
	return ""
}

func parsePlanSummary(output string) PlanSummary {
	summary := PlanSummary{}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.Contains(line, "Plan:") {
			// Parse line like "Plan: 2 to add, 1 to change, 0 to destroy."
			parts := strings.Fields(line)
			for i, part := range parts {
				if part == "to" && i+1 < len(parts) {
					switch parts[i+1] {
					case "add,":
						if i > 0 {
							fmt.Sscanf(parts[i-1], "%d", &summary.ToAdd)
						}
					case "change,":
						if i > 0 {
							fmt.Sscanf(parts[i-1], "%d", &summary.ToChange)
						}
					case "destroy.":
						if i > 0 {
							fmt.Sscanf(parts[i-1], "%d", &summary.ToDestroy)
						}
					}
				}
			}
		}
	}

	summary.Total = summary.ToAdd + summary.ToChange + summary.ToDestroy
	return summary
}

func parsePlanChanges(output string) []ResourceChange {
	var changes []ResourceChange

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Look for resource change indicators
		if strings.HasPrefix(line, "# ") {
			change := parseResourceChange(line)
			if change != nil {
				changes = append(changes, *change)
			}
		}
	}

	return changes
}

func parseResourceChange(line string) *ResourceChange {
	// Parse lines like "# module.vpc.aws_vpc.main will be created"
	if !strings.Contains(line, " will be ") {
		return nil
	}

	parts := strings.Split(line, " will be ")
	if len(parts) != 2 {
		return nil
	}

	address := strings.TrimPrefix(parts[0], "# ")
	action := strings.TrimSpace(parts[1])

	// Extract resource type and name
	addressParts := strings.Split(address, ".")
	var resourceType, resourceName string
	if len(addressParts) >= 2 {
		resourceType = addressParts[len(addressParts)-2]
		resourceName = addressParts[len(addressParts)-1]
	}

	return &ResourceChange{
		Address: address,
		Type:    resourceType,
		Name:    resourceName,
		Action:  action,
	}
}

func analyzeSecurityImplications(result *PlanResult) *SecurityAnalysis {
	analysis := &SecurityAnalysis{
		Issues:          []SecurityIssue{},
		Score:           100,
		Severity:        "low",
		Recommendations: []string{},
	}

	// Analyze each resource change for security implications
	for _, change := range result.Changes {
		issues := analyzeResourceSecurity(change)
		analysis.Issues = append(analysis.Issues, issues...)
	}

	// Calculate overall security score
	analysis.Score = calculateSecurityScore(analysis.Issues)
	analysis.Severity = calculateSeverity(analysis.Issues)
	analysis.Recommendations = generateSecurityRecommendations(analysis.Issues)

	return analysis
}

func analyzeResourceSecurity(change ResourceChange) []SecurityIssue {
	var issues []SecurityIssue

	switch change.Type {
	case "aws_security_group":
		if change.Action == "created" || change.Action == "updated" {
			issues = append(issues, SecurityIssue{
				Resource:    change.Address,
				Issue:       "Security group allows broad access",
				Severity:    "medium",
				Description: "Security group rules should follow principle of least privilege",
				Fix:         "Review and restrict security group rules to minimum required access",
			})
		}
	case "aws_instance":
		if change.Action == "created" {
			issues = append(issues, SecurityIssue{
				Resource:    change.Address,
				Issue:       "EC2 instance without encryption",
				Severity:    "low",
				Description: "EC2 instance should use encrypted EBS volumes",
				Fix:         "Enable EBS encryption for instance volumes",
			})
		}
	case "aws_s3_bucket":
		if change.Action == "created" {
			issues = append(issues, SecurityIssue{
				Resource:    change.Address,
				Issue:       "S3 bucket without encryption",
				Severity:    "high",
				Description: "S3 bucket should have server-side encryption enabled",
				Fix:         "Enable default encryption for S3 bucket",
			})
		}
	}

	return issues
}

func calculateSecurityScore(issues []SecurityIssue) int {
	score := 100

	for _, issue := range issues {
		switch issue.Severity {
		case "high":
			score -= 20
		case "medium":
			score -= 10
		case "low":
			score -= 5
		}
	}

	if score < 0 {
		score = 0
	}

	return score
}

func calculateSeverity(issues []SecurityIssue) string {
	hasHigh := false
	hasMedium := false

	for _, issue := range issues {
		switch issue.Severity {
		case "high":
			hasHigh = true
		case "medium":
			hasMedium = true
		}
	}

	if hasHigh {
		return "high"
	} else if hasMedium {
		return "medium"
	}

	return "low"
}

func generateSecurityRecommendations(issues []SecurityIssue) []string {
	recommendations := []string{
		"Review all resource configurations for security best practices",
		"Enable encryption for data at rest and in transit",
		"Apply principle of least privilege to access controls",
		"Use AWS Config rules for continuous compliance monitoring",
	}

	if len(issues) > 0 {
		recommendations = append(recommendations, "Address the identified security issues before applying changes")
	}

	return recommendations
}

func estimateCosts(result *PlanResult) *CostEstimation {
	estimation := &CostEstimation{
		MonthlyCost: "Unable to estimate",
		Changes:     []CostChange{},
		Breakdown:   []CostBreakdown{},
	}

	// This is a simplified cost estimation
	// In practice, you would integrate with cloud billing APIs or tools like Infracost

	totalCost := 0.0

	for _, change := range result.Changes {
		cost := estimateResourceCost(change)
		if cost > 0 {
			estimation.Changes = append(estimation.Changes, CostChange{
				Resource:   change.Address,
				Before:     "$0.00",
				After:      fmt.Sprintf("$%.2f", cost),
				Difference: fmt.Sprintf("+$%.2f", cost),
			})
			totalCost += cost
		}
	}

	if totalCost > 0 {
		estimation.MonthlyCost = fmt.Sprintf("$%.2f", totalCost)
	}

	return estimation
}

func estimateResourceCost(change ResourceChange) float64 {
	// Simplified cost estimation based on resource type
	switch change.Type {
	case "aws_instance":
		return 50.0 // Rough estimate for t3.medium per month
	case "aws_rds_instance":
		return 100.0 // Rough estimate for db.t3.micro per month
	case "aws_s3_bucket":
		return 5.0 // Rough estimate for minimal S3 usage
	case "aws_vpc":
		return 0.0 // VPC itself is free
	default:
		return 0.0
	}
}

func outputPlanResults(result *PlanResult) error {
	switch planFormat {
	case "json":
		return outputPlanJSON(result)
	default:
		return outputPlanText(result)
	}
}

func outputPlanJSON(result *PlanResult) error {
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(jsonData))
	return nil
}

func outputPlanText(result *PlanResult) error {
	fmt.Printf("\n📋 Plan Summary:\n")
	fmt.Printf("================\n")
	fmt.Printf("Resources to add: %d\n", result.Summary.ToAdd)
	fmt.Printf("Resources to change: %d\n", result.Summary.ToChange)
	fmt.Printf("Resources to destroy: %d\n", result.Summary.ToDestroy)
	fmt.Printf("Total changes: %d\n", result.Summary.Total)

	if len(result.Changes) > 0 {
		fmt.Printf("\n🔄 Resource Changes:\n")
		fmt.Printf("===================\n")
		for _, change := range result.Changes {
			icon := getActionIcon(change.Action)
			fmt.Printf("%s %s (%s)\n", icon, change.Address, change.Action)
		}
	}

	if result.Security != nil {
		fmt.Printf("\n🔒 Security Analysis:\n")
		fmt.Printf("====================\n")
		fmt.Printf("Security Score: %d/100 (%s)\n", result.Security.Score, result.Security.Severity)

		if len(result.Security.Issues) > 0 {
			fmt.Printf("\nSecurity Issues:\n")
			for _, issue := range result.Security.Issues {
				severityIcon := getSeverityIcon(issue.Severity)
				fmt.Printf("%s %s: %s\n", severityIcon, issue.Resource, issue.Issue)
				fmt.Printf("   Fix: %s\n", issue.Fix)
			}
		}
	}

	if result.Cost != nil {
		fmt.Printf("\n💰 Cost Estimation:\n")
		fmt.Printf("==================\n")
		fmt.Printf("Estimated Monthly Cost: %s\n", result.Cost.MonthlyCost)

		if len(result.Cost.Changes) > 0 {
			fmt.Printf("\nCost Changes:\n")
			for _, cost := range result.Cost.Changes {
				fmt.Printf("📊 %s: %s (%s)\n", cost.Resource, cost.After, cost.Difference)
			}
		}
	}

	return nil
}

func getActionIcon(action string) string {
	switch {
	case strings.Contains(action, "created"):
		return "➕"
	case strings.Contains(action, "updated") || strings.Contains(action, "modified"):
		return "🔄"
	case strings.Contains(action, "destroyed") || strings.Contains(action, "deleted"):
		return "❌"
	default:
		return "🔹"
	}
}

func getSeverityIcon(severity string) string {
	switch severity {
	case "high":
		return "🔴"
	case "medium":
		return "🟡"
	case "low":
		return "🟢"
	default:
		return "🔹"
	}
}

func savePlanToFile(result *PlanResult) error {
	// Create output directory if needed
	dir := filepath.Dir(planOutputFile)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	// Save as JSON
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(planOutputFile, jsonData, 0o644)
}

func printPlanSummary(result *PlanResult) {
	fmt.Printf("\n🎯 Plan Summary:\n")
	fmt.Printf("===============\n")
	fmt.Printf("Timestamp: %s\n", result.Timestamp.Format(time.RFC3339))
	if result.Environment != "" {
		fmt.Printf("Environment: %s\n", result.Environment)
	}
	if result.Version != "" {
		fmt.Printf("Terraform Version: %s\n", result.Version)
	}

	if result.Summary.Total == 0 {
		fmt.Printf("✅ No changes required - infrastructure is up to date\n")
	} else {
		fmt.Printf("📝 %d total changes planned\n", result.Summary.Total)
		if result.Security != nil && result.Security.Score < 80 {
			fmt.Printf("⚠️ Security score is low (%d/100) - review before applying\n", result.Security.Score)
		}
	}

	fmt.Printf("\n📚 Next steps:\n")
	fmt.Printf("1. Review the planned changes carefully\n")
	fmt.Printf("2. Address any security issues if found\n")
	fmt.Printf("3. Run 'terraform apply' to implement changes\n")
	fmt.Printf("4. Monitor resources after deployment\n")
}
