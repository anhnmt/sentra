package store

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"time"

	bolt "go.etcd.io/bbolt"
)

var (
	bucketDevices = []byte("devices")
	bucketScans   = []byte("scans")
	bucketMatches = []byte("matches")
	bucketLogs    = []byte("logs")
)

// Store wraps a bbolt database for persistent scan history.
type Store struct {
	db *bolt.DB
}

// Open opens (or creates) the bbolt database at path.
func Open(path string) (*Store, error) {
	db, err := bolt.Open(path, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, fmt.Errorf("open db %s: %w", path, err)
	}
	s := &Store{db: db}
	if err := s.initBuckets(); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) initBuckets() error {
	return s.db.Update(func(tx *bolt.Tx) error {
		for _, name := range [][]byte{bucketDevices, bucketScans, bucketMatches, bucketLogs} {
			if _, err := tx.CreateBucketIfNotExists(name); err != nil {
				return fmt.Errorf("create bucket %s: %w", name, err)
			}
		}
		return nil
	})
}

// Close closes the underlying database.
func (s *Store) Close() error {
	return s.db.Close()
}

// GetScan retrieves a scan record by ID.
func (s *Store) GetScan(id string) (*ScanRecord, error) {
	var rec *ScanRecord
	err := s.db.View(func(tx *bolt.Tx) error {
		data := tx.Bucket(bucketScans).Get([]byte(id))
		if data == nil {
			return nil
		}
		return json.Unmarshal(data, &rec)
	})
	return rec, err
}

// GetAllScanIDs returns all scan IDs sorted by most recent first.
func (s *Store) GetAllScanIDs() ([]string, error) {
	var ids []string
	err := s.db.View(func(tx *bolt.Tx) error {
		cur := tx.Bucket(bucketScans).Cursor()
		for k, _ := cur.Last(); k != nil; k, _ = cur.Prev() {
			ids = append(ids, string(k))
		}
		return nil
	})
	return ids, err
}

// GetMatches retrieves all matches for a scan.
func (s *Store) GetMatches(scanID string) ([]MatchRecord, error) {
	var matches []MatchRecord
	err := s.db.View(func(tx *bolt.Tx) error {
		sub := tx.Bucket(bucketMatches).Bucket([]byte(scanID))
		if sub == nil {
			return nil
		}
		cur := sub.Cursor()
		for k, v := cur.First(); k != nil; k, v = cur.Next() {
			var m MatchRecord
			if err := json.Unmarshal(v, &m); err != nil {
				return err
			}
			matches = append(matches, m)
		}
		return nil
	})
	return matches, err
}

// GetDevice retrieves device info by hostname.
func (s *Store) GetDevice(hostname string) (*DeviceInfo, error) {
	var dev *DeviceInfo
	err := s.db.View(func(tx *bolt.Tx) error {
		data := tx.Bucket(bucketDevices).Get([]byte(hostname))
		if data == nil {
			return nil
		}
		return json.Unmarshal(data, &dev)
	})
	return dev, err
}

// itob encodes uint64 as 8-byte big-endian (for sequential bucket keys).
func itob(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, v)
	return b
}
