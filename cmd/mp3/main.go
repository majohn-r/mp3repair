package main

import (
	"fmt"
	"mp3/internal"
	"mp3/internal/commands"
	"os"
	"runtime/debug"
	"time"

	"github.com/majohn-r/output"
)

// these variables' values are injected by the build process - do not rename them!
var (
	version  = "unknown version!" // semantic version; the build reads this from build/version.txt
	creation string               // build timestamp in RFC3339 format (2006-01-02T15:04:05Z07:00)
)

func main() {
	os.Exit(exec(internal.InitLogging, os.Args))
}

func exec(logInit func(output.Bus) bool, cmdLine []string) (returnValue int) {
	returnValue = 1
	o := output.NewDefaultBus(internal.ProductionLogger{})
	if logInit(o) {
		returnValue = run(o, debug.ReadBuildInfo, cmdLine)
	}
	report(o, returnValue)
	return
}

func report(o output.Bus, returnValue int) {
	if returnValue != 0 {
		o.WriteCanonicalError("%q version %s, created at %s, failed", internal.AppName, version, creation)
	}
}

func run(o output.Bus, f func() (*debug.BuildInfo, bool), cmdlineArgs []string) (returnValue int) {
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
	o.Log(output.Info, "execution ends", map[string]any{
		"duration": time.Since(startTime),
		"exitCode": returnValue,
	})
	return
}

func logBegin(o output.Bus, goVersion string, dependencies, cmdLineArgs []string) {
	o.Log(output.Info, "execution starts", map[string]any{
		"version":      version,
		"timeStamp":    creation,
		"go version":   goVersion,
		"dependencies": dependencies,
		"args":         cmdLineArgs,
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
