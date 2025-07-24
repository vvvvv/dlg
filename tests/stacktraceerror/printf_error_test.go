//go:build dlg

package dlg_test_stacktrace_error

import (
	"errors"
	"regexp"
	"strings"
	"testing"

	"github.com/vvvvv/dlg"
	"github.com/vvvvv/dlg/tests/internal"
)

var funcName = "dlg/tests/stacktraceerror.TestPrintfStackTraceOnError"

func TestPrintfStackTraceOnError(t *testing.T) {
	out := internal.CaptureOutput(func() {
		dlg.Printf("message with error: %v", errors.New("some error"))
	})

	matched, _ := regexp.MatchString(`.*goroutine \d+ \[running\].*`, out)
	containsThisFn := strings.Contains(out, funcName)

	if !matched || !containsThisFn {
		t.Errorf("Output doesn't contain a stack trace: Got: %q", out)
	}
}
