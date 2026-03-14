package yara

import (
	"io/fs"
	"os"
)

const (
	minFileSize = 4 << 10  // 4KB
	maxFileSize = 30 << 20 // 30MB
)

func IsEligible(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return IsEligibleInfo(info)
}

func IsEligibleInfo(info fs.FileInfo) bool {
	size := info.Size()
	return size >= minFileSize && size <= maxFileSize
}
