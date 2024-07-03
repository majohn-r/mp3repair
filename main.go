/*
Copyright © 2021 Marc Johnson (marc.johnson27591@gmail.com)
*/
//go:generate goversioninfo -icon=mp3repair.ico
package main

import (
	"mp3repair/cmd"
)

var executor = cmd.Execute

func main() {
	elevationControl := cmd.NewElevationControl()
	if !elevationControl.WillRunElevated() {
		elevationControl.ConfigureExit()
		executor()
	}
}
