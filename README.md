# mergesiding

Several AI agents (or people) can work on the same git repo at once without stepping on each other. Each task gets its own folder next to the project. When a task is finished, changes go onto `main` (or your usual branch) **one task at a time**.

Ways to use it:

1. **Command line** — works anywhere you have git  
2. **MCP** — any client that speaks the [Model Context Protocol](https://modelcontextprotocol.io/) (Cursor, Claude Desktop, and other MCP hosts)

中文：[README.zh-CN.md](README.zh-CN.md) · [MCP 配置](docs/mcp.zh-CN.md) · [任务目录放哪](docs/worktree-layout.zh-CN.md)

## Install

You do **not** need to install Go.

1. Open [Releases](https://github.com/sunChenXuan/mergesiding/releases).
2. Download the zip/tarball for your system and unzip it.
3. Either:
   - put `mergesiding` / `mergesiding.exe` on your PATH, or  
   - keep the full path to the file for your MCP config.

Only people changing the source need this:

```bash
git clone https://github.com/sunChenXuan/mergesiding.git
cd mergesiding
go build -o mergesiding ./cmd/mergesiding
```

## MCP (any compatible host)

mergesiding exposes tools over **stdio MCP**. Register **two** servers with the same binary and different roles. Details: [docs/mcp.md](docs/mcp.md).

```json
{
  "mcpServers": {
    "mergesiding-writer": {
      "command": "mergesiding",
      "args": ["mcp"],
      "env": { "MERGESIDING_MCP_ROLE": "writer" }
    },
    "mergesiding-scheduler": {
      "command": "mergesiding",
      "args": ["mcp"],
      "env": { "MERGESIDING_MCP_ROLE": "scheduler" }
    }
  }
}
```

Exact config file location depends on your host (for example Cursor’s MCP settings, or Claude Desktop’s `claude_desktop_config.json`). If the host can’t find `mergesiding`, set `"command"` to the full path (e.g. `C:/Tools/mergesiding.exe`).

- **writer** — start tasks, mark ready, status, abort, cleanup (no merge)  
- **scheduler** — integrate / integrate_all, plus status / abort / cleanup  

MCP is host-agnostic. Example (Cursor): MCP alone does **not** add `/mergesiding` — install the companion Skill (`skills/mergesiding/SKILL.md` → `~/.cursor/skills/mergesiding/`) or paste [cursor-rule-snippet.md](cursor-rule-snippet.md) into User Rules. Details: [docs/mcp.md § Companion Skill](docs/mcp.md#companion-skill--rule-example-cursor).

## Command line

```text
mergesiding start --slug foo --repo C:\path\to\your\repo
```

It prints a folder path. Do all edits and commits **there** (branch `task/foo`). Don’t change your normal project folder for that task.

When you’re done and the folder is clean:

```text
mergesiding ready --slug foo
mergesiding integrate
```

`ready` finishes **this** task: do not keep editing that worktree. Need more changes afterward → wait until integrate (or `abort` if you must drop it), then `start` a **new** slug / worktree.

`integrate` is meant for you (or a small scheduler), not for every coding agent.

## Where the extra folders live

By default, next to the project—not inside it:

```text
projects/
  my-app/                 ← your normal clone
  .agent-git-worktrees/
    my-app/
      foo/                ← the task folder start created
```

More detail: [docs/worktree-layout.md](docs/worktree-layout.md)

## Optional project settings

Put `.mergesiding.json` in the repo root if you want:

```json
{
  "integrationBranch": "main",
  "verify": ["go test ./..."],
  "worktreeRoot": "../.agent-git-worktrees/my-app"
}
```

`verify` commands run before merge. If they fail, merge stops until someone fixes the task.

## License

[MIT](LICENSE) © 2026 sunChenXuan
