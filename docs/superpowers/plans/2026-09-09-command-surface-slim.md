# Slim command / MCP surface Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Remove redundant `list` / `integrate_all` surface, split abort vs cleanup, add MCP role `full` (default when role env unset; no `all` role alias), and align EN/ZH docs + Skill + snippet to the same permissions-only narrative.

**Architecture:** Keep CLI and dual-ish MCP registration; extend `Role` with `full`, default `RoleFromEnv` to `full` when unset, register writerâˆªscheduler tools for `full`. Slim MCP tool list and abort API so disk cleanup only goes through `cleanup`. Docs state role permissions without pitching a preferred role.

**Tech Stack:** Go 1.x, mark3labs/mcp-go, existing `go test ./...`, Markdown docs under `docs/` and `skills/`.

## Global Constraints

- Spec: `docs/superpowers/specs/2026-09-09-command-surface-slim-design.md`
- Breaking OK: delete `list`, `mergesiding_list`, `mergesiding_integrate_all`; abort rejects cleanup flags
- Canonical full role name: `full`; unset/empty role env â†?`full`; value `all` aliases to `full`
- Integrate tool boolean argument remains named `all` (not the role)
- Docs: permissions only â€?no â€œprefer full/splitâ€?marketing copy; stating unsetâ†’`full` as fact is OK
- Do not change status machine, multi-repo, verify, worktree layout, or auto-cleanup after integrate
- Method-level comments on new exported funcs/consts per project rules

---

## File map

| File | Responsibility |
|------|----------------|
| `internal/mcp/role.go` | Role enum, `RoleFromEnv`, `ToolNames` |
| `internal/mcp/role_test.go` | Role/default/tool-list tests |
| `internal/mcp/tools.go` | MCP tool registration + handlers |
| `internal/mcp/server.go` | Role â†?register + instructions |
| `internal/mcp/instructions_*.md` | Host instructions per role |
| `internal/abort/abort.go` | Abort without cleanup flags |
| `internal/abort/abort_test.go` | Abort unit tests (new) |
| `internal/cli/cli.go` | CLI: drop `list`; abort flag rejection; help |
| `README.md`, `README.zh-CN.md`, `docs/mcp.md`, `docs/mcp.zh-CN.md`, `skills/mergesiding/SKILL.md`, `cursor-rule-snippet.md` | User-facing alignment |

---

### Task 1: MCP role `full` + default unset + ToolNames slim

**Files:**
- Modify: `internal/mcp/role.go`
- Modify: `internal/mcp/role_test.go`

**Interfaces:**
- Produces: `RoleFull Role = "full"`; `RoleFromEnv() (Role, error)` returns `RoleFull` when both env vars empty; `"all"` maps to `RoleFull`; `ToolNames` has no `mergesiding_list` / `mergesiding_integrate_all`; `ToolNames(RoleFull)` = writer tools without list + `mergesiding_integrate`

- [ ] **Step 1: Write the failing tests**

Replace/extend `internal/mcp/role_test.go`:

```go
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
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/mcp/ -count=1 -run "TestFull|TestRoleFromEnvDefault|TestRoleFromEnvFull|TestWriterAndScheduler|TestSchedulerHasIntegrate"`

Expected: FAIL (RoleFull undefined and/or ToolNames still lists removed tools / empty env errors)

- [ ] **Step 3: Implement role.go**

```go
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
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/mcp/ -count=1`

Expected: PASS for role tests (server/tools may still compile; if tools.go still references old names, fix only in later tasks â€?role_test should pass)

- [ ] **Step 5: Commit**

```bash
git add internal/mcp/role.go internal/mcp/role_test.go
git commit -m "Add MCP full role (default when unset) and slim ToolNames."
```

---

### Task 2: Slim MCP tools + register `full`

**Files:**
- Modify: `internal/mcp/tools.go`
- Modify: `internal/mcp/server.go`
- Create: `internal/mcp/instructions_full.md`
- Modify: `internal/mcp/instructions_writer.md`
- Modify: `internal/mcp/instructions_scheduler.md`

**Interfaces:**
- Consumes: `RoleFull`, `ToolNames` from Task 1
- Produces: no `mergesiding_list` / `mergesiding_integrate_all`; `mergesiding_integrate` optional bool `all`; `mergesiding_abort` without cleanup bools; `registerFullTools` = writer + integrate

