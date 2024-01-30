/*
Copyright © 2021 Marc Johnson (marc.johnson27591@gmail.com)
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

/*
The **about** command provides information about the **mp3** program, including:

- The program version
- The build timestamp
- Copyright information
- The version of go used to compile the code
- A list of dependencies and their versions
*/

const (
	aboutCommand = "about"
	appName      = "mp3" // the name of the application
	firstYear    = 2021  // the year when development of this application began
)

var (
	Version  = "unknown version!" // semantic version
	Creation string               // build timestamp in RFC3339 format (2006-01-02T15:04:05Z07:00)
	// AboutCmd represents the about command
	AboutCmd = &cobra.Command{
		Use:   aboutCommand,
		Short: "Provides information about the mp3 program",
		Long: fmt.Sprintf("%q", aboutCommand) + ` provides the following information about the mp3 program:

* The program version
* Copyright information
* Build information:
  * The build timestamp
  * The version of go used to compile the code
  * A list of dependencies and their versions`,
		Run: AboutRun,
	}
)

func AboutRun(_ *cobra.Command, _ []string) {
	o := BusGetter()
	LogCommandStart(o, aboutCommand, map[string]any{})
	GenerateAboutContent(o)
	Exit(Success)
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
