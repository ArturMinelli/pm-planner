package config

import (
	"testing"
	"time"
)

func TestResolveMaxDailyExtraWork(t *testing.T) {
	got, err := ResolveMaxDailyExtraWork(&File{})
	if err != nil {
		t.Fatal(err)
	}
	if got != 3*time.Hour {
		t.Fatalf("default cap: %s", got)
	}

	got, err = ResolveMaxDailyExtraWork(&File{MaxDailyExtraMinutes: 90})
	if err != nil {
		t.Fatal(err)
	}
	if got != 90*time.Minute {
		t.Fatalf("custom cap: %s", got)
	}
}

func TestResolveMaxDailyExtraWorkRejectsInvalidValues(t *testing.T) {
	for _, minutes := range []int{-1, 24*60 + 1} {
		if _, err := ResolveMaxDailyExtraWork(&File{MaxDailyExtraMinutes: minutes}); err == nil {
			t.Fatalf("expected error for %d minutes", minutes)
		}
	}
}

func TestMaxDailyExtraMinutesRoundTripsThroughConfig(t *testing.T) {
	path := t.TempDir() + "/config.yaml"
	if err := Save(path, &File{MaxDailyExtraMinutes: 75}); err != nil {
		t.Fatal(err)
	}
	got, err := Read(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.MaxDailyExtraMinutes != 75 {
		t.Fatalf("saved cap: %d", got.MaxDailyExtraMinutes)
	}
}
