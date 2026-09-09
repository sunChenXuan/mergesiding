# MCP setup

中文：[mcp.zh-CN.md](mcp.zh-CN.md)

mergesiding speaks **stdio MCP**. Any host that can launch a local MCP server can use it — not limited to one IDE.

You register **two** servers: same program, different `MERGESIDING_MCP_ROLE`.

## Writer

Start a task, mark ready, status, abort, cleanup.  
**Cannot** merge.

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
