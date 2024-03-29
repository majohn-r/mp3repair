/*
Copyright © 2021 Marc Johnson (marc.johnson27591@gmail.com)
*/
package cmd

import (
	"fmt"
	"strings"

	cmd_toolkit "github.com/majohn-r/cmd-toolkit"
	"github.com/majohn-r/output"
	"github.com/spf13/cobra"
)

const (
	aboutCommand = "about"
	appName      = "mp3repair" // the name of the application
	author       = "Marc Johnson"
	firstYear    = 2021 // the year when development of this application began
)

var (
	// semantic version
	Version = "unknown version!"
	// build timestamp in RFC3339 format (2006-01-02T15:04:05Z07:00)
	Creation string
	// AboutCmd represents the about command
	AboutCmd = &cobra.Command{
		Use:   aboutCommand,
		Short: "Provides information about the " + appName + " program",
		Long: fmt.Sprintf("%q", aboutCommand) +
			` provides the following information about the ` + appName + ` program:

* The program version and build timestamp
* Copyright information
* Build information:
  * The version of go used to compile the code
  * A list of dependencies and their versions
* The directory where log files are written
* The full path of the application configuration file and whether it exists`,
		RunE: AboutRun,
	}
)

func AboutRun(_ *cobra.Command, _ []string) error {
	o := BusGetter()
	LogCommandStart(o, aboutCommand, map[string]any{})
	o.WriteConsole(strings.Join(cmd_toolkit.FlowerBox(GatherOutput(o)), "\n"))
	return nil
}

func GatherOutput(o output.Bus) []string {
	goVersion, buildDependencies := InterpretBuildData()
	// 9: 1 each for
	// - app name
	// - copyright
	// - build information header
	// - go version
	// - log file location
	// - configuration file status
	// - and up to 3 for elevation status
	lines := make([]string, 0, 9+len(buildDependencies))
	lines = append(lines,
		cmd_toolkit.DecoratedAppName(appName, Version, Creation),
		cmd_toolkit.Copyright(o, firstYear, Creation, author),
		cmd_toolkit.BuildInformationHeader(),
		cmd_toolkit.FormatGoVersion(goVersion))
	lines = append(lines, cmd_toolkit.FormatBuildDependencies(buildDependencies)...)
	lines = append(lines, fmt.Sprintf("Log files are written to %s", LogPath()))
	if path, exists := configFile(); exists {
		lines = append(lines, fmt.Sprintf("Configuration file %s exists", path))
	} else {
		lines = append(lines, fmt.Sprintf("Configuration file %s does not yet exist", path))
	}
	elevationData := NewElevationControl().Status(appName)
	lines = append(lines, elevationData[0])
	if len(elevationData) > 1 {
		for _, s := range elevationData[1:] {
			lines = append(lines, fmt.Sprintf(" - %s", s))
		}
	}
	return lines
}

func InitializeAbout(version, creation string) {
	Version = version
	Creation = creation
	InitBuildData(Version, Creation)
	SetFirstYear(firstYear)
}

func init() {
	RootCmd.AddCommand(AboutCmd)
}
