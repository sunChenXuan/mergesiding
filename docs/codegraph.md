# codegraph × mergesiding worktrees

中文：[codegraph.zh-CN.md](codegraph.zh-CN.md)

mergesiding and codegraph are separate tools. They are **not** mutually exclusive — you can use both in the same session. This page is only for agents/hosts that already have codegraph; others can ignore it.

New task worktrees do **not** inherit a `.codegraph/` index. `mergesiding start` does **not** run `codegraph init`.

## Decision tree

| Situation | Action |
|-----------|--------|
| Default: explore **existing** code | Call codegraph with `projectPath=<absolute path to the human base checkout>` (base must already have been `codegraph init`’d) |
| Right after every `start` | Do **not** run `codegraph init -i` on the new worktree (large repos can be very slow) |
| Need the graph to follow **unmerged worktree edits**, and the user asks or the diff is large | Run `codegraph init -i` **in that worktree**, then use `projectPath=<that worktree absolute path>` |
| Copy `.codegraph/` from base → worktree | **Forbidden** |
| One `.codegraph` above the repo to “cover” many worktrees | **Forbidden** (easy to mix trees; cross-drive layouts can’t wrap them anyway) |

## Why not borrow another tree’s index

codegraph indexes belong to **one** working tree. Pointing at or copying another worktree’s `.codegraph` yields wrong or stale symbols. If you must follow the current tree’s unmerged changes, **init locally** in that worktree — do not borrow.

For everyday “where is X / how does Y work” while coding in a mergesiding worktree, prefer the **base** index via `projectPath` instead of init’ing every task folder.

## Cursor

Keep the chat workspace on the human base checkout. For linked worktrees, do **not** call `move_agent_to_root` / `move_agent_to_cloned_root`. Edit via absolute paths and Shell `working_directory` set to the worktree path printed by `start`.
