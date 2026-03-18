package core

import "context"

type MatchResult struct {
	DetectorName string
	RuleName     string
	Target       string
	Metadata     map[string]any
}

type Detector interface {
	Name() string
	Scan(ctx context.Context, target string, data []byte) ([]MatchResult, error)
}
