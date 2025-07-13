package ansible

import (
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
	Short: "Ansible 플레이북 및 역할 생성",
	Long: `Ansible 플레이북, 역할, 템플릿을 자동 생성합니다.

다양한 서버 설정을 위한 플레이북 생성:
- 웹 서버 설정 (Apache, Nginx)
- 데이터베이스 설정 (MySQL, PostgreSQL, MongoDB)
- 애플리케이션 배포 (Docker, Node.js, Python)
- 보안 설정 (방화벽, SSL, 사용자 관리)
- 모니터링 설정 (Prometheus, Grafana)
- 시스템 기본 설정 (패키지, 서비스, 구성 파일)

Examples:
  gz ansible generate --type webserver --target nginx
  gz ansible generate --type database --target mysql --environment production
  gz ansible generate --type application --target docker --with-vault
  gz ansible generate --role common --tasks user-management,security`,
	Run: runGenerate,
}

var (
	generateType       string
	generateTarget     string
	generateRole       string
	generateTasks      []string
	generateVars       []string
	genEnvironment     string
	outputPath         string
	withVault          bool
	withHandlers       bool
	withTemplates      bool
	withDefaults       bool
	genInventoryGroups []string
	playbookName       string
	description        string
)

func init() {
	GenerateCmd.Flags().StringVarP(&generateType, "type", "t", "", "플레이북 타입 (webserver, database, application, security, monitoring)")
	GenerateCmd.Flags().StringVar(&generateTarget, "target", "", "대상 기술 (nginx, apache, mysql, postgresql, docker, etc.)")
	GenerateCmd.Flags().StringVarP(&generateRole, "role", "r", "", "역할 이름")
	GenerateCmd.Flags().StringSliceVar(&generateTasks, "tasks", []string{}, "생성할 태스크 목록")
	GenerateCmd.Flags().StringSliceVar(&generateVars, "vars", []string{}, "변수 정의 (key=value)")
	GenerateCmd.Flags().StringVarP(&genEnvironment, "environment", "e", "development", "대상 환경")
	GenerateCmd.Flags().StringVarP(&outputPath, "output", "o", ".", "출력 디렉터리")
	GenerateCmd.Flags().BoolVar(&withVault, "with-vault", false, "Ansible Vault 암호화 활성화")
	GenerateCmd.Flags().BoolVar(&withHandlers, "with-handlers", true, "핸들러 생성")
	GenerateCmd.Flags().BoolVar(&withTemplates, "with-templates", true, "템플릿 파일 생성")
	GenerateCmd.Flags().BoolVar(&withDefaults, "with-defaults", true, "기본값 파일 생성")
	GenerateCmd.Flags().StringSliceVar(&genInventoryGroups, "inventory-groups", []string{"webservers"}, "인벤토리 그룹")
	GenerateCmd.Flags().StringVarP(&playbookName, "name", "n", "", "플레이북 이름")
	GenerateCmd.Flags().StringVarP(&description, "description", "d", "", "플레이북 설명")
}

// Playbook represents an Ansible playbook structure
type Playbook struct {
	Name       string                 `yaml:"name"`
	Hosts      interface{}            `yaml:"hosts"`
	BecomeUser string                 `yaml:"become_user,omitempty"`
	Become     bool                   `yaml:"become,omitempty"`
	Vars       map[string]interface{} `yaml:"vars,omitempty"`
	VarsFiles  []string               `yaml:"vars_files,omitempty"`
	Tasks      []Task                 `yaml:"tasks,omitempty"`
	Roles      []interface{}          `yaml:"roles,omitempty"`
	Handlers   []Handler              `yaml:"handlers,omitempty"`
	PreTasks   []Task                 `yaml:"pre_tasks,omitempty"`
	PostTasks  []Task                 `yaml:"post_tasks,omitempty"`
}

// Task represents an Ansible task
type Task struct {
	Name         string                 `yaml:"name"`
	Module       string                 `yaml:",inline"`
	Params       map[string]interface{} `yaml:",inline"`
	When         string                 `yaml:"when,omitempty"`
	Loop         interface{}            `yaml:"loop,omitempty"`
	Register     string                 `yaml:"register,omitempty"`
	Notify       []string               `yaml:"notify,omitempty"`
	Tags         []string               `yaml:"tags,omitempty"`
	Become       bool                   `yaml:"become,omitempty"`
	BecomeUser   string                 `yaml:"become_user,omitempty"`
	IgnoreErrors bool                   `yaml:"ignore_errors,omitempty"`
}

// Handler represents an Ansible handler
type Handler struct {
	Name   string                 `yaml:"name"`
	Module string                 `yaml:",inline"`
	Params map[string]interface{} `yaml:",inline"`
	Listen string                 `yaml:"listen,omitempty"`
}

