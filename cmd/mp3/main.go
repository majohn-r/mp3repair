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
	if initEnv(os.Stderr, internal.LookupEnvVars) {
		if initLogging(os.Stderr, internal.TemporaryFileFolder()) {
			returnValue = run(os.Args)
		}
	}
	report(os.Stderr, returnValue)
	os.Exit(returnValue)
}

const (
	fkCommandLineArguments = "args"
	fkDuration             = "duration"
	fkExitCode             = "exitCode"
	fkTimeStamp            = "timeStamp"
	fkVersion              = "version"
	logDirName             = "logs"
	statusFormat           = "%s version %s, created at %s, failed\n"
)

func report(w io.Writer, returnValue int) {
	if returnValue != 0 {
		fmt.Fprintf(w, statusFormat, internal.AppName, version, creation)
	}
}

func run(cmdlineArgs []string) (returnValue int) {
	returnValue = 1
	startTime := time.Now()
	logrus.WithFields(logrus.Fields{
		fkVersion:              version,
		fkTimeStamp:            creation,
		fkCommandLineArguments: cmdlineArgs,
	}).Info(internal.LI_BEGIN_EXECUTION)
	if cmd, args, ok := subcommands.ProcessCommand(os.Stderr, internal.ApplicationDataPath(), cmdlineArgs); ok {
		if cmd.Exec(os.Stdout, os.Stderr, args) {
			returnValue = 0
		}
	}
	logrus.WithFields(logrus.Fields{
		fkDuration: time.Since(startTime),
		fkExitCode: returnValue,
	}).Info(internal.LI_END_EXECUTION)
	return
}

func initEnv(w io.Writer, lookup func(w io.Writer) bool) bool {
	return lookup(w)
}

// exposed so that unit tests can close the writer!
var logger *cronowriter.CronoWriter

func initLogging(w io.Writer, parentDir string) bool {
	path := filepath.Join(internal.CreateAppSpecificPath(parentDir), logDirName)
	if err := os.MkdirAll(path, 0755); err != nil {
		fmt.Fprintf(w, internal.USER_CANNOT_CREATE_DIRECTORY, path, err)
		return false
	}
	logger = internal.ConfigureLogging(path)
	logrus.SetOutput(logger)
	internal.CleanupLogFiles(w, path)
	return true
}
