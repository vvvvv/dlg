//go:build dlg

package stacktraceregiononerror_test

import (
	"fmt"
	"testing"

	"github.com/vvvvv/dlg"
	"github.com/vvvvv/dlg/tests/internal"
)

func noTrace() {
	dlg.StartTrace()
	defer dlg.StopTrace()

	dlg.Printf("don't trace this")
}

func traceOnError() {
	dlg.StartTrace()
	err := fmt.Errorf("this error")

	dlg.Printf("don't trace this")
	dlg.Printf("trace %v", err)
	dlg.StopTrace()

	dlg.Printf("don't trace this")
}

func TestPrintfStackTraceRegion(t *testing.T) {
	type exp struct {
		line  string
		trace bool
	}

	tcs := []struct {
		name string
		fn   func()
		exp  []exp
	}{
		{
			name: "don't trace if no error is given",
			fn:   noTrace,
			exp: []exp{
				{
					"don't trace this", false,
				},
			},
		},
		{
			name: "trace on error",
			fn:   traceOnError,
			exp: []exp{
				{
					"don't trace this", false,
				},
				{
					"trace this error", true,
				},
				{
					"don't trace this", false,
				},
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			out := internal.CaptureOutput(tc.fn)
			lines := internal.ParseLines([]byte(out))

			if len(lines) != len(tc.exp) {
				t.Fatalf("Testcase must contain all output; expected: %v ; got: %v", len(lines), len(tc.exp))
			}

			for i := 0; i < len(tc.exp); i++ {
				want := tc.exp[i]
				got := lines[i]

				if want.line != got.Line() || want.trace != got.HasTrace() {
					t.Errorf("Mismatch: want: %q (stacktrace: %v) ; got: %q (stacktrace: %v)\n%+v", want.line, want.trace, got.Line(), got.HasTrace(), got)
				}
			}
		})
	}
}
