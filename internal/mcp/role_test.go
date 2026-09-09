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
		if n == "mergesiding_integrate_all" || n == "mergesiding_list" {
			t.Fatalf("scheduler must not expose removed tool %s", n)
		}
	}
	if !found {
		t.Fatal("scheduler missing mergesiding_integrate")
	}
}

func TestFullHasStartReadyIntegrate(t *testing.T) {
	have := map[string]bool{}
	for _, n := range ToolNames(RoleFull) {
		have[n] = true
		if n == "mergesiding_list" || n == "mergesiding_integrate_all" {
			t.Fatalf("full must not expose removed tool %s", n)
		}
	}
	for _, need := range []string{"mergesiding_start", "mergesiding_ready", "mergesiding_integrate"} {
		if !have[need] {
			t.Fatalf("full missing %s", need)
		}
	}
}

func TestWriterAndSchedulerHaveNoList(t *testing.T) {
	for _, role := range []Role{RoleWriter, RoleScheduler} {
		for _, n := range ToolNames(role) {
			if n == "mergesiding_list" {
				t.Fatalf("%s must not expose list", role)
			}
		}
	}
}

func TestRoleFromEnvDefaultFull(t *testing.T) {
	t.Setenv(EnvRole, "")
	t.Setenv(EnvRoleLegacy, "")
	r, err := RoleFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if r != RoleFull {
		t.Fatalf("got %q want full", r)
	}
}

func TestRoleFromEnvFull(t *testing.T) {
	for _, v := range []string{"full", "FULL"} {
		t.Setenv(EnvRole, v)
		t.Setenv(EnvRoleLegacy, "")
		r, err := RoleFromEnv()
		if err != nil {
			t.Fatalf("%s: %v", v, err)
		}
		if r != RoleFull {
			t.Fatalf("%s: got %q want full", v, r)
		}
	}
}

func TestInvalidRoleError(t *testing.T) {
	t.Setenv(EnvRole, "nope")
	t.Setenv(EnvRoleLegacy, "")
	if _, err := RoleFromEnv(); err == nil {
		t.Fatal("expected error")
	}
}
