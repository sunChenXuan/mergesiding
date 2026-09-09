---
name: mergesiding
description: >-
  Use when multiple agents may edit the same git repo in parallel, or when
  starting/readying/integrating isolated mergesiding tasks (CLI or any MCP host).
---

# mergesiding

## When to use

- Two or more agents will change the same repository
- User asks to start an isolated task worktree / mark ready / integrate

## MCP vs CLI

If the host has mergesiding MCP configured, use those tools; otherwise CLI: `mergesiding`.

MCP roles differ by **tool access** only (see [docs/mcp.md](../../docs/mcp.md)):

- **writer** — start, ready, status, abort, cleanup (no integrate)
- **scheduler** — status, abort, cleanup, integrate (`all` arg for batch; no start/ready)
- **full** — all tools above; unset/empty `MERGESIDING_MCP_ROLE` → full

## Rules

1. Do **not** edit the human base checkout for parallel work.
2. Start: MCP `mergesiding_start` or `mergesiding start --slug <slug> --repo <abs> [--worktree-root <path>] [--writer-id <id>] [--brief <path>]`
3. Default worktrees: `<repo-parent>/.agent-git-worktrees/<repo-name>/<slug>` (override with `--worktree-root`, `MERGESIDING_WORKTREE_ROOT`, or `.mergesiding.json`)
4. Commit only inside printed worktree paths on branch `task/<slug>`.
5. When done and clean: `mergesiding_ready` / `mergesiding ready --slug <slug>` — that **closes this task**. Do not keep editing that worktree after `ready`.
6. More changes later: wait for integrate (or `abort` to drop the task), then `start` a **new** slug / worktree. One slug → one writer → one shot until ready.
7. Remind the user: nothing is on the integration branch until integrate.
8. Never merge into the integration branch yourself unless asked to run scheduler/full integrate tools (`mergesiding_integrate` / `mergesiding integrate`; batch via `all: true` or CLI `--all`).
9. `abort` does not remove worktrees or branches.
10. Optional disk cleanup: `cleanup` with `--remove-worktree` / `--delete-branch` when status is `done` or after give-up.
11. On escalate under `MERGESIDING_HOME/escalate`: resolve conflicts only, continue rebase, then `ready` again.
12. Do not use `git checkout --ours/--theirs` unless the brief explicitly requires it.
13. If using Cursor: never call `move_agent_to_root` / `move_agent_to_cloned_root` for linked worktrees.

## Docs

- EN: README + [docs/mcp.md](../../docs/mcp.md) + docs/worktree-layout.md
- ZH: README.zh-CN.md + docs/*.zh-CN.md
