// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package errors

import (
	sterrors "errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWrap(t *testing.T) {
	cause := sterrors.New("configuration file not found: /etc/gzh.yaml")

	t.Run("표지와 원인을 모두 남긴다", func(t *testing.T) {
		err := Wrap(cause, ErrConfigNotFound)

		require.Error(t, err)
		assert.Equal(t, "config not found: configuration file not found: /etc/gzh.yaml", err.Error())
		assert.ErrorIs(t, err, ErrConfigNotFound)
		assert.ErrorIs(t, err, cause)
	})

	t.Run("같은 표지를 두 번 붙이지 않는다", func(t *testing.T) {
		// 파사드가 감싸고, 서비스가 감싸고, 명령이 또 감싸는 실제 경로다.
		// 예전에는 "config not found:"가 세 번 겹쳐 나왔다.
		once := Wrap(cause, ErrConfigNotFound)
		twice := Wrap(once, ErrConfigNotFound)
		thrice := Wrap(twice, ErrConfigNotFound)

		assert.Equal(t, once.Error(), thrice.Error())
		assert.ErrorIs(t, thrice, ErrConfigNotFound)
		assert.ErrorIs(t, thrice, cause)
	})

	t.Run("사슬 깊은 곳에 있어도 알아본다", func(t *testing.T) {
		// 중간에 다른 설명이 끼어도 표지는 이미 사슬에 있다.
		inner := Wrap(cause, ErrConfigNotFound)
		middle := fmt.Errorf("failed to load config: %w", inner)

		err := Wrap(middle, ErrConfigNotFound)

		assert.Equal(t, middle.Error(), err.Error())
		assert.ErrorIs(t, err, ErrConfigNotFound)
	})

	t.Run("다른 표지는 그대로 덧붙인다", func(t *testing.T) {
		err := Wrap(Wrap(cause, ErrConfigNotFound), ErrInvalidConfig)

		assert.Equal(t, "invalid config: config not found: configuration file not found: /etc/gzh.yaml", err.Error())
		assert.ErrorIs(t, err, ErrInvalidConfig)
		assert.ErrorIs(t, err, ErrConfigNotFound)
	})

	t.Run("nil 인자", func(t *testing.T) {
		assert.Equal(t, ErrConfigNotFound, Wrap(nil, ErrConfigNotFound))
		assert.Equal(t, cause, Wrap(cause, nil))
		assert.NoError(t, Wrap(nil, nil))
	})
}
