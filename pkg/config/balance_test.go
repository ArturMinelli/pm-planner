package config

import "testing"

func TestResolveBalanceCreditMultiplierDefault(t *testing.T) {
	got, err := ResolveBalanceCreditMultiplier(&File{})
	if err != nil {
		t.Fatal(err)
	}
	if got != DefaultBalanceCreditMultiplier {
		t.Fatalf("default multiplier: %v", got)
	}
}

func TestResolveBalanceCreditMultiplierCustom(t *testing.T) {
	got, err := ResolveBalanceCreditMultiplier(&File{BalanceCreditMultiplier: 2.0})
	if err != nil {
		t.Fatal(err)
	}
	if got != 2.0 {
		t.Fatalf("custom multiplier: %v", got)
	}
}

func TestResolveBalanceCreditMultiplierRejectsOutOfRange(t *testing.T) {
	for _, m := range []float64{0.5, 3.1, -1.0} {
		if _, err := ResolveBalanceCreditMultiplier(&File{BalanceCreditMultiplier: m}); err == nil {
			t.Fatalf("expected error for multiplier %v", m)
		}
	}
}

func TestBalanceCreditMultiplierRoundTripsThroughConfig(t *testing.T) {
	path := t.TempDir() + "/config.yaml"
	if err := Save(path, &File{BalanceCreditMultiplier: 2.0}); err != nil {
		t.Fatal(err)
	}
	got, err := Read(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.BalanceCreditMultiplier != 2.0 {
		t.Fatalf("saved multiplier: %v", got.BalanceCreditMultiplier)
	}
}
