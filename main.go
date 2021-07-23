package main

import (
	"fmt"
	"os"

	"github.com/kyverno/policy-reporter/cmd"
)

func main() {
	if err := cmd.NewCLI().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
