package commands

import (
	"flag"
	"fmt"
	"mp3/internal"
	"runtime/debug"
	"time"

	"github.com/majohn-r/output"
)

func init() {
	addCommandData(aboutCommandName, commandData{isDefault: false, initFunction: newAboutCmd})
}

var (
	// these are set by InitBuildData
	appVersion        string
	buildTimestamp    string
	goVersion         string
	buildDependencies []string
)

// GoVersion returns the version of Go used to compile the program
func GoVersion() string {
	return goVersion
}

// BuildDependencies returns information about the dependencies used to compile
// the program
func BuildDependencies() []string {
	return buildDependencies
}

const (
	aboutCommandName = "about"

	author    = "Marc Johnson"
	firstYear = 2021 // the year that development began
)

// InitBuildData captures information about how the program was compiled, the
// version of the program, and the timestamp for when the program was built.
func InitBuildData(f func() (*debug.BuildInfo, bool), version, creation string) {
	if b, ok := f(); ok {
		goVersion = b.GoVersion
		for _, d := range b.Deps {
			buildDependencies = append(buildDependencies, fmt.Sprintf("%s %s", d.Path, d.Version))
		}
	} else {
		goVersion = "unknown"
	}
	appVersion = version
	buildTimestamp = creation
}

type aboutCmd struct {
}

func newAboutCmd(o output.Bus, _ *internal.Configuration, _ *flag.FlagSet) (CommandProcessor, bool) {
	return &aboutCmd{}, true
}

// Exec runs the command. The args parameter is ignored, and the methid always
// returns true.
func (a *aboutCmd) Exec(o output.Bus, args []string) (ok bool) {
	logStart(o, aboutCommandName, map[string]any{})
	var s []string
	s = append(s,
		fmt.Sprintf("%s version %s, built on %s", internal.AppName, appVersion, translateTimestamp(buildTimestamp)),
		formatCopyright(firstYear, finalYear(o, buildTimestamp)))
	s = append(s, formatBuildData()...)
	reportAbout(o, s)
	return true
}

func translateTimestamp(s string) string {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return s
	}
	return t.Format("Monday, January 2 2006, 15:04:05 MST")
}

func reportAbout(o output.Bus, lines []string) {
	max := 0
	for _, s := range lines {
		if len(s) > max {
			max = len([]rune(s))
		}
	}
	var formattedLines []string
	for _, s := range lines {
		b := make([]rune, max)
		i := 0
		for _, s1 := range s {
			b[i] = s1
			i++
		}
		for ; i < max; i++ {
			b[i] = ' '
		}
		formattedLines = append(formattedLines, string(b))
	}
	verticalLine := make([]rune, max)
	for i := 0; i < max; i++ {
		verticalLine[i] = '-'
	}
	header := string(verticalLine)
	o.WriteConsole("+-%s-+\n", header)
	for _, s := range formattedLines {
		o.WriteConsole("| %s |\n", s)
	}
	o.WriteConsole("+-%s-+\n", header)
}

func formatBuildData() []string {
	var s []string
	s = append(s, "Build Information", fmt.Sprintf(" - Go version: %s", goVersion))
	for _, dep := range buildDependencies {
		s = append(s, fmt.Sprintf(" - Dependency: %s", dep))
	}
	return s
}

func formatCopyright(firstYear, lastYear int) string {
	if lastYear <= firstYear {
		return fmt.Sprintf("Copyright © %d %s", firstYear, author)
	}
	return fmt.Sprintf("Copyright © %d-%d %s", firstYear, lastYear, author)

}

func finalYear(o output.Bus, timestamp string) int {
	var y = firstYear
	if t, err := time.Parse(time.RFC3339, timestamp); err != nil {
		o.WriteCanonicalError("The build time %q cannot be parsed: %v", timestamp, err)
		o.Log(output.Error, "parse error", map[string]any{
			"error": err,
			"value": timestamp,
		})
	} else {
		y = t.Year()
	}
	return y
}
