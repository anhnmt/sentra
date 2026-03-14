package yara

import (
	"bytes"
	"io/fs"
	"os"
)

const (
	minFileSize = 4 << 10  // 4KB
	maxFileSize = 30 << 20 // 30MB
	magicReadN  = 8        // chỉ đọc 8 byte đầu
)

// skipMagic — các format binary không cần YARA scan
var skipMagic = [][]byte{
	{0xFF, 0xD8, 0xFF},       // JPEG
	{0x89, 0x50, 0x4E, 0x47}, // PNG
	{0x47, 0x49, 0x46},       // GIF
	{0x42, 0x4D},             // BMP
	{0x49, 0x44, 0x33},       // MP3
	{0xFF, 0xFB},             // MP3 (no ID3)
	{0x00, 0x00, 0x00, 0x20, 0x66, 0x74, 0x79, 0x70}, // MP4
	{0x1A, 0x45, 0xDF, 0xA3},                         // MKV/WebM
	{0x52, 0x49, 0x46, 0x46},                         // AVI/WAV (RIFF)
	{0x66, 0x74, 0x79, 0x70},                         // MOV
}

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

// HasSkipMagic check 8 byte đầu của data — gọi sau khi đã đọc file
// tránh scan toàn bộ file binary không liên quan
func HasSkipMagic(data []byte) bool {
	if len(data) < magicReadN {
		return false
	}
	header := data[:magicReadN]
	for _, magic := range skipMagic {
		if bytes.HasPrefix(header, magic) {
			return true
		}
	}
	return false
}
