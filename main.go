package main

import (
	"errors"
	"fmt"
	"os"
	"saggy"
)

func main() {
	// Invoke the CLI
	if err := saggy.CLI(os.Args); err != nil {
		var SilentError *saggy.SilentError
		if errors.As(err, &SilentError) {
			os.Exit(SilentError.ExitCode)
		}
		var cliErr *saggy.CLIError
		if errors.As(err, &cliErr) {
			fmt.Fprintln(os.Stderr, err.Error())
			if cliErr.PrintUsage {
				fmt.Fprint(os.Stderr, saggy.USAGE_TEXT)
			}
			os.Exit(cliErr.Code)
		}
		saggyError := &saggy.SaggyError{}
		if errors.As(err, &saggyError) {
			fmt.Fprintln(os.Stderr, saggyError.Error())
			os.Exit(2)
		}

		fmt.Fprintln(os.Stderr, err)
		os.Exit(3)
	}
	os.Exit(0)
}
