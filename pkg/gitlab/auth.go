package gitlab

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Gizzahub/gzh-cli/internal/httpclient"
)

// configuredToken 은 런타임에 주입된 GitLab 토큰이다.
var configuredToken string

// cachedUsername 은 토큰으로 조회한 사용자명을 캐시한다.
var cachedUsername string

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

// getCurrentUser 는 현재 토큰의 사용자 정보를 가져온다.
func getCurrentUser(ctx context.Context) (string, error) {
	if configuredToken == "" {
		return "", fmt.Errorf("no token configured")
	}

	// 캐시된 사용자명이 있으면 반환
	if cachedUsername != "" {
		return cachedUsername, nil
	}

	url := buildAPIURL("user")
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	addAuthHeader(req)

	client := httpclient.GetGlobalClient("gitlab")
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get user info: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get user info: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var user struct {
		Username string `json:"username"`
	}

	if err := json.Unmarshal(body, &user); err != nil {
		return "", fmt.Errorf("failed to parse user info: %w", err)
	}

	// 사용자명 캐시
	cachedUsername = user.Username
	return cachedUsername, nil
}
