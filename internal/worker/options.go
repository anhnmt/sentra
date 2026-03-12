package worker

import "runtime"

type Options struct {
	Size int
}

func DefaultOptions() *Options {
	return &Options{
		Size: runtime.NumCPU() * 2,
	}
}
