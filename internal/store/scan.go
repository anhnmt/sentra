package store

import (
	"encoding/json"
	"fmt"
	"time"

	bolt "go.etcd.io/bbolt"
)

// ScanRecord captures metadata for a single scan session.
type ScanRecord struct {
	ID         string    `json:"id"`
	StartedAt  time.Time `json:"started_at"`
	FinishedAt time.Time `json:"finished_at,omitempty"`
	Target     string    `json:"target"`
	RulesDir   string    `json:"rules_dir"`
	Workers    int       `json:"workers"`
	Scanned    int64     `json:"scanned"`
	Skipped    int64     `json:"skipped"`
	MatchCount int64     `json:"match_count"`
	ErrorCount int64     `json:"error_count"`
	Status     string    `json:"status"` // running | completed | canceled | error
}

// MatchRecord stores a single YARA detection within a scan.
type MatchRecord struct {
	ScanID       string         `json:"scan_id"`
	DetectorName string         `json:"detector_name"`
	RuleName     string         `json:"rule_name"`
	Target       string         `json:"target"`
	Metadata     map[string]any `json:"metadata,omitempty"`
	DetectedAt   time.Time      `json:"detected_at"`
}

// ScanSession is an active scan that can record matches incrementally.
type ScanSession struct {
	store  *Store
	record ScanRecord
}

// BeginScan persists a new scan record with status "running" and returns a ScanSession.
func (s *Store) BeginScan(rec ScanRecord) (*ScanSession, error) {
	data, err := json.Marshal(rec)
	if err != nil {
		return nil, fmt.Errorf("marshal scan: %w", err)
	}
	if err := s.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketScans).Put([]byte(rec.ID), data)
	}); err != nil {
		return nil, err
	}
	return &ScanSession{store: s, record: rec}, nil
}

// RecordMatch appends a match to the scan's match bucket using auto-increment keys.
func (sess *ScanSession) RecordMatch(m MatchRecord) error {
	data, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("marshal match: %w", err)
	}
	return sess.store.db.Update(func(tx *bolt.Tx) error {
		sub, err := tx.Bucket(bucketMatches).CreateBucketIfNotExists([]byte(sess.record.ID))
		if err != nil {
			return fmt.Errorf("create match bucket: %w", err)
		}
		seq, _ := sub.NextSequence()
		return sub.Put(itob(seq), data)
	})
}

// Finish updates the scan record with final statistics and status.
func (sess *ScanSession) Finish(scanned, skipped, matches, errors int64, status string) error {
	sess.record.FinishedAt = time.Now()
	sess.record.Scanned = scanned
	sess.record.Skipped = skipped
	sess.record.MatchCount = matches
	sess.record.ErrorCount = errors
	sess.record.Status = status

	data, err := json.Marshal(sess.record)
	if err != nil {
		return fmt.Errorf("marshal scan finish: %w", err)
	}
	return sess.store.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketScans).Put([]byte(sess.record.ID), data)
	})
}
