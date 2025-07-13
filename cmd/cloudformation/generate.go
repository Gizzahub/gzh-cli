package cloudformation

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// GenerateCmd represents the generate command
var GenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate CloudFormation templates",
	Long: `Generate CloudFormation templates for various AWS resources.

Supports generation of:
- VPC and networking infrastructure
- EC2 instances and Auto Scaling groups
- RDS databases and clusters
- Lambda functions and API Gateway
- S3 buckets and CloudFront distributions
- IAM roles and policies
- EKS clusters and ECS services
- Load balancers and target groups

Examples:
  gz cloudformation generate --resource vpc --environment production
  gz cloudformation generate --template lambda-api --runtime nodejs18.x
  gz cloudformation generate --stack my-infrastructure --with-outputs`,
	Run: runGenerate,
}

var (
	resourceType   string
	templateName   string
	stackName      string
	environment    string
	region         string
	outputPath     string
	format         string
	withOutputs    bool
	withParameters bool
	withMetadata   bool
	runtime        string
	instanceType   string
	dbEngine       string
	parameters     []string
	generateTags   []string
)

func init() {
	GenerateCmd.Flags().StringVarP(&resourceType, "resource", "r", "", "Resource type to generate (vpc, ec2, rds, lambda, etc.)")
	GenerateCmd.Flags().StringVarP(&templateName, "template", "t", "", "Template name/type")
	GenerateCmd.Flags().StringVarP(&stackName, "stack", "s", "", "CloudFormation stack name")
	GenerateCmd.Flags().StringVarP(&environment, "environment", "e", "dev", "Target environment")
	GenerateCmd.Flags().StringVar(&region, "region", "us-west-2", "AWS region")
	GenerateCmd.Flags().StringVarP(&outputPath, "output", "o", ".", "Output directory")
	GenerateCmd.Flags().StringVar(&format, "format", "yaml", "Output format (yaml, json)")
	GenerateCmd.Flags().BoolVar(&withOutputs, "with-outputs", true, "Include outputs section")
	GenerateCmd.Flags().BoolVar(&withParameters, "with-parameters", true, "Include parameters section")
	GenerateCmd.Flags().BoolVar(&withMetadata, "with-metadata", false, "Include metadata section")
	GenerateCmd.Flags().StringVar(&runtime, "runtime", "", "Runtime for Lambda functions")
	GenerateCmd.Flags().StringVar(&instanceType, "instance-type", "t3.micro", "EC2 instance type")
	GenerateCmd.Flags().StringVar(&dbEngine, "db-engine", "mysql", "RDS database engine")
	GenerateCmd.Flags().StringSliceVar(&parameters, "params", []string{}, "Template parameters")
	GenerateCmd.Flags().StringSliceVar(&generateTags, "tags", []string{}, "Resource tags")
}

// CloudFormationTemplate represents a CloudFormation template
type CloudFormationTemplate struct {
	AWSTemplateFormatVersion string                 `json:"AWSTemplateFormatVersion,omitempty" yaml:"AWSTemplateFormatVersion,omitempty"`
	Description              string                 `json:"Description,omitempty" yaml:"Description,omitempty"`
	Metadata                 map[string]interface{} `json:"Metadata,omitempty" yaml:"Metadata,omitempty"`
	Parameters               map[string]Parameter   `json:"Parameters,omitempty" yaml:"Parameters,omitempty"`
	Mappings                 map[string]interface{} `json:"Mappings,omitempty" yaml:"Mappings,omitempty"`
	Conditions               map[string]interface{} `json:"Conditions,omitempty" yaml:"Conditions,omitempty"`
	Transform                interface{}            `json:"Transform,omitempty" yaml:"Transform,omitempty"`
	Resources                map[string]Resource    `json:"Resources" yaml:"Resources"`
	Outputs                  map[string]Output      `json:"Outputs,omitempty" yaml:"Outputs,omitempty"`
}

