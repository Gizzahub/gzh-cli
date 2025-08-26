package gitlab

import "net/http"

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
