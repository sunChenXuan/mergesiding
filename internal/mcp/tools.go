package mcp

import (
	"context"
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/sunChenXuan/mergesiding/internal/abort"
	"github.com/sunChenXuan/mergesiding/internal/cleanup"
	"github.com/sunChenXuan/mergesiding/internal/integrate"
	"github.com/sunChenXuan/mergesiding/internal/paths"
	"github.com/sunChenXuan/mergesiding/internal/ready"
	"github.com/sunChenXuan/mergesiding/internal/start"
	"github.com/sunChenXuan/mergesiding/internal/status"
)

func jsonResult(v any) (*mcp.CallToolResult, error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(string(b)), nil
}

func errResult(err error) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultError(err.Error()), nil
}

func registerWriterTools(s *server.MCPServer) {
	s.AddTool(mcp.NewTool("mergesiding_start",
		mcp.WithDescription("PRIMARY for new parallel work — create isolated task worktrees. Returns worktree paths. Do not edit human base checkout."),
		mcp.WithString("slug", mcp.Required(), mcp.Description("Task slug (branch task/<slug>)")),
		mcp.WithArray("repos", mcp.Required(), mcp.Description("Absolute repo paths"), mcp.Items(map[string]any{"type": "string"})),
		mcp.WithString("worktree_root", mcp.Description("Optional shared worktree root (abs or relative to each repo)")),
		mcp.WithString("writer_id", mcp.Description("Optional writer id for resume hooks")),
		mcp.WithString("brief", mcp.Description("Optional brief path")),
	), handleStart)

	s.AddTool(mcp.NewTool("mergesiding_ready",
		mcp.WithDescription("Mark task ready when worktrees are clean and have commits ahead. Enqueues for scheduler integrate."),
		mcp.WithString("slug", mcp.Required(), mcp.Description("Task slug")),
	), handleReady)

	registerSharedTools(s)
}

func registerSchedulerTools(s *server.MCPServer) {
	registerSharedTools(s)
	s.AddTool(mcp.NewTool("mergesiding_integrate",
		mcp.WithDescription("PRIMARY — integrate one ready queue head, or resume a slug (including blocked_partial)."),
		mcp.WithString("slug", mcp.Description("Optional slug; omit to peek ready queue")),
	), handleIntegrate)

	s.AddTool(mcp.NewTool("mergesiding_integrate_all",
		mcp.WithDescription("Loop integrate until queue empty or stop_batch (awaiting_writer/blocked/blocked_partial/lock)."),
	), handleIntegrateAll)
}

func registerSharedTools(s *server.MCPServer) {
	s.AddTool(mcp.NewTool("mergesiding_status",
		mcp.WithDescription("List tasks or detail one slug (escalate_path, recovery_checklist). Prefer for blocked diagnosis."),
		mcp.WithString("slug", mcp.Description("Optional task slug for detail")),
	), handleStatus)

	s.AddTool(mcp.NewTool("mergesiding_list",
		mcp.WithDescription("Alias of status without slug — list all tasks."),
	), handleList)

	s.AddTool(mcp.NewTool("mergesiding_abort",
		mcp.WithDescription("Abandon a task (active/ready/awaiting_writer/blocked/blocked_partial). Optional git cleanup flags."),
		mcp.WithString("slug", mcp.Required(), mcp.Description("Task slug")),
		mcp.WithBoolean("remove_worktree", mcp.Description("Remove worktrees")),
		mcp.WithBoolean("delete_branch", mcp.Description("Delete task branches")),
	), handleAbort)

	s.AddTool(mcp.NewTool("mergesiding_cleanup",
		mcp.WithDescription("Remove git artifacts for done/aborted tasks. Requires remove_worktree and/or delete_branch."),
		mcp.WithString("slug", mcp.Required(), mcp.Description("Task slug")),
		mcp.WithBoolean("remove_worktree", mcp.Description("Remove worktrees")),
		mcp.WithBoolean("delete_branch", mcp.Description("Delete task branches")),
	), handleCleanup)
}

func handleStart(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	slug, _ := args["slug"].(string)
	var repos []string
	if raw, ok := args["repos"].([]any); ok {
		for _, r := range raw {
			if s, ok := r.(string); ok {
				repos = append(repos, s)
			}
		}
	}
	opt := start.Options{Slug: slug, Repos: repos}
	if v, ok := args["worktree_root"].(string); ok && v != "" {
		opt.SharedRoot = &v
	}
	if v, ok := args["writer_id"].(string); ok && v != "" {
		opt.WriterID = &v
	}
	if v, ok := args["brief"].(string); ok && v != "" {
		opt.BriefPath = &v
	}
	task, err := start.Start(paths.Default(), opt)
	if err != nil {
		return errResult(err)
	}
	return jsonResult(task)
}

func handleReady(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	slug, _ := req.GetArguments()["slug"].(string)
	task, err := ready.MarkReady(paths.Default(), slug)
	if err != nil {
		return errResult(err)
	}
	return jsonResult(task)
}

func handleStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var slugPtr *string
	if v, ok := req.GetArguments()["slug"].(string); ok && v != "" {
		slugPtr = &v
	}
	payload, err := status.Payload(paths.Default(), slugPtr)
	if err != nil {
		return errResult(err)
	}
	return jsonResult(payload)
}

func handleList(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	payload, err := status.Payload(paths.Default(), nil)
	if err != nil {
		return errResult(err)
	}
	return jsonResult(payload)
}

func handleAbort(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	slug, _ := args["slug"].(string)
	rw, _ := args["remove_worktree"].(bool)
	db, _ := args["delete_branch"].(bool)
	task, err := abort.Abort(paths.Default(), slug, rw, db)
	if err != nil {
		return errResult(err)
	}
	return jsonResult(task)
}

func handleCleanup(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	slug, _ := args["slug"].(string)
	rw, _ := args["remove_worktree"].(bool)
	db, _ := args["delete_branch"].(bool)
	actions, err := cleanup.Run(paths.Default(), slug, rw, db)
	if err != nil {
		return errResult(err)
	}
	return jsonResult(map[string]any{"actions": actions})
}

func handleIntegrate(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var slugPtr *string
	if v, ok := req.GetArguments()["slug"].(string); ok && v != "" {
		slugPtr = &v
	}
	out := integrate.IntegrateOne(paths.Default(), slugPtr)
	return jsonResult(out.ToDict())
}

func handleIntegrateAll(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	results, summary := integrate.IntegrateAll(paths.Default())
	mapped := make([]map[string]any, 0, len(results))
	for _, r := range results {
		mapped = append(mapped, r.ToDict())
	}
	return jsonResult(map[string]any{"results": mapped, "summary": summary})
}