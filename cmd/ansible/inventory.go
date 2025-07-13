package ansible

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// InventoryCmd represents the inventory command
var InventoryCmd = &cobra.Command{
	Use:   "inventory",
	Short: "Ansible 인벤토리 파일 관리",
	Long: `Ansible 인벤토리 파일을 생성하고 관리합니다.

인벤토리 관리 기능:
- 정적 인벤토리 파일 생성 (INI/YAML 형식)
- 동적 인벤토리 스크립트 생성
- 호스트 그룹 관리
- 호스트 변수 설정
- 클라우드 인스턴스 자동 검색
- SSH 연결 테스트

Examples:
  gz ansible inventory create --format ini --groups webservers,databases
  gz ansible inventory create --format yaml --cloud aws --region us-west-2
  gz ansible inventory add-host --group webservers --host 192.168.1.10
  gz ansible inventory test --host all`,
	Run: runInventory,
}

var (
	inventoryAction  string
	inventoryFormat  string
	inventoryGroups  []string
	inventoryHosts   []string
	inventoryFile    string
	cloudProvider    string
	cloudRegion      string
	cloudTags        []string
	hostGroup        string
	hostAddress      string
	hostVars         []string
	sshUser          string
	sshKey           string
	sshPort          int
	testConnection   bool
	generateDynamic  bool
	includeVariables bool
)

func init() {
	InventoryCmd.Flags().StringVarP(&inventoryAction, "action", "a", "create", "액션 (create, add-host, remove-host, test, list)")
	InventoryCmd.Flags().StringVarP(&inventoryFormat, "format", "f", "ini", "인벤토리 형식 (ini, yaml)")
	InventoryCmd.Flags().StringSliceVarP(&inventoryGroups, "groups", "g", []string{"webservers"}, "호스트 그룹")
	InventoryCmd.Flags().StringSliceVar(&inventoryHosts, "hosts", []string{}, "호스트 목록")
	InventoryCmd.Flags().StringVarP(&inventoryFile, "file", "i", "inventory", "인벤토리 파일 경로")
	InventoryCmd.Flags().StringVar(&cloudProvider, "cloud", "", "클라우드 제공자 (aws, gcp, azure)")
	InventoryCmd.Flags().StringVar(&cloudRegion, "region", "", "클라우드 리전")
	InventoryCmd.Flags().StringSliceVar(&cloudTags, "cloud-tags", []string{}, "클라우드 인스턴스 필터 태그")
	InventoryCmd.Flags().StringVar(&hostGroup, "group", "", "호스트 그룹 이름")
	InventoryCmd.Flags().StringVar(&hostAddress, "host", "", "호스트 주소")
	InventoryCmd.Flags().StringSliceVar(&hostVars, "host-vars", []string{}, "호스트 변수 (key=value)")
	InventoryCmd.Flags().StringVar(&sshUser, "ssh-user", "ubuntu", "SSH 사용자")
	InventoryCmd.Flags().StringVar(&sshKey, "ssh-key", "~/.ssh/id_rsa", "SSH 개인키 경로")
	InventoryCmd.Flags().IntVar(&sshPort, "ssh-port", 22, "SSH 포트")
	InventoryCmd.Flags().BoolVar(&testConnection, "test", false, "SSH 연결 테스트")
	InventoryCmd.Flags().BoolVar(&generateDynamic, "dynamic", false, "동적 인벤토리 생성")
	InventoryCmd.Flags().BoolVar(&includeVariables, "include-vars", true, "호스트/그룹 변수 포함")
}

// Inventory represents an Ansible inventory structure
type Inventory struct {
	All      AllGroup               `yaml:"all,omitempty"`
	Children map[string]Group       `yaml:",inline"`
	Meta     map[string]interface{} `yaml:"_meta,omitempty"`
}

// AllGroup represents the special 'all' group
type AllGroup struct {
	Vars     map[string]interface{} `yaml:"vars,omitempty"`
	Children []string               `yaml:"children,omitempty"`
}

