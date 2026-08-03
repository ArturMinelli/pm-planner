package config

import (
	"strings"
	"testing"
)

func TestResolvePlannerAnchors_partialAndInvalid(t *testing.T) {
	f := &File{
		Planner: &PlannerAnchors{
			In1:  "09:00",
			Out1: "bad",
			In2:  "14:00",
		},
	}
	got, err := ResolvePlannerAnchors(f)
	if err != nil {
		t.Fatal(err)
	}
	if got[0] != "09:00" {
		t.Fatalf("in1: got %q", got[0])
	}
	if got[1] != "12:00" {
		t.Fatalf("out1 invalid should fallback: got %q", got[1])
	}
	if got[2] != "14:00" {
		t.Fatalf("in2: got %q", got[2])
	}
	if got[3] != "18:00" {
		t.Fatalf("out2 empty should builtin: got %q", got[3])
	}
}

func TestValidatePlannerAnchors_rejectsOrder(t *testing.T) {
	err := ValidatePlannerAnchors([]string{"08:00", "12:00", "12:10", "18:00"})
	if err == nil || !strings.Contains(err.Error(), "15 minutes") {
		t.Fatalf("expected gap error, got %v", err)
	}
}

func TestValidatePlannerAnchors_acceptsCustom(t *testing.T) {
	anchors := []string{"07:30", "11:30", "12:30", "17:00"}
	if err := ValidatePlannerAnchors(anchors); err != nil {
		t.Fatal(err)
	}
}

func TestSave_rejectsInvalidPlanner(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/config.yaml"
	f := &File{
		Planner: &PlannerAnchors{In1: "99:99"},
	}
	err := Save(path, f)
	if err == nil {
		t.Fatal("expected save error for invalid planner time")
	}
}
