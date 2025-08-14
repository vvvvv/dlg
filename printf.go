//go:build dlg

package dlg

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

const stackBufSize = 1024

var (
	// Initial Printf buffer size
	bufSize = 128

	// DLG_STACKTRACE and DLG_COLORS must be set using linker flags only:
	// e.g. go build -tags dlg -ldflags "-X github.com/vvvvv/dlg.DLG_STACKTRACE=ALWAYS"
	// Packages importing dlg MUST NOT read from or write to this variable - doing so won't have any effect and will result in a compilation error when the dlg build tag is omitted.
	DLG_STACKTRACE = ""
	DLG_COLOR      = ""

	termColor []byte
)

// Include stack trace on error or on every call to Printf
var stackflags = 0

const (
	onerror = 1 << iota
	always
	region
)

// Buffers for Printf
var bufPool = sync.Pool{
	New: func() interface{} { return make([]byte, 0, bufSize) },
}

func Printf(f string, v ...any) {
	b := bufPool.Get().([]byte)

	formatInfo(&b)
	if len(v) == 0 && strings.IndexByte(f, '%') < 0 {
		// If there's no formatting we take a fast path
		b = append(b, f...)
		b = append(b, '\n')
	} else {
		f += "\n" // Without this we get an 'non-constant format string in call' error when v is left empty. Annoying
		b = fmt.Appendf(b, f, v...)
	}

	if stackflags != 0 &&
		((stackflags&onerror != 0 && hasError(v)) ||
			(stackflags&always != 0)) {

		if (stackflags&region != 0 && inTracingRegion(1)) || (stackflags&region == 0) {
			writeStack(&b)
		}
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

// Set on package init
var timeStart time.Time

// formatInfo appends timestamp, elapsed time, and source location to the buffer.
func formatInfo(buf *[]byte) {
	now := time.Now().UTC()
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

func elapsed(buf *[]byte, since *string) {
	*buf = append(*buf, " ["...)
	*buf = append(*buf, *since...)
	*buf = append(*buf, "] "...)
}

// pad i with zeros according to the specified width
func pad(buf *[]byte, i int, width int) {
	width -= 1
	var b [20]byte
	bp := len(b) - 1
	// _ = b[bp]
	for ; (i >= 10 || width >= 1) && bp >= 0; width, bp = width-1, bp-1 {
		q := i / 10
		b[bp] = byte('0' + i - q*10)
		i = q
	}
	b[bp] = byte('0' + i)
	*buf = append(*buf, b[bp:]...)
}

// callsite Appends the filename and line number.
// It optionally colors the output.
func callsite(buf *[]byte) {
	// Calldepth skips n frames for reporting the correct file and line number
	// 0 = runtime -> extern.go
	// 1 = callsite -> printf.go
	// 2 = formatInfo -> printf.go
	// 3 = Printf -> printf.go
	// 4 = callerFn
	const calldepth = 4
	pcs := make([]uintptr, 1)
	n := runtime.Callers(calldepth, pcs)

	fileName := "no_file"
	lineNr := 0
	if n != 0 {
		frames := runtime.CallersFrames(pcs)
		frame, _ := frames.Next()

		fileName = frame.File
		for i := len(fileName) - 1; i > 0; i-- {
			if fileName[i] == '/' {
				fileName = fileName[i+1:]
				break
			}
		}

		lineNr = frame.Line
	}

	// File name:line number
	colorizeOrDont(buf)
	*buf = append(*buf, fileName...)
	*buf = append(*buf, ':')
	pad(buf, lineNr, -1)
	resetColorOrDont(buf)
	*buf = append(*buf, ": "...)
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

// writeOutput is the default output writer.
// This function gets set by SetOutput.
var writeOutput = func(buf []byte) (n int, err error) {
	return os.Stderr.Write(buf)
}

// Protects calls to SetOutput
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

func env(name string) (v string, ok bool) {
	const envPrefix = "DLG_"
	v, ok = os.LookupEnv(envPrefix + name)
	return strings.ToLower(v), ok
}

func init() {
	defer func() {
		timeStart = time.Now().UTC()
	}()

	// Warmup runtime.Caller and pre init the buffer pool.
	defer func() {
		// Warmup runtime.Caller
		runtime.Caller(0)

		// Pre init buffer pool
		bufPool.Put(make([]byte, 0, bufSize))

		// Initialize trace region store
		callers := make([]caller, 0, 16)
		callersStore.Store(callers)
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
- DLG_STACKTRACE=REGION  show stack traces in trace regions 
- DLG_NO_WARN=1          disable this message (use at your own risk)

`)
	}

	// Check if output should get colorized
	if color, ok := colorArgToTermColor(DLG_COLOR); ok {
		if _, noColor := os.LookupEnv("NO_COLOR"); noColor {
			// Respect NO_COLOR
		} else {
			termColor = color
			colorizeOrDont = colorize
			resetColorOrDont = colorReset
		}
	}

	// check if stack traces should get generated
	stacktrace := DLG_STACKTRACE
	if stacktrace == "" {
		if stacktrace, ok = env("STACKTRACE"); !ok {
			return
		}
	}

	// TODO: Should we fail hard if there's an invalid stack trace argument?
	// Notify the user about unrecognized stack trace arguments but use valid ones regardless.
	var err error
	stackflags, err = parseTraceArgs(stacktrace)
	if stackflags != 0 {
		// Increase initial buffer size to accommodate stack traces
		bufSize += int(stackBufSize)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, " dlg: Invalid Argument %v\n", err)
	}
}

func parseTraceArgs(arg string) (stackflags int, err error) {
	args := strings.Split(strings.ToLower(arg), ",")

	if len(args) > 2 {
		err = fmt.Errorf("DLG_STACKTRACE: too many arguments")
		return
	}

	// Should be a []error but errors.Join separates errors with newlines.
	// We want them seperated by ',' so we have to fall back to strings.Join here.
	var invalidArgsErr []string

	for i := 0; i < len(args); i++ {
		flag, parseErr := parseTraceOption(args[i])
		if parseErr != nil {
			invalidArgsErr = append(invalidArgsErr, parseErr.Error())
		} else {
			stackflags |= flag
		}
	}

	if len(invalidArgsErr) > 0 {
		err = fmt.Errorf("DLG_STACKTRACE: %s.\n", strings.Join(invalidArgsErr, ", "))
	}
	return
}

func parseTraceOption(opt string) (stackflag int, err error) {
	const (
		traceOptError  = "err"
		traceOptAlways = "alw"
		traceOptRegion = "reg"
	)

	err = fmt.Errorf("invalid argument %q", opt)
	if len(opt) < 3 {
		return
	}

	switch opt[:3] {
	case traceOptError:
		stackflag |= onerror
		err = nil
	case traceOptAlways:
		stackflag |= always
		err = nil
	case traceOptRegion:
		stackflag |= region
		err = nil
	}

	return
}
