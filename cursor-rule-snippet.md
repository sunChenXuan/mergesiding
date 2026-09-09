<!-- 简版 rule；完整版见 skills/mergesiding/SKILL.md -->

# mergesiding agent notes (optional)

Daily path: **`start` → `ready` → `integrate`**. Full workflow: [skills/mergesiding/SKILL.md](skills/mergesiding/SKILL.md).

MCP roles differ by tool access (writer / scheduler / full) — see [docs/mcp.md](docs/mcp.md).

- Do **not** edit the human base checkout for parallel work.
- Commit only in printed worktree paths on `task/<slug>`.
- `abort` does not remove worktrees; use `cleanup` after `done` or `aborted`.
- Writers must not merge; scheduler or full role owns `integrate`.
- If using Cursor: do not call `move_agent_to_root` / `move_agent_to_cloned_root` for linked worktrees.
- If using codegraph: default `projectPath` = human base; do not `codegraph init -i` on every start — [docs/codegraph.md](docs/codegraph.md).
