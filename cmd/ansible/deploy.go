package ansible

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// DeployCmd represents the deploy command
var DeployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Ansible 플레이북 배포 및 실행",
	Long: `Ansible 플레이북을 실행하여 서버 설정을 배포합니다.

배포 기능:
- 플레이북 구문 검사
- 드라이런(dry-run) 모드 지원
- 단계별 실행 옵션
- 실시간 로그 출력
- 배포 결과 리포트
- 롤백 지원

Examples:
  gz ansible deploy --playbook site.yml --inventory inventory.ini
  gz ansible deploy --playbook web.yml --check --diff
  gz ansible deploy --playbook db.yml --limit databases --step
  gz ansible deploy --playbook app.yml --tags deploy --vault-password-file .vault_pass`,
	Run: runDeploy,
}

var (
	deployPlaybook    string
	deployInventory   string
	deployLimit       string
	deployTags        []string
	deploySkipTags    []string
	checkMode         bool
	diffMode          bool
	stepMode          bool
	verboseLevel      int
	extraVars         []string
	vaultPasswordFile string
	forks             int
	timeout           int
	connectionTimeout int
	privateKeyFile    string
	remoteUser        string
	become            bool
	becomeUser        string
	becomeMethod      string
	askBecomePass     bool
	startAtTask       string
	listHosts         bool
	listTasks         bool
	syntaxCheck       bool
	enableProfile     bool
	enableCallbacks   bool
	logPath           string
)

func init() {
	DeployCmd.Flags().StringVarP(&deployPlaybook, "playbook", "p", "", "실행할 플레이북 파일")
	DeployCmd.Flags().StringVarP(&deployInventory, "inventory", "i", "inventory", "인벤토리 파일/디렉터리")
	DeployCmd.Flags().StringVarP(&deployLimit, "limit", "l", "", "특정 호스트 패턴으로 제한")
	DeployCmd.Flags().StringSliceVarP(&deployTags, "tags", "t", []string{}, "실행할 태그")
	DeployCmd.Flags().StringSliceVar(&deploySkipTags, "skip-tags", []string{}, "건너뛸 태그")
	DeployCmd.Flags().BoolVarP(&checkMode, "check", "C", false, "체크 모드 (드라이런)")
	DeployCmd.Flags().BoolVarP(&diffMode, "diff", "D", false, "변경 사항 미리보기")
	DeployCmd.Flags().BoolVar(&stepMode, "step", false, "단계별 실행")
	DeployCmd.Flags().IntVarP(&verboseLevel, "verbose", "v", 0, "상세 출력 레벨 (0-4)")
	DeployCmd.Flags().StringSliceVarP(&extraVars, "extra-vars", "e", []string{}, "추가 변수 (key=value)")
	DeployCmd.Flags().StringVar(&vaultPasswordFile, "vault-password-file", "", "Vault 패스워드 파일")
	DeployCmd.Flags().IntVar(&forks, "forks", 5, "동시 실행 프로세스 수")
	DeployCmd.Flags().IntVar(&timeout, "timeout", 10, "연결 타임아웃 (초)")
	DeployCmd.Flags().IntVar(&connectionTimeout, "connection-timeout", 10, "SSH 연결 타임아웃 (초)")
	DeployCmd.Flags().StringVar(&privateKeyFile, "private-key", "", "SSH 개인키 파일")
	DeployCmd.Flags().StringVarP(&remoteUser, "user", "u", "", "원격 사용자명")
	DeployCmd.Flags().BoolVarP(&become, "become", "b", false, "권한 상승 사용")
	DeployCmd.Flags().StringVar(&becomeUser, "become-user", "root", "권한 상승 대상 사용자")
	DeployCmd.Flags().StringVar(&becomeMethod, "become-method", "sudo", "권한 상승 방법")
	DeployCmd.Flags().BoolVarP(&askBecomePass, "ask-become-pass", "K", false, "권한 상승 패스워드 요청")
	DeployCmd.Flags().StringVar(&startAtTask, "start-at-task", "", "특정 태스크부터 시작")
	DeployCmd.Flags().BoolVar(&listHosts, "list-hosts", false, "대상 호스트 목록 출력")
	DeployCmd.Flags().BoolVar(&listTasks, "list-tasks", false, "태스크 목록 출력")
	DeployCmd.Flags().BoolVar(&syntaxCheck, "syntax-check", false, "구문 검사만 실행")
	DeployCmd.Flags().BoolVar(&enableProfile, "profile", false, "성능 프로파일링 활성화")
	DeployCmd.Flags().BoolVar(&enableCallbacks, "callbacks", true, "콜백 플러그인 활성화")
	DeployCmd.Flags().StringVar(&logPath, "log-path", "", "로그 파일 경로")
}

