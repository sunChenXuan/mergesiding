# MCP setup

中文：[mcp.zh-CN.md](mcp.zh-CN.md)

mergesiding speaks **stdio MCP**. Any host that can launch a local MCP server can use it — not limited to one IDE.

Each MCP server instance exposes a subset of tools, controlled by `MERGESIDING_MCP_ROLE` (or legacy `AGENT_GIT_MCP_ROLE`). If both are unset or empty, the role is **`full`**.

## Role matrix

| Role | Tools |
|------|-------|
| **writer** | `start`, `ready`, `status`, `abort`, `cleanup` |
| **scheduler** | `status`, `abort`, `cleanup`, `integrate` (`all` arg for batch) |
| **full** | all of the above (start through integrate) |

`abort` marks a task abandoned and removes it from the ready queue; it does **not** remove worktrees or delete branches.

Disk cleanup is separate: run `cleanup` with `--remove-worktree` and/or `--delete-branch` when status is `done` or after give-up.

## Writer

Start a task, mark ready, status, abort, cleanup.  
**Cannot** merge. After `ready`, that task is finished for the writer — further work needs a new `start` (after integrate or abort).

```json
"mergesiding-writer": {
  "command": "mergesiding",
  "args": ["mcp"],
  "env": { "MERGESIDING_MCP_ROLE": "writer" }
}
```

## Scheduler

Merge with `integrate` (optional `all: true` to drain the ready queue), plus status / abort / cleanup.  
**Cannot** start or ready tasks.

```json
"mergesiding-scheduler": {
  "command": "mergesiding",
  "args": ["mcp"],
  "env": { "MERGESIDING_MCP_ROLE": "scheduler" }
}
```

## Full

Every MCP tool: start, ready, status, abort, cleanup, integrate (with optional `all` for batch).  
Same tool set when `MERGESIDING_MCP_ROLE` is unset or empty.

```json
"mergesiding": {
  "command": "mergesiding",
  "args": ["mcp"]
}
```

Or explicitly:

```json
"mergesiding": {
  "command": "mergesiding",
  "args": ["mcp"],
  "env": { "MERGESIDING_MCP_ROLE": "full" }
}
```

You may register writer + scheduler (split roles) or a single full server — whichever matches how you assign tools to agents.

Put config in whatever file your MCP host uses (examples: Cursor MCP settings, Claude Desktop `claude_desktop_config.json`). Restart or reload the host after saving.

## If the host can’t find `mergesiding`

Set `command` to the full path. On Windows, forward slashes avoid escape issues:

```json
"command": "C:/Tools/mergesiding.exe"
```

macOS/Linux example: `/Users/you/bin/mergesiding`.

## Companion Skill / Rule (example: Cursor)

**MCP ≠ Skill.** mergesiding MCP only exposes tools (`start` / `ready` / `status` / `abort` / `cleanup`; scheduler and full also have `integrate`). It works with **any** MCP host. A host-specific Skill or rule is separate: it teaches workflow (don’t edit the human base checkout, commit only in the worktree, after `ready` let a human/scheduler `integrate`).

Example — [Cursor](https://cursor.com) Skills (`SKILL.md` / `/mergesiding`) are **not** registered by MCP. Installing `mcp.json` alone does **not** add `/mergesiding`. Contrast: some Cursor Plugins declare `"skills": "./skills/"` in `plugin.json` and ship skills with the plugin; mergesiding today is binary + MCP, so install a Skill/Rule separately (a bundled Plugin is a possible future; not shipping yet).

After MCP works, on Cursor ask your AI (or copy yourself):

```text
skills/mergesiding/SKILL.md  →  ~/.cursor/skills/mergesiding/SKILL.md
```

Windows: `%USERPROFILE%\.cursor\skills\mergesiding\SKILL.md`.  
Fallback on that host: paste [cursor-rule-snippet.md](../cursor-rule-snippet.md) into User Rules (and never call `move_agent_to_root` on a linked worktree).

Other MCP hosts: use the same rules via whatever system prompt / project instructions they support ([cursor-rule-snippet.md](../cursor-rule-snippet.md) as text). Natural-language MCP calls work without a Skill; for multi-agent parallel work, install host-side instructions.

## Optional environment variables

- `MERGESIDING_MCP_ROLE` — `writer`, `scheduler`, or `full` (unset/empty → `full`)
- `MERGESIDING_HOME` — task state directory (default `~/.mergesiding`)
- `MERGESIDING_WORKTREE_ROOT` — override where task folders are created
- `MERGESIDING_RESUME_CMD` — optional shell when a conflict needs the writer again; supports `{writer_id}`. Runs via `sh -c` / `cmd /C`. `writer_id` must be shell-safe (`[A-Za-z0-9._:@/+-]`); otherwise the hook is skipped.

Legacy `AGENT_GIT_*` names still work.

## Common issues

- Invalid role → server won’t start; use `writer`, `scheduler`, or `full` only
- Writer has no integrate tools → intentional (use scheduler or full for merge)
- Unexpected task folder location → [worktree-layout.md](worktree-layout.md)
- Expected `/mergesiding` after MCP-only setup (Cursor example) → install the companion Skill above
