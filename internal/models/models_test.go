package models

import "testing"

func TestTaskStatusValues(t *testing.T) {
	if StatusActive != "active" {
		t.Fatalf("StatusActive = %q", StatusActive)
	}
	if StatusBlockedPartial != "blocked_partial" {
		t.Fatalf("StatusBlockedPartial = %q", StatusBlockedPartial)
	}
}
