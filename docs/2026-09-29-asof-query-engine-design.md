# ASOF-QUERY-ENGINE — Rust read engine over the MERIDIAN ledger: Design

*2026-09-29 — output of the brainstorm over the 09-28 discussion. This is the
validated design for the next local build.
Nothing here is claimable until its gate is green and every twin is red. The
build starts from `6ddc4f6` (`docs: twins handoff, P3 stays partial`, main);
pick-up measures drift from there. Rule carried from every prior seed: if a
property can't get a twin, it isn't a property — it's a hope. Build the twin
first.*

*Amended 2026-09-29, the same day, after the design was read against the tree
at `6ddc4f6`. §0a lists every change and what in the tree forced it. A1–A11
follow the operator's rulings on that review. A12–A16 were found while
amending and are **unruled**: they are written in so the document is
coherent, and each can be struck on its own.*

---

## 0. Decisions locked in the brainstorm

| Question | Decision |
|---|---|
| Scope | **P8 + P9 + effective-date axis.** A Rust read engine that is byte-identical to the Go fold (P8), stays byte-identical under concurrent reads (P9), and answers bitemporal `AsOf(seq, effective)` reads. |
| Rename | **GitHub repo renamed to `asof-query-engine` now**, on this build's first commit. The Go module path `github.com/hossainpazooki/meridian` **stays until P8 is green**; the flip is its own commit, after P8's row lands. The old URL redirects; immutable handoff/learning entries keep it. |
| Layout | **Go moves to `ledger/`, its gate code with it (`ledger/gates/`); Rust lives in `engine/`.** `fixtures/`, `docs/`, `STATUS.md`, `README.md`, CI and the language-neutral half of `gates/` (runner, claimability, import pin, expectations, conformance pack, `out/`) stay at root and are shared. The move is a measured no-change: every existing live row keeps its `content_hash`. *(amended, A2)* |
| P9 claim | **Fidelity gated at fixture scale; numbers reported, unclaimed.** Throughput and latency come from a pinned `--scale` feed, land in `gates/out/bench.json` and a "Measured, unclaimed" section of STATUS.md, and never enter a verdict row. |
| Oracle | **Go stays the single authority.** The Go `asof` gains `--effective` so it can produce every read the engine produces; the engine is never its own oracle. The generator's naive fold stays import-pinned, names neither the ledger nor the engine, and gains exactly what measuring the new counts requires: the effective filter and one mode per modelled defect. *(amended, A9)* |
| Read architecture (engine) | **Pre-apply checkpoints + tail extension; the apply pass runs on every read.** A checkpoint every `K` events holds what the fold knows *before* it resolves amendments or applies anything. `AsOf(V, E)` restores the nearest checkpoint `≤ V` whose whole prefix is visible at `E`, extends it with the visible tail, then resolves, sorts and applies. Pure recompute stays available (`--no-index`) as the in-process control the P8 twins compare against. *(amended, A1, A12)* |
| Snapshot format | **Unchanged when `E` is omitted.** A read with `E` supplied adds exactly one canonical key, `effective`, to the snapshot. P1–P7 never supply it, so their bytes and hashes do not move. |
| Surfaces | Verdict surfaces stay `meridian-lane1-p8`, `meridian-lane1-p9`. The rename does not rename rows; the surface is the instrument, the repo is the address. |
| Bench schedule | The engine's own `bench` subcommand runs the schedule; the Go CLI is timed only as the recompute control. No Go changes for benchmarking. |

## 0a. Amendments, 2026-09-29

Evidence is from the tree at `6ddc4f6` unless a row says otherwise.

