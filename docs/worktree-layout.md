# Where task folders go

中文：[worktree-layout.zh-CN.md](worktree-layout.zh-CN.md)

When you `start` a task, mergesiding creates another checkout of the repo (a git worktree) in its own folder. Your normal project folder stays as-is.

## Default

If the repo is `.../projects/my-app`, the task `foo` lands here:

```text
projects/
  my-app/                 ← keep using this as usual
  .agent-git-worktrees/
    my-app/
      foo/                ← work here for that task
```

So: next to the project, under `.agent-git-worktrees/<repo-name>/<task-name>`.

Because it’s outside the repo, you usually don’t need to touch `.gitignore`.

## Choosing another place

You can pass a folder when starting, set `MERGESIDING_WORKTREE_ROOT`, or put `worktreeRoot` in `.mergesiding.json`.

- Full path (e.g. `D:\wt\my-app`) → use that folder, then put the task name under it.
- Relative path (e.g. `../.agent-git-worktrees/my-app`) → resolved from the **repo root**, not from whatever directory Cursor happened to start in.
- Nothing set → default above.

If you start several repos at once with one shared root (CLI/MCP `--worktree-root` / `worktree_root`, or `MERGESIDING_WORKTREE_ROOT`), each repo gets its own subfolder: `<shared>/<repo-name>/<task-name>`.

Per-repo `worktreeRoot` in `.mergesiding.json` does **not** add an extra repo-name segment (the config already points at that repo’s folder).

## If you put folders inside the repo

Add that path to `.gitignore`, or git will see a mess of nested checkouts.

## What to edit

Only the folder `start` printed, on branch `task/<name>`. Don’t do the parallel task work in your everyday clone.

## codegraph (optional)

Task worktrees do not get a codegraph index automatically. If you use codegraph MCP in the same session, see [codegraph.md](codegraph.md) (default: `projectPath` = human base checkout; do not `codegraph init -i` on every `start`).
