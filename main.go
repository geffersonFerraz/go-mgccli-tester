package main

import (
	"fmt"
	"os"

	"github.com/geffersonFerraz/go-mgccli-tester/cmd"
)

func main() {
	err := cmd.Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}
