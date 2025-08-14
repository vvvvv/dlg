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

/*
StartTrace begins a tracing region.

Tracing regions allow you to limit stack trace generation to specific parts of your code.
When DLG_STACKTRACE is set to "REGION,ALWAYS", Printf calls inside an active tracing region
will always include a stack trace. When set to "REGION,ERROR", only Printf calls inside the
region that include an error argument will include a stack trace.

A tracing region can be started with an optional key (any comparable value). If a key is
provided, only a matching StopTrace call with the same key will end that region. Without a key,
StopTrace ends the most recent active tracing region started in the same scope.

Tracing regions follow LIFO (last-in, first-out) order, and their scope is tied to the function
that called StartTrace. Even if StopTrace is called in a nested function, the region remains
active until closed in the starting scope-unless explicitly stopped by a matching key.

In builds without the dlg tag, StartTrace is a no-op.
*/
func StartTrace(v ...any) {}

/*
StopTrace ends a tracing region.

If called without arguments, StopTrace ends the most recent active tracing region started
in the same scope, regardless of whether it was keyed. If a key is provided, StopTrace ends
only the region that was started with the matching key, allowing you to close regions from
other functions or scopes.

Tracing regions are closed in LIFO (last-in, first-out) order.

In builds without the dlg tag, StopTrace is a no-op.
*/
func StopTrace(v ...any) {}