| # | The brainstorm said | The tree shows | The design now says |
|---|---|---|---|
| A1 | Checkpoint the fold state every `K` events, replay the tail, then apply the effective filter. | The fold is a batch: it resolves every amendment, sorts every event by `(effective, seq)`, then applies. On base, the amendment at seq 48 rewrites the split at seq 22 (ratio 2 → 3), which has already been applied in any checkpoint between them. A filter cannot be applied after the fold: applied state has no record of which events made it. | Checkpoints hold pre-apply state; the filter runs before resolution; the apply pass runs per read (§4). Operator ruling. |
| A2 | `gates/` stays at root; Go moves to `ledger/`. | The gate package (`gates/*.go`) and `gates/p2nondet/main.go` import `internal/*`. Go refuses an `internal` import from outside the tree rooted at its parent, and with `go.mod` in `ledger/` the root `gates/` is outside the module. | Go gate code moves to `ledger/gates/`; scripts, pack, expectations and `out/` stay in root `gates/` (§2). Operator ruling. |
| A3 | P9 twin `hash_from_stale_read` expects 120. | Each thread repeated one `(V, E)` pair for all its reads, so the previous response's hash was the correct hash: the twin would have been green. | Threads rotate through the pairs, one step per read (§5, P9). |
| A4 | P9 twin `shared_scratch_state` expects exactly 112. | Shared fold state is not distinct per pair: every event with seq ≤ 21 is effective on or before 2026-01-17, so at V1 two different `E` fold the same events, and the count depended on which thread wrote last. | The twin shares whole responses, which differ for any two distinct pairs by construction (§5, P9). |
| A5 | `effective_dates` has one date "between the action and its amendment's effective dates". | The amendment (seq 48) and its action (seq 22) carry the same effective date, 2026-01-17. Only the V axis separates them. | The three dates are placed around the two *actions*, and the guard is stated over which actions are visible (§5, P8). |
| A6 | `checkpoint_tail_dropped`'s count is derivable from `viewpoints` alone. | Whether dropping the tail changes a byte depends on whether the tail is visible at `E` and on what it does. | Measured by the generator, guarded `> 0` (§5, P8). |
| A7 | `chain_refused` uses `fixtures/p2/mutated-unrechained`. | No such path. The unrechained feed is `fixtures/p2/tampered`; the ledger CLI refuses it at seq 23. | `fixtures/p2/tampered` (§5, P8). |
| A8 | Engine fold state is `i128`. | The ledger holds `int64` and refuses an event on overflow at eleven points of the apply pass; each refusal is in the snapshot. Only the division's product is wide (`big.Int`). | State is `i64` with checked arithmetic and the ledger's refusals; the product alone is `i128` (§4). |
| A9 | The generator's naive fold "gains nothing". | §5 has the generator measure twin counts that depend on `E`; `naive_fold` already takes a `mode` for modelled defects. | It gains the filter and the modes, and stays import-pinned (§0, §7). |
| A10 | "The README claim-language grep extends to…", including a bare `x`. | No such grep is a gate step. The only one is a command quoted in a plan document. A bare `x` matches most prose. | A new gate step with a negative control; the multiplier is a pattern, not a letter (§6). |
| A11 | Seq ranges `[0..V]`; events apply "in feed order"; twin binaries in `gates/bin/`; CI measured on `ubuntu-24.04`. | Seq starts at 1. The fold applies in `(effective, seq)` order. Build output is root `bin/` (gitignored). The workflow names `ubuntu-latest`. | Corrected in place. |
| A12 | *(unruled)* — | A pre-apply checkpoint built without a filter has already deduplicated fills over records a filter may remove. | A checkpoint is usable for `(V, E)` only when its whole prefix is visible at `E` (§4). |
| A13 | *(unruled)* The harness sets `K = 2` for one twin; default `K` is 1024. | Base has 71 events. At `K = 1024` the live cell has only the empty checkpoint, and `index_equals_recompute` compares recompute with itself. | The live cell and every P8 twin run at the manifest's `p8.checkpoint_every`, 2 on base (§5, P8). |
| A14 | *(unruled)* P8 evaluates 3 × 3 = 9 pairs, all with `E`. | The read that exists today, `AsOf(V)` without `E`, was compared by no check. | The universe is `viewpoints × (effective_dates + the unfiltered read)`, 3 × 4 = 12 (§5). |
| A15 | *(unruled)* P9 has two twins. | Crediting is per check (`gates/claimability.py`): a live check no twin drives nonzero is unfalsified and blocks CLAIMABLE, as it does for P3. `no_read_creates_or_writes` had no twin. | A third P9 twin; 36 rows, not 34 (§5, §7). |
| A16 | *(unruled)* — | What a read does with an amendment whose action `E` hides was not stated. | Stated as the consequence of filtering records: refused `unknown_action` (§1). Unexercised on base, and named as a limit (§5). |

One correction to the review itself: it counted "2 effective-date inversions"
on base as evidence for A1. Both are at seq 48–49, and they are the amendment
and an absorbed duplicate fill. A1 rests on the amendment.

