package main

import (
	"fmt"
	"mp3/internal"
	"mp3/internal/commands"
	"os"
	"time"
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
	o := internal.NewOutputDevice(os.Stdout, os.Stderr)
	if internal.InitLogging(o) {
		returnValue = run(o, os.Args)
	}
	report(o, returnValue)
	os.Exit(returnValue)
}

const (
	fkCommandLineArguments = "args"
	fkDuration             = "duration"
	fkExitCode             = "exitCode"
	fkTimeStamp            = "timeStamp"
	fkVersion              = "version"
	statusFormat           = "%s version %s, created at %s, failed\n"
)

func report(o internal.OutputBus, returnValue int) {
	if returnValue != 0 {
		fmt.Fprintf(o.ErrorWriter(), statusFormat, internal.AppName, version, creation)
	}
}

func run(o internal.OutputBus, cmdlineArgs []string) (returnValue int) {
	returnValue = 1
	startTime := time.Now()
	o.LogWriter().Info(internal.LI_BEGIN_EXECUTION, map[string]interface{}{
		fkVersion:              version,
		fkTimeStamp:            creation,
		fkCommandLineArguments: cmdlineArgs,
	})
	if cmd, args, ok := commands.ProcessCommand(o, cmdlineArgs); ok {
		if cmd.Exec(o, args) {
			returnValue = 0
		}
	}
	o.LogWriter().Info(internal.LI_END_EXECUTION, map[string]interface{}{
		fkDuration: time.Since(startTime),
		fkExitCode: returnValue,
	})
	return
}
