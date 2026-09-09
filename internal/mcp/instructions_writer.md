# mergesiding — multi-agent git worktree isolation with serial integrate

mergesiding isolates parallel AI/coding agents in sibling git worktrees, then
serially rebases, verifies, and merges onto the integration branch under a lock.

Consult these tools for start/ready/status (writer) or integrate (scheduler).
Do NOT edit the human base checkout for parallel work. Do NOT call Cursor
move_agent_to_root / move_agent_to_cloned_root for linked worktrees — keep the
chat root and use absolute worktree paths + Shell working_directory.

## Tool selection by intent

- **Start an isolated parallel task** → `mergesiding_start` (PRIMARY for new work)
- **Task done and worktree clean** → `mergesiding_ready` (closes this task; do not keep editing that worktree)
- **More changes after ready** → wait for integrate (or `mergesiding_abort`), then `mergesiding_start` a **new** slug
- **Inspect progress / blocked / escalate** → `mergesiding_status` (omit slug to list all tasks)
- **Abandon a task** → `mergesiding_abort` (does not remove worktrees; use `mergesiding_cleanup`)
- **Remove DONE/ABORTED git artifacts** → `mergesiding_cleanup` (needs explicit flags)
- **Integrate / merge** → NOT available in writer profile; scheduler owns integrate

## Common chains

- New work: `mergesiding_start` → edit/commit ONLY in returned worktree paths → `mergesiding_ready` → ask human/scheduler to integrate
- Follow-up work: after integrate (or abort), `mergesiding_start` with a new slug / new worktree
- After conflict escalate: resolve rebase in worktree → `mergesiding_ready` → scheduler integrates again
- Diagnose: `mergesiding_status` with slug (read escalate_path / recovery_checklist)

## Anti-patterns

- Do not edit the human base checkout for parallel tasks
- Do not keep editing a worktree after `mergesiding_ready`
- Do not merge into the integration branch yourself
- Do not use `git checkout --ours/--theirs` unless the brief requires it
- Do not call workspace root-switch tools for linked worktrees
- Do not invent integrate calls — this profile has no integrate tools

## Limitations

- Metadata lives under MERGESIDING_HOME (default ~/.mergesiding)
- Default worktrees: `<repo-parent>/.agent-git-worktrees/<repo>/<slug>`
- Relative worktreeRoot resolves against the repo root
- Verify failures and dirty base → blocked; conflicts → awaiting_writer + escalate file