type Parameter struct {
	Type                  string      `json:"Type" yaml:"Type"`
	Description           string      `json:"Description,omitempty" yaml:"Description,omitempty"`
	Default               interface{} `json:"Default,omitempty" yaml:"Default,omitempty"`
	AllowedValues         []string    `json:"AllowedValues,omitempty" yaml:"AllowedValues,omitempty"`
	AllowedPattern        string      `json:"AllowedPattern,omitempty" yaml:"AllowedPattern,omitempty"`
	ConstraintDescription string      `json:"ConstraintDescription,omitempty" yaml:"ConstraintDescription,omitempty"`
	MinLength             int         `json:"MinLength,omitempty" yaml:"MinLength,omitempty"`
	MaxLength             int         `json:"MaxLength,omitempty" yaml:"MaxLength,omitempty"`
	MinValue              int         `json:"MinValue,omitempty" yaml:"MinValue,omitempty"`
	MaxValue              int         `json:"MaxValue,omitempty" yaml:"MaxValue,omitempty"`
	NoEcho                bool        `json:"NoEcho,omitempty" yaml:"NoEcho,omitempty"`
}

type Resource struct {
	Type                string                 `json:"Type" yaml:"Type"`
	Properties          map[string]interface{} `json:"Properties,omitempty" yaml:"Properties,omitempty"`
	DependsOn           interface{}            `json:"DependsOn,omitempty" yaml:"DependsOn,omitempty"`
	Metadata            map[string]interface{} `json:"Metadata,omitempty" yaml:"Metadata,omitempty"`
	CreationPolicy      map[string]interface{} `json:"CreationPolicy,omitempty" yaml:"CreationPolicy,omitempty"`
	UpdatePolicy        map[string]interface{} `json:"UpdatePolicy,omitempty" yaml:"UpdatePolicy,omitempty"`
	DeletionPolicy      string                 `json:"DeletionPolicy,omitempty" yaml:"DeletionPolicy,omitempty"`
	UpdateReplacePolicy string                 `json:"UpdateReplacePolicy,omitempty" yaml:"UpdateReplacePolicy,omitempty"`
	Condition           string                 `json:"Condition,omitempty" yaml:"Condition,omitempty"`
}

type Output struct {
	Description string      `json:"Description,omitempty" yaml:"Description,omitempty"`
	Value       interface{} `json:"Value" yaml:"Value"`
	Export      *Export     `json:"Export,omitempty" yaml:"Export,omitempty"`
	Condition   string      `json:"Condition,omitempty" yaml:"Condition,omitempty"`
}

type Export struct {
	Name interface{} `json:"Name" yaml:"Name"`
}

// TemplateSpec holds template generation specifications
type TemplateSpec struct {
	ResourceType   string
	TemplateName   string
	StackName      string
	Environment    string
	Region         string
	Format         string
	WithOutputs    bool
	WithParameters bool
	WithMetadata   bool
	Runtime        string
	InstanceType   string
	DBEngine       string
	Parameters     map[string]string
	Tags           map[string]string
}

