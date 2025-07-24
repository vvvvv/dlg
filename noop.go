//go:build !dlg

package dlg

import (
	"io"
)

/*
Printf writes a formatted message to standard error when built with the dlg build tag. Formatting uses the same verbs as fmt (see https://pkg.go.dev/fmt#hdr-Printing).
It also supports optional stack trace generation, configurable at runtime via environment variables.

In builds without the dlg tag, Printf is a no-op.
*/
func Printf(fmt string, v ...any) {}

/*
SetOutput sets the output destination for Printf.
While Printf itself is safe for concurrent use, this guarantee does not extend to custom writers.
To ensure concurrency safety, the provided writer should implement [sync.Locker].

Calls to SetOutput should ideally be made at program initialization as they affect the logger globally.
Changing outputs during program execution, while concurrency-safe, may cause logs to temporarily continue appearing in the previous output. Eventually, all logs will be written to the new output.
For consistent output behavior, avoid changing writers after logging has started.
*/

func SetOutput(w io.Writer) {}
