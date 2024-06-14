package cmd

import (
	"fmt"
	"runtime/debug"
	"strings"

	cmdtoolkit "github.com/majohn-r/cmd-toolkit"
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
	// Version is the application's semantic version
	Version = "unknown version!"
	// Creation is the application's build timestamp in RFC3339 format (2006-01-02T15:04:05Z07:00)
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
	CachedGoVersion         string
	CachedBuildDependencies []string
)

func AboutRun(_ *cobra.Command, _ []string) error {
	o := BusGetter()
	LogCommandStart(o, aboutCommand, map[string]any{})
	o.WriteConsole(strings.Join(cmdtoolkit.FlowerBox(AcquireAboutData(o)), "\n"))
	return nil
}

func AcquireAboutData(o output.Bus) []string {
	CachedGoVersion, CachedBuildDependencies = InterpretBuildData(debug.ReadBuildInfo)
	// 9: 1 each for
	// - app name
	// - copyright
	// - build information header
	// - go version
	// - log file location
	// - configuration file status
	// - and up to 3 for elevation status
	lines := make([]string, 0, 9+len(CachedBuildDependencies))
	lines = append(lines,
		cmdtoolkit.DecoratedAppName(appName, Version, Creation),
		cmdtoolkit.Copyright(o, firstYear, Creation, author),
		"Build Information",
		cmdtoolkit.FormatGoVersion(CachedGoVersion))
	lines = append(lines, cmdtoolkit.FormatBuildDependencies(CachedBuildDependencies)...)
	lines = append(lines, fmt.Sprintf("Log files are written to %s", LogPath()))
	path, exists := configFile()
	switch {
	case exists:
		lines = append(lines, fmt.Sprintf("Configuration file %s exists", path))
	default:
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

type AboutMaker struct {
	SoftwareVersion string
	CreationDate    string
}

func (maker AboutMaker) InitializeAbout() {
	Version = maker.SoftwareVersion
	Creation = maker.CreationDate
}

func init() {
	RootCmd.AddCommand(AboutCmd)
}
