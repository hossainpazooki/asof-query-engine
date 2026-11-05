package snapshot

import (
	"bytes"
	"strings"
	"testing"

	"github.com/hossainpazooki/meridian/internal/fold"
)

func emptyState() fold.State {
	return fold.State{Seq: 1, Positions: map[string]fold.Position{}, Valuations: map[string]fold.Valuation{},
		Unevaluable: []fold.Unevaluable{}, Absorbed: []fold.Absorbed{}, Refusals: []fold.Refusal{}}
}

// A read with E supplied adds exactly one canonical key, "effective", sorted
// into place; with E omitted the bytes are what Build produces today.
func TestBuildAtAddsExactlyTheEffectiveKey(t *testing.T) {
	s := emptyState()
	_, plain, plainHash, err := Build(s, "sha256:abc")
	if err != nil {
		t.Fatal(err)
	}
	_, same, sameHash, err := BuildAt(s, "sha256:abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(plain, same) || plainHash != sameHash {
		t.Fatal("BuildAt with no date must equal Build byte for byte")
	}
	doc, b, h, err := BuildAt(s, "sha256:abc", "2026-01-09")
	if err != nil {
		t.Fatal(err)
	}
	if doc["effective"] != "2026-01-09" {
		t.Fatalf("doc effective = %v", doc["effective"])
	}
	want := `"dividend_income":0,"effective":"2026-01-09","feed_prefix_hash":"sha256:abc"`
	if !strings.Contains(string(b), want) {
		t.Fatalf("effective must sort between dividend_income and feed_prefix_hash:\n%s", b)
	}
	if h == plainHash {
		t.Fatal("adding the effective key must change the hash")
	}
	// One key, nothing else: removing it restores the plain bytes.
	restored := strings.Replace(string(b), `"effective":"2026-01-09",`, "", 1)
	if restored != string(plain) {
		t.Fatalf("BuildAt changed more than the effective key:\n%s\n%s", restored, plain)
	}
}
