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
