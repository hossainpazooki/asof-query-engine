package gates

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/hossainpazooki/meridian/internal/feed"
)

func binary(t *testing.T) string {
	t.Helper()
	if b := os.Getenv("MERIDIAN_BIN"); b != "" {
		return b
	}
	bin := filepath.Join(t.TempDir(), "meridian")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	if out, err := exec.Command("go", "build", "-o", bin, "../cmd/meridian").CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, out)
	}
	return bin
}

// p2NondetBinary builds ledger/gates/p2nondet, P2's non-deterministic replay twin
// (see its package doc), into a temp dir. It is never MERIDIAN_BIN: the live
// cell always runs the production binary.
func p2NondetBinary(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "p2nondet")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	if out, err := exec.Command("go", "build", "-o", bin, "./p2nondet").CombinedOutput(); err != nil {
		t.Fatalf("build p2nondet: %v\n%s", err, out)
	}
	return bin
}

// freshProcessSnapshot invokes the CLI as two separate subprocess calls (a
// fresh process for each) and returns the pinned-hash-format string
// ("sha256:<hex>", first whitespace field of `snapshot`'s stdout) and the raw
// canonical snapshot bytes `asof` writes to stdout. Both commands replay the
// same feed prefix independently, so any statefulness leaking between calls
// within one process can never be mistaken for determinism here.
func freshProcessSnapshot(t *testing.T, bin, feedPath string) (hash string, bytes []byte) {
	t.Helper()
	out, err := exec.Command(bin, "asof", "--feed", feedPath).Output()
	if err != nil {
		t.Fatalf("asof: %v", err)
	}
	h, err := exec.Command(bin, "snapshot", "--feed", feedPath, "--out", t.TempDir()).Output()
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	return strings.Fields(string(h))[0], out
}

// replaysIdentical is the comparator behind fresh_process_identical, factored
// out so TestReplaysIdenticalDiscriminates can drive it directly with every
// combination of equal and different hashes and bytes. The twin
// per_process_nonce_in_snapshot drives it end to end instead, through p2Check,
// on a binary whose two fresh-process replays really differ.
func replaysIdentical(h1 string, b1 []byte, h2 string, b2 []byte) bool {
	return h1 == h2 && bytes.Equal(b1, b2)
}

// p2Check computes P2's three checks for one binary over one feed. The live
// cell and every twin call it (the shape of p7Check), so a key is the same
// predicate over the same kind of input in every row that reports it, and a
// nonzero count means the same thing in every row: the feed or the binary
// under test failed that check.
//
// chain_verifies is decided first, in-process, over the feed alone. When the
// chain does not verify, the CLI refuses the feed (exit 2) and no snapshot
// exists, so fresh_process_identical and pinned_hash_match are left out of
// the counts: a 0 would claim they were checked and matched, a 1 that they
// were checked and diverged. A live row reduced that way is RED, and Emit
// refuses it.
func p2Check(t *testing.T, bin, feedPath, pin string) (Counts, string, error) {
	t.Helper()
	c := Counts{Checks: map[string]int64{}, Evaluated: map[string]int64{}}
	// chain_verifies' denominator is the number of records the chain walk
	// actually examined, which is the whole feed only when the walk reached
	// the end of it. A walk that stops at a break examined the records up to
	// and including the one it stopped on and nothing after it, so the
	// denominator is filled in below, from the walk, rather than from the
	// manifest's end_seq.
	c.Checks["chain_verifies"], c.Evaluated["chain_verifies"] = 0, 0
	// feed.Open runs os.MkdirAll and opens O_CREATE|O_RDWR
	// (docs/learnings/2026-09-03-feed-open-is-read-write.md), so calling it on
	// a path that does not exist CREATES an empty feed and then reports a
	// clean, empty ledger. A gate must not write into fixtures/, and a missing
	// fixture must not be convertible into a zero that reads like a pass:
	// stat first and fail here.
	if _, err := os.Stat(feedPath); err != nil {
		c.Checks["chain_verifies"] = 1
		return c, "", err
	}
	f, err := feed.Open(feedPath)
	if err != nil {
		c.Checks["chain_verifies"] = 1
		// The walk stopped at the first record that broke the chain. Open
		// accepts only contiguous seqs starting at 1, so every record it
		// accepted sits at its own seq and the seq it stopped on is the count
		// of records it examined. (The one shape where those two numbers
		// differ is a gap, whose ChainError carries the offending record's
		// own seq rather than its position; no twin here plants a gap, and
		// the tampered twin asserts the exact break seq it does plant.) Any
		// other failure examined no records at all, and a 0 denominator is
		// refused by Emit rather than published as a pass.
		var ce *feed.ChainError
		if errors.As(err, &ce) {
			c.Evaluated["chain_verifies"] = ce.Seq
		}
		return c, "", err
	}
	c.Evaluated["chain_verifies"] = f.Len()
	f.Close()
	h1, b1 := freshProcessSnapshot(t, bin, feedPath)
	h2, b2 := freshProcessSnapshot(t, bin, feedPath)
	// fresh_process_identical examines two things about the pair of fresh
	// runs: hash equality and byte equality.
	c.Checks["fresh_process_identical"], c.Evaluated["fresh_process_identical"] = 0, 2
	if !replaysIdentical(h1, b1, h2, b2) {
		c.Checks["fresh_process_identical"] = 1
	}
	// pinned_hash_match examines one thing: the first run's hash against the
	// pin.
	c.Checks["pinned_hash_match"], c.Evaluated["pinned_hash_match"] = 0, 1
	if h1 != pin {
		c.Checks["pinned_hash_match"] = 1
	}
	return c, h1, nil
}

