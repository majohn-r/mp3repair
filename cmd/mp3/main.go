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
	fieldKeyCommandLineArguments = "args"
	fieldKeyDuration             = "duration"
	fieldKeyExitCode             = "exitCode"
	fieldKeyTImeStamp            = "timeStamp"
	fieldKeyVersion              = "version"
	fieldKeyGoVersion            = "go version"
	fieldKeyDependencies         = "dependencies"
	statusFormat                 = "%q version %s, created at %s, failed"
)

func exec(logInit func(internal.OutputBus) bool, cmdLine []string) (returnValue int) {
	returnValue = 1
	o := internal.NewDefaultOutputBus(internal.ProductionLogger{})
	if logInit(o) {
		returnValue = run(o, debug.ReadBuildInfo, cmdLine)
	}
	report(o, returnValue)
	return
}

func report(o internal.OutputBus, returnValue int) {
	if returnValue != 0 {
		o.WriteCanonicalError(statusFormat, internal.AppName, version, creation)
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
	o.Log(internal.Info, internal.LogInfoEndExecution, map[string]any{
		fieldKeyDuration: time.Since(startTime),
		fieldKeyExitCode: returnValue,
	})
	return
}

func logBegin(o internal.OutputBus, goVersion string, dependencies []string, cmdLineArgs []string) {
	o.Log(internal.Info, internal.LogInfoBeginExecution, map[string]any{
		fieldKeyVersion:              version,
		fieldKeyTImeStamp:            creation,
		fieldKeyGoVersion:            goVersion,
		fieldKeyDependencies:         dependencies,
		fieldKeyCommandLineArguments: cmdLineArgs,
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
