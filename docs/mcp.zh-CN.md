# MCP 配置

English: [mcp.md](mcp.md)

mergesiding 走的是 **stdio MCP**。只要宿主能启动本地 MCP 服务，就可以用——**不限某一款 IDE**。

需要配 **两条**：同一个程序，不同的 `MERGESIDING_MCP_ROLE`。

## Writer（写代码）

能：开任务、标记做完、看状态、放弃、清理。  
**不能**：合进主分支。`ready` 之后这次任务对写者就算结束——后续改动要重新 `start`（在 integrate 或 abort 之后）。

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

## 配套 Skill / 规则（示例：Cursor）

**MCP ≠ Skill。** mergesiding 的 MCP 只提供工具（`start` / `ready` / `status` / `abort` / `cleanup`；scheduler 另有 `integrate`），适用于**任意** MCP 宿主。宿主侧的 Skill/规则是另一回事：约定用法（不要改 human base checkout、只在 worktree 提交、`ready` 后由人/scheduler `integrate`）。

示例 — [Cursor](https://cursor.com) 的 Skill（`SKILL.md` / `/mergesiding`）**不会**由 MCP 自动注册。只装 `mcp.json` **不会**出现 `/mergesiding`。对照：有的 Cursor Plugin 在 `plugin.json` 里声明 `"skills": "./skills/"`，装插件就有 skill；mergesiding 当前是二进制 + MCP，需要单独装 Skill/规则（将来可做成 Plugin；目前未提供安装命令）。

writer/scheduler MCP 配好后，若使用 Cursor，建议让 AI（或自己）复制：

```text
skills/mergesiding/SKILL.md  →  ~/.cursor/skills/mergesiding/SKILL.md
```

Windows：`%USERPROFILE%\.cursor\skills\mergesiding\SKILL.md`。  
该宿主备选：把 [cursor-rule-snippet.md](../cursor-rule-snippet.md) 写入用户规则（且不要对 linked worktree 调 `move_agent_to_root`）。

其它 MCP 宿主：把同样约定写进各自的系统提示 / 项目说明即可（可用 [cursor-rule-snippet.md](../cursor-rule-snippet.md) 当文案）。没有 Skill 也能自然语言调 MCP；多 Agent 并行强烈建议装宿主侧说明。

## 可选环境变量

- `MERGESIDING_MCP_ROLE`：`writer` 或 `scheduler`（跑 `mcp` 时必填）
- `MERGESIDING_HOME`：任务状态目录，默认 `~/.mergesiding`
- `MERGESIDING_WORKTREE_ROOT`：任务目录建在哪
- `MERGESIDING_RESUME_CMD`：冲突后可选钩子，可用 `{writer_id}`。通过 `sh -c` / `cmd /C` 执行；`writer_id` 必须是安全字符（`[A-Za-z0-9._:@/+-]`），否则跳过钩子。

旧名 `AGENT_GIT_*` 仍然认。

## 常见情况

- 没设角色 → 起不来，检查环境变量
- writer 看不到合并工具 → 正常
- 任务目录位置不对 → [worktree-layout.zh-CN.md](worktree-layout.zh-CN.md)
- 只装了 MCP 却没有 `/mergesiding`（Cursor 示例）→ 按上文安装配套 Skill
