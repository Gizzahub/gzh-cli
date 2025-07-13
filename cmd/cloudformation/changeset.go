package cloudformation

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/spf13/cobra"
)

// ChangeSetCmd represents the changeset command
var ChangeSetCmd = &cobra.Command{
	Use:   "changeset",
	Short: "Manage CloudFormation change sets",
	Long: `Manage CloudFormation change sets for safe stack updates.

Change sets allow you to preview changes before applying them to your stack:
- Create change sets to preview modifications
- Compare different change sets
- Execute or delete change sets
- Analyze impact and resource replacements
- Support for both stack updates and creation

Examples:
  gz cloudformation changeset create --stack my-stack --template template.yaml
  gz cloudformation changeset list --stack my-stack
  gz cloudformation changeset describe --stack my-stack --changeset my-changeset
  gz cloudformation changeset execute --stack my-stack --changeset my-changeset
  gz cloudformation changeset delete --stack my-stack --changeset my-changeset`,
	Run: runChangeSet,
}

var (
	changesetAction     string
	changesetName       string
	changesetStack      string
	changesetTemplate   string
	changesetParams     string
	changesetRegion     string
	changesetProfile    string
	changesetTags       []string
	changesetCaps       []string
	includeNestedStacks bool
	changesetType       string
	resourcesToImport   string
	clientToken         string
	roleArn             string
	notificationArns    []string
	outputFormat        string
	verbose             bool
)

func init() {
	ChangeSetCmd.Flags().StringVarP(&changesetAction, "action", "a", "create", "Action: create, list, describe, execute, delete")
	ChangeSetCmd.Flags().StringVarP(&changesetName, "changeset", "c", "", "Change set name")
	ChangeSetCmd.Flags().StringVarP(&changesetStack, "stack", "s", "", "Stack name")
	ChangeSetCmd.Flags().StringVarP(&changesetTemplate, "template", "t", "", "Template file path")
	ChangeSetCmd.Flags().StringVarP(&changesetParams, "parameters", "p", "", "Parameters file path")
	ChangeSetCmd.Flags().StringVarP(&changesetRegion, "region", "r", "", "AWS region")
	ChangeSetCmd.Flags().StringVar(&changesetProfile, "profile", "", "AWS profile")
	ChangeSetCmd.Flags().StringSliceVar(&changesetTags, "tags", []string{}, "Stack tags (key=value)")
	ChangeSetCmd.Flags().StringSliceVar(&changesetCaps, "capabilities", []string{}, "Stack capabilities")
	ChangeSetCmd.Flags().BoolVar(&includeNestedStacks, "include-nested", false, "Include nested stacks in change set")
	ChangeSetCmd.Flags().StringVar(&changesetType, "type", "UPDATE", "Change set type (UPDATE, CREATE, IMPORT)")
	ChangeSetCmd.Flags().StringVar(&resourcesToImport, "import-resources", "", "Resources to import (JSON file)")
	ChangeSetCmd.Flags().StringVar(&clientToken, "client-token", "", "Client request token")
	ChangeSetCmd.Flags().StringVar(&roleArn, "role-arn", "", "IAM role ARN")
	ChangeSetCmd.Flags().StringSliceVar(&notificationArns, "notification-arns", []string{}, "SNS notification ARNs")
	ChangeSetCmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table, json, yaml)")
	ChangeSetCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")

	ChangeSetCmd.MarkFlagRequired("stack")
}

// ChangeSetSummary represents change set information
type ChangeSetSummary struct {
	ChangeSetName    string             `json:"changeset_name"`
	ChangeSetId      string             `json:"changeset_id"`
	Status           string             `json:"status"`
	StatusReason     string             `json:"status_reason,omitempty"`
	CreationTime     *time.Time         `json:"creation_time"`
	Description      string             `json:"description,omitempty"`
	ExecutionStatus  string             `json:"execution_status,omitempty"`
	StackName        string             `json:"stack_name"`
	ChangeSetType    string             `json:"changeset_type"`
	Changes          []types.Change     `json:"changes,omitempty"`
	Parameters       []types.Parameter  `json:"parameters,omitempty"`
	Tags             []types.Tag        `json:"tags,omitempty"`
	Capabilities     []types.Capability `json:"capabilities,omitempty"`
	NotificationARNs []string           `json:"notification_arns,omitempty"`
}

