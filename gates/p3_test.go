package gates

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/hossainpazooki/meridian/internal/snapshot"
)

// threeHistoriesViolation reports 1 when q1, q2, q3 (the AAA quantity at
// V1, V2, V3, as their canonical decimal strings) do not represent three
// DISTINCT values, 0 otherwise — 1 for any collapse, of two viewpoints or
// of all three. Factored out of p3Check so its arithmetic can be exercised
// directly by TestP3ThreeHistoriesDiscriminates, table-style; the
// viewpoint-ignored twin in TestP3PointInTimeActions drives it through the
// document path. fixtures/generate.py plants the same predicate
// (three_histories_violation).
func threeHistoriesViolation(q1, q2, q3 string) int64 {
	qtys := map[string]bool{q1: true, q2: true, q3: true}
	if len(qtys) != 3 {
		return 1
	}
	return 0
}

// p3Check evaluates point-in-time corporate-action folding at three
// viewpoints (V1: before the action, V2: after the original terms but
// before the amendment, V3: after the amendment). docs is keyed by
// viewpoint name and may carry a twin's document at any viewpoint (each
// twin replaces the viewpoints it names; the ledger's own documents fill
// the rest); expected holds the honest golden for each viewpoint. t is
// used only to fail loudly, with a message naming what was missing, if a
// document is malformed — never to panic on a bad type assertion.
func p3Check(t *testing.T, m Manifest, docs map[string]snapshot.Doc, expected map[string]snapshot.Doc) Counts {
	t.Helper()
	c := NewCounts("viewpoint_V1", "viewpoint_V2", "viewpoint_V3", "three_histories",
		"positions_match_manifest", "unevaluable_match_manifest")
	qtyAt := map[string]string{}
	for _, v := range []string{"V1", "V2", "V3"} {
		ms := snapshot.Diff(expected[v], docs[v])
		c.Evaluated["viewpoint_"+v] = int64(snapshot.Leaves(expected[v]))
		c.Checks["viewpoint_"+v] = int64(len(ms))

		// A gate must refuse a malformed document loudly and legibly, not
		// die in an unchecked type assertion — the same fail-closed
		// discipline the fold and reconcile apply. Every step below uses
		// the comma-ok form and names exactly what was missing or
		// mis-shaped.
		pos, ok := docs[v]["positions"].(map[string]any)
		if !ok {
			t.Fatalf("p3: viewpoint %s document has no \"positions\" object", v)
		}
		aaaAny, ok := pos["AAA"]
		if !ok {
			t.Fatalf("p3: viewpoint %s document has no position for instrument AAA", v)
		}
		aaa, ok := aaaAny.(map[string]any)
		if !ok {
			t.Fatalf("p3: viewpoint %s AAA position is not an object (got %T)", v, aaaAny)
		}
		qtyAny, ok := aaa["qty"]
		if !ok {
			t.Fatalf("p3: viewpoint %s AAA position has no \"qty\" field", v)
		}
		qty, ok := qtyAny.(json.Number)
		if !ok {
			t.Fatalf("p3: viewpoint %s AAA qty is not a json.Number (got %T)", v, qtyAny)
		}
		qtyAt[v] = string(qty)

		// snapshot.Diff only walks the golden's keys, so it cannot by
		// itself catch a doc that invents a position or drops an
		// instrument from "unevaluable" — close that hole with its own
		// set-equality check against the manifest's published lists,
		// keyed by viewpoint (V1/V2/V3), and SUMMED ACROSS ALL THREE
		// under a single check name since that is how the planted
		// expected_violations are published (one count each, not
		// per-viewpoint). Read "positions_match_manifest: 0" and
		// "unevaluable_match_manifest: 0" as "zero mismatches total
		// across V1+V2+V3," not as a single comparison — a non-zero
		// value says one of the three viewpoints diverged, but not by
		// itself which one (viewpoint_V1/V2/V3 pinpoint that). SetEquality
		// overwrites rather than adds, so accumulate through a scratch
		// Counts per viewpoint.
		//
		// positions_match_manifest asserts over the position sets
		// themselves (got and want ARE the universe), so it uses plain
		// SetEquality. unevaluable_match_manifest asserts, for every
		// position in the ledger at that viewpoint, whether its
		// unevaluable-or-not status matches the golden — its universe is
		// the position set, not the union of two unevaluable lists — so
		// it must use SetEqualityOverUniverse with that viewpoint's
		// published positions as the universe (harness rule documented
		// on SetEqualityOverUniverse in gates/manifest.go, kept
		// consistent with the other property gates).
		tmp := NewCounts("positions_match_manifest", "unevaluable_match_manifest")
		SetEquality(tmp, "positions_match_manifest", PositionKeys(docs[v]), m.Strs("positions_at", v))
		SetEqualityOverUniverse(tmp, "unevaluable_match_manifest", UnevaluableInstruments(docs[v]), m.Strs("unevaluable_at", v), m.Strs("positions_at", v))
		c.Evaluated["positions_match_manifest"] += tmp.Evaluated["positions_match_manifest"]
		c.Checks["positions_match_manifest"] += tmp.Checks["positions_match_manifest"]
		c.Evaluated["unevaluable_match_manifest"] += tmp.Evaluated["unevaluable_match_manifest"]
		c.Checks["unevaluable_match_manifest"] += tmp.Checks["unevaluable_match_manifest"]
	}
	// Asserts the three viewpoints really do yield three DISTINCT AAA
	// quantities, not merely that each differs from its own golden.
	//
	// It cannot move alone. The three goldens carry three distinct AAA
	// quantities (fixtures/generate.py dies otherwise), so any document set
	// that collapses two histories leaves at least one viewpoint off its
	// golden and moves viewpoint_V1/V2/V3 as well. The viewpoint-ignored
	// twin (every viewpoint answered with the end-of-feed snapshot) is the
	// collapse this check exists to report: it reads 1 there, and 0 on the
	// live row and on every other twin. TestP3ThreeHistoriesDiscriminates
	// covers the two-viewpoint collapses no twin plants.
	c.Evaluated["three_histories"] = 1
	c.Checks["three_histories"] = threeHistoriesViolation(qtyAt["V1"], qtyAt["V2"], qtyAt["V3"])
	return c
}

