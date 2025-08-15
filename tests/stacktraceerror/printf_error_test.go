//go:build dlg

package dlg_test_stacktrace_error

import (
	"errors"
	"testing"

	"github.com/vvvvv/dlg"
	"github.com/vvvvv/dlg/tests/internal"
)

func TestPrintfStackTraceOnError(t *testing.T) {
	out := internal.CaptureOutput(func() {
		dlg.Printf("message with error: %v", errors.New("some error"))
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
		"message with error: some error", true,
	}

	if got.Line() != want.line || got.HasTrace() != want.trace {
		t.Errorf("Mismatch: want: %q (stacktrace: %v) ; got: %q (stacktrace: %v)", want.line, want.trace, got.Line(), got.HasTrace())
	}
}
