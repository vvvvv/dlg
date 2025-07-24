<div align="center">
<h1>dlg <a href="https://github.com/vvvvv/dlg/actions/workflows/tests.yml"><img src="https://github.com/vvvvv/dlg/actions/workflows/tests.yml/badge.svg?branch=main" /></a></h1>
  <h3>Printf-Style Debugging with Zero-Cost in Production Builds</h3>
</div>

**dlg** provides a minimal API for printf-style debugging - a lightweight logger that completely vanishes from production builds while providing rich debugging capabilities during development.  
When built without the `dlg` tag, all logging calls disappear entirely from your binary, resulting in no runtime overhead.

### Why dlg?
- 🚀 **True zero-cost abstraction** - Logging calls completely disappear from production binaries  
- ⚡️ **Near-zero overhead** - Performance-focused design for debug builds  
- 🔍 **Smart stack traces** - Runtime-configurable stack trace generation  
- 🔒 **Concurrent-safe by design** - Custom writers simply implement `sync.Locker` to be safe
- ✨ **Minimalist API** - Exposes just two functions, `Printf` and `SetOutput`

### The Magic of Zero-Cost

```go
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
```

When compiled without the `dlg` build tag:

- All calls to `dlg` compile to empty functions
- Go linker completely eliminates these no-ops
- Final binary contains no trace of logging code
- Zero memory overhead, zero CPU impact

### Getting Started
```bash
go get github.com/vvvvv/dlg
```


### Activating Debug Mode

Enable debug features with the `dlg` build tag:
```bash
# Production build (no logging)
go build -o app

# Debug build 
go build -tags dlg -o app-debug
```

#### Debug build output:

```
./app-debug


* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * *
* * * * * * * * * * * * * *  DEBUG BUILD  * * * * * * * * * * * * * *
* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * *
- DLG_STACKTRACE=ERROR   show stack traces on errors
- DLG_STACKTRACE=ALWAYS  show stack traces always
- DLG_NO_WARN=1          disable this message (use at your own risk)

starting...
01:28:27 [2µs] main.go:16: executing risky operation
01:28:27 [21µs] main.go:19: something failed: unexpected error
01:28:27 [23µs] main.go:22: continuing
```

#### Stack Trace output:

```bash
DLG_STACKTRACE=ERROR ./app-debug

# [Debug Banner omitted]
starting...
01:31:34 [2µs] main.go:16: executing risky operation
01:31:34 [21µs] main.go:19: something failed: unexpected error
goroutine 1 [running]:
github.com/vvvvv/dlg.writeStack(0x14000104ec0)
    /Users/v/src/go/src/github.com/vvvvv/dlg/printf.go:86 +0x84
github.com/vvvvv/dlg.Printf({0x1002d25f4, 0x14}, {0x14000104f18, 0x1, 0x1})
    /Users/v/src/go/src/github.com/vvvvv/dlg/printf.go:50 +0x170
main.main()
    /Users/v/src/go/src/github.com/vvvvv/dlg/examples/example01/main.go:19 +0xc0

01:31:34 [38µs] main.go:22: continuing
```

#### Configuration

```bash
# Runtime configuration
DLG_STACKTRACE=ERROR  ./app-debug   # Traces on errors only
DLG_STACKTRACE=ALWAYS ./app-debug   # Traces on every call

# Compile-time configuration
go build -tags dlg -ldflags "-X 'github.com/vvvvv/dlg.DLG_STACKTRACE=ERROR'"
```

#### Suppressing the startup banner
```bash
DLG_NO_WARN=1 ./app-debug
```
**The debug banner cannot be disabled via linker flags. This prevents accidental deployment of debug builds to production.**

#### Concurrency Safety

While `dlg.Printf` is safe for concurrent use, custom writers should implement [sync.Locker](see https://pkg.go.dev/sync#Locker).

```go
package main

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/vvvvv/dlg"
)

type SafeBuffer struct {
	bytes.Buffer
	sync.Mutex
}

func main() {
	sb := &SafeBuffer{}
	dlg.SetOutput(sb) // Now fully concurrency-safe!

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for n := 0; n < 5; n++ {
				dlg.Printf("from goroutine #%v: message %v", i, n)
			}
		}()
	}
	wg.Wait()
	fmt.Print(sb.Buffer.String())
}
```