// TestP3ThreeHistoriesDiscriminates proves threeHistoriesViolation's
// arithmetic directly — the same shape as TestSetEqualityTable in
// gates/manifest.go. The viewpoint-ignored twin in TestP3PointInTimeActions
// plants one collapse (all three viewpoints, through real documents); this
// table also covers each two-viewpoint collapse, which no twin plants.
func TestP3ThreeHistoriesDiscriminates(t *testing.T) {
	cases := []struct {
		name           string
		q1, q2, q3     string
		wantViolations int64
	}{
		{"three distinct values (the live shape)", "26", "61", "86", 0},
		{"V1 and V2 collapse", "26", "26", "86", 1},
		{"V1 and V3 collapse", "26", "61", "26", 1},
		{"V2 and V3 collapse", "26", "86", "86", 1},
		{"all three collapse", "50", "50", "50", 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := threeHistoriesViolation(tc.q1, tc.q2, tc.q3); got != tc.wantViolations {
				t.Fatalf("threeHistoriesViolation(%q, %q, %q) = %d, want %d", tc.q1, tc.q2, tc.q3, got, tc.wantViolations)
			}
		})
	}
}

func TestP3PointInTimeActions(t *testing.T) {
	m := LoadManifest(t)
	vp := map[string]int64{"V1": m.Int("viewpoints", "V1"), "V2": m.Int("viewpoints", "V2"), "V3": m.Int("viewpoints", "V3")}
	expected, docs := map[string]snapshot.Doc{}, map[string]snapshot.Doc{}
	var hashV3 string
	for v, seq := range vp {
		expected[v] = LoadDoc(t, filepath.Join(FixturesDir, "base", "expected", v+".json"))
		r := ReadFixture(t, "base/feed.jsonl", seq)
		docs[v] = r.Doc
		if v == "V3" {
			hashV3 = r.Hash
		}
	}
	params := map[string]any{"viewpoints": vp, "action_seq": m.Int("action", "seq"), "amendment_seq": m.Int("action", "amendment_seq"),
		"original_ratio": m.Int("action", "original_ratio"), "amended_ratio": m.Int("action", "amended_ratio")}

	c := p3Check(t, m, docs, expected)
	Emit(t, Row{Prop: 3, Cell: "live", Scope: "fixtures/base at three viewpoints around the amended split", ContentHash: hashV3,
		Basis: "sha256 of canonical snapshot bytes at V3", Rows: 3, Params: params, Counts: c})

	leaked := map[string]snapshot.Doc{"V1": docs["V1"], "V2": LoadDoc(t, filepath.Join(FixturesDir, "p3", "twin", "V2.json")), "V3": docs["V3"]}
	ct := p3Check(t, m, leaked, expected)
	raw, _ := json.Marshal(leaked["V2"])
	Emit(t, Row{Prop: 3, Cell: "twin", Scope: "V2 replaced by a snapshot that leaked the amended terms", ContentHash: "sha256:" + sha256Hex(raw),
		Basis: "sha256 of the leaked V2 document as decoded and re-marshaled", Rows: 3, Params: params, Counts: ct, Planted: ptr(m.Planted("p3"))})

	// Twins beyond the leak, one per viewpoint check the leak cannot move.
	// Each substitutes the generator's documents at the viewpoints it names
	// and keeps the ledger's own at the rest; the planted counts are the
	// manifest's p3.<key>.
	p3Twin(t, m, docs, expected, params, "twin_effective_date", "twin-effective-date", []string{"V1", "V2"},
		"V1 and V2 replaced by snapshots that admitted corporate actions by effective date, not by feed position")
	p3Twin(t, m, docs, expected, params, "twin_stale_terms", "twin-stale-terms", []string{"V3"},
		"V3 replaced by a snapshot that never applied the amendment")
	p3Twin(t, m, docs, expected, params, "twin_viewpoint_ignored", "twin-viewpoint-ignored", []string{"V1", "V2"},
		"V1 and V2 replaced by the end-of-feed snapshot: the requested viewpoint ignored")
}

