//go:build !windows

package main

import (
	"fmt"

	"golang.org/x/sys/unix"
)

func getFreeDiskSpace(path string) (uint64, error) {
	var stat unix.Statfs_t
	if err := unix.Statfs(path, &stat); err != nil {
		return 0, fmt.Errorf("failed to call Statfs: %v", err)
	}

	var bavail = stat.Bavail
	// nolint:staticcheck
	if stat.Bavail < 0 {
		bavail = 0
	}
	//nolint:unconvert
	return uint64(bavail) * uint64(stat.Bsize), nil
}
