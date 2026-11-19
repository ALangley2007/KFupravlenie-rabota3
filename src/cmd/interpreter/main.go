package main

import (
	"os"

	"kf3/src/cli"
)

func main() {
	os.Exit(cli.RunInterpreter(os.Args[1:], os.Stderr))
}
