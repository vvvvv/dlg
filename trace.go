//go:build dlg

package dlg

import (
	"reflect"
	"runtime"
	"sync"
	"sync/atomic"
)

const maxFrames = 64

var pcPool = sync.Pool{
	New: func() any { return make([]uintptr, maxFrames) },
}

// writeStack appends a formatted stack trace to the provided byte buffer.
// It captures stack frames starting from the actual caller until reaching go internal frames.
// The stack trace contains:
// 1. The caller function name (e.g. main.main() )
// 2. The file path and line number (e.g. main.go:69)
// 3. The PC offset from the function entry in hexadecimal
func writeStack(buf *[]byte) {
	// calldepth skips n frames to report the correct file and line number
	// 0 = runtime -> extern.go
	// 1 = writeStack -> trace.go
	// 2 = Printf -> printf.go
	// 3 = callerFn
	const calldepth = 3

	pcs := pcPool.Get().([]uintptr)

	n := runtime.Callers(calldepth, pcs)
	if n > maxFrames {
		n = maxFrames
	}
	pcs = pcs[:n]

	frames := runtime.CallersFrames(pcs)
	for {
		frame, more := frames.Next()

		fnName := frame.Function
		if isAtRuntimeCalldepth(fnName) {
			// If we've reached go internal frames don't descend any deeper.
			break
		}

		if fnName == "" {
			fnName = "unknown"
		}

		// Caller function name
		// e.g. main.main()
		*buf = append(*buf, fnName...)
		*buf = append(*buf, "()\n\t"...)

		// File name:line number
		*buf = append(*buf, frame.File...)
		*buf = append(*buf, ':')
		pad(buf, frame.Line, -1)

		// PC offset in hex
		off := uintptr(0)
		if frame.Entry != 0 && frame.PC >= frame.Entry {
			off = frame.PC - frame.Entry
		}
		*buf = append(*buf, " +0x"...)
		appendHex(buf, uint64(off))
		*buf = append(*buf, '\n')

		if !more {
			break
		}
	}

	pcPool.Put(pcs[:cap(pcs)])
}

// isAtRuntimeCalldepth checks if a frame's function is at runtime level depth.
func isAtRuntimeCalldepth(fn string) bool {
	return fn == "runtime.main" || fn == "runtime.goexit" || fn == "testing.tRunner"
}

// appendHex converts n into hexadecimal for n >= 0 and appends it to buf.
// Negative numbers are not handled.
func appendHex(buf *[]byte, n uint64) {
	const hexd = "0123456789abcdef"
	var b [16]byte
	i := len(b) - 1

	_ = hexd[i]
	for ; (n > 0xF) && i >= 0; i-- {
		b[i] = hexd[n&0xF]
		n = n >> 4
	}
	b[i] = hexd[n&0xF]
	*buf = append(*buf, b[i:]...)
}

type caller struct {
	key            []any
	id             string
	pc             uintptr
	lpc            uintptr
	runFuncForPC   uintptr
	frameEntry     uintptr
	frameFuncEntry uintptr
}

var (
	callersMu    sync.RWMutex
	callersStore atomic.Value
	traceCount   int32
)

func deleteItemAt[T any](s []T, idx int) []T {
	_ = s[idx]
	res := make([]T, len(s)-1)
	copy(res[:idx], s[:idx])
	copy(res[idx:], s[idx+1:])
	return res
}

func StartTrace(v ...any) {
	startTrace(2, v)
}

// startTrace marks the current caller as a tracing region.
//
// skip controls how many stack frames above startTrace to skip before capturing
// the region entry in order to get the actual callsite.
// key is an optional identifier which may later be used to stop the matching region via stopTrace.
//
// Internally the function records the caller's function identifier and entry PC,
// appending a new entry to callersStore.
//
// On error this function fails silently.
func startTrace(skip int, key []any) {
	pc := make([]uintptr, 1)
	n := runtime.Callers(skip+1, pc)
	if n == 0 {
		return
	}

	pc = pc[:n]
	frames := runtime.CallersFrames(pc)

	frame, _ := frames.Next()
	if frame.Function == "" || frame.Func == nil {
		// We cannot identify the function.
		// This may happen in FFI.
		return
	}

	c := caller{key: key, id: frame.Function, pc: frame.Entry}

	callersMu.Lock()
	defer callersMu.Unlock()

	callers := callersStore.Load().([]caller)

	newCallers := make([]caller, len(callers)+1)
	copy(newCallers, callers)
	newCallers[len(callers)] = c
	callersStore.Store(newCallers)

	atomic.AddInt32(&traceCount, 1)
}

func StopTrace(v ...any) {
	stopTrace(2, v)
}

// stopTrace closes a previously started tracing region.
//
// It closes the the most recent started region.
// If key is non-nil, the most recent region whose key matches is closed.
//
// On success callersStore is updated.
// On error this function fails silently.
func stopTrace(skip int, key []any) {
	if tc := atomic.LoadInt32(&traceCount); tc == 0 {
		// TODO: maybe panic here?
		return
	}

	var newCallers []caller

	if key != nil {
		// Check if we find a region with the matching key.
		// If we don't find one return.
		callersMu.Lock()
		defer callersMu.Unlock()
		callers := callersStore.Load().([]caller)
		for i := len(callers) - 1; i >= 0; i-- {
			c := callers[i]

			if reflect.DeepEqual(c.key, key) {
				// Found it.
				newCallers = deleteItemAt(callers, i)
				callersStore.Store(newCallers)
				atomic.AddInt32(&traceCount, -1)
				return
			}
		}

		// TODO: should this panic or fail silently?
		return
	}

	pc := make([]uintptr, 1)
	n := runtime.Callers(skip+1, pc)
	if n == 0 {
		return
	}

	pc = pc[:n]
	frames := runtime.CallersFrames(pc)

	frame, _ := frames.Next()
	if frame.Function == "" || frame.Func == nil {
		// We cannot identify the function.
		// This may happen in FFI.
		return
	}

	callersMu.Lock()
	defer callersMu.Unlock()
	callers := callersStore.Load().([]caller)

	// Check if this frame has an open region.
	for i := len(callers) - 1; i >= 0; i-- {
		c := callers[i]
		if c.id == frame.Function && c.pc == frame.Entry {
			// Found it.
			newCallers = deleteItemAt(callers, i)
			callersStore.Store(newCallers)
			atomic.AddInt32(&traceCount, -1)
			return
		}
	}
	return
}

// inTracingRegion reports whether any frame in the current call stack is inside a tracked tracing region.
func inTracingRegion(skip int) bool {
	callers := callersStore.Load().([]caller)
	if len(callers) == 0 {
		return false
	}

	pcs := make([]uintptr, maxFrames)
	n := runtime.Callers(skip+1, pcs[:])
	if n == 0 {
		return false
	}

	frames := runtime.CallersFrames(pcs[:n])
	for i := 0; i < n; i++ {
		frame, more := frames.Next()
		if frame.Func == nil {
			// We don't know which function we're in.
			// This may happen in FFI
			// TODO: should we just break here?
			continue
		}
		entry := frame.Entry

		for j := len(callers) - 1; j >= 0; j-- {
			if callers[j].pc == entry {
				return true
			}
		}
		if !more {
			break
		}
	}
	return false
}
