# mergesiding scheduler — serial integrate

mergesiding scheduler tools process the ready queue: rebase → verify → merge
under an exclusive lock. Writers own start/ready; you own integrate.

## Tool selection by intent

- **Integrate next ready task** → `mergesiding_integrate` (PRIMARY)
- **Drain queue until empty or stop** → `mergesiding_integrate` with `all=true`
- **Resume BLOCKED_PARTIAL / specific slug** → `mergesiding_integrate` with slug
- **Inspect / recovery** → `mergesiding_status` (omit slug to list all tasks)
- **Abandon** → `mergesiding_abort` (does not remove worktrees; use `mergesiding_cleanup`)
- **Cleanup DONE/ABORTED** → `mergesiding_cleanup`

## Common chains

- After writer ready: `mergesiding_integrate` or `mergesiding_integrate` with `all=true`
- On awaiting_writer / blocked: stop; wait for writer ready; integrate again
- On blocked_partial: `mergesiding_integrate` with that slug (do not rely on queue alone)

## Anti-patterns

- Do not ask writers to merge into integration themselves
- Do not skip escalate files when status is awaiting_writer / blocked
- Do not auto-rollback already-merged repos on blocked_partial
- codegraph pairing is writer-side; if relevant, see docs/codegraph.md (do not init on every start)

## Limitations

- One integrate lock; concurrent integrate fails with lock held
- stop_batch on awaiting_writer / blocked / blocked_partial
