package store

import (
	"encoding/json"
	"fmt"
	"time"

	bolt "go.etcd.io/bbolt"
)

// DeviceInfo holds system information collected at startup.
type DeviceInfo struct {
	Hostname    string    `json:"hostname"`
	OS          string    `json:"os"`
	Arch        string    `json:"arch"`
	CPUModel    string    `json:"cpu_model"`
	CPUCores    int       `json:"cpu_cores"`
	TotalRAMMB  uint64    `json:"total_ram_mb"`
	CollectedAt time.Time `json:"collected_at"`
}

// SaveDevice persists device info, keyed by hostname.
func (s *Store) SaveDevice(info DeviceInfo) error {
	data, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("marshal device: %w", err)
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketDevices).Put([]byte(info.Hostname), data)
	})
}
