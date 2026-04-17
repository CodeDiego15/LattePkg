package main

import (
	"fmt"
	"os"

	"github.com/DiegoDev2/Fleet/internal/cli/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "armada: %s\n", err)
		os.Exit(1)
	}
}
