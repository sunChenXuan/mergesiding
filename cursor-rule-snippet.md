<!-- 简版 rule；完整版见 skills/mergesiding/SKILL.md -->

# mergesiding agent notes (optional)

Works with CLI or any MCP host (`mergesiding-writer` / `mergesiding-scheduler`).

- Prefer MCP tools when the host has them configured; otherwise use `mergesiding` CLI.
- Do **not** edit the human base checkout for parallel work.
- Default worktrees: `<repo-parent>/.agent-git-worktrees/<repo-name>/<slug>`
- Commit only in printed worktree paths on `task/<slug>`.
- When done and clean: `mergesiding ready --slug <slug>`
- Remind the user to run `mergesiding integrate` (scheduler/human); writers must not merge into the integration branch themselves.
- Writer MCP has no integrate tools.
- On escalate under `MERGESIDING_HOME/escalate`: resolve conflicts only, continue rebase, then `ready` again.
- Do not use `git checkout --ours/--theirs` unless the brief explicitly requires it.
- If using Cursor: do not call `move_agent_to_root` / `move_agent_to_cloned_root` for linked worktrees.
