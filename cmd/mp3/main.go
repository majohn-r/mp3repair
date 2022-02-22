package main

import (
	"fmt"
	"mp3/internal"
	"mp3/internal/subcommands"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

func main() {
	returnValue := 1
	if initEnv() {
		if initLogging() {
			if cmd, args, err := subcommands.ProcessCommand(os.Args); err != nil {
				fmt.Fprintln(os.Stderr, err)
				logrus.Error(err)
			} else {
				cmd.Exec(args)
				returnValue = 0
			}
		}
	}
	os.Exit(returnValue)
}

func initEnv() bool {
	if errors := internal.LookupEnvVars(); len(errors) > 0 {
		fmt.Fprintln(os.Stderr, "1 or more environment variables unset")
		for _, e := range errors {
			fmt.Fprintln(os.Stderr, e)
		}
		return false
	}
	return true
}

func initLogging() bool {
	path := filepath.Join(internal.TmpFolder, "mp3", "logs")
	if err := os.MkdirAll(path, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "cannot create path '%s': %v\n", path, err)
		return false
	}
	logrus.SetOutput(internal.ConfigureLogging(path))
	internal.CleanupLogFiles(path)
	return true
}