## 1. What "as-of" means from this build on

Two timelines, one clock — unchanged. Feed sequence is knowledge time and the
only clock; every event carries its domain effective date. This build exposes
the second axis on reads:

- `AsOf(V)` — fold events `[1..V]`, apply each at its effective date. This is
  today's read, byte for byte.
- `AsOf(V, E)` — the same fold over the records of `[1..V]` **whose effective
  date is `≤ E`**. "What did the ledger say, at knowledge point V, the
  portfolio was on date E." `E` removes records before the fold sees them and
  changes nothing about how the remaining ones apply, which is
  `(effective, seq)` order, as today. Events known at V but effective after E
  are invisible at `(V, E)` by construction, the same mechanism that makes a
  price learned after V invisible at V.

An amendment (`action_amendment`) is a later event. The fold resolves
amendments before it applies anything: every visible action first, then every
visible amendment in seq order, overwriting terms. `(V, E)` with V before the
amendment's seq sees the original terms; with V at or after it, the corrected
terms, applied at the action's effective date. No special machinery, same as
today.

The amendment's own effective date decides one thing only: whether the
amendment record is visible at `E`. An amendment visible at `(V, E)` whose
action is not visible there is refused `unknown_action`, which is what the
fold does today with an amendment of an action it has not seen, and the
refusal is in the snapshot. That is a consequence of filtering records, not a
rule added for it.

The snapshot for `(V, E)` is the canonical snapshot as today plus one key,
`"effective": "YYYY-MM-DD"`, sorted into place. `feed_prefix_hash` still names
the visible prefix `[1..V]`, read from the unfiltered feed; `E` is a filter
over it, not a different prefix.

## 2. Repo shape after the move

```
asof-query-engine/
  ledger/                 Go: the authority (module path unchanged until P8)
    cmd/meridian/         CLI: append, replay, snapshot, asof [--effective], reconcile, serve
    internal/{feed,fold,snapshot,asof,canon,reconcile,reader,readgrpc}
    api/meridian/v1/      read.proto, generated *.pb.go
    gates/                p1..p9 tests, manifest.go, verdict.go, p2nondet/
    buf.yaml buf.gen.yaml go.mod go.sum
  engine/                 Rust: the read engine (crate `asof-engine`)
    Cargo.toml Cargo.lock rust-toolchain.toml
    src/{feed,fold,snapshot,index,asof,cli}.rs  (one module, one purpose)
    benches/              criterion or hand-rolled; see §6
  fixtures/               generate.py (+ --scale, + effective dates, + P8/P9 counts), checked-in feeds
    base/  p1/ .. p6/  bench/    bench/feed.jsonl is generated, gitignored, pinned by fixtures/bench/feed.sha256
  gates/                  run.sh, claimability.py, importpin.py, claimwords.py, expect.json, conformance/, out/
  bin/                    build output, gitignored: ledger CLI, engine, one engine binary per twin
  docs/  STATUS.md  README.md  .github/
```

The gate code that is Go lives inside the Go module because it imports the
module's `internal` packages, and Go allows that only from inside the tree
that holds them. What stays in root `gates/` is what either language's rows
pass through.

Move mechanics, first commit of the build:

1. `git mv` every Go path into `ledger/`: `api/`, `cmd/`, `internal/`,
   `go.mod`, `go.sum`, `buf.yaml`, `buf.gen.yaml`, and from `gates/` the `.go`
   files and `p2nondet/`. `gates/run.sh` runs `go vet`, `go build`, `go test`
   and the proto-fresh check from `ledger/`; the gate package's fixtures path
   gains one level; CI paths follow.
2. Re-run `sh gates/run.sh`. Gate for the move: **all 27 rows byte-identical
   to `6ddc4f6`'s rows** except `gate_sha`/`gate_worktree`, compared row by
   row. A move that changes a `content_hash` is a bug, not a migration.
3. `gh repo rename asof-query-engine` — **operator action, not the session's.**
   The session updates `README.md`, `STATUS.md`, `.github/` and `docs/*-design.md`
   links to the new URL and sets the local remote. Handoff and learning entries
   are immutable and are not touched; GitHub's redirect covers them.

