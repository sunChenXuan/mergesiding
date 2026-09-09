# 任务目录放在哪

English: [worktree-layout.md](worktree-layout.md)

`start` 之后，mergesiding 会再检出一份仓库到单独目录（git worktree）。你平时用的那个项目文件夹可以照旧，不要拿它做并行任务。

## 默认位置

仓库在 `.../projects/my-app`，任务名是 `foo` 时：

```text
projects/
  my-app/                 ← 继续当日常工作区
  .agent-git-worktrees/
    my-app/
      foo/                ← 这个任务在这里改
```

也就是：在项目旁边，`.agent-git-worktrees/<仓库名>/<任务名>`。

目录在仓库外面，一般不用改 `.gitignore`。

## 想换地方

可以在启动时指定目录，或设环境变量 `MERGESIDING_WORKTREE_ROOT`，或在 `.mergesiding.json` 里写 `worktreeRoot`。

- 写成完整路径（如 `D:\wt\my-app`）→ 就用那里，下面再跟任务名
- 写成相对路径（如 `../.agent-git-worktrees/my-app`）→ 相对**仓库根目录**计算，不要按 Cursor 当前工作目录猜
- 什么都不写 → 用上面的默认

一次开多个仓库、又共用一个根目录时：`<共用根>/<仓库名>/<任务名>`。

## 如果非要塞进仓库里面

记得把那条路径写进 `.gitignore`，否则 git 会很乱。

## 改哪里

只改 `start` 打印出来的那个目录，分支是 `task/<名字>`。并行任务别动你日常打开的那份克隆。
