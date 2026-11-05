package main

import (
	"path/filepath"
	"strings"
	"testing"
)

// `asof --effective YYYY-MM-DD` reads (V, E): records effective after E are
// hidden before the fold, and the snapshot carries the effective key. With
// the flag omitted the output is today's, byte for byte. --seq is unchanged.
func TestAsOfEffectiveFlag(t *testing.T) {
	bin := build(t)
	feedPath := filepath.Join(t.TempDir(), "feed.jsonl")
	run(t, bin, "append", "--feed", feedPath, "--type", "fill", "--id", "e1", "--effective", "2026-01-05",
		"--payload", `{"instrument":"AAA","price":1000,"qty":100,"side":"buy","trade_id":"T-1","venue":"X"}`)
	run(t, bin, "append", "--feed", feedPath, "--type", "price", "--id", "e2", "--effective", "2026-01-12", "--payload", `{"instrument":"AAA","price":1100}`)

	plain, code := run(t, bin, "asof", "--feed", feedPath, "--seq", "2")
	if code != 0 || strings.Contains(plain, `"effective"`) {
		t.Fatalf("%d %s", code, plain)
	}
	at, code := run(t, bin, "asof", "--feed", feedPath, "--seq", "2", "--effective", "2026-01-09")
	if code != 0 {
		t.Fatalf("%d %s", code, at)
	}
	if !strings.Contains(at, `"effective":"2026-01-09"`) || !strings.Contains(at, `"feed_seq":2`) {
		t.Fatalf("missing effective key or feed_seq: %s", at)
	}
	if !strings.Contains(at, `"valuation":null`) || !strings.Contains(at, `"no_price_in_prefix"`) {
		t.Fatalf("the price effective 2026-01-12 must be hidden at E=2026-01-09: %s", at)
	}
	if !strings.Contains(plain, `"price":1100`) {
		t.Fatalf("without --effective the price is visible: %s", plain)
	}
	out, code := run(t, bin, "asof", "--feed", feedPath, "--seq", "2", "--effective", "2026-1-9")
	if code != 1 || !strings.Contains(out, "effective") {
		t.Fatalf("a malformed --effective must exit 1 and say so: %d %s", code, out)
	}
}