// ChangeSetComparison represents comparison between change sets
type ChangeSetComparison struct {
	ChangeSet1      string                  `json:"changeset_1"`
	ChangeSet2      string                  `json:"changeset_2"`
	AddedChanges    []types.Change          `json:"added_changes"`
	RemovedChanges  []types.Change          `json:"removed_changes"`
	ModifiedChanges []ChangeComparison      `json:"modified_changes"`
	ImpactAnalysis  ChangeSetImpactAnalysis `json:"impact_analysis"`
}

// ChangeComparison represents differences between individual changes
type ChangeComparison struct {
	Resource    string       `json:"resource"`
	OldChange   types.Change `json:"old_change"`
	NewChange   types.Change `json:"new_change"`
	Differences []string     `json:"differences"`
}

// ChangeSetImpactAnalysis represents impact analysis of change set
type ChangeSetImpactAnalysis struct {
	ResourceReplacements int      `json:"resource_replacements"`
	ServiceInterruptions int      `json:"service_interruptions"`
	SecurityChanges      int      `json:"security_changes"`
	CostImpact           string   `json:"cost_impact"`
	RiskLevel            string   `json:"risk_level"`
	Recommendations      []string `json:"recommendations"`
	AffectedResources    []string `json:"affected_resources"`
}

func runChangeSet(cmd *cobra.Command, args []string) {
	if changesetStack == "" {
		fmt.Printf("❌ Stack name is required\n")
		cmd.Help()
		os.Exit(1)
	}

	fmt.Printf("🔄 Managing CloudFormation Change Sets\n")
	fmt.Printf("📦 Stack: %s\n", changesetStack)
	fmt.Printf("🎯 Action: %s\n", changesetAction)

	// Initialize AWS config
	ctx := context.Background()
	cfg, err := initChangeSetAWSConfig(ctx)
	if err != nil {
		fmt.Printf("❌ Failed to initialize AWS config: %v\n", err)
		os.Exit(1)
	}

	client := cloudformation.NewFromConfig(cfg)

	// Execute action
	switch changesetAction {
	case "create":
		err = createChangeSet(ctx, client)
	case "list":
		err = listChangeSetsForStack(ctx, client)
	case "describe":
		err = describeChangeSet(ctx, client)
	case "execute":
		err = executeChangeSet(ctx, client)
	case "delete":
		err = deleteChangeSet(ctx, client)
	case "compare":
		err = compareChangeSets(ctx, client)
	case "analyze":
		err = analyzeChangeSetImpact(ctx, client)
	default:
		fmt.Printf("❌ Unknown action: %s\n", changesetAction)
		fmt.Printf("Available actions: create, list, describe, execute, delete, compare, analyze\n")
		os.Exit(1)
	}

	if err != nil {
		fmt.Printf("❌ Operation failed: %v\n", err)
		os.Exit(1)
	}
}

func initChangeSetAWSConfig(ctx context.Context) (aws.Config, error) {
	var options []func(*config.LoadOptions) error

	if changesetRegion != "" {
		options = append(options, config.WithRegion(changesetRegion))
	}

	if changesetProfile != "" {
		options = append(options, config.WithSharedConfigProfile(changesetProfile))
	}

	return config.LoadDefaultConfig(ctx, options...)
}

