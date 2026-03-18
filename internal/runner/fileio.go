package runner

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/edsrzf/mmap-go"
)

const (
	mmapThreshold = 512 * 1024 // 512KB
	batchSize     = 16
)

type scanItem struct {
	path    string
	data    []byte
	cleanup func()
}

func readFile(path string) ([]byte, func(), error) {
	noop := func() {}

	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrPermission) {
			return nil, noop, nil
		}
		return nil, noop, fmt.Errorf("stat %s: %w", path, err)
	}

	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrPermission) {
			return nil, noop, nil
		}
		return nil, noop, fmt.Errorf("open %s: %w", path, err)
	}

	if info.Size() < mmapThreshold {
		defer f.Close()
		data, err := io.ReadAll(f)
		if err != nil {
			return nil, noop, fmt.Errorf("read %s: %w", path, err)
		}
		return data, noop, nil
	}

	mapped, err := mmap.Map(f, mmap.RDONLY, 0)
	if err != nil {
		f.Close()
		return readFallback(path, noop)
	}
	return mapped, func() { mapped.Unmap(); f.Close() }, nil
}

func readFallback(path string, noop func()) ([]byte, func(), error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, noop, fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		return nil, noop, fmt.Errorf("read %s: %w", path, err)
	}
	return data, noop, nil
}
