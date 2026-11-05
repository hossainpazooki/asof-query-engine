// Package asof answers "what did the ledger know at seq V" by replaying the
// prefix [1..V] through the fold on every call, and "what was in force on
// date E" by removing the records effective after E before that replay
// (ReadAt). Nothing is cached; nothing derived exists outside this call.
package asof

import (
	"fmt"
	"time"

	"github.com/hossainpazooki/meridian/internal/feed"
	"github.com/hossainpazooki/meridian/internal/fold"
	"github.com/hossainpazooki/meridian/internal/snapshot"
)

type Result struct {
	State            fold.State
	Doc              snapshot.Doc
	Bytes            []byte
	Hash, PrefixHash string
	Seq              int64
	// Effective is the date the read filtered by; "" when none was supplied.
	Effective string
}

// Read opens feedPath (verifying its chain), folds [1..seq] and builds the
// snapshot. seq < 0 selects the last record.
func Read(feedPath string, seq int64) (Result, error) { return ReadAt(feedPath, seq, "") }

// ReadAt is Read at (seq, effective): records of [1..seq] whose effective
// date is after effective (YYYY-MM-DD) are removed before the fold. The
// prefix hash and feed_seq still name seq, read from the unfiltered feed;
// the snapshot gains one key, "effective". effective == "" is Read, byte
// for byte.
func ReadAt(feedPath string, seq int64, effective string) (Result, error) {
	f, err := feed.Open(feedPath)
	if err != nil {
		return Result{}, err
	}
	defer f.Close()
	return ReadFromAt(f, seq, effective)
}

// ReadFrom folds [1..seq] of an already-open feed and builds the snapshot.
// seq < 0 selects the last record. It exists so a caller that has already
// opened the feed (to check its length, say) does not pay a second scan.
func ReadFrom(f *feed.Feed, seq int64) (Result, error) { return ReadFromAt(f, seq, "") }

// ReadFromAt is ReadAt over an already-open feed.
func ReadFromAt(f *feed.Feed, seq int64, effective string) (Result, error) {
	if effective != "" {
		if _, err := time.Parse("2006-01-02", effective); err != nil {
			return Result{}, fmt.Errorf("asof: effective date %q is not YYYY-MM-DD", effective)
		}
	}
	if seq < 0 {
		seq = f.Len()
	}
	ph, err := f.PrefixHash(seq)
	if err != nil {
		return Result{}, err
	}
	records := f.Records()
	if effective != "" {
		// The filter is over the prefix: Seq <= seq keeps the fold's own
		// selection, and the date compare is lexical, which orders
		// YYYY-MM-DD correctly. The feed checks only that a record's
		// effective is present, so a record written with another shape
		// compares as text; the ledger CLI writes this shape.
		kept := make([]feed.Record, 0, len(records))
		for _, r := range records {
			if r.Seq <= seq && r.Effective <= effective {
				kept = append(kept, r)
			}
		}
		records = kept
	}
	st, err := fold.Fold(records, seq)
	if err != nil {
		return Result{}, err
	}
	doc, b, h, err := snapshot.BuildAt(st, ph, effective)
	if err != nil {
		return Result{}, err
	}
	return Result{State: st, Doc: doc, Bytes: b, Hash: h, PrefixHash: ph, Seq: seq, Effective: effective}, nil
}
