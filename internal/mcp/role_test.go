package mcp

import (
	"strings"
	"testing"
)

func TestWriterToolsExcludeIntegrate(t *testing.T) {
	for _, n := range ToolNames(RoleWriter) {
		if strings.Contains(n, "integrate") {
			t.Fatalf("writer must not expose %s", n)
		}
	}
}

func TestSchedulerHasIntegrate(t *testing.T) {
	found := false
	for _, n := range ToolNames(RoleScheduler) {
		if n == "mergesiding_integrate" {
			found = true
		}
	}
	if !found {
		t.Fatal("scheduler missing mergesiding_integrate")
	}
}

func TestInvalidRoleError(t *testing.T) {
	t.Setenv(EnvRole, "nope")
	t.Setenv(EnvRoleLegacy, "")
	if _, err := RoleFromEnv(); err == nil {
		t.Fatal("expected error")
	}
}
