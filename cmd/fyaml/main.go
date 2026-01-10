package main

import (
	"errors"
	"os"

	"github.com/jksmth/fyaml"
	"github.com/jksmth/fyaml/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		if errors.Is(err, fyaml.ErrCheckMismatch) {
			os.Exit(2)
		}
		os.Exit(1)
	}
}
