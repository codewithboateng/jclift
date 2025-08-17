package fuzz

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/codewithboateng/jclift/internal/parser"
)

// Fuzz the parser with arbitrary content to ensure we never panic.
// We wrap the data in a tiny JCL scaffold to pass basic shape checks.
func FuzzParseNoPanic(f *testing.F) {
	seeds := [][]byte{
		[]byte("//A JOB\n//S EXEC PGM=IEFBR14\n"),
		[]byte("//PAY JOB\n//* COMMENT\n"),
		[]byte("garbage-but-should-not-panic\n"),
	}
	for _, s := range seeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		dir := t.TempDir()
		content := append([]byte("//FZ JOB (1),'F',CLASS=A,MSGCLASS=X\n"), data...)
		if err := os.WriteFile(filepath.Join(dir, "fuzz.jcl"), content, 0o644); err != nil {
			t.Skipf("write failed: %v", err)
		}
		_, _ = parser.Parse(dir) // we only assert "no panic"
	})
}
