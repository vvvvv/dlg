/*
Package dlg implements a zero-cost debug logger that compiles down to a no-op unless built using the build tag dlg.
The package exports just two functions: dlg.Printf(format string, v ...any) and dlg.SetOutput(w io.Writer).
The format string uses the same verbs as fmt (see https://pkg.go.dev/fmt#hdr-Printing).

Printf defaults to writing to standard error but SetOutput may be used to change the output destination.
While Printf itself is safe for concurrent use, this guarantee does not extend to custom output writers.
To ensure concurrency safety, the provided writer should implement sync.Locker (see https://pkg.go.dev/sync#Locker).

Printf also supports optional stack trace generation, either if Printf contains an error, or on every call.
Stack traces can be activated either at runtime by setting Environment variables or at compile time via a linker flag.

Usage:

	package main

	import (
		"fmt"

		"github.com/vvvvv/dlg"
	)

	func risky() error {
		return fmt.Errorf("unexpected error")
	}

	func main() {
		fmt.Println("starting...")

		dlg.Printf("executing risky operation")
		err := risky()
		if err != nil {
			dlg.Printf("something failed: %v", err)
		}

		dlg.Printf("continuing")
	}

Compiling without the dlg build tag:

	go build -o example
	./example

	starting...

Compiling with dlg activated:

	go build -tags=dlg -o example
	./example

	​ * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * *
	​ * * * * * * * * * * * * * *  DEBUG BUILD  * * * * * * * * * * * * * *
	​ * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * *
	​ - DLG_STACKTRACE=ERROR   show stack traces on errors
	​ - DLG_STACKTRACE=ALWAYS  show stack traces always
	​ - DLG_NO_WARN=1          disable this message (use at your own risk)

	starting...
	01:28:27 [2µs] main.go:16: executing risky operation
	01:28:27 [21µs] main.go:19: something failed: unexpected error
	01:28:27 [23µs] main.go:22: continuing

Note: To abbreviate the example output below, the debug banner has been omitted.

Including stack traces in the output we simply set an environment variable when executing the binary:

	DLG_STACKTRACE=ERROR ./example

	starting...
	01:31:34 [2µs] main.go:16: executing risky operation
	01:31:34 [21µs] main.go:19: something failed: unexpected error
	main.main()
		/Users/v/src/go/src/github.com/vvvvv/dlg/examples/example01/main.go:19 +0xc0

	01:31:34 [38µs] main.go:22: continuing

Including stack traces every time Printf is called:

	DLG_STACKTRACE=ALWAYS ./example

	starting...
	01:35:47 [2µs] main.go:16: executing risky operation
	main.main()
		/Users/v/src/go/src/github.com/vvvvv/dlg/examples/example01/main.go:16 +0x6c

	01:35:47 [34µs] main.go:19: something failed: unexpected error
	main.main()
		/Users/v/src/go/src/github.com/vvvvv/dlg/examples/example01/main.go:19 +0xc0

	01:35:47 [41µs] main.go:22: continuing
	main.main()
		/Users/v/src/go/src/github.com/vvvvv/dlg/examples/example01/main.go:22 +0xdc

Compiling with stack traces activated:

	go build -tags dlg -ldflags "-X 'github.com/vvvvv/dlg.DLG_STACKTRACE=ERROR'"

Package Configuration:

Outputs a stack trace if any of the arguments passed to Printf is of type error:

	export DLG_STACKTRACE=ERROR
	// or
	-ldflags "-X 'github.com/vvvvv/dlg.DLG_STACKTRACE=ERROR'"

Outputs a stack trace *on every call* to Printf, regardless of the arguments:

	export DLG_STACKTRACE=ALWAYS
	// or
	-ldflags "-X 'github.com/vvvvv/dlg.DLG_STACKTRACE=ALWAYS'"

Suppresses the debug banner printed at startup:

	// Note: DLG_NO_WARN cannot be set via -ldflags. This decision was made so debug builds will never accidentally land in production.
	DLG_NO_WARN=1
*/
package dlg
