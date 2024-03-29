/*
Copyright © 2021 Marc Johnson (marc.johnson27591@gmail.com)
*/
//go:generate goversioninfo -icon=mp3repair.ico
package main

import (
	"mp3repair/cmd"
)

var (
	executor = cmd.Execute
	// the following variables are set by the build process; the variable names
	// are known to that process, so do not casually change them
	version  string // semantic version
	creation string // build timestamp in RFC3339 format (2006-01-02T15:04:05Z07:00)
)

func main() {
	elevationControl := cmd.NewElevationControl()
	if !elevationControl.WillRunElevated() {
		elevationControl.ConfigureExit()
		cmd.InitializeAbout(version, creation)
		executor()
	}
}
