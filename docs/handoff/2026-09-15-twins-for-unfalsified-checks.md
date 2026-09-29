# Handoff -- twins for the unfalsified checks; P2's twin split under the live names

2026-09-15. Newest MERIDIAN commit this brief describes: **`df4cf0f`**
(`docs: conformance section and re-vendor handoff`, main). Pick-up measures drift from
here. Everything below is **uncommitted** on top of it.

This entry supersedes **Open / next item 2** of
`2026-09-15-conformance-re-vendored.md`, which listed the unfalsified checks of
five properties as not started. Four of the five are done, by twins; the fifth
(P3) is ruled PARTIAL and has nothing to build. Its other items stand.

**The governing text** is DATUM, a private governing text for the operator's
data-platform repositories. This repo vendors its conformance pack and cites it
for nothing else; every claim below is checkable from this repo, its CI, and
the vendored pack.

## Current state

- **built, uncommitted** -- six new twin rows and one new twin binary. P1 gains
  `fill_identity_key_drops_trade_id`; P2's single combined twin becomes three
  feed twins that run the live row's own check function
  (`fill_price_mutated_rechained`, `buy_and_split_reordered_rechained`,
  `split_ratio_edited_not_rechained`) plus `per_process_nonce_in_snapshot`; P3
  gains `actions_admitted_by_effective_date`, `stale_original_terms_at_V3` and
  `viewpoint_ignored_end_of_feed_served`; P4 gains
  `unpriceable_position_suppressed`; P6 gains `price_event_withheld`. P5 and P7
  are untouched. Six properties derive CLAIMABLE, P3 is PARTIAL.
  re-verify: `sh gates/run.sh 2>&1 | tail -1` -> `ok conformance pack agrees: claimable=6`
- **built, uncommitted** -- 27 verdict rows over seven surfaces (7 live, 20
  twins), both derivations agreeing property by property and on the unfalsified
  checks.
  re-verify: `node gates/conformance/check.mjs gates/out --verify-pin --expect gates/expect.json | tail -1` -> `ok 27 rows conform, 7 surfaces`
- **measured, no live row changed** -- the seven live rows keep the check keys,
  `evaluated` denominators, `result`, `content_hash` and `scope` they had at
  `df4cf0f`, compared row by row before and after. Six of the seven carry the
  committed snapshot pin as their `content_hash`; P6's carries its own.
  re-verify: `git diff --stat -- fixtures/base/feed.jsonl fixtures/base/snapshot.sha256` -> no output, and `python -c "import json,glob; print(sorted({json.load(open(f))['content_hash'] for f in glob.glob('gates/out/*-live-*.json')}))"` -> the pin plus P6's one hash
- **found red, fixed, green** -- `p2Check` called `feed.Open` before any
  existence check, and `feed.Open` creates a missing feed. On a scratch copy
  with `fixtures/base/feed.jsonl` renamed away, the pre-fix gate left a 0-byte
  `fixtures/base/feed.jsonl` behind and failed with `P2 live cell is RED:
  map[chain_verifies:0 fresh_process_identical:0 pinned_hash_match:1]`; the
  post-fix gate leaves no file and fails at emit with `P2 live check
  "chain_verifies" has no evaluated denominator (evaluated=0)`. The real
  `fixtures/base/` was never touched: the rename and both runs happened in
  copies outside the repo.
  re-verify: `git grep -n 'os.Stat(feedPath)' -- gates/p2_test.go` -> one line, above the `feed.Open` call
- **built, uncommitted** -- `chain_verifies`' denominator now comes from the
  chain walk, not from the manifest: 71 where the walk reaches the end of the
  feed, 23 (the planted break seq) in the tampered twin, which previously
  published 71 for a walk that stopped at 23.
  re-verify: `python -c "import json,glob;print([(json.load(open(f))['planted']['mutation'], json.load(open(f))['evaluated']['chain_verifies']) for f in glob.glob('gates/out/*p2-twin*.json')])"` -> `split_ratio_edited_not_rechained` at 23, the other three at 71