// Group represents a host group
type Group struct {
	Hosts    map[string]Host        `yaml:"hosts,omitempty"`
	Vars     map[string]interface{} `yaml:"vars,omitempty"`
	Children []string               `yaml:"children,omitempty"`
}

// Host represents a single host
type Host struct {
	AnsibleHost              string `yaml:"ansible_host,omitempty"`
	AnsibleUser              string `yaml:"ansible_user,omitempty"`
	AnsibleSSHPrivateKeyFile string `yaml:"ansible_ssh_private_key_file,omitempty"`
	AnsiblePort              int    `yaml:"ansible_port,omitempty"`
	AnsibleConnection        string `yaml:"ansible_connection,omitempty"`
	// Custom variables can be added here
	Vars map[string]interface{} `yaml:",inline"`
}

// DynamicInventory represents a dynamic inventory script
type DynamicInventory struct {
	Provider string                 `yaml:"provider"`
	Region   string                 `yaml:"region,omitempty"`
	Filters  map[string]interface{} `yaml:"filters,omitempty"`
	Groups   map[string]string      `yaml:"groups,omitempty"`
	Keyed    []string               `yaml:"keyed_groups,omitempty"`
}

func runInventory(cmd *cobra.Command, args []string) {
	fmt.Printf("📋 Ansible 인벤토리 관리\n")
	fmt.Printf("🎯 액션: %s\n", inventoryAction)

	switch inventoryAction {
	case "create":
		if err := createInventory(); err != nil {
			fmt.Printf("❌ 인벤토리 생성 실패: %v\n", err)
			os.Exit(1)
		}
	case "add-host":
		if err := addHostToInventory(); err != nil {
			fmt.Printf("❌ 호스트 추가 실패: %v\n", err)
			os.Exit(1)
		}
	case "remove-host":
		if err := removeHostFromInventory(); err != nil {
			fmt.Printf("❌ 호스트 제거 실패: %v\n", err)
			os.Exit(1)
		}
	case "test":
		if err := testInventoryConnection(); err != nil {
			fmt.Printf("❌ 연결 테스트 실패: %v\n", err)
			os.Exit(1)
		}
	case "list":
		if err := listInventory(); err != nil {
			fmt.Printf("❌ 인벤토리 목록 실패: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Printf("❌ 알 수 없는 액션: %s\n", inventoryAction)
		cmd.Help()
		os.Exit(1)
	}

	fmt.Printf("✅ 인벤토리 작업 완료\n")
}

func createInventory() error {
	fmt.Printf("🏗️ 인벤토리 생성: %s (형식: %s)\n", inventoryFile, inventoryFormat)

	if generateDynamic {
		return createDynamicInventory()
	}

	inventory := &Inventory{
		Children: make(map[string]Group),
		Meta: map[string]interface{}{
			"hostvars": make(map[string]interface{}),
		},
	}

	// Set up all group
	inventory.All = AllGroup{
		Vars: map[string]interface{}{
			"ansible_user":                 sshUser,
			"ansible_ssh_private_key_file": sshKey,
			"ansible_port":                 sshPort,
			"ansible_ssh_common_args":      "-o StrictHostKeyChecking=no",
		},
		Children: inventoryGroups,
	}

	// Create groups
	for _, groupName := range inventoryGroups {
		group := Group{
			Hosts: make(map[string]Host),
			Vars:  make(map[string]interface{}),
		}

		// Add group-specific variables
		switch groupName {
		case "webservers":
			group.Vars["http_port"] = 80
			group.Vars["https_port"] = 443
		case "databases":
			group.Vars["db_port"] = 3306
		case "loadbalancers":
			group.Vars["backend_port"] = 8080
		}

		inventory.Children[groupName] = group
	}

	// Add hosts if provided
	if len(inventoryHosts) > 0 {
		if err := addHostsToInventory(inventory); err != nil {
			return err
		}
	}

	// Discover cloud instances if cloud provider specified
	if cloudProvider != "" {
		if err := discoverCloudInstances(inventory); err != nil {
			return fmt.Errorf("클라우드 인스턴스 검색 실패: %w", err)
		}
	}

	// Write inventory file
	return writeInventoryFile(inventory)
}

func addHostsToInventory(inventory *Inventory) error {
	for i, hostAddr := range inventoryHosts {
		groupName := inventoryGroups[0] // Default to first group
		if i < len(inventoryGroups) {
			groupName = inventoryGroups[i]
		}

		host := Host{
			AnsibleHost:              hostAddr,
			AnsibleUser:              sshUser,
			AnsibleSSHPrivateKeyFile: sshKey,
			AnsiblePort:              sshPort,
			Vars:                     make(map[string]interface{}),
		}

		// Parse host variables
		for _, hostVar := range hostVars {
			parts := strings.SplitN(hostVar, "=", 2)
			if len(parts) == 2 {
				host.Vars[parts[0]] = parts[1]
			}
		}

		// Add host to group
		if group, exists := inventory.Children[groupName]; exists {
			group.Hosts[hostAddr] = host
			inventory.Children[groupName] = group
		}

		fmt.Printf("✅ 호스트 추가: %s -> %s\n", hostAddr, groupName)
	}

	return nil
}

func discoverCloudInstances(inventory *Inventory) error {
	fmt.Printf("☁️ 클라우드 인스턴스 검색: %s\n", cloudProvider)

	switch cloudProvider {
	case "aws":
		return discoverAWSInstances(inventory)
	case "gcp":
		return discoverGCPInstances(inventory)
	case "azure":
		return discoverAzureInstances(inventory)
	default:
		return fmt.Errorf("지원하지 않는 클라우드 제공자: %s", cloudProvider)
	}
}

func discoverAWSInstances(inventory *Inventory) error {
	// This is a placeholder implementation
	// In a real implementation, you would use AWS SDK to discover EC2 instances
	fmt.Printf("🔍 AWS EC2 인스턴스 검색 중...\n")

	// Example instances (in real implementation, fetch from AWS API)
	exampleInstances := []map[string]interface{}{
		{
			"id":         "i-1234567890abcdef0",
			"public_ip":  "203.0.113.12",
			"private_ip": "10.0.1.12",
			"tags": map[string]string{
				"Name":        "web-server-1",
				"Environment": "production",
				"Role":        "webserver",
			},
		},
		{
			"id":         "i-0987654321fedcba0",
			"public_ip":  "203.0.113.13",
			"private_ip": "10.0.1.13",
			"tags": map[string]string{
				"Name":        "db-server-1",
				"Environment": "production",
				"Role":        "database",
			},
		},
	}

	for _, instance := range exampleInstances {
		tags := instance["tags"].(map[string]string)
		instanceRole := tags["Role"]

		// Find appropriate group
		groupName := "ungrouped"
		if instanceRole == "webserver" {
			groupName = "webservers"
		} else if instanceRole == "database" {
			groupName = "databases"
		}

		// Ensure group exists
		if _, exists := inventory.Children[groupName]; !exists {
			inventory.Children[groupName] = Group{
				Hosts: make(map[string]Host),
				Vars:  make(map[string]interface{}),
			}
		}

		// Create host
		hostName := tags["Name"]
		host := Host{
			AnsibleHost:              instance["public_ip"].(string),
			AnsibleUser:              sshUser,
			AnsibleSSHPrivateKeyFile: sshKey,
			AnsiblePort:              sshPort,
			Vars: map[string]interface{}{
				"instance_id":    instance["id"],
				"private_ip":     instance["private_ip"],
				"environment":    tags["Environment"],
				"cloud_provider": "aws",
			},
		}

		// Add to group
		group := inventory.Children[groupName]
		group.Hosts[hostName] = host
		inventory.Children[groupName] = group

		fmt.Printf("✅ AWS 인스턴스 추가: %s (%s) -> %s\n", hostName, instance["public_ip"], groupName)
	}

	return nil
}

func discoverGCPInstances(inventory *Inventory) error {
	fmt.Printf("🔍 GCP 인스턴스 검색 중...\n")
	// Placeholder for GCP instance discovery
	return fmt.Errorf("GCP 인스턴스 검색은 아직 구현되지 않았습니다")
}

func discoverAzureInstances(inventory *Inventory) error {
	fmt.Printf("🔍 Azure 인스턴스 검색 중...\n")
	// Placeholder for Azure instance discovery
	return fmt.Errorf("Azure 인스턴스 검색은 아직 구현되지 않았습니다")
}

func writeInventoryFile(inventory *Inventory) error {
	switch inventoryFormat {
	case "yaml":
		return writeYAMLInventory(inventory)
	case "ini":
		return writeINIInventory(inventory)
	default:
		return fmt.Errorf("지원하지 않는 형식: %s", inventoryFormat)
	}
}

func writeYAMLInventory(inventory *Inventory) error {
	filename := inventoryFile
	if !strings.HasSuffix(filename, ".yml") && !strings.HasSuffix(filename, ".yaml") {
		filename += ".yml"
	}

	data, err := yaml.Marshal(inventory)
	if err != nil {
		return err
	}

	fmt.Printf("📝 YAML 인벤토리 파일 생성: %s\n", filename)
	return os.WriteFile(filename, data, 0o644)
}

func writeINIInventory(inventory *Inventory) error {
	filename := inventoryFile
	if !strings.HasSuffix(filename, ".ini") {
		filename += ".ini"
	}

	var content strings.Builder

	// Write groups and hosts
	for groupName, group := range inventory.Children {
		content.WriteString(fmt.Sprintf("[%s]\n", groupName))

		for hostName, host := range group.Hosts {
			hostLine := hostName
			if host.AnsibleHost != "" && host.AnsibleHost != hostName {
				hostLine += fmt.Sprintf(" ansible_host=%s", host.AnsibleHost)
			}
			if host.AnsibleUser != "" {
				hostLine += fmt.Sprintf(" ansible_user=%s", host.AnsibleUser)
			}
			if host.AnsibleSSHPrivateKeyFile != "" {
				hostLine += fmt.Sprintf(" ansible_ssh_private_key_file=%s", host.AnsibleSSHPrivateKeyFile)
			}
			if host.AnsiblePort != 0 && host.AnsiblePort != 22 {
				hostLine += fmt.Sprintf(" ansible_port=%d", host.AnsiblePort)
			}

			// Add custom variables
			for key, value := range host.Vars {
				hostLine += fmt.Sprintf(" %s=%v", key, value)
			}

			content.WriteString(hostLine + "\n")
		}
		content.WriteString("\n")

		// Write group variables
		if len(group.Vars) > 0 {
			content.WriteString(fmt.Sprintf("[%s:vars]\n", groupName))
			for key, value := range group.Vars {
				content.WriteString(fmt.Sprintf("%s=%v\n", key, value))
			}
			content.WriteString("\n")
		}
	}

	// Write all group variables
	if len(inventory.All.Vars) > 0 {
		content.WriteString("[all:vars]\n")
		for key, value := range inventory.All.Vars {
			content.WriteString(fmt.Sprintf("%s=%v\n", key, value))
		}
		content.WriteString("\n")
	}

	fmt.Printf("📝 INI 인벤토리 파일 생성: %s\n", filename)
	return os.WriteFile(filename, []byte(content.String()), 0o644)
}

func createDynamicInventory() error {
	fmt.Printf("⚡ 동적 인벤토리 스크립트 생성\n")

	dynamicConfig := DynamicInventory{
		Provider: cloudProvider,
		Region:   cloudRegion,
		Filters:  make(map[string]interface{}),
		Groups:   make(map[string]string),
	}

	// Parse cloud tags as filters
	for _, tag := range cloudTags {
		parts := strings.SplitN(tag, "=", 2)
		if len(parts) == 2 {
			dynamicConfig.Filters[fmt.Sprintf("tag:%s", parts[0])] = parts[1]
		}
	}

	// Set up group mappings
	dynamicConfig.Groups["webservers"] = "tag:Role=webserver"
	dynamicConfig.Groups["databases"] = "tag:Role=database"

	// Write dynamic inventory configuration
	configFile := inventoryFile + "_dynamic.yml"
	data, err := yaml.Marshal(dynamicConfig)
	if err != nil {
		return err
	}

	if err := os.WriteFile(configFile, data, 0o644); err != nil {
		return err
	}

	// Create dynamic inventory script
	scriptContent := generateDynamicInventoryScript(cloudProvider)
	scriptFile := inventoryFile + "_dynamic.py"

	if err := os.WriteFile(scriptFile, []byte(scriptContent), 0o755); err != nil {
		return err
	}

	fmt.Printf("✅ 동적 인벤토리 생성 완료\n")
	fmt.Printf("📄 설정 파일: %s\n", configFile)
	fmt.Printf("🐍 스크립트 파일: %s\n", scriptFile)

	return nil
}

func generateDynamicInventoryScript(provider string) string {
	switch provider {
	case "aws":
		return `#!/usr/bin/env python3
"""
AWS EC2 동적 인벤토리 스크립트
"""
import json
import boto3
import sys
from botocore.exceptions import BotoCoreError, ClientError

def get_ec2_inventory():
    try:
        ec2 = boto3.client('ec2')
        response = ec2.describe_instances(
            Filters=[
                {'Name': 'instance-state-name', 'Values': ['running']}
            ]
        )
        
        inventory = {
            '_meta': {'hostvars': {}},
            'all': {'children': []},
            'ungrouped': {'hosts': []}
        }
        
        for reservation in response['Reservations']:
            for instance in reservation['Instances']:
                instance_id = instance['InstanceId']
                public_ip = instance.get('PublicIpAddress', '')
                private_ip = instance.get('PrivateIpAddress', '')
                
                # Get instance tags
                tags = {tag['Key']: tag['Value'] for tag in instance.get('Tags', [])}
                instance_name = tags.get('Name', instance_id)
                
                # Determine group based on Role tag
                role = tags.get('Role', 'ungrouped')
                if role == 'webserver':
                    group_name = 'webservers'
                elif role == 'database':
                    group_name = 'databases'
                else:
                    group_name = 'ungrouped'
                
                # Ensure group exists
                if group_name not in inventory:
                    inventory[group_name] = {'hosts': []}
                    if group_name != 'ungrouped':
                        inventory['all']['children'].append(group_name)
                
                # Add host to group
                inventory[group_name]['hosts'].append(instance_name)
                
                # Add host variables
                inventory['_meta']['hostvars'][instance_name] = {
                    'ansible_host': public_ip or private_ip,
                    'ansible_user': 'ubuntu',
                    'instance_id': instance_id,
                    'private_ip': private_ip,
                    'public_ip': public_ip,
                    'instance_type': instance['InstanceType'],
                    'availability_zone': instance['Placement']['AvailabilityZone'],
                    'tags': tags
                }
        
        return inventory
        
    except (BotoCoreError, ClientError) as e:
        print(f"AWS 오류: {e}", file=sys.stderr)
        return {}

if __name__ == '__main__':
    if '--list' in sys.argv:
        print(json.dumps(get_ec2_inventory(), indent=2))
    elif '--host' in sys.argv:
        print(json.dumps({}))
    else:
        print("Usage: {} --list | --host <hostname>".format(sys.argv[0]))
        sys.exit(1)
`
	default:
		return fmt.Sprintf(`#!/usr/bin/env python3
"""
%s 동적 인벤토리 스크립트 (미구현)
"""
import json
import sys

def get_inventory():
    return {
        '_meta': {'hostvars': {}},
        'all': {'children': []},
        'ungrouped': {'hosts': []}
    }

if __name__ == '__main__':
    if '--list' in sys.argv:
        print(json.dumps(get_inventory(), indent=2))
    elif '--host' in sys.argv:
        print(json.dumps({}))
    else:
        print("Usage: {} --list | --host <hostname>".format(sys.argv[0]))
        sys.exit(1)
`, strings.Title(provider))
	}
}

func addHostToInventory() error {
	if hostAddress == "" || hostGroup == "" {
		return fmt.Errorf("호스트 주소와 그룹이 필요합니다")
	}

	fmt.Printf("➕ 호스트 추가: %s -> %s\n", hostAddress, hostGroup)

	// Read existing inventory
	inventory, err := readInventoryFile()
	if err != nil {
		return err
	}

	// Ensure group exists
	if _, exists := inventory.Children[hostGroup]; !exists {
		inventory.Children[hostGroup] = Group{
			Hosts: make(map[string]Host),
			Vars:  make(map[string]interface{}),
		}
	}

	// Create host
	host := Host{
		AnsibleHost:              hostAddress,
		AnsibleUser:              sshUser,
		AnsibleSSHPrivateKeyFile: sshKey,
		AnsiblePort:              sshPort,
		Vars:                     make(map[string]interface{}),
	}

	// Parse host variables
	for _, hostVar := range hostVars {
		parts := strings.SplitN(hostVar, "=", 2)
		if len(parts) == 2 {
			host.Vars[parts[0]] = parts[1]
		}
	}

	// Add to group
	group := inventory.Children[hostGroup]
	group.Hosts[hostAddress] = host
	inventory.Children[hostGroup] = group

	// Write back to file
	return writeInventoryFile(inventory)
}

func removeHostFromInventory() error {
	if hostAddress == "" {
		return fmt.Errorf("제거할 호스트 주소가 필요합니다")
	}

	fmt.Printf("➖ 호스트 제거: %s\n", hostAddress)

	// Read existing inventory
	inventory, err := readInventoryFile()
	if err != nil {
		return err
	}

	// Remove host from all groups
	found := false
	for groupName, group := range inventory.Children {
		if _, exists := group.Hosts[hostAddress]; exists {
			delete(group.Hosts, hostAddress)
			inventory.Children[groupName] = group
			found = true
			fmt.Printf("✅ %s 그룹에서 호스트 제거됨\n", groupName)
		}
	}

	if !found {
		return fmt.Errorf("호스트를 찾을 수 없습니다: %s", hostAddress)
	}

	// Write back to file
	return writeInventoryFile(inventory)
}

func testInventoryConnection() error {
	fmt.Printf("🔗 인벤토리 연결 테스트\n")

	// Read inventory
	inventory, err := readInventoryFile()
	if err != nil {
		return err
	}

	// Test connections
	for groupName, group := range inventory.Children {
		fmt.Printf("\n📋 그룹: %s\n", groupName)

		for hostName, host := range group.Hosts {
			fmt.Printf("  🖥️  %s (%s) ... ", hostName, host.AnsibleHost)

			if testSSHConnection(host.AnsibleHost, host.AnsibleUser, host.AnsibleSSHPrivateKeyFile, host.AnsiblePort) {
				fmt.Printf("✅ 연결 성공\n")
			} else {
				fmt.Printf("❌ 연결 실패\n")
			}
		}
	}

	return nil
}

func testSSHConnection(host, user, keyFile string, port int) bool {
	if port == 0 {
		port = 22
	}

	// Simple TCP connection test
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), 5*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func listInventory() error {
	fmt.Printf("📋 인벤토리 목록\n")

	// Read inventory
	inventory, err := readInventoryFile()
	if err != nil {
		return err
	}

	// Display inventory structure
	fmt.Printf("\n🗂️  그룹 및 호스트:\n")
	for groupName, group := range inventory.Children {
		fmt.Printf("\n📁 %s (%d 호스트)\n", groupName, len(group.Hosts))

		for hostName, host := range group.Hosts {
			fmt.Printf("  🖥️  %s", hostName)
			if host.AnsibleHost != "" && host.AnsibleHost != hostName {
				fmt.Printf(" -> %s", host.AnsibleHost)
			}
			fmt.Printf("\n")
		}

		// Show group variables
		if len(group.Vars) > 0 {
			fmt.Printf("  📝 그룹 변수:\n")
			for key, value := range group.Vars {
				fmt.Printf("    %s: %v\n", key, value)
			}
		}
	}

	// Show all variables
	if len(inventory.All.Vars) > 0 {
		fmt.Printf("\n🌐 전역 변수:\n")
		for key, value := range inventory.All.Vars {
			fmt.Printf("  %s: %v\n", key, value)
		}
	}

	return nil
}

func readInventoryFile() (*Inventory, error) {
	filename := inventoryFile

	// Try different extensions
	possibleFiles := []string{
		filename,
		filename + ".yml",
		filename + ".yaml",
		filename + ".ini",
	}

	var data []byte
	var err error
	var format string

	for _, file := range possibleFiles {
		data, err = os.ReadFile(file)
		if err == nil {
			if strings.HasSuffix(file, ".yml") || strings.HasSuffix(file, ".yaml") {
				format = "yaml"
			} else {
				format = "ini"
			}
			filename = file
			break
		}
	}

	if err != nil {
		return nil, fmt.Errorf("인벤토리 파일을 찾을 수 없습니다: %s", filename)
	}

	inventory := &Inventory{
		Children: make(map[string]Group),
	}

	if format == "yaml" {
		if err := yaml.Unmarshal(data, inventory); err != nil {
			return nil, fmt.Errorf("YAML 파싱 오류: %w", err)
		}
	} else {
		if err := parseINIInventory(string(data), inventory); err != nil {
			return nil, fmt.Errorf("INI 파싱 오류: %w", err)
		}
	}

	return inventory, nil
}

func parseINIInventory(content string, inventory *Inventory) error {
	scanner := bufio.NewScanner(strings.NewReader(content))

	var currentGroup string
	var inVarsSection bool

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		// Group header
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			groupHeader := line[1 : len(line)-1]

			if strings.HasSuffix(groupHeader, ":vars") {
				currentGroup = strings.TrimSuffix(groupHeader, ":vars")
				inVarsSection = true
			} else {
				currentGroup = groupHeader
				inVarsSection = false

				// Create group if it doesn't exist
				if _, exists := inventory.Children[currentGroup]; !exists {
					inventory.Children[currentGroup] = Group{
						Hosts: make(map[string]Host),
						Vars:  make(map[string]interface{}),
					}
				}
			}
			continue
		}

		if currentGroup == "" {
			continue
		}

		if inVarsSection {
			// Parse variable line
			if parts := strings.SplitN(line, "=", 2); len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])

				if currentGroup == "all" {
					if inventory.All.Vars == nil {
						inventory.All.Vars = make(map[string]interface{})
					}
					inventory.All.Vars[key] = value
				} else {
					group := inventory.Children[currentGroup]
					group.Vars[key] = value
					inventory.Children[currentGroup] = group
				}
			}
		} else {
			// Parse host line
			parts := strings.Fields(line)
			if len(parts) == 0 {
				continue
			}

			hostName := parts[0]
			host := Host{
				Vars: make(map[string]interface{}),
			}

			// Parse host variables
			for _, part := range parts[1:] {
				if keyValue := strings.SplitN(part, "=", 2); len(keyValue) == 2 {
					key := keyValue[0]
					value := keyValue[1]

					switch key {
					case "ansible_host":
						host.AnsibleHost = value
					case "ansible_user":
						host.AnsibleUser = value
					case "ansible_ssh_private_key_file":
						host.AnsibleSSHPrivateKeyFile = value
					case "ansible_port":
						if port := parseInt(value); port > 0 {
							host.AnsiblePort = port
						}
					default:
						host.Vars[key] = value
					}
				}
			}

			// Add host to group
			group := inventory.Children[currentGroup]
			group.Hosts[hostName] = host
			inventory.Children[currentGroup] = group
		}
	}

	return scanner.Err()
}

func parseInt(s string) int {
	var result int
	fmt.Sscanf(s, "%d", &result)
	return result
}
