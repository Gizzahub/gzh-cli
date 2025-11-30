// Copyright (c) 2025 Archmagece
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
func Wrap(err, target error) error {
	if err == nil {
		return target
	}
	if target == nil {
		return err
	}
	return fmt.Errorf("%w: %w", target, err)
}