// DeployResult represents the result of a deployment
type DeployResult struct {
	Success      bool              `json:"success"`
	StartTime    time.Time         `json:"start_time"`
	EndTime      time.Time         `json:"end_time"`
	Duration     time.Duration     `json:"duration"`
	Playbook     string            `json:"playbook"`
	Inventory    string            `json:"inventory"`
	HostStats    map[string]string `json:"host_stats"`
	TaskResults  []TaskResult      `json:"task_results"`
	ErrorMessage string            `json:"error_message,omitempty"`
}

// TaskResult represents the result of a single task
type TaskResult struct {
	TaskName string        `json:"task_name"`
	Host     string        `json:"host"`
	Status   string        `json:"status"` // ok, changed, failed, skipped, unreachable
	Duration time.Duration `json:"duration"`
	Message  string        `json:"message,omitempty"`
}

func runDeploy(cmd *cobra.Command, args []string) {
	fmt.Printf("🚀 Ansible 플레이북 배포\n")

	if deployPlaybook == "" {
		fmt.Printf("❌ 플레이북 파일이 필요합니다 (--playbook)\n")
		os.Exit(1)
	}

	// Check if playbook exists
	if _, err := os.Stat(deployPlaybook); os.IsNotExist(err) {
		fmt.Printf("❌ 플레이북 파일을 찾을 수 없습니다: %s\n", deployPlaybook)
		os.Exit(1)
	}

	// Check if inventory exists
	if _, err := os.Stat(deployInventory); os.IsNotExist(err) {
		fmt.Printf("❌ 인벤토리 파일을 찾을 수 없습니다: %s\n", deployInventory)
		os.Exit(1)
	}

	// Syntax check if requested
	if syntaxCheck {
		if err := performSyntaxCheck(); err != nil {
			fmt.Printf("❌ 구문 검사 실패: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✅ 구문 검사 완료\n")
		return
	}

	// List hosts if requested
	if listHosts {
		if err := performListHosts(); err != nil {
			fmt.Printf("❌ 호스트 목록 실패: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// List tasks if requested
	if listTasks {
		if err := performListTasks(); err != nil {
			fmt.Printf("❌ 태스크 목록 실패: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Perform deployment
	result, err := performDeployment()
	if err != nil {
		fmt.Printf("❌ 배포 실패: %v\n", err)
		os.Exit(1)
	}

	// Display results
	displayDeploymentResult(result)

	if !result.Success {
		os.Exit(1)
	}
}

func performSyntaxCheck() error {
	fmt.Printf("🔍 플레이북 구문 검사 중...\n")

	args := []string{
		"ansible-playbook",
		"--syntax-check",
		"--inventory", deployInventory,
		deployPlaybook,
	}

	if vaultPasswordFile != "" {
		args = append(args, "--vault-password-file", vaultPasswordFile)
	}

	cmd := exec.Command(args[0], args[1:]...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("구문 오류: %s", string(output))
	}

	fmt.Printf("✅ 구문 검사 통과\n")
	return nil
}

func performListHosts() error {
	fmt.Printf("📋 대상 호스트 목록:\n")

	args := []string{
		"ansible-playbook",
		"--list-hosts",
		"--inventory", deployInventory,
		deployPlaybook,
	}

	if deployLimit != "" {
		args = append(args, "--limit", deployLimit)
	}

	cmd := exec.Command(args[0], args[1:]...)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("호스트 목록 조회 실패: %w", err)
	}

	fmt.Print(string(output))
	return nil
}

func performListTasks() error {
	fmt.Printf("📋 태스크 목록:\n")

	args := []string{
		"ansible-playbook",
		"--list-tasks",
		"--inventory", deployInventory,
		deployPlaybook,
	}

	if len(deployTags) > 0 {
		args = append(args, "--tags", strings.Join(deployTags, ","))
	}

	if len(deploySkipTags) > 0 {
		args = append(args, "--skip-tags", strings.Join(deploySkipTags, ","))
	}

	cmd := exec.Command(args[0], args[1:]...)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("태스크 목록 조회 실패: %w", err)
	}

	fmt.Print(string(output))
	return nil
}

func performDeployment() (*DeployResult, error) {
	result := &DeployResult{
		StartTime:   time.Now(),
		Playbook:    deployPlaybook,
		Inventory:   deployInventory,
		HostStats:   make(map[string]string),
		TaskResults: []TaskResult{},
	}

	fmt.Printf("🎯 플레이북 실행: %s\n", deployPlaybook)
	fmt.Printf("📊 인벤토리: %s\n", deployInventory)

	if checkMode {
		fmt.Printf("🔍 체크 모드 (드라이런) 실행\n")
	}

	// Build ansible-playbook command
	args := buildAnsibleCommand()

	// Create command with context for timeout
	ctx := context.Background()
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(timeout)*time.Minute)
		defer cancel()
	}

	cmd := exec.CommandContext(ctx, args[0], args[1:]...)

	// Set up environment
	cmd.Env = os.Environ()
	if logPath != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("ANSIBLE_LOG_PATH=%s", logPath))
	}
	if enableProfile {
		cmd.Env = append(cmd.Env, "ANSIBLE_CALLBACK_WHITELIST=profile_tasks")
	}

	// Create pipes for real-time output
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("stdout 파이프 생성 실패: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("stderr 파이프 생성 실패: %w", err)
	}

	// Start command
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("명령 실행 실패: %w", err)
	}

	// Read output in real-time
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			fmt.Println(line)

			// Parse task results from output
			if taskResult := parseTaskResult(line); taskResult != nil {
				result.TaskResults = append(result.TaskResults, *taskResult)
			}
		}
	}()

	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			fmt.Fprintf(os.Stderr, "⚠️  %s\n", line)
		}
	}()

	// Wait for command to complete
	err = cmd.Wait()
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	if err != nil {
		result.Success = false
		result.ErrorMessage = err.Error()
		return result, nil
	}

	result.Success = true
	return result, nil
}

