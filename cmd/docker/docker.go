package docker

import (
	"github.com/spf13/cobra"
)

// DockerCmd represents the docker command.
var DockerCmd = &cobra.Command{
	Use:   "docker",
	Short: "컨테이너 이미지 관리 및 자동화",
	Long: `Docker 컨테이너 이미지 빌드, 배포, 관리를 자동화합니다.

컨테이너 이미지 관리 기능:
- 멀티 아키텍처 이미지 빌드 (amd64, arm64)
- 이미지 레지스트리 관리 및 배포
- 취약점 스캔 및 보안 검사
- 이미지 최적화 및 크기 분석
- CI/CD 파이프라인 통합
- 자동 태깅 및 버전 관리

사용 가능한 명령어:
  dockerfile  최적화된 Dockerfile 생성
  build       이미지 자동 빌드 및 배포
  scan        보안 취약점 스캔
  optimize    이미지 최적화
  registry    레지스트리 관리
  pipeline    CI/CD 파이프라인 생성`,
}

func init() {
	DockerCmd.AddCommand(DockerfileCmd)
	DockerCmd.AddCommand(BuildCmd)
	DockerCmd.AddCommand(ScanCmd)
	DockerCmd.AddCommand(OptimizeCmd)
	DockerCmd.AddCommand(RegistryCmd)
	DockerCmd.AddCommand(PipelineCmd)
}