// p3Twin emits one P3 twin row: the generator's documents under
// fixtures/p3/<dir>/<V>.json replace the ledger's own at each viewpoint in
// subst, the ledger's documents (docs) fill the rest, and the row is judged
// against the manifest's p3.<key> plant.
func p3Twin(t *testing.T, m Manifest, docs, expected map[string]snapshot.Doc, params map[string]any, key, dir string, subst []string, scope string) {
	t.Helper()
	twin := map[string]snapshot.Doc{"V1": docs["V1"], "V2": docs["V2"], "V3": docs["V3"]}
	replaced := map[string]snapshot.Doc{}
	for _, v := range subst {
		d := LoadDoc(t, filepath.Join(FixturesDir, "p3", dir, v+".json"))
		twin[v], replaced[v] = d, d
	}
	c := p3Check(t, m, twin, expected)
	raw, err := json.Marshal(replaced)
	if err != nil {
		t.Fatalf("p3 %s: substituted documents do not marshal: %v", key, err)
	}
	Emit(t, Row{Prop: 3, Cell: "twin", Scope: scope, ContentHash: "sha256:" + sha256Hex(raw),
		Basis: "sha256 of the substituted documents, keyed by viewpoint, as decoded and re-marshaled", Rows: 3, Params: params, Counts: c,
		Planted: ptr(m.Planted("p3", key))})
}
