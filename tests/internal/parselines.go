package internal

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

type logline struct {
	line  string
	trace string
}

func (l logline) String() string {
	if l.trace != "" {
		return fmt.Sprintf("%s\n%s\n", l.line, l.trace)
	}
	return fmt.Sprintf("%s\n", l.line)
}

func (l logline) Line() string {
	lines := logSingleLineRegexp.FindStringSubmatch(l.line)
	if len(lines) > 1 {
		return strings.TrimSpace(lines[1])
	}

	return ""
}

func (l logline) HasTrace() bool {
	return l.trace != ""
}

var (
	// Gotta love writing regexes
	// This matches:
	//                                v everything until there (including possible stack traces)
	//  00:09:45 [4µs] main.go:16: foo 01:19:55 [8s] main.go:36: bar
	// logLineRegexp = regexp.MustCompile(`\d{2}:\d{2}:\d{2}\s+\[\d+\.?\d*.?s\]\s+\w+\.go:\d+:\s+.*?`)
	// logLineRegexp = regexp.MustCompile(`\d{2}:\d{2}:\d{2}\s+\[\d+\.?\d*.?s\]\s+\S+\.go:\d+:\s+.*?`)
	logLineRegexp = regexp.MustCompile(`\d{2}:\d{2}:\d{2}\s+\[[^\]]+\]\s+\S+\.go:\d+:\s+.*?`)
	// traceRegexp   = regexp.MustCompile(`goroutine \d+ \[running\]:`)
	traceRegexp = regexp.MustCompile(`(?:\S+\([^)]*\)\s+\S+\.go:\d+\s+\+0x[0-9A-Fa-f]+\s+)+`)
	// traceRegexp = regexp.MustCompile(`(?:[A-Za-z0-9_/.-]+\S+.go:\d+\s+\+0x[0-9A-Fa-f]+\s*)+`)

	logSingleLineRegexp = regexp.MustCompile(`\d{2}:\d{2}:\d{2}\s+\[\d+\.?\d*.?s\]\s+\w+\.go:\d+:\s(.*)$`)
)

// ParseLines parses dlg.Printf output into line, trace.
// This is done to simplify tests and should never be used outside of tests.
func ParseLines(b []byte) []logline {
	b = bytes.ReplaceAll(b, []byte("\n"), []byte(" "))
	loglines := make([]logline, 0, 32)

	for len(b) != 0 {
		// Find log line
		lineLoc := logLineRegexp.FindIndex(b)

		if lineLoc == nil {
			break
		}
		var (
			from int
			to   int
		)

		from = lineLoc[0]

		if next := logLineRegexp.FindIndex(b[from+1:]); next == nil {
			// No other match can be found.
			// This means we're at the last line -> b[from:len(b)]
			to = len(b)
		} else {
			// Slice to the beginning of the next match
			to = from + 1 + next[0]
		}

		if traceLoc := traceRegexp.FindIndex(b[from:to]); traceLoc == nil {
			// There's no stack trace
			loglines = append(loglines, logline{
				line: string(b[from:to]),
			})
		} else {
			// Found stack trace which means:
			// line: starts at 'from' ; ends at the beginning of the stacktrace
			// trace: starts at the beginning of the stacktrace ; ends at 'to'
			loglines = append(loglines, logline{
				line:  string(b[from:traceLoc[0]]),
				trace: string(b[traceLoc[0]:to]),
			})
		}

		b = b[to:]
	}

	return loglines
}

// // working partially
// func parseLines(b []byte) []logline {
// 	// Replace all newlines in order to have a single line of log output
// 	b = bytes.ReplaceAll(b, []byte("\n"), []byte(" "))
// 	r := bytes.NewReader(b)
//
// 	scanner := bufio.NewScanner(r)
// 	scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
// 		locs := logLineRegexp.FindAllIndex(data, -1)
//
// 		switch {
// 		case len(locs) >= 2:
// 			from, to := locs[0][0], locs[1][0]
// 			return to, data[from:to], nil
//
// 		case len(locs) == 1:
// 			if atEOF {
// 				from := locs[0][0]
// 				return len(data), data[from:], nil
// 			}
// 			// need more data to know the end boundary
// 			return 0, nil, nil
//
// 		default:
// 			if atEOF {
// 				return 0, nil, nil
// 			}
// 			return 0, nil, nil
// 		}
// 	})
//
// 	loglines := make([]logline, 0, 32)
// 	for scanner.Scan() {
// 		line := scanner.Bytes()
// 		loc := traceRegexp.FindIndex(line)
// 		if loc == nil {
// 			// Not a line with stack trace
// 			loglines = append(loglines, logline{line: string(line)})
// 		} else {
// 			loglines = append(loglines, logline{
// 				line:  string(line[:loc[0]]),
// 				trace: string(line[loc[0]:]),
// 			})
// 		}
// 	}
//
// 	return loglines
// }
//
// // almost working
// func parseLines(b []byte) []logline {
// 	// Replace all newlines in order to have a single line of log output
// 	b = bytes.ReplaceAll(b, []byte("\n"), []byte(" "))
// 	r := bytes.NewReader(b)
//
// 	scanner := bufio.NewScanner(r)
// 	scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
// 		locs := logLineRegexp.FindAllIndex(data, -1)
// 		if locs == nil || len(locs) < 2 {
// 			return 0, data, bufio.ErrFinalToken
// 		}
//
// 		// from is the start of the match.
// 		// to is the start of the next match.
// 		// This way possible stack trace lines get included.
// 		from, to := locs[0][0], locs[1][0]
//
// 		return to, data[from:to], nil
// 	})
//
// 	loglines := make([]logline, 0, 32)
// 	for scanner.Scan() {
// 		line := scanner.Bytes()
// 		loc := traceRegexp.FindIndex(line)
// 		if loc == nil {
// 			// Not a line with stack trace
// 			loglines = append(loglines, logline{line: string(line)})
// 		} else {
// 			loglines = append(loglines, logline{
// 				line:  string(line[:loc[0]]),
// 				trace: string(line[loc[0]:]),
// 			})
// 		}
// 	}
//
// 	return loglines
// }
