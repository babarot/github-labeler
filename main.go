package main

import (
	"fmt"
	"os"

	"github.com/jessevdk/go-flags"
)

// These variables are set in Goreleaser
var (
	Version  = "unset"
	Revision = "unset"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	var opt Option
	args, err := flags.ParseArgs(&opt, args)
	if err != nil {
		return 2
	}

	cli := CLI{
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Option: opt,
	}

	if err := cli.Run(args); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}

	return 0
}
