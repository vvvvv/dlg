//go:build dlg

package dlg_test

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/vvvvv/dlg"
	"github.com/vvvvv/dlg/tests/internal"
)

var (
	bannerRegexp  = regexp.MustCompile(`^[\*\s]+DEBUG BUILD[\*\s]+$`)
	logLineRegexp = regexp.MustCompile(`\d{2}:\d{2}:\d{2} \[.*\] printf_test\.go:\d+: test message`)
)

func TestPrintfBasic(t *testing.T) {
	out := internal.CaptureOutput(func() {
		dlg.Printf("test %s", "message")
	})

	matched := logLineRegexp.MatchString(out)
	if !matched {
		t.Errorf("Output format mismatch. Got: %q ; Want: %q", out, "test message")
	}
}

func TestPrintfNoDebugBanner(t *testing.T) {
	out := internal.CaptureOutput(func() {
		dlg.Printf("different %s message", "test")
	})

	matched := bannerRegexp.MatchString(out)
	if matched {
		t.Errorf("Expected no Debug Banner: Got: %q", out)
	}
}

func TestSetOutput(t *testing.T) {
	var buf bytes.Buffer

	dlg.SetOutput(&buf)
	defer dlg.SetOutput(os.Stderr)

	want := "custom output target"

	noOut := internal.CaptureOutput(func() {
		dlg.Printf(want)
	})

	if strings.Contains(noOut, want) {
		t.Error("Output was written to stderr but shouldn't")
	}

	out := buf.String()

	if !strings.Contains(out, want) {
		t.Errorf("Expected output in Writer: Got: %q ; Want: %q", out, want)
	}
}

func TestPrintfConcurrentWriter(t *testing.T) {
	buf := struct {
		sync.Mutex
		bytes.Buffer
	}{}

	dlg.SetOutput(&buf)
	defer dlg.SetOutput(os.Stderr)

	n := 100

	var wg sync.WaitGroup
	for i := range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			dlg.Printf("message from #%v", i)
		}()
	}
	wg.Wait()

	logLines := strings.Split(buf.String(), "\n")
	logLines = logLines[:len(logLines)-1] // last element contains empty string

	if len(logLines) != n {
		t.Errorf("Expected %v log lines but got: %v", n, len(logLines))
	}

	for n := 0; n < len(logLines); n++ {
		found := false
		want := fmt.Sprintf("message from %v", n)
		for _, line := range logLines {
			if strings.ContainsAny(line, want) {
				found = true
			}
		}

		if !found {
			t.Errorf("Expected log line %q not in buffer.", want)
		}

	}
}

var encoder = base64.StdEncoding.WithPadding(base64.NoPadding)

func randomString(n int) string {
	randomBytes := make([]byte, 128)
	_, err := rand.Read(randomBytes)
	if err != nil {
		panic(err)
	}
	dst := make([]byte, base64.StdEncoding.EncodedLen(len(randomBytes)))

	encoder.Encode(dst, randomBytes)
	return string(dst[:n])
}

func randomStrings(n int) []string {
	s := make([]string, 1000)

	for i := 0; i < len(s); i++ {
		s[i] = randomString(n)
	}

	return s
}

func BenchmarkPrintf8(b *testing.B) {
	var buf bytes.Buffer
	dlg.SetOutput(&buf)

	s := randomStrings(8)

	for i := 0; i < b.N; i++ {
		buf.Reset()
		dlg.Printf(s[i%len(s)])
	}
}

func BenchmarkPrintf16(b *testing.B) {
	var buf bytes.Buffer
	dlg.SetOutput(&buf)

	s := randomStrings(16)

	for i := 0; i < b.N; i++ {
		buf.Reset()
		dlg.Printf(s[i%len(s)])
	}
}

func BenchmarkPrintf32(b *testing.B) {
	var buf bytes.Buffer
	dlg.SetOutput(&buf)

	s := randomStrings(32)

	for i := 0; i < b.N; i++ {
		buf.Reset()
		dlg.Printf(s[i%len(s)])
	}
}

func BenchmarkPrintf64(b *testing.B) {
	var buf bytes.Buffer
	dlg.SetOutput(&buf)

	s := randomStrings(64)

	for i := 0; i < b.N; i++ {
		buf.Reset()
		dlg.Printf(s[i%len(s)])
	}
}

func BenchmarkPrintf128(b *testing.B) {
	var buf bytes.Buffer
	dlg.SetOutput(&buf)

	s := randomStrings(128)

	for i := 0; i < b.N; i++ {
		buf.Reset()
		dlg.Printf(s[i%len(s)])
	}
}

type safeBuffer struct {
	sync.Mutex
	bytes.Buffer
}

var safeBuf = &safeBuffer{}

func BenchmarkPrintf128Parallel(b *testing.B) {
	s := randomStrings(128)
	strCount := len(s)
	dlg.SetOutput(safeBuf)

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			if i == strCount {
				i = 0
			}
			str := s[i]
			dlg.Printf(str)
		}
	})
}
