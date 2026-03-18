package store

import (
	"encoding/binary"
	"fmt"
	"time"

	bolt "go.etcd.io/bbolt"
)

var (
	bucketDevices = []byte("devices")
	bucketScans   = []byte("scans")
	bucketMatches = []byte("matches")
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
		for _, name := range [][]byte{bucketDevices, bucketScans, bucketMatches} {
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

// itob encodes uint64 as 8-byte big-endian (for sequential bucket keys).
func itob(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, v)
	return b
}
