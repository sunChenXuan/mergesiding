// Package mcp implements the dual-role mergesiding MCP stdio server.
package mcp

import (
	"fmt"
	"os"
	"strings"
)

// Role is the MCP tool profile.
type Role string

const (
	RoleWriter    Role = "writer"
	RoleScheduler Role = "scheduler"
)

// EnvRole is the preferred role environment variable.
const EnvRole = "MERGESIDING_MCP_ROLE"

// EnvRoleLegacy is accepted for migration / Cursor configs using AGENT_GIT_MCP_ROLE.
const EnvRoleLegacy = "AGENT_GIT_MCP_ROLE"

// RoleFromEnv reads MERGESIDING_MCP_ROLE or AGENT_GIT_MCP_ROLE.
func RoleFromEnv() (Role, error) {
	v := strings.TrimSpace(os.Getenv(EnvRole))
	if v == "" {
		v = strings.TrimSpace(os.Getenv(EnvRoleLegacy))
	}
	switch Role(strings.ToLower(v)) {
	case RoleWriter:
		return RoleWriter, nil
	case RoleScheduler:
		return RoleScheduler, nil
	default:
		return "", fmt.Errorf("%s must be writer or scheduler (got %q)", EnvRole, v)
	}
}

// ToolNames returns tool names registered for a role.
func ToolNames(role Role) []string {
	switch role {
	case RoleWriter:
		return []string{
			"mergesiding_start",
			"mergesiding_ready",
			"mergesiding_status",
			"mergesiding_list",
			"mergesiding_abort",
			"mergesiding_cleanup",
		}
	case RoleScheduler:
		return []string{
			"mergesiding_status",
			"mergesiding_list",
			"mergesiding_abort",
			"mergesiding_cleanup",
			"mergesiding_integrate",
			"mergesiding_integrate_all",
		}
	default:
		return nil
	}
}
