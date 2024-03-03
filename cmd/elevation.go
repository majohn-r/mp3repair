package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
)

const (
	adminPermissionVar        = "MP3_RUNS_AS_ADMIN"
	defaultEnvironmentSetting = true
)

func environmentPermits() bool {
	if value, ok := LookupEnv(adminPermissionVar); ok {
		// interpret value as bool
		if boolValue, err := strconv.ParseBool(value); err != nil {
			fmt.Printf("Value %q of environment variable %q is neither true nor false\n", value, adminPermissionVar)
		} else {
			return boolValue
		}
	}
	return defaultEnvironmentSetting
}

func redirectedIO() bool {
	for _, device := range []*os.File{os.Stdin, os.Stderr, os.Stdout} {
		fd := device.Fd()
		if !IsTerminal(fd) && !IsCygwinTerminal(fd) {
			return true
		}
	}
	return false
}

func processIsElevated() bool {
	t := GetCurrentProcessToken()
	return IsElevated(t)
}

type ElevationControl struct {
	adminPermitted bool
	elevated       bool
	redirected     bool
}

func NewElevationControl() *ElevationControl {
	return &ElevationControl{
		adminPermitted: environmentPermits(),
		elevated:       processIsElevated(),
		redirected:     redirectedIO(),
	}
}

func (ec *ElevationControl) canElevate() bool {
	if ec.elevated {
		return false // already there, so, no
	}
	if ec.redirected {
		return false // redirection will be lost, so, no
	}
	return ec.adminPermitted // ok, obey the environment variable then
}

func (ec *ElevationControl) ConfigureExit() {
	if ec.elevated {
		originalExit := Exit
		Exit = func(code int) {
			fmt.Printf("Exiting with exit code %d\n", code)
			var name string
			fmt.Printf("Press enter to close the window...\n")
			Scanf("%s", &name)
			originalExit(code)
		}
	}
}

func mergeArguments(args []string) string {
	merged := ""
	if len(args) > 1 {
		merged = strings.Join(args[1:], " ")
	}
	return merged
}

// credit: https://gist.github.com/jerblack/d0eb182cc5a1c1d92d92a4c4fcc416c6

func runElevated() {
	verb := "runas"
	exe, _ := os.Executable()
	cwd, _ := os.Getwd()
	args := mergeArguments(os.Args)
	verbPtr, _ := syscall.UTF16PtrFromString(verb)
	exePtr, _ := syscall.UTF16PtrFromString(exe)
	cwdPtr, _ := syscall.UTF16PtrFromString(cwd)
	argPtr, _ := syscall.UTF16PtrFromString(args)
	var showCmd int32 = syscall.SW_NORMAL
	ShellExecute(0, verbPtr, exePtr, argPtr, cwdPtr, showCmd)
}

func (ec *ElevationControl) WillRunElevated() bool {
	if ec.canElevate() {
		runElevated()
		return true
	}
	return false
}

func (ec *ElevationControl) Status(appName string) []string {
	results := []string{}
	if ec.elevated {
		results = append(results, fmt.Sprintf("%s is running with elevated privileges", appName))
	} else {
		results = append(results, fmt.Sprintf("%s is not running with elevated privileges", appName))
		if ec.redirected {
			results = append(results, "At least one of stdin, stdout, and stderr has been redirected")
		}
		if !ec.adminPermitted {
			results = append(results, fmt.Sprintf("The environment variable %s evaluates as false", adminPermissionVar))
		}
	}
	return results
}
