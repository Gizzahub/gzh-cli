package ansible

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/term"
)

// VaultCmd represents the vault command
var VaultCmd = &cobra.Command{
	Use:   "vault",
	Short: "Ansible Vault 암호화 관리",
	Long: `Ansible Vault를 사용하여 민감한 정보를 안전하게 암호화하고 관리합니다.

Vault 관리 기능:
- 변수 파일 암호화/복호화
- 인라인 문자열 암호화
- 패스워드 변경
- Vault ID 관리
- 키 파일 생성 및 관리
- 자동 백업 및 복원

Examples:
  gz ansible vault encrypt --file secrets.yml
  gz ansible vault decrypt --file secrets.yml --output secrets_plain.yml
  gz ansible vault encrypt-string --string "mysecret" --name "db_password"
  gz ansible vault create --file vault_vars.yml
  gz ansible vault rekey --file secrets.yml`,
	Run: runVault,
}

var (
	vaultAction       string
	vaultFile         string
	vaultOutput       string
	vaultString        string
	vaultName          string
	vaultPassword      string
	vaultPasswordPath  string
	vaultId            string
	vaultFormat       string
	createBackup      bool
	force             bool
	generateKey       bool
	keyLength         int
)

func init() {
	VaultCmd.Flags().StringVarP(&vaultAction, "action", "a", "encrypt", "액션 (encrypt, decrypt, encrypt-string, create, rekey, view)")
	VaultCmd.Flags().StringVarP(&vaultFile, "file", "f", "", "대상 파일 경로")
	VaultCmd.Flags().StringVarP(&vaultOutput, "output", "o", "", "출력 파일 경로")
	VaultCmd.Flags().StringVarP(&vaultString, "string", "s", "", "암호화할 문자열")
	VaultCmd.Flags().StringVarP(&vaultName, "name", "n", "", "변수 이름")
	VaultCmd.Flags().StringVar(&vaultPassword, "password", "", "Vault 패스워드")
	VaultCmd.Flags().StringVar(&vaultPasswordPath, "password-file", "", "패스워드 파일 경로")
	VaultCmd.Flags().StringVar(&vaultId, "vault-id", "default", "Vault ID")
	VaultCmd.Flags().StringVar(&vaultFormat, "format", "yaml", "출력 형식 (yaml, json)")
	VaultCmd.Flags().BoolVar(&createBackup, "backup", true, "백업 파일 생성")
	VaultCmd.Flags().BoolVar(&force, "force", false, "강제 실행")
	VaultCmd.Flags().BoolVar(&generateKey, "generate-key", false, "키 파일 생성")
	VaultCmd.Flags().IntVar(&keyLength, "key-length", 32, "생성할 키 길이")
}

// VaultHeader represents the Ansible Vault file header
const VaultHeader = "$ANSIBLE_VAULT;1.1;AES256"

// VaultData represents encrypted vault data structure
type VaultData struct {
	Header string `yaml:"-"`
	Data   string `yaml:"-"`
}

// VaultMetadata represents vault file metadata
type VaultMetadata struct {
	VaultId     string `yaml:"vault_id"`
	Created     string `yaml:"created"`
	Description string `yaml:"description,omitempty"`
}

