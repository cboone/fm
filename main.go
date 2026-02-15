package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/cboone/fm/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		if !errors.Is(err, cmd.ErrSilent) {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		}
		os.Exit(1)
	}
}
