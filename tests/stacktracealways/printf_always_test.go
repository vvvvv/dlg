//go:build dlg

package dlg_test_stacktrace_always

import (
	"regexp"
	"strings"
	"testing"

	"github.com/vvvvv/dlg"
	"github.com/vvvvv/dlg/tests/internal"
)

var funcName = "tests/stacktracealways.TestPrintfStackTraceAlways"

func TestPrintfStackTraceAlways(t *testing.T) {
	out := internal.CaptureOutput(func() {
		dlg.Printf("test message")
	})

	matched, _ := regexp.MatchString(`.*goroutine \d+ \[running\].*`, out)
	containsThisFn := strings.Contains(out, funcName)

	if !matched || !containsThisFn {
		t.Errorf("Output doesn't contain a stack trace: Got: %q", out)
	}
}
