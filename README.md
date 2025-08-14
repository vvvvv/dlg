<div align="center">
    <h1>dlg <a href="https://github.com/vvvvv/dlg/actions/workflows/tests.yml"><img src="https://github.com/vvvvv/dlg/actions/workflows/tests.yml/badge.svg?branch=main" /></a></h1>
    <h3><em>delog - /diːˈlɑːɡ/ </em></h3>
    <h3>Printf-Style Debugging with Zero-Cost in Production Builds</h3>
</div>

**dlg** provides a minimal API for printf-style debugging - a lightweight logger that completely vanishes from production builds while providing rich debugging capabilities during development.  
When built without the `dlg` tag, all logging calls disappear entirely from your binary, resulting in no runtime overhead.

### Why dlg?
- 🚀 **True zero-cost abstraction** - Logging calls completely disappear from production binaries  
- ⚡️ **Near-zero overhead** - Performance-focused design for debug builds  
- 🔍 **Smart stack traces** - Runtime-configurable stack trace generation  
- 🔒 **Concurrency-safe by design** - Custom writers simply implement `sync.Locker` to be safe
- ✨ **Minimalist API** - Only `Printf`, and a couple of utility functions
- 🎨 **Colorize Output** - Highlight output for better visibility in noisy output

### The Magic of Zero-Cost

When compiled without the `dlg` build tag:

- All calls to `dlg` compile to empty functions
- Go linker completely eliminates these no-ops
- Final binary contains no trace of logging code
- Zero memory overhead, zero CPU impact

