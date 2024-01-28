/*
Copyright © 2021 Marc Johnson (marc.johnson27591@gmail.com)
*/
package cmd

import (
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
	AppName      = "mp3" // the name of the application
	FirstYear    = 2021  // the year when development of this application began
)

var (
	Version  = "unknown version!" // semantic version
	Creation string               // build timestamp in RFC3339 format (2006-01-02T15:04:05Z07:00)
	// aboutCmd represents the about command
	aboutCmd = &cobra.Command{
		Use:                   aboutCommand,
		DisableFlagsInUseLine: true,
		Short:                 "Provides information about the program",
		Long: `Provides the following information about the program:
	
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
	CommandStartLogger(o, aboutCommand, map[string]any{})
	AboutContentGenerator(o)
}

func InitializeAbout(version, creation string) {
	Version = version
	Creation = creation
	InitBuildDataFunc(Version, Creation)
	FirstYearSetFunc(FirstYear)
}

func init() {
	rootCmd.AddCommand(aboutCmd)
}