- [ ] **Step 1: Update tools.go registration**

In `registerWriterTools`: keep start/ready; call `registerSharedTools` (shared must NOT register list).

In `registerSharedTools`: only status, abort, cleanup â€?delete `mergesiding_list` tool and `handleList`.

In `registerSchedulerTools`: shared + integrate with optional `all`:

```go
s.AddTool(mcp.NewTool("mergesiding_integrate",
	mcp.WithDescription("PRIMARY â€?integrate one ready queue head, or resume a slug (including blocked_partial). Set all=true to drain the ready queue until empty or stop_batch."),
	mcp.WithString("slug", mcp.Description("Optional slug; omit to peek ready queue")),
	mcp.WithBoolean("all", mcp.Description("If true, loop integrate like former integrate_all (ignore slug)")),
), handleIntegrate)
```

Delete `mergesiding_integrate_all` tool and `handleIntegrateAll` as a separate registration (fold into handleIntegrate).

Update `handleIntegrate`:

```go
func handleIntegrate(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	if all, _ := args["all"].(bool); all {
		results, summary := integrate.IntegrateAll(paths.Default())
		mapped := make([]map[string]any, 0, len(results))
		for _, r := range results {
			mapped = append(mapped, r.ToDict())
		}
		return jsonResult(map[string]any{"results": mapped, "summary": summary})
	}
	var slugPtr *string
	if v, ok := args["slug"].(string); ok && v != "" {
		slugPtr = &v
	}
	out := integrate.IntegrateOne(paths.Default(), slugPtr)
	return jsonResult(out.ToDict())
}
```

Update `handleAbort` â€?stop reading cleanup booleans; call `abort.Abort(paths.Default(), slug)` once Task 3 changes the signature. **For this task only**, if Abort still has 4 params, pass `false, false` temporarily OR do Task 3 first in the same session before compiling â€?prefer completing Task 3 abort signature before finishing this stepâ€™s compile.

Abort tool schema: remove `remove_worktree` / `delete_branch` properties.

Add:

```go
func registerFullTools(s *server.MCPServer) {
	registerWriterTools(s)
	// Writer already registered shared; add integrate only once.
	s.AddTool(mcp.NewTool("mergesiding_integrate",
		mcp.WithDescription("PRIMARY â€?integrate one ready queue head, or resume a slug (including blocked_partial). Set all=true to drain the ready queue until empty or stop_batch."),
		mcp.WithString("slug", mcp.Description("Optional slug; omit to peek ready queue")),
		mcp.WithBoolean("all", mcp.Description("If true, loop integrate like former integrate_all (ignore slug)")),
	), handleIntegrate)
}
```

**Important:** `registerWriterTools` currently calls `registerSharedTools`. If `registerFullTools` calls `registerWriterTools` then adds integrate, do **not** also call `registerSchedulerTools` (would double-register shared tools). Refactor if needed:

```go
func registerWriterTools(s *server.MCPServer) {
	// start + ready + shared
}
func registerSchedulerTools(s *server.MCPServer) {
	registerSharedTools(s)
	registerIntegrateTool(s)
}
func registerFullTools(s *server.MCPServer) {
	registerWriterTools(s)
	registerIntegrateTool(s)
}
```

Extract `registerIntegrateTool` to avoid duplicating the tool definition.

- [ ] **Step 2: Update server.go**

```go
//go:embed instructions_full.md
var instructionsFull string

func Run() error {
	role, err := RoleFromEnv()
	if err != nil {
		return err
	}
	instr := instructionsFull
	switch role {
	case RoleWriter:
		instr = instructionsWriter
	case RoleScheduler:
		instr = instructionsScheduler
	case RoleFull:
		instr = instructionsFull
	}
	s := server.NewMCPServer("mergesiding", "0.1.0", server.WithInstructions(instr))
	switch role {
	case RoleWriter:
		registerWriterTools(s)
	case RoleScheduler:
		registerSchedulerTools(s)
	case RoleFull:
		registerFullTools(s)
	default:
		return fmt.Errorf("unsupported role %q", role)
	}
	return server.ServeStdio(s)
}
```

- [ ] **Step 3: Instructions markdown**

Create `instructions_full.md`: combine writer+scheduler intent selection; use `mergesiding_status` only (no list); integrate via `mergesiding_integrate` / `all=true`; state permissions factually.