func createChangeSet(ctx context.Context, client *cloudformation.Client) error {
	if changesetName == "" {
		changesetName = fmt.Sprintf("%s-changeset-%d", changesetStack, time.Now().Unix())
	}

	fmt.Printf("📋 Creating change set: %s\n", changesetName)

	// Read template
	if changesetTemplate == "" {
		return fmt.Errorf("template file is required for creating change set")
	}

	templateBody, err := os.ReadFile(changesetTemplate)
	if err != nil {
		return fmt.Errorf("failed to read template file: %w", err)
	}

	// Parse parameters
	parameters, err := parseChangeSetParameters()
	if err != nil {
		return fmt.Errorf("failed to parse parameters: %w", err)
	}

	// Parse tags
	tags, err := parseChangeSetTags()
	if err != nil {
		return fmt.Errorf("failed to parse tags: %w", err)
	}

	// Parse capabilities
	capabilities := parseChangeSetCapabilities()

	// Parse resources to import
	var resourcesToImportList []types.ResourceToImport
	if resourcesToImport != "" {
		resourcesToImportList, err = parseResourcesToImport()
		if err != nil {
			return fmt.Errorf("failed to parse resources to import: %w", err)
		}
	}

	input := &cloudformation.CreateChangeSetInput{
		ChangeSetName:       aws.String(changesetName),
		StackName:           aws.String(changesetStack),
		TemplateBody:        aws.String(string(templateBody)),
		Parameters:          parameters,
		Tags:                tags,
		Capabilities:        capabilities,
		ChangeSetType:       types.ChangeSetType(changesetType),
		IncludeNestedStacks: aws.Bool(includeNestedStacks),
	}

	if len(resourcesToImportList) > 0 {
		input.ResourcesToImport = resourcesToImportList
	}

	if roleArn != "" {
		input.RoleARN = aws.String(roleArn)
	}

	if clientToken != "" {
		// Note: ClientRequestToken is not available in CreateChangeSetInput
		// input.ClientRequestToken = aws.String(clientToken)
	}

	if len(notificationArns) > 0 {
		input.NotificationARNs = notificationArns
	}

	// Create change set
	output, err := client.CreateChangeSet(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to create change set: %w", err)
	}

	fmt.Printf("✅ Change set created successfully\n")
	fmt.Printf("🆔 Change Set ID: %s\n", aws.ToString(output.Id))

	// Wait for change set creation to complete
	fmt.Printf("⏳ Waiting for change set creation to complete...\n")
	if err := waitForChangeSetCreation(ctx, client, changesetStack, changesetName); err != nil {
		return err
	}

	// Describe the created change set
	return describeSpecificChangeSet(ctx, client, changesetStack, changesetName)
}

func listChangeSetsForStack(ctx context.Context, client *cloudformation.Client) error {
	fmt.Printf("📋 Listing change sets for stack: %s\n", changesetStack)

	input := &cloudformation.ListChangeSetsInput{
		StackName: aws.String(changesetStack),
	}

	output, err := client.ListChangeSets(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to list change sets: %w", err)
	}

	if len(output.Summaries) == 0 {
		fmt.Printf("No change sets found for stack %s\n", changesetStack)
		return nil
	}

	// Convert to our summary format
	var summaries []ChangeSetSummary
	for _, summary := range output.Summaries {
		summaries = append(summaries, ChangeSetSummary{
			ChangeSetName:   aws.ToString(summary.ChangeSetName),
			ChangeSetId:     aws.ToString(summary.ChangeSetId),
			Status:          string(summary.Status),
			StatusReason:    aws.ToString(summary.StatusReason),
			CreationTime:    summary.CreationTime,
			Description:     aws.ToString(summary.Description),
			ExecutionStatus: string(summary.ExecutionStatus),
			StackName:       aws.ToString(summary.StackName),
			ChangeSetType:   "UPDATE", // Default value since ChangeSetType field is not available
		})
	}

	return outputChangeSetList(summaries)
}

func describeChangeSet(ctx context.Context, client *cloudformation.Client) error {
	if changesetName == "" {
		return fmt.Errorf("changeset name is required for describe action")
	}

	return describeSpecificChangeSet(ctx, client, changesetStack, changesetName)
}

func describeSpecificChangeSet(ctx context.Context, client *cloudformation.Client, stackName, changeSetName string) error {
	fmt.Printf("📋 Describing change set: %s\n", changeSetName)

	input := &cloudformation.DescribeChangeSetInput{
		ChangeSetName: aws.String(changeSetName),
		StackName:     aws.String(stackName),
	}

	output, err := client.DescribeChangeSet(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to describe change set: %w", err)
	}

	summary := ChangeSetSummary{
		ChangeSetName:    aws.ToString(output.ChangeSetName),
		ChangeSetId:      aws.ToString(output.ChangeSetId),
		Status:           string(output.Status),
		StatusReason:     aws.ToString(output.StatusReason),
		CreationTime:     output.CreationTime,
		Description:      aws.ToString(output.Description),
		ExecutionStatus:  string(output.ExecutionStatus),
		StackName:        aws.ToString(output.StackName),
		ChangeSetType:    "UPDATE", // Default value since ChangeSetType field is not available
		Changes:          output.Changes,
		Parameters:       output.Parameters,
		Tags:             output.Tags,
		Capabilities:     output.Capabilities,
		NotificationARNs: output.NotificationARNs,
	}

	return outputChangeSetDetails(summary)
}

