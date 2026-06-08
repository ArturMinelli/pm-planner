package reminder

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type DeliveryStore interface {
	WasDelivered(id string) bool
	MarkDelivered(id string, at time.Time) error
}

type FileStore struct {
	path string
	mu   sync.Mutex
}

type deliveryState struct {
	Delivered map[string]time.Time `json:"delivered"`
}

func NewFileStore(path string) *FileStore {
	return &FileStore{path: path}
}

func (s *FileStore) WasDelivered(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	state, err := s.read()
	if err != nil {
		return false
	}
	_, ok := state.Delivered[id]
	return ok
}

func (s *FileStore) MarkDelivered(id string, at time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	state, err := s.read()
	if err != nil {
		return err
	}
	if state.Delivered == nil {
		state.Delivered = make(map[string]time.Time)
	}
	state.Delivered[id] = at
	return s.write(state)
}

func (s *FileStore) PruneBefore(date string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	state, err := s.read()
	if err != nil {
		return err
	}
	for id := range state.Delivered {
		if strings.Compare(id, date+"|") < 0 {
			delete(state.Delivered, id)
		}
	}
	return s.write(state)
}

func (s *FileStore) read() (deliveryState, error) {
	var state deliveryState
	b, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			state.Delivered = make(map[string]time.Time)
			return state, nil
		}
		return state, err
	}
	if len(b) == 0 {
		state.Delivered = make(map[string]time.Time)
		return state, nil
	}
	if err := json.Unmarshal(b, &state); err != nil {
		return state, err
	}
	if state.Delivered == nil {
		state.Delivered = make(map[string]time.Time)
	}
	return state, nil
}

func (s *FileStore) write(state deliveryState) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0700); err != nil {
		return err
	}
	b, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(s.path), ".pm-reminders-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	_, writeErr := tmp.Write(b)
	closeErr := tmp.Close()
	if writeErr != nil {
		_ = os.Remove(tmpPath)
		return writeErr
	}
	if closeErr != nil {
		_ = os.Remove(tmpPath)
		return closeErr
	}
	if err := os.Chmod(tmpPath, 0600); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}
	if err := os.Rename(tmpPath, s.path); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}
	return nil
}