Update writer/scheduler instruction files: remove `mergesiding_list` and `mergesiding_integrate_all`; point batch integrate to `mergesiding_integrate` with `all=true`; note abort does not remove worktrees (use cleanup).

- [ ] **Step 4: Run tests**

Run: `go test ./internal/mcp/ -count=1`

Expected: PASS (may fail until Task 3 if Abort signature already changed â€?order Task 3 next if compile errors)

- [ ] **Step 5: Commit**

```bash
git add internal/mcp/tools.go internal/mcp/server.go internal/mcp/instructions_writer.md internal/mcp/instructions_scheduler.md internal/mcp/instructions_full.md
git commit -m "Slim MCP tools; register full role; fold integrate_all into all flag."
```

---

### Task 3: Abort status-only + CLI reject cleanup flags

**Files:**
- Modify: `internal/abort/abort.go`
- Create: `internal/abort/abort_test.go`
- Modify: `internal/cli/cli.go` (`cmdAbort`, help, remove `list`)
- Modify: `internal/mcp/tools.go` (`handleAbort` call site) if still using old signature

**Interfaces:**
- Produces: `func Abort(p paths.Paths, taskSlug string) (*models.TaskRecord, error)` â€?no cleanup params; does not call `cleanup.CleanupTask`

- [ ] **Step 1: Write failing abort tests**

`internal/abort/abort_test.go` â€?use a temp MERGESIDING_HOME / store pattern from existing `store` or `start` tests if available. Minimal approach:

```go
package abort_test

import (
	"testing"

	"github.com/sunChenXuan/mergesiding/internal/abort"
	"github.com/sunChenXuan/mergesiding/internal/models"
	"github.com/sunChenXuan/mergesiding/internal/paths"
	"github.com/sunChenXuan/mergesiding/internal/store"
)

// Use t.TempDir() for MERGESIDING_HOME via paths helper if the codebase supports
// constructing paths.Paths with custom Home; otherwise follow store_test patterns.
```

Inspect `internal/store/store_test.go` and `internal/paths/paths.go` for how tests set home. Mirror that: create an `active` task JSON, call `Abort`, assert status `aborted`, assert worktree paths untouched (no need to create real git if Abort no longer calls cleanup â€?just assert no error and status).

Also add a CLI-level check in a small test or manual step: if no CLI test harness exists, add table-driven logic in `cmdAbort` and verify via `go test` only on abort package; then manually run CLI in Step 4.

Simpler abort unit test without git:

```go
func TestAbortSetsAbortedWithoutCleanupArgs(t *testing.T) {
	// Arrange: Paths with temp home, Save TaskRecord{Slug:"t", Status:active, Repos:[]}
	// Act: abort.Abort(p, "t")
	// Assert: status aborted; function signature has 2 params only
}
```

Copy temp-home setup from `internal/store/store_test.go`.

- [ ] **Step 2: Run test â€?expect fail if signature still has cleanup or cleanup still invoked**

Run: `go test ./internal/abort/ -count=1`

- [ ] **Step 3: Change Abort API**

```go
// Abort abandons a task: set ABORTED and dequeue. Does not remove worktrees or branches;
// callers must use cleanup after abort when disk cleanup is desired.
func Abort(p paths.Paths, taskSlug string) (*models.TaskRecord, error) {
	// same status switch + queue remove + save as today
	// DELETE the block: if removeWorktree || deleteBranch { cleanup.CleanupTask(...) }
	// REMOVE cleanup import if unused
}
```

Update all callers (`cli.cmdAbort`, `mcp.handleAbort`) to `abort.Abort(p, slug)`.

- [ ] **Step 4: CLI â€?reject old flags; drop list**

In `Execute` switch: change `case "status", "list":` to `case "status":` only; remove the `if cmd == "list"` branch.

In `cmdAbort`:

```go
func cmdAbort(p paths.Paths, args []string) int {
	if hasFlag(args, "--remove-worktree") || hasFlag(args, "--delete-branch") {
		fmt.Fprintln(os.Stderr, "abort no longer removes git artifacts; use: mergesiding cleanup --slug <slug> --remove-worktree and/or --delete-branch")
		return 2
	}
	slug := flagValue(args, "--slug")
	// ...
	task, err := abort.Abort(p, slug)
	// ...
}
```

Update `printHelp` â€?remove `list` line; abort line without cleanup flags.

- [ ] **Step 5: Run tests**

