// Package punchlog keeps a local registry of clock-ins reconciled against PontoMais.
//
// Known limitation: the reminder daemon process does not run the Syncer; only the
// desktop app / pm-dev process does, so records only sync while one of those runs.
package punchlog

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"

	"pm-cli/pkg/api"
	"pm-cli/pkg/config"
)

// Record statuses and sources (JSON contract shared with the frontend).
const (
	StatusPending     = "pending"
	StatusSynced      = "synced"
	StatusUnconfirmed = "unconfirmed"
	StatusFailed      = "failed"

	SourceApp    = "app"
	SourceRemote = "remote"
)

// Record is one clock-in known locally.
type Record struct {
	LocalID        string     `json:"localId"`
	Date           string     `json:"date"`
	Time           string     `json:"time"`
	CreatedAt      time.Time  `json:"createdAt"`
	Source         string     `json:"source"`
	Status         string     `json:"status"`
	RemoteID       int64      `json:"remoteId,omitempty"`
	Address        string     `json:"address,omitempty"`
	Latitude       float64    `json:"latitude"`
	Longitude      float64    `json:"longitude"`
	LocationSource string     `json:"locationSource,omitempty"`
	Disabled       bool       `json:"disabled,omitempty"`
	Checks         int        `json:"checks"`
	LastCheckedAt  *time.Time `json:"lastCheckedAt,omitempty"`
	SyncedAt       *time.Time `json:"syncedAt,omitempty"`
}

// Data is the persisted file content.
type Data struct {
	Records       []Record   `json:"records"`
	LastSyncAt    *time.Time `json:"lastSyncAt,omitempty"`
	LastSyncError string     `json:"lastSyncError,omitempty"`
}

// Store is a JSON file store; every mutation is a locked read-modify-write.
type Store struct {
	path string
	mu   sync.Mutex
}

// NewStore returns a store at path ("" makes every operation fail).
func NewStore(path string) *Store { return &Store{path: path} }

// DefaultStore returns the store in the config dir (punches.json).
func DefaultStore() *Store {
	dir, err := config.DefaultDir()
	if err != nil {
		return NewStore("")
	}
	return NewStore(filepath.Join(dir, "punches.json"))
}

// NewLocalID returns a random record id.
func NewLocalID() string {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

// Read returns a snapshot of the data.
func (s *Store) Read() (Data, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.load()
}

// Update applies fn to the data and saves when fn reports a change.
func (s *Store) Update(fn func(d *Data) (changed bool)) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	d, err := s.load()
	if err != nil {
		return err
	}
	if !fn(&d) {
		return nil
	}
	return s.save(d)
}

func (s *Store) load() (Data, error) {
	var d Data
	if s.path == "" {
		return d, errors.New("punchlog: no storage path")
	}
	b, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return d, nil
	}
	if err != nil {
		return d, err
	}
	if err := json.Unmarshal(b, &d); err != nil {
		// Keep the corrupt file for inspection and start empty.
		_ = os.Rename(s.path, s.path+".bak")
		return Data{}, nil
	}
	return d, nil
}

func (s *Store) save(d Data) error {
	out, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return err
	}
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".punches-*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	_, werr := tmp.Write(out)
	cerr := tmp.Close()
	if err := errors.Join(werr, cerr, os.Chmod(tmpPath, 0600)); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}
	if err := os.Rename(tmpPath, s.path); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}
	return nil
}

// AddPending saves a new pending app record and returns its localId.
func (s *Store) AddPending(r Record) (string, error) {
	r.LocalID, r.Source, r.Status = NewLocalID(), SourceApp, StatusPending
	return r.LocalID, s.Update(func(d *Data) bool {
		d.Records = append(d.Records, r)
		return true
	})
}

// MarkSynced links localID to the remote punch p. A remote-sourced duplicate of the
// same remoteId (added by a concurrent sync) is dropped so the id stays unique.
func (s *Store) MarkSynced(localID string, p api.Punch, at time.Time) error {
	return s.Update(func(d *Data) bool {
		kept := d.Records[:0]
		for _, r := range d.Records {
			if r.LocalID != localID && r.RemoteID == p.ID && r.Source == SourceRemote {
				continue
			}
			kept = append(kept, r)
		}
		d.Records = kept
		for i := range d.Records {
			r := &d.Records[i]
			if r.LocalID != localID {
				continue
			}
			r.Status, r.RemoteID, r.SyncedAt = StatusSynced, p.ID, &at
			if p.Date != "" && p.Time != "" {
				r.Date, r.Time = p.Date, p.Time
			}
			return true
		}
		return true
	})
}

// MarkFailed marks localID as definitively not registered.
func (s *Store) MarkFailed(localID string) error {
	return s.Update(func(d *Data) bool {
		for i := range d.Records {
			if d.Records[i].LocalID == localID {
				d.Records[i].Status = StatusFailed
				return true
			}
		}
		return false
	})
}
