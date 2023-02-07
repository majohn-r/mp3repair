package main

import (
	"mp3/internal/commands"
	"os"
	"runtime/debug"

	tools "github.com/majohn-r/cmd-toolkit"
	"github.com/majohn-r/output"
)

// these variables' values are injected by the build process - do not rename them!
var (
	// semantic version; the build reads this from build/version.txt
	version = "unknown version!"
	// build timestamp in RFC3339 format (2006-01-02T15:04:05Z07:00)
	creation       string
	defaultCommand                                                                                                                = "list"
	execFunc       func(output.Bus, func(output.Bus) bool, func() (*debug.BuildInfo, bool), string, string, string, []string) int = tools.Execute
	exitFunc       func(int)                                                                                                      = os.Exit
)

func main() {
	commands.DeclareDefault(defaultCommand)
	exitCode := execFunc(output.NewDefaultBus(tools.ProductionLogger{}), tools.InitLogging, debug.ReadBuildInfo, "mp3", version, creation, os.Args)
	exitFunc(exitCode)
}