Run: `go test ./internal/abort/ ./internal/cli/ ./internal/mcp/ ./... -count=1`

Expected: PASS (packages without tests still compile)

Manual smoke (optional but recommended):

```bash
go run ./cmd/mergesiding abort --slug x --remove-worktree
```

Expected: exit 2 and the stderr message (slug may also fail load â€?flag check should run first)

- [ ] **Step 6: Commit**

```bash
git add internal/abort/abort.go internal/abort/abort_test.go internal/cli/cli.go internal/mcp/tools.go
git commit -m "Make abort status-only; reject cleanup flags; drop CLI list."
```

---

### Task 4: Align docs, Skill, snippet

**Files:**
- Modify: `README.md`, `README.zh-CN.md`
- Modify: `docs/mcp.md`, `docs/mcp.zh-CN.md`
- Modify: `skills/mergesiding/SKILL.md`
- Modify: `cursor-rule-snippet.md`

**Interfaces:**
- Consumes: role matrix + abort/cleanup contract from spec

- [ ] **Step 1: Update MCP docs (EN + ZH)**

For each role section, list tools only:

- writer: start, ready, status, abort, cleanup
- scheduler: status, abort, cleanup, integrate (`all` arg for batch)
- full: all of the above; note unset/empty `MERGESIDING_MCP_ROLE` â†?full; `all` accepted as alias of `full`

Remove integrate_all / list mentions. State abort does not clean disk; cleanup after done/aborted.

Config examples may show writer+scheduler and/or a single server without role env / with `full` â€?without â€œrecommendedâ€?wording.

- [ ] **Step 2: Update README EN + ZH**

Same tool narrative; command examples without `list`; abort without cleanup flags; mention cleanup separately.

- [ ] **Step 3: Update Skill + snippet**

Skill: MCP roles differ by tool access (link mcp docs); writer has no integrate; full has all tools; abort then optional cleanup; no list/integrate_all.

Snippet: short pointer to Skill + `start â†?ready â†?integrate` + one line on MCP role tool access â†?`docs/mcp.md`.

- [ ] **Step 4: Search verification**

Run (PowerShell):

```powershell
Select-String -Path README.md,README.zh-CN.md,docs\mcp.md,docs\mcp.zh-CN.md,skills\mergesiding\SKILL.md,cursor-rule-snippet.md,internal\mcp\instructions_*.md -Pattern "integrate_all|mergesiding_list|mergesiding list|abort.*remove-worktree|prefer full|recommended.*role" -CaseSensitive:$false
```

Expected: no hits that recommend removed surface or preferred-role marketing (factual â€œunset defaults to fullâ€?is OK).

- [ ] **Step 5: Full test suite**

Run: `go test ./... -count=1`

Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add README.md README.zh-CN.md docs/mcp.md docs/mcp.zh-CN.md skills/mergesiding/SKILL.md cursor-rule-snippet.md
git commit -m "Align docs and Skill with slim surface and full MCP role."
```

---

### Task 5: Spec/plan consistency commit (if needed)

**Files:**
- Modify: `docs/superpowers/specs/2026-09-09-command-surface-slim-design.md` (already includes `full` + default)
- This plan file

- [ ] **Step 1: Confirm spec and code agree on `full` / unset default (no all alias)**

- [ ] **Step 2: Commit any remaining doc/spec drift**

```bash
git add docs/superpowers/specs/2026-09-09-command-surface-slim-design.md docs/superpowers/plans/2026-09-09-command-surface-slim.md
git commit -m "Add implementation plan for command surface slim-down."
```

---

## Spec coverage checklist

| Spec item | Task |
|-----------|------|
| Remove CLI list / MCP list | 2, 3 |
| Remove integrate_all; integrate `all` arg | 2 |
| abort status-only; CLI reject flags; MCP schema drop | 2, 3 |
| Role `full` + unset default (no all alias) | 1, 2 |
| Docs permissions-only + EN/ZH + Skill + snippet | 4 |
| Tests for roles/abort/CLI | 1, 3, 4 |
| No status machine / auto-cleanup / dual-role removal | honored (non-goals) |

## Self-review notes

- No TBD placeholders in tasks
- Integrate arg `all` vs role `full` distinguished throughout
- `registerFullTools` must not double-register shared tools
- Abort signature change coordinated between Tasks 2â€? (implement Abort API before MCP compile finishes)
