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

// LogLevel represents log severity levels.
type LogLevel string

const (
	LevelDebug LogLevel = "debug"
	LevelInfo  LogLevel = "info"
	LevelWarn  LogLevel = "warn"
	LevelError LogLevel = "error"
)

// LogRecord stores a single log entry.
type LogRecord struct {
	ScanID   string         `json:"scan_id,omitempty"`
	Level    LogLevel       `json:"level"`
	Message  string         `json:"message"`
	Fields   map[string]any `json:"fields,omitempty"`
	LoggedAt time.Time      `json:"logged_at"`
}

// ScanSession is an active scan that can record matches and logs incrementally.
type ScanSession struct {
	store  *Store
	Record ScanRecord // exported for external access
}

// ID returns the scan session's unique identifier.
func (sess *ScanSession) ID() string {
	return sess.Record.ID
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
	return &ScanSession{store: s, Record: rec}, nil
}

// RecordMatch appends a match to the scan's match bucket using auto-increment keys.
func (sess *ScanSession) RecordMatch(m MatchRecord) error {
	data, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("marshal match: %w", err)
	}
	return sess.store.db.Update(func(tx *bolt.Tx) error {
		sub, err := tx.Bucket(bucketMatches).CreateBucketIfNotExists([]byte(sess.Record.ID))
		if err != nil {
			return fmt.Errorf("create match bucket: %w", err)
		}
		seq, _ := sub.NextSequence()
		return sub.Put(itob(seq), data)
	})
}

// RecordLog appends a log entry to the logs bucket.
func (sess *ScanSession) RecordLog(level LogLevel, message string, fields map[string]any) error {
	rec := LogRecord{
		ScanID:   sess.Record.ID,
		Level:    level,
		Message:  message,
		Fields:   fields,
		LoggedAt: time.Now(),
	}
	data, err := json.Marshal(rec)
	if err != nil {
		return fmt.Errorf("marshal log: %w", err)
	}
	return sess.store.db.Update(func(tx *bolt.Tx) error {
		sub, err := tx.Bucket(bucketLogs).CreateBucketIfNotExists([]byte(sess.Record.ID))
		if err != nil {
			return fmt.Errorf("create log bucket: %w", err)
		}
		seq, _ := sub.NextSequence()
		return sub.Put(itob(seq), data)
	})
}

// RecordLog stores a log entry without requiring a scan session.
func (s *Store) RecordLog(scanID string, level LogLevel, message string, fields map[string]any) error {
	rec := LogRecord{
		ScanID:   scanID,
		Level:    level,
		Message:  message,
		Fields:   fields,
		LoggedAt: time.Now(),
	}
	data, err := json.Marshal(rec)
	if err != nil {
		return fmt.Errorf("marshal log: %w", err)
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		sub, err := tx.Bucket(bucketLogs).CreateBucketIfNotExists([]byte(scanID))
		if err != nil {
			return fmt.Errorf("create log bucket: %w", err)
		}
		seq, _ := sub.NextSequence()
		return sub.Put(itob(seq), data)
	})
}

// Finish updates the scan record with final statistics and status.
func (sess *ScanSession) Finish(scanned, skipped, matches, errors int64, status string) error {
	sess.Record.FinishedAt = time.Now()
	sess.Record.Scanned = scanned
	sess.Record.Skipped = skipped
	sess.Record.MatchCount = matches
	sess.Record.ErrorCount = errors
	sess.Record.Status = status

	data, err := json.Marshal(sess.Record)
	if err != nil {
		return fmt.Errorf("marshal scan finish: %w", err)
	}
	return sess.store.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketScans).Put([]byte(sess.Record.ID), data)
	})
}
