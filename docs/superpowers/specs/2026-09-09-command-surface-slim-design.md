# Design: Slim mergesiding command / MCP surface

Date: 2026-09-09  
Status: approved for implementation planning  
Scope: remove redundant CLI/MCP surface; clarify abort vs cleanup; add MCP `full` role (default when unset); align docs

## Goal

Reduce overlapping operations without changing the core product model: parallel isolated worktrees, serial integrate, optional MCP role split.

Happy path stays: `start → ready → integrate`.

## Non-goals

- Collapsing MCP to a single mandatory role (writer/scheduler/`full` all remain valid)
- Changing the task status machine
- Multi-repo, `writer_id`, `brief`, resume hooks, verify, or worktree layout
- Auto-cleanup after successful integrate
- Building a Cursor Plugin / new install UX
- Marketing or “recommended default” framing for any MCP role (default-when-unset is a behavior fact, not a pitch)

## Decisions (locked)

| Topic | Choice |
|-------|--------|
| Compatibility | Breaking OK — delete redundant surface, no soft aliases |
| abort vs cleanup | Split: abort = status only; cleanup = disk only |
| After `done` | Manual `cleanup` still required |
| MCP roles | Keep `writer` / `scheduler`; add `full` (full tool set) |
| Role default | If `MERGESIDING_MCP_ROLE` and legacy env are both unset/empty → `full` |
| Role naming | Canonical name is `full` only — no `all` role alias (integrate tool arg `all` is unrelated) |
| Role docs | Document each role’s permissions only — do not push a preferred role |
| Docs | Update README, MCP docs, Skill, and snippet together |

## Approach

Minimal breaking slim-down (Approach 1 from brainstorm), plus `full` role:

1. Remove `list` / `mergesiding_list` (use `status` with no slug).
2. Remove `mergesiding_integrate_all`; add optional `all` on `mergesiding_integrate` (CLI keeps `integrate --all`).
3. Strip cleanup flags from `abort`; only `cleanup` removes worktrees/branches.
4. Accept `MERGESIDING_MCP_ROLE=full` with the union of writer + scheduler tools; empty/unset → `full`. No role alias `all`.
5. Realign user-facing docs to the slim command narrative; role sections list permissions only.

## Command / MCP surface

### Main path (unchanged)

`start` → `ready` → `integrate`

### Role matrix after change

| Command / tool | Writer | Scheduler | Full |
|----------------|--------|-----------|------|
| `start` / `mergesiding_start` | yes | no | yes |
| `ready` / `mergesiding_ready` | yes | no | yes |
| `status` / `mergesiding_status` | yes | yes | yes |
| `abort` / `mergesiding_abort` | yes | yes | yes |
| `cleanup` / `mergesiding_cleanup` | yes | yes | yes |
| `integrate` / `mergesiding_integrate` | no | yes (`all` arg optional) | yes (`all` arg optional) |

`MERGESIDING_MCP_ROLE` (and legacy `AGENT_GIT_MCP_ROLE`):

- unset / empty → `full`
- `writer` | `scheduler` | `full` only (no `all` alias)

### Removals

- CLI subcommand `list`
- MCP tool `mergesiding_list`
- MCP tool `mergesiding_integrate_all`

### Adjustments

- `status` with no `--slug` / no `slug` argument lists all tasks (former `list` behavior).
- MCP `mergesiding_integrate` gains optional boolean `all` (`true` = former integrate_all). This is a **tool argument**, not the role name (`full`).
- CLI `integrate --all` remains; mutually exclusive with `--slug` (unchanged).

### Role semantics (for docs — permissions only)

- **writer** — start, ready, status, abort, cleanup. No integrate.
- **scheduler** — status, abort, cleanup, integrate (optional batch via tool arg `all`). No start/ready.
- **full** — every MCP tool above (start, ready, status, abort, cleanup, integrate). Same when role env is unset.

Do **not** frame docs as “prefer full” or “prefer split.” State factually that unset role env means `full`. Worktree isolation addresses concurrent file edits; role split only controls which tools an MCP client can call. Serial integrate remains enforced by the integrate lock/queue regardless of role.

CLI is unchanged: one binary exposes all subcommands; MCP roles do not apply to CLI.

MCP server instructions: writer and scheduler keep role-specific instruction embeds; `full` uses a combined instructions file (or writer∪scheduler guidance without implying a preferred setup).

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
- `cursor-rule-snippet.md` — shorten to a pointer to the Skill plus the three-step path and a one-line note that MCP roles differ by tool access (point to mcp docs for the matrix)

Shared narrative:

- Daily: `start → ready → integrate`
- Side paths: `status`, `abort`, `cleanup`
- Explicit: abort does not touch disk; after `done`/`aborted`, cleanup is manual
- MCP: for each role, **only state which tools it has**; mention unset → `full` as a factual default; no “recommended for most users” pitch

Out of doc scope for this change: large status-machine diagrams, Plugin packaging.

## Breaking changes (release note)

1. `mergesiding list` removed → use `mergesiding status`
2. MCP `mergesiding_list` removed → use `mergesiding_status`
3. MCP `mergesiding_integrate_all` removed → `mergesiding_integrate` with `all: true`
4. `abort` no longer accepts cleanup flags (CLI errors if passed; MCP schema drops them) — use `cleanup` after abort
5. Additive / behavior change: unset `MERGESIDING_MCP_ROLE` now defaults to `full` (previously errored); value `full` accepted

## Testing

- Role / `ToolNames` tests: no `list` / `integrate_all`; integrate supports `all` arg
- `RoleFull` exposes start+ready+integrate; writer still excludes integrate; scheduler still lacks start/ready
- `RoleFromEnv`: empty → `full`; `full` → `full`; `all` is invalid; invalid role still errors
- abort: no cleanup side effects; passing old flags fails on CLI
- cleanup: still works for `done` / `aborted` with flags
- CLI help and parsing: no `list`; `integrate --all` still works
- Docs: permissions matrix accurate; unset→full stated as fact; no preferred-role pitch; no recommended `list` / `integrate_all` / “abort removes worktree”

## Success criteria

- Writer / scheduler / full tool lists match the matrix above
- Unset role env starts MCP with full tools
- abort cannot delete git artifacts; cleanup remains the only disk path
- EN/ZH README + MCP docs + Skill + snippet describe the same command flow
- MCP docs state role permissions without promoting a preferred role
- Existing tests updated; suite green
