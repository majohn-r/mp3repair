package main

import (
	"mp3/internal/commands"
	"os"

	tools "github.com/majohn-r/cmd-toolkit"
	"github.com/majohn-r/output"
)

// these variables' values are injected by the build process - do not rename them!
var (
	// semantic version; the build reads this from build/version.txt
	version = "unknown version!"
	// build timestamp in RFC3339 format (2006-01-02T15:04:05Z07:00)
	creation string
	// these are variables in order to allow unit testing to inject
	// test-friendly functions
	execFunc func(output.Bus, int, string, string, string, []string) int = tools.Execute
	exitFunc func(int)                                                   = os.Exit
)

const (
	appName        = "mp3"
	defaultCommand = "list"
	firstYear      = 2021
)

func main() {
	commands.DeclareDefault(defaultCommand)
	exitCode := execFunc(output.NewDefaultBus(tools.ProductionLogger{}), firstYear, appName, version, creation, os.Args)
	exitFunc(exitCode)
}
