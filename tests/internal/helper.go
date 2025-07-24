package internal

import (
	"bytes"
	"io"
	"os"
)

func CaptureOutput(fn func()) string {
	stderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w
	defer func() {
		os.Stderr = stderr
	}()

	fn()

	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)

	return buf.String()
}
