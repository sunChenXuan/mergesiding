# mergesiding — full profile (writer + integrate)

mergesiding isolates parallel AI/coding agents in sibling git worktrees, then
serially rebases, verifies, and merges onto the integration branch under a lock.

This profile exposes every MCP tool: start/ready (writer) plus integrate (scheduler).
Do NOT edit the human base checkout for parallel work. Do NOT call Cursor
move_agent_to_root / move_agent_to_cloned_root for linked worktrees — keep the
chat root and use absolute worktree paths + Shell working_directory.

## Tool selection by intent

- **Start an isolated parallel task** → `mergesiding_start` (PRIMARY for new work)
- **Task done and worktree clean** → `mergesiding_ready`
- **Integrate next ready task** → `mergesiding_integrate` (PRIMARY for merge)
- **Drain queue until empty or stop** → `mergesiding_integrate` with `all=true`
- **Resume BLOCKED_PARTIAL / specific slug** → `mergesiding_integrate` with slug
- **Inspect / blocked / escalate** → `mergesiding_status` (omit slug to list all)
- **Abandon a task** → `mergesiding_abort` (does not remove worktrees; use cleanup)
- **Remove DONE/ABORTED git artifacts** → `mergesiding_cleanup` (needs explicit flags)

## Common chains

- New work: `mergesiding_start` → edit/commit ONLY in returned worktree paths → `mergesiding_ready` → `mergesiding_integrate`
- Batch integrate: `mergesiding_integrate` with `all=true` after writers ready
- After conflict escalate: resolve rebase in worktree → `mergesiding_ready` → `mergesiding_integrate`
- Diagnose: `mergesiding_status` with slug (read escalate_path / recovery_checklist)

## Anti-patterns

- Do not edit the human base checkout for parallel tasks
- Do not merge into the integration branch yourself outside these tools
- Do not use `git checkout --ours/--theirs` unless the brief requires it
- Do not call workspace root-switch tools for linked worktrees
- Do not expect `mergesiding_abort` to remove worktrees — use `mergesiding_cleanup`
- If using codegraph while editing a worktree: default `projectPath` = human base checkout; do not `codegraph init -i` on every start (see docs/codegraph.md)

## Limitations

- Metadata lives under MERGESIDING_HOME (default ~/.mergesiding)
- Default worktrees: `<repo-parent>/.agent-git-worktrees/<repo>/<slug>`
- Relative worktreeRoot resolves against the repo root
- Verify failures and dirty base → blocked; conflicts → awaiting_writer + escalate file
- One integrate lock; concurrent integrate fails with lock held
- stop_batch on awaiting_writer / blocked / blocked_partial
