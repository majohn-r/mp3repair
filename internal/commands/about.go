package commands

import (
	"flag"
	"fmt"
	"mp3/internal"
	"time"
)

// AboutData contains data about the image itself.
type AboutData struct {
	// AppVersion is the semantic version of the application.
	AppVersion string
	// BuildTimeStamp is the time when the application was built and is
	// formatted per time.RFC3339.
	BuildTimestamp string
}

// BuildData contains data about how the image was built.
type BuildData struct {
	// GoVersion contains the version of go used to build the image.
	GoVersion string
	// Dependencies contains information about the modules used to build the
	// image. Each element in the slice consists of a module's path and its
	// version.
	Dependencies []string
}

var (
	// AboutSettings is set by main's run method.
	AboutSettings AboutData
	// AboutBuildData is set by main's run method.
	AboutBuildData *BuildData
)

const (
	author    = "Marc Johnson"
	firstYear = 2021 // the year that development began
)

type aboutCmd struct {
	n string // command name, probably "about"
}

func (v *aboutCmd) name() string {
	return v.n
}

func newAboutCmd(o internal.OutputBus, c *internal.Configuration, fSet *flag.FlagSet) (CommandProcessor, bool) {
	return &aboutCmd{n: fSet.Name()}, true
}

// Exec runs the command. The args parameter is ignored, and the methid always
// returns true.
func (v *aboutCmd) Exec(o internal.OutputBus, args []string) (ok bool) {
	o.LogWriter().Info(internal.LI_EXECUTING_COMMAND, map[string]interface{}{fkCommandName: v.name()})
	var elements []string
	timeStamp := translateTimestamp(AboutSettings.BuildTimestamp)
	description := fmt.Sprintf("%s version %s, built on %s", internal.AppName, AboutSettings.AppVersion, timeStamp)
	elements = append(elements, description)
	lastYear := finalYear(o, AboutSettings.BuildTimestamp)
	copyright := formatCopyright(firstYear, lastYear)
	elements = append(elements, copyright)
	elements = append(elements, "Build Information")
	b := formatBuildData(AboutBuildData)
	elements = append(elements, b...)
	reportAbout(o, elements)
	return true
}

func translateTimestamp(t string) string {
	rT, err := time.Parse(time.RFC3339, t)
	if err != nil {
		return t
	}
	return rT.Format("Monday, January 2 2006, 15:04:05 MST")
}

func reportAbout(o internal.OutputBus, data []string) {
	maxLineLength := 0
	for _, s := range data {
		if len(s) > maxLineLength {
			maxLineLength = len([]rune(s))
		}
	}
	var formattedData []string
	for _, s := range data {
		b := make([]rune, maxLineLength)
		i := 0
		for _, s1 := range s {
			b[i] = s1
			i++
		}
		for ; i < maxLineLength; i++ {
			b[i] = ' '
		}
		formattedData = append(formattedData, string(b))
	}
	bHeader := make([]rune, maxLineLength)
	for i := 0; i < maxLineLength; i++ {
		bHeader[i] = '-'
	}
	header := string(bHeader)
	o.WriteConsole(false, "+-%s-+\n", header)
	for _, s := range formattedData {
		o.WriteConsole(false, "| %s |\n", s)
	}
	o.WriteConsole(false, "+-%s-+\n", header)
}

func formatBuildData(bD *BuildData) []string {
	var s []string
	s = append(s, fmt.Sprintf(" - Go version: %s", bD.GoVersion))
	for _, dep := range bD.Dependencies {
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

func finalYear(o internal.OutputBus, timestamp string) int {
	var y = firstYear
	if t, err := time.Parse(time.RFC3339, timestamp); err != nil {
		o.WriteError(internal.USER_CANNOT_PARSE_TIMESTAMP, timestamp, err)
		o.LogWriter().Error("parse error", map[string]interface{}{
			internal.FK_ERROR: err,
			internal.FK_VALUE: timestamp,
		})
	} else {
		y = t.Year()
	}
	return y
}
