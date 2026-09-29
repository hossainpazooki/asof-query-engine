# STATUS

State of record for MERIDIAN. The README defers to this file; nothing is
claimable anywhere unless it is claimable here.

- **2026-08-31** -- Design locked (`docs/2026-08-31-design.md`); repo scaffolded.
  **Build not started** -- sequencing vs. a competing build is an open pipeline
  decision. No code, no fixtures, no gates existed. Every cell in the table
  below was UNCLAIMED as of this date.

- **2026-09-01** -- Lane 1 built. `sh gates/run.sh` ends
  `ok lane1 claimable=6/6`, over **15 verdict rows** in `gates/out/`
  (P1, P2, P3, P5: one live + one twin each; P4: one live + two twins; P6: one
  live + three twins). Every live cell GREEN, every twin cell RED for exactly
  its planted reason with exact counts.

  **The evidence is committed and pushed.** Corrected 2026-09-03 at
  pick-up: an earlier version of this paragraph said the build lived only
  in an untracked working tree on `lane1-build` and that nothing had been
  pushed. That was true when written. The build landed in four
  evidence-first commits (`527f571` core, `b7f8cf5` fixtures, `70479fe`
  gates + CI, `be1101d` docs) and merged to `main` as PR #1 at `f8825ba`;
  `main` is level with `origin/main`. In this project only the human
  operator writes git history.

  **CI has now run, and P2's cross-OS leg is measured.** Corrected
  2026-09-01 after the merge: an earlier version of this entry said CI had
  never run and that the cross-OS leg was unmeasured, which was true when it
  was written and is no longer. `gates` run
  [33559490763](https://github.com/hossainpazooki/meridian/actions/runs/33559490763)
  on the merge commit `f8825ba` succeeded on `ubuntu-24.04` under Python
  3.14, printing `ok fixtures deterministic and fresh`, `ok import-pin
  self-test (all negative controls caught)`, and `ok lane1 claimable=6/6`
  with the same six rows. So the pinned snapshot hash reproduces on Linux as
  well as Windows, and the fixtures regenerate byte-identically there.
  Two earlier runs (33558902926 push, 33559063650 pull_request) also passed.

  Verdict rows are **regenerated on every run and are not committed** --
  `/gates/out/` is in `.gitignore`. They are build output, not a record; the
  record is this file. Twin counts per property live in
  `fixtures/base/manifest.json` under `p<N>.twin.expected_violations`.

- **2026-09-03** -- P7 (wire fidelity of the gRPC read API) built. `sh
  gates/run.sh` ends `ok lane1 claimable=7/7` over **18 verdict rows** (the
  15 above plus P7: one live + two twins). The API is `Head` / `AsOf` /
  `Reconcile`, read-only by construction of `api/meridian/v1/read.proto`
  (a descriptor test pins exactly those three unary methods). The gate is a
  gRPC client over an in-process listener that rehashes the bytes it
  receives and recomputes locally; twin 1 serves `fixtures/p2/mutated` as
  if it were base (self-consistent hashes, wrong content), twin 2 serves the
  right bytes under the wrong `snapshot_hash`. Design:
  `docs/2026-09-03-grpc-read-api-design.md`.

- **2026-09-06** -- **BASELINE registration retired; conformance to the
  governing text planned.** The "copy, not a translation" premise was
  measured: all 18
  rows from the 2026-09-06T16:53Z run, renamed to BASELINE's filename rule
  and hash-bound in a scratch source file, are refused by BASELINE's
  `check-ledger.mjs` at baseline `775978f` (every row: `rows must equal
  evaluated.no_future_accepted`, a PARALLAX check name bound into the
  checker; P4, P6, P7: `second twin cell`, the checker holds one twin per
  surface and lane), and `build.mjs` refuses a second surface. Evidence:
  `docs/learnings/2026-09-06-baseline-checker-refuses-meridian-rows.md`.
  The operator then ruled that BASELINE is not a catalog and that MERIDIAN
  is governed by **DATUM**, a private governing text for the operator's
  data-platform repositories (one discipline, one verdict-row schema, a
  conformance pack each governed repo vendors at a pinned revision). It is
  not public, so this file cites it for the ruling and for nothing else.
  Nothing in this repo's gates changed: `sh gates/run.sh` at `a087ab2`
  still ends `ok lane1 claimable=7/7` over 18 rows, and CI run 33817051877
  on `a087ab2` is green. **Not built:** the governing text's conformance
  pack does not exist yet, so vendoring it, the
  `gate_sha`/`gate_worktree`/`schema` emitter change, and the generated
  claimability block are all planned and blocked on it. Seed:
  `docs/handoff/2026-09-06-datum-adoption-seed.md`. *Reworded 2026-09-09:
  the governing text's private location and revision were removed from
  this entry; the ruling and the plan are unchanged.*

- **2026-09-09** -- **Conformance adopted; emitter writes the shared row.**
  The governing text's conformance pack is vendored under `gates/datum/`
  (checker, self-test, fixtures, reason vocabulary, a copy of the row
  schema) and bound file by file to one commit of the private governing
  text by `gates/datum/PIN`. Order of events, each measured: the vendored
  self-test ran green here (`13 positive, 46 negative, 17 reasons, 13
  rules mutated`); the checker over this repo's 18 rows was **red** with
  exactly five refusals per row (missing `schema`, `gate_sha`,
  `gate_worktree`; unknown `parallax_sha`, `parallax_worktree`) and
  nothing else; `TestEmit` was made red on the new key list; `Emit` in
  `gates/verdict.go` now writes the row's `schema` identifier and the
  emitter's commit and tree state as `gate_sha` / `gate_worktree`; the
  same checker over the regenerated rows exits 0 with seven surfaces
  CLAIMABLE and twin counts 1/1/1/2/1/3/2, the table `claimability.py`
  prints. `gates/run.sh` now runs the pack's self-test before the gates
  and its checker over `gates/out` after them with the pin verified, and
  fails if the two derivations disagree on how many surfaces are
  CLAIMABLE. The claimability table below is **generated** from the rows
  (`claimability.py --render`, between markers) and compared by `--check`
  in `run.sh`; nothing in it is typed. Two things found on the way and
  recorded in the governing text's repo: Go reserves `vendor/`, so the
  pack lives under `gates/`; the pack's self-test read its schema from a
  path that exists only in its home repo, fixed there. **Not yet:** the
  pin names a commit of the private repo that contains that fix, so
  `PIN` is written once that commit exists; CI here has not run with the
  pack; no MERIDIAN row has been copied into the pack's real-fixture
  corpus (that copy needs rows emitted from a clean tree at a pushed sha).
  Learnings: `docs/learnings/2026-09-09-emitter-writes-the-shared-row.md`.
  *Corrected 2026-09-15: none of the three "Not yet" items holds any
  longer. The pin was written in `35142fa` (`fix: write the pack pin, CI
  was red on an empty pin`, 2026-09-10). CI run
  [34438961409](https://github.com/hossainpazooki/meridian/actions/runs/34438961409),
  a push to `main` at `97d815c`, a descendant of `35142fa`, succeeded and
  printed `ok 18 rows conform, 7 surfaces` and `ok conformance pack agrees:
  claimable=7` from the pack's steps. The pack re-vendored on 2026-09-15
  carries one P7 live and one P7 twin row of this repo in its real-fixture
  corpus, byte-identical to the rows this repo emitted with `gate_sha`
  `97d815c` and `gate_worktree` clean.*
  *Moved 2026-09-15: the pack now lives under `gates/conformance/`, pinned
  by `gates/conformance/PIN`, and the count-only agreement check described
  above is replaced by a property-by-property comparison; see the
  2026-09-15 entry. The paths above are the record of where it was.*
  *Reworded 2026-09-15: a schema identifier literal was removed from this
  entry; the fact is unchanged.*

- **2026-09-15** -- **Conformance pack re-vendored; both derivations
  credit per check.** The pack was re-vendored from a pushed commit of the
  governing text into `gates/conformance/`, the directory the pack's own
  self-test builds as the shape a governed repo ships; the previous
  vendored directory was deleted. The commit is the first line of
  `gates/conformance/PIN` and is not repeated here. Order of events, each
  measured: the vendored tree was diffed against an archive of that commit
  (no difference; the schema file, placed beside `check.mjs`, compared
  byte for byte on its own); `--verify-pin` refused before the pin existed
  (`PIN missing beside check.mjs`, exit 2) and, after `--write-pin`, alone
  printed `ok pin verified`; the vendored self-test printed `ok
  conformance: 19 positive, 50 negative, 2 real, 20 reasons, 16 rules
  mutated, 1 crediting rule mutated`.

  The re-vendored pack credits per check (see Crediting rule). Over the 18
  rows emitted at `97d815c`, before `gates/claimability.py` changed, the
  pack derived CLAIMABLE for P5 and P7 only, while `gates/claimability.py`
  printed `ok lane1 claimable=7/7`: the two derivations disagreed.
  `gates/claimability.py` then gained a `--self-test`, run failing against
  its previous derivation first, and after it the same per-check rule,
  `--json` and `--agree`; its generated table gained an Unfalsified column.
  `gates/run.sh` now verifies the pin alone before the pack's self-test,
  runs `claimability.py --self-test` before the gates, runs the pack over
  the rows with `--verify-pin --expect gates/expect.json` (every surface
  and its exact twin mutations), and fails unless both derivations give
  every property the same status and the same unfalsified checks, printing
  both sides when they do not. No Go code, fixture or verdict-row field
  changed. `sh gates/run.sh` exited 0; its last 12 lines, unedited:

      ok lane1 claimable=2/7
      == conformance pack over the rows
      meridian-lane1-p1 lane1 PARTIAL (live GREEN, 1 twin; unfalsified: positions_match_manifest, unevaluable_match_manifest)
      meridian-lane1-p2 lane1 PARTIAL (live GREEN, 1 twin; unfalsified: chain_verifies, fresh_process_identical, pinned_hash_match)
      meridian-lane1-p3 lane1 PARTIAL (live GREEN, 1 twin; unfalsified: positions_match_manifest, three_histories, unevaluable_match_manifest, viewpoint_V1, viewpoint_V3)
      meridian-lane1-p4 lane1 PARTIAL (live GREEN, 2 twins; unfalsified: positions_match_manifest)
      meridian-lane1-p5 lane1 CLAIMABLE (live GREEN, 1 twin)
      meridian-lane1-p6 lane1 PARTIAL (live GREEN, 3 twins; unfalsified: unevaluable_match_golden)
      meridian-lane1-p7 lane1 CLAIMABLE (live GREEN, 2 twins)
      ok 18 rows conform, 7 surfaces
      == agreement
      ok conformance pack agrees: claimable=2

- **2026-09-15** -- **Twins for the unfalsified checks.** Six twin rows and one
  twin binary were added so that every live check of P1, P2, P4 and P6 is driven
  nonzero by a twin of its own property. `sh gates/run.sh` exits 0 and ends `ok
  conformance pack agrees: claimable=6` over **27 verdict rows** and seven
  surfaces; P3 is the only property short of CLAIMABLE. **No live row
  changed** -- the seven live rows of the entry above keep their check keys,
  their `evaluated` denominators, their `result`, their `content_hash` and
  their `scope`, compared row by row before and after.

  Per property:

  - **P1** -- new twin `fill_identity_key_drops_trade_id`: the at-most-once
    identity key reads `trade_id` as null, so every fill on a venue shares the
    first fill's identity and only that instrument keeps a position. The
    planted duplicate is still absorbed and the planted collision still
    refused, so the twin is confined to over-refusal. It drives
    `positions_match_manifest` to 4 and `unevaluable_match_manifest` to 1, each
    over an evaluated universe of 5. PARTIAL -> CLAIMABLE.
  - **P2** -- the one combined twin is replaced by three that each run the live
    row's own check function over one planted feed
    (`fill_price_mutated_rechained`, `buy_and_split_reordered_rechained`,
    `split_ratio_edited_not_rechained`), plus a fourth,
    `per_process_nonce_in_snapshot`, described below. `chain_verifies`,
    `fresh_process_identical` and `pinned_hash_match` are each driven nonzero.
    PARTIAL -> CLAIMABLE.
  - **P3** -- three new twins. `actions_admitted_by_effective_date` admits
    actions and amendments from beyond the visible prefix by their `effective`
    date, valid time read as knowledge time, and moves `viewpoint_V1` to 1 of
    14. `stale_original_terms_at_V3` never applies the amendment, so the action
    keeps the terms it was first published with, and moves `viewpoint_V3` to 3
    of 14. `viewpoint_ignored_end_of_feed_served` answers V1 and V2 with what
    the ledger knows at `end_seq`, collapsing the three histories into one, and
    is the twin that drives `three_histories` to 1 of 1.
    `positions_match_manifest` and `unevaluable_match_manifest` stay
    unfalsified; see below. P3 stays PARTIAL.
  - **P4** -- new twin `unpriceable_position_suppressed` deletes the withheld
    instrument's holding and its `unevaluable` declaration together, the
    tidier way to hide an unpriceable holding, and drives
    `positions_match_manifest` and `unevaluable_match_manifest` to 1 each of 5.
    PARTIAL -> CLAIMABLE.
  - **P6** -- new twin `price_event_withheld` withholds the first price event,
    so the fold must report that instrument unevaluable while the golden
    fixture states `unevaluable: []`; `unevaluable_match_golden` goes to 1 of
    3. PARTIAL -> CLAIMABLE.
  - **P5** and **P7** are unchanged.

  **Three defects in the new gate code, each found and fixed before the gate
  ran green**, each recorded because each would have shipped a wrong number or
  a wrong side effect:

  1. `p2Check` called `feed.Open` before any existence check, and `feed.Open`
     CREATES a missing feed
     (`docs/learnings/2026-09-03-feed-open-is-read-write.md`). Measured on a
     scratch copy of the tree with `fixtures/base/feed.jsonl` renamed away:
     before the fix the run left a **0-byte** `fixtures/base/feed.jsonl`
     behind and failed with `P2 live cell is RED: map[chain_verifies:0
     fresh_process_identical:0 pinned_hash_match:1]` -- a gate writing into
     `fixtures/` and then measuring the empty ledger it had just created.
     After the fix the file is absent and the run fails at emit with `P2 live
     check "chain_verifies" has no evaluated denominator (evaluated=0)`. The
     gate stats the feed first.
  2. The tampered twin published `evaluated` `chain_verifies` 71 while its
     chain walk stops at the planted break. `Evaluated` means the universe the
     check actually examined (`gates/verdict.go`), so the denominator now comes
     from the walk: 71 in the rows whose walk reaches the end of the feed, and
     **23** -- the break seq -- in that twin. The comment in
     `fixtures/generate.py` that said the denominator was `end_seq` in every P2
     row is corrected in place.
  3. The nonce twin's `content_hash` was its own snapshot, which by
     construction differs on every run, so the row pinned nothing. It now
     carries the sha256 of the base feed file with CRLF normalized to LF -- the
     input the twin replayed, hashed the way the tampered twin hashes its feed
     -- and its basis says the twin binary's snapshot bytes differ on every run
     by construction and are not recorded.

  **What `per_process_nonce_in_snapshot` credits, and what it does not.** The
  twin is a separate binary, `gates/p2nondet`: the production fold and snapshot
  plus a 128-bit `crypto/rand` nonce stamped into the snapshot document in
  every process. Nothing in `cmd/` or `internal/` changes for it, and the
  production binary the live cell runs never contains that code; its
  `planted.mutated_rows` is 0, because no input record changed. **What it
  credits is that the gate reports non-determinism when the replayer carries
  it** -- `fresh_process_identical` goes to 1 on a replayer whose two
  fresh-process runs really differ. It does **not** show that the production
  ledger can be non-deterministic, and nothing here claims it does.

  *Note 2026-09-15: this resolves the P2 item of finding C28 in
  `docs/2026-09-01-lane1-build-ledger.md`, which named `fresh_process_identical`
  unfalsified and routed it deliberately to a comparator-discriminates test
  rather than a twin, on the reasoning that planting non-determinism would mean
  breaking the thing under test. It does not: the defect is planted in the
  replayer, not in the fold, the way P7's twins plant theirs in the server. The
  build ledger is an immutable record and is not edited. The source sentence
  saying no honest test in this build could demonstrate the negation is
  superseded in `gates/p2_test.go` by the twin's own doc comment, which states
  what the twin shows and what it does not.*

  *Note 2026-09-15: this reverses decision C29 in the same build ledger, which
  routed P2's inverted twin polarity as a presentation fix -- a note in the
  row's `scope` -- and not a logic change. It is a logic change. P2's three
  feed twins now run the live row's own check function under the live row's own
  check names, so a nonzero count in a P2 twin row means what it means in every
  other twin row: the artifact failed that check. The polarity note is gone
  from the row's `scope`, and the tampered twin must still break at the planted
  seq or the gate fails.*

  **P3 stays PARTIAL, and that is a fact about the feed, not unfinished work.**
  `positions_match_manifest` and `unevaluable_match_manifest` are unfalsifiable
  by construction on this feed, because the position set and the unevaluable
  set are the same at V1, V2 and V3, so no point-in-time corporate-action
  defect can move either. Dropping the two checks from P3's live row was
  considered and rejected: P3 evaluates them over every viewpoint (evaluated
  15 = 5 instruments x 3 viewpoints, from the rows) where P1 and P4 evaluate
  them at V3 only (evaluated 5), so removal would lose the V1 and V2 coverage.
  Extending the base feed was rejected because the defect it would plant -- an
  action applied to an instrument never held -- is not a point-in-time defect.

  `sh gates/run.sh` exited 0; its last 14 lines, unedited:

      P6   | GREEN | RED*,RED*,RED*,RED* | YES
      P7   | GREEN | RED*,RED*          | YES
      ok lane1 claimable=6/7
      == conformance pack over the rows
      meridian-lane1-p1 lane1 CLAIMABLE (live GREEN, 2 twins)
      meridian-lane1-p2 lane1 CLAIMABLE (live GREEN, 4 twins)
      meridian-lane1-p3 lane1 PARTIAL (live GREEN, 4 twins; unfalsified: positions_match_manifest, unevaluable_match_manifest)
      meridian-lane1-p4 lane1 CLAIMABLE (live GREEN, 3 twins)
      meridian-lane1-p5 lane1 CLAIMABLE (live GREEN, 1 twin)
      meridian-lane1-p6 lane1 CLAIMABLE (live GREEN, 4 twins)
      meridian-lane1-p7 lane1 CLAIMABLE (live GREEN, 2 twins)
      ok 27 rows conform, 7 surfaces
      == agreement
      ok conformance pack agrees: claimable=6

## Crediting rule

A property is **CLAIMABLE** only when all three halves below hold, as
mechanically enforced by `gates/claimability.py` (which independently
re-derives each verdict from the row's own contents rather than trusting
the row's `result` label) and by the vendored conformance pack, whose
derivation `gates/run.sh` requires to agree with it property by property.
`Emit` in `gates/verdict.go` also refuses, at emit time, a row that breaks
a per-row condition of the first two halves (no checks, a missing
denominator, a RED live, a twin not RED as planted); it writes one row at a
time, so exactly one live row per property is enforced only by the two
derivations:

- **Live** -- exactly one live row for the property, with `result` GREEN. A
  live row whose `checks` map is empty, or that carries any check with a
  missing or non-positive `evaluated` denominator, is refused rather than
  credited: a gate that examined nothing is not a gate that passed.
- **Twin** -- **every** twin row RED, with its `checks` map **exactly equal**
  to its `planted.expected_violations` (full dict equality, over the union of
  both key sets, so an expectation with no computed check and a computed check
  with no expectation are both refusals), and at least one check non-zero.
- **Per check** -- every check key the live row reports is set nonzero by
  at least one twin row that is RED as planted. A live check that no such
  twin drives nonzero is **unfalsified**: its 0 has never been shown able
  to be anything else. Any unfalsified check makes the property
  **PARTIAL**, never CLAIMABLE, and is named in the generated table below.
  A property with an UNEVALUABLE row, its live row included, derives
  **UNEVALUABLE** and is not credited per check at all.

*Corrected 2026-09-15: this rule had two halves until the per-check half
was added by operator ruling; both derivations enforce it as of the
2026-09-15 entry. `Emit` writes one row at a time, so it cannot apply it.*

**Most properties have more than one twin, and all of them must hold** -- one
red twin does not credit a property that plants three defects. A gate that has
never run red proves nothing. *Corrected 2026-09-15: this line read "P4 and P7
have two twins each and P6 has three". It is stale as of the entry above -- P1
and P7 have two, P4 three, P2, P3 and P6 four, P5 one. Those numbers are not
retyped as a rule here: the generated table below derives each property's twin
count from its rows, `gates/expect.json` names every surface and its exact twin
mutations, and `gates/run.sh` refuses the rows if they do not match it.*

Verdicts are emitted as `GATE_VERDICT` rows in BASELINE's ledger schema
(surface, lane, cell, result, per-check planted-vs-caught counts, repo sha +
worktree state, content hash + basis, replay command; twin rows carry a 17th
key `planted`), so earned cells can register into the BASELINE catalog as a
copy, not a translation. *Corrected 2026-09-06: the schema sentence holds
(live-row key set equals BASELINE's 16 keys exactly); the registration
clause does not -- BASELINE's checker refuses the rows and the operator
retired registration (see the 2026-09-06 entry). The rows' target is now
the governing text's conformance pack, which keeps these field names except
`parallax_sha`/`parallax_worktree` (renamed to the emitter-neutral
`gate_sha`/`gate_worktree`) and adds a `schema` key; that emitter change is
not made.* *Corrected 2026-09-09: the emitter change is made; see the
2026-09-09 entry. Rows carry `schema`, `gate_sha`, `gate_worktree`; the
vendored pack under `gates/datum/` checks every row in `gates/run.sh`.*
*Moved 2026-09-15: the vendored pack is now under `gates/conformance/`.*

## Claimability -- Lane 1 (local, Go core)

<!-- meridian:claimability:begin -->
generated by `python gates/claimability.py gates/out --render` from the rows of one run; `gates/run.sh` compares it with `--check STATUS.md`

| # | Property | Live | Twin | Status | Unfalsified |
|---|---|---|---|---|---|
| P1 | At-most-once fill ingestion (2 twins) | GREEN | RED | CLAIMABLE | - |
| P2 | Deterministic replay, byte-identical snapshot (4 twins) | GREEN | RED | CLAIMABLE | - |
| P3 | PIT-correct corporate actions (incl. amendment) (4 twins) | GREEN | RED | PARTIAL | `positions_match_manifest`, `unevaluable_match_manifest` |
| P4 | Fail-closed valuation (3 twins) | GREEN | RED | CLAIMABLE | - |
| P5 | Reconciliation proven able to fail | GREEN | RED | CLAIMABLE | - |
| P6 | Portfolio math (average cost, P&L) (4 twins) | GREEN | RED | CLAIMABLE | - |
| P7 | Wire fidelity of the gRPC read API (2 twins) | GREEN | RED | CLAIMABLE | - |
<!-- meridian:claimability:end -->

Every RED above is red **for its planted reason with its exact planted
counts**; a merely-red twin does not credit a cell.

The Twin column holds one word per property because that is the schema
`gates/claimability.py` parses, but **most properties have more than one twin
row**, and a single RED there means **every** one of them went red as planted
-- the per-twin rows are in `gates/out/` and the per-property twin counts are
in the pack's output pasted in the dated entry above.
*Corrected 2026-09-15: this paragraph read "P4 and P7 have two twin rows each
and P6 has three", and the sentence that followed it listed P4's two, P6's
three and P7's two twin mutations by name. Both are stale as of that entry.
The list of twin mutations is not repeated here: `gates/expect.json` is where
every surface's exact twin mutations are named, and `gates/run.sh` refuses the
rows if they do not match it, so a copy here could go stale without any gate
noticing.*

The Unfalsified column is generated with the rest of the table: for each
property, the checks its live row reports that no twin RED as planted sets
nonzero. A name there is what holds that property at PARTIAL.

## Honest limits

Measured facts about what the seven cells above do and do not establish. These
are not caveats added for modesty -- each one was found by measurement during
the build, and each bounds a claim that would otherwise be read as stronger
than the evidence. **The first four trace to an entry in
`docs/2026-09-01-lane1-build-ledger.md`, which records the measurement that
established each, and the fifth (P7) to the 2026-09-03 design and final
review** -- so they are checkable rather than merely asserted. The
sixth (no production claim) is not a finding at all: it is a scope wall
declared in `docs/2026-08-31-design.md` section 6 before any code existed, and
it is checkable a different way -- by the absence of anything in the repo that
would falsify it. *Noted 2026-09-15: the third bullet's list of unfalsified
checks is now the pack's own output over the rows, pasted there.*

- **P5 demonstrates CROSS-IMPLEMENTATION AGREEMENT, not independent
  verification.** The Python naive fold is a **same-contract
  reimplementation**: it and the Go fold were written from the same written
  contract, so a contract-level misunderstanding is reproduced identically on
  both sides and the gate stays green. The import-pin
  (`gates/importpin.py`) gives **structural** independence -- no shared code --
  and never epistemic independence. `fixtures/p6/golden.json` is the only
  artifact in the build that escapes this circularity, because a human derived
  its twelve leaves from the arithmetic. P5's claim is "reconciles against an
  independently-implemented fold", never "independently verified correct".

- **The feed's hash chain does not protect its tail.** A payload edit to the
  **last** record, a truncation to a shorter valid prefix, and a tail record
  replaced by a blank line are all **accepted** by `feed.Open`. Tail integrity
  comes from the externally pinned snapshot hash
  (`fixtures/base/snapshot.sha256`), not from the chain. The package comment
  in `internal/feed/feed.go` states four such limits with the reasoning and the
  measured shapes; read it there rather than trusting a summary.

- **Live checks that no twin of their own property drives nonzero.** *Corrected
  2026-09-15: this bullet said three checks were never falsified anywhere
  in the build (`fresh_process_identical`, `three_histories`,
  `unevaluable_match_golden`). That undercounted. The re-vendored pack, over
  the 18 rows emitted at `97d815c`, names these live checks unfalsified --
  set nonzero by no twin RED as planted -- per property (excerpt: lines 1-4
  and 6 of the output of `node gates/conformance/check.mjs gates/out
  --verify-pin --expect gates/expect.json`):*

      meridian-lane1-p1 lane1 PARTIAL (live GREEN, 1 twin; unfalsified: positions_match_manifest, unevaluable_match_manifest)
      meridian-lane1-p2 lane1 PARTIAL (live GREEN, 1 twin; unfalsified: chain_verifies, fresh_process_identical, pinned_hash_match)
      meridian-lane1-p3 lane1 PARTIAL (live GREEN, 1 twin; unfalsified: positions_match_manifest, three_histories, unevaluable_match_manifest, viewpoint_V1, viewpoint_V3)
      meridian-lane1-p4 lane1 PARTIAL (live GREEN, 2 twins; unfalsified: positions_match_manifest)
      meridian-lane1-p6 lane1 PARTIAL (live GREEN, 3 twins; unfalsified: unevaluable_match_golden)

  Each reads 0 in its property's live row and reads 0, or is absent, in
  every twin row of that property (`unevaluable_match_manifest` is set
  nonzero only by a P4 twin, which credits nothing for P1 or P3), so by this
  project's own rule
  its 0 is not evidence, and under per-check crediting each holds its
  property at PARTIAL. The three named before each have a test proving that
  its **comparator discriminates** -- driven with knowingly different
  inputs, it reports the difference (`TestReplaysIdenticalDiscriminates`,
  `TestP3ThreeHistoriesDiscriminates`, `TestSetEqualityOverUniverseTable`).
  That is **not** the same as a twin proving the **ledger can produce the
  defect**. No non-determinism was planted into the fold, deliberately; what
  is demonstrated is that the check could report one, not that the system
  could commit one.

  *Corrected again 2026-09-15, by the twin set in the entry above: the list is
  now one property and two checks. Over the 27 rows of that run, the pack's
  per-property output in full (`node gates/conformance/check.mjs gates/out
  --verify-pin --expect gates/expect.json`):*

      meridian-lane1-p1 lane1 CLAIMABLE (live GREEN, 2 twins)
      meridian-lane1-p2 lane1 CLAIMABLE (live GREEN, 4 twins)
      meridian-lane1-p3 lane1 PARTIAL (live GREEN, 4 twins; unfalsified: positions_match_manifest, unevaluable_match_manifest)
      meridian-lane1-p4 lane1 CLAIMABLE (live GREEN, 3 twins)
      meridian-lane1-p5 lane1 CLAIMABLE (live GREEN, 1 twin)
      meridian-lane1-p6 lane1 CLAIMABLE (live GREEN, 4 twins)
      meridian-lane1-p7 lane1 CLAIMABLE (live GREEN, 2 twins)

  *P3 is the only PARTIAL, and its two checks are unfalsifiable by
  construction on this feed -- the position set and the unevaluable set are the
  same at V1, V2 and V3 -- not unfalsified for want of a twin; the entry above
  gives the reasoning and why neither check is dropped.
  `three_histories` and `unevaluable_match_golden` are now each driven nonzero
  by a twin of their own property, so the comparator-discriminates tests named
  above are no longer what carries them; those tests remain, covering cases no
  twin reaches.
  The sentence "No non-determinism was planted into the fold, deliberately" is
  superseded for P2, and only in the half a reader is most likely to overread:
  non-determinism IS now planted, but in a twin binary (`gates/p2nondet`) that
  calls the production fold and then stamps a per-process nonce into the
  snapshot document it marshals -- the fold itself is untouched, nothing in
  `cmd/` or `internal/` changed for it, and the live cell still runs the
  production binary. `fresh_process_identical` goes to 1 on that binary. What
  is still NOT shown is that the production ledger could commit a
  non-deterministic replay; no claim anywhere in this repo says it could.*

- **Fixture reproducibility is interpreter-dependent.** The fixtures were
  generated on **Python 3.14**, the only interpreter on the build machine, and
  CI pins that exact version. `fixtures/generate.py` drives its stream through
  `random.Random`, whose `randint` and `choice` are **not** contractually
  stable across CPython versions (only `getrandbits`/`random()` are). On a
  different interpreter MINOR the fixtures may not reproduce.
  `fixtures/generate_test.sh` is the detector -- and a failure there would
  present as fixture staleness, not as an interpreter difference.
  Narrowed 2026-09-01: the fixtures are now known to reproduce byte-identically
  on **two operating systems** under Python 3.14 (Windows locally, ubuntu-24.04
  in CI run 33559490763). What remains unverified is any OTHER Python minor;
  the OS leg is no longer in doubt.

- **P7 measures an in-process listener, not a network.** Its verdict rows
  come from a `bufconn` transport inside one test process. Nothing is
  claimed about TCP behaviour under load, TLS, authentication,
  authorisation, concurrency, or backpressure. `meridian serve` over
  loopback TCP is exercised by one smoke test (`Head` answered over the
  wire) that emits no row and ends the process with a kill, so graceful
  stop on SIGINT/SIGTERM is untested (Windows cannot deliver those signals
  to a child process). "Read-only" is a statement about the proto -- three
  methods, all reads, pinned by a descriptor test -- not about the serving
  process's filesystem permissions.
  Measured at final review 2026-09-03, none fixed, all stated: (1) a read
  opens the feed read-write -- `feed.Open` runs `os.MkdirAll` and opens with
  `O_CREATE|O_RDWR`, so `serve` needs write permission on the feed to answer
  reads and would fail on a read-only mount; (2) the exists-then-open guard
  in `FeedReader` has a race the one-shot CLI did not: delete the feed
  between the stat and the open and the server creates an empty feed and
  answers a clean empty ledger; (3) no message-size bound is set, so gRPC's
  4 MB client default caps the snapshot `AsOf` can deliver, and a ledger
  past it would fail with a code absent from the design's status table (the
  base feed is 16 KB; the boundary is unmeasured); (4) a statement's
  `as_of_seq` is parsed and never compared to the requested seq, in the
  gRPC path exactly as in the CLI -- pre-existing, now remotely reachable;
  (5) two comparison legs are never driven non-zero by any twin: the
  `records` leg of `head_matches_local` and the `compared` leg of
  `reconcile_matches_local` (both twins keep 71 records and compare 11
  fields), so their zeros are not evidence; (6) the `proto fresh` step had
  only ever run on Windows when this was written. Narrowed 2026-09-03 after
  the push: CI runs 33815809956 (push) and 33815870881 (pull_request) on
  `4d2b21d`, `ubuntu-24.04`, Go 1.26, Python 3.14, printed `ok proto fresh`
  (buf and its module graph fetched from the Go module proxy on the cold
  runner and built there, about 75 s between the import-pin line and the
  proto-fresh line) and
  `ok lane1 claimable=7/7` with the P7 row, so the codegen pin and the
  seven cells reproduce on Linux as well as Windows.
  *Noted 2026-09-15: P7 derives CLAIMABLE under per-check crediting, and
  that does not reach (5). Crediting reads one integer per check key, and
  `p7Check` in `gates/p7_test.go` folds both head fields (`records` and
  `prefix_hash`) into `head_matches_local`, and the mismatch-list
  difference plus the `compared` field into `reconcile_matches_local`. The
  wrong-feed twin sets both keys nonzero through their other legs (it keeps
  the record count, per `fixtures/generate.py`), so both keys are credited
  while the `records` and `compared` legs stay undriven.*

- **No production claim.** Synthetic, versioned fixtures only. The ledger has
  run nowhere that matters, against no market data, no custodian, and no
  counterparty.

## Lanes 2-3 (empty on purpose)

| Lane | Surface | Live | Twin | Status |
|---|---|---|---|---|
| 2 | ClickHouse as-of read surface (behind Reader protocol) | UNCLAIMED | UNCLAIMED | not promised |
| 3 | Kafka feed transport (guarantees re-proven end-to-end) | UNCLAIMED | UNCLAIMED | not promised |

## Deferred decisions

- Build order vs. intent-workbench -- **resolved** 2026-09-01 by the operator
  invoking the Lane 1 build. No other deferred decision is treated as resolved.
- **D2, chain-covered residue** -- a measured, better feed-chain contract
  (`prev` covering everything since the last record) that closes junk
  injection and the lazy tail forgery with no bound, is byte-identical for
  clean feeds, and therefore migrates rather than rewrites. Deliberately **not
  taken** in Lane 1 because `prev` is a cross-language contract and the Python
  generator was mid-hardening. Awaiting an operator decision; see
  `docs/handoff/2026-09-01-lane1-build.md`.
- gRPC read API -- **resolved** 2026-09-03: built read-only as P7; see the
  dated entry above and `docs/2026-09-03-grpc-read-api-design.md`.
- BASELINE registration of the claimable cells -- **retired** 2026-09-06 by
  operator ruling (BASELINE is not a catalog); replaced by conformance to the
  governing text, planned and blocked on its pack. See the dated entry above.
  *Corrected 2026-09-15: neither planned nor blocked any longer. The pack is
  vendored (see the 2026-09-09 entry), its pin was written in `35142fa`, and
  it runs in `gates/run.sh` from `gates/conformance/` (see the 2026-09-15
  entry).*
- Cross-language byte-identical twin -- v2 candidate once the snapshot format
  is stable.
- Whether `canon.Marshal` should ever accept non-ASCII (it refuses today, by
  decision, since the spec restricts every string to ASCII).
