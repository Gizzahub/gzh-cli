// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package gitea

import (
	"net/http"
)

// configuredToken 은 런타임에 주입된 Gitea 토큰이다.
var configuredToken string

// SetToken 은 API 요청에 사용할 토큰을 설정한다.
func SetToken(token string) {
	configuredToken = token
}

// addAuthHeader 는 토큰이 있으면 Authorization 헤더를 추가한다.
// Gitea는 "token <token>" 형식을 사용한다.
func addAuthHeader(req *http.Request) {
	if configuredToken == "" {
		return
	}
	req.Header.Set("Authorization", "token "+configuredToken)
}

// GetToken 은 현재 설정된 토큰을 반환한다.
func GetToken() string {
	return configuredToken
}
