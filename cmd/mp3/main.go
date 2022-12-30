package main

import (
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

func exec(logInit func(output.Bus) bool, cmdLine []string) (exitCode int) {
	exitCode = 1
	o := output.NewDefaultBus(internal.ProductionLogger{})
	if logInit(o) && internal.InitApplicationPath(o) {
		exitCode = run(o, debug.ReadBuildInfo, cmdLine)
	}
	report(o, exitCode)
	return
}

func report(o output.Bus, exitCode int) {
	if exitCode != 0 {
		o.WriteCanonicalError("%q version %s, created at %s, failed", internal.AppName, version, creation)
	}
}

func run(o output.Bus, f func() (*debug.BuildInfo, bool), args []string) (exitCode int) {
	exitCode = 1
	start := time.Now()
	// initialize about command
	commands.InitBuildData(f, version, creation)
	o.Log(output.Info, "execution starts", map[string]any{
		"version":      version,
		"timeStamp":    creation,
		"goVersion":    commands.GoVersion(),
		"dependencies": commands.BuildDependencies(),
		"args":         args,
	})
	if cmd, args, ok := commands.ProcessCommand(o, args); ok {
		if cmd.Exec(o, args) {
			exitCode = 0
		}
	}
	o.Log(output.Info, "execution ends", map[string]any{
		"duration": time.Since(start),
		"exitCode": exitCode,
	})
	return
}
