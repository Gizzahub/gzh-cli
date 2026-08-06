// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

package errors

import (
	sterrors "errors"
	"fmt"
)

var (
	// ErrConfigNotFound indicates that configuration could not be located.
	ErrConfigNotFound = sterrors.New("config not found")
	// ErrInvalidConfig indicates the provided configuration is invalid.
	ErrInvalidConfig = sterrors.New("invalid config")
	// ErrConfigNotLoaded indicates no configuration has been loaded.
	ErrConfigNotLoaded = sterrors.New("no configuration loaded")
)

// Wrap annotates err with target to allow errors.Is/As checks on target while
// preserving the original error as the cause.
//
// target이 이미 err의 사슬에 있으면 err를 그대로 돌려준다. target은 분류
// 표지라서 두 번 붙여도 새로 알려주는 것이 없고, 계층마다 같은 표지를
// 붙이면 말이 겹쳐 나온다. 실제로 설정 파일을 못 찾으면 파사드와 서비스와
// 명령이 차례로 감싸서 "config not found: config not found: config not
// found: configuration file not found: /경로"가 나왔다.
func Wrap(err, target error) error {
	if err == nil {
		return target
	}

	if target == nil {
		return err
	}

	if sterrors.Is(err, target) {
		return err
	}

	return fmt.Errorf("%w: %w", target, err)
}
