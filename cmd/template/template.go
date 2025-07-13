package template

import (
	"github.com/spf13/cobra"
)

// TemplateCmd represents the template command
var TemplateCmd = &cobra.Command{
	Use:   "template",
	Short: "템플릿 저장소 및 마켓플레이스 관리",
	Long: `템플릿 마켓플레이스 시스템 관리 도구.

템플릿 관리 기능:
- 템플릿 저장소 구조 관리
- 메타데이터 스키마 검증
- 버전 관리 및 의존성 해결
- 커뮤니티 템플릿 공유
- 기업용 프라이빗 마켓플레이스

사용 가능한 명령어:
  init         템플릿 저장소 초기화
  validate     템플릿 메타데이터 검증
  publish      템플릿 퍼블리시
  search       템플릿 검색
  install      템플릿 설치
  marketplace  마켓플레이스 관리`,
	Aliases: []string{"tpl"},
}

func init() {
	TemplateCmd.AddCommand(InitCmd)
	TemplateCmd.AddCommand(ValidateCmd)
	TemplateCmd.AddCommand(PublishCmd)
	TemplateCmd.AddCommand(SearchCmd)
	TemplateCmd.AddCommand(InstallCmd)
	TemplateCmd.AddCommand(MarketplaceCmd)
}