// TestP2DeterministicReplay: the live cell folds fixtures/base/feed.jsonl
// twice, each in its own subprocess, and requires the two runs to be
// byte-identical, the first to match the pinned hash, and the feed's chain to
// verify. Four twins run the same p2Check, each over one planted defect:
//
//   - fill_price_mutated_rechained: fixtures/p2/mutated, a valid chain that is
//     not the pinned feed (pinned_hash_match);
//   - buy_and_split_reordered_rechained: fixtures/p2/reordered, the same for
//     two swapped events (pinned_hash_match);
//   - split_ratio_edited_not_rechained: fixtures/p2/tampered, one record
//     edited in place, so the chain breaks at the next seq (chain_verifies);
//   - per_process_nonce_in_snapshot: fixtures/base replayed by gates/p2nondet,
//     the production fold plus a per-process nonce in the snapshot
//     (fresh_process_identical, pinned_hash_match).
func TestP2DeterministicReplay(t *testing.T) {
	m := LoadManifest(t)
	bin := binary(t)
	base := filepath.Join(FixturesDir, "base", "feed.jsonl")
	pinRaw, err := os.ReadFile(filepath.Join(FixturesDir, "base", "snapshot.sha256"))
	if err != nil {
		t.Fatalf("pin missing: run `cd ledger && go run ./cmd/meridian snapshot --feed ../fixtures/base/feed.jsonl --out <dir>` and write the hash to fixtures/base/snapshot.sha256")
	}
	pin := strings.TrimSpace(string(pinRaw))
	end := m.Int("end_seq")

	c, h1, _ := p2Check(t, bin, base, pin)
	params := map[string]any{"pinned": pin, "viewpoint": end}
	Emit(t, Row{Prop: 2, Cell: "live", Scope: "fixtures/base folded twice in fresh processes", ContentHash: h1,
		Basis: "sha256 of canonical snapshot bytes", Rows: end, Params: params, Counts: c})

	// Twins 1 and 2: valid, re-chained feeds that are not the pinned feed.
	cm, hm, _ := p2Check(t, bin, filepath.Join(FixturesDir, "p2", "mutated", "feed.jsonl"), pin)
	Emit(t, Row{Prop: 2, Cell: "twin", Scope: "fixtures/p2/mutated (one fill price +1, re-chained) folded twice in fresh processes, vs the base pin", ContentHash: hm,
		Basis: "sha256 of canonical snapshot bytes", Rows: end, Params: map[string]any{"pinned": pin, "mutated_seq": m.Int("p2", "mutated", "seq")},
		Counts: cm, Planted: ptr(m.Planted("p2", "twin_mutated"))})
	cr, hr, _ := p2Check(t, bin, filepath.Join(FixturesDir, "p2", "reordered", "feed.jsonl"), pin)
	Emit(t, Row{Prop: 2, Cell: "twin", Scope: "fixtures/p2/reordered (a buy and the split swapped, re-chained) folded twice in fresh processes, vs the base pin", ContentHash: hr,
		Basis: "sha256 of canonical snapshot bytes", Rows: end, Params: map[string]any{"pinned": pin, "reordered_seqs": m.Ints("p2", "reordered", "seqs")},
		Counts: cr, Planted: ptr(m.Planted("p2", "twin_reordered"))})

	// Twin 3: one record edited in place, not re-chained. chain_verifies
	// counts any refusal, exactly as the live cell does; the guard keeps the
	// twin RED for its planted reason, a break at the record after the edited
	// one and nowhere else.
	tampered := filepath.Join(FixturesDir, "p2", "tampered", "feed.jsonl")
	ctp, _, err := p2Check(t, bin, tampered, pin)
	breakAt := m.Int("p2", "tampered", "break_at_seq")
	var ce *feed.ChainError
	if !errors.As(err, &ce) || ce.Seq != breakAt {
		t.Fatalf("P2 tampered twin: want a chain break at seq %d, got %v", breakAt, err)
	}
	raw, err := os.ReadFile(tampered)
	if err != nil {
		t.Fatal(err)
	}
	Emit(t, Row{Prop: 2, Cell: "twin", Scope: "fixtures/p2/tampered (split ratio edited in place, not re-chained): chain check only; the CLI refuses a feed whose chain does not verify, so no snapshot exists to replay or pin",
		ContentHash: "sha256:" + sha256Hex(bytes.ReplaceAll(raw, []byte("\r\n"), []byte("\n"))), Basis: "sha256 of the tampered feed file, CRLF normalized to LF", Rows: end,
		Params: map[string]any{"tampered_seq": m.Int("p2", "tampered", "seq"), "break_at_seq": breakAt, "observed_break_seq": ce.Seq},
		Counts: ctp, Planted: ptr(m.Planted("p2", "twin_tampered"))})

	// Twin 4: the base feed replayed by a binary whose output depends on the
	// process, not only on the feed. The row's content hash is the INPUT the
	// twin replayed -- the base feed file, hashed the way the tampered twin
	// above hashes its feed -- and not the twin's own snapshot: those bytes
	// carry a fresh nonce in every process, so a row recording them would
	// carry a different content hash on every run and pin nothing. The basis
	// says so rather than leaving a reader to infer it.
	cn, _, _ := p2Check(t, p2NondetBinary(t), base, pin)
	baseRaw, err := os.ReadFile(base)
	if err != nil {
		t.Fatal(err)
	}
	Emit(t, Row{Prop: 2, Cell: "twin", Scope: "fixtures/base folded twice in fresh processes by gates/p2nondet (the production fold and snapshot plus a crypto/rand nonce drawn in every process), vs the base pin",
		ContentHash: "sha256:" + sha256Hex(bytes.ReplaceAll(baseRaw, []byte("\r\n"), []byte("\n"))),
		Basis:       "sha256 of the base feed file, CRLF normalized to LF; the twin binary's snapshot bytes differ on every run by construction and are not recorded", Rows: end, Params: params,
		Counts: cn, Planted: ptr(m.Planted("p2", "twin_nondeterministic"))})
}

