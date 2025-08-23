package main

import (
	"fmt"
	"os"

	"github.com/griffin/go-shellify/cmd/shellify"
)

func main() {
	if err := shellify.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}