# codegraph × mergesiding worktrees

English: [codegraph.md](codegraph.md)

mergesiding 与 codegraph 是两套工具，**不互斥**，可在同一会话里同时用。本页只给已经接了 codegraph 的 Agent/宿主看；不用 codegraph 的可以忽略。

新建的任务 worktree **不会**自动带上 `.codegraph/` 索引。`mergesiding start` **不会**跑 `codegraph init`。

## 决策树

| 场景 | 做法 |
|------|------|
| 默认：探索**已有**代码 | codegraph 调用加 `projectPath=<人的 base 仓库绝对路径>`（前提是 base 已做过 `codegraph init`） |
| 每次 `start` 之后 | **不要**对每个新 worktree 执行 `codegraph init -i`（大仓可能很慢） |
| 需要图跟着 **worktree 里未合入的改动**，且用户明确要求 / 改动已经很大 | 在该 worktree 目录执行 `codegraph init -i`，再用 `projectPath=<该 worktree 绝对路径>` |
| 从 base **拷贝** `.codegraph/` 到 worktree | **禁止** |
| 在仓库**上层**建一份 `.codegraph`「包住」多个 worktree | **禁止**（易混树；跨盘布局也包不住） |

## 为什么不要借别的树的索引

codegraph 索引属于**某一**工作树。借用或拷贝另一棵 worktree 的 `.codegraph` 会得到错误或过期符号。若必须跟着当前树的未合入改动，应在该 worktree **本地 init**，不要借索引。

在 mergesiding worktree 里日常查「X 在哪 / Y 怎么走」时，优先用 base 的索引 + `projectPath`，而不是给每个任务目录都 init 一遍。

## Cursor

聊天工作区根保持在人的 base checkout。对 linked worktree **不要**调用 `move_agent_to_root` / `move_agent_to_cloned_root`；用绝对路径改文件，Shell 的 `working_directory` 设为 `start` 打印的 worktree 路径。
