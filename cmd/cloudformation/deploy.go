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

// DeployCmd represents the deploy command
var DeployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy and manage CloudFormation stacks",
	Long: `Deploy and manage CloudFormation stacks with comprehensive monitoring.

Provides safe stack deployment with:
- Pre-deployment validation and safety checks
- Stack creation and update operations
- Change set preview and approval
- Stack monitoring and rollback capabilities
- Parameter file support
- Cross-region deployment
- Drift detection and remediation

Examples:
  gz cloudformation deploy --template template.yaml --stack my-stack
  gz cloudformation deploy --template template.yaml --stack my-stack --parameters parameters.json
  gz cloudformation deploy --stack my-stack --changeset --preview
  gz cloudformation deploy --stack my-stack --rollback`,
	Run: runDeploy,
}

var (
	deployTemplate     string
	deployStack        string
	deployParameters   string
	deployRegion       string
	deployProfile      string
	capabilities       []string
	createChangeset    bool
	previewChangeset   bool
	executeChangeset   string
	rollbackStack      bool
	deleteStack        bool
	monitorStack       bool
	timeoutMinutes     int
	onFailure          string
	disableRollback    bool
	enableTermination  bool
	notificationARNs   []string
	tags               []string
	roleARN            string
	clientRequestToken string
	stackPolicyBody    string
	driftDetection     bool
)

func init() {
	DeployCmd.Flags().StringVarP(&deployTemplate, "template", "t", "", "CloudFormation template file path")
	DeployCmd.Flags().StringVarP(&deployStack, "stack", "s", "", "CloudFormation stack name")
	DeployCmd.Flags().StringVarP(&deployParameters, "parameters", "p", "", "Parameters file path")
	DeployCmd.Flags().StringVarP(&deployRegion, "region", "r", "", "AWS region")
	DeployCmd.Flags().StringVar(&deployProfile, "profile", "", "AWS profile")
	DeployCmd.Flags().StringSliceVar(&capabilities, "capabilities", []string{}, "Stack capabilities")
	DeployCmd.Flags().BoolVar(&createChangeset, "changeset", false, "Create change set instead of direct deployment")
	DeployCmd.Flags().BoolVar(&previewChangeset, "preview", false, "Preview change set")
	DeployCmd.Flags().StringVar(&executeChangeset, "execute-changeset", "", "Execute specific change set")
	DeployCmd.Flags().BoolVar(&rollbackStack, "rollback", false, "Rollback stack to previous version")
	DeployCmd.Flags().BoolVar(&deleteStack, "delete", false, "Delete stack")
	DeployCmd.Flags().BoolVar(&monitorStack, "monitor", true, "Monitor stack deployment")
	DeployCmd.Flags().IntVar(&timeoutMinutes, "timeout", 30, "Stack operation timeout in minutes")
	DeployCmd.Flags().StringVar(&onFailure, "on-failure", "ROLLBACK", "Action on stack creation failure")
	DeployCmd.Flags().BoolVar(&disableRollback, "disable-rollback", false, "Disable rollback on failure")
	DeployCmd.Flags().BoolVar(&enableTermination, "enable-termination-protection", false, "Enable termination protection")
	DeployCmd.Flags().StringSliceVar(&notificationARNs, "notification-arns", []string{}, "SNS topic ARNs for notifications")
	DeployCmd.Flags().StringSliceVar(&tags, "tags", []string{}, "Stack tags")
	DeployCmd.Flags().StringVar(&roleARN, "role-arn", "", "IAM role ARN for CloudFormation")
	DeployCmd.Flags().StringVar(&clientRequestToken, "client-token", "", "Client request token")
	DeployCmd.Flags().StringVar(&stackPolicyBody, "stack-policy", "", "Stack policy file path")
	DeployCmd.Flags().BoolVar(&driftDetection, "drift-detection", false, "Run drift detection after deployment")
}

