//go:build darwin

package main

import (
	"strings"
	"testing"
)

func TestRenderLaunchAgentEscapesExecutable(t *testing.T) {
	got := renderLaunchAgent(`/Applications/PM & Planner.app/Contents/MacOS/pm-desktop`)
	if !strings.Contains(got, "com.pmplanner.reminders") {
		t.Fatalf("missing label:\n%s", got)
	}
	if !strings.Contains(got, "--daemon") {
		t.Fatalf("missing daemon argument:\n%s", got)
	}
	if !strings.Contains(got, "PM &amp; Planner.app") {
		t.Fatalf("executable should be XML escaped:\n%s", got)
	}
}
