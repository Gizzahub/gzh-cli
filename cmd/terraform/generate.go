package terraform

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// GenerateCmd represents the generate command
var GenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate Terraform configurations and modules",
	Long: `Generate Terraform configurations and modules for various cloud providers.

Supports generation of:
- Provider-specific modules (AWS, GCP, Azure)
- Network infrastructure (VPC, subnets, security groups)
- Compute resources (EC2, GKE, AKS clusters)
- Storage solutions (S3, Cloud Storage, Blob Storage)
- Database configurations (RDS, Cloud SQL, Cosmos DB)
- Kubernetes deployment manifests
- Terraform backends and state management

Examples:
  gz terraform generate --provider aws --resource vpc
  gz terraform generate --provider gcp --resource gke-cluster
  gz terraform generate --module networking --environment production`,
	Run: runGenerate,
}

var (
	provider     string
	resource     string
	module       string
	environment  string
	region       string
	outputPath   string
	backend      string
	withModules  bool
	withRemote   bool
	withWorkflow bool
	variables    []string
	tags         []string
)

func init() {
	GenerateCmd.Flags().StringVarP(&provider, "provider", "p", "aws", "Cloud provider (aws, gcp, azure)")
	GenerateCmd.Flags().StringVarP(&resource, "resource", "r", "", "Resource type to generate")
	GenerateCmd.Flags().StringVarP(&module, "module", "m", "", "Module type to generate")
	GenerateCmd.Flags().StringVarP(&environment, "environment", "e", "dev", "Target environment")
	GenerateCmd.Flags().StringVar(&region, "region", "", "Cloud region")
	GenerateCmd.Flags().StringVarP(&outputPath, "output", "o", ".", "Output directory")
	GenerateCmd.Flags().StringVar(&backend, "backend", "local", "Backend type (local, s3, gcs)")
	GenerateCmd.Flags().BoolVar(&withModules, "with-modules", true, "Include module structure")
	GenerateCmd.Flags().BoolVar(&withRemote, "with-remote-state", false, "Include remote state configuration")
	GenerateCmd.Flags().BoolVar(&withWorkflow, "with-workflow", false, "Include CI/CD workflow files")
	GenerateCmd.Flags().StringSliceVar(&variables, "vars", []string{}, "Variables to include")
	GenerateCmd.Flags().StringSliceVar(&tags, "tags", []string{}, "Default tags to apply")
}

// TerraformSpec holds terraform generation specifications
type TerraformSpec struct {
	Provider     string
	Resource     string
	Module       string
	Environment  string
	Region       string
	Backend      string
	WithModules  bool
	WithRemote   bool
	WithWorkflow bool
	Variables    map[string]interface{}
	Tags         map[string]string
}

// ModuleConfig represents a Terraform module configuration
type ModuleConfig struct {
	Name         string                 `yaml:"name"`
	Description  string                 `yaml:"description"`
	Version      string                 `yaml:"version"`
	Provider     string                 `yaml:"provider"`
	Resources    []ResourceConfig       `yaml:"resources"`
	Variables    map[string]interface{} `yaml:"variables"`
	Outputs      map[string]interface{} `yaml:"outputs"`
	Dependencies []string               `yaml:"dependencies"`
}

type ResourceConfig struct {
	Type        string                 `yaml:"type"`
	Name        string                 `yaml:"name"`
	Properties  map[string]interface{} `yaml:"properties"`
	DependsOn   []string               `yaml:"depends_on,omitempty"`
	Count       interface{}            `yaml:"count,omitempty"`
	ForEach     interface{}            `yaml:"for_each,omitempty"`
	Lifecycle   map[string]interface{} `yaml:"lifecycle,omitempty"`
	Provisioner map[string]interface{} `yaml:"provisioner,omitempty"`
}

