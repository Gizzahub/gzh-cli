//go:build !windows

package update

import (
	"syscall"
)

// getDiskSpace returns available disk space in bytes for Unix-like systems.
func getDiskSpace(path string) (uint64, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return 0, err
	}

	return stat.Bavail * uint64(stat.Bsize), nil
}
