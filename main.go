package main

import (
	"fmt"
	"os"

	"github.com/heguangyu1989/celo/cmd"
	"github.com/heguangyu1989/celo/pkg/p"
)

func main() {
	if err := cmd.Execute(); err != nil {
		p.Error(fmt.Sprintf("Error: %v", err))
		os.Exit(1)
	}
}