func executeChangeSet(ctx context.Context, client *cloudformation.Client) error {
	if changesetName == "" {
		return fmt.Errorf("changeset name is required for execute action")
	}

	fmt.Printf("▶️ Executing change set: %s\n", changesetName)

	input := &cloudformation.ExecuteChangeSetInput{
		ChangeSetName: aws.String(changesetName),
		StackName:     aws.String(changesetStack),
	}

	if clientToken != "" {
		// Note: ClientRequestToken is not available in CreateChangeSetInput
		// input.ClientRequestToken = aws.String(clientToken)
	}

	_, err := client.ExecuteChangeSet(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to execute change set: %w", err)
	}

	fmt.Printf("✅ Change set execution initiated\n")
	fmt.Printf("👀 Monitor stack status to track progress\n")

	return nil
}

func deleteChangeSet(ctx context.Context, client *cloudformation.Client) error {
	if changesetName == "" {
		return fmt.Errorf("changeset name is required for delete action")
	}

	fmt.Printf("🗑️ Deleting change set: %s\n", changesetName)

	input := &cloudformation.DeleteChangeSetInput{
		ChangeSetName: aws.String(changesetName),
		StackName:     aws.String(changesetStack),
	}

	_, err := client.DeleteChangeSet(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to delete change set: %w", err)
	}

	fmt.Printf("✅ Change set deleted successfully\n")
	return nil
}

func compareChangeSets(ctx context.Context, client *cloudformation.Client) error {
	// This would require two changeset names to compare
	// For now, implement a basic comparison framework
	fmt.Printf("🔍 Change set comparison feature\n")
	fmt.Printf("This feature compares two change sets to show differences\n")
	// Implementation would require additional parameters for two changeset names
	return fmt.Errorf("comparison feature requires two changeset names - not yet implemented")
}

func analyzeChangeSetImpact(ctx context.Context, client *cloudformation.Client) error {
	if changesetName == "" {
		return fmt.Errorf("changeset name is required for analyze action")
	}

	fmt.Printf("🔍 Analyzing change set impact: %s\n", changesetName)

	// Get change set details
	input := &cloudformation.DescribeChangeSetInput{
		ChangeSetName: aws.String(changesetName),
		StackName:     aws.String(changesetStack),
	}

	output, err := client.DescribeChangeSet(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to describe change set: %w", err)
	}

	// Analyze impact
	analysis := analyzeImpact(output.Changes)
	return outputImpactAnalysis(analysis)
}

func parseChangeSetParameters() ([]types.Parameter, error) {
	if changesetParams == "" {
		return nil, nil
	}

	data, err := os.ReadFile(changesetParams)
	if err != nil {
		return nil, fmt.Errorf("failed to read parameters file: %w", err)
	}

	var paramFile struct {
		Parameters map[string]string `json:"Parameters"`
	}

	if err := json.Unmarshal(data, &paramFile); err != nil {
		return nil, fmt.Errorf("failed to parse parameters file: %w", err)
	}

	var parameters []types.Parameter
	for key, value := range paramFile.Parameters {
		parameters = append(parameters, types.Parameter{
			ParameterKey:   aws.String(key),
			ParameterValue: aws.String(value),
		})
	}

	return parameters, nil
}

func parseChangeSetTags() ([]types.Tag, error) {
	var tags []types.Tag

	for _, tag := range changesetTags {
		parts := strings.SplitN(tag, "=", 2)
		if len(parts) == 2 {
			tags = append(tags, types.Tag{
				Key:   aws.String(parts[0]),
				Value: aws.String(parts[1]),
			})
		}
	}

	return tags, nil
}

func parseChangeSetCapabilities() []types.Capability {
	var capabilities []types.Capability
	for _, cap := range changesetCaps {
		capabilities = append(capabilities, types.Capability(cap))
	}
	return capabilities
}

