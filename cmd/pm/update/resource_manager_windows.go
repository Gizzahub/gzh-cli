//go:build windows
// +build windows

package update

import (
	"golang.org/x/sys/windows"
	"unsafe"
)

// getDiskSpace returns available disk space in bytes for Windows
func getDiskSpace(path string) (uint64, error) {
	pathPtr, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return 0, err
	}
	
	var freeBytesAvailable uint64
	var totalNumberOfBytes uint64
	var totalNumberOfFreeBytes uint64
	
	err = windows.GetDiskFreeSpaceEx(
		pathPtr,
		(*uint64)(unsafe.Pointer(&freeBytesAvailable)),
		(*uint64)(unsafe.Pointer(&totalNumberOfBytes)),
		(*uint64)(unsafe.Pointer(&totalNumberOfFreeBytes)),
	)
	
	if err != nil {
		return 0, err
	}
	
	return freeBytesAvailable, nil
}