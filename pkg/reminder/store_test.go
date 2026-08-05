package reminder

import (
	"path/filepath"
	"testing"
	"time"
)

func TestFileStorePersistsDelivery(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	store := NewFileStore(path)
	id := ReminderID("2026-06-08", SlotKey(0, false), 15)
	if store.WasDelivered(id) {
		t.Fatal("fresh store should be empty")
	}
	if err := store.MarkDelivered(id, time.Now()); err != nil {
		t.Fatal(err)
	}
	reopened := NewFileStore(path)
	if !reopened.WasDelivered(id) {
		t.Fatal("delivery should persist")
	}
}

func TestFileStorePruneBefore(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	store := NewFileStore(path)
	oldID := ReminderID("2026-06-07", SlotKey(0, false), 15)
	keepID := ReminderID("2026-06-08", SlotKey(0, false), 15)
	if err := store.MarkDelivered(oldID, time.Now()); err != nil {
		t.Fatal(err)
	}
	if err := store.MarkDelivered(keepID, time.Now()); err != nil {
		t.Fatal(err)
	}
	if err := store.PruneBefore("2026-06-08"); err != nil {
		t.Fatal(err)
	}
	if store.WasDelivered(oldID) {
		t.Fatal("old delivery should be pruned")
	}
	if !store.WasDelivered(keepID) {
		t.Fatal("same-day delivery should remain")
	}
}