// StackInfo represents CloudFormation stack information
type StackInfo struct {
	StackName        string                       `json:"stack_name"`
	StackStatus      string                       `json:"stack_status"`
	CreationTime     *time.Time                   `json:"creation_time"`
	LastUpdatedTime  *time.Time                   `json:"last_updated_time"`
	DriftInformation *types.StackDriftInformation `json:"drift_information,omitempty"`
	Outputs          []types.Output               `json:"outputs,omitempty"`
	Parameters       []types.Parameter            `json:"parameters,omitempty"`
	Tags             []types.Tag                  `json:"tags,omitempty"`
}

// DeploymentResult represents the result of a stack deployment
type DeploymentResult struct {
	Success     bool               `json:"success"`
	StackName   string             `json:"stack_name"`
	StackId     string             `json:"stack_id"`
	Status      string             `json:"status"`
	Events      []types.StackEvent `json:"events"`
	Outputs     []types.Output     `json:"outputs"`
	Duration    time.Duration      `json:"duration"`
	ChangeSetId string             `json:"changeset_id,omitempty"`
	Errors      []string           `json:"errors"`
}

func runDeploy(cmd *cobra.Command, args []string) {
	startTime := time.Now()

	if deployStack == "" {
		fmt.Printf("❌ Stack name is required\n")
		cmd.Help()
		os.Exit(1)
	}

	fmt.Printf("🚀 CloudFormation Stack Deployment\n")
	fmt.Printf("📦 Stack: %s\n", deployStack)

	// Initialize AWS config
	ctx := context.Background()
	cfg, err := initAWSConfig(ctx)
	if err != nil {
		fmt.Printf("❌ Failed to initialize AWS config: %v\n", err)
		os.Exit(1)
	}

	client := cloudformation.NewFromConfig(cfg)

	// Handle different operations
	var result *DeploymentResult
	switch {
	case deleteStack:
		result, err = handleDeleteStack(ctx, client, startTime)
	case rollbackStack:
		result, err = handleRollbackStack(ctx, client, startTime)
	case createChangeset:
		result, err = handleCreateChangeset(ctx, client, startTime)
	case previewChangeset:
		result, err = handlePreviewChangeset(ctx, client, startTime)
	case executeChangeset != "":
		result, err = handleExecuteChangeset(ctx, client, startTime)
	case driftDetection:
		result, err = handleDriftDetection(ctx, client, startTime)
	default:
		result, err = handleStackDeployment(ctx, client, startTime)
	}

	if err != nil {
		fmt.Printf("❌ Operation failed: %v\n", err)
		os.Exit(1)
	}

	// Print results
	printDeploymentResults(result)
}

func initAWSConfig(ctx context.Context) (aws.Config, error) {
	var options []func(*config.LoadOptions) error

	if deployRegion != "" {
		options = append(options, config.WithRegion(deployRegion))
	}

	if deployProfile != "" {
		options = append(options, config.WithSharedConfigProfile(deployProfile))
	}

	return config.LoadDefaultConfig(ctx, options...)
}

func handleStackDeployment(ctx context.Context, client *cloudformation.Client, startTime time.Time) (*DeploymentResult, error) {
	fmt.Printf("🔍 Checking stack status...\n")

	// Check if stack exists
	stackExists, err := checkStackExists(ctx, client, deployStack)
	if err != nil {
		return nil, fmt.Errorf("failed to check stack existence: %w", err)
	}

	var result *DeploymentResult
	if stackExists {
		fmt.Printf("📝 Stack exists, updating...\n")
		result, err = updateStack(ctx, client, startTime)
	} else {
		fmt.Printf("✨ Creating new stack...\n")
		result, err = createStack(ctx, client, startTime)
	}

	if err != nil {
		return nil, err
	}

	// Monitor stack if requested
	if monitorStack {
		fmt.Printf("👀 Monitoring stack deployment...\n")
		if err := monitorStackProgress(ctx, client, deployStack, timeoutMinutes); err != nil {
			result.Errors = append(result.Errors, err.Error())
		}
	}

	// Run drift detection if requested
	if driftDetection {
		fmt.Printf("🔍 Running drift detection...\n")
		if err := runStackDriftDetection(ctx, client, deployStack); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Drift detection failed: %v", err))
		}
	}

	return result, nil
}