The module-path flip (`go.mod`, every import, the proto's `go_package`) is a
separate commit after P8's live row lands, gated the same way: 36 rows, all
`content_hash` values unchanged.

## 3. Ledger changes (Go) — the minimum that keeps Go the oracle

- `internal/asof`: `Read(feed, seq)` gains `ReadAt(feed, seq, effective)`.
  `Read` calls `ReadAt` with no effective date; no behaviour change.
- `internal/fold`: **untouched.** `ReadAt` removes the records effective
  after `E` and hands the rest to `fold.Fold`, whose contract is already
  stated over `Seq` and over any slice a caller hands it, contiguous or not.
  The fold does not know about `E`. The prefix hash is read from the
  unfiltered feed.
- `internal/snapshot`: canonical encoder emits `effective` when set. When
  unset, bytes are unchanged; a test asserts the base fixture's V1/V2/V3
  snapshots still hash to the committed pin.
- `cmd/meridian asof`: gains `--effective YYYY-MM-DD`. `--seq` unchanged.
- `internal/reader` / gRPC: **untouched.** `AsOfRequest` does not gain an
  effective field in this build; that is a P7 change and would reopen P7's
  rows. Named non-goal, §8.

## 4. Engine (Rust) — components and boundaries

Each module answers: what it does, how it is used, what it depends on.

- **`feed`** — reads a hash-chained JSONL feed, verifies every link, refuses
  a broken chain with a `ChainError` carrying the seq. **Read-only, never
  creates**: a missing path is `NotFound`, mirroring `reader.ErrNotFound` —
  the "feed.Open creates an empty feed and answers green over nothing" hole
  is closed on this side by construction. Depends on nothing internal.
- **`fold`** — port of the Go fold, pass for pass: decode and deduplicate,
  resolve amendments, sort by `(effective, seq)`, apply, value. Integer minor
  units held as `i64` with checked arithmetic; where the ledger refuses an
  event because a product or a running total does not fit, the engine refuses
  it with the same refusal record. State is `(total_cost, qty)` per instrument
  plus cash, realized P&L, dividend income, refusal and absorbed records. The
  single division, `relieved = round_half_even(total_cost × qty_sold / total_qty)`,
  is one function in one file with a doc comment naming its Go twin; its
  product is the one `i128` in the crate. No floats. Idempotency key derived
  from the same identity fields, same hash.
- **`snapshot`** — canonical JSON: sorted keys, fixed integer encodings,
  optional `effective`. Byte-equal to Go's encoder on every fixture. sha256
  over the bytes, `sha256:`-prefixed like the CLI.
- **`index`** — pre-apply checkpoints every `K` events, held as
  `Arc<Checkpoint>`, immutable. The checkpoint at seq `k` holds what the fold
  knows about `[1..k]` before it resolves amendments or applies anything: the
  first valid action per `action_id`, the amendments in seq order, the fill
  dedupe map, the decoded fills and prices, the absorbed duplicates, the
  decode-time refusals, and `max_effective`, the latest effective date in
  `[1..k]`. Each is a function of the records in seq order and of nothing
  later, so a checkpoint extended with `(k..V]` equals the pre-apply state of
  `[1..V]` decoded from zero. Built once per feed open; the feed is opened,
  chain-verified and decoded once per process, not per request. Depends on
  `feed`, `fold`.
- **`asof`** — `read(V, E)`: restore the nearest checkpoint `k ≤ V` whose
  whole prefix is visible at `E` (`max_effective ≤ E`, or `E` omitted; the
  empty checkpoint at 0 always qualifies), extend it with the records of
  `(k..V]` visible at `E`, then resolve amendments, sort, apply, value and
  serialize. `read_recompute(V, E)` does the same from the empty checkpoint;
  same signature, used by the P8 harness as the in-process control. `&self`
  and `Sync`; no interior mutability outside the index's immutable
  checkpoints.
- **`cli`** — `asof-engine head|asof --seq N [--effective D] [--no-index]|bench`.
  Output formats match the Go CLI's where the command exists there (`head`,
  `asof`): same `prefix_hash`, `snapshot_hash`, same bytes on stdout.

**What the index buys, stated before anyone measures it.** The index saves
opening, chain verification, decoding, canonicalization and hashing. It does
not save the apply pass, which costs the visible events on every read. An
index that checkpoints applied state is the faster design and the one this
build does not take: on this feed it is wrong whenever a later record changes
how an earlier one applies (§10).

