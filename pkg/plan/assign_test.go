package plan

import (
	"reflect"
	"testing"
)

func TestAssignStampsToPlannerSlotsFillsSlotsInPunchOrder(t *testing.T) {
	stamps := []string{"09:02", "13:18", "15:45"}
	anchors := []string{"08:00", "12:00", "13:30", "18:00"}

	result := assignStampsToPlannerSlots(stamps, anchors)

	expected := []string{"09:02", "13:18", "15:45", ""}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestAssignStampsToPlannerSlotsLonePunchGoesToFirstSlot(t *testing.T) {
	stamps := []string{"13:45"}
	anchors := []string{"08:00", "12:00", "13:30", "18:00"}

	result := assignStampsToPlannerSlots(stamps, anchors)

	expected := []string{"13:45", "", "", ""}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestAssignStampsToPlannerSlotsFourStampsFillAllSlots(t *testing.T) {
	stamps := []string{"08:02", "12:05", "13:35", "17:58"}
	anchors := []string{"08:00", "12:00", "13:30", "18:00"}

	result := assignStampsToPlannerSlots(stamps, anchors)

	expected := []string{"08:02", "12:05", "13:35", "17:58"}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestAssignStampsToPlannerSlotsSixStampsThreeJourneys(t *testing.T) {
	anchors := []string{"08:00", "12:00", "13:00", "14:00", "14:30", "17:30"}
	stamps := []string{"08:01", "11:59", "13:02", "13:58", "14:29", "17:31"}

	result := assignStampsToPlannerSlots(stamps, anchors)

	expected := []string{"08:01", "11:59", "13:02", "13:58", "14:29", "17:31"}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
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