func parseResourcesToImport() ([]types.ResourceToImport, error) {
	data, err := os.ReadFile(resourcesToImport)
	if err != nil {
		return nil, fmt.Errorf("failed to read resources to import file: %w", err)
	}

	var resources []types.ResourceToImport
	if err := json.Unmarshal(data, &resources); err != nil {
		return nil, fmt.Errorf("failed to parse resources to import: %w", err)
	}

	return resources, nil
}

func analyzeImpact(changes []types.Change) ChangeSetImpactAnalysis {
	analysis := ChangeSetImpactAnalysis{
		Recommendations:   []string{},
		AffectedResources: []string{},
	}

	for _, change := range changes {
		if change.ResourceChange != nil {
			rc := change.ResourceChange
			resourceId := aws.ToString(rc.LogicalResourceId)
			analysis.AffectedResources = append(analysis.AffectedResources, resourceId)

			// Count replacements
			if rc.Replacement == types.ReplacementTrue {
				analysis.ResourceReplacements++
			}

			// Check for service interruptions
			if rc.Replacement == types.ReplacementTrue || change.Type == "Remove" {
				analysis.ServiceInterruptions++
			}

			// Check for security-related changes
			resourceType := aws.ToString(rc.ResourceType)
			if strings.Contains(resourceType, "IAM") || strings.Contains(resourceType, "Security") {
				analysis.SecurityChanges++
			}
		}
	}

	// Determine risk level
	if analysis.ResourceReplacements > 5 || analysis.SecurityChanges > 0 {
		analysis.RiskLevel = "HIGH"
		analysis.Recommendations = append(analysis.Recommendations, "Review all resource replacements carefully")
		analysis.Recommendations = append(analysis.Recommendations, "Test in development environment first")
	} else if analysis.ResourceReplacements > 0 || analysis.ServiceInterruptions > 0 {
		analysis.RiskLevel = "MEDIUM"
		analysis.Recommendations = append(analysis.Recommendations, "Monitor affected resources after deployment")
	} else {
		analysis.RiskLevel = "LOW"
		analysis.Recommendations = append(analysis.Recommendations, "Changes appear safe to deploy")
	}

	// Estimate cost impact
	if analysis.ResourceReplacements > 0 {
		analysis.CostImpact = "MEDIUM - Resource replacements may incur temporary costs"
	} else {
		analysis.CostImpact = "LOW - No resource replacements expected"
	}

	return analysis
}

func outputChangeSetList(summaries []ChangeSetSummary) error {
	switch outputFormat {
	case "json":
		return outputJSON(summaries)
	case "yaml":
		return outputYAML(summaries)
	default:
		return outputTable(summaries)
	}
}

func outputChangeSetDetails(summary ChangeSetSummary) error {
	switch outputFormat {
	case "json":
		return outputJSON(summary)
	case "yaml":
		return outputYAML(summary)
	default:
		return outputChangeSetTable(summary)
	}
}

func outputImpactAnalysis(analysis ChangeSetImpactAnalysis) error {
	switch outputFormat {
	case "json":
		return outputJSON(analysis)
	case "yaml":
		return outputYAML(analysis)
	default:
		return outputImpactTable(analysis)
	}
}

func outputTable(summaries []ChangeSetSummary) error {
	fmt.Printf("\n📋 Change Sets:\n")
	fmt.Printf("═══════════════════════════════════════════════════════════════════════════════════════\n")
	fmt.Printf("%-30s %-20s %-15s %-20s\n", "NAME", "STATUS", "TYPE", "CREATED")
	fmt.Printf("───────────────────────────────────────────────────────────────────────────────────────\n")

	for _, summary := range summaries {
		createdTime := ""
		if summary.CreationTime != nil {
			createdTime = summary.CreationTime.Format("2006-01-02 15:04")
		}

		fmt.Printf("%-30s %-20s %-15s %-20s\n",
			truncateString(summary.ChangeSetName, 30),
			summary.Status,
			summary.ChangeSetType,
			createdTime)
	}

	fmt.Printf("\n💡 Use 'describe' action to see detailed changes\n")
	return nil
}

