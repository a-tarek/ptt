package main

import (
	"os"

	"github.com/a-tarek/ptt/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
