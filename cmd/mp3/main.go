package main

import (
	"fmt"
	"mp3/internal"
	"mp3/internal/commands"
	"os"
	"runtime/debug"
	"time"
)

// these variables' values are injected by the mage build
var (
	version  = "unknown version!" // semantic version; read by the mage build from version.txt
	creation string               // build timestamp in RFC3339 format (2006-01-02T15:04:05Z07:00)
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
	fkDependencies         = "dependencies"
	statusFormat           = "%q version %s, created at %s, failed"
)

func exec(logInit func(internal.OutputBus) bool, cmdLine []string) (returnValue int) {
	returnValue = 1
	o := internal.NewOutputBus()
	if logInit(o) {
		returnValue = run(o, debug.ReadBuildInfo, cmdLine)
	}
	report(o, returnValue)
	return
}

func report(o internal.OutputBus, returnValue int) {
	if returnValue != 0 {
		o.WriteError(statusFormat, internal.AppName, version, creation)
	}
}

func run(o internal.OutputBus, f func() (*debug.BuildInfo, bool), cmdlineArgs []string) (returnValue int) {
	returnValue = 1
	startTime := time.Now()
	// initialize about command
	commands.AboutBuildData = createBuildData(f)
	commands.AboutSettings = commands.AboutData{
		AppVersion:     version,
		BuildTimestamp: creation,
	}
	logBegin(o, commands.AboutBuildData.GoVersion, commands.AboutBuildData.Dependencies, cmdlineArgs)
	if cmd, args, ok := commands.ProcessCommand(o, cmdlineArgs); ok {
		if cmd.Exec(o, args) {
			returnValue = 0
		}
	}
	o.LogWriter().Info(internal.LogInfoEndExecution, map[string]any{
		fkDuration: time.Since(startTime),
		fkExitCode: returnValue,
	})
	return
}

func logBegin(o internal.OutputBus, goVersion string, dependencies []string, cmdLineArgs []string) {
	o.LogWriter().Info(internal.LogInfoBeginExecution, map[string]any{
		fkVersion:              version,
		fkTimeStamp:            creation,
		fkGoVersion:            goVersion,
		fkDependencies:         dependencies,
		fkCommandLineArguments: cmdLineArgs,
	})
}

func createBuildData(f func() (*debug.BuildInfo, bool)) *commands.BuildData {
	bD := &commands.BuildData{}
	if b, ok := f(); ok {
		bD.GoVersion = b.GoVersion
		for _, d := range b.Deps {
			bD.Dependencies = append(bD.Dependencies, fmt.Sprintf("%s %s", d.Path, d.Version))
		}
	} else {
		bD.GoVersion = "unknown"
	}
	return bD
}
