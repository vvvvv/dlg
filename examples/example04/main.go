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
		dlg.Printf("something failed but we don't trace it: %v", err)
	}

	dlg.StartTrace()
	dlg.Printf("now we trace it")
	dlg.Printf("where did that error come from again?: %v", err)
	dlg.StopTrace()

	dlg.Printf("continuing")
}
