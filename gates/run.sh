#!/bin/sh
# MERIDIAN Lane 1 gate runner. Every step must pass; nothing here is optional.
set -eu
cd "$(dirname "$0")/.."
PY="${PYTHON:-python}"

echo "== fixtures: deterministic + fresh"
sh fixtures/generate_test.sh

echo "== import-pin (+ negative control)"
"$PY" gates/importpin.py
"$PY" gates/importpin.py --self-test

echo "== proto fresh"
# Regenerate into a scratch dir and byte-compare with the committed files:
# committed generated code must be exactly what the pinned tools produce.
# The Go module lives in ledger/, so every Go step runs with ledger/ as its
# working directory; the scratch dir is ledger/.protofresh.
rm -rf ledger/.protofresh && mkdir -p ledger/.protofresh
(cd ledger && go tool buf generate -o .protofresh)
gen=$(cd ledger && find .protofresh/api -name '*.pb.go' | sort)
[ -n "$gen" ] || { echo "FAIL proto fresh: buf generate produced no *.pb.go"; exit 1; }
for g in $gen; do
  c="${g#.protofresh/}"
  cmp "ledger/$g" "ledger/$c" || { echo "FAIL generated ledger/$c stale: run (cd ledger && go tool buf generate)"; exit 1; }
done
for c in $(cd ledger && find api -name '*.pb.go' | sort); do
  [ -f "ledger/.protofresh/$c" ] || { echo "FAIL committed ledger/$c has no regenerated counterpart: run (cd ledger && go tool buf generate)"; exit 1; }
done
rm -rf ledger/.protofresh
echo "ok proto fresh"

echo "== go vet"
(cd ledger && go vet ./...)

echo "== build"
mkdir -p bin
BIN="$PWD/bin/meridian"
case "$(uname -s 2>/dev/null || echo unknown)" in MINGW*|MSYS*|CYGWIN*) BIN="$BIN.exe" ;; esac
(cd ledger && go build -o "$BIN" ./cmd/meridian)

echo "== conformance pack pin"
# gates/conformance/PIN binds every vendored file of the conformance pack, by
# LF-normalised sha256, to one pushed commit of its governing text. Verified
# alone, before any vendored code is trusted to run: exit 2 names a file that
# is missing, unlisted or altered, and nothing below runs.
node gates/conformance/check.mjs --verify-pin

echo "== conformance pack self-test"
# The vendored checker's own negative controls, mutation included, run in the
# shape this repo ships (PIN beside it), before it is trusted over anything:
# a checker that has only ever said yes proves nothing. Node 20+, no
# dependencies.
node gates/conformance/test.mjs --mutate

echo "== claimability self-test"
# claimability.py's planted cases, in memory: per-check crediting (a live
# check that no twin RED as planted sets nonzero blocks CLAIMABLE; an
# UNEVALUABLE row makes the property UNEVALUABLE) and every kind of
# difference the agreement step below exists to report.
"$PY" gates/claimability.py --self-test

echo "== tests + gates"
rm -rf gates/out && mkdir -p gates/out
OUT="$PWD/gates/out"
(cd ledger && MERIDIAN_BIN="$BIN" MERIDIAN_VERDICT_DIR="$OUT" MERIDIAN_RUNNER="${MERIDIAN_RUNNER:-local}" go test ./... -count=1)

echo "== claimability"
# --status: STATUS.md may not mark a property the rows do not support.
# --check: the claimability table in STATUS.md is generated from the rows
# (between markers) and must equal a fresh rendering; status is derived,
# never authored.
"$PY" gates/claimability.py gates/out --status STATUS.md --check STATUS.md

echo "== conformance pack over the rows"
# The rows against the pinned pack. --verify-pin again, so this run is
# against the pinned bytes and not only the self-test above; --expect
# refuses a cell the rows lack, a cell they hold unexpectedly, or a twin
# mutation set other than the one gates/expect.json lists. Exit 1 refuses and
# exit 2 is unevaluable; set -e stops on either, with the pack's own message.
node gates/conformance/check.mjs gates/out --verify-pin --expect gates/expect.json

echo "== agreement"
# Two independent derivations of the same table: claimability.py and the
# pack. For every property they must give the same status and the same
# unfalsified checks; any difference fails with both --json outputs printed.
# Disagreement is a finding, not something to align by loosening either.
# The pack runs again only for --json, after the run above has exited 0, so a
# refusal is never hidden inside a JSON file. bin/ is build output.
rm -f bin/conformance-pack.json bin/claimability.json
node gates/conformance/check.mjs gates/out --verify-pin --expect gates/expect.json --json > bin/conformance-pack.json
"$PY" gates/claimability.py gates/out --json > bin/claimability.json
"$PY" gates/claimability.py --agree bin/conformance-pack.json bin/claimability.json
