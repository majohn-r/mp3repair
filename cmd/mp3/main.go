package main

import (
	"mp3/internal/commands"
	"os"
	"strconv"

	tools "github.com/majohn-r/cmd-toolkit"
	"github.com/majohn-r/output"
)

var (
	// the following variables are set by the build process; the variable names
	// are known to that process, so do not casually change them
	version        = "unknown version!" // semantic version
	creation       string               // build timestamp in RFC3339 format (2006-01-02T15:04:05Z07:00)
	appName        string               // the name of the application
	defaultCommand string               // the name of the default command
	firstYear      string               // the year when development of this application began
	// these are variables in order to allow unit testing to inject
	// test-friendly functions
	execFunc func(output.Bus, int, string, string, string, []string) int = tools.Execute
	exitFunc func(int)                                                   = os.Exit
	bus      output.Bus                                                  = output.NewDefaultBus(tools.ProductionLogger{})
)

func main() {
	exitCode := 1
	if beginningYear, err := strconv.Atoi(firstYear); err != nil {
		bus.WriteCanonicalError("The value of firstYear %q is not valid: %v\n", firstYear, err)
	} else {
		commands.DeclareDefault(defaultCommand)
		exitCode = execFunc(bus, beginningYear, appName, version, creation, os.Args)
	}
	exitFunc(exitCode)
}
