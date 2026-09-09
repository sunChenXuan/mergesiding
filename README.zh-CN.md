# mergesiding

English: [README.md](README.md)

几个 AI（或几个人）要同时改同一个仓库时，容易互相覆盖。mergesiding 给每个任务在项目旁边单独开一个目录；做完之后，再**一个一个**合回你平时用的分支（比如 `main`）。

怎么用：

1. **命令行** — 有 git 就能用  
2. **MCP** — 任何支持 [Model Context Protocol](https://modelcontextprotocol.io/) 的客户端都可以接（Cursor、Claude Desktop，以及其它 MCP 宿主）

更细的说明：

- MCP 怎么配：[docs/mcp.zh-CN.md](docs/mcp.zh-CN.md)
- 任务目录放哪：[docs/worktree-layout.zh-CN.md](docs/worktree-layout.zh-CN.md)
- 与 codegraph 配合（可选）：[docs/codegraph.zh-CN.md](docs/codegraph.zh-CN.md)

## 怎么安装

**不用装 Go，也不用自己编译。**

1. 打开 [Releases](https://github.com/sunChenXuan/mergesiding/releases)
2. 下载你系统对应的压缩包，解压
3. 二选一：
   - 把 `mergesiding`（Windows 是 `mergesiding.exe`）放到 PATH 里；或
   - 记下完整路径，填进你的 MCP 配置

只有改源码的人才需要本地 `go build`。

## MCP（通用，不限某一款软件）

通过 **stdio MCP** 提供工具。用同一个程序配一条或多条，角色不同。细节见 [docs/mcp.zh-CN.md](docs/mcp.zh-CN.md)。

拆分 writer + scheduler：

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

或一条服务暴露全部工具（未设角色 env 时默认为 `full`）：

```json
{
  "mcpServers": {
    "mergesiding": {
      "command": "mergesiding",
      "args": ["mcp"]
    }
  }
}
```

配置文件写在哪，取决于你用的宿主（例如 Cursor 的 MCP 设置，或 Claude Desktop 的 `claude_desktop_config.json`）。找不到命令时，把 `"command"` 改成完整路径，例如 `D:/Tools/mergesiding.exe`。

各角色 MCP 工具：

- **writer**：开任务、标记做完、看状态、放弃、清理（不能合并）  
- **scheduler**：状态、放弃、清理、合并（`integrate` 的 `all` 参数可批量；不能 start/ready）  
- **full**：以上全部；`MERGESIDING_MCP_ROLE` 未设置或为空 → `full`  

`abort` 不会删 worktree 或分支；任务 `done` 或 `aborted` 后需要清理磁盘时再跑 `cleanup`。

MCP 不限某一款宿主。示例（Cursor）：只装 MCP **不会**出现 `/mergesiding`——再装配套 Skill（`skills/mergesiding/SKILL.md` → `~/.cursor/skills/mergesiding/`），或把 [cursor-rule-snippet.md](cursor-rule-snippet.md) 写入用户规则。说明：[docs/mcp.zh-CN.md § 配套 Skill](docs/mcp.zh-CN.md#配套-skill--规则示例cursor)。

## 命令行怎么用

```text
mergesiding start --slug foo --repo C:\path\to\your\repo
```

屏幕上会打印一个目录。这次任务的改动、提交都放在那个目录里（分支名是 `task/foo`）。你平时打开的那个项目目录先别动。

做完并且目录是干净的：

```text
mergesiding ready --slug foo
mergesiding integrate
```

批量合并：`mergesiding integrate --all`。

`ready` 表示**这次任务已经交卷**：不要再改那个 worktree。后面还要改 → 等合入（或必须放弃时用 `abort`），再 `start` **新的** slug / worktree。

`abort` 只改任务状态，不会删 worktree。`aborted` 或 `done` 后要清磁盘，单独跑 `cleanup`：

```text
mergesiding abort --slug foo
mergesiding cleanup --slug foo --remove-worktree --delete-branch
```

`integrate` 建议你自己跑，或交给单独的调度；不要让每个写代码的 Agent 随便合并。

## 任务目录默认在哪

在项目旁边，不在仓库里面：

```text
projects/
  my-app/                 ← 你平时用的克隆
  .agent-git-worktrees/
    my-app/
      foo/                ← start 建出来的任务目录
```

更细的布局说明：[docs/worktree-layout.zh-CN.md](docs/worktree-layout.zh-CN.md)。同会话用 codegraph 时见 [docs/codegraph.zh-CN.md](docs/codegraph.zh-CN.md)。

## 可选配置

仓库根目录可以放 `.mergesiding.json`：

```json
{
  "integrationBranch": "main",
  "verify": ["go test ./..."],
  "worktreeRoot": "../.agent-git-worktrees/my-app"
}
```

合入前会跑 `verify` 里的命令；失败就不会合进去。

## 许可证

[MIT](LICENSE) © 2026 sunChenXuan
