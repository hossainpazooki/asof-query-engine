package fold

import (
	"testing"

	"github.com/hossainpazooki/meridian/internal/feed"
)

// Step 2 (the effective-date axis): asof.ReadAt hands Fold the records of
// [1..V] with those effective after E removed, and upTo still names V. Such
// a slice is shorter than V. Fold's contract is stated over Seq (see
// TestFoldSelectsPrefixBySeqNotIndex), so the slice's length is not a bound
// on upTo; the only out-of-range upTo is a negative one. Before step 2 the
// check `upTo > len(records)` refused exactly this slice.
func TestFoldAcceptsUpToBeyondSliceLength(t *testing.T) {
	rs := records(
		fill("2026-01-05", "AAA", "buy", 10, 100, "T-1"),
		fill("2026-01-06", "AAA", "buy", 5, 100, "T-2"),
		fill("2026-01-07", "AAA", "buy", 1, 100, "T-3"),
	)
	filtered := []feed.Record{rs[0], rs[2]} // seq 1 and seq 3; seq 2 is hidden
	s, err := Fold(filtered, 3)
	if err != nil {
		t.Fatalf("Fold over a 2-record slice with upTo 3 must fold by Seq, got: %v", err)
	}
	if s.Seq != 3 {
		t.Fatalf("State.Seq must be upTo (3), got %d", s.Seq)
	}
	if s.Positions["AAA"].Qty != 11 {
		t.Fatalf("seq 1 and seq 3 fold to qty 11, got %d", s.Positions["AAA"].Qty)
	}
	if _, err := Fold(filtered, -1); err == nil {
		t.Fatal("negative upTo must still error")
	}
}
