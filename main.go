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
	// clilog.Env = "GOMI_LOG"
	// clilog.SetOutput()
	// defer log.Printf("[INFO] finish main function")
	//
	// log.Printf("[INFO] Version: %s (%s)", Version, Revision)
	// log.Printf("[INFO] gomiPath: %s", gomiPath)
	// log.Printf("[INFO] inventoryPath: %s", inventoryPath)
	// log.Printf("[INFO] Args: %#v", args)

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
