package config

import (
	"fmt"
	"strings"
	"time"

	"pm-cli/pkg/plan"
)

const minAnchorGapMinutes = 15

func (p *PlannerAnchors) anySet() bool {
	if p == nil {
		return false
	}
	return strings.TrimSpace(p.In1) != "" ||
		strings.TrimSpace(p.Out1) != "" ||
		strings.TrimSpace(p.In2) != "" ||
		strings.TrimSpace(p.Out2) != ""
}

// ResolvePlannerAnchors merges config with built-in defaults. Invalid stored values fall back per slot.
func ResolvePlannerAnchors(f *File) ([]string, error) {
	return mergePlannerAnchors(f, false)
}

func mergePlannerAnchors(f *File, strict bool) ([]string, error) {
	builtinArray := plan.BuiltinAnchors()
	out := []string{builtinArray[0], builtinArray[1], builtinArray[2], builtinArray[3]}
	if f == nil || f.Planner == nil {
		return out, nil
	}
	fields := []*string{&f.Planner.In1, &f.Planner.Out1, &f.Planner.In2, &f.Planner.Out2}
	for i, field := range fields {
		s := strings.TrimSpace(*field)
		if s == "" {
			continue
		}
		if !validHHMM(s) {
			if strict {
				return out, fmt.Errorf("invalid time %q for slot %d", s, i+1)
			}
			continue
		}
		out[i] = s
	}
	return out, nil
}

func validHHMM(s string) bool {
	_, err := time.Parse("15:04", s)
	return err == nil
}

func hhmmToMinutes(s string) (int, bool) {
	t, err := time.Parse("15:04", s)
	if err != nil {
		return 0, false
	}
	return t.Hour()*60 + t.Minute(), true
}

// ValidatePlannerAnchors checks four HH:MM values are strictly increasing with at least minAnchorGapMinutes between slots.
func ValidatePlannerAnchors(anchors []string) error {
	if len(anchors) != 4 {
		return fmt.Errorf("expected 4 anchor slots, got %d", len(anchors))
	}
	mins := make([]int, 4)
	for i, s := range anchors {
		if !validHHMM(s) {
			return fmt.Errorf("slot %d: invalid time %q", i+1, s)
		}
		m, ok := hhmmToMinutes(s)
		if !ok {
			return fmt.Errorf("slot %d: invalid time %q", i+1, s)
		}
		mins[i] = m
	}
	for i := 1; i < 4; i++ {
		if mins[i] <= mins[i-1] {
			return fmt.Errorf("times must be strictly increasing (slot %d after slot %d)", i+1, i)
		}
		if mins[i]-mins[i-1] < minAnchorGapMinutes {
			return fmt.Errorf("at least %d minutes required between consecutive slots", minAnchorGapMinutes)
		}
	}
	return nil
}
