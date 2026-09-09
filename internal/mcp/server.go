package mcp

import (
	_ "embed"
	"fmt"

	"github.com/mark3labs/mcp-go/server"
)

//go:embed instructions_writer.md
var instructionsWriter string

//go:embed instructions_scheduler.md
var instructionsScheduler string

// Run starts the stdio MCP server for the role in MERGESIDING_MCP_ROLE.
func Run() error {
	role, err := RoleFromEnv()
	if err != nil {
		return err
	}
	instr := instructionsWriter
	if role == RoleScheduler {
		instr = instructionsScheduler
	}
	s := server.NewMCPServer(
		"mergesiding",
		"0.1.0",
		server.WithInstructions(instr),
	)
	switch role {
	case RoleWriter:
		registerWriterTools(s)
	case RoleScheduler:
		registerSchedulerTools(s)
	default:
		return fmt.Errorf("unsupported role %q", role)
	}
	return server.ServeStdio(s)
}
