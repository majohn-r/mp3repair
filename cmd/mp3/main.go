package main

import (
	"mp3/internal"
	"mp3/internal/commands"
	"os"
	"time"
)

// these variables' values are injected by the mage build
var (
	version   string = "unknown version!"   // semantic version; read by the mage build from version.txt
	creation  string                        // build timestamp in RFC3339 format (2006-01-02T15:04:05Z07:00)
	goVersion string = "unknown go version" // go version; read by the mage build from 'go version'
)

func main() {
	os.Exit(exec(internal.InitLogging, os.Args))
}

const (
	fkCommandLineArguments = "args"
	fkDuration             = "duration"
	fkExitCode             = "exitCode"
	fkTimeStamp            = "timeStamp"
	fkVersion              = "version"
	fkGoVersion            = "go version"
	statusFormat           = "%q version %s, created at %s, with %s, failed"
)

func exec(logInit func(internal.OutputBus) bool, cmdLine []string) (returnValue int) {
	returnValue = 1
	o := internal.NewOutputDevice()
	if logInit(o) {
		returnValue = run(o, cmdLine)
	}
	report(o, returnValue)
	return
}

func report(o internal.OutputBus, returnValue int) {
	if returnValue != 0 {
		o.WriteError(statusFormat, internal.AppName, version, creation, goVersion)
	}
}

func run(o internal.OutputBus, cmdlineArgs []string) (returnValue int) {
	returnValue = 1
	startTime := time.Now()
	o.LogWriter().Info(internal.LI_BEGIN_EXECUTION, map[string]interface{}{
		fkVersion:              version,
		fkTimeStamp:            creation,
		fkGoVersion:            goVersion,
		fkCommandLineArguments: cmdlineArgs,
	})
	// initialize about command
	commands.AboutSettings = commands.AboutData{
		AppVersion:     version,
		BuildTimestamp: creation,
	}
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
