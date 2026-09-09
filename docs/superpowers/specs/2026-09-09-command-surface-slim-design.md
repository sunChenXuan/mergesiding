# Design: Slim mergesiding command / MCP surface

Date: 2026-09-09  
Status: approved for implementation planning  
Scope: remove redundant CLI/MCP surface; clarify abort vs cleanup; align docs

## Goal

Reduce overlapping operations without changing the core product model: parallel isolated worktrees, serial integrate, writer/scheduler role split.

Happy path stays: `start → ready → integrate`.

## Non-goals

- Merging writer and scheduler into one MCP server
- Changing the task status machine
- Multi-repo, `writer_id`, `brief`, resume hooks, verify, or worktree layout
- Auto-cleanup after successful integrate
- Building a Cursor Plugin / new install UX

## Decisions (locked)

| Topic | Choice |
|-------|--------|
| Compatibility | Breaking OK — delete redundant surface, no soft aliases |
| abort vs cleanup | Split: abort = status only; cleanup = disk only |
| After `done` | Manual `cleanup` still required |
| Dual MCP roles | Keep writer / scheduler |
| Docs | Update README, MCP docs, Skill, and snippet together |

## Approach

Minimal breaking slim-down (Approach 1 from brainstorm):

1. Remove `list` / `mergesiding_list` (use `status` with no slug).
2. Remove `mergesiding_integrate_all`; add optional `all` on `mergesiding_integrate` (CLI keeps `integrate --all`).
3. Strip cleanup flags from `abort`; only `cleanup` removes worktrees/branches.
4. Realign all user-facing docs to the same narrative.

## Command / MCP surface

### Main path (unchanged)

`start` → `ready` → `integrate`

### Role matrix after change

| Command / tool | Writer | Scheduler |
|----------------|--------|-----------|
| `start` / `mergesiding_start` | yes | no |
| `ready` / `mergesiding_ready` | yes | no |
| `status` / `mergesiding_status` | yes | yes |
| `abort` / `mergesiding_abort` | yes | yes |
| `cleanup` / `mergesiding_cleanup` | yes | yes |
| `integrate` / `mergesiding_integrate` | no | yes (`all` optional) |

### Removals

- CLI subcommand `list`
- MCP tool `mergesiding_list`
- MCP tool `mergesiding_integrate_all`

### Adjustments

- `status` with no `--slug` / no `slug` argument lists all tasks (former `list` behavior).
- MCP `mergesiding_integrate` gains optional boolean `all` (`true` = former integrate_all).
- CLI `integrate --all` remains; mutually exclusive with `--slug` (unchanged).

### Dual-role rationale (unchanged)

Two MCP configs with the same binary keep writers from seeing `integrate`. Full automation needs both roles; Writer-only + human/CLI `integrate` remains a valid minimal setup. Pure CLI needs neither MCP server.

## abort / cleanup contract

### `abort --slug <s>`

- Allowed statuses: `active`, `ready`, `awaiting_writer`, `blocked`, `blocked_partial` (unchanged).
- Effects: set status `aborted`, remove from ready queue.
- Does **not** accept `--remove-worktree` or `--delete-branch`.
- Does **not** remove worktrees or delete branches.
- CLI: if those flags appear, **error and exit non-zero** (do not silently ignore).
- MCP: remove the properties from the tool schema so hosts stop offering them; no silent disk cleanup path remains on abort.

### `cleanup --slug <s> (--remove-worktree and/or --delete-branch)`

- Allowed statuses: `done`, `aborted` only.
- At least one cleanup flag required (unchanged).
- Successful integrate leaves `done` artifacts until explicit `cleanup`.

### Typical sequences

- Drop work: `abort` → optional `cleanup`
- After merge: `integrate` → optional `cleanup`

## Documentation alignment

Update in one pass so narratives match:

- `README.md`, `README.zh-CN.md`
- `docs/mcp.md`, `docs/mcp.zh-CN.md`
- `skills/mergesiding/SKILL.md`
- `cursor-rule-snippet.md` — shorten to a pointer to the Skill plus the three-step path and dual-role note (avoid a third full rulebook)

Shared narrative:

- Daily: `start → ready → integrate`
- Side paths: `status`, `abort`, `cleanup`
- Explicit: abort does not touch disk; after `done`/`aborted`, cleanup is manual

Out of doc scope for this change: large status-machine diagrams, Plugin packaging.

## Breaking changes (release note)

1. `mergesiding list` removed → use `mergesiding status`
2. MCP `mergesiding_list` removed → use `mergesiding_status`
3. MCP `mergesiding_integrate_all` removed → `mergesiding_integrate` with `all: true`
4. `abort` no longer accepts cleanup flags (CLI errors if passed; MCP schema drops them) — use `cleanup` after abort

## Testing

- Role / `ToolNames` tests: no `list` / `integrate_all`; integrate supports `all`
- abort: no cleanup side effects; passing old flags fails
- cleanup: still works for `done` / `aborted` with flags
- CLI help and parsing: no `list`; `integrate --all` still works
- Docs: post-change search should not recommend `list`, `integrate_all`, or “abort removes worktree”

## Success criteria

- Writer/scheduler tool lists match the matrix above
- abort cannot delete git artifacts; cleanup remains the only disk path
- EN/ZH README + MCP docs + Skill + snippet describe the same flow
- Existing tests updated; suite green
