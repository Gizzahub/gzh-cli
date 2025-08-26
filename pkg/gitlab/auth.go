package gitlab

import (
	"fmt"
	"net/http"
)

// configuredToken 은 런타임에 주입된 GitLab 토큰이다.
var configuredToken string

// SetToken 은 API 요청에 사용할 토큰을 설정한다.
func SetToken(token string) {
	configuredToken = token
}

// addAuthHeader 는 토큰이 있으면 PRIVATE-TOKEN 헤더를 추가한다.
func addAuthHeader(req *http.Request) {
	if configuredToken == "" {
		return
	}
	req.Header.Set("PRIVATE-TOKEN", configuredToken)
}

// tokenGuidanceMessage 는 필요한 토큰 권한과 토큰 생성 페이지 URL을 안내한다.
// 한국어 안내 메시지로 출력되도록 구성.
func tokenGuidanceMessage() string {
	base := getWebBaseURL()
	// GitLab 개인 액세스 토큰 생성 페이지
	patURL := base + "/-/profile/personal_access_tokens"
	return "필요 토큰 권한: read_api, read_repository. 토큰 생성: " + patURL
}

// formatGuidanceBox 는 가독성 좋은 박스 형태로 가이드를 포매팅한다.
func formatGuidanceBox(title, content string) string {
	return fmt.Sprintf(`
┌─ %s ─────────────────────────────────────────────────────────────────────────┐
│ %s
└─────────────────────────────────────────────────────────────────────────────────┘`, title, content)
}

// accessGuidanceMessage 는 접근 실패 시 종합 가이드를 제공한다.
// - 필요한 토큰 권한, 발급 URL
// - 최소 프로젝트/그룹 역할 요구사항
func accessGuidanceMessage() string {
	base := getWebBaseURL()
	patURL := base + "/-/profile/personal_access_tokens"

	content := fmt.Sprintf(`토큰 발급: %s
│ 필요 권한: read_api, read_repository
│ 최소 역할: Reporter 이상 (그룹 내 모든 프로젝트)`, patURL)

	return formatGuidanceBox("GitLab 인증 가이드", content)
}
