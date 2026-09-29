# Handoff -- the Go module moved under ledger/; the read-engine design is amended and its build is at step 1 of 6

2026-09-29. Newest commit this brief describes: **`4fa75b4`**
(`docs: record the move under ledger/`, main, pushed). Pick-up measures drift
from here. Nothing described below is uncommitted except this entry, its three
learnings entries and the two index rows.

This entry supersedes **Open / next item 1** of
`2026-09-15-twins-for-unfalsified-checks.md` (commit and watch CI): that work
landed as `f1d0729`, `9f1978b`, `6ddc4f6`. Its other items stand.

## Current state

- **built, committed, CI green** -- the Go module (`api/`, `cmd/`, `internal/`,
  `go.mod`, `go.sum`, `buf.yaml`, `buf.gen.yaml`) and its Go gate code (every
  `.go` file that was in `gates/`, with `p2nondet/`) are under `ledger/`
  (`d3e216a`). Root `gates/` keeps `run.sh`, `claimability.py`, `importpin.py`,
  `expect.json`, `conformance/`, `out/`.
  re-verify: `git ls-files '*.go' | grep -v '^ledger/' | wc -l` -> `0`, and `(cd ledger && go list ./... | wc -l)` -> `12`
- **measured** -- the gate is green on the moved tree, on a clean worktree at
  `4fa75b4`, run twice.
  re-verify: `sh gates/run.sh 2>&1 | tail -1` -> `ok conformance pack agrees: claimable=6`
- **measured** -- the move changed no verdict row. The 27 rows after it are the
  same multiset as the 27 rows of a clean run at `1a99587`, outside
  `gate_sha`, `gate_worktree` and `ran_at`.
  re-verify, after a gate run: the `re-verify:` line of
  `docs/learnings/2026-09-29-verdict-rows-differ-in-ran-at.md` -> `27 9119d176a76a349af7dc43e04bb1d234880ea1b43bd2f49c1b330942d01a7cf8`
- **measured** -- the rewritten `proto fresh` step can still fail: one byte
  appended to `ledger/api/meridian/v1/read.pb.go` made the gate stop with
  `FAIL generated ledger/api/meridian/v1/read.pb.go stale`. The file was
  restored from a copy and the next run was green.
  re-verify: `git grep -n 'cmp "ledger/' -- gates/run.sh` -> one line, inside the `for g in $gen` loop
- **measured** -- CI is green on `4fa75b4` (run 36639525532, ubuntu), and the
  one CI line the move added resolved: the run's setup-go step logs
  `cache-dependency-path: ledger/go.sum` and `Cache restored successfully`.
  STATUS.md's 2026-09-29 entry says "CI has not run on the move"; that was true
  when it was written and is a dated record.
  re-verify: `gh run view 36639525532 --json headSha,conclusion -q '.headSha+" "+.conclusion'` -> `4fa75b4f4592dd77a809ab07bee870738856ff9c success`
- **done by the operator** -- the repository is renamed `asof-query-engine`.
  Both earlier names redirect. The Go module path and the verdict surfaces
  still say `meridian`, by decision (below).
  re-verify: `git remote get-url origin` -> `https://github.com/hossainpazooki/asof-query-engine.git`
- **written, committed** -- the design for the build,
  `docs/2026-09-29-asof-query-engine-design.md` (`1a99587`). Its section 0a
  lists sixteen amendments made after the design was read against the tree.
  A1-A11 follow operator rulings. **A12-A16 are unruled.**
  re-verify: `grep -c '(unruled)' docs/2026-09-29-asof-query-engine-design.md` -> `5`
- **not started** -- steps 2 to 6 of the design's build sequence (its
  section 9): the effective-date axis in Go, the Rust engine and P8, P9, the
  module-path flip, the bench. There is no `engine/` directory and no Rust.
  re-verify: `ls engine 2>&1 | head -1` -> no such directory, and `git grep -n 'func ReadAt' -- ledger/` -> no output

## Locked decisions

1. **Engine checkpoints hold pre-apply state; the apply pass runs on every
   read.** Reason: the fold is a batch. It resolves amendments, sorts by
   `(effective, seq)`, then applies, so a later record can change how an
   earlier one applied; on base the amendment at seq 48 rewrites the split at
   seq 22. A checkpoint of applied state between them is wrong for any later
   viewpoint. Operator ruling, 2026-09-29; design A1.
2. **Go gate code lives inside the Go module, at `ledger/gates/`.** Reason: it
   imports the module's `internal/` packages, and Go permits that only from
   inside the tree that holds them. Operator ruling; design A2.
3. **The design document in this repo carries no landing section.** Reason:
   the repository is public. Operator ruling.
4. **Go is the oracle.** The engine's correctness is equality with the ledger;
   the engine is never its own oracle. Reason: one authority, already gated by
   P1-P7. From the design, unchanged.
5. **The module path stays `github.com/hossainpazooki/meridian` until P8's live
   row lands; the flip is its own commit.** Reason: the flip touches every
   import, and the design gates it as a no-change over all rows, the engine's
   included. From the design, unchanged.
6. **Verdict surfaces stay `meridian-lane1-pN` through the rename.** Reason:
   the surface names the instrument, the repository name is an address. From
   the design, unchanged.