func runGenerate(cmd *cobra.Command, args []string) {
	// Parse variables
	varsMap := make(map[string]interface{})
	for _, v := range variables {
		parts := strings.SplitN(v, "=", 2)
		if len(parts) == 2 {
			varsMap[parts[0]] = parts[1]
		}
	}

	// Parse tags
	tagsMap := make(map[string]string)
	for _, t := range tags {
		parts := strings.SplitN(t, "=", 2)
		if len(parts) == 2 {
			tagsMap[parts[0]] = parts[1]
		}
	}

	spec := TerraformSpec{
		Provider:     provider,
		Resource:     resource,
		Module:       module,
		Environment:  environment,
		Region:       region,
		Backend:      backend,
		WithModules:  withModules,
		WithRemote:   withRemote,
		WithWorkflow: withWorkflow,
		Variables:    varsMap,
		Tags:         tagsMap,
	}

	fmt.Printf("🏗️ Generating Terraform configuration for %s\n", provider)
	if environment != "" {
		fmt.Printf("📦 Environment: %s\n", environment)
	}

	// Create output directory
	if err := os.MkdirAll(outputPath, 0o755); err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	// Generate main configuration
	if err := generateMainConfig(spec); err != nil {
		fmt.Printf("Error generating main configuration: %v\n", err)
		os.Exit(1)
	}

	// Generate provider configuration
	if err := generateProviderConfig(spec); err != nil {
		fmt.Printf("Error generating provider configuration: %v\n", err)
		os.Exit(1)
	}

	// Generate variables
	if err := generateVariables(spec); err != nil {
		fmt.Printf("Error generating variables: %v\n", err)
		os.Exit(1)
	}

	// Generate outputs
	if err := generateOutputs(spec); err != nil {
		fmt.Printf("Error generating outputs: %v\n", err)
		os.Exit(1)
	}

	// Generate backend configuration if remote
	if spec.WithRemote {
		if err := generateBackend(spec); err != nil {
			fmt.Printf("Error generating backend: %v\n", err)
			os.Exit(1)
		}
	}

	// Generate modules if requested
	if spec.WithModules {
		if err := generateModules(spec); err != nil {
			fmt.Printf("Error generating modules: %v\n", err)
			os.Exit(1)
		}
	}

	// Generate workflow files if requested
	if spec.WithWorkflow {
		if err := generateWorkflow(spec); err != nil {
			fmt.Printf("Error generating workflow: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Printf("✅ Terraform configuration generated successfully\n")
	fmt.Printf("📁 Output directory: %s\n", outputPath)
	fmt.Printf("\n📝 Next steps:\n")
	fmt.Printf("1. Review generated configuration files\n")
	fmt.Printf("2. Initialize Terraform: terraform init\n")
	fmt.Printf("3. Plan deployment: terraform plan\n")
	fmt.Printf("4. Apply configuration: terraform apply\n")
}

func generateMainConfig(spec TerraformSpec) error {
	var content strings.Builder

	// Terraform version and required providers
	content.WriteString("terraform {\n")
	content.WriteString("  required_version = \">= 1.0\"\n")
	content.WriteString("  required_providers {\n")

	switch spec.Provider {
	case "aws":
		content.WriteString("    aws = {\n")
		content.WriteString("      source  = \"hashicorp/aws\"\n")
		content.WriteString("      version = \"~> 5.0\"\n")
		content.WriteString("    }\n")
	case "gcp":
		content.WriteString("    google = {\n")
		content.WriteString("      source  = \"hashicorp/google\"\n")
		content.WriteString("      version = \"~> 4.0\"\n")
		content.WriteString("    }\n")
	case "azure":
		content.WriteString("    azurerm = {\n")
		content.WriteString("      source  = \"hashicorp/azurerm\"\n")
		content.WriteString("      version = \"~> 3.0\"\n")
		content.WriteString("    }\n")
	}

	content.WriteString("  }\n")

	// Backend configuration
	if spec.WithRemote {
		content.WriteString("\n  backend \"" + spec.Backend + "\" {\n")
		switch spec.Backend {
		case "s3":
			content.WriteString("    bucket = \"terraform-state-" + spec.Environment + "\"\n")
			content.WriteString("    key    = \"" + spec.Environment + "/terraform.tfstate\"\n")
			content.WriteString("    region = \"" + getDefaultRegion(spec.Provider) + "\"\n")
		case "gcs":
			content.WriteString("    bucket = \"terraform-state-" + spec.Environment + "\"\n")
			content.WriteString("    prefix = \"" + spec.Environment + "\"\n")
		}
		content.WriteString("  }\n")
	}

	content.WriteString("}\n\n")

	// Local values
	content.WriteString("locals {\n")
	content.WriteString("  environment = \"" + spec.Environment + "\"\n")
	content.WriteString("  region      = var.region\n")
	content.WriteString("  \n")
	content.WriteString("  common_tags = {\n")
	content.WriteString("    Environment = local.environment\n")
	content.WriteString("    ManagedBy   = \"terraform\"\n")
	content.WriteString("    Project     = var.project_name\n")
	for k, v := range spec.Tags {
		content.WriteString(fmt.Sprintf("    %s = \"%s\"\n", k, v))
	}
	content.WriteString("  }\n")
	content.WriteString("}\n\n")

	// Generate resource blocks based on specification
	if spec.Resource != "" {
		content.WriteString(generateResourceBlocks(spec))
	} else if spec.Module != "" {
		content.WriteString(generateModuleBlocks(spec))
	}

	return os.WriteFile(filepath.Join(outputPath, "main.tf"), []byte(content.String()), 0o644)
}

func generateProviderConfig(spec TerraformSpec) error {
	var content strings.Builder

	switch spec.Provider {
	case "aws":
		content.WriteString("provider \"aws\" {\n")
		if spec.Region != "" {
			content.WriteString("  region = \"" + spec.Region + "\"\n")
		} else {
			content.WriteString("  region = var.aws_region\n")
		}
		content.WriteString("\n")
		content.WriteString("  default_tags {\n")
		content.WriteString("    tags = local.common_tags\n")
		content.WriteString("  }\n")
		content.WriteString("}\n")

	case "gcp":
		content.WriteString("provider \"google\" {\n")
		content.WriteString("  project = var.gcp_project\n")
		if spec.Region != "" {
			content.WriteString("  region  = \"" + spec.Region + "\"\n")
		} else {
			content.WriteString("  region  = var.gcp_region\n")
		}
		content.WriteString("}\n")

	case "azure":
		content.WriteString("provider \"azurerm\" {\n")
		content.WriteString("  features {}\n")
		content.WriteString("}\n")
	}

	return os.WriteFile(filepath.Join(outputPath, "providers.tf"), []byte(content.String()), 0o644)
}

func generateVariables(spec TerraformSpec) error {
	var content strings.Builder

	// Common variables
	content.WriteString("variable \"project_name\" {\n")
	content.WriteString("  description = \"Name of the project\"\n")
	content.WriteString("  type        = string\n")
	content.WriteString("}\n\n")

	content.WriteString("variable \"environment\" {\n")
	content.WriteString("  description = \"Environment name\"\n")
	content.WriteString("  type        = string\n")
	content.WriteString("  default     = \"" + spec.Environment + "\"\n")
	content.WriteString("}\n\n")

	// Provider-specific variables
	switch spec.Provider {
	case "aws":
		content.WriteString("variable \"aws_region\" {\n")
		content.WriteString("  description = \"AWS region\"\n")
		content.WriteString("  type        = string\n")
		content.WriteString("  default     = \"" + getDefaultRegion(spec.Provider) + "\"\n")
		content.WriteString("}\n\n")

		if spec.Resource == "vpc" || spec.Module == "networking" {
			content.WriteString("variable \"vpc_cidr\" {\n")
			content.WriteString("  description = \"CIDR block for VPC\"\n")
			content.WriteString("  type        = string\n")
			content.WriteString("  default     = \"10.0.0.0/16\"\n")
			content.WriteString("}\n\n")
		}

	case "gcp":
		content.WriteString("variable \"gcp_project\" {\n")
		content.WriteString("  description = \"GCP project ID\"\n")
		content.WriteString("  type        = string\n")
		content.WriteString("}\n\n")

		content.WriteString("variable \"gcp_region\" {\n")
		content.WriteString("  description = \"GCP region\"\n")
		content.WriteString("  type        = string\n")
		content.WriteString("  default     = \"" + getDefaultRegion(spec.Provider) + "\"\n")
		content.WriteString("}\n\n")

	case "azure":
		content.WriteString("variable \"azure_location\" {\n")
		content.WriteString("  description = \"Azure location\"\n")
		content.WriteString("  type        = string\n")
		content.WriteString("  default     = \"" + getDefaultRegion(spec.Provider) + "\"\n")
		content.WriteString("}\n\n")
	}

	// Add custom variables
	for k, v := range spec.Variables {
		content.WriteString(fmt.Sprintf("variable \"%s\" {\n", k))
		content.WriteString("  description = \"Custom variable\"\n")
		content.WriteString("  type        = string\n")
		content.WriteString(fmt.Sprintf("  default     = \"%v\"\n", v))
		content.WriteString("}\n\n")
	}

	return os.WriteFile(filepath.Join(outputPath, "variables.tf"), []byte(content.String()), 0o644)
}

func generateOutputs(spec TerraformSpec) error {
	var content strings.Builder

	switch spec.Provider {
	case "aws":
		if spec.Resource == "vpc" || spec.Module == "networking" {
			content.WriteString("output \"vpc_id\" {\n")
			content.WriteString("  description = \"ID of the VPC\"\n")
			content.WriteString("  value       = aws_vpc.main.id\n")
			content.WriteString("}\n\n")

			content.WriteString("output \"vpc_cidr_block\" {\n")
			content.WriteString("  description = \"CIDR block of the VPC\"\n")
			content.WriteString("  value       = aws_vpc.main.cidr_block\n")
			content.WriteString("}\n\n")
		}

	case "gcp":
		if spec.Resource == "vpc" || spec.Module == "networking" {
			content.WriteString("output \"network_name\" {\n")
			content.WriteString("  description = \"Name of the VPC network\"\n")
			content.WriteString("  value       = google_compute_network.main.name\n")
			content.WriteString("}\n\n")
		}

	case "azure":
		if spec.Resource == "resource-group" || spec.Module == "networking" {
			content.WriteString("output \"resource_group_name\" {\n")
			content.WriteString("  description = \"Name of the resource group\"\n")
			content.WriteString("  value       = azurerm_resource_group.main.name\n")
			content.WriteString("}\n\n")
		}
	}

	return os.WriteFile(filepath.Join(outputPath, "outputs.tf"), []byte(content.String()), 0o644)
}

func generateBackend(spec TerraformSpec) error {
	var content strings.Builder

	content.WriteString("# Backend configuration for remote state\n")
	content.WriteString("# This file should be customized for your specific backend setup\n\n")

	switch spec.Backend {
	case "s3":
		content.WriteString("# To use this backend, run:\n")
		content.WriteString("# terraform init -backend-config=\"bucket=your-terraform-state-bucket\"\n")
		content.WriteString("# terraform init -backend-config=\"key=" + spec.Environment + "/terraform.tfstate\"\n")
		content.WriteString("# terraform init -backend-config=\"region=your-region\"\n\n")

		// Generate S3 bucket creation script
		bucketContent := `#!/bin/bash
# Script to create S3 bucket for Terraform state

BUCKET_NAME="terraform-state-` + spec.Environment + `"
REGION="` + getDefaultRegion(spec.Provider) + `"

aws s3 mb s3://$BUCKET_NAME --region $REGION
aws s3api put-bucket-versioning --bucket $BUCKET_NAME --versioning-configuration Status=Enabled
aws s3api put-bucket-encryption --bucket $BUCKET_NAME --server-side-encryption-configuration '{
  "Rules": [
    {
      "ApplyServerSideEncryptionByDefault": {
        "SSEAlgorithm": "AES256"
      }
    }
  ]
}'

echo "S3 bucket $BUCKET_NAME created successfully"
`
		os.WriteFile(filepath.Join(outputPath, "create-s3-backend.sh"), []byte(bucketContent), 0o755)

	case "gcs":
		content.WriteString("# To use this backend, run:\n")
		content.WriteString("# terraform init -backend-config=\"bucket=your-terraform-state-bucket\"\n")
		content.WriteString("# terraform init -backend-config=\"prefix=" + spec.Environment + "\"\n\n")
	}

	return os.WriteFile(filepath.Join(outputPath, "backend.tf.example"), []byte(content.String()), 0o644)
}

func generateModules(spec TerraformSpec) error {
	modulesDir := filepath.Join(outputPath, "modules")
	if err := os.MkdirAll(modulesDir, 0o755); err != nil {
		return err
	}

	// Generate standard modules based on provider
	switch spec.Provider {
	case "aws":
		return generateAWSModules(modulesDir, spec)
	case "gcp":
		return generateGCPModules(modulesDir, spec)
	case "azure":
		return generateAzureModules(modulesDir, spec)
	}

	return nil
}

func generateAWSModules(modulesDir string, spec TerraformSpec) error {
	// VPC module
	vpcDir := filepath.Join(modulesDir, "vpc")
	if err := os.MkdirAll(vpcDir, 0o755); err != nil {
		return err
	}

	vpcContent := `resource "aws_vpc" "main" {
  cidr_block           = var.cidr_block
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = merge(var.tags, {
    Name = "${var.name_prefix}-vpc"
  })
}

resource "aws_subnet" "public" {
  count = length(var.public_subnets)

  vpc_id                  = aws_vpc.main.id
  cidr_block              = var.public_subnets[count.index]
  availability_zone       = data.aws_availability_zones.available.names[count.index]
  map_public_ip_on_launch = true

  tags = merge(var.tags, {
    Name = "${var.name_prefix}-public-${count.index + 1}"
    Type = "Public"
  })
}

resource "aws_subnet" "private" {
  count = length(var.private_subnets)

  vpc_id            = aws_vpc.main.id
  cidr_block        = var.private_subnets[count.index]
  availability_zone = data.aws_availability_zones.available.names[count.index]

  tags = merge(var.tags, {
    Name = "${var.name_prefix}-private-${count.index + 1}"
    Type = "Private"
  })
}

resource "aws_internet_gateway" "main" {
  vpc_id = aws_vpc.main.id

  tags = merge(var.tags, {
    Name = "${var.name_prefix}-igw"
  })
}

data "aws_availability_zones" "available" {
  state = "available"
}`

	if err := os.WriteFile(filepath.Join(vpcDir, "main.tf"), []byte(vpcContent), 0o644); err != nil {
		return err
	}

	vpcVariables := `variable "name_prefix" {
  description = "Prefix for resource names"
  type        = string
}

variable "cidr_block" {
  description = "CIDR block for VPC"
  type        = string
  default     = "10.0.0.0/16"
}

variable "public_subnets" {
  description = "CIDR blocks for public subnets"
  type        = list(string)
  default     = ["10.0.1.0/24", "10.0.2.0/24"]
}

variable "private_subnets" {
  description = "CIDR blocks for private subnets"
  type        = list(string)
  default     = ["10.0.3.0/24", "10.0.4.0/24"]
}

variable "tags" {
  description = "Tags to apply to resources"
  type        = map(string)
  default     = {}
}`

	if err := os.WriteFile(filepath.Join(vpcDir, "variables.tf"), []byte(vpcVariables), 0o644); err != nil {
		return err
	}

	vpcOutputs := `output "vpc_id" {
  description = "ID of the VPC"
  value       = aws_vpc.main.id
}

output "vpc_cidr_block" {
  description = "CIDR block of the VPC"
  value       = aws_vpc.main.cidr_block
}

output "public_subnet_ids" {
  description = "IDs of the public subnets"
  value       = aws_subnet.public[*].id
}

output "private_subnet_ids" {
  description = "IDs of the private subnets"
  value       = aws_subnet.private[*].id
}

output "internet_gateway_id" {
  description = "ID of the Internet Gateway"
  value       = aws_internet_gateway.main.id
}`

	return os.WriteFile(filepath.Join(vpcDir, "outputs.tf"), []byte(vpcOutputs), 0o644)
}

func generateGCPModules(modulesDir string, spec TerraformSpec) error {
	// Network module
	networkDir := filepath.Join(modulesDir, "network")
	if err := os.MkdirAll(networkDir, 0o755); err != nil {
		return err
	}

	networkContent := `resource "google_compute_network" "main" {
  name                    = var.network_name
  auto_create_subnetworks = false
  description             = "VPC network for ${var.project_name}"
}

resource "google_compute_subnetwork" "subnets" {
  for_each = var.subnets

  name          = each.key
  network       = google_compute_network.main.id
  ip_cidr_range = each.value.cidr
  region        = each.value.region

  dynamic "secondary_ip_range" {
    for_each = lookup(each.value, "secondary_ranges", {})
    content {
      range_name    = secondary_ip_range.key
      ip_cidr_range = secondary_ip_range.value
    }
  }

  private_ip_google_access = lookup(each.value, "private_google_access", true)
}`

	if err := os.WriteFile(filepath.Join(networkDir, "main.tf"), []byte(networkContent), 0o644); err != nil {
		return err
	}

	return nil
}

func generateAzureModules(modulesDir string, spec TerraformSpec) error {
	// Resource Group module
	rgDir := filepath.Join(modulesDir, "resource-group")
	if err := os.MkdirAll(rgDir, 0o755); err != nil {
		return err
	}

	rgContent := `resource "azurerm_resource_group" "main" {
  name     = var.resource_group_name
  location = var.location

  tags = var.tags
}`

	if err := os.WriteFile(filepath.Join(rgDir, "main.tf"), []byte(rgContent), 0o644); err != nil {
		return err
	}

	return nil
}

func generateWorkflow(spec TerraformSpec) error {
	workflowDir := filepath.Join(outputPath, ".github", "workflows")
	if err := os.MkdirAll(workflowDir, 0o755); err != nil {
		return err
	}

	workflowContent := `name: Terraform CI/CD

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

env:
  TF_VERSION: 1.5.0

jobs:
  terraform:
    name: Terraform
    runs-on: ubuntu-latest
    
    steps:
    - name: Checkout
      uses: actions/checkout@v3

    - name: Setup Terraform
      uses: hashicorp/setup-terraform@v2
      with:
        terraform_version: ${{ env.TF_VERSION }}

    - name: Terraform Format
      run: terraform fmt -check

    - name: Terraform Init
      run: terraform init

    - name: Terraform Validate
      run: terraform validate

    - name: Terraform Plan
      run: terraform plan -no-color
      env:`

	switch spec.Provider {
	case "aws":
		workflowContent += `
        AWS_ACCESS_KEY_ID: ${{ secrets.AWS_ACCESS_KEY_ID }}
        AWS_SECRET_ACCESS_KEY: ${{ secrets.AWS_SECRET_ACCESS_KEY }}`
	case "gcp":
		workflowContent += `
        GOOGLE_CREDENTIALS: ${{ secrets.GCP_SA_KEY }}`
	case "azure":
		workflowContent += `
        ARM_CLIENT_ID: ${{ secrets.ARM_CLIENT_ID }}
        ARM_CLIENT_SECRET: ${{ secrets.ARM_CLIENT_SECRET }}
        ARM_SUBSCRIPTION_ID: ${{ secrets.ARM_SUBSCRIPTION_ID }}
        ARM_TENANT_ID: ${{ secrets.ARM_TENANT_ID }}`
	}

	workflowContent += `

    - name: Terraform Apply
      if: github.ref == 'refs/heads/main' && github.event_name == 'push'
      run: terraform apply -auto-approve
      env:`

	switch spec.Provider {
	case "aws":
		workflowContent += `
        AWS_ACCESS_KEY_ID: ${{ secrets.AWS_ACCESS_KEY_ID }}
        AWS_SECRET_ACCESS_KEY: ${{ secrets.AWS_SECRET_ACCESS_KEY }}`
	case "gcp":
		workflowContent += `
        GOOGLE_CREDENTIALS: ${{ secrets.GCP_SA_KEY }}`
	case "azure":
		workflowContent += `
        ARM_CLIENT_ID: ${{ secrets.ARM_CLIENT_ID }}
        ARM_CLIENT_SECRET: ${{ secrets.ARM_CLIENT_SECRET }}
        ARM_SUBSCRIPTION_ID: ${{ secrets.ARM_SUBSCRIPTION_ID }}
        ARM_TENANT_ID: ${{ secrets.ARM_TENANT_ID }}`
	}

	return os.WriteFile(filepath.Join(workflowDir, "terraform.yml"), []byte(workflowContent), 0o644)
}

func generateResourceBlocks(spec TerraformSpec) string {
	var content strings.Builder

	switch spec.Provider {
	case "aws":
		content.WriteString(generateAWSResources(spec))
	case "gcp":
		content.WriteString(generateGCPResources(spec))
	case "azure":
		content.WriteString(generateAzureResources(spec))
	}

	return content.String()
}

func generateAWSResources(spec TerraformSpec) string {
	switch spec.Resource {
	case "vpc":
		return `# VPC
resource "aws_vpc" "main" {
  cidr_block           = var.vpc_cidr
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = merge(local.common_tags, {
    Name = "${var.project_name}-vpc"
  })
}

# Public Subnet
resource "aws_subnet" "public" {
  vpc_id                  = aws_vpc.main.id
  cidr_block              = "10.0.1.0/24"
  availability_zone       = data.aws_availability_zones.available.names[0]
  map_public_ip_on_launch = true

  tags = merge(local.common_tags, {
    Name = "${var.project_name}-public-subnet"
    Type = "Public"
  })
}

# Private Subnet
resource "aws_subnet" "private" {
  vpc_id            = aws_vpc.main.id
  cidr_block        = "10.0.2.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]

  tags = merge(local.common_tags, {
    Name = "${var.project_name}-private-subnet"
    Type = "Private"
  })
}

# Internet Gateway
resource "aws_internet_gateway" "main" {
  vpc_id = aws_vpc.main.id

  tags = merge(local.common_tags, {
    Name = "${var.project_name}-igw"
  })
}

# Data source for availability zones
data "aws_availability_zones" "available" {
  state = "available"
}
`
	case "ec2":
		return `# EC2 Instance
resource "aws_instance" "main" {
  ami           = data.aws_ami.ubuntu.id
  instance_type = var.instance_type

  vpc_security_group_ids = [aws_security_group.main.id]
  subnet_id              = aws_subnet.public.id

  tags = merge(local.common_tags, {
    Name = "${var.project_name}-instance"
  })
}

# Security Group
resource "aws_security_group" "main" {
  name_prefix = "${var.project_name}-sg"
  vpc_id      = aws_vpc.main.id

  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = local.common_tags
}

# Data source for Ubuntu AMI
data "aws_ami" "ubuntu" {
  most_recent = true
  owners      = ["099720109477"] # Canonical

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-focal-20.04-amd64-server-*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }
}
`
	default:
		return "# Add your AWS resources here\n"
	}
}

func generateGCPResources(spec TerraformSpec) string {
	switch spec.Resource {
	case "vpc":
		return `# VPC Network
resource "google_compute_network" "main" {
  name                    = "${var.project_name}-vpc"
  auto_create_subnetworks = false
  description             = "VPC network for ${var.project_name}"
}

# Subnet
resource "google_compute_subnetwork" "main" {
  name          = "${var.project_name}-subnet"
  network       = google_compute_network.main.id
  ip_cidr_range = "10.0.1.0/24"
  region        = var.gcp_region

  private_ip_google_access = true
}
`
	case "gke-cluster":
		return `# GKE Cluster
resource "google_container_cluster" "main" {
  name     = "${var.project_name}-gke"
  location = var.gcp_region

  # We can't create a cluster with no node pool defined, but we want to only use
  # separately managed node pools. So we create the smallest possible default
  # node pool and immediately delete it.
  remove_default_node_pool = true
  initial_node_count       = 1

  network    = google_compute_network.main.id
  subnetwork = google_compute_subnetwork.main.id
}

# Node Pool
resource "google_container_node_pool" "main" {
  name       = "${var.project_name}-nodes"
  location   = var.gcp_region
  cluster    = google_container_cluster.main.name
  node_count = 2

  node_config {
    preemptible  = true
    machine_type = "e2-medium"

    # Google recommends custom service accounts that have cloud-platform scope and permissions granted via IAM Roles.
    oauth_scopes = [
      "https://www.googleapis.com/auth/cloud-platform"
    ]
  }
}
`
	default:
		return "# Add your GCP resources here\n"
	}
}

func generateAzureResources(spec TerraformSpec) string {
	switch spec.Resource {
	case "resource-group":
		return `# Resource Group
resource "azurerm_resource_group" "main" {
  name     = "${var.project_name}-rg"
  location = var.azure_location

  tags = local.common_tags
}
`
	case "virtual-network":
		return `# Virtual Network
resource "azurerm_virtual_network" "main" {
  name                = "${var.project_name}-vnet"
  address_space       = ["10.0.0.0/16"]
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name

  tags = local.common_tags
}

# Subnet
resource "azurerm_subnet" "main" {
  name                 = "${var.project_name}-subnet"
  resource_group_name  = azurerm_resource_group.main.name
  virtual_network_name = azurerm_virtual_network.main.name
  address_prefixes     = ["10.0.1.0/24"]
}
`
	default:
		return "# Add your Azure resources here\n"
	}
}

func generateModuleBlocks(spec TerraformSpec) string {
	var content strings.Builder

	switch spec.Module {
	case "networking":
		content.WriteString("# Networking Module\n")
		content.WriteString("module \"networking\" {\n")
		content.WriteString("  source = \"./modules/vpc\"\n")
		content.WriteString("\n")
		content.WriteString("  name_prefix = var.project_name\n")
		content.WriteString("  tags        = local.common_tags\n")
		content.WriteString("}\n\n")

	case "compute":
		content.WriteString("# Compute Module\n")
		content.WriteString("module \"compute\" {\n")
		content.WriteString("  source = \"./modules/compute\"\n")
		content.WriteString("\n")
		content.WriteString("  name_prefix = var.project_name\n")
		content.WriteString("  vpc_id      = module.networking.vpc_id\n")
		content.WriteString("  subnet_ids  = module.networking.public_subnet_ids\n")
		content.WriteString("  tags        = local.common_tags\n")
		content.WriteString("}\n\n")
	}

	return content.String()
}

func getDefaultRegion(provider string) string {
	switch provider {
	case "aws":
		return "us-west-2"
	case "gcp":
		return "us-central1"
	case "azure":
		return "East US"
	default:
		return ""
	}
}
