//nolint:testpackage // White-box testing needed for internal function access
package synclone

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	gerrors "github.com/gizzahub/gzh-cli/internal/errors"
)

// TestIsRateLimitError는 요청 한도 판정이 감싸인 오류에도 닿는지 본다.
//
// 예전 코드는 `err.Error() == "rate limit"`이었다. 이 저장소 어디에도 Error()가
// 정확히 그 문자열인 오류가 없어서 한 번도 참이 된 적이 없다.
func TestIsRateLimitError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "RecoverableError로 이미 분류된 것",
			err: gerrors.NewRecoverableError(
				gerrors.ErrorTypeRateLimit, "GitHub API rate limit exceeded", nil, true,
			),
			want: true,
		},
		{
			name: "RecoverableError를 한 번 더 감싼 것",
			err: fmt.Errorf("failed to list repositories: %w",
				gerrors.NewRecoverableError(
					gerrors.ErrorTypeRateLimit, "GitHub API rate limit exceeded", nil, true,
				)),
			want: true,
		},
		{
			name: "문자열만 있는 것",
			err:  fmt.Errorf("403 Forbidden: API rate limit exceeded for user"),
			want: true,
		},
		{
			name: "예전 코드가 유일하게 잡던 모양",
			err:  fmt.Errorf("rate limit"),
			want: true,
		},
		{
			name: "관계없는 오류",
			err:  fmt.Errorf("connection refused"),
			want: false,
		},
		{
			name: "취소는 한도 초과가 아니다",
			err:  context.Canceled,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isRateLimitError(tt.err))
		})
	}
}

// TestClassifyGithubError는 run()의 분류 갈래가 감싸인 취소를 알아보는지 본다.
//
// 예전 코드는 `err.Error() == "context canceled"`였다. fmt.Errorf로 한 번만
// 감싸도 문자열이 달라져서 network로 잘못 분류됐다.
func TestClassifyGithubError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want gerrors.ErrorType
	}{
		{"맨 취소", context.Canceled, gerrors.ErrorTypeTimeout},
		{
			"감싸인 취소",
			fmt.Errorf("clone failed for repo foo: %w", context.Canceled),
			gerrors.ErrorTypeTimeout,
		},
		{"기한 초과", context.DeadlineExceeded, gerrors.ErrorTypeTimeout},
		{
			"감싸인 한도 초과",
			fmt.Errorf("failed to get repositories: %w",
				gerrors.NewRecoverableError(
					gerrors.ErrorTypeRateLimit, "API rate limit exceeded", nil, true,
				)),
			gerrors.ErrorTypeRateLimit,
		},
		{"그 밖", fmt.Errorf("dial tcp: connection refused"), gerrors.ErrorTypeNetwork},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, classifyGithubError(tt.err))
		})
	}
}