// TestReplaysIdenticalDiscriminates drives replaysIdentical directly, the same
// shape as TestSetEqualityTable in manifest_test.go, over every combination
// of equal and different hashes and bytes. The twin
// per_process_nonce_in_snapshot in TestP2DeterministicReplay shows the whole
// pipeline (two real subprocess pairs, this comparator, Emit) reporting a
// really non-deterministic replay as not identical, but only in the case
// where both legs differ; this table covers the one-leg cases that twin
// cannot reach.
//
// Neither this table nor that twin shows that the production ledger can
// produce two different fresh-process replays of one feed. The twin binary
// is not the production binary, which is what makes it a twin: the live
// cell's fresh_process_identical: 0 is credited because the same check went
// to 1 on a binary that really is non-deterministic, not because the
// production fold was broken to show it.
func TestReplaysIdenticalDiscriminates(t *testing.T) {
	cases := []struct {
		name          string
		h1            string
		b1            []byte
		h2            string
		b2            []byte
		wantIdentical bool
	}{
		{name: "same hash, same bytes — identical", h1: "sha256:aaa", b1: []byte("x"), h2: "sha256:aaa", b2: []byte("x"), wantIdentical: true},
		{name: "same hash, both empty bytes — identical", h1: "sha256:empty", b1: nil, h2: "sha256:empty", b2: nil, wantIdentical: true},
		{name: "different hash, same bytes — not identical", h1: "sha256:aaa", b1: []byte("x"), h2: "sha256:bbb", b2: []byte("x"), wantIdentical: false},
		{name: "same hash, different bytes — not identical", h1: "sha256:aaa", b1: []byte("x"), h2: "sha256:aaa", b2: []byte("y"), wantIdentical: false},
		{name: "different hash, different bytes — not identical", h1: "sha256:aaa", b1: []byte("x"), h2: "sha256:bbb", b2: []byte("y"), wantIdentical: false},
		{name: "different length bytes — not identical", h1: "sha256:aaa", b1: []byte("x"), h2: "sha256:aaa", b2: []byte("xx"), wantIdentical: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := replaysIdentical(tc.h1, tc.b1, tc.h2, tc.b2); got != tc.wantIdentical {
				t.Fatalf("replaysIdentical(%q, %v, %q, %v) = %v, want %v", tc.h1, tc.b1, tc.h2, tc.b2, got, tc.wantIdentical)
			}
		})
	}
}
