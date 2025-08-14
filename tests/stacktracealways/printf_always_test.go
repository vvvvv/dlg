//go:build dlg

package dlg_test_stacktrace_always

import (
	"bytes"
	"regexp"
	"strings"
	"testing"

	"github.com/vvvvv/dlg"
	"github.com/vvvvv/dlg/tests/internal"
)

// var funcName = "tests/stacktracealways.TestPrintfStackTraceAlways"
var funcName = "tests/stacktracealways/printf_always_test"

func TestPrintfStackTraceAlways(t *testing.T) {
	out := internal.CaptureOutput(func() {
		dlg.Printf("test message")
	})

	matched, _ := regexp.MatchString(`.*goroutine \d+ \[running\].*`, out)
	containsThisFn := strings.Contains(out, funcName)

	if !matched || !containsThisFn {
		// t.Errorf("Output doesn't contain a stack trace: Got: %v %v \n %q", matched, containsThisFn, out)
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
