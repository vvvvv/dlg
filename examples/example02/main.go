package main

import (
	"bytes"
	"fmt"
	"io"
	"sync"

	"github.com/vvvvv/dlg"
)

type SafeWriter struct {
	w  io.Writer
	mu sync.Mutex
}

func (s *SafeWriter) Write(b []byte) (int, error) {
	return s.w.Write(b)
}

func (s *SafeWriter) Lock() {
	s.mu.Lock()
}

func (s *SafeWriter) Unlock() {
	s.mu.Unlock()
}

func main() {
	// Using a bytes.Buffer for demostrative purposes
	sink := &bytes.Buffer{}
	// But we could also do:
	//	sink, err := os.Create("debug.log")
	//	if err != nil {
	//		panic(err)
	//	}
	//	defer sink.Close()

	writer := &SafeWriter{
		w: sink,
	}

	dlg.SetOutput(writer)

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

	fmt.Print(sink.String())
}
