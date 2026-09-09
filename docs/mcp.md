# MCP setup

中文：[mcp.zh-CN.md](mcp.zh-CN.md)

mergesiding speaks **stdio MCP**. Any host that can launch a local MCP server can use it — not limited to one IDE.

You register **two** servers: same program, different `MERGESIDING_MCP_ROLE`.

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

Merge with `integrate` / `integrate_all`, plus status / abort / cleanup.

```json
"mergesiding-scheduler": {
  "command": "mergesiding",
  "args": ["mcp"],
  "env": { "MERGESIDING_MCP_ROLE": "scheduler" }
}
```

Put this in whatever config file your MCP host uses (examples: Cursor MCP settings, Claude Desktop `claude_desktop_config.json`). Restart or reload the host after saving.

## If the host can’t find `mergesiding`

Set `command` to the full path. On Windows, forward slashes avoid escape issues:

```json
"command": "C:/Tools/mergesiding.exe"
```

macOS/Linux example: `/Users/you/bin/mergesiding`.

## Companion Skill / Rule (example: Cursor)

**MCP ≠ Skill.** mergesiding MCP only exposes tools (`start` / `ready` / `status` / `abort` / `cleanup`; scheduler also has `integrate`). It works with **any** MCP host. A host-specific Skill or rule is separate: it teaches workflow (don’t edit the human base checkout, commit only in the worktree, after `ready` let a human/scheduler `integrate`).

Example — [Cursor](https://cursor.com) Skills (`SKILL.md` / `/mergesiding`) are **not** registered by MCP. Installing `mcp.json` alone does **not** add `/mergesiding`. Contrast: some Cursor Plugins declare `"skills": "./skills/"` in `plugin.json` and ship skills with the plugin; mergesiding today is binary + MCP, so install a Skill/Rule separately (a bundled Plugin is a possible future; not shipping yet).

After writer/scheduler MCP works, on Cursor ask your AI (or copy yourself):

```text
skills/mergesiding/SKILL.md  →  ~/.cursor/skills/mergesiding/SKILL.md
```

Windows: `%USERPROFILE%\.cursor\skills\mergesiding\SKILL.md`.  
Fallback on that host: paste [cursor-rule-snippet.md](../cursor-rule-snippet.md) into User Rules (and never call `move_agent_to_root` on a linked worktree).

Other MCP hosts: use the same rules via whatever system prompt / project instructions they support ([cursor-rule-snippet.md](../cursor-rule-snippet.md) as text). Natural-language MCP calls work without a Skill; for multi-agent parallel work, install host-side instructions.

## Optional environment variables

- `MERGESIDING_MCP_ROLE` — `writer` or `scheduler` (required for `mcp`)
- `MERGESIDING_HOME` — task state directory (default `~/.mergesiding`)
- `MERGESIDING_WORKTREE_ROOT` — override where task folders are created
- `MERGESIDING_RESUME_CMD` — optional shell when a conflict needs the writer again; supports `{writer_id}`. Runs via `sh -c` / `cmd /C`. `writer_id` must be shell-safe (`[A-Za-z0-9._:@/+-]`); otherwise the hook is skipped.

Legacy `AGENT_GIT_*` names still work.

## Common issues

- Missing role → server won’t start; set `MERGESIDING_MCP_ROLE`
- Writer has no integrate tools → intentional
- Unexpected task folder location → [worktree-layout.md](worktree-layout.md)
- Expected `/mergesiding` after MCP-only setup (Cursor example) → install the companion Skill above
