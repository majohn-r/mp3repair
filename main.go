/*
Copyright © 2021 Marc Johnson (marc.johnson27591@gmail.com)
*/
//go:generate goversioninfo -icon=mp3repair.ico
package main

import (
	cmdtoolkit "github.com/majohn-r/cmd-toolkit"
	"mp3repair/cmd"
)

var executor = cmd.Execute

func main() {
	elevationControl := cmdtoolkit.NewElevationControlWithEnvVar(
		cmd.ElevatedPrivilegesPermissionVar,
		cmd.DefaultElevatedPrivilegesPermission,
	)
	if !elevationControl.WillRunElevated() {
		cmd.Exit = elevationControl.ConfigureExit(cmd.Exit)
		executor()
	}
}