func runVault(cmd *cobra.Command, args []string) {
	fmt.Printf("🔐 Ansible Vault 관리\n")
	fmt.Printf("🎯 액션: %s\n", vaultAction)

	switch vaultAction {
	case "encrypt":
		if err := encryptVaultFile(); err != nil {
			fmt.Printf("❌ 암호화 실패: %v\n", err)
			os.Exit(1)
		}
	case "decrypt":
		if err := decryptVaultFile(); err != nil {
			fmt.Printf("❌복호화 실패: %v\n", err)
			os.Exit(1)
		}
	case "encrypt-string":
		if err := encryptVaultString(); err != nil {
			fmt.Printf("❌ 문자열 암호화 실패: %v\n", err)
			os.Exit(1)
		}
	case "create":
		if err := createVaultFile(); err != nil {
			fmt.Printf("❌ Vault 파일 생성 실패: %v\n", err)
			os.Exit(1)
		}
	case "rekey":
		if err := rekeyVaultFile(); err != nil {
			fmt.Printf("❌ 키 변경 실패: %v\n", err)
			os.Exit(1)
		}
	case "view":
		if err := viewVaultFile(); err != nil {
			fmt.Printf("❌ 파일 보기 실패: %v\n", err)
			os.Exit(1)
		}
	case "generate-key":
		if err := generateVaultKey(); err != nil {
			fmt.Printf("❌ 키 생성 실패: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Printf("❌ 알 수 없는 액션: %s\n", vaultAction)
		cmd.Help()
		os.Exit(1)
	}

	fmt.Printf("✅ Vault 작업 완료\n")
}

func encryptVaultFile() error {
	if vaultFile == "" {
		return fmt.Errorf("암호화할 파일이 필요합니다 (--file)")
	}

	fmt.Printf("🔒 파일 암호화: %s\n", vaultFile)

	// Check if file exists
	if _, err := os.Stat(vaultFile); os.IsNotExist(err) {
		return fmt.Errorf("파일을 찾을 수 없습니다: %s", vaultFile)
	}

	// Check if already encrypted
	if isVaultFile(vaultFile) && !force {
		return fmt.Errorf("파일이 이미 암호화되어 있습니다 (--force로 강제 실행 가능)")
	}

	// Get password
	password, err := getVaultPassword()
	if err != nil {
		return err
	}

	// Read original file
	data, err := os.ReadFile(vaultFile)
	if err != nil {
		return fmt.Errorf("파일 읽기 실패: %w", err)
	}

	// Create backup if requested
	if createBackup {
		backupFile := vaultFile + ".backup"
		if err := os.WriteFile(backupFile, data, 0o644); err != nil {
			fmt.Printf("⚠️ 백업 생성 실패: %v\n", err)
		} else {
			fmt.Printf("💾 백업 생성: %s\n", backupFile)
		}
	}

	// Encrypt data
	encryptedData, err := encryptData(data, password)
	if err != nil {
		return fmt.Errorf("암호화 실패: %w", err)
	}

	// Write encrypted file
	outputFile := vaultFile
	if vaultOutput != "" {
		outputFile = vaultOutput
	}

	if err := writeVaultFile(outputFile, encryptedData); err != nil {
		return fmt.Errorf("암호화된 파일 쓰기 실패: %w", err)
	}

	fmt.Printf("✅ 파일 암호화 완료: %s\n", outputFile)
	return nil
}

func decryptVaultFile() error {
	if vaultFile == "" {
		return fmt.Errorf("복호화할 파일이 필요합니다 (--file)")
	}

	fmt.Printf("🔓 파일 복호화: %s\n", vaultFile)

	// Check if file is encrypted
	if !isVaultFile(vaultFile) {
		return fmt.Errorf("Vault 파일이 아닙니다: %s", vaultFile)
	}

	// Get password
	password, err := getVaultPassword()
	if err != nil {
		return err
	}

	// Read encrypted file
	encryptedData, err := readVaultFile(vaultFile)
	if err != nil {
		return fmt.Errorf("Vault 파일 읽기 실패: %w", err)
	}

	// Decrypt data
	decryptedData, err := decryptData(encryptedData, password)
	if err != nil {
		return fmt.Errorf("복호화 실패: %w", err)
	}

	// Write decrypted file
	outputFile := vaultFile
	if vaultOutput != "" {
		outputFile = vaultOutput
	} else {
		// Remove .vault extension if exists
		if strings.HasSuffix(outputFile, ".vault") {
			outputFile = strings.TrimSuffix(outputFile, ".vault")
		} else {
			outputFile += ".decrypted"
		}
	}

	if err := os.WriteFile(outputFile, decryptedData, 0o644); err != nil {
		return fmt.Errorf("복호화된 파일 쓰기 실패: %w", err)
	}

	fmt.Printf("✅ 파일 복호화 완료: %s\n", outputFile)
	return nil
}

func encryptVaultString() error {
	if vaultString == "" {
		return fmt.Errorf("암호화할 문자열이 필요합니다 (--string)")
	}

	fmt.Printf("🔒 문자열 암호화\n")

	// Get password
	password, err := getVaultPassword()
	if err != nil {
		return err
	}

	// Encrypt string
	encryptedData, err := encryptData([]byte(vaultString), password)
	if err != nil {
		return fmt.Errorf("문자열 암호화 실패: %w", err)
	}

	// Format as Ansible vault string
	vaultString := formatVaultString(encryptedData, vaultName)

	fmt.Printf("✅ 암호화된 문자열:\n")
	fmt.Println(vaultString)

	// Save to file if specified
	if vaultOutput != "" {
		if err := os.WriteFile(vaultOutput, []byte(vaultString), 0o644); err != nil {
			return fmt.Errorf("파일 저장 실패: %w", err)
		}
		fmt.Printf("💾 파일로 저장: %s\n", vaultOutput)
	}

	return nil
}

func createVaultFile() error {
	if vaultFile == "" {
		return fmt.Errorf("생성할 파일 경로가 필요합니다 (--file)")
	}

	fmt.Printf("📝 Vault 파일 생성: %s\n", vaultFile)

	// Check if file already exists
	if _, err := os.Stat(vaultFile); err == nil && !force {
		return fmt.Errorf("파일이 이미 존재합니다 (--force로 강제 생성 가능)")
	}

	// Create directory if needed
	if err := os.MkdirAll(filepath.Dir(vaultFile), 0o755); err != nil {
		return fmt.Errorf("디렉터리 생성 실패: %w", err)
	}

	// Create template content
	templateContent := createVaultTemplate()

	// Get password
	password, err := getVaultPassword()
	if err != nil {
		return err
	}

	// Encrypt template
	encryptedData, err := encryptData([]byte(templateContent), password)
	if err != nil {
		return fmt.Errorf("템플릿 암호화 실패: %w", err)
	}

	// Write vault file
	if err := writeVaultFile(vaultFile, encryptedData); err != nil {
		return fmt.Errorf("Vault 파일 생성 실패: %w", err)
	}

	fmt.Printf("✅ Vault 파일 생성 완료: %s\n", vaultFile)
	fmt.Printf("📝 파일을 편집하려면: ansible-vault edit %s\n", vaultFile)

	return nil
}

func rekeyVaultFile() error {
	if vaultFile == "" {
		return fmt.Errorf("키를 변경할 파일이 필요합니다 (--file)")
	}

	fmt.Printf("🔑 Vault 키 변경: %s\n", vaultFile)

	// Check if file is encrypted
	if !isVaultFile(vaultFile) {
		return fmt.Errorf("Vault 파일이 아닙니다: %s", vaultFile)
	}

	// Get old password
	fmt.Print("기존 패스워드를 입력하세요: ")
	oldPassword, err := readPassword()
	if err != nil {
		return fmt.Errorf("기존 패스워드 입력 실패: %w", err)
	}

	// Get new password
	fmt.Print("새 패스워드를 입력하세요: ")
	newPassword, err := readPassword()
	if err != nil {
		return fmt.Errorf("새 패스워드 입력 실패: %w", err)
	}

	// Confirm new password
	fmt.Print("새 패스워드를 다시 입력하세요: ")
	confirmPassword, err := readPassword()
	if err != nil {
		return fmt.Errorf("패스워드 확인 입력 실패: %w", err)
	}

	if newPassword != confirmPassword {
		return fmt.Errorf("새 패스워드가 일치하지 않습니다")
	}

	// Read and decrypt with old password
	encryptedData, err := readVaultFile(vaultFile)
	if err != nil {
		return fmt.Errorf("Vault 파일 읽기 실패: %w", err)
	}

	decryptedData, err := decryptData(encryptedData, oldPassword)
	if err != nil {
		return fmt.Errorf("기존 패스워드로 복호화 실패: %w", err)
	}

	// Create backup
	if createBackup {
		backupFile := vaultFile + ".backup"
		if err := writeVaultFile(backupFile, encryptedData); err != nil {
			fmt.Printf("⚠️ 백업 생성 실패: %v\n", err)
		} else {
			fmt.Printf("💾 백업 생성: %s\n", backupFile)
		}
	}

	// Encrypt with new password
	newEncryptedData, err := encryptData(decryptedData, newPassword)
	if err != nil {
		return fmt.Errorf("새 패스워드로 암호화 실패: %w", err)
	}

	// Write with new encryption
	if err := writeVaultFile(vaultFile, newEncryptedData); err != nil {
		return fmt.Errorf("새로 암호화된 파일 쓰기 실패: %w", err)
	}

	fmt.Printf("✅ Vault 키 변경 완료\n")
	return nil
}

func viewVaultFile() error {
	if vaultFile == "" {
		return fmt.Errorf("보려는 파일이 필요합니다 (--file)")
	}

	fmt.Printf("👁️ Vault 파일 보기: %s\n", vaultFile)

	// Check if file is encrypted
	if !isVaultFile(vaultFile) {
		// If not encrypted, just display the file
		data, err := os.ReadFile(vaultFile)
		if err != nil {
			return fmt.Errorf("파일 읽기 실패: %w", err)
		}
		fmt.Print(string(data))
		return nil
	}

	// Get password
	password, err := getVaultPassword()
	if err != nil {
		return err
	}

	// Read and decrypt
	encryptedData, err := readVaultFile(vaultFile)
	if err != nil {
		return fmt.Errorf("Vault 파일 읽기 실패: %w", err)
	}

	decryptedData, err := decryptData(encryptedData, password)
	if err != nil {
		return fmt.Errorf("복호화 실패: %w", err)
	}

	// Display content
	fmt.Println("\n--- Vault 파일 내용 ---")
	fmt.Print(string(decryptedData))
	fmt.Println("--- 끝 ---")

	return nil
}

func generateVaultKey() error {
	fmt.Printf("🔑 Vault 키 파일 생성 (길이: %d)\n", keyLength)

	// Generate random key
	key := make([]byte, keyLength)
	if _, err := rand.Read(key); err != nil {
		return fmt.Errorf("키 생성 실패: %w", err)
	}

	keyHex := hex.EncodeToString(key)

	// Output file
	outputFile := vaultOutput
	if outputFile == "" {
		outputFile = ".vault_key"
	}

	// Write key file
	if err := os.WriteFile(outputFile, []byte(keyHex), 0o600); err != nil {
		return fmt.Errorf("키 파일 저장 실패: %w", err)
	}

	fmt.Printf("✅ 키 파일 생성 완료: %s\n", outputFile)
	fmt.Printf("🔒 키: %s\n", keyHex)
	fmt.Printf("\n⚠️  보안 주의사항:\n")
	fmt.Printf("- 키 파일을 안전한 위치에 보관하세요\n")
	fmt.Printf("- 키 파일을 Git에 커밋하지 마세요\n")
	fmt.Printf("- .gitignore에 키 파일을 추가하세요\n")

	return nil
}

func getVaultPassword() (string, error) {
	if vaultPassword != "" {
		return vaultPassword, nil
	}

	if vaultPasswordPath != "" {
		data, err := os.ReadFile(vaultPasswordPath)
		if err != nil {
			return "", fmt.Errorf("패스워드 파일 읽기 실패: %w", err)
		}
		return strings.TrimSpace(string(data)), nil
	}

	// Prompt for password
	fmt.Print("Vault 패스워드를 입력하세요: ")
	password, err := readPassword()
	if err != nil {
		return "", fmt.Errorf("패스워드 입력 실패: %w", err)
	}

	return password, nil
}

func readPassword() (string, error) {
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", err
	}
	fmt.Println() // New line after password input
	return string(bytePassword), nil
}

func isVaultFile(filename string) bool {
	file, err := os.Open(filename)
	if err != nil {
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		return strings.HasPrefix(scanner.Text(), VaultHeader)
	}

	return false
}

func encryptData(data []byte, password string) (string, error) {
	// Generate salt
	salt := make([]byte, 32)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	// Derive key using PBKDF2
	key := pbkdf2.Key([]byte(password), salt, 10000, 32, sha256.New)

	// Create cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// Generate IV
	iv := make([]byte, aes.BlockSize)
	if _, err := rand.Read(iv); err != nil {
		return "", err
	}

	// Encrypt data
	stream := cipher.NewCFBEncrypter(block, iv)
	encrypted := make([]byte, len(data))
	stream.XORKeyStream(encrypted, data)

	// Combine salt + iv + encrypted data
	combined := append(salt, iv...)
	combined = append(combined, encrypted...)

	return hex.EncodeToString(combined), nil
}

func decryptData(encryptedHex, password string) ([]byte, error) {
	// Decode hex
	combined, err := hex.DecodeString(encryptedHex)
	if err != nil {
		return nil, err
	}

	if len(combined) < 32+aes.BlockSize {
		return nil, fmt.Errorf("암호화된 데이터가 너무 짧습니다")
	}

	// Extract components
	salt := combined[:32]
	iv := combined[32 : 32+aes.BlockSize]
	encrypted := combined[32+aes.BlockSize:]

	// Derive key
	key := pbkdf2.Key([]byte(password), salt, 10000, 32, sha256.New)

	// Create cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Decrypt
	stream := cipher.NewCFBDecrypter(block, iv)
	decrypted := make([]byte, len(encrypted))
	stream.XORKeyStream(decrypted, encrypted)

	return decrypted, nil
}

func writeVaultFile(filename, encryptedData string) error {
	content := fmt.Sprintf("%s;%s\n%s", VaultHeader, vaultId, encryptedData)
	return os.WriteFile(filename, []byte(content), 0o644)
}

func readVaultFile(filename string) (string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(data), "\n")
	if len(lines) < 2 {
		return "", fmt.Errorf("잘못된 Vault 파일 형식")
	}

	// Verify header
	if !strings.HasPrefix(lines[0], VaultHeader) {
		return "", fmt.Errorf("잘못된 Vault 헤더")
	}

	// Return encrypted data (everything after first line)
	return strings.Join(lines[1:], "\n"), nil
}

