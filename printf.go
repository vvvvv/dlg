//go:build dlg

package dlg

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var (
	// Initial Printf buffer size
	bufSize = 128

	// DLG_STACKTRACE must be set using linker flags only:
	// e.g. go build -tags dlg -ldflags "-X github.com/vvvvv/dlg.DLG_STACKTRACE=ALWAYS"
	// Packages importing dlg MUST NOT read from or write to this variable - doing so will result in a compilation error when the dlg build tag is omitted.
	DLG_STACKTRACE = ""
)

// Include stack trace on error or on every call to Printf
var stackflags = 0

const (
	onerror = 1 << iota
	always
)

// Buffers for Printf
var bufPool = sync.Pool{
	New: func() interface{} { return make([]byte, 0, bufSize) },
}

func Printf(f string, v ...any) {
	b := bufPool.Get().([]byte)

	formatInfo(&b)
	f += "\n" // Without this we get an 'non-constant format string in call' error when v is left empty. Annoying
	b = fmt.Appendf(b, f, v...)

	if stackflags != 0 &&
		((stackflags&onerror != 0 && hasError(v)) ||
			(stackflags&always != 0)) {
		writeStack(&b)
	}

	writeOutput(b)

	// Remove buffers with a capacity greater than 32kb from the sync.Pool in order to keep the footprint small
	if cap(b) >= (1 << 15) {
		b = nil
	} else {
		b = b[:0]
	}

	bufPool.Put(b)
}

var stackBufSize uint32 = 1024

const maxStackTraceBufferSize uint32 = 1 << 20 // 1MB

// writeStack appends the stack trace to the buffer.
// Stack traces exceeding maxStackTraceBufferSize are truncated.
//
// Implementation detail:
// writeStack uses stackBufSize as the initial buffer size for stack traces.
// This value may be increased during execution if a stack trace exceeds the current buffer capacity.
//
// Updates to stackBufSize are guarded by sync/atomic to ensure safe concurrent modification.
// IMPORTANT: stackBufSize may contain stale values as reads are NOT guarded by sync/atomic!
// This is acceptable because stackBufSize is only used as a starting hint for buffer allocation.
// This is a performance optimization to save the overhead of calls to atomic.Read - eventually the correct value is going to get read.
func writeStack(buf *[]byte) {
	var b []byte
	var n int
	bufn := stackBufSize
	for ; bufn <= maxStackTraceBufferSize; bufn *= 2 {
		b = make([]byte, bufn)
		n = runtime.Stack(b, false)
		if n < len(b) {
			*buf = append(*buf, b[:n]...)
			*buf = append(*buf, '\n')
			atomic.StoreUint32(&stackBufSize, bufn)
			return
		}
	}

	*buf = append(*buf, b[:n]...)
	*buf = append(*buf, " [...] [ TRUNCATED ]\n"...)
	// TODO: should we call atomic.StoreUint32 here or treat this very large stack trace as an outlier?
}

// hasError returns whether any argument is an error.
func hasError(args []any) bool {
	for i := len(args) - 1; i >= 0; i-- {
		if _, ok := args[i].(error); ok {
			return true
		}
	}
	return false
}

// Calldepth for reporting file and line number
// 0 = callsite()
// 1 = formatInfo()
// 2 = Printf()
// 3 = caller of Printf
const calldepth = 3

func callsite(buf *[]byte) {
	_, file, line, ok := runtime.Caller(calldepth)
	if !ok {
		file = "no_file"
		line = 0
	} else {
		for i := len(file) - 1; i > 0; i-- {
			if file[i] == '/' {
				file = file[i+1:]
				break
			}
		}
	}

	*buf = append(*buf, file...)
	*buf = append(*buf, ':')
	pad(buf, line, -1)
	*buf = append(*buf, ": "...)
}

// Set on package init
var timeStart time.Time

// formatInfo appends timestamp, elapsed time, and source location to the buffer.
func formatInfo(buf *[]byte) {
	now := time.Now().UTC()
	// s := now.Sub(timeStart)
	since := now.Sub(timeStart).String()

	// Time in HH:MM:SS
	h, min, sec := now.Clock()
	padTime(buf, h, ':')
	padTime(buf, min, ':')
	padTime(buf, sec, 0)

	// Elapsed time
	elapsed(buf, &since)

	// Source file, line number
	callsite(buf)
}