7. **A "rows unchanged" comparison excludes exactly three fields** --
   `gate_sha`, `gate_worktree`, `ran_at` -- and compares rows as a multiset,
   never by file name. Reason: those three name the run; everything else names
   what was measured. The design's text names two of them and is wrong on that
   point (Open / next, item 4).
8. **Carried, unchanged:** every decision of
   `2026-09-15-twins-for-unfalsified-checks.md`, including the process lock --
   say "one more commit coming" before pushing, because the operator merges
   within minutes.

## Reuse map

- `gates/run.sh` -- the pattern for a step that belongs to one language's
  tree: `(cd ledger && ...)`, with `BIN` and the verdict directory computed as
  absolute paths under the root before the subshell. The engine's steps take
  the same shape with `engine/`.
- `ledger/gates/manifest.go`, `FixturesDir` -- the one path from the gate
  package to `fixtures/`. Five tests outside the gate package
  (`ledger/internal/{asof,reader,readgrpc,reconcile}` and
  `ledger/cmd/meridian`) reach `fixtures/` by their own relative paths; any
  further change of depth has to touch them too.
- The row-multiset line (`docs/learnings/2026-09-29-verdict-rows-differ-in-ran-at.md`,
  `re-verify:`) -- the gate for steps 2 and 5 of the build ("P1-P7 rows
  unchanged"). It reads only P1-P7, so it stays valid once P8 and P9 rows
  exist.
- `docs/2026-09-29-asof-query-engine-design.md`, section 0a -- each amendment
  with what the tree showed. Section 5 -- every twin with its expected counts.
  Section 9 -- the build sequence and its per-step gate.
- `docs/learnings/2026-09-29-base-feed-effective-axis-shape.md` -- what the
  base feed can and cannot show on the effective-date axis; read it before
  choosing `effective_dates`.
- `fixtures/generate.py`, `naive_fold(records, up_to, mode=...)` -- defects are
  already modelled as modes of the generator's own fold; the counts the design
  calls "measured by the generator" are new modes there.

## Invariants

- **The gates run only through `sh gates/run.sh`, from the repository root.**
  It clears `gates/out/`; a bare `go test` appends, and duplicates read as
  refusals from the pack and as `duplicate verdict rows` from
  `claimability.py`.
- **`go` commands run from `ledger/`.** There is no `go.mod` at the root.
- **No step of `gates/run.sh` is removed, reordered or made conditional.** The
  move kept every step header in order; so must every later step.
- **A gate never writes into `fixtures/`.**
- **The generator names neither the ledger nor the engine.**
  `gates/importpin.py` enforces the first today; the design (section 7) adds
  the second when `engine/` exists.
- **Nothing is CLAIMABLE anywhere unless STATUS.md says so**, and its table is
  generated from the rows. The README carries no counts.
- **Handoff and learnings entries are immutable.** A wrong one is superseded
  by a new dated entry. Older dated entries in STATUS.md keep the paths they
  were written with.
- **Exit 2 from the pack is unevaluable and is never coerced.**
- **This repository is public; its governing text is private.** Tracked files
  cite only what a reader can check here, and carry no home-directory or
  scratch path.
- **Only the operator writes git history.**

## Open / next

1. **Rulings on A12-A16, before step 2 is built.** A16 (an amendment visible
   at a date that hides its action is refused `unknown_action`) and A14 (the
   read universe includes the read with no date, twelve pairs on base) decide
   what step 2's generator and tests assert. A12, A13 and A15 can wait for
   step 3 and step 4.
2. **Step 2, the effective-date axis in Go:** `ReadAt`, `--effective`, the
   `effective` snapshot key, the generator's `effective_dates`, its filter and
   its visible-actions guard. Gate: the row-multiset line prints the same
   hash, and the snapshot pin test still holds with no date supplied. The fold
   itself is not edited; `ReadAt` filters the records it hands to it.
3. **STATUS.md owes one dated note:** CI ran green on the move
   (run 36639525532) and the cache path resolved. It belongs to the next
   change that touches STATUS.md.
4. **The design owes one amendment:** its move gate and its module-flip gate
   name two excluded fields; the measured comparison needs three
   (decision 7). If A15 is struck, its row count for the flip gate is 35,
   not 36.
5. **Operator's call, not taken here:** three path pointers inside notes dated
   2026-09-15 in STATUS.md's Honest limits were re-pointed to `ledger/` by the
   move's docs pass, and the 2026-09-29 entry says so. Revert them if dated
   notes should keep the paths they were written with.
6. **Operator's call, not taken here:** tracked entries that predate this
   work contain home-directory paths, in a public repository
   (`git grep -l -i 'C:.Users' -- docs` lists them). They are immutable under
   the learnings rule, so the route is a superseding entry or a history
   decision, not an edit.
7. **Unchanged from the previous handoff:** the optional third P7 twin for the
   two folded legs; lanes 2-3; the chain-covered-residue decision; generator
   RNG via `getrandbits`. The learnings form gate is red on one pre-existing
   entry, `2026-09-01-vacuity-guard-denominator.md`, and on nothing else.

**How this was built, for whoever weighs the evidence.** The design review and
its amendments were one session's reading of the tree. The move was made by
one builder agent, gated twice by an integration agent, and attacked by two
skeptics, both of which returned not refuted; every measured line above was
then re-run by the orchestrating session itself, on the committed tree, and
that re-run is what this brief reports.
