# MCP 配置

English: [mcp.md](mcp.md)

mergesiding 走的是 **stdio MCP**。只要宿主能启动本地 MCP 服务，就可以用——**不限某一款 IDE**。

需要配 **两条**：同一个程序，不同的 `MERGESIDING_MCP_ROLE`。

## Writer（写代码）

能：开任务、标记做完、看状态、放弃、清理。  
**不能**：合进主分支。

```json
"mergesiding-writer": {
  "command": "mergesiding",
  "args": ["mcp"],
  "env": { "MERGESIDING_MCP_ROLE": "writer" }
}
```

## Scheduler（合并）

能跑 `integrate` / `integrate_all`，也能看状态、放弃、清理。

```json
"mergesiding-scheduler": {
  "command": "mergesiding",
  "args": ["mcp"],
  "env": { "MERGESIDING_MCP_ROLE": "scheduler" }
}
```

配置写在你使用的 MCP 宿主的配置文件里即可（例如 Cursor 的 MCP 设置，或 Claude Desktop 的 `claude_desktop_config.json`）。保存后按宿主要求重启或重载。

## 提示找不到 mergesiding 时

把 `command` 写成完整路径。Windows 用正斜杠更省事：

```json
"command": "D:/Tools/mergesiding.exe"
```

## 可选环境变量

- `MERGESIDING_MCP_ROLE`：`writer` 或 `scheduler`（跑 `mcp` 时必填）
- `MERGESIDING_HOME`：任务状态目录，默认 `~/.mergesiding`
- `MERGESIDING_WORKTREE_ROOT`：任务目录建在哪
- `MERGESIDING_RESUME_CMD`：冲突后可选钩子，可用 `{writer_id}`

旧名 `AGENT_GIT_*` 仍然认。

## 常见情况

- 没设角色 → 起不来，检查环境变量
- writer 看不到合并工具 → 正常
- 任务目录位置不对 → [worktree-layout.zh-CN.md](worktree-layout.zh-CN.md)