- **built, uncommitted** -- the nonce twin row's `content_hash` is the sha256
  of the base feed file with CRLF normalized to LF, the input it replayed, not
  its own snapshot (which carries a fresh nonce in every process and would make
  the row unstable across runs). Its basis says so.
  re-verify: `python -c "import json,glob,hashlib;h='sha256:'+hashlib.sha256(open('fixtures/base/feed.jsonl','rb').read().replace(b'\r\n',b'\n')).hexdigest();print([json.load(open(f))['content_hash']==h for f in glob.glob('gates/out/*p2-twin*.json') if json.load(open(f))['planted']['mutation']=='per_process_nonce_in_snapshot'])"` -> `[True]`
- **built, uncommitted** -- fixtures regenerate deterministically with the new
  twin trees and manifest keys.
  re-verify: `sh fixtures/generate_test.sh` -> `ok fixtures deterministic and fresh`
- **built, uncommitted** -- STATUS.md: a new dated entry (per-property changes,
  the three defects, the C28 P2 resolution and the C29 reversal as dated notes,
  the P3 ruling, the run's last 14 lines); dated in-place corrections where the
  twin counts were typed and in the unfalsified-checks honest limit, whose
  pasted pack output is now the current one; the generated claimability block
  re-rendered. `gates/claimability.py` and `fixtures/generate.py` comments
  corrected where they named three P6 twins, an `end_seq` denominator, or an
  absent `unevaluable` key in the P6 golden.
  re-verify: `python gates/claimability.py gates/out --status STATUS.md --check STATUS.md` -> `ok STATUS.md generated claimability block is fresh`
- **checked, not edited** -- README.md. Every sentence about twins and
  properties was read against the rows: it carries no counts, it names no twin
  mutations, and its claimability sentence ("every check the live gate reports
  has been driven nonzero by at least one of those twins") is exactly the rule
  the run enforces. Nothing in it is false, so it is unchanged.
  re-verify: `git diff --stat -- README.md` -> no output
- **not run** -- CI on any of this; it runs on the operator's push.

## Locked decisions

1. **A twin may plant its defect in a sibling the gate runs in place of the
   component under test.** `gates/p2nondet` is the production fold and snapshot
   plus a per-process `crypto/rand` nonce in the snapshot document; nothing in
   `cmd/` or `internal/` changes for it, and the live cell always runs the
   production binary. Operator ruling. Its `planted.mutated_rows` is 0 because
   no input record changed.
2. **What that twin credits is bounded, and the bound is written down.** It
   credits that the gate reports non-determinism when the replayer carries it.
   It does not show that the production ledger can be non-deterministic, and no
   text in this repo may say it does.
3. **P2's twins run the live row's own check function under the live row's own
   check names.** This reverses C29, which routed the inverted polarity as a
   presentation fix. A tampered feed must still break at the planted seq or the
   gate fails. Operator ruling.
4. **P3 stays PARTIAL and names exactly two checks.** `positions_match_manifest`
   and `unevaluable_match_manifest` are unfalsifiable by construction on this
   feed; neither is dropped, because P3 is the only property that evaluates
   them at V1 and V2. Operator ruling; the reasoning is in STATUS.md and in
   `docs/learnings/2026-09-15-p3-set-checks-cannot-move-on-this-feed.md`.
5. **`Evaluated` is measured, never assumed.** A denominator read from the
   manifest describes the fixture; the check's denominator is what the check
   examined. Where the two can differ, the gate takes the measured one.
6. **Carried, unchanged:** every decision of `2026-09-15-conformance-re-vendored.md`,
   including `gates/expect.json` as part of the pack's run, per-property
   agreement, "fix majors, list minors", and the process lock -- say "one more
   commit coming" before pushing, because the operator merges within minutes.

## Reuse map

- `gates/p2_test.go`, `p2Check` -- one function for the live cell and every P2
  twin; the shape to copy when a property's twins should report the live row's
  own keys. `p2NondetBinary` builds the twin binary and is never `MERIDIAN_BIN`.
- `gates/p2nondet/main.go` -- the twin-binary pattern: a `main` under `gates/`
  that answers only the subcommands the gate calls, in the production CLI's
  output shapes, with one planted defect and a package doc stating it.
- `fixtures/generate.py`, the manifest block at the end -- every twin's
  `expected_violations`, measured here and compared by `Emit`. A new twin
  starts there and in `gates/expect.json`.
- `gates/expect.json` -- the only place twin mutations are enumerated. STATUS.md
  deliberately no longer lists them.
- `gates/claimability.py` -- `unfalsified()`, `derive()`, `--render`, `--check`,
  `--agree`, `--self-test`.
- `docs/2026-09-01-lane1-build-ledger.md` -- findings C1-C29, immutable. C28 and
  C29 are resolved and reversed respectively by dated notes in STATUS.md, not by
  edits there.

## Invariants

- **`run.sh` clears `gates/out/`; a bare `go test` appends.** Duplicates read as
  refusals from the pack and as `duplicate verdict rows` from `claimability.py`.
- **A gate never writes into `fixtures/`.** `feed.Open` creates and opens
  read-write, so every gate that takes a feed path stats it first.
- **Disagreement between `Emit`, `claimability.py` and the pack is a finding.**
  Do not align by loosening any of the three.
- **Crediting reads one integer per check key.** Legs folded inside a key stay
  invisible to it; P7's `records` and `compared` legs remain undriven under a
  CLAIMABLE.
- **Exit 2 from the pack is unevaluable and is never coerced.**
- **Nothing is CLAIMABLE anywhere unless STATUS.md says so**, and the generated
  block derives that from the rows.
- **The governing text is private; this repo cites only what a reader can
  check.** One naming per document; no paths, revisions or rule numbers.
- **Only the operator writes git history.**

## Open / next

1. **Commit and push; watch CI.** New directories are untracked and need
   `git add`: `gates/p2nondet/`, `fixtures/p1/twin-identity-collapse/`,
   `fixtures/p3/twin-effective-date/`, `fixtures/p3/twin-stale-terms/`,
   `fixtures/p3/twin-viewpoint-ignored/`, `fixtures/p4/twin-unpriced-suppressed/`,
   `fixtures/p6/twin-unpriced/`. Modified: `STATUS.md`, `README.md` is NOT
   modified, `fixtures/base/manifest.json`, `fixtures/generate.py`,
   `gates/claimability.py`, `gates/expect.json`, `gates/p1_test.go`,
   `gates/p2_test.go`, `gates/p3_test.go`, `gates/p4_test.go`,
   `gates/p6_test.go`, plus the two docs entries and their indexes.
2. **P3: nothing to build.** Superseding item 2 of the previous handoff for this
   property: the two checks are unfalsifiable on this feed and stay named. If
   the operator ever wants P3 CLAIMABLE, the only honest routes are a feed whose
   position or unevaluable set genuinely varies by viewpoint under a
   point-in-time defect, or a ruling that the two checks are not part of what P3
   claims. Neither is taken.
3. **Unchanged from the previous handoff.** The optional third P7 twin for the
   two folded legs; lanes 2-3; the chain-covered-residue decision; generator
   RNG via `getrandbits`. (The previous handoff also listed stale comments in
   `gates/verdict.go` and `gates/verdict_test.go` naming the old vendored
   directory; both name `gates/conformance/` at the commit this entry
   describes, so there is nothing to fix.)

**Known limits (minors, listed, not fixed):** `per_process_nonce_in_snapshot`
is red as planted with probability 1 - 2^-128 rather than by construction --
two equal 128-bit draws would leave `fresh_process_identical` at 0 while
`pinned_hash_match` stays 1 (the nonce field is always in the hashed bytes),
so the row would still be red but not as planted and the gate would refuse
it, which is the safe direction but is still a probabilistic red as planted;
`TestReplaysIdenticalDiscriminates` stays because the twin drives both legs of
`fresh_process_identical` together and cannot reach the one-leg cases. The
tampered twin's denominator is the chain error's seq, which equals the records
walked for a prev mismatch but would overstate it for a gap error; no twin here
plants a gap, and the twin asserts its exact break seq. `gates/p2nondet` is
built only by the P2 gate, so `go vet` and `go build ./...` cover it but nothing
else exercises it. STATUS.md's 2026-09-01 entry still records the twin counts of
that date, as a dated record, and older handoffs keep the pre-move pack path in
their re-verify lines.
