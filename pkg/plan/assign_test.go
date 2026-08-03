package plan

import (
	"reflect"
	"testing"
)

func TestAssignStampsToPlannerSlotsLoneMiddayPunchGoesToSaida1(t *testing.T) {
	stamps := []string{"12:28"}
	anchors := []string{"08:00", "12:00", "13:30", "18:00"}

	result := assignStampsToPlannerSlots(stamps, anchors)

	if result[0] != "" {
		t.Errorf("slot 0 (Entrada1): expected empty, got %q", result[0])
	}
	if result[1] != "12:28" {
		t.Errorf("slot 1 (Saída1): expected %q, got %q", "12:28", result[1])
	}
	if result[2] != "" {
		t.Errorf("slot 2 (Entrada2): expected empty, got %q", result[2])
	}
	if result[3] != "" {
		t.Errorf("slot 3 (Saída2): expected empty, got %q", result[3])
	}
}

func TestAssignStampsToPlannerSlotsFourStampsEachCloseToAnchor(t *testing.T) {
	stamps := []string{"08:02", "12:05", "13:35", "17:58"}
	anchors := []string{"08:00", "12:00", "13:30", "18:00"}

	result := assignStampsToPlannerSlots(stamps, anchors)

	expected := []string{"08:02", "12:05", "13:35", "17:58"}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestAssignStampsToPlannerSlotsSixStampsThreeJourneys(t *testing.T) {
	// 6 slots derived from base ["08:00","12:00","13:00","14:00"] + extra journey:
	// slot 4 (entry3) = 14:00 + 30min break = 14:30
	// slot 5 (exit3)  = 14:30 + 3h share    = 17:30
	anchors := []string{"08:00", "12:00", "13:00", "14:00", "14:30", "17:30"}
	stamps := []string{"08:01", "11:59", "13:02", "13:58", "14:29", "17:31"}

	result := assignStampsToPlannerSlots(stamps, anchors)

	expected := []string{"08:01", "11:59", "13:02", "13:58", "14:29", "17:31"}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestAssignStampsToPlannerSlotsFewerStampsThanSlots(t *testing.T) {
	stamps := []string{"13:45"}
	anchors := []string{"08:00", "12:00", "13:30", "18:00"}

	result := assignStampsToPlannerSlots(stamps, anchors)

	if len(result) != 4 {
		t.Fatalf("expected length 4, got %d", len(result))
	}
	if result[2] != "13:45" {
		t.Errorf("stamp near 13:30 anchor should land at slot 2, got result=%v", result)
	}
}

func TestAssignStampsToPlannerSlotsEmptyStampsReturnsEmptySlots(t *testing.T) {
	result := assignStampsToPlannerSlots([]string{}, []string{"08:00", "12:00"})
	for slotIndex, slotValue := range result {
		if slotValue != "" {
			t.Errorf("slot %d: expected empty, got %q", slotIndex, slotValue)
		}
	}
}