func elapsed(buf *[]byte, since *string) {
	*buf = append(*buf, " ["...)
	*buf = append(*buf, *since...)
	*buf = append(*buf, "] "...)
}

// padTime formats hours, minutes, seconds as two digit values
// with an optional delimiter suffix.
func padTime(buf *[]byte, i int, delim byte) {
	if i < 10 {
		*buf = append(*buf, '0')
		*buf = append(*buf, byte('0'+i))
	} else {
		q := i / 10
		*buf = append(*buf, byte('0'+q))
		*buf = append(*buf, byte('0'+i-(q*10)))
	}
	if delim != 0 {
		*buf = append(*buf, delim)
	}
}

// pad i with zeros according to the specified width
func pad(buf *[]byte, i int, width int) {
	width -= 1
	var b [20]byte
	bp := len(b) - 1
	_ = b[bp]
	for ; (i >= 10 || width >= 1) && bp >= 0; width, bp = width-1, bp-1 {
		q := i / 10
		b[bp] = byte('0' + i - q*10)
		i = q
	}
	b[bp] = byte('0' + i)
	*buf = append(*buf, b[bp:]...)
}

var writeOutput = func(buf []byte) (n int, err error) {
	return os.Stderr.Write(buf)
}

var mu sync.Mutex

// SetOutput sets the output destination for Printf.
// Defaults to os.Stderr.
func SetOutput(w io.Writer) {
	mu.Lock()
	defer mu.Unlock()

	if w == nil {
		w = io.Discard
	}

	if locker, ok := w.(sync.Locker); ok {
		writeOutput = func(buf []byte) (n int, err error) {
			locker.Lock()
			n, err = w.Write(buf)
			locker.Unlock()
			return
		}
		return
	}

	writeOutput = func(buf []byte) (int, error) {
		return w.Write(buf)
	}
}

const envPrefix = "DLG_"

func env(name string) (v string, ok bool) {
	v, ok = os.LookupEnv(envPrefix + name)
	return strings.ToLower(v), ok
}

func init() {
	defer func() {
		timeStart = time.Now().UTC()
	}()

	// Warmup runtime.Caller and pre init the buffer pool.
	defer func() {
		runtime.Caller(0)
		bufPool.Put(make([]byte, 0, bufSize))
	}()

	// Controls whether the debug banner is printed at startup.
	// Set to any value (even empty) other than "0" to suppress the message.
	noWarning, ok := env("NO_WARN")
	if !ok || noWarning == "0" {
		// TODO: decide whether to check if we're running in a tty and if this should cause the flag to be ignored
		fmt.Fprint(
			os.Stderr,
			`* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * *
* * * * * * * * * * * * * *  DEBUG BUILD  * * * * * * * * * * * * * *
* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * *
- DLG_STACKTRACE=ERROR   show stack traces on errors
- DLG_STACKTRACE=ALWAYS  show stack traces always
- DLG_NO_WARN=1          disable this message (use at your own risk)

`)
	}

	// check if stack traces should get generated
	stacktrace := DLG_STACKTRACE
	if stacktrace == "" {
		if stacktrace, ok = env("STACKTRACE"); !ok {
			return
		}
	}
	stacktrace = strings.TrimSpace(strings.ToLower(stacktrace))

	unrecognizedArg := true
	if len(stacktrace) >= 3 {
		switch stacktrace[:3] {
		case "err":
			stackflags |= onerror
			bufSize += int(stackBufSize)
			unrecognizedArg = false
		case "alw":
			stackflags |= always
			bufSize += int(stackBufSize)
			unrecognizedArg = false
		default:
		}
	}
	if len(stacktrace) != 0 && unrecognizedArg {
		fmt.Fprintf(os.Stderr, "dlg: DLG_STACKTRACE set to '%s'. Did you mean 'ERROR' or 'ALWAYS'?\n\n", stacktrace)
	}
}