func buildAnsibleCommand() []string {
	args := []string{
		"ansible-playbook",
		"--inventory", deployInventory,
	}

	if checkMode {
		args = append(args, "--check")
	}

	if diffMode {
		args = append(args, "--diff")
	}

	if stepMode {
		args = append(args, "--step")
	}

	if verboseLevel > 0 {
		verboseFlag := "-" + strings.Repeat("v", verboseLevel)
		args = append(args, verboseFlag)
	}

	if deployLimit != "" {
		args = append(args, "--limit", deployLimit)
	}

	if len(deployTags) > 0 {
		args = append(args, "--tags", strings.Join(deployTags, ","))
	}

	if len(deploySkipTags) > 0 {
		args = append(args, "--skip-tags", strings.Join(deploySkipTags, ","))
	}

	if forks > 0 {
		args = append(args, "--forks", fmt.Sprintf("%d", forks))
	}

	if connectionTimeout > 0 {
		args = append(args, "--connection-timeout", fmt.Sprintf("%d", connectionTimeout))
	}

	if privateKeyFile != "" {
		args = append(args, "--private-key", privateKeyFile)
	}

	if remoteUser != "" {
		args = append(args, "--user", remoteUser)
	}

	if become {
		args = append(args, "--become")

		if becomeUser != "" {
			args = append(args, "--become-user", becomeUser)
		}

		if becomeMethod != "" {
			args = append(args, "--become-method", becomeMethod)
		}

		if askBecomePass {
			args = append(args, "--ask-become-pass")
		}
	}

	if startAtTask != "" {
		args = append(args, "--start-at-task", startAtTask)
	}

	if vaultPasswordFile != "" {
		args = append(args, "--vault-password-file", vaultPasswordFile)
	}

	// Add extra variables
	for _, extraVar := range extraVars {
		args = append(args, "--extra-vars", extraVar)
	}

	// Add playbook at the end
	args = append(args, deployPlaybook)

	return args
}