Non-goals inside the crate: no gRPC server, no `unsafe`, no SIMD, no async
runtime, no writes.

## 5. Properties and twins

Each property is claimable only when its live cell passes **and** every twin
fails for exactly the planted reason, with exact counts owned by
`fixtures/base/manifest.json`, **and** every check the live cell reports is
driven nonzero by at least one of those twins. `Emit` refuses drift between
the twin's expectation in code and the manifest, as for P1–P7.

**The read universe.** `effective_dates` is a new manifest list of three
dates, placed by rule: the day before the first action's effective date, the
day before the second action's, and the feed's last effective date. On base
those are 2026-01-16, 2026-01-22 and 2026-02-13, around the split `CA-0001`
(seq 22, effective 2026-01-17) and the dividend `CA-0002` (seq 33, effective
2026-01-23). The generator asserts that at V3 the three dates make visible no
action, the split alone, and both: the P3 three-histories argument on the E
axis, stated over actions because fills alone would make any three dates
differ. A **pair** is a viewpoint with one of the three dates or with no date:
3 × 4 = 12 on base. Any two distinct pairs differ in bytes, by `feed_seq` or
by `effective`.

### P8 — Cross-language byte identity

**Property.** For every read the ledger can answer, the engine answers with
the same bytes.

| Check | Evaluated | Violation counts |
|---|---|---|
| `snapshot_bytes_equal_go` | pairs (12 on base) | pairs where engine bytes ≠ Go bytes |
| `snapshot_hash_equal_go` | same | pairs where the engine's claimed hash ≠ Go's |
| `index_equals_recompute` | same | pairs where `read(V,E)` ≠ `read_recompute(V,E)` in-process |
| `chain_refused` | 1 | 0 if the engine refuses `fixtures/p2/tampered`; 1 if it opens it |

The live cell and every twin run at `K` = the manifest's
`p8.checkpoint_every`, 2 on base. At the default the base feed would hold one
checkpoint, the empty one, and `index_equals_recompute` would compare a
function with itself.

Twins (manifest keys `p8.twin_*`; each is a feature-flagged build of the
engine, `--features twin-<name>`, so the live binary contains no twin code).
Every count below that is not a constant is **measured by `generate.py`**,
which models the defect as a mode of its own fold and guards the count `> 0`,
or the twin proves nothing. The engine's claimed hash is over the bytes it
returns, so wherever a twin moves `snapshot_bytes_equal_go` it moves
`snapshot_hash_equal_go` by the same count.

- **`rounding_half_up`** — the single division rounds half-up. Expected:
  `snapshot_bytes_equal_go` = number of pairs whose visible history contains a
  sell whose relief hits a `.5` boundary; the generator plants at least one
  such sell. `index_equals_recompute` 0: both paths round the same wrong way.
- **`checkpoint_tail_dropped`** — `read` restores the checkpoint and skips the
  extension; the response still names the requested `V`. Expected:
  `index_equals_recompute` = pairs where the visible tail changes the
  snapshot; `snapshot_bytes_equal_go` moves by the same count.
- **`effective_axis_ignored`** — `E` is parsed and written into the snapshot,
  and no record is removed. Expected: `snapshot_bytes_equal_go` = pairs where
  removing the records `E` hides changes the snapshot. Pairs with no date are
  never among them. `index_equals_recompute` 0.
- **`chain_unverified`** — `feed` skips link verification. Expected:
  `chain_refused` = 1, everything else 0 (the tampered feed is never used for
  the byte checks, so no other check moves).

Live cell: base fixture, every check 0. Scope `"engine vs ledger CLI,
fixtures/base, 3 viewpoints × (3 effective dates + unfiltered), K=2"`;
`content_hash` = Go's V3 snapshot hash; basis `"sha256 of canonical snapshot
bytes as produced by the ledger CLI"`.

**Honest limits, on-page.** No fixture makes the ledger refuse on overflow, so
the engine's overflow refusals are equal to the ledger's by unit test and by
no row. No read on base shows an amendment whose action `E` hides (§1): the
base amendment and its action share an effective date. At `K` = 2 every tail
is one record long.

### P9 — Concurrent-read fidelity

