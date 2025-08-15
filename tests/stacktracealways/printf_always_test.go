//go:build dlg

package dlg_test_stacktrace_always

import (
	"bytes"
	"testing"

	"github.com/vvvvv/dlg"
	"github.com/vvvvv/dlg/tests/internal"
)

func TestPrintfStackTraceAlways(t *testing.T) {
	out := internal.CaptureOutput(func() {
		dlg.Printf("test message")
	})

	lines := internal.ParseLines([]byte(out))

	if len(lines) > 1 {
		// This should only happen if there's something wrong with internal.ParseLines
		t.Fatalf("Too many lines: %+v", lines)
	}

	got := lines[0]

	want := struct {
		line  string
		trace bool
	}{
		"test message", true,
	}

	if got.Line() != want.line || got.HasTrace() != want.trace {
		t.Errorf("Mismatch: want: %q (stacktrace: %v) ; got: %q (stacktrace: %v)", want.line, want.trace, got.Line(), got.HasTrace())
	}
}

func BenchmarkPrintfTraceAlways16(b *testing.B) {
	var buf bytes.Buffer
	dlg.SetOutput(&buf)

	s := internal.RandomStrings(16)

	for i := 0; i < b.N; i++ {
		buf.Reset()
		dlg.Printf(s[i%len(s)])
	}
}

func BenchmarkPrintfTraceAlways64(b *testing.B) {
	var buf bytes.Buffer
	dlg.SetOutput(&buf)

	s := internal.RandomStrings(64)

	for i := 0; i < b.N; i++ {
		buf.Reset()
		dlg.Printf(s[i%len(s)])
	}
}

func BenchmarkPrintfTraceAlways128(b *testing.B) {
	var buf bytes.Buffer
	dlg.SetOutput(&buf)

	s := internal.RandomStrings(128)

	for i := 0; i < b.N; i++ {
		buf.Reset()
		dlg.Printf(s[i%len(s)])
	}
}

func BenchmarkPrintfWithFormattingTraceAlways16(b *testing.B) {
	var buf bytes.Buffer
	dlg.SetOutput(&buf)

	s := internal.RandomStrings(16)

	for i := 0; i < b.N; i++ {
		buf.Reset()
		dlg.Printf(s[i%len(s)], i)
	}
}

func BenchmarkPrintfWithFormattingTraceAlways64(b *testing.B) {
	var buf bytes.Buffer
	dlg.SetOutput(&buf)

	s := internal.RandomStrings(64)

	for i := 0; i < b.N; i++ {
		buf.Reset()
		dlg.Printf(s[i%len(s)], i)
	}
}

func BenchmarkPrintfWithFormattingTraceAlways128(b *testing.B) {
	var buf bytes.Buffer
	dlg.SetOutput(&buf)

	s := internal.RandomStrings(128)

	for i := 0; i < b.N; i++ {
		buf.Reset()
		dlg.Printf(s[i%len(s)], i)
	}
}