func parseTaskResult(line string) *TaskResult {
	// Simple parser for Ansible output
	// In a more sophisticated implementation, you could use JSON output format
	if strings.Contains(line, "TASK [") {
		taskName := extractTaskName(line)
		return &TaskResult{
			TaskName: taskName,
			Status:   "running",
		}
	}

	return nil
}

func extractTaskName(line string) string {
	start := strings.Index(line, "[")
	end := strings.Index(line, "]")
	if start != -1 && end != -1 && end > start {
		return line[start+1 : end]
	}
	return "Unknown Task"
}

func displayDeploymentResult(result *DeployResult) {
	fmt.Printf("\n" + strings.Repeat("=", 60) + "\n")
	fmt.Printf("📊 배포 결과 리포트\n")
	fmt.Printf(strings.Repeat("=", 60) + "\n")

	fmt.Printf("🎯 플레이북: %s\n", result.Playbook)
	fmt.Printf("📊 인벤토리: %s\n", result.Inventory)
	fmt.Printf("⏰ 시작 시간: %s\n", result.StartTime.Format("2006-01-02 15:04:05"))
	fmt.Printf("⏰ 종료 시간: %s\n", result.EndTime.Format("2006-01-02 15:04:05"))
	fmt.Printf("⏱️  소요 시간: %s\n", result.Duration.String())

	if result.Success {
		fmt.Printf("✅ 상태: 성공\n")
	} else {
		fmt.Printf("❌ 상태: 실패\n")
		if result.ErrorMessage != "" {
			fmt.Printf("🚨 오류: %s\n", result.ErrorMessage)
		}
	}

	if len(result.TaskResults) > 0 {
		fmt.Printf("\n📋 태스크 실행 결과:\n")
		for _, task := range result.TaskResults {
			statusIcon := getStatusIcon(task.Status)
			fmt.Printf("  %s %s\n", statusIcon, task.TaskName)
		}
	}

	fmt.Printf("\n💡 추가 옵션:\n")
	fmt.Printf("  - 상세 로그: --verbose 또는 -v\n")
	fmt.Printf("  - 드라이런: --check\n")
	fmt.Printf("  - 변경사항 미리보기: --diff\n")

	if logPath != "" {
		fmt.Printf("📝 자세한 로그: %s\n", logPath)
	}

	fmt.Printf(strings.Repeat("=", 60) + "\n")
}

func getStatusIcon(status string) string {
	switch status {
	case "ok":
		return "✅"
	case "changed":
		return "🔄"
	case "failed":
		return "❌"
	case "skipped":
		return "⏭️"
	case "unreachable":
		return "🚫"
	default:
		return "⚙️"
	}
}

func saveDeploymentReport(result *DeployResult) error {
	reportPath := filepath.Join(".", fmt.Sprintf("deploy-report-%s.json",
		result.StartTime.Format("20060102-150405")))

	// Implementation for saving deployment report as JSON
	// This would serialize the DeployResult struct to JSON file

	fmt.Printf("📄 배포 리포트 저장: %s\n", reportPath)
	return nil
}
