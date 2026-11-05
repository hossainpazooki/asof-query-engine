package asof

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hossainpazooki/meridian/internal/canon"
	"github.com/hossainpazooki/meridian/internal/feed"
)

// effectiveFeed writes a four-record feed: a fill (2026-01-05), a split
// (2026-01-10), the split's amendment (2026-01-10), a price (2026-01-12).
func effectiveFeed(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "feed.jsonl")
	f, err := feed.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	f.Append("fill", "e1", "2026-01-05", map[string]any{"instrument": "AAA", "price": json.Number("1000"), "qty": json.Number("100"), "side": "buy", "trade_id": "T-1", "venue": "X"})
	f.Append("action", "e2", "2026-01-10", map[string]any{"action_id": "CA-1", "announced": "2026-01-08", "instrument": "AAA", "kind": "split", "processed": "2026-01-10", "ratio": json.Number("2")})
	f.Append("action_amendment", "e3", "2026-01-10", map[string]any{"action_id": "CA-1", "ratio": json.Number("3")})
	f.Append("price", "e4", "2026-01-12", map[string]any{"instrument": "AAA", "price": json.Number("1100")})
	return path
}

func TestReadAtFiltersRecordsEffectiveAfterE(t *testing.T) {
	path := effectiveFeed(t)
	full, err := Read(path, 4)
	if err != nil {
		t.Fatal(err)
	}

	// E before the split: the fill alone; the price is hidden too, so AAA is
	// unevaluable. feed_seq and the prefix hash still name V = 4.
	r, err := ReadAt(path, 4, "2026-01-09")
	if err != nil {
		t.Fatal(err)
	}
	if r.State.Positions["AAA"].Qty != 100 || len(r.State.Refusals) != 0 || len(r.State.Unevaluable) != 1 {
		t.Fatalf("at E=2026-01-09: %+v", r.State)
	}
	if r.Seq != 4 || r.Doc["feed_seq"].(json.Number) != "4" || r.PrefixHash != full.PrefixHash || r.Doc["feed_prefix_hash"] != full.PrefixHash {
		t.Fatalf("E is a filter over the prefix, not a different prefix: %+v", r)
	}
	if r.Doc["effective"] != "2026-01-09" || r.Effective != "2026-01-09" {
		t.Fatalf("snapshot must carry the effective key: %v", r.Doc["effective"])
	}
	if !strings.Contains(string(r.Bytes), `"effective":"2026-01-09"`) {
		t.Fatalf("bytes must carry the effective key: %s", r.Bytes)
	}

	// E on the split's date: split and amendment both visible (same
	// effective date), corrected terms apply; the price is still hidden.
	r, err = ReadAt(path, 4, "2026-01-10")
	if err != nil {
		t.Fatal(err)
	}
	if r.State.Positions["AAA"].Qty != 300 || len(r.State.Unevaluable) != 1 {
		t.Fatalf("at E=2026-01-10: %+v", r.State)
	}

	// E at the last record: the same fold as Read, plus exactly one key.
	r, err = ReadAt(path, 4, "2026-01-12")
	if err != nil {
		t.Fatal(err)
	}
	if r.State.Positions["AAA"].Qty != 300 || len(r.State.Unevaluable) != 0 {
		t.Fatalf("at E=2026-01-12: %+v", r.State)
	}
	if r.Hash == full.Hash {
		t.Fatal("a read with E must not share the hash of the read without it")
	}
	stripped := map[string]any{}
	for k, v := range r.Doc {
		if k != "effective" {
			stripped[k] = v
		}
	}
	a, _ := canon.Marshal(stripped)
	b, _ := canon.Marshal(full.Doc)
	if !bytes.Equal(a, b) {
		t.Fatalf("with E at the last date the documents differ beyond the effective key:\n%s\n%s", a, b)
	}

	// No E: byte for byte today's read.
	r, err = ReadAt(path, 4, "")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(r.Bytes, full.Bytes) || r.Hash != full.Hash || r.Effective != "" {
		t.Fatal("ReadAt with no date must equal Read")
	}
	if _, ok := r.Doc["effective"]; ok {
		t.Fatal("no effective key when no date is supplied")
	}

	// A malformed E is refused as such, never compared lexically.
	for _, bad := range []string{"2026-1-9", "20260109", "2026-01-32", "yesterday"} {
		if _, err := ReadAt(path, 4, bad); err == nil {
			t.Fatalf("malformed effective date %q must error", bad)
		}
	}
	if _, err := ReadAt(path, 5, "2026-01-09"); err == nil {
		t.Fatal("seq beyond feed must still error with E")
	}
}

