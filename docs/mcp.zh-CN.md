# MCP 配置

English: [mcp.md](mcp.md)

mergesiding 走的是 **stdio MCP**。只要宿主能启动本地 MCP 服务，就可以用——**不限某一款 IDE**。

每个 MCP 实例通过 `MERGESIDING_MCP_ROLE`（或旧名 `AGENT_GIT_MCP_ROLE`）决定暴露哪些工具。两者都未设置或为空时，角色为 **`full`**。

## 角色与工具

| 角色 | 工具 |
|------|------|
| **writer** | `start`、`ready`、`status`、`abort`、`cleanup` |
| **scheduler** | `status`、`abort`、`cleanup`、`integrate`（`all` 参数可批量） |
| **full** | 以上全部（含 start 到 integrate） |

`abort` 只改任务状态、移出就绪队列，**不会**删 worktree 或分支。

磁盘清理单独进行：状态为 `done` 或已放弃后，再跑 `cleanup`，并带上 `--remove-worktree` 和/或 `--delete-branch`。

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

能跑 `integrate`（可选 `all: true` 批量处理就绪队列），也能看状态、放弃、清理。  
**不能**：开任务或标记 ready。

```json
"mergesiding-scheduler": {
  "command": "mergesiding",
  "args": ["mcp"],
  "env": { "MERGESIDING_MCP_ROLE": "scheduler" }
}
```

## Full（全部工具）

所有 MCP 工具：start、ready、status、abort、cleanup、integrate（`all` 可批量）。  
`MERGESIDING_MCP_ROLE` 未设置或为空时，也是这一套工具。

```json
"mergesiding": {
  "command": "mergesiding",
  "args": ["mcp"]
}
```

或显式指定：

```json
"mergesiding": {
  "command": "mergesiding",
  "args": ["mcp"],
  "env": { "MERGESIDING_MCP_ROLE": "full" }
}
```

可以配 writer + scheduler（拆分角色），也可以只配一条 full 服务——按你怎么给 Agent 分工具即可。

配置写在你使用的 MCP 宿主的配置文件里即可（例如 Cursor 的 MCP 设置，或 Claude Desktop 的 `claude_desktop_config.json`）。保存后按宿主要求重启或重载。

## 提示找不到 mergesiding 时

把 `command` 写成完整路径。Windows 用正斜杠更省事：

```json
"command": "D:/Tools/mergesiding.exe"
```

## 配套 Skill / 规则（示例：Cursor）

**MCP ≠ Skill。** mergesiding 的 MCP 只提供工具（`start` / `ready` / `status` / `abort` / `cleanup`；scheduler 与 full 另有 `integrate`），适用于**任意** MCP 宿主。宿主侧的 Skill/规则是另一回事：约定用法（不要改 human base checkout、只在 worktree 提交、`ready` 后由人/scheduler `integrate`）。

示例 — [Cursor](https://cursor.com) 的 Skill（`SKILL.md` / `/mergesiding`）**不会**由 MCP 自动注册。只装 `mcp.json` **不会**出现 `/mergesiding`。对照：有的 Cursor Plugin 在 `plugin.json` 里声明 `"skills": "./skills/"`，装插件就有 skill；mergesiding 当前是二进制 + MCP，需要单独装 Skill/规则（将来可做成 Plugin；目前未提供安装命令）。

MCP 配好后，若使用 Cursor，建议让 AI（或自己）复制：

```text
skills/mergesiding/SKILL.md  →  ~/.cursor/skills/mergesiding/SKILL.md
```

Windows：`%USERPROFILE%\.cursor\skills\mergesiding\SKILL.md`。  
该宿主备选：把 [cursor-rule-snippet.md](../cursor-rule-snippet.md) 写入用户规则（且不要对 linked worktree 调 `move_agent_to_root`）。

其它 MCP 宿主：把同样约定写进各自的系统提示 / 项目说明即可（可用 [cursor-rule-snippet.md](../cursor-rule-snippet.md) 当文案）。没有 Skill 也能自然语言调 MCP；多 Agent 并行强烈建议装宿主侧说明。

## 可选环境变量

- `MERGESIDING_MCP_ROLE`：`writer`、`scheduler` 或 `full`（未设置/为空 → `full`）
- `MERGESIDING_HOME`：任务状态目录，默认 `~/.mergesiding`
- `MERGESIDING_WORKTREE_ROOT`：任务目录建在哪
- `MERGESIDING_RESUME_CMD`：冲突后可选钩子，可用 `{writer_id}`。通过 `sh -c` / `cmd /C` 执行；`writer_id` 必须是安全字符（`[A-Za-z0-9._:@/+-]`），否则跳过钩子。

旧名 `AGENT_GIT_*` 仍然认。

## 常见情况

- 无效角色 → 起不来；只能用 `writer`、`scheduler` 或 `full`
- writer 看不到合并工具 → 正常（合并用 scheduler 或 full）
- 任务目录位置不对 → [worktree-layout.zh-CN.md](worktree-layout.zh-CN.md)
- 在 mergesiding worktree 里同时用 codegraph → [codegraph.zh-CN.md](codegraph.zh-CN.md)
- 只装了 MCP 却没有 `/mergesiding`（Cursor 示例）→ 按上文安装配套 Skill
