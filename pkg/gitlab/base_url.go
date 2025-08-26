package gitlab

import (
	"os"
	"strings"
)

// configuredBaseAPIURL 는 런타임에 주입된 GitLab API 베이스 URL이다.
// 환경변수나 기본값보다 우선한다.
var configuredBaseAPIURL string

// SetBaseURL 은 사용자가 제공한 base_url을 받아 API v4 엔드포인트로 정규화한다.
// 예) https://gitlab.company.com -> https://gitlab.company.com/api/v4
func SetBaseURL(baseURL string) {
	configuredBaseAPIURL = normalizeAPIBase(baseURL)
}

// getBaseAPIURL 은 우선순위에 따라 GitLab API의 베이스 URL을 반환한다.
// 1) SetBaseURL로 주입된 값
// 2) GITLAB_BASE_URL (root, /api/v4 붙임)
// 3) GITLAB_API_URL (완전한 API URL)
// 4) GZH_GITLAB_API (완전한 API URL)
// 5) 기본값 https://gitlab.com/api/v4
func getBaseAPIURL() string {
	if configuredBaseAPIURL != "" {
		return configuredBaseAPIURL
	}

	if v := os.Getenv("GITLAB_BASE_URL"); v != "" {
		return normalizeAPIBase(v)
	}
	if v := os.Getenv("GITLAB_API_URL"); v != "" {
		return trimTrailingSlash(v)
	}
	if v := os.Getenv("GZH_GITLAB_API"); v != "" {
		return trimTrailingSlash(v)
	}
	return "https://gitlab.com/api/v4"
}

func normalizeAPIBase(u string) string {
	u = trimTrailingSlash(u)
	// 이미 /api/ 경로를 포함하면 그대로 사용
	if strings.Contains(u, "/api/") {
		return u
	}
	return u + "/api/v4"
}

func trimTrailingSlash(s string) string {
	return strings.TrimSuffix(s, "/")
}

// buildAPIURL 은 엔드포인트를 베이스 API URL과 합성한다.
func buildAPIURL(endpoint string) string {
	base := getBaseAPIURL()
	endpoint = strings.TrimPrefix(endpoint, "/")
	return base + "/" + endpoint
}

func getWebBaseURL() string {
	base := getBaseAPIURL()
	if idx := strings.Index(base, "/api/"); idx != -1 {
		return base[:idx]
	}
	// fallback: common suffix trims
	base = strings.TrimSuffix(base, "/api/v4")
	base = strings.TrimSuffix(base, "/api")
	return trimTrailingSlash(base)
}
