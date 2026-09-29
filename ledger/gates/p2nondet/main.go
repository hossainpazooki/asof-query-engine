// Command p2nondet is P2's non-deterministic replay twin. ledger/gates/p2_test.go
// builds it into a temp directory and runs the live cell's own checks with
// it; nothing else builds or ships it, and nothing in cmd/ or internal/
// changes for it, so the production binary the live cell runs never contains
// this code.
//
// It is the production fold and snapshot (internal/asof) with one planted
// defect: every process stamps the snapshot document with a fresh 128-bit
// nonce from crypto/rand before the canonical bytes are marshaled, so the
// bytes depend on something no feed record supplies. Two fresh-process
// replays of the same feed therefore differ unless two independent 128-bit
// draws collide (probability 2^-128 for each of the two legs the gate
// compares, hash and bytes).
//
// It answers only the two subcommands the P2 gate calls, in the production
// CLI's output shapes: `asof --feed F` writes the snapshot bytes to stdout;
// `snapshot --feed F --out D` writes <D>/<hex>.json and prints
// "sha256:<hex> <path>". A feed whose chain does not verify exits 2, as the
// production CLI does.
package main

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/hossainpazooki/meridian/internal/asof"
	"github.com/hossainpazooki/meridian/internal/canon"
	"github.com/hossainpazooki/meridian/internal/feed"
	"github.com/hossainpazooki/meridian/internal/snapshot"
)

func main() { os.Exit(run(os.Args[1:])) }

func fail(err error) int {
	fmt.Fprintln(os.Stderr, "error:", err)
	var ce *feed.ChainError
	if errors.As(err, &ce) {
		return 2
	}
	return 1
}

func run(args []string) int {
	if len(args) == 0 || (args[0] != "asof" && args[0] != "snapshot") {
		fmt.Fprintln(os.Stderr, "usage: p2nondet <asof|snapshot> --feed F [--out D]")
		return 1
	}
	fs := flag.NewFlagSet(args[0], flag.ContinueOnError)
	feedPath := fs.String("feed", "", "feed path")
	out := fs.String("out", "", "output directory (snapshot only)")
	if err := fs.Parse(args[1:]); err != nil {
		return 1
	}
	// feed.Open creates a missing file; a read-only command must not.
	if _, err := os.Stat(*feedPath); err != nil {
		return fail(err)
	}
	r, err := asof.Read(*feedPath, -1)
	if err != nil {
		return fail(err)
	}
	// The planted defect: a per-process value in the document.
	var nonce [16]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return fail(err)
	}
	r.Doc["replay_nonce"] = hex.EncodeToString(nonce[:])
	b, err := canon.Marshal(r.Doc)
	if err != nil {
		return fail(err)
	}
	b = append(b, '\n')
	h := "sha256:" + canon.SHA256Hex(b)
	if args[0] == "asof" {
		os.Stdout.Write(b)
		return 0
	}
	p, err := snapshot.Write(*out, b, h)
	if err != nil {
		return fail(err)
	}
	fmt.Printf("%s %s\n", h, p)
	return 0
}