// Role represents an Ansible role structure
type Role struct {
	Name         string
	Description  string
	Dependencies []string
	Tasks        []Task
	Handlers     []Handler
	Vars         map[string]interface{}
	Defaults     map[string]interface{}
	Templates    map[string]string
	Files        map[string]string
}

func runGenerate(cmd *cobra.Command, args []string) {
	fmt.Printf("🎭 Ansible 플레이북 생성기\n")

	if generateRole != "" {
		fmt.Printf("📋 역할 생성: %s\n", generateRole)
		if err := generateAnsibleRole(); err != nil {
			fmt.Printf("❌ 역할 생성 실패: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if generateType == "" {
		fmt.Printf("❌ 플레이북 타입이 필요합니다 (--type)\n")
		cmd.Help()
		os.Exit(1)
	}

	fmt.Printf("📦 타입: %s\n", generateType)
	if generateTarget != "" {
		fmt.Printf("🎯 대상: %s\n", generateTarget)
	}
	fmt.Printf("🌍 환경: %s\n", genEnvironment)

	// Create output directory
	if err := os.MkdirAll(outputPath, 0o755); err != nil {
		fmt.Printf("❌ 출력 디렉터리 생성 실패: %v\n", err)
		os.Exit(1)
	}

	// Generate playbook
	if err := generatePlaybook(); err != nil {
		fmt.Printf("❌ 플레이북 생성 실패: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Ansible 플레이북 생성 완료\n")
	fmt.Printf("📁 출력 경로: %s\n", outputPath)
	fmt.Printf("\n📝 다음 단계:\n")
	fmt.Printf("1. 인벤토리 파일 생성: gz ansible inventory create\n")
	fmt.Printf("2. 플레이북 구문 확인: ansible-playbook --syntax-check playbook.yml\n")
	fmt.Printf("3. 플레이북 실행: ansible-playbook -i inventory playbook.yml\n")
}

func generatePlaybook() error {
	// Parse variables
	vars := parseVariables()

	// Create playbook based on type
	playbook, err := createPlaybookByType(vars)
	if err != nil {
		return err
	}

	// Set playbook name
	if playbookName == "" {
		playbookName = fmt.Sprintf("%s-%s", generateType, genEnvironment)
	}

	// Write main playbook
	playbookFile := filepath.Join(outputPath, "playbook.yml")
	if err := writePlaybook(playbook, playbookFile); err != nil {
		return err
	}

	// Generate supporting files
	if err := generateSupportingFiles(vars); err != nil {
		return err
	}

	// Generate vault files if requested
	if withVault {
		if err := generateVaultFiles(vars); err != nil {
			return err
		}
	}

	return nil
}

func parseVariables() map[string]interface{} {
	vars := make(map[string]interface{})

	for _, v := range generateVars {
		parts := strings.SplitN(v, "=", 2)
		if len(parts) == 2 {
			vars[parts[0]] = parts[1]
		}
	}

	// Add environment-specific defaults
	vars["environment"] = genEnvironment
	vars["ansible_user"] = "{{ ansible_user | default('ubuntu') }}"
	vars["become"] = true

	return vars
}

func createPlaybookByType(vars map[string]interface{}) (*Playbook, error) {
	playbook := &Playbook{
		Name:      fmt.Sprintf("%s 설정 플레이북", strings.Title(generateType)),
		Hosts:     genInventoryGroups,
		Become:    true,
		Vars:      vars,
		VarsFiles: []string{fmt.Sprintf("vars/%s.yml", genEnvironment)},
	}

	if description != "" {
		playbook.Name = description
	}

	switch generateType {
	case "webserver":
		return createWebServerPlaybook(playbook)
	case "database":
		return createDatabasePlaybook(playbook)
	case "application":
		return createApplicationPlaybook(playbook)
	case "security":
		return createSecurityPlaybook(playbook)
	case "monitoring":
		return createMonitoringPlaybook(playbook)
	default:
		return createGenericPlaybook(playbook)
	}
}

func createWebServerPlaybook(playbook *Playbook) (*Playbook, error) {
	target := generateTarget
	if target == "" {
		target = "nginx"
	}

	switch target {
	case "nginx":
		playbook.Tasks = []Task{
			{
				Name:   "시스템 패키지 업데이트",
				Module: "apt",
				Params: map[string]interface{}{
					"update_cache":     true,
					"cache_valid_time": 3600,
				},
				Tags: []string{"packages"},
			},
			{
				Name:   "Nginx 설치",
				Module: "apt",
				Params: map[string]interface{}{
					"name":  "nginx",
					"state": "present",
				},
				Tags:   []string{"nginx", "install"},
				Notify: []string{"restart nginx"},
			},
			{
				Name:   "Nginx 설정 파일 복사",
				Module: "template",
				Params: map[string]interface{}{
					"src":    "nginx.conf.j2",
					"dest":   "/etc/nginx/nginx.conf",
					"backup": true,
				},
				Tags:   []string{"nginx", "config"},
				Notify: []string{"restart nginx"},
			},
			{
				Name:   "사이트 설정 복사",
				Module: "template",
				Params: map[string]interface{}{
					"src":  "site.conf.j2",
					"dest": "/etc/nginx/sites-available/{{ site_name | default('default') }}",
				},
				Tags:   []string{"nginx", "sites"},
				Notify: []string{"restart nginx"},
			},
			{
				Name:   "사이트 활성화",
				Module: "file",
				Params: map[string]interface{}{
					"src":   "/etc/nginx/sites-available/{{ site_name | default('default') }}",
					"dest":  "/etc/nginx/sites-enabled/{{ site_name | default('default') }}",
					"state": "link",
				},
				Tags:   []string{"nginx", "sites"},
				Notify: []string{"restart nginx"},
			},
			{
				Name:   "기본 사이트 비활성화",
				Module: "file",
				Params: map[string]interface{}{
					"path":  "/etc/nginx/sites-enabled/default",
					"state": "absent",
				},
				Tags: []string{"nginx", "cleanup"},
				When: "site_name is defined",
			},
			{
				Name:   "방화벽에서 HTTP/HTTPS 허용",
				Module: "ufw",
				Params: map[string]interface{}{
					"rule":  "allow",
					"port":  "{{ item }}",
					"proto": "tcp",
				},
				Loop: []string{"80", "443"},
				Tags: []string{"firewall"},
			},
			{
				Name:   "Nginx 서비스 시작 및 활성화",
				Module: "systemd",
				Params: map[string]interface{}{
					"name":    "nginx",
					"state":   "started",
					"enabled": true,
				},
				Tags: []string{"nginx", "service"},
			},
		}

		if withHandlers {
			playbook.Handlers = []Handler{
				{
					Name:   "restart nginx",
					Module: "systemd",
					Params: map[string]interface{}{
						"name":  "nginx",
						"state": "restarted",
					},
				},
				{
					Name:   "reload nginx",
					Module: "systemd",
					Params: map[string]interface{}{
						"name":  "nginx",
						"state": "reloaded",
					},
				},
			}
		}

	case "apache":
		playbook.Tasks = []Task{
			{
				Name:   "Apache 설치",
				Module: "apt",
				Params: map[string]interface{}{
					"name":  "apache2",
					"state": "present",
				},
				Tags:   []string{"apache", "install"},
				Notify: []string{"restart apache"},
			},
			{
				Name:   "Apache 모듈 활성화",
				Module: "apache2_module",
				Params: map[string]interface{}{
					"name":  "{{ item }}",
					"state": "present",
				},
				Loop:   []string{"rewrite", "ssl", "headers"},
				Tags:   []string{"apache", "modules"},
				Notify: []string{"restart apache"},
			},
			{
				Name:   "가상 호스트 설정",
				Module: "template",
				Params: map[string]interface{}{
					"src":  "vhost.conf.j2",
					"dest": "/etc/apache2/sites-available/{{ site_name | default('000-default') }}.conf",
				},
				Tags:   []string{"apache", "vhost"},
				Notify: []string{"restart apache"},
			},
			{
				Name:   "Apache 서비스 시작",
				Module: "systemd",
				Params: map[string]interface{}{
					"name":    "apache2",
					"state":   "started",
					"enabled": true,
				},
				Tags: []string{"apache", "service"},
			},
		}
	}

	return playbook, nil
}

func createDatabasePlaybook(playbook *Playbook) (*Playbook, error) {
	target := generateTarget
	if target == "" {
		target = "mysql"
	}

	switch target {
	case "mysql":
		playbook.Tasks = []Task{
			{
				Name:   "MySQL 서버 설치",
				Module: "apt",
				Params: map[string]interface{}{
					"name":  []string{"mysql-server", "mysql-client", "python3-pymysql"},
					"state": "present",
				},
				Tags: []string{"mysql", "install"},
			},
			{
				Name:   "MySQL 보안 설정",
				Module: "mysql_user",
				Params: map[string]interface{}{
					"name":              "root",
					"password":          "{{ mysql_root_password }}",
					"login_unix_socket": "/var/run/mysqld/mysqld.sock",
					"host":              "localhost",
					"state":             "present",
				},
				Tags: []string{"mysql", "security"},
			},
			{
				Name:   "데이터베이스 생성",
				Module: "mysql_db",
				Params: map[string]interface{}{
					"name":           "{{ mysql_database }}",
					"state":          "present",
					"login_user":     "root",
					"login_password": "{{ mysql_root_password }}",
				},
				Tags: []string{"mysql", "database"},
				When: "mysql_database is defined",
			},
			{
				Name:   "데이터베이스 사용자 생성",
				Module: "mysql_user",
				Params: map[string]interface{}{
					"name":           "{{ mysql_user }}",
					"password":       "{{ mysql_password }}",
					"priv":           "{{ mysql_database }}.*:ALL",
					"state":          "present",
					"login_user":     "root",
					"login_password": "{{ mysql_root_password }}",
				},
				Tags: []string{"mysql", "user"},
				When: "mysql_user is defined",
			},
		}

	case "postgresql":
		playbook.Tasks = []Task{
			{
				Name:   "PostgreSQL 설치",
				Module: "apt",
				Params: map[string]interface{}{
					"name":  []string{"postgresql", "postgresql-contrib", "python3-psycopg2"},
					"state": "present",
				},
				Tags: []string{"postgresql", "install"},
			},
			{
				Name:   "PostgreSQL 서비스 시작",
				Module: "systemd",
				Params: map[string]interface{}{
					"name":    "postgresql",
					"state":   "started",
					"enabled": true,
				},
				Tags: []string{"postgresql", "service"},
			},
			{
				Name:   "데이터베이스 생성",
				Module: "postgresql_db",
				Params: map[string]interface{}{
					"name":  "{{ postgres_database }}",
					"state": "present",
				},
				Become:     true,
				BecomeUser: "postgres",
				Tags:       []string{"postgresql", "database"},
				When:       "postgres_database is defined",
			},
		}
	}

	return playbook, nil
}

func createApplicationPlaybook(playbook *Playbook) (*Playbook, error) {
	target := generateTarget
	if target == "" {
		target = "docker"
	}

	switch target {
	case "docker":
		playbook.Tasks = []Task{
			{
				Name:   "Docker 의존성 패키지 설치",
				Module: "apt",
				Params: map[string]interface{}{
					"name":  []string{"apt-transport-https", "ca-certificates", "curl", "gnupg", "lsb-release"},
					"state": "present",
				},
				Tags: []string{"docker", "dependencies"},
			},
			{
				Name:   "Docker GPG 키 추가",
				Module: "apt_key",
				Params: map[string]interface{}{
					"url":   "https://download.docker.com/linux/ubuntu/gpg",
					"state": "present",
				},
				Tags: []string{"docker", "repository"},
			},
			{
				Name:   "Docker 저장소 추가",
				Module: "apt_repository",
				Params: map[string]interface{}{
					"repo":  "deb [arch=amd64] https://download.docker.com/linux/ubuntu {{ ansible_distribution_release }} stable",
					"state": "present",
				},
				Tags: []string{"docker", "repository"},
			},
			{
				Name:   "Docker CE 설치",
				Module: "apt",
				Params: map[string]interface{}{
					"name":         []string{"docker-ce", "docker-ce-cli", "containerd.io", "docker-compose-plugin"},
					"state":        "present",
					"update_cache": true,
				},
				Tags: []string{"docker", "install"},
			},
			{
				Name:   "사용자를 docker 그룹에 추가",
				Module: "user",
				Params: map[string]interface{}{
					"name":   "{{ ansible_user }}",
					"groups": "docker",
					"append": true,
				},
				Tags: []string{"docker", "user"},
			},
			{
				Name:   "Docker 서비스 시작",
				Module: "systemd",
				Params: map[string]interface{}{
					"name":    "docker",
					"state":   "started",
					"enabled": true,
				},
				Tags: []string{"docker", "service"},
			},
		}

	case "nodejs":
		playbook.Tasks = []Task{
			{
				Name:   "NodeSource GPG 키 추가",
				Module: "apt_key",
				Params: map[string]interface{}{
					"url":   "https://deb.nodesource.com/gpgkey/nodesource.gpg.key",
					"state": "present",
				},
				Tags: []string{"nodejs", "repository"},
			},
			{
				Name:   "NodeSource 저장소 추가",
				Module: "apt_repository",
				Params: map[string]interface{}{
					"repo":  "deb https://deb.nodesource.com/node_{{ nodejs_version | default('18') }}.x {{ ansible_distribution_release }} main",
					"state": "present",
				},
				Tags: []string{"nodejs", "repository"},
			},
			{
				Name:   "Node.js 설치",
				Module: "apt",
				Params: map[string]interface{}{
					"name":         "nodejs",
					"state":        "present",
					"update_cache": true,
				},
				Tags: []string{"nodejs", "install"},
			},
		}
	}

	return playbook, nil
}

func createSecurityPlaybook(playbook *Playbook) (*Playbook, error) {
	playbook.Tasks = []Task{
		{
			Name:   "시스템 패키지 업데이트",
			Module: "apt",
			Params: map[string]interface{}{
				"upgrade":          "yes",
				"update_cache":     true,
				"cache_valid_time": 3600,
			},
			Tags: []string{"security", "updates"},
		},
		{
			Name:   "보안 패키지 설치",
			Module: "apt",
			Params: map[string]interface{}{
				"name":  []string{"ufw", "fail2ban", "unattended-upgrades"},
				"state": "present",
			},
			Tags: []string{"security", "packages"},
		},
		{
			Name:   "UFW 방화벽 기본 정책 설정",
			Module: "ufw",
			Params: map[string]interface{}{
				"direction": "{{ item.direction }}",
				"policy":    "{{ item.policy }}",
			},
			Loop: []map[string]string{
				{"direction": "incoming", "policy": "deny"},
				{"direction": "outgoing", "policy": "allow"},
			},
			Tags: []string{"security", "firewall"},
		},
		{
			Name:   "SSH 포트 허용",
			Module: "ufw",
			Params: map[string]interface{}{
				"rule":  "allow",
				"port":  "{{ ssh_port | default('22') }}",
				"proto": "tcp",
			},
			Tags: []string{"security", "ssh"},
		},
		{
			Name:   "방화벽 활성화",
			Module: "ufw",
			Params: map[string]interface{}{
				"state": "enabled",
			},
			Tags: []string{"security", "firewall"},
		},
		{
			Name:   "SSH 보안 설정",
			Module: "lineinfile",
			Params: map[string]interface{}{
				"path":   "/etc/ssh/sshd_config",
				"regexp": "^{{ item.regexp }}",
				"line":   "{{ item.line }}",
				"backup": true,
			},
			Loop: []map[string]string{
				{"regexp": "PermitRootLogin", "line": "PermitRootLogin no"},
				{"regexp": "PasswordAuthentication", "line": "PasswordAuthentication no"},
				{"regexp": "X11Forwarding", "line": "X11Forwarding no"},
			},
			Tags:   []string{"security", "ssh"},
			Notify: []string{"restart ssh"},
		},
		{
			Name:   "Fail2ban 서비스 시작",
			Module: "systemd",
			Params: map[string]interface{}{
				"name":    "fail2ban",
				"state":   "started",
				"enabled": true,
			},
			Tags: []string{"security", "fail2ban"},
		},
	}

	if withHandlers {
		playbook.Handlers = []Handler{
			{
				Name:   "restart ssh",
				Module: "systemd",
				Params: map[string]interface{}{
					"name":  "ssh",
					"state": "restarted",
				},
			},
		}
	}

	return playbook, nil
}

func createMonitoringPlaybook(playbook *Playbook) (*Playbook, error) {
	playbook.Tasks = []Task{
		{
			Name:   "모니터링 디렉터리 생성",
			Module: "file",
			Params: map[string]interface{}{
				"path":  "/opt/monitoring",
				"state": "directory",
				"mode":  "0755",
			},
			Tags: []string{"monitoring", "setup"},
		},
		{
			Name:   "Node Exporter 다운로드",
			Module: "get_url",
			Params: map[string]interface{}{
				"url":  "https://github.com/prometheus/node_exporter/releases/download/v{{ node_exporter_version | default('1.6.1') }}/node_exporter-{{ node_exporter_version | default('1.6.1') }}.linux-amd64.tar.gz",
				"dest": "/tmp/node_exporter.tar.gz",
			},
			Tags: []string{"monitoring", "node_exporter"},
		},
		{
			Name:   "Node Exporter 압축 해제",
			Module: "unarchive",
			Params: map[string]interface{}{
				"src":        "/tmp/node_exporter.tar.gz",
				"dest":       "/opt/monitoring/",
				"remote_src": true,
				"creates":    "/opt/monitoring/node_exporter",
				"extra_opts": []string{"--strip-components=1"},
			},
			Tags: []string{"monitoring", "node_exporter"},
		},
		{
			Name:   "Node Exporter 실행 권한 설정",
			Module: "file",
			Params: map[string]interface{}{
				"path": "/opt/monitoring/node_exporter",
				"mode": "0755",
			},
			Tags: []string{"monitoring", "node_exporter"},
		},
		{
			Name:   "Node Exporter 서비스 파일 생성",
			Module: "template",
			Params: map[string]interface{}{
				"src":  "node_exporter.service.j2",
				"dest": "/etc/systemd/system/node_exporter.service",
			},
			Tags:   []string{"monitoring", "service"},
			Notify: []string{"restart node_exporter"},
		},
		{
			Name:   "Node Exporter 서비스 시작",
			Module: "systemd",
			Params: map[string]interface{}{
				"name":          "node_exporter",
				"state":         "started",
				"enabled":       true,
				"daemon_reload": true,
			},
			Tags: []string{"monitoring", "service"},
		},
	}

	return playbook, nil
}

func createGenericPlaybook(playbook *Playbook) (*Playbook, error) {
	// Create basic tasks based on provided task list
	if len(generateTasks) > 0 {
		for _, taskName := range generateTasks {
			task := Task{
				Name:   fmt.Sprintf("Execute %s", taskName),
				Module: "debug",
				Params: map[string]interface{}{
					"msg": fmt.Sprintf("This is a placeholder for %s task", taskName),
				},
				Tags: []string{strings.ToLower(taskName)},
			}
			playbook.Tasks = append(playbook.Tasks, task)
		}
	} else {
		playbook.Tasks = []Task{
			{
				Name:   "기본 시스템 정보 수집",
				Module: "setup",
				Params: map[string]interface{}{},
				Tags:   []string{"facts"},
			},
			{
				Name:   "시스템 상태 확인",
				Module: "command",
				Params: map[string]interface{}{
					"cmd": "uptime",
				},
				Register: "system_uptime",
				Tags:     []string{"status"},
			},
			{
				Name:   "시스템 정보 출력",
				Module: "debug",
				Params: map[string]interface{}{
					"var": "system_uptime.stdout",
				},
				Tags: []string{"info"},
			},
		}
	}

	return playbook, nil
}

func writePlaybook(playbook *Playbook, filename string) error {
	data, err := yaml.Marshal([]Playbook{*playbook})
	if err != nil {
		return fmt.Errorf("YAML 마샬링 실패: %w", err)
	}

	return os.WriteFile(filename, data, 0o644)
}

func generateSupportingFiles(vars map[string]interface{}) error {
	// Create directory structure
	dirs := []string{
		"vars",
		"templates",
		"files",
		"group_vars",
		"host_vars",
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(filepath.Join(outputPath, dir), 0o755); err != nil {
			return err
		}
	}

	// Generate vars file
	varsFile := filepath.Join(outputPath, "vars", fmt.Sprintf("%s.yml", genEnvironment))
	if err := generateVarsFile(vars, varsFile); err != nil {
		return err
	}

	// Generate templates if requested
	if withTemplates {
		if err := generateTemplateFiles(); err != nil {
			return err
		}
	}

	// Generate group vars
	if err := generateGroupVars(); err != nil {
		return err
	}

	return nil
}

func generateVarsFile(vars map[string]interface{}, filename string) error {
	// Add type-specific variables
	switch generateType {
	case "webserver":
		if generateTarget == "nginx" {
			vars["nginx_user"] = "www-data"
			vars["nginx_worker_processes"] = "auto"
			vars["nginx_worker_connections"] = 1024
			vars["nginx_keepalive_timeout"] = 65
			vars["site_name"] = "example.com"
			vars["document_root"] = "/var/www/html"
		}
	case "database":
		if generateTarget == "mysql" {
			vars["mysql_root_password"] = "{{ vault_mysql_root_password }}"
			vars["mysql_database"] = "app_db"
			vars["mysql_user"] = "app_user"
			vars["mysql_password"] = "{{ vault_mysql_password }}"
		}
	case "monitoring":
		vars["node_exporter_version"] = "1.6.1"
		vars["node_exporter_port"] = 9100
	}

	data, err := yaml.Marshal(vars)
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0o644)
}

func generateTemplateFiles() error {
	templatesDir := filepath.Join(outputPath, "templates")

	switch generateType {
	case "webserver":
		if generateTarget == "nginx" {
			// Nginx main config template
			nginxConf := `user {{ nginx_user }};
worker_processes {{ nginx_worker_processes }};
pid /run/nginx.pid;

events {
    worker_connections {{ nginx_worker_connections }};
}

http {
    include /etc/nginx/mime.types;
    default_type application/octet-stream;
    
    sendfile on;
    keepalive_timeout {{ nginx_keepalive_timeout }};
    
    include /etc/nginx/sites-enabled/*;
}`
			if err := os.WriteFile(filepath.Join(templatesDir, "nginx.conf.j2"), []byte(nginxConf), 0o644); err != nil {
				return err
			}

			// Site config template
			siteConf := `server {
    listen 80;
    server_name {{ site_name }};
    root {{ document_root }};
    index index.html index.htm index.php;
    
    location / {
        try_files $uri $uri/ =404;
    }
    
    location ~ \.php$ {
        include snippets/fastcgi-php.conf;
        fastcgi_pass unix:/var/run/php/php8.1-fpm.sock;
    }
    
    location ~ /\.ht {
        deny all;
    }
}`
			if err := os.WriteFile(filepath.Join(templatesDir, "site.conf.j2"), []byte(siteConf), 0o644); err != nil {
				return err
			}
		}
	case "monitoring":
		// Node Exporter service template
		serviceTemplate := `[Unit]
Description=Node Exporter
After=network.target

[Service]
Type=simple
User=prometheus
ExecStart=/opt/monitoring/node_exporter --web.listen-address=:{{ node_exporter_port | default('9100') }}
Restart=always

[Install]
WantedBy=multi-user.target`
		if err := os.WriteFile(filepath.Join(templatesDir, "node_exporter.service.j2"), []byte(serviceTemplate), 0o644); err != nil {
			return err
		}
	}

	return nil
}

func generateGroupVars() error {
	groupVarsDir := filepath.Join(outputPath, "group_vars")

	for _, group := range genInventoryGroups {
		groupVars := map[string]interface{}{
			"ansible_user":                 "ubuntu",
			"ansible_ssh_private_key_file": "~/.ssh/id_rsa",
		}

		if group == "webservers" {
			groupVars["http_port"] = 80
			groupVars["https_port"] = 443
		} else if group == "databases" {
			groupVars["db_port"] = 3306
		}

		data, err := yaml.Marshal(groupVars)
		if err != nil {
			return err
		}

		filename := filepath.Join(groupVarsDir, fmt.Sprintf("%s.yml", group))
		if err := os.WriteFile(filename, data, 0o644); err != nil {
			return err
		}
	}

	return nil
}

func generateVaultFiles(vars map[string]interface{}) error {
	vaultDir := filepath.Join(outputPath, "vault")
	if err := os.MkdirAll(vaultDir, 0o755); err != nil {
		return err
	}

	// Generate vault variables based on type
	vaultVars := make(map[string]interface{})

	switch generateType {
	case "database":
		if generateTarget == "mysql" {
			vaultVars["vault_mysql_root_password"] = "change_me_mysql_root_password"
			vaultVars["vault_mysql_password"] = "change_me_mysql_user_password"
		} else if generateTarget == "postgresql" {
			vaultVars["vault_postgres_password"] = "change_me_postgres_password"
		}
	case "application":
		vaultVars["vault_app_secret_key"] = "change_me_secret_key"
		vaultVars["vault_api_token"] = "change_me_api_token"
	}

	if len(vaultVars) > 0 {
		data, err := yaml.Marshal(vaultVars)
		if err != nil {
			return err
		}

		vaultFile := filepath.Join(vaultDir, fmt.Sprintf("%s.yml", genEnvironment))
		if err := os.WriteFile(vaultFile, data, 0o644); err != nil {
			return err
		}

		// Create vault instructions
		instructions := fmt.Sprintf(`# Ansible Vault 사용 방법

## 1. Vault 파일 암호화
ansible-vault encrypt %s

## 2. Vault 파일 편집
ansible-vault edit %s

## 3. 플레이북 실행 시 Vault 패스워드 사용
ansible-playbook -i inventory playbook.yml --ask-vault-pass

## 4. Vault 패스워드 파일 사용
echo "your_vault_password" > .vault_pass
ansible-playbook -i inventory playbook.yml --vault-password-file .vault_pass

## 주의: .vault_pass 파일은 .gitignore에 추가하세요!
`, vaultFile, vaultFile)

		readmeFile := filepath.Join(vaultDir, "README.md")
		return os.WriteFile(readmeFile, []byte(instructions), 0o644)
	}

	return nil
}

func generateAnsibleRole() error {
	fmt.Printf("🏗️ 역할 생성: %s\n", generateRole)

	roleDir := filepath.Join(outputPath, "roles", generateRole)

	// Create role directory structure
	roleDirs := []string{
		"tasks",
		"handlers",
		"templates",
		"files",
		"vars",
		"defaults",
		"meta",
	}

	for _, dir := range roleDirs {
		if err := os.MkdirAll(filepath.Join(roleDir, dir), 0o755); err != nil {
			return err
		}
	}

	// Generate role metadata
	if err := generateRoleMeta(roleDir); err != nil {
		return err
	}

	// Generate tasks
	if err := generateRoleTasks(roleDir); err != nil {
		return err
	}

	// Generate defaults
	if err := generateRoleDefaults(roleDir); err != nil {
		return err
	}

	// Generate handlers if requested
	if withHandlers {
		if err := generateRoleHandlers(roleDir); err != nil {
			return err
		}
	}

	fmt.Printf("✅ 역할 생성 완료: %s\n", roleDir)
	return nil
}

func generateRoleMeta(roleDir string) error {
	meta := map[string]interface{}{
		"galaxy_info": map[string]interface{}{
			"author":              "Generated by gz-ansible",
			"description":         fmt.Sprintf("Ansible role for %s", generateRole),
			"license":             "MIT",
			"min_ansible_version": "2.9",
			"platforms": []map[string]interface{}{
				{
					"name":     "Ubuntu",
					"versions": []string{"20.04", "22.04"},
				},
			},
			"galaxy_tags": []string{generateRole, "automation"},
		},
		"dependencies": []string{},
	}

	data, err := yaml.Marshal(meta)
	if err != nil {
		return err
	}

	metaFile := filepath.Join(roleDir, "meta", "main.yml")
	return os.WriteFile(metaFile, data, 0o644)
}

func generateRoleTasks(roleDir string) error {
	var tasks []Task

	if len(generateTasks) > 0 {
		for _, taskName := range generateTasks {
			task := Task{
				Name:   fmt.Sprintf("Execute %s", strings.Title(taskName)),
				Module: "debug",
				Params: map[string]interface{}{
					"msg": fmt.Sprintf("Executing %s task", taskName),
				},
				Tags: []string{strings.ToLower(taskName)},
			}
			tasks = append(tasks, task)
		}
	} else {
		// Default tasks for common roles
		switch generateRole {
		case "common":
			tasks = []Task{
				{
					Name:   "시스템 패키지 업데이트",
					Module: "apt",
					Params: map[string]interface{}{
						"update_cache":     true,
						"cache_valid_time": 3600,
					},
					Tags: []string{"packages"},
				},
				{
					Name:   "필수 패키지 설치",
					Module: "apt",
					Params: map[string]interface{}{
						"name":  "{{ common_packages }}",
						"state": "present",
					},
					Tags: []string{"packages"},
				},
			}
		case "security":
			tasks = []Task{
				{
					Name:   "보안 업데이트 설치",
					Module: "apt",
					Params: map[string]interface{}{
						"upgrade":      "safe",
						"update_cache": true,
					},
					Tags: []string{"security"},
				},
				{
					Name:   "방화벽 설정",
					Module: "ufw",
					Params: map[string]interface{}{
						"state":     "enabled",
						"policy":    "deny",
						"direction": "incoming",
					},
					Tags: []string{"firewall"},
				},
			}
		default:
			tasks = []Task{
				{
					Name:   fmt.Sprintf("%s 역할 기본 작업", strings.Title(generateRole)),
					Module: "debug",
					Params: map[string]interface{}{
						"msg": fmt.Sprintf("This is the main task for %s role", generateRole),
					},
				},
			}
		}
	}

	data, err := yaml.Marshal(tasks)
	if err != nil {
		return err
	}

	tasksFile := filepath.Join(roleDir, "tasks", "main.yml")
	return os.WriteFile(tasksFile, data, 0o644)
}

func generateRoleDefaults(roleDir string) error {
	defaults := make(map[string]interface{})

	// Add role-specific defaults
	switch generateRole {
	case "common":
		defaults["common_packages"] = []string{
			"curl", "wget", "git", "vim", "htop", "tree",
		}
	case "security":
		defaults["ufw_rules"] = []map[string]interface{}{
			{"port": "22", "rule": "allow", "proto": "tcp"},
		}
	default:
		defaults[fmt.Sprintf("%s_enabled", generateRole)] = true
	}

	// Add user-defined variables
	for _, v := range generateVars {
		parts := strings.SplitN(v, "=", 2)
		if len(parts) == 2 {
			defaults[parts[0]] = parts[1]
		}
	}

	data, err := yaml.Marshal(defaults)
	if err != nil {
		return err
	}

	defaultsFile := filepath.Join(roleDir, "defaults", "main.yml")
	return os.WriteFile(defaultsFile, data, 0o644)
}

func generateRoleHandlers(roleDir string) error {
	handlers := []Handler{
		{
			Name:   "restart service",
			Module: "systemd",
			Params: map[string]interface{}{
				"name":  "{{ service_name }}",
				"state": "restarted",
			},
		},
		{
			Name:   "reload service",
			Module: "systemd",
			Params: map[string]interface{}{
				"name":  "{{ service_name }}",
				"state": "reloaded",
			},
		},
	}

	data, err := yaml.Marshal(handlers)
	if err != nil {
		return err
	}

	handlersFile := filepath.Join(roleDir, "handlers", "main.yml")
	return os.WriteFile(handlersFile, data, 0o644)
}
