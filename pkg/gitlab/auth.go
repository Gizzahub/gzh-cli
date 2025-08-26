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
