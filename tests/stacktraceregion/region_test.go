//go:build dlg

package stacktraceregion_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/vvvvv/dlg"
	"github.com/vvvvv/dlg/tests/internal"
)

func noTrace() {
	dlg.Printf("no trace")
}

func traceRegion() {
	dlg.StartTrace()
	defer dlg.StopTrace()

	dlg.Printf("trace this")
}

func traceRegionUntilStop() {
	dlg.StartTrace()

	dlg.Printf("trace this")
	dlg.StopTrace()

	dlg.Printf("don't trace this")
}

func traceClosure() {
	dlg.StartTrace()

	dlg.Printf("trace this")

	f := func() {
		dlg.Printf("trace closure 1")
	}

	g := func() {
		dlg.Printf("don't trace closure 2")
	}

	h := func() {
		dlg.StartTrace()
		defer dlg.StopTrace()

		dlg.Printf("trace closure 2")
	}

	dlg.Printf("trace this too")

	f() // trace this
	dlg.StopTrace()
	g() // don't trace this
	h() // trace this

	dlg.Printf("don't trace this")
}

func traceDeepStack() {
	f := func() {
		dlg.StartTrace()
		defer dlg.StopTrace()
		func() {
			func() {
				func() {
					dlg.Printf("trace deep stack call")
				}()
			}()
		}()
	}

	f()
}

func stopTraceOnlyAtScope() {
	dlg.StartTrace()

	f := func() {
		dlg.Printf("trace closure")
		dlg.StopTrace()
	}

	f()
	dlg.Printf("trace this")

	dlg.StopTrace()

	dlg.Printf("don't trace this")
}

func stopTraceIgnoreUnnecessaryCalls() {
	dlg.StopTrace()
	dlg.StopTrace()
	dlg.StartTrace()
	dlg.Printf("trace this")
	dlg.StopTrace()
	dlg.Printf("don't trace this")
}

func stopTraceOnKeySimple() {
	dlg.StartTrace(1)

	f := func() {
		dlg.Printf("trace this")
		dlg.StopTrace(1)
	}

	f()

	dlg.Printf("don't trace this")
}

func stopTraceOnKeyStruct() {
	dlg.StartTrace(struct{}{})

	dlg.Printf("trace this")

	dlg.StopTrace("invalid key")

	dlg.Printf("trace this too")

	dlg.StopTrace(struct{}{})

	dlg.Printf("don't trace this")
}

func stopTraceOnlyOnValidKey() {
	dlg.StartTrace("foo")

	dlg.Printf("trace this")

	dlg.StopTrace("invalid key")
	dlg.Printf("trace this too")

	dlg.StopTrace("foo")
	dlg.Printf("don't trace this")
}

func stopTraceNoKeyStopsActiveRegion() {
	dlg.StartTrace("abc")

	dlg.Printf("trace this")

	dlg.StopTrace()

	dlg.Printf("don't trace this")
}

func stopTraceNoKeyStopsOnlyScopeRegion() {
	dlg.StartTrace("abc")

	dlg.Printf("trace this")

	f := func() {
		dlg.StopTrace()

		dlg.Printf("trace this too")
	}

	f()

	dlg.Printf("trace this aswell")

	dlg.StopTrace()
	dlg.Printf("don't trace this")
}

func startTracingRegionOrPrintf(start bool, key any) {
	if start {
		dlg.StartTrace(key)
		return
	}
	dlg.Printf("start region or printf")
}