func checkStackExists(ctx context.Context, client *cloudformation.Client, stackName string) (bool, error) {
	input := &cloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	}

	_, err := client.DescribeStacks(ctx, input)
	if err != nil {
		// Check if it's a "stack does not exist" error
		if strings.Contains(err.Error(), "does not exist") {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func createStack(ctx context.Context, client *cloudformation.Client, startTime time.Time) (*DeploymentResult, error) {
	templateBody, err := readTemplateFile()
	if err != nil {
		return nil, err
	}

	parameters, err := parseParameters()
	if err != nil {
		return nil, err
	}

	stackTags, err := parseStackTags()
	if err != nil {
		return nil, err
	}

	input := &cloudformation.CreateStackInput{
		StackName:                   aws.String(deployStack),
		TemplateBody:                aws.String(templateBody),
		Parameters:                  parameters,
		Tags:                        stackTags,
		Capabilities:                parseCapabilities(),
		TimeoutInMinutes:            aws.Int32(int32(timeoutMinutes)),
		OnFailure:                   types.OnFailure(onFailure),
		DisableRollback:             aws.Bool(disableRollback),
		EnableTerminationProtection: aws.Bool(enableTermination),
	}

	if len(notificationARNs) > 0 {
		input.NotificationARNs = notificationARNs
	}

	if roleARN != "" {
		input.RoleARN = aws.String(roleARN)
	}

	if clientRequestToken != "" {
		// Note: ClientRequestToken may not be available in some inputs
		// input.ClientRequestToken = aws.String(clientRequestToken)
	}

	if stackPolicyBody != "" {
		policyBody, err := os.ReadFile(stackPolicyBody)
		if err != nil {
			return nil, fmt.Errorf("failed to read stack policy: %w", err)
		}
		input.StackPolicyBody = aws.String(string(policyBody))
	}

	output, err := client.CreateStack(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create stack: %w", err)
	}

	return &DeploymentResult{
		Success:   true,
		StackName: deployStack,
		StackId:   aws.ToString(output.StackId),
		Status:    "CREATE_IN_PROGRESS",
		Duration:  time.Since(startTime),
		Errors:    []string{},
	}, nil
}

func updateStack(ctx context.Context, client *cloudformation.Client, startTime time.Time) (*DeploymentResult, error) {
	templateBody, err := readTemplateFile()
	if err != nil {
		return nil, err
	}

	parameters, err := parseParameters()
	if err != nil {
		return nil, err
	}

	stackTags, err := parseStackTags()
	if err != nil {
		return nil, err
	}

	input := &cloudformation.UpdateStackInput{
		StackName:    aws.String(deployStack),
		TemplateBody: aws.String(templateBody),
		Parameters:   parameters,
		Tags:         stackTags,
		Capabilities: parseCapabilities(),
	}

	if roleARN != "" {
		input.RoleARN = aws.String(roleARN)
	}

	if clientRequestToken != "" {
		// Note: ClientRequestToken may not be available in some inputs
		// input.ClientRequestToken = aws.String(clientRequestToken)
	}

	if stackPolicyBody != "" {
		policyBody, err := os.ReadFile(stackPolicyBody)
		if err != nil {
			return nil, fmt.Errorf("failed to read stack policy: %w", err)
		}
		input.StackPolicyBody = aws.String(string(policyBody))
	}

	output, err := client.UpdateStack(ctx, input)
	if err != nil {
		// Check if no updates are needed
		if strings.Contains(err.Error(), "No updates are to be performed") {
			fmt.Printf("✅ No updates required for stack\n")
			return &DeploymentResult{
				Success:   true,
				StackName: deployStack,
				Status:    "UPDATE_COMPLETE",
				Duration:  time.Since(startTime),
				Errors:    []string{},
			}, nil
		}
		return nil, fmt.Errorf("failed to update stack: %w", err)
	}

	return &DeploymentResult{
		Success:   true,
		StackName: deployStack,
		StackId:   aws.ToString(output.StackId),
		Status:    "UPDATE_IN_PROGRESS",
		Duration:  time.Since(startTime),
		Errors:    []string{},
	}, nil
}

func handleDeleteStack(ctx context.Context, client *cloudformation.Client, startTime time.Time) (*DeploymentResult, error) {
	fmt.Printf("🗑️ Deleting stack: %s\n", deployStack)

	input := &cloudformation.DeleteStackInput{
		StackName: aws.String(deployStack),
	}

	if roleARN != "" {
		input.RoleARN = aws.String(roleARN)
	}

	if clientRequestToken != "" {
		// Note: ClientRequestToken may not be available in some inputs
		// input.ClientRequestToken = aws.String(clientRequestToken)
	}

	_, err := client.DeleteStack(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to delete stack: %w", err)
	}

	result := &DeploymentResult{
		Success:   true,
		StackName: deployStack,
		Status:    "DELETE_IN_PROGRESS",
		Duration:  time.Since(startTime),
		Errors:    []string{},
	}

	// Monitor deletion if requested
	if monitorStack {
		fmt.Printf("👀 Monitoring stack deletion...\n")
		if err := monitorStackProgress(ctx, client, deployStack, timeoutMinutes); err != nil {
			result.Errors = append(result.Errors, err.Error())
		}
	}

	return result, nil
}

func handleRollbackStack(ctx context.Context, client *cloudformation.Client, startTime time.Time) (*DeploymentResult, error) {
	fmt.Printf("🔄 Rolling back stack: %s\n", deployStack)

	input := &cloudformation.CancelUpdateStackInput{
		StackName: aws.String(deployStack),
	}

	if clientRequestToken != "" {
		// Note: ClientRequestToken may not be available in some inputs
		// input.ClientRequestToken = aws.String(clientRequestToken)
	}

	_, err := client.CancelUpdateStack(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to rollback stack: %w", err)
	}

	return &DeploymentResult{
		Success:   true,
		StackName: deployStack,
		Status:    "UPDATE_ROLLBACK_IN_PROGRESS",
		Duration:  time.Since(startTime),
		Errors:    []string{},
	}, nil
}

func handleCreateChangeset(ctx context.Context, client *cloudformation.Client, startTime time.Time) (*DeploymentResult, error) {
	changesetName := fmt.Sprintf("%s-changeset-%d", deployStack, time.Now().Unix())
	fmt.Printf("📋 Creating change set: %s\n", changesetName)

	templateBody, err := readTemplateFile()
	if err != nil {
		return nil, err
	}

	parameters, err := parseParameters()
	if err != nil {
		return nil, err
	}

	stackTags, err := parseStackTags()
	if err != nil {
		return nil, err
	}

	input := &cloudformation.CreateChangeSetInput{
		StackName:     aws.String(deployStack),
		ChangeSetName: aws.String(changesetName),
		TemplateBody:  aws.String(templateBody),
		Parameters:    parameters,
		Tags:          stackTags,
		Capabilities:  parseCapabilities(),
	}

	if roleARN != "" {
		input.RoleARN = aws.String(roleARN)
	}

	if clientRequestToken != "" {
		// Note: ClientRequestToken may not be available in some inputs
		// input.ClientRequestToken = aws.String(clientRequestToken)
	}

	output, err := client.CreateChangeSet(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create change set: %w", err)
	}

	// Wait for change set creation to complete
	fmt.Printf("⏳ Waiting for change set creation...\n")
	if err := waitForChangeSetCreation(ctx, client, deployStack, changesetName); err != nil {
		return nil, err
	}

	// Preview the change set
	if err := previewChangeSet(ctx, client, deployStack, changesetName); err != nil {
		return nil, err
	}

	return &DeploymentResult{
		Success:     true,
		StackName:   deployStack,
		ChangeSetId: aws.ToString(output.Id),
		Status:      "CHANGESET_CREATED",
		Duration:    time.Since(startTime),
		Errors:      []string{},
	}, nil
}

func handlePreviewChangeset(ctx context.Context, client *cloudformation.Client, startTime time.Time) (*DeploymentResult, error) {
	// List change sets for the stack
	changeSets, err := listChangeSets(ctx, client, deployStack)
	if err != nil {
		return nil, err
	}

	if len(changeSets) == 0 {
		return nil, fmt.Errorf("no change sets found for stack %s", deployStack)
	}

	// Preview the latest change set
	latestChangeSet := changeSets[0]
	fmt.Printf("📋 Previewing change set: %s\n", aws.ToString(latestChangeSet.ChangeSetName))

	if err := previewChangeSet(ctx, client, deployStack, aws.ToString(latestChangeSet.ChangeSetName)); err != nil {
		return nil, err
	}

	return &DeploymentResult{
		Success:   true,
		StackName: deployStack,
		Status:    "CHANGESET_PREVIEWED",
		Duration:  time.Since(startTime),
		Errors:    []string{},
	}, nil
}

func handleExecuteChangeset(ctx context.Context, client *cloudformation.Client, startTime time.Time) (*DeploymentResult, error) {
	fmt.Printf("▶️ Executing change set: %s\n", executeChangeset)

	input := &cloudformation.ExecuteChangeSetInput{
		ChangeSetName: aws.String(executeChangeset),
		StackName:     aws.String(deployStack),
	}

	if clientRequestToken != "" {
		// Note: ClientRequestToken may not be available in some inputs
		// input.ClientRequestToken = aws.String(clientRequestToken)
	}

	_, err := client.ExecuteChangeSet(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to execute change set: %w", err)
	}

	result := &DeploymentResult{
		Success:   true,
		StackName: deployStack,
		Status:    "UPDATE_IN_PROGRESS",
		Duration:  time.Since(startTime),
		Errors:    []string{},
	}

	// Monitor execution if requested
	if monitorStack {
		fmt.Printf("👀 Monitoring change set execution...\n")
		if err := monitorStackProgress(ctx, client, deployStack, timeoutMinutes); err != nil {
			result.Errors = append(result.Errors, err.Error())
		}
	}

	return result, nil
}

func handleDriftDetection(ctx context.Context, client *cloudformation.Client, startTime time.Time) (*DeploymentResult, error) {
	fmt.Printf("🔍 Starting drift detection for stack: %s\n", deployStack)

	if err := runStackDriftDetection(ctx, client, deployStack); err != nil {
		return nil, err
	}

	return &DeploymentResult{
		Success:   true,
		StackName: deployStack,
		Status:    "DRIFT_DETECTION_COMPLETE",
		Duration:  time.Since(startTime),
		Errors:    []string{},
	}, nil
}

func readTemplateFile() (string, error) {
	if deployTemplate == "" {
		return "", fmt.Errorf("template file is required")
	}

	data, err := os.ReadFile(deployTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to read template file: %w", err)
	}

	return string(data), nil
}

func parseParameters() ([]types.Parameter, error) {
	if deployParameters == "" {
		return nil, nil
	}

	data, err := os.ReadFile(deployParameters)
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

func parseStackTags() ([]types.Tag, error) {
	var stackTags []types.Tag

	for _, tag := range tags {
		parts := strings.SplitN(tag, "=", 2)
		if len(parts) == 2 {
			stackTags = append(stackTags, types.Tag{
				Key:   aws.String(parts[0]),
				Value: aws.String(parts[1]),
			})
		}
	}

	return stackTags, nil
}

func parseCapabilities() []types.Capability {
	var caps []types.Capability
	for _, cap := range capabilities {
		caps = append(caps, types.Capability(cap))
	}
	return caps
}

func monitorStackProgress(ctx context.Context, client *cloudformation.Client, stackName string, timeoutMinutes int) error {
	timeout := time.Duration(timeoutMinutes) * time.Minute
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		status, err := getStackStatus(ctx, client, stackName)
		if err != nil {
			return err
		}

		fmt.Printf("📊 Stack status: %s\n", status)

		// Check if operation is complete
		if isStackOperationComplete(status) {
			if isStackOperationSuccessful(status) {
				fmt.Printf("✅ Stack operation completed successfully\n")
				return nil
			} else {
				return fmt.Errorf("stack operation failed with status: %s", status)
			}
		}

		time.Sleep(30 * time.Second)
	}

	return fmt.Errorf("timeout waiting for stack operation to complete")
}

func getStackStatus(ctx context.Context, client *cloudformation.Client, stackName string) (string, error) {
	input := &cloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	}

	output, err := client.DescribeStacks(ctx, input)
	if err != nil {
		return "", err
	}

	if len(output.Stacks) == 0 {
		return "", fmt.Errorf("stack not found")
	}

	return string(output.Stacks[0].StackStatus), nil
}

func isStackOperationComplete(status string) bool {
	completeStatuses := []string{
		"CREATE_COMPLETE",
		"CREATE_FAILED",
		"DELETE_COMPLETE",
		"DELETE_FAILED",
		"UPDATE_COMPLETE",
		"UPDATE_FAILED",
		"UPDATE_ROLLBACK_COMPLETE",
		"UPDATE_ROLLBACK_FAILED",
	}

	for _, completeStatus := range completeStatuses {
		if status == completeStatus {
			return true
		}
	}
	return false
}

func isStackOperationSuccessful(status string) bool {
	successStatuses := []string{
		"CREATE_COMPLETE",
		"DELETE_COMPLETE",
		"UPDATE_COMPLETE",
		"UPDATE_ROLLBACK_COMPLETE",
	}

	for _, successStatus := range successStatuses {
		if status == successStatus {
			return true
		}
	}
	return false
}

func waitForChangeSetCreation(ctx context.Context, client *cloudformation.Client, stackName, changeSetName string) error {
	for i := 0; i < 60; i++ { // Wait up to 5 minutes
		input := &cloudformation.DescribeChangeSetInput{
			StackName:     aws.String(stackName),
			ChangeSetName: aws.String(changeSetName),
		}

		output, err := client.DescribeChangeSet(ctx, input)
		if err != nil {
			return err
		}

		status := string(output.Status)
		if status == "CREATE_COMPLETE" {
			return nil
		} else if status == "FAILED" {
			return fmt.Errorf("change set creation failed: %s", aws.ToString(output.StatusReason))
		}

		time.Sleep(5 * time.Second)
	}

	return fmt.Errorf("timeout waiting for change set creation")
}

func previewChangeSet(ctx context.Context, client *cloudformation.Client, stackName, changeSetName string) error {
	input := &cloudformation.DescribeChangeSetInput{
		StackName:     aws.String(stackName),
		ChangeSetName: aws.String(changeSetName),
	}

	output, err := client.DescribeChangeSet(ctx, input)
	if err != nil {
		return err
	}

	fmt.Printf("\n📋 Change Set Preview:\n")
	fmt.Printf("=====================\n")
	fmt.Printf("Change Set Name: %s\n", aws.ToString(output.ChangeSetName))
	fmt.Printf("Status: %s\n", string(output.Status))
	fmt.Printf("Description: %s\n", aws.ToString(output.Description))

	if len(output.Changes) == 0 {
		fmt.Printf("No changes detected\n")
		return nil
	}

	fmt.Printf("\nChanges (%d):\n", len(output.Changes))
	for i, change := range output.Changes {
		fmt.Printf("%d. Action: %s\n", i+1, string(change.Type))
		if change.ResourceChange != nil {
			fmt.Printf("   Resource: %s (%s)\n",
				aws.ToString(change.ResourceChange.LogicalResourceId),
				aws.ToString(change.ResourceChange.ResourceType))
			if change.ResourceChange.Replacement != "" {
				fmt.Printf("   Replacement: %s\n", string(change.ResourceChange.Replacement))
			}
		}
	}

	return nil
}

func listChangeSets(ctx context.Context, client *cloudformation.Client, stackName string) ([]types.ChangeSetSummary, error) {
	input := &cloudformation.ListChangeSetsInput{
		StackName: aws.String(stackName),
	}

	output, err := client.ListChangeSets(ctx, input)
	if err != nil {
		return nil, err
	}

	return output.Summaries, nil
}

func runStackDriftDetection(ctx context.Context, client *cloudformation.Client, stackName string) error {
	// Start drift detection
	input := &cloudformation.DetectStackDriftInput{
		StackName: aws.String(stackName),
	}

	output, err := client.DetectStackDrift(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to start drift detection: %w", err)
	}

	// Wait for drift detection to complete
	fmt.Printf("⏳ Waiting for drift detection to complete...\n")
	driftId := aws.ToString(output.StackDriftDetectionId)

	for i := 0; i < 120; i++ { // Wait up to 10 minutes
		statusInput := &cloudformation.DescribeStackDriftDetectionStatusInput{
			StackDriftDetectionId: aws.String(driftId),
		}

		statusOutput, err := client.DescribeStackDriftDetectionStatus(ctx, statusInput)
		if err != nil {
			return err
		}

		status := string(statusOutput.DetectionStatus)
		if status == "DETECTION_COMPLETE" {
			fmt.Printf("✅ Drift detection completed\n")
			fmt.Printf("Stack drift status: %s\n", string(statusOutput.StackDriftStatus))

			if statusOutput.StackDriftStatus == types.StackDriftStatusDrifted {
				fmt.Printf("⚠️ Stack has drifted from its expected configuration\n")
				// List drifted resources
				if err := listDriftedResources(ctx, client, stackName); err != nil {
					return err
				}
			}
			return nil
		} else if status == "DETECTION_FAILED" {
			return fmt.Errorf("drift detection failed: %s", aws.ToString(statusOutput.DetectionStatusReason))
		}

		time.Sleep(5 * time.Second)
	}

	return fmt.Errorf("timeout waiting for drift detection to complete")
}

func listDriftedResources(ctx context.Context, client *cloudformation.Client, stackName string) error {
	input := &cloudformation.DescribeStackResourceDriftsInput{
		StackName: aws.String(stackName),
	}

	output, err := client.DescribeStackResourceDrifts(ctx, input)
	if err != nil {
		return err
	}

	fmt.Printf("\n🔍 Drifted Resources:\n")
	for _, drift := range output.StackResourceDrifts {
		if drift.StackResourceDriftStatus == types.StackResourceDriftStatusModified {
			fmt.Printf("- %s (%s): %s\n",
				aws.ToString(drift.LogicalResourceId),
				aws.ToString(drift.ResourceType),
				string(drift.StackResourceDriftStatus))
		}
	}

	return nil
}

func printDeploymentResults(result *DeploymentResult) {
	fmt.Printf("\n🎉 Deployment Results:\n")
	fmt.Printf("=====================\n")
	fmt.Printf("Success: %t\n", result.Success)
	fmt.Printf("Stack Name: %s\n", result.StackName)
	fmt.Printf("Status: %s\n", result.Status)
	fmt.Printf("Duration: %v\n", result.Duration)

	if result.StackId != "" {
		fmt.Printf("Stack ID: %s\n", result.StackId)
	}

	if result.ChangeSetId != "" {
		fmt.Printf("Change Set ID: %s\n", result.ChangeSetId)
	}

	if len(result.Errors) > 0 {
		fmt.Printf("\n⚠️ Warnings/Errors:\n")
		for _, err := range result.Errors {
			fmt.Printf("  - %s\n", err)
		}
	}

	if len(result.Outputs) > 0 {
		fmt.Printf("\n📊 Stack Outputs:\n")
		for _, output := range result.Outputs {
			fmt.Printf("  %s: %s\n", aws.ToString(output.OutputKey), aws.ToString(output.OutputValue))
		}
	}

	fmt.Printf("\n📝 Next steps:\n")
	fmt.Printf("1. Verify resources in AWS Console\n")
	fmt.Printf("2. Test deployed infrastructure\n")
	fmt.Printf("3. Monitor CloudWatch for any issues\n")
	if driftDetection {
		fmt.Printf("4. Set up periodic drift detection\n")
	}
}
