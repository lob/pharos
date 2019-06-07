package main

import (
	"os"

	"github.com/lob/pharos/pkg/pharos/cmd"
)

func main() {
	err := cmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
