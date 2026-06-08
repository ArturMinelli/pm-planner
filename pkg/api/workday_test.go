package api

import (
	"testing"
	"time"
)

func TestParseRetryAfterSeconds(t *testing.T) {
	if got := parseRetryAfter("42"); got != 42*time.Second {
		t.Fatalf("retry after: got %s", got)
	}
}

func TestParseRetryAfterPastDate(t *testing.T) {
	if got := parseRetryAfter("Mon, 02 Jan 2006 15:04:05 GMT"); got != 0 {
		t.Fatalf("past retry date should be ignored, got %s", got)
	}
}