func tracingRegionAcrossGoroutines() {
	key := "key"
	started := make(chan struct{})
	go func() {
		// start a region inside go routine
		startTracingRegionOrPrintf(true, key)
		close(started)
	}()

	<-started // region is active

	dlg.Printf("don't trace this")

	startTracingRegionOrPrintf(false, nil)

	dlg.StopTrace(key)

	startTracingRegionOrPrintf(false, nil)
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
			name: "don't trace if no region is active",
			fn:   noTrace,
			exp: []exp{
				{
					"no trace", false,
				},
			},
		},
		{
			name: "trace in region",
			fn:   traceRegion,
			exp: []exp{
				{
					"trace this", true,
				},
			},
		},
		{
			name: "trace in region until stop",
			fn:   traceRegionUntilStop,
			exp: []exp{
				{
					"trace this", true,
				},
				{
					"don't trace this", false,
				},
			},
		},
		{
			name: "trace closure",
			fn:   traceClosure,
			exp: []exp{
				{
					"trace this", true,
				},
				{
					"trace this too", true,
				},
				{
					"trace closure 1", true,
				},
				{
					"don't trace closure 2", false,
				},
				{
					"trace closure 2", true,
				},
				{
					"don't trace this", false,
				},
			},
		},
		{
			name: "trace deep stack",
			fn:   traceDeepStack,
			exp: []exp{
				{
					"trace deep stack call", true,
				},
			},
		},
		{
			name: "ignore unnecessary calls to StopTrace",
			fn:   stopTraceIgnoreUnnecessaryCalls,
			exp: []exp{
				{
					"trace this", true,
				},
				{
					"don't trace this", false,
				},
			},
		},
		{
			name: "stop tracing if StopTrace was called from the same scope",
			fn:   stopTraceOnlyAtScope,
			exp: []exp{
				{
					"trace closure", true,
				},
				{
					"trace this", true,
				},
				{
					"don't trace this", false,
				},
			},
		},
		{
			name: "stop tracing if called with correct key regardless of scope",
			fn:   stopTraceOnKeySimple,
			exp: []exp{
				{
					"trace this", true,
				},
				{
					"don't trace this", false,
				},
			},
		},
		{
			name: "accept any key to start and stop a region",
			fn:   stopTraceOnKeyStruct,
			exp: []exp{
				{
					"trace this", true,
				},
				{
					"trace this too", true,
				},
				{
					"don't trace this", false,
				},
			},
		},
		{
			name: "stop tracing only on valid key",
			fn:   stopTraceOnlyOnValidKey,
			exp: []exp{
				{
					"trace this", true,
				},
				{
					"trace this too", true,
				},
				{
					"don't trace this", false,
				},
			},
		},
		{
			name: "stop tracing active region with key even if no key is given",
			fn:   stopTraceNoKeyStopsActiveRegion,
			exp: []exp{
				{
					"trace this", true,
				},
				{
					"don't trace this", false,
				},
			},
		},
		{
			name: "stop tracing active region with key only when called from the same scope",
			fn:   stopTraceNoKeyStopsOnlyScopeRegion,
			exp: []exp{
				{
					"trace this", true,
				},
				{
					"trace this too", true,
				},
				{
					"trace this aswell", true,
				},
				{
					"don't trace this", false,
				},
			},
		},
		{
			name: "tracing region stays active across go routines",
			fn:   tracingRegionAcrossGoroutines,
			exp: []exp{
				{"don't trace this", false},
				{"start region or printf", true},
				{"start region or printf", false},
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			out := internal.CaptureOutput(tc.fn)
			lines := internal.ParseLines([]byte(out))

			if len(lines) != len(tc.exp) {
				for _, l := range lines {
					// 	// this doesn't find the very last line if it isn't a stacktrace i believe.
					// 	// regex issue
					fmt.Printf("line: %v \n", l.Line())
				}
				fmt.Printf("OUT: %v\n", out)
				t.Fatalf("Testcase must contain all output; expected: %v ; got: %v", len(lines), len(tc.exp))
			}

			for i := 0; i < len(tc.exp); i++ {
				want := tc.exp[i]
				got := lines[i]

				// t.Errorf("got: %#v", got)
				if want.line != got.Line() || want.trace != got.HasTrace() {
					t.Errorf("Mismatch: want: %q (stacktrace: %v) ; got: %q (stacktrace: %v)", want.line, want.trace, got.Line(), got.HasTrace())
				}
			}
		})
	}
}

func BenchmarkPrintfWithRegion16(b *testing.B) {
	var buf bytes.Buffer
	dlg.SetOutput(&buf)

	s := internal.RandomStrings(8)

	dlg.StartTrace()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		dlg.Printf(s[i%len(s)])
	}
	dlg.StopTrace()
}

func BenchmarkPrintfWithRegion64(b *testing.B) {
	var buf bytes.Buffer
	dlg.SetOutput(&buf)

	s := internal.RandomStrings(8)

	dlg.StartTrace()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		dlg.Printf(s[i%len(s)])
	}
	dlg.StopTrace()
}

func BenchmarkPrintfWithRegionStartStop16(b *testing.B) {
	var buf bytes.Buffer
	dlg.SetOutput(&buf)

	s := internal.RandomStrings(16)

	for i := 0; i < b.N; i++ {
		buf.Reset()
		dlg.StartTrace()
		dlg.Printf(s[i%len(s)])
		dlg.StopTrace()
	}
}

func BenchmarkPrintfWithRegionStartStop64(b *testing.B) {
	var buf bytes.Buffer
	dlg.SetOutput(&buf)

	s := internal.RandomStrings(64)

	for i := 0; i < b.N; i++ {
		buf.Reset()
		dlg.StartTrace()
		dlg.Printf(s[i%len(s)])
		dlg.StopTrace()
	}
}

func BenchmarkPrintfWithRegionStartStopWithKey16(b *testing.B) {
	var buf bytes.Buffer
	dlg.SetOutput(&buf)

	s := internal.RandomStrings(16)
	key := "key"

	for i := 0; i < b.N; i++ {
		buf.Reset()
		dlg.StartTrace(key)
		dlg.Printf(s[i%len(s)])
		dlg.StopTrace(key)
	}
}

func BenchmarkPrintfWithRegionStartStopWithKey64(b *testing.B) {
	var buf bytes.Buffer
	dlg.SetOutput(&buf)

	s := internal.RandomStrings(64)
	key := "key"

	for i := 0; i < b.N; i++ {
		buf.Reset()
		dlg.StartTrace(key)
		dlg.Printf(s[i%len(s)])
		dlg.StopTrace(key)
	}
}
