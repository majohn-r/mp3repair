package main

import (
	"fmt"
	"mp3/internal"
	"mp3/internal/subcommands"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/utahta/go-cronowriter"
)

func main() {
	returnValue := 1
	if initEnv(internal.LookupEnvVars) {
		if initLogging(internal.TmpFolder) {
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

func initEnv(lookup func() []error) bool {
	if errors := lookup(); len(errors) > 0 {
		fmt.Fprintln(os.Stderr, "1 or more environment variables unset")
		for _, e := range errors {
			fmt.Fprintln(os.Stderr, e)
		}
		return false
	}
	return true
}

// exposed so that unit tests can close the writer!
var logger *cronowriter.CronoWriter

func initLogging(parentDir string) bool {
	path := filepath.Join(parentDir, "mp3", "logs")
	if err := os.MkdirAll(path, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "cannot create path '%s': %v\n", path, err)
		return false
	}
	logger = internal.ConfigureLogging(path)
	logrus.SetOutput(logger)
	internal.CleanupLogFiles(path)
	return true
}