For the full technical breakdown, see [True Zero-Cost Elimination.](#true-zero-cost-elimination)

### Getting Started
```bash
go get github.com/vvvvv/dlg
```

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

### Activating Debug Mode

Enable debug features with the `dlg` build tag:
```bash
# Production build (no logging)
go build -o app

# Debug build 
go build -tags dlg -o app-debug
```

**Normal Output**  

```bash
$ go build -o app
./app
```

```
starting...
```

**Debug Build Output**

```bash
go build -tags dlg -o app-debug
./app-debug
```

```
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

**Stack Trace Output**

```bash
go build -tags dlg -o app-debug
DLG_STACKTRACE=ERROR ./app-debug
```

```
# [Debug Banner omitted]
starting...
01:31:34 [2µs] main.go:16: executing risky operation
01:31:34 [21µs] main.go:19: something failed: unexpected error
main.main()
    /Users/v/src/go/src/github.com/vvvvv/dlg/examples/example01/main.go:19 +0xc0
01:31:34 [38µs] main.go:22: continuing
```

### Tracing Regions <sup>experimental</sup>
Sometimes you only want stack traces for a specific area of your code while investigating an issue.
Tracing regions let you define those boundaries.

`dlg.StartTrace()` begins a tracing region, and `dlg.StopTrace()` ends it.  
When `DLG_STACKTRACE` is set to `REGION,ALWAYS`, `dlg.Printf` will print a stack trace only if the current call stack contains a function that's inside an active tracing region.  
Similarly, when set to `REGION,ERROR`, stack traces are printed inside tracing regions only if an error is passed to `dlg.Printf`.

#### Basic Usage

The simplest usage is to start and stop a trace around the code you want to inspect.  
Any `dlg.Printf` calls made in that tracing region will include stack traces.

Let's start with the most basic example:

```go
func main(){
    dlg.StartTrace()
    dlg.Printf("foobar")
    dlg.StopTrace()
}
```

**Output *`DLG_STACKTRACE=REGION,ALWAYS`***

```
16:14:39 [10µs] main.go:8: foobar
main.main()
    /Users/v/src/go/src/github.com/vvvvv/dlg/examples/example08/main.go:8 +0x4b
```

#### Tracing Across Functions

Tracing isn't limited to a single function. Once a tracing region is started, it covers all functions called within that region until it's stopped.
This means you can start tracing in main() and see traces for calls deeper in the stack.

```go 

func foo(){
    dlg.Printf("hello from foo")
}

func main(){
    dlg.Printf("outside of tracing region")

    dlg.StartTrace()
    dlg.Printf("started tracing")
    foo()
    dlg.StopTrace()
}
```

**Output *`DLG_STACKTRACE=REGION,ALWAYS`***

```
16:19:35 [1µs] main.go:11: outside of tracing region
16:19:35 [30µs] main.go:14: started tracing
main.main()
        /Users/v/src/go/src/github.com/vvvvv/dlg/examples/example08/main.go:14 +0x67
16:19:35 [43µs] main.go:7: hello from foo
main.foo()
        /Users/v/src/go/src/github.com/vvvvv/dlg/examples/example08/main.go:7 +0x87
main.main()
        /Users/v/src/go/src/github.com/vvvvv/dlg/examples/example08/main.go:15 +0x68
```


#### Error Tracing

If you only want stack traces when an error occurs inside a tracing region, set `DLG_STACKTRACE=REGION,ERROR`.
In this mode, traces appear only for `dlg.Printf` calls that include an error argument.

```go 

func main(){
    dlg.StartTrace()

    dlg.Printf("starting...")

    err := fmt.Errorf("this is an error") 

    dlg.Printf("oh no an error: %v", err)

    dlg.StopTrace()
}


```

**Output *`DLG_STACKTRACE=REGION,ERROR`***

```
16:24:20 [3µs] main.go:15: starting...
16:24:20 [29µs] main.go:19: oh no an error: this is an error
main.main()
    /Users/v/src/go/src/github.com/vvvvv/dlg/examples/example08/main.go:19 +0x97
```

#### Understanding Tracing Region Scopes

A tracing region is tied to the function scope that called `StartTrace`.
If you call `StopTrace()` inside a nested function, the tracing region remains active as the region was started from the outer scope.

```go
func main(){
    dlg.StartTrace()

    dlg.Printf("starting...")

    fn := func(){
        dlg.Printf("hello from fn")

        dlg.StopTrace()
    }

    fn()

    dlg.Printf("this will still produce a stack trace")

    dlg.StopTrace()
}

```

**Output *`DLG_STACKTRACE=REGION,ALWAYS`***

```
16:28:27 [12µs] main.go:15: starting...
main.main()
        /Users/v/src/go/src/github.com/vvvvv/dlg/examples/example08/main.go:15 +0x4b
16:28:27 [43µs] main.go:18: hello from fn
main.main.func1()
        /Users/v/src/go/src/github.com/vvvvv/dlg/examples/example08/main.go:18 +0x6b
main.main()
        /Users/v/src/go/src/github.com/vvvvv/dlg/examples/example08/main.go:22 +0x4c
16:28:27 [49µs] main.go:24: this will still produce a stack trace
main.main()
        /Users/v/src/go/src/github.com/vvvvv/dlg/examples/example08/main.go:24 +0x9f
```


#### Stopping a Tracing Region from Another Function

To close a tracing region from a nested function, you need to start and stop it with a matching key.
This allows you to end exactly the tracing region you intended, even from different scopes.


```go
func main(){
    dlg.StartTrace(1)

    dlg.Printf("starting...")

    fn := func(){
        dlg.Printf("hello from fn")

        dlg.StopTrace(1)
    }

    fn()

    dlg.Printf("this won't trace")
}

```

**Output *`DLG_STACKTRACE=REGION,ALWAYS`***

```
16:34:07 [9µs] main.go:15: starting...
main.main()
        /Users/v/src/go/src/github.com/vvvvv/dlg/examples/example08/main.go:15 +0x6b
16:34:07 [33µs] main.go:18: hello from fn
main.main.func1()
        /Users/v/src/go/src/github.com/vvvvv/dlg/examples/example08/main.go:18 +0x8b
main.main()
        /Users/v/src/go/src/github.com/vvvvv/dlg/examples/example08/main.go:23 +0x6c
16:34:07 [48µs] main.go:25: this won't trace
```

#### Choosing a Key

You can use any type as a tracing key: integers, strings, floats, even structs.
For clarity, it's best to keep keys simple, such as short strings or integers.

```go

dlg.StartTrace("foo")
...
dlg.StopTrace("foo")


dlg.StartTrace(7.2)
...
dlg.StopTrace(7.2)


dlg.StartTrace(struct{name string}{name: "tracing region"})
...

dlg.StopTrace(struct{name string}{name: "tracing region"})
```

#### Stopping Without a Key

`StopTrace()` without arguments will end the most recent active tracing region, even if it was started with a key - as long as you call it from the same scope.

```go
func main(){
    dlg.StartTrace(1)

    dlg.Printf("this will trace")

    dlg.StopTrace()

    dlg.Printf("this won't trace")
}
```

**Output *`DLG_STACKTRACE=REGION,ALWAYS`***

```
16:47:04 [10µs] main.go:14: this will trace
main.main()
        /Users/v/src/go/src/github.com/vvvvv/dlg/examples/example08/main.go:14 +0x6b
16:47:04 [34µs] main.go:18: this won't trace
```

> 💡 All tracing regions, whether keyed or not, are closed in LIFO *(last-in, first-out)*  order.

#### ⚠️ Why You Should Avoid `defer StopTrace()`

It might be tempting to wrap `dlg.StopTrace()` in a `defer`, but don't.
The Go compiler cannot eliminate `defer` calls. Even something as trivial as `defer func(){}()` remains as a real function call in the compiled binary.
If you want true zero-cost elimination, call `StopTrace` directly.

*For more examples of tracing regions, see /tests/stacktraceregion/region_test.go.*


### Concurrency Safety for Custom Writers

While `dlg.Printf` is safe for concurrent use, custom writers should implement [sync.Locker](https://pkg.go.dev/sync#Locker).

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

### Configuration

`dlg` can be configured at runtime *(environment variables)* or at compile time *(linker flags)*.
Compile-time settings win over runtime.
Settings configured at compile time cannot be overridden at runtime.

|  Variable          | Runtime-configurable | Compile-time-configurable | Description                             |
| ------------------ | -------------------- | --------------------------| --------------------------------------- |
| DLG_STACKTRACE     | ✔︎                    | ✔︎                         | Controls when stack traces are shown    |
| DLG_COLOR          | ✘                    | ✔︎                         | Sets output color for file/line         |
| DLG_NO_WARN        | ✔︎                    | ✘                         | Suppresses debug banner                 |


**DLG_STACKTRACE - Controls when to generate stack traces**

*Runtime:*
```bash
# Errors only
DLG_STACKTRACE=ERROR         ./app-debug
# Every call
DLG_STACKTRACE=ALWAYS        ./app-debug
# Errors within tracing region
DLG_STACKTRACE=REGION,ERROR  ./app-debug
# Every call within tracing region
DLG_STACKTRACE=REGION,ALWAYS ./app-debug
```


*Compile-time:*
```bash
go build -tags dlg -ldflags "-X 'github.com/vvvvv/dlg.DLG_STACKTRACE=ERROR'"
go build -tags dlg -ldflags "-X 'github.com/vvvvv/dlg.DLG_STACKTRACE=REGION,ALWAYS'"
```

**DLG_NO_WARN - Suppress the debug startup banner**  

*Runtime:*
```bash
DLG_NO_WARN=1 ./app-debug
```
> The debug banner cannot be disabled via linker flags. This prevents accidental deployment of debug builds to production.

**DLG_COLOR - Highlight file name & line number**  

*Compile-time:*
```bash
# Set the color to ANSI color red.
go build -tags dlg -ldflags "-X 'github.com/vvvvv/dlg.DLG_COLOR=red'"
# Set the color to ANSI color 4.
go build -tags dlg -ldflags "-X 'github.com/vvvvv/dlg.DLG_COLOR=4'"
# Set raw ANSI color.
go build -tags dlg -ldflags "-X 'github.com/vvvvv/dlg.DLG_COLOR=\033[38;2;250;3;250m'"
```
> This setting respects the `NO_COLOR` convention

Valid color values: 
- Named: *black, red, green, yellow, blue, magenta, cyan, white*
- ANSI: *0 - 255*
- Raw ANSI color escape sequences

### True Zero-Cost Elimination

The term "zero-cost" isn't just a claim - it's a verifiable compiler behavior. When dlg is disabled, the Go toolchain performs complete dead code elimination.

Consider this simple program:

```go
package main

import (
	"fmt"
	"github.com/vvvvv/dlg"
)

func main() {
	fmt.Println("hello world")
	dlg.Printf("hello from dlg")
}
```

When built *without* the `dlg` tag:

```bash
go build -o production_binary
```

The resulting disassembly (via `go tool objdump -s main.main production_binary`) shows:


```assembly
  ... [function prologue] ...
  main.go:10        0x10009c6a8      b00001a5      ADRP 217088(PC), R5
  main.go:10        0x10009c6ac      913480a5      ADD $3360, R5, R5
  main.go:10        0x10009c6b0      f9001fe5      MOVD R5, 56(RSP)
  main.go:10        0x10009c6b4      f0000265      ADRP 323584(PC), R5
  main.go:10        0x10009c6b8      9135a0a5      ADD $3432, R5, R5
  main.go:10        0x10009c6bc      f90023e5      MOVD R5, 64(RSP)
  print.go:314      0x10009c6c0      b00006db      ADRP 888832(PC), R27
  print.go:314      0x10009c6c4      f9479761      MOVD 3880(R27), R1
  print.go:314      0x10009c6c8      90000280      ADRP 327680(PC), R0
  print.go:314      0x10009c6cc      910c6000      ADD $792, R0, R0
  print.go:314      0x10009c6d0      9100e3e2      ADD $56, RSP, R2
  print.go:314      0x10009c6d4      b24003e3      ORR $1, ZR, R3
  print.go:314      0x10009c6d8      aa0303e4      MOVD R3, R4
  print.go:314      0x10009c6dc      97ffecf5      CALL fmt.Fprintln(SB)    ; Only this call (fmt.Println) remains
  main.go:12        0x10009c6e0      f85f83fd      MOVD -8(RSP), R29
  main.go:12        0x10009c6e4      f84507fe      MOVD.P 80(RSP), R30
  main.go:12        0x10009c6e8      d65f03c0      RET
  main.go:9         0x10009c6ec      aa1e03e3      MOVD R30, R3
  main.go:9         0x10009c6f0      97ff3bbc      CALL runtime.morestack_noctxt.abi0(SB)
  main.go:9         0x10009c6f4      17ffffe7      JMP main.main(SB)
  main.go:9         0x10009c6f8      00000000      ?
```


The compiler eliminates `dlg` as if it was never imported.  

However, the Go compiler can only eliminate `dlg.Printf` calls if it can prove that the arguments themselves have no side effects and are not used elsewhere.  
This has two important implications:
1. **Referenced variables stay:** If a variable is used outside the `dlg.Printf` call, the computation remains - but the call itself is still removed.
2. **Function calls are evaluated:** Even if `dlg.Printf` is eliminated, any argument expressions with potential side effects (e.g. function calls) are still evaluated.

Let's look at practical examples.

**✅ Referenced Variables - Printf Eliminated, Variable Remains**
```go
res := 69 * 42             // Used later -> remains
dlg.Printf("res: %v", res) // Eliminated
fmt.Println("result: ", res)
```

```assembly
; res remains (used by fmt.Println)
0x10009c6c8  d2816a40  MOVD $2898, R0   ; 69*42=2898 stored
...
; fmt.Println remains
0x10009c6fc  97ffeced  CALL fmt.Fprintln(SB)
```

**✅ Unused Expressions - Fully Eliminated**

```go
res := 69 * 42                  // Eliminated
dlg.Printf("res: %v", res)      // Eliminated
dlg.Printf("calc: %v", 69 * 42) // Eliminated
```

```assembly
; Entire function reduced to a single return:
0x100067b60  d65f03c0  RET
```

**⚠️ Function Calls - Still Evaluated[^1]**

```go
// The call to fmt.Errorf is evaluated but the call to dlg.Printf is still eliminated
dlg.Printf("call to fn: %v", fmt.Errorf("some error"))
```


```assembly
; fmt.Errorf is still executed
0x10009f21c  913ec000  ADD $4016, R0, R0
0x10009f230  97ffd8e0  CALL fmt.Errorf(SB)
```


### ⚡️Rule of Thumb:
**Avoid placing function calls or expensive computations directly inside `dlg.Printf`.**

As long as you follow this principle, `dlg` maintains its promise:  
***No instructions.*  
*No references.*  
*Zero memory allocations.*  
*Zero CPU cycles used.*  
*identical binary size to code without `dlg`.*  
*True zero-cost.***

[^1]: There's a bit more nuance to this - if a function is side-effect free and returns a basic type (e.g., `int`, `string`), the compiler may still eliminate the function call.
