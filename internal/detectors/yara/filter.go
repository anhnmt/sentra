package yara

import (
	"os"
)

const (
	minFileSize = 4 << 10  // 4KB
	maxFileSize = 30 << 20 // 30MB
)

func isEligible(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	size := info.Size()
	return size >= minFileSize && size <= maxFileSize
}