**Property.** Under concurrent reads, every response is the bytes a serial
read would have produced, and the harness can tell when it is not.

**Schedule (deterministic).** `T` threads (8 on base), `R` rounds (16). In
round `r`, thread `t` reads pair `(t + r) mod 12`, so the pairs read in one
round are all different and a thread never reads the same pair twice running.
A barrier releases each round's reads at once. Every response is compared to
the serial `read(V, E)` bytes computed before the first barrier. Before
releasing anything the harness checks that the 12 serial hashes are pairwise
distinct, and reports the schedule unevaluable if they are not: both byte
twins below are exact only because of it. Universe = `T × R` = 128.

| Check | Evaluated | Violation counts |
|---|---|---|
| `concurrent_bytes_equal_serial` | 128 | responses ≠ serial bytes |
| `concurrent_hash_equal_serial` | 128 | claimed hash ≠ serial hash |
| `no_read_creates_or_writes` | 1 | 1 if the feed's bytes or mtime changed, or any file appeared under the feed's directory, during the schedule |

Twins:

- **`shared_scratch_state`** — the engine serializes every response into one
  shared buffer, lock-guarded so the twin has no data race and needs no
  `unsafe`. A second barrier sits between "write" and "return", so every
  thread returns what the last writer left. Expected:
  `concurrent_bytes_equal_serial` = `(T − 1) × R` = 112, exactly:
  in each round the last writer is right and the other seven are wrong,
  whichever thread wrote last, because the eight pairs of a round have eight
  different responses. `concurrent_hash_equal_serial` = 112, the hash being
  over the returned bytes. A twin whose count depends on the scheduler is not
  a twin; this one's does not.
- **`hash_from_stale_read`** — the engine returns fresh bytes with the hash of
  the previous response on the same thread; a thread's first read has no
  previous response and carries its own hash. Expected:
  `concurrent_hash_equal_serial` = `T × (R − 1)` = 120; bytes check 0. Shows
  the hash leg is load-bearing under concurrency, P7-twin-2's argument moved
  in-process.
- **`sidecar_written_on_open`** — the engine writes an index file beside the
  feed when it opens it. Expected: `no_read_creates_or_writes` = 1; both byte
  checks 0. Without it the third check has never been seen to fail and P9
  would derive PARTIAL.

Live cell: every check 0. Scope names `T`, `R`, `K`, and the fixture.

**Honest limit, on-page:** P9 measures threads in one process on one machine.
Nothing is claimed about a network, a server, backpressure, or a scheduler
other than the one CI ran.

## 6. Numbers — measured, unclaimed

- `fixtures/generate.py --scale N` emits a seeded feed of `N` events
  (`fixtures/bench/feed.jsonl`, gitignored) and writes its sha256 to
  `fixtures/bench/feed.sha256` (committed). Default `N` = 1,000,000.
- `asof-engine bench --feed fixtures/bench/feed.jsonl --threads T --reads R`
  refuses to run unless the feed hashes to the pin (`exit 2`, unevaluable —
  the same posture as an empty conformance pin). It reports p50/p99 latency
  and reads/s for `read` and `read_recompute`, and times
  `ledger/cmd/meridian asof` as the recompute control.
- Output: `gates/out/bench.json` (not a verdict row; the conformance checker
  ignores it by name) and a `## Measured, unclaimed` table in STATUS.md,
  regenerated by `claimability.py` from `bench.json`, carrying machine, OS,
  toolchain, `K`, and the feed pin.
- **A new gate step, `gates/claimwords.py`.** No claim-language check is a
  gate step today. This one refuses performance vocabulary in `README.md` and
  in `STATUS.md` outside that generated section: `faster`, `slower`,
  `throughput`, `latency`, `p50`, `p99`, `reads/s`, and a number written
  against `x` or `×` with no space (`3x`, `10×`). It carries a `--self-test`
  that plants each and requires the refusal. Both files pass it as they stand.
  Numbers appear in exactly one place and are labelled unclaimed there.
- `bench` runs in `run.sh` behind `BENCH=1`; CI does not set it. CI numbers
  from a shared runner would be noise labelled as measurement.

## 7. Gates, CI, state of record

- `gates/run.sh` gains `== engine build` (`cargo build --locked --release`
  in `engine/`, plus each twin feature build into `bin/`), `== claim words`,
  `== p8`, `== p9`, and `== bench` (when `BENCH=1`). Its Go steps run from
  `ledger/`. `proto fresh` is unchanged but for that.
