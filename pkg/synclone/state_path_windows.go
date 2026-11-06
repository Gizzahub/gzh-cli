// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package bulkclone

import (
	"errors"
	"os"
	"unicode/utf16"

	"golang.org/x/sys/windows"
)

// Windows는 긴 구성 요소에 ERROR_INVALID_NAME을 반환할 수 있어 실제 볼륨 제한을 확인한다.
func legacyWindowsComponentTooLong(parent *os.Root, component string, lookupErr error) (bool, error) {
	if !errors.Is(lookupErr, windows.ERROR_INVALID_NAME) {
		return false, nil
	}

	dir, err := parent.Open(".")
	if err != nil {
		return false, err
	}
	defer func() { _ = dir.Close() }()
	conn, err := dir.SyscallConn()
	if err != nil {
		return false, err
	}

	var maxComponent uint32
	var volumeErr error
	if err := conn.Control(func(fd uintptr) {
		volumeErr = windows.GetVolumeInformationByHandle(windows.Handle(fd), nil, 0, nil, &maxComponent, nil, nil, 0)
	}); err != nil {
		return false, err
	}
	if volumeErr != nil {
		return false, volumeErr
	}
	if maxComponent == 0 {
		return false, errors.New("volume returned a zero maximum component length")
	}
	return stateComponentExceedsWindowsLimit(component, maxComponent), nil
}

// Windows 구성 요소 제한은 UTF-16 코드 단위 수로 계산한다.
func stateComponentExceedsWindowsLimit(component string, maximum uint32) bool {
	return len(utf16.Encode([]rune(component))) > int(maximum)
}