func outputChangeSetTable(summary ChangeSetSummary) error {
	fmt.Printf("\n📋 Change Set Details:\n")
	fmt.Printf("══════════════════════════════════════════════════════════════════════════════════\n")
	fmt.Printf("Name: %s\n", summary.ChangeSetName)
	fmt.Printf("Status: %s\n", summary.Status)
	fmt.Printf("Type: %s\n", summary.ChangeSetType)
	fmt.Printf("Stack: %s\n", summary.StackName)

	if summary.CreationTime != nil {
		fmt.Printf("Created: %s\n", summary.CreationTime.Format("2006-01-02 15:04:05"))
	}

	if summary.Description != "" {
		fmt.Printf("Description: %s\n", summary.Description)
	}

	if summary.StatusReason != "" {
		fmt.Printf("Status Reason: %s\n", summary.StatusReason)
	}

	// Show changes
	if len(summary.Changes) > 0 {
		fmt.Printf("\n🔄 Changes (%d):\n", len(summary.Changes))
		fmt.Printf("───────────────────────────────────────────────────────────────────────────────────\n")
		fmt.Printf("%-5s %-15s %-30s %-25s %-15s\n", "#", "ACTION", "RESOURCE", "TYPE", "REPLACEMENT")
		fmt.Printf("───────────────────────────────────────────────────────────────────────────────────\n")

		for i, change := range summary.Changes {
			action := string(change.Type) // Using Type instead of Action
			resource := ""
			resourceType := ""
			replacement := ""

			if change.ResourceChange != nil {
				resource = aws.ToString(change.ResourceChange.LogicalResourceId)
				resourceType = aws.ToString(change.ResourceChange.ResourceType)
				replacement = string(change.ResourceChange.Replacement)
			}

			fmt.Printf("%-5d %-15s %-30s %-25s %-15s\n",
				i+1,
				action,
				truncateString(resource, 30),
				truncateString(resourceType, 25),
				replacement)
		}
	}

	// Show parameters
	if len(summary.Parameters) > 0 {
		fmt.Printf("\n📝 Parameters (%d):\n", len(summary.Parameters))
		for _, param := range summary.Parameters {
			fmt.Printf("  %s: %s\n", aws.ToString(param.ParameterKey), aws.ToString(param.ParameterValue))
		}
	}

	// Show capabilities
	if len(summary.Capabilities) > 0 {
		fmt.Printf("\n🔐 Required Capabilities:\n")
		for _, cap := range summary.Capabilities {
			fmt.Printf("  - %s\n", string(cap))
		}
	}

	return nil
}

func outputImpactTable(analysis ChangeSetImpactAnalysis) error {
	fmt.Printf("\n🔍 Change Set Impact Analysis:\n")
	fmt.Printf("════════════════════════════════════════════════════════════════════════════════\n")
	fmt.Printf("Risk Level: %s\n", getRiskIcon(analysis.RiskLevel)+analysis.RiskLevel)
	fmt.Printf("Resource Replacements: %d\n", analysis.ResourceReplacements)
	fmt.Printf("Service Interruptions: %d\n", analysis.ServiceInterruptions)
	fmt.Printf("Security Changes: %d\n", analysis.SecurityChanges)
	fmt.Printf("Cost Impact: %s\n", analysis.CostImpact)

	if len(analysis.AffectedResources) > 0 {
		fmt.Printf("\n📦 Affected Resources (%d):\n", len(analysis.AffectedResources))
		for _, resource := range analysis.AffectedResources {
			fmt.Printf("  - %s\n", resource)
		}
	}

	if len(analysis.Recommendations) > 0 {
		fmt.Printf("\n💡 Recommendations:\n")
		for _, rec := range analysis.Recommendations {
			fmt.Printf("  - %s\n", rec)
		}
	}

	return nil
}

func outputJSON(data interface{}) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(jsonData))
	return nil
}

func outputYAML(data interface{}) error {
	// For simplicity, using JSON output for now
	// Full YAML implementation would require yaml package
	return outputJSON(data)
}

func getRiskIcon(level string) string {
	switch level {
	case "HIGH":
		return "🔴 "
	case "MEDIUM":
		return "🟡 "
	case "LOW":
		return "🟢 "
	default:
		return "⚪ "
	}
}

func truncateString(str string, maxLen int) string {
	if len(str) <= maxLen {
		return str
	}
	return str[:maxLen-3] + "..."
}
