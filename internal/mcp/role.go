// Package mcp implements the mergesiding MCP stdio server (writer/scheduler/full roles).
package mcp

import (
	"fmt"
	"os"
	"strings"
)

// Role is the MCP tool profile.
type Role string

const (
	// RoleWriter exposes start/ready/status/abort/cleanup (no integrate).
	RoleWriter Role = "writer"
	// RoleScheduler exposes status/abort/cleanup/integrate (no start/ready).
	RoleScheduler Role = "scheduler"
	// RoleFull exposes every mergesiding MCP tool (start through integrate).
	RoleFull Role = "full"
)

// EnvRole is the preferred role environment variable.
const EnvRole = "MERGESIDING_MCP_ROLE"

// EnvRoleLegacy is accepted for migration / Cursor configs using AGENT_GIT_MCP_ROLE.
const EnvRoleLegacy = "AGENT_GIT_MCP_ROLE"

// RoleFromEnv reads MERGESIDING_MCP_ROLE or AGENT_GIT_MCP_ROLE.
// Empty/unset defaults to RoleFull. Only writer|scheduler|full are valid.
func RoleFromEnv() (Role, error) {
	v := strings.TrimSpace(os.Getenv(EnvRole))
	if v == "" {
		v = strings.TrimSpace(os.Getenv(EnvRoleLegacy))
	}
	if v == "" {
		return RoleFull, nil
	}
	switch Role(strings.ToLower(v)) {
	case RoleWriter:
		return RoleWriter, nil
	case RoleScheduler:
		return RoleScheduler, nil
	case RoleFull:
		return RoleFull, nil
	default:
		return "", fmt.Errorf("%s must be writer, scheduler, or full (got %q)", EnvRole, v)
	}
}

// ToolNames returns tool names registered for a role.
func ToolNames(role Role) []string {
	shared := []string{
		"mergesiding_status",
		"mergesiding_abort",
		"mergesiding_cleanup",
	}
	switch role {
	case RoleWriter:
		return append([]string{
			"mergesiding_start",
			"mergesiding_ready",
		}, shared...)
	case RoleScheduler:
		return append(append([]string{}, shared...), "mergesiding_integrate")
	case RoleFull:
		return append([]string{
			"mergesiding_start",
			"mergesiding_ready",
		}, append(append([]string{}, shared...), "mergesiding_integrate")...)
	default:
		return nil
	}
}