func formatVaultString(encryptedData, varName string) string {
	if varName != "" {
		return fmt.Sprintf("%s: !vault |\n  %s;%s\n  %s",
			varName, VaultHeader, vaultId, encryptedData)
	}
	return fmt.Sprintf("!vault |\n  %s;%s\n  %s",
		VaultHeader, vaultId, encryptedData)
}

func createVaultTemplate() string {
	switch vaultFormat {
	case "yaml":
		return `---
# Ansible Vault 변수 파일
# 이 파일을 편집하여 암호화할 변수를 추가하세요

# 데이터베이스 설정
vault_db_password: "change_me_database_password"
vault_db_root_password: "change_me_root_password"

# 애플리케이션 설정  
vault_app_secret_key: "change_me_secret_key"
vault_api_token: "change_me_api_token"

# 외부 서비스 설정
vault_email_password: "change_me_email_password"
vault_service_api_key: "change_me_service_key"

# SSH 키 및 인증서
vault_ssl_private_key: |
  -----BEGIN PRIVATE KEY-----
  (여기에 실제 개인키 내용 입력)
  -----END PRIVATE KEY-----

# 클라우드 서비스 인증
vault_aws_access_key: "change_me_aws_access_key"
vault_aws_secret_key: "change_me_aws_secret_key"
`
	case "json":
		return `{
  "_comment": "Ansible Vault 변수 파일 - JSON 형식",
  "vault_db_password": "change_me_database_password",
  "vault_db_root_password": "change_me_root_password",
  "vault_app_secret_key": "change_me_secret_key",
  "vault_api_token": "change_me_api_token",
  "vault_email_password": "change_me_email_password",
  "vault_service_api_key": "change_me_service_key",
  "vault_aws_access_key": "change_me_aws_access_key",
  "vault_aws_secret_key": "change_me_aws_secret_key"
}`
	default:
		return "# Ansible Vault 변수 파일\nvault_secret=change_me_secret_value\n"
	}
}