func runGenerate(cmd *cobra.Command, args []string) {
	// Parse parameters
	paramsMap := make(map[string]string)
	for _, p := range parameters {
		parts := strings.SplitN(p, "=", 2)
		if len(parts) == 2 {
			paramsMap[parts[0]] = parts[1]
		}
	}

	// Parse tags
	tagsMap := make(map[string]string)
	for _, t := range generateTags {
		parts := strings.SplitN(t, "=", 2)
		if len(parts) == 2 {
			tagsMap[parts[0]] = parts[1]
		}
	}

	// Default values
	if stackName == "" {
		if resourceType != "" {
			stackName = fmt.Sprintf("%s-%s-stack", resourceType, environment)
		} else if templateName != "" {
			stackName = fmt.Sprintf("%s-%s-stack", templateName, environment)
		} else {
			stackName = fmt.Sprintf("infrastructure-%s-stack", environment)
		}
	}

	spec := TemplateSpec{
		ResourceType:   resourceType,
		TemplateName:   templateName,
		StackName:      stackName,
		Environment:    environment,
		Region:         region,
		Format:         format,
		WithOutputs:    withOutputs,
		WithParameters: withParameters,
		WithMetadata:   withMetadata,
		Runtime:        runtime,
		InstanceType:   instanceType,
		DBEngine:       dbEngine,
		Parameters:     paramsMap,
		Tags:           tagsMap,
	}

	fmt.Printf("🏗️ Generating CloudFormation template\n")
	fmt.Printf("📦 Stack: %s\n", stackName)
	fmt.Printf("🌍 Environment: %s\n", environment)
	fmt.Printf("📍 Region: %s\n", region)

	// Create output directory
	if err := os.MkdirAll(outputPath, 0o755); err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	// Generate template
	template, err := generateTemplate(spec)
	if err != nil {
		fmt.Printf("Error generating template: %v\n", err)
		os.Exit(1)
	}

	// Write template file
	filename := fmt.Sprintf("%s.%s", stackName, spec.Format)
	filepath := filepath.Join(outputPath, filename)

	if err := writeTemplate(template, filepath, spec.Format); err != nil {
		fmt.Printf("Error writing template: %v\n", err)
		os.Exit(1)
	}

	// Generate parameter files
	if spec.WithParameters {
		if err := generateParameterFiles(spec); err != nil {
			fmt.Printf("Error generating parameter files: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Printf("✅ CloudFormation template generated successfully\n")
	fmt.Printf("📁 Template file: %s\n", filepath)
	fmt.Printf("\n📝 Next steps:\n")
	fmt.Printf("1. Review the generated template\n")
	fmt.Printf("2. Validate template: aws cloudformation validate-template --template-body file://%s\n", filename)
	fmt.Printf("3. Deploy stack: aws cloudformation deploy --template-file %s --stack-name %s\n", filename, stackName)
}

func generateTemplate(spec TemplateSpec) (*CloudFormationTemplate, error) {
	template := &CloudFormationTemplate{
		AWSTemplateFormatVersion: "2010-09-09",
		Description:              fmt.Sprintf("CloudFormation template for %s (%s environment)", spec.StackName, spec.Environment),
		Resources:                make(map[string]Resource),
	}

	// Add metadata if requested
	if spec.WithMetadata {
		template.Metadata = map[string]interface{}{
			"AWS::CloudFormation::Designer": map[string]interface{}{
				"GeneratedBy": "gzh-manager-go",
				"Version":     "1.0.0",
			},
		}
	}

	// Add parameters if requested
	if spec.WithParameters {
		template.Parameters = generateParameters(spec)
	}

	// Generate resources based on type
	var err error
	switch spec.ResourceType {
	case "vpc":
		err = addVPCResources(template, spec)
	case "ec2":
		err = addEC2Resources(template, spec)
	case "rds":
		err = addRDSResources(template, spec)
	case "lambda":
		err = addLambdaResources(template, spec)
	case "s3":
		err = addS3Resources(template, spec)
	case "iam":
		err = addIAMResources(template, spec)
	case "ecs":
		err = addECSResources(template, spec)
	case "eks":
		err = addEKSResources(template, spec)
	default:
		// Generate based on template name if resource type not specified
		if spec.TemplateName != "" {
			err = addTemplateResources(template, spec)
		} else {
			err = addBasicInfrastructure(template, spec)
		}
	}

	if err != nil {
		return nil, err
	}

	// Add outputs if requested
	if spec.WithOutputs {
		template.Outputs = generateOutputs(template, spec)
	}

	return template, nil
}

func generateParameters(spec TemplateSpec) map[string]Parameter {
	params := make(map[string]Parameter)

	// Common parameters
	params["Environment"] = Parameter{
		Type:          "String",
		Description:   "Environment name",
		Default:       spec.Environment,
		AllowedValues: []string{"dev", "staging", "production"},
	}

	params["ProjectName"] = Parameter{
		Type:        "String",
		Description: "Name of the project",
		Default:     "MyProject",
	}

	// Resource-specific parameters
	switch spec.ResourceType {
	case "vpc":
		params["VpcCidr"] = Parameter{
			Type:           "String",
			Description:    "CIDR block for VPC",
			Default:        "10.0.0.0/16",
			AllowedPattern: `^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])(\/([0-9]|[1-2][0-9]|3[0-2]))$`,
		}
	case "ec2":
		params["InstanceType"] = Parameter{
			Type:          "String",
			Description:   "EC2 instance type",
			Default:       spec.InstanceType,
			AllowedValues: []string{"t3.micro", "t3.small", "t3.medium", "t3.large"},
		}
		params["KeyPairName"] = Parameter{
			Type:        "String",
			Description: "Name of EC2 Key Pair",
		}
	case "rds":
		params["DBEngine"] = Parameter{
			Type:          "String",
			Description:   "Database engine",
			Default:       spec.DBEngine,
			AllowedValues: []string{"mysql", "postgres", "mariadb"},
		}
		params["DBInstanceClass"] = Parameter{
			Type:        "String",
			Description: "RDS instance class",
			Default:     "db.t3.micro",
		}
		params["MasterUsername"] = Parameter{
			Type:        "String",
			Description: "Master username for RDS instance",
			Default:     "admin",
		}
		params["MasterPassword"] = Parameter{
			Type:        "String",
			Description: "Master password for RDS instance",
			NoEcho:      true,
			MinLength:   8,
			MaxLength:   41,
		}
	}

	return params
}

func addVPCResources(template *CloudFormationTemplate, spec TemplateSpec) error {
	// VPC
	template.Resources["VPC"] = Resource{
		Type: "AWS::EC2::VPC",
		Properties: map[string]interface{}{
			"CidrBlock":          map[string]interface{}{"Ref": "VpcCidr"},
			"EnableDnsHostnames": true,
			"EnableDnsSupport":   true,
			"Tags": []map[string]interface{}{
				{
					"Key":   "Name",
					"Value": map[string]interface{}{"Fn::Sub": "${ProjectName}-vpc"},
				},
				{
					"Key":   "Environment",
					"Value": map[string]interface{}{"Ref": "Environment"},
				},
			},
		},
	}

	// Internet Gateway
	template.Resources["InternetGateway"] = Resource{
		Type: "AWS::EC2::InternetGateway",
		Properties: map[string]interface{}{
			"Tags": []map[string]interface{}{
				{
					"Key":   "Name",
					"Value": map[string]interface{}{"Fn::Sub": "${ProjectName}-igw"},
				},
			},
		},
	}

	// Attach Gateway
	template.Resources["AttachGateway"] = Resource{
		Type: "AWS::EC2::VPCGatewayAttachment",
		Properties: map[string]interface{}{
			"VpcId":             map[string]interface{}{"Ref": "VPC"},
			"InternetGatewayId": map[string]interface{}{"Ref": "InternetGateway"},
		},
	}

	// Public Subnet
	template.Resources["PublicSubnet"] = Resource{
		Type: "AWS::EC2::Subnet",
		Properties: map[string]interface{}{
			"VpcId":               map[string]interface{}{"Ref": "VPC"},
			"CidrBlock":           "10.0.1.0/24",
			"AvailabilityZone":    map[string]interface{}{"Fn::Select": []interface{}{0, map[string]interface{}{"Fn::GetAZs": ""}}},
			"MapPublicIpOnLaunch": true,
			"Tags": []map[string]interface{}{
				{
					"Key":   "Name",
					"Value": map[string]interface{}{"Fn::Sub": "${ProjectName}-public-subnet"},
				},
			},
		},
	}

	// Private Subnet
	template.Resources["PrivateSubnet"] = Resource{
		Type: "AWS::EC2::Subnet",
		Properties: map[string]interface{}{
			"VpcId":            map[string]interface{}{"Ref": "VPC"},
			"CidrBlock":        "10.0.2.0/24",
			"AvailabilityZone": map[string]interface{}{"Fn::Select": []interface{}{1, map[string]interface{}{"Fn::GetAZs": ""}}},
			"Tags": []map[string]interface{}{
				{
					"Key":   "Name",
					"Value": map[string]interface{}{"Fn::Sub": "${ProjectName}-private-subnet"},
				},
			},
		},
	}

	// Route Table
	template.Resources["PublicRouteTable"] = Resource{
		Type: "AWS::EC2::RouteTable",
		Properties: map[string]interface{}{
			"VpcId": map[string]interface{}{"Ref": "VPC"},
			"Tags": []map[string]interface{}{
				{
					"Key":   "Name",
					"Value": map[string]interface{}{"Fn::Sub": "${ProjectName}-public-rt"},
				},
			},
		},
	}

	// Public Route
	template.Resources["PublicRoute"] = Resource{
		Type:      "AWS::EC2::Route",
		DependsOn: "AttachGateway",
		Properties: map[string]interface{}{
			"RouteTableId":         map[string]interface{}{"Ref": "PublicRouteTable"},
			"DestinationCidrBlock": "0.0.0.0/0",
			"GatewayId":            map[string]interface{}{"Ref": "InternetGateway"},
		},
	}

	// Subnet Route Table Association
	template.Resources["PublicSubnetRouteTableAssociation"] = Resource{
		Type: "AWS::EC2::SubnetRouteTableAssociation",
		Properties: map[string]interface{}{
			"SubnetId":     map[string]interface{}{"Ref": "PublicSubnet"},
			"RouteTableId": map[string]interface{}{"Ref": "PublicRouteTable"},
		},
	}

	return nil
}

func addEC2Resources(template *CloudFormationTemplate, spec TemplateSpec) error {
	// Security Group
	template.Resources["SecurityGroup"] = Resource{
		Type: "AWS::EC2::SecurityGroup",
		Properties: map[string]interface{}{
			"GroupDescription": "Security group for EC2 instance",
			"VpcId":            map[string]interface{}{"Ref": "VPC"},
			"SecurityGroupIngress": []map[string]interface{}{
				{
					"IpProtocol": "tcp",
					"FromPort":   22,
					"ToPort":     22,
					"CidrIp":     "0.0.0.0/0",
				},
				{
					"IpProtocol": "tcp",
					"FromPort":   80,
					"ToPort":     80,
					"CidrIp":     "0.0.0.0/0",
				},
			},
		},
	}

	// EC2 Instance
	template.Resources["EC2Instance"] = Resource{
		Type: "AWS::EC2::Instance",
		Properties: map[string]interface{}{
			"ImageId":      map[string]interface{}{"Ref": "LatestAmiId"},
			"InstanceType": map[string]interface{}{"Ref": "InstanceType"},
			"KeyName":      map[string]interface{}{"Ref": "KeyPairName"},
			"SecurityGroupIds": []interface{}{
				map[string]interface{}{"Ref": "SecurityGroup"},
			},
			"SubnetId": map[string]interface{}{"Ref": "PublicSubnet"},
			"Tags": []map[string]interface{}{
				{
					"Key":   "Name",
					"Value": map[string]interface{}{"Fn::Sub": "${ProjectName}-instance"},
				},
			},
		},
	}

	// Add parameter for latest AMI
	if template.Parameters == nil {
		template.Parameters = make(map[string]Parameter)
	}
	template.Parameters["LatestAmiId"] = Parameter{
		Type:    "AWS::SSM::Parameter::Value<AWS::EC2::Image::Id>",
		Default: "/aws/service/ami-amazon-linux-latest/amzn2-ami-hvm-x86_64-gp2",
	}

	return nil
}

func addRDSResources(template *CloudFormationTemplate, spec TemplateSpec) error {
	// DB Subnet Group
	template.Resources["DBSubnetGroup"] = Resource{
		Type: "AWS::RDS::DBSubnetGroup",
		Properties: map[string]interface{}{
			"DBSubnetGroupDescription": "Subnet group for RDS instance",
			"SubnetIds": []interface{}{
				map[string]interface{}{"Ref": "PrivateSubnet"},
				map[string]interface{}{"Ref": "PrivateSubnet2"},
			},
			"Tags": []map[string]interface{}{
				{
					"Key":   "Name",
					"Value": map[string]interface{}{"Fn::Sub": "${ProjectName}-db-subnet-group"},
				},
			},
		},
	}

	// DB Security Group
	template.Resources["DBSecurityGroup"] = Resource{
		Type: "AWS::EC2::SecurityGroup",
		Properties: map[string]interface{}{
			"GroupDescription": "Security group for RDS instance",
			"VpcId":            map[string]interface{}{"Ref": "VPC"},
			"SecurityGroupIngress": []map[string]interface{}{
				{
					"IpProtocol":            "tcp",
					"FromPort":              3306,
					"ToPort":                3306,
					"SourceSecurityGroupId": map[string]interface{}{"Ref": "SecurityGroup"},
				},
			},
		},
	}

	// RDS Instance
	template.Resources["RDSInstance"] = Resource{
		Type: "AWS::RDS::DBInstance",
		Properties: map[string]interface{}{
			"DBInstanceIdentifier": map[string]interface{}{"Fn::Sub": "${ProjectName}-db"},
			"DBInstanceClass":      map[string]interface{}{"Ref": "DBInstanceClass"},
			"Engine":               map[string]interface{}{"Ref": "DBEngine"},
			"MasterUsername":       map[string]interface{}{"Ref": "MasterUsername"},
			"MasterUserPassword":   map[string]interface{}{"Ref": "MasterPassword"},
			"AllocatedStorage":     20,
			"DBSubnetGroupName":    map[string]interface{}{"Ref": "DBSubnetGroup"},
			"VPCSecurityGroups": []interface{}{
				map[string]interface{}{"Ref": "DBSecurityGroup"},
			},
			"BackupRetentionPeriod": 7,
			"MultiAZ":               false,
			"StorageType":           "gp2",
			"StorageEncrypted":      true,
		},
		DeletionPolicy: "Snapshot",
	}

	return nil
}

func addLambdaResources(template *CloudFormationTemplate, spec TemplateSpec) error {
	// Lambda Execution Role
	template.Resources["LambdaExecutionRole"] = Resource{
		Type: "AWS::IAM::Role",
		Properties: map[string]interface{}{
			"AssumeRolePolicyDocument": map[string]interface{}{
				"Version": "2012-10-17",
				"Statement": []map[string]interface{}{
					{
						"Effect": "Allow",
						"Principal": map[string]interface{}{
							"Service": "lambda.amazonaws.com",
						},
						"Action": "sts:AssumeRole",
					},
				},
			},
			"ManagedPolicyArns": []string{
				"arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole",
			},
		},
	}

	// Lambda Function
	runtime := spec.Runtime
	if runtime == "" {
		runtime = "nodejs18.x"
	}

	template.Resources["LambdaFunction"] = Resource{
		Type: "AWS::Lambda::Function",
		Properties: map[string]interface{}{
			"FunctionName": map[string]interface{}{"Fn::Sub": "${ProjectName}-function"},
			"Runtime":      runtime,
			"Handler":      "index.handler",
			"Role":         map[string]interface{}{"Fn::GetAtt": []string{"LambdaExecutionRole", "Arn"}},
			"Code": map[string]interface{}{
				"ZipFile": "exports.handler = async (event) => { return { statusCode: 200, body: 'Hello World!' }; };",
			},
			"Timeout": 30,
		},
	}

	return nil
}

func addS3Resources(template *CloudFormationTemplate, spec TemplateSpec) error {
	// S3 Bucket
	template.Resources["S3Bucket"] = Resource{
		Type: "AWS::S3::Bucket",
		Properties: map[string]interface{}{
			"BucketName": map[string]interface{}{"Fn::Sub": "${ProjectName}-${Environment}-bucket"},
			"BucketEncryption": map[string]interface{}{
				"ServerSideEncryptionConfiguration": []map[string]interface{}{
					{
						"ServerSideEncryptionByDefault": map[string]interface{}{
							"SSEAlgorithm": "AES256",
						},
					},
				},
			},
			"PublicAccessBlockConfiguration": map[string]interface{}{
				"BlockPublicAcls":       true,
				"BlockPublicPolicy":     true,
				"IgnorePublicAcls":      true,
				"RestrictPublicBuckets": true,
			},
			"VersioningConfiguration": map[string]interface{}{
				"Status": "Enabled",
			},
		},
	}

	return nil
}

func addIAMResources(template *CloudFormationTemplate, spec TemplateSpec) error {
	// IAM Role
	template.Resources["IAMRole"] = Resource{
		Type: "AWS::IAM::Role",
		Properties: map[string]interface{}{
			"RoleName": map[string]interface{}{"Fn::Sub": "${ProjectName}-role"},
			"AssumeRolePolicyDocument": map[string]interface{}{
				"Version": "2012-10-17",
				"Statement": []map[string]interface{}{
					{
						"Effect": "Allow",
						"Principal": map[string]interface{}{
							"Service": "ec2.amazonaws.com",
						},
						"Action": "sts:AssumeRole",
					},
				},
			},
		},
	}

	// IAM Policy
	template.Resources["IAMPolicy"] = Resource{
		Type: "AWS::IAM::Policy",
		Properties: map[string]interface{}{
			"PolicyName": map[string]interface{}{"Fn::Sub": "${ProjectName}-policy"},
			"PolicyDocument": map[string]interface{}{
				"Version": "2012-10-17",
				"Statement": []map[string]interface{}{
					{
						"Effect": "Allow",
						"Action": []string{
							"s3:GetObject",
							"s3:PutObject",
						},
						"Resource": "*",
					},
				},
			},
			"Roles": []interface{}{
				map[string]interface{}{"Ref": "IAMRole"},
			},
		},
	}

	return nil
}

func addECSResources(template *CloudFormationTemplate, spec TemplateSpec) error {
	// ECS Cluster
	template.Resources["ECSCluster"] = Resource{
		Type: "AWS::ECS::Cluster",
		Properties: map[string]interface{}{
			"ClusterName": map[string]interface{}{"Fn::Sub": "${ProjectName}-cluster"},
		},
	}

	// ECS Task Definition
	template.Resources["ECSTaskDefinition"] = Resource{
		Type: "AWS::ECS::TaskDefinition",
		Properties: map[string]interface{}{
			"Family":                  "web-app",
			"NetworkMode":             "awsvpc",
			"RequiresCompatibilities": []string{"FARGATE"},
			"Cpu":                     "256",
			"Memory":                  "512",
			"ExecutionRoleArn":        map[string]interface{}{"Ref": "ECSExecutionRole"},
			"ContainerDefinitions": []map[string]interface{}{
				{
					"Name":  "web",
					"Image": "nginx:latest",
					"PortMappings": []map[string]interface{}{
						{
							"ContainerPort": 80,
							"Protocol":      "tcp",
						},
					},
					"Essential": true,
					"LogConfiguration": map[string]interface{}{
						"LogDriver": "awslogs",
						"Options": map[string]interface{}{
							"awslogs-group":         map[string]interface{}{"Ref": "LogGroup"},
							"awslogs-region":        map[string]interface{}{"Ref": "AWS::Region"},
							"awslogs-stream-prefix": "ecs",
						},
					},
				},
			},
		},
	}

	return nil
}

func addEKSResources(template *CloudFormationTemplate, spec TemplateSpec) error {
	// EKS Service Role
	template.Resources["EKSServiceRole"] = Resource{
		Type: "AWS::IAM::Role",
		Properties: map[string]interface{}{
			"AssumeRolePolicyDocument": map[string]interface{}{
				"Version": "2012-10-17",
				"Statement": []map[string]interface{}{
					{
						"Effect": "Allow",
						"Principal": map[string]interface{}{
							"Service": "eks.amazonaws.com",
						},
						"Action": "sts:AssumeRole",
					},
				},
			},
			"ManagedPolicyArns": []string{
				"arn:aws:iam::aws:policy/AmazonEKSClusterPolicy",
			},
		},
	}

	// EKS Cluster
	template.Resources["EKSCluster"] = Resource{
		Type: "AWS::EKS::Cluster",
		Properties: map[string]interface{}{
			"Name":    map[string]interface{}{"Fn::Sub": "${ProjectName}-cluster"},
			"Version": "1.28",
			"RoleArn": map[string]interface{}{"Fn::GetAtt": []string{"EKSServiceRole", "Arn"}},
			"ResourcesVpcConfig": map[string]interface{}{
				"SubnetIds": []interface{}{
					map[string]interface{}{"Ref": "PublicSubnet"},
					map[string]interface{}{"Ref": "PrivateSubnet"},
				},
			},
		},
	}

	return nil
}

func addTemplateResources(template *CloudFormationTemplate, spec TemplateSpec) error {
	// Handle specific template types
	switch spec.TemplateName {
	case "lambda-api":
		if err := addLambdaResources(template, spec); err != nil {
			return err
		}
		// Add API Gateway resources
		return addAPIGatewayResources(template, spec)
	case "web-app":
		if err := addVPCResources(template, spec); err != nil {
			return err
		}
		if err := addEC2Resources(template, spec); err != nil {
			return err
		}
		return addLoadBalancerResources(template, spec)
	default:
		return addBasicInfrastructure(template, spec)
	}
}

func addAPIGatewayResources(template *CloudFormationTemplate, spec TemplateSpec) error {
	// API Gateway
	template.Resources["ApiGateway"] = Resource{
		Type: "AWS::ApiGateway::RestApi",
		Properties: map[string]interface{}{
			"Name": map[string]interface{}{"Fn::Sub": "${ProjectName}-api"},
		},
	}

	// API Gateway Resource
	template.Resources["ApiResource"] = Resource{
		Type: "AWS::ApiGateway::Resource",
		Properties: map[string]interface{}{
			"RestApiId": map[string]interface{}{"Ref": "ApiGateway"},
			"ParentId":  map[string]interface{}{"Fn::GetAtt": []string{"ApiGateway", "RootResourceId"}},
			"PathPart":  "hello",
		},
	}

	// API Gateway Method
	template.Resources["ApiMethod"] = Resource{
		Type: "AWS::ApiGateway::Method",
		Properties: map[string]interface{}{
			"RestApiId":         map[string]interface{}{"Ref": "ApiGateway"},
			"ResourceId":        map[string]interface{}{"Ref": "ApiResource"},
			"HttpMethod":        "GET",
			"AuthorizationType": "NONE",
			"Integration": map[string]interface{}{
				"Type":                  "AWS_PROXY",
				"IntegrationHttpMethod": "POST",
				"Uri": map[string]interface{}{
					"Fn::Sub": "arn:aws:apigateway:${AWS::Region}:lambda:path/2015-03-31/functions/${LambdaFunction.Arn}/invocations",
				},
			},
		},
	}

	return nil
}

func addLoadBalancerResources(template *CloudFormationTemplate, spec TemplateSpec) error {
	// Application Load Balancer
	template.Resources["LoadBalancer"] = Resource{
		Type: "AWS::ElasticLoadBalancingV2::LoadBalancer",
		Properties: map[string]interface{}{
			"Name":   map[string]interface{}{"Fn::Sub": "${ProjectName}-alb"},
			"Scheme": "internet-facing",
			"Type":   "application",
			"Subnets": []interface{}{
				map[string]interface{}{"Ref": "PublicSubnet"},
			},
			"SecurityGroups": []interface{}{
				map[string]interface{}{"Ref": "SecurityGroup"},
			},
		},
	}

	// Target Group
	template.Resources["TargetGroup"] = Resource{
		Type: "AWS::ElasticLoadBalancingV2::TargetGroup",
		Properties: map[string]interface{}{
			"Name":     map[string]interface{}{"Fn::Sub": "${ProjectName}-tg"},
			"Port":     80,
			"Protocol": "HTTP",
			"VpcId":    map[string]interface{}{"Ref": "VPC"},
			"Targets": []map[string]interface{}{
				{
					"Id":   map[string]interface{}{"Ref": "EC2Instance"},
					"Port": 80,
				},
			},
		},
	}

	return nil
}

func addBasicInfrastructure(template *CloudFormationTemplate, spec TemplateSpec) error {
	// Add VPC by default
	return addVPCResources(template, spec)
}

func generateOutputs(template *CloudFormationTemplate, spec TemplateSpec) map[string]Output {
	outputs := make(map[string]Output)

	// Generate outputs based on resources
	for name, resource := range template.Resources {
		switch resource.Type {
		case "AWS::EC2::VPC":
			outputs["VpcId"] = Output{
				Description: "VPC ID",
				Value:       map[string]interface{}{"Ref": name},
				Export: &Export{
					Name: map[string]interface{}{"Fn::Sub": "${AWS::StackName}-VpcId"},
				},
			}
		case "AWS::EC2::Instance":
			outputs["InstanceId"] = Output{
				Description: "EC2 Instance ID",
				Value:       map[string]interface{}{"Ref": name},
			}
			outputs["PublicIP"] = Output{
				Description: "EC2 Instance Public IP",
				Value:       map[string]interface{}{"Fn::GetAtt": []string{name, "PublicIp"}},
			}
		case "AWS::RDS::DBInstance":
			outputs["DatabaseEndpoint"] = Output{
				Description: "RDS Instance Endpoint",
				Value:       map[string]interface{}{"Fn::GetAtt": []string{name, "Endpoint.Address"}},
			}
		case "AWS::Lambda::Function":
			outputs["LambdaFunctionArn"] = Output{
				Description: "Lambda Function ARN",
				Value:       map[string]interface{}{"Fn::GetAtt": []string{name, "Arn"}},
			}
		case "AWS::S3::Bucket":
			outputs["BucketName"] = Output{
				Description: "S3 Bucket Name",
				Value:       map[string]interface{}{"Ref": name},
			}
		}
	}

	return outputs
}

func writeTemplate(template *CloudFormationTemplate, filepath, format string) error {
	var data []byte
	var err error

	switch format {
	case "json":
		data, err = json.MarshalIndent(template, "", "  ")
	case "yaml":
		data, err = yaml.Marshal(template)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	if err != nil {
		return err
	}

	return os.WriteFile(filepath, data, 0o644)
}

func generateParameterFiles(spec TemplateSpec) error {
	// Generate parameter file for each environment
	environments := []string{"dev", "staging", "production"}

	for _, env := range environments {
		params := map[string]interface{}{
			"Parameters": map[string]interface{}{
				"Environment": env,
				"ProjectName": spec.StackName,
			},
		}

		filename := fmt.Sprintf("%s-parameters-%s.json", spec.StackName, env)
		filepath := filepath.Join(outputPath, filename)

		data, err := json.MarshalIndent(params, "", "  ")
		if err != nil {
			return err
		}

		if err := os.WriteFile(filepath, data, 0o644); err != nil {
			return err
		}
	}

	return nil
}
