package yara

import (
	"bytes"
	"io/fs"
	"os"
)

const magicReadN = 8

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

func IsEligible(path string, minSize, maxSize int64) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return IsEligibleInfo(info, minSize, maxSize)
}

func IsEligibleInfo(info fs.FileInfo, minSize, maxSize int64) bool {
	size := info.Size()
	return size >= minSize && size <= maxSize
}

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
