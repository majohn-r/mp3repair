/*
Copyright © 2021 Marc Johnson (marc.johnson27591@gmail.com)
*/
//go:generate goversioninfo -icon=mp3repair.ico
package main

import (
	"fmt"
	"mp3repair/cmd"
	"os"

	cmdtoolkit "github.com/majohn-r/cmd-toolkit"
)

var (
	executor = cmd.Execute
	scanf    = fmt.Scanf
)

const selfPromotionMarker = "autoPromoted"

func main() {
	var selfPromoted bool
	os.Args, selfPromoted = isSelfPromoted(os.Args)
	elevationControl := cmdtoolkit.NewElevationControlWithEnvVar(
		cmd.ElevatedPrivilegesPermissionVar,
		cmd.DefaultElevatedPrivilegesPermission,
	)
	// add self-promoted flag to os.Args
	originalArgs := os.Args
	os.Args = injectSelfPromotion(os.Args)
	if !elevationControl.WillRunElevated() {
		configureExit(selfPromoted)
		os.Args = originalArgs
		executor()
	}
}

func injectSelfPromotion(args []string) []string {
	if len(args) == 0 {
		return args // no sense trying
	}
	newArgs := make([]string, 0, 1+len(args))
	newArgs = append(newArgs, args[0], selfPromotionMarker)
	if len(args) > 1 {
		newArgs = append(newArgs, args[1:]...)
	}
	return newArgs
}

func isSelfPromoted(args []string) ([]string, bool) {
	if len(args) > 1 && args[1] == selfPromotionMarker {
		newArgs := make([]string, 0, len(args)-1)
		newArgs = append(newArgs, args[0])
		if len(args) > 2 {
			newArgs = append(newArgs, args[2:]...)
		}
		return newArgs, true
	}
	return args, false
}

func configureExit(selfPromoted bool) {
	if selfPromoted {
		originalExit := cmd.Exit
		cmd.Exit = func(code int) {
			fmt.Printf("Exiting with exit code %d\n", code)
			var name string
			fmt.Printf("Press enter to close the window...\n")
			_, _ = scanf("%s", &name)
			originalExit(code)
		}
	}
}