// A16: an amendment visible at (V, E) whose action E hides is refused
// unknown_action, the fold's existing refusal for an amendment of an action
// it has not seen. The base fixture cannot show this (its amendment shares
// its action's effective date); this synthetic feed can.
func TestReadAtAmendmentWhoseActionIsHiddenIsRefusedUnknownAction(t *testing.T) {
	path := filepath.Join(t.TempDir(), "feed.jsonl")
	f, err := feed.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	f.Append("fill", "e1", "2026-01-05", map[string]any{"instrument": "AAA", "price": json.Number("1000"), "qty": json.Number("100"), "side": "buy", "trade_id": "T-1", "venue": "X"})
	f.Append("action", "e2", "2026-02-01", map[string]any{"action_id": "CA-1", "announced": "2026-01-08", "instrument": "AAA", "kind": "split", "processed": "2026-02-01", "ratio": json.Number("2")})
	f.Append("action_amendment", "e3", "2026-01-20", map[string]any{"action_id": "CA-1", "ratio": json.Number("3")})
	f.Close()

	r, err := ReadAt(path, 3, "2026-01-25")
	if err != nil {
		t.Fatal(err)
	}
	if len(r.State.Refusals) != 1 || r.State.Refusals[0].Kind != "unknown_action" || r.State.Refusals[0].EventID != "e3" {
		t.Fatalf("amendment with its action hidden must be refused unknown_action: %+v", r.State.Refusals)
	}
	if r.State.Positions["AAA"].Qty != 100 {
		t.Fatalf("the hidden split must not apply: %+v", r.State.Positions)
	}
	full, err := ReadAt(path, 3, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(full.State.Refusals) != 0 || full.State.Positions["AAA"].Qty != 300 {
		t.Fatalf("without E the amendment resolves: %+v", full.State)
	}
}

// The committed base fixture: with no date supplied every viewpoint's bytes
// are unchanged (the P1-P7 rows do not move), the full read still hashes to
// the committed pin, the universe of (viewpoint, date-or-none) pairs the
// manifest defines has 12 distinct documents (A14's size; distinct by
// construction, since feed_seq or the effective key differ), and, the one
// check that is evidence of separation, at V3 with the effective key
// stripped the three dates fold to three different states while the last
// date folds to the unfiltered read.
func TestReadAtOnBaseFixtureKeepsThePinAndSeparatesStates(t *testing.T) {
	base := filepath.Join("..", "..", "..", "fixtures", "base")
	feedPath := filepath.Join(base, "feed.jsonl")
	raw, err := os.ReadFile(filepath.Join(base, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var m struct {
		Viewpoints     map[string]int64 `json:"viewpoints"`
		EffectiveDates []string         `json:"effective_dates"`
	}
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	if len(m.Viewpoints) != 3 || len(m.EffectiveDates) != 3 {
		t.Fatalf("manifest must name three viewpoints and three effective dates: %+v", m)
	}
	pinRaw, err := os.ReadFile(filepath.Join(base, "snapshot.sha256"))
	if err != nil {
		t.Fatal(err)
	}
	pin := strings.TrimSpace(string(pinRaw))

	seen := map[string]string{}
	for name, v := range m.Viewpoints {
		plain, err := Read(feedPath, v)
		if err != nil {
			t.Fatal(err)
		}
		noDate, err := ReadAt(feedPath, v, "")
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(plain.Bytes, noDate.Bytes) {
			t.Fatalf("%s: ReadAt with no date must equal Read", name)
		}
		seen[plain.Hash] = name + "/none"
		for _, e := range m.EffectiveDates {
			r, err := ReadAt(feedPath, v, e)
			if err != nil {
				t.Fatal(err)
			}
			if r.Doc["effective"] != e {
				t.Fatalf("%s at %s: effective key %v", name, e, r.Doc["effective"])
			}
			if prev, dup := seen[r.Hash]; dup {
				t.Fatalf("pairs %s and %s/%s share a hash", prev, name, e)
			}
			seen[r.Hash] = name + "/" + e
		}
	}
	if len(seen) != 12 {
		t.Fatalf("expected 12 distinct pairs, got %d", len(seen))
	}
	stateAt := func(e string) string {
		r, err := ReadAt(feedPath, m.Viewpoints["V3"], e)
		if err != nil {
			t.Fatal(err)
		}
		delete(r.Doc, "effective")
		b, err := canon.Marshal(r.Doc)
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}
	none, s0, s1, s2 := stateAt(""), stateAt(m.EffectiveDates[0]), stateAt(m.EffectiveDates[1]), stateAt(m.EffectiveDates[2])
	if s0 == s1 || s1 == s2 || s0 == s2 {
		t.Fatal("at V3 the three effective dates must fold to three different states once the effective key is stripped")
	}
	if s2 != none {
		t.Fatal("at V3 the feed's last effective date must fold to the same state as the unfiltered read")
	}
	full, err := Read(feedPath, -1)
	if err != nil {
		t.Fatal(err)
	}
	if full.Hash != pin {
		t.Fatalf("full read hash %s differs from the committed pin %s", full.Hash, pin)
	}
}