- `ledger/gates/p8_test.go` and `p9_test.go` drive the engine binaries and
  emit their rows through the existing `Emit`.
- `gates/importpin.py`: the generator may name the engine no more than the
  ledger. Its forbidden literals gain `engine/`, `.rs` and `cargo`, with the
  negative controls to match.
- `gates/claimability.py`: `PROPS = [1..9]`, success line
  `ok lane1 claimable=N/9`; the overclaim check on STATUS.md counts nine.
- Row count: 27 + P8 (1 live + 4 twins) + P9 (1 live + 3 twins) = **36**.
  The learning `2026-09-15-p3-set-checks-cannot-move-on-this-feed.md` carries
  the "27 rows" figure; a new entry supersedes it with `kills:`.
- `rust-toolchain.toml` pins the channel; `Cargo.lock` committed; `--locked`
  everywhere. CI adds `dtolnay/rust-toolchain` at the pinned version, caches
  `~/.cargo` and `engine/target`. Cross-OS: the CI runner (`ubuntu-latest`,
  the image recorded per run) is measured; the operator's machine is a local
  run, not a row.
- STATUS.md: dated 2026-09-29 entry; P8 and P9 rows appear only once both
  cells are earned and CI is green; the "cross-language byte-identical twin"
  line under deferred decisions moves to resolved.
- Handoff entry at close, per `docs/handoff/HANDOFF.md` convention.

## 8. Scope walls (README before code)

- Go is the oracle. The engine's correctness is defined as equality with the
  ledger; a disagreement is an engine bug until a Go-side twin proves otherwise.
- No Rust gRPC server, no change to `read.proto`, no P7 rows touched. The
  engine implementing `Reader` behind `serve` is a later property.
- No writes from the engine, ever. `feed` cannot create, append, or truncate.
- No floats, no `unsafe`, no SIMD, no async.
- Numbers are measured and unclaimed. Performance vocabulary in claim
  positions stays forbidden; §6 names the one place it may appear.
- Named non-goals carried from 08-31: lot selection, FX, symbol changes,
  spin-offs, mergers, fractional shares, market connectivity, production.
- Lane 2 (ClickHouse) and Lane 3 (Kafka) remain promised by no one.

## 9. Build sequence

1. **Move + rename** — `ledger/` move, Go gate code included; gate: 27 rows
   unchanged. Operator renames the repo; session updates links and remote.
   One commit.
2. **Effective axis in Go** — `ReadAt`, `--effective`, `effective` key,
   generator's `effective_dates`, its filter and its visible-actions guard.
   Gate: P1–P7 rows unchanged; new unit tests green. One commit.
3. **Engine core → P8** — `feed`, `fold`, `snapshot`, `asof::read_recompute`,
   CLI `head`/`asof --no-index`; P8 live with `index_equals_recompute`
   trivially 0; then `index` and `read`; four twins; P8 claimable. Two commits
   (core, then index + twins).
4. **Concurrency → P9** — harness, three twins. One commit.
5. **Module-path flip** — gate: 36 rows unchanged. One commit.
6. **Bench** — `--scale`, pin, `bench`, STATUS section, `claimwords.py`.
   One commit. Handoff and learning entries close the build.

## 10. What this document does not decide

- `K`'s default beyond 1024, and whether checkpoints are memory-only or
  spilled to disk (memory-only for this build; disk is a v2 question).
- Whether the gate should also run at a `K` that makes tails longer than one
  record.
- Applied-state checkpoints with a validity rule: the faster index, not taken
  in this build.
- A twin for the visibility rule itself (a checkpoint restored although `E`
  hides part of its prefix). On base at `K` = 2, four of the twelve pairs
  have such a checkpoint for the live `index_equals_recompute` to compare;
  no twin plants the defect.
- A fixture in which an amendment is visible and its action is not.
- Whether the engine implements `Reader` behind `serve` (would reopen P7).
- The `AsOfRequest.effective` proto field (same reason).
- Whether P9's `T` and `R` scale with cores in CI (fixed at 8 × 16 for this
  build; a row that depends on the runner's core count is not reproducible).
- Lane 2 / Lane 3 timing.
