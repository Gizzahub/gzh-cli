package ansible

import (
	"github.com/spf13/cobra"
)

// AnsibleCmd represents the ansible command
var AnsibleCmd = &cobra.Command{
	Use:   "ansible",
	Short: "Ansible 플레이북 및 인벤토리 관리",
	Long: `Ansible 자동화 도구 관리 및 플레이북 생성.

서버 설정 자동화를 위한 Ansible 도구:
- 플레이북 생성 및 템플릿 관리
- 인벤토리 파일 자동 생성
- 역할(Role) 정의 및 관리
- Ansible Vault 암호화 통합
- 다중 환경 설정 지원

사용 가능한 명령어:
  generate     Ansible 플레이북 및 역할 생성
  inventory    인벤토리 파일 관리
  vault        Ansible Vault 암호화 관리
  deploy       플레이북 배포 및 실행`,
	Aliases: []string{"ans"},
}

func init() {
	AnsibleCmd.AddCommand(GenerateCmd)
	AnsibleCmd.AddCommand(InventoryCmd)
	AnsibleCmd.AddCommand(VaultCmd)
	AnsibleCmd.AddCommand(DeployCmd)
}
