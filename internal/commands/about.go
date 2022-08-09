package commands

import (
	"flag"
	"fmt"
	"mp3/internal"
	"runtime/debug"
	"time"
)

type AboutData struct {
	AppVersion     string
	BuildTimestamp string
}

var AboutSettings AboutData

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

func (v *aboutCmd) Exec(o internal.OutputBus, args []string) (ok bool) {
	o.LogWriter().Info(internal.LI_EXECUTING_COMMAND, map[string]interface{}{fkCommandName: v.name()})
	var elements []string
	description := fmt.Sprintf("%s version %s, created on %s", internal.AppName, AboutSettings.AppVersion, AboutSettings.BuildTimestamp)
	elements = append(elements, description)
	lastYear := finalYear(o, AboutSettings.BuildTimestamp)
	copyright := formatCopyright(firstYear, lastYear)
	elements = append(elements, copyright)
	elements = append(elements, "Build Information")
	b := formatBuildData(debug.ReadBuildInfo)
	elements = append(elements, b...)
	reportAbout(o, elements)
	return true
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

func formatBuildData(f func() (*debug.BuildInfo, bool)) []string {
	var s []string
	if b, ok := f(); ok {
		goVersion := b.GoVersion
		s = append(s, fmt.Sprintf(" - Go version: %s", goVersion))
		dependencies := b.Deps
		for _, dep := range dependencies {
			s = append(s, fmt.Sprintf(" - Dependency: %s %s", dep.Path, dep.Version))
		}
	} else {
		s = append(s, " - None available")
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
