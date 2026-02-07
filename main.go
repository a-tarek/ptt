package main

import (
	"os"

	"github.com/ahmedelarabyy/wt/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
