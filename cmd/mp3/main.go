package main

import (
	"fmt"
	"io"
	"mp3/internal"
	"mp3/internal/subcommands"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/utahta/go-cronowriter"
)

// these variables' values are injected by the mage build
var (
	// semantic version; read by the mage build from version.txt
	version string = "unknown version!"
	// build timestamp in RFC3339 format (2006-01-02T15:04:05Z07:00)
	creation string
)

func main() {
	returnValue := 1
	if initEnv(internal.LookupEnvVars) {
		if initLogging(internal.TemporaryFileFolder()) {
			returnValue = run(os.Args)
		}
	}
	report(os.Stderr, returnValue)
	os.Exit(returnValue)
}

const (
	statusFormat = "mp3 version %s, created at %s, failed\n"
	logDirName   = "logs"
)

func report(w io.Writer, returnValue int) {
	if returnValue != 0 {
		fmt.Fprintf(w, statusFormat, version, creation)
	}
}

func run(cmdlineArgs []string) (returnValue int) {
	returnValue = 1
	startTime := time.Now()
	logrus.WithFields(logrus.Fields{"version": version, "created": creation}).Info("begin execution")
	defer func() {
		logrus.WithFields(logrus.Fields{"duration": time.Since(startTime)}).Info("end execution")
	}()
	if cmd, args, err := subcommands.ProcessCommand(internal.ApplicationDataPath(), cmdlineArgs); err != nil {
		fmt.Fprintln(os.Stderr, err)
		logrus.Error(err)
	} else {
		cmd.Exec(os.Stdout, args)
		returnValue = 0
	}
	return
}

func initEnv(lookup func() []error) bool {
	if errors := lookup(); len(errors) > 0 {
		fmt.Fprintln(os.Stderr, internal.LOG_ENV_ISSUES_DETECTED)
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
	path := filepath.Join(internal.CreateAppSpecificPath(parentDir), logDirName)
	if err := os.MkdirAll(path, 0755); err != nil {
		fmt.Fprintf(os.Stderr, internal.USER_CANNOT_CREATE_DIRECTORY, path, err)
		return false
	}
	logger = internal.ConfigureLogging(path)
	logrus.SetOutput(logger)
	internal.CleanupLogFiles(path)
	return true
}
