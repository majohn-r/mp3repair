package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"

	"github.com/majohn-r/output"
)

const (
	adminPermissionVar        = "MP3REPAIR_RUNS_AS_ADMIN"
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

func redirectedDescriptor(fd uintptr) bool {
	if !IsTerminal(fd) && !IsCygwinTerminal(fd) {
		return true
	}
	return false
}

func stderrState() bool {
	return redirectedDescriptor(os.Stderr.Fd())
}

func stdinState() bool {
	return redirectedDescriptor(os.Stdin.Fd())
}

func stdoutState() bool {
	return redirectedDescriptor(os.Stdout.Fd())
}

func processIsElevated() bool {
	t := GetCurrentProcessToken()
	return IsElevated(t)
}

type ElevationControl struct {
	adminPermitted   bool
	elevated         bool
	stderrRedirected bool
	stdinRedirected  bool
	stdoutRedirected bool
}

func NewElevationControl() *ElevationControl {
	return &ElevationControl{
		adminPermitted:   environmentPermits(),
		elevated:         processIsElevated(),
		stderrRedirected: stderrState(),
		stdinRedirected:  stdinState(),
		stdoutRedirected: stdoutState(),
	}
}

func (ec *ElevationControl) redirected() bool {
	return ec.stderrRedirected || ec.stdinRedirected || ec.stdoutRedirected
}

func (ec *ElevationControl) canElevate() bool {
	if ec.elevated {
		return false // already there, so, no
	}
	if ec.redirected() {
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

func runElevated() (status bool) {
	verb := "runas"
	exe, _ := os.Executable()
	cwd, _ := os.Getwd()
	args := mergeArguments(os.Args)
	verbPtr, _ := syscall.UTF16PtrFromString(verb)
	exePtr, _ := syscall.UTF16PtrFromString(exe)
	cwdPtr, _ := syscall.UTF16PtrFromString(cwd)
	argPtr, _ := syscall.UTF16PtrFromString(args)
	var showCmd int32 = syscall.SW_NORMAL
	// https://github.com/majohn-r/mp3repair/issues/157 if ShellExecute returns
	// no error, assume the user accepted admin privileges and return true
	// status
	if err := ShellExecute(0, verbPtr, exePtr, argPtr, cwdPtr, showCmd); err == nil {
		status = true
	}
	return
}

func (ec *ElevationControl) WillRunElevated() bool {
	if ec.canElevate() {
		// https://github.com/majohn-r/mp3repair/issues/157 if privileges can be
		// elevated successfully, return true, else assume user declined and
		// return false.
		return runElevated()
	}
	return false
}

func (ec *ElevationControl) Log(o output.Bus, level output.Level) {
	o.Log(level, "elevation state", map[string]any{
		"elevated":          ec.elevated,
		"admin_permission":  ec.adminPermitted,
		"stderr_redirected": ec.stderrRedirected,
		"stdin_redirected":  ec.stdinRedirected,
		"stdout_redirected": ec.stdoutRedirected,
	})
}

func (ec *ElevationControl) Status(appName string) []string {
	results := make([]string, 0, 3)
	if ec.elevated {
		results = append(results, fmt.Sprintf("%s is running with elevated privileges", appName))
	} else {
		results = append(results, fmt.Sprintf("%s is not running with elevated privileges", appName))
		if ec.redirected() {
			redirectedIO := make([]string, 0, 3)
			if ec.stderrRedirected {
				redirectedIO = append(redirectedIO, "stderr")
			}
			if ec.stdinRedirected {
				redirectedIO = append(redirectedIO, "stdin")
			}
			if ec.stdoutRedirected {
				redirectedIO = append(redirectedIO, "stdout")
			}
			switch len(redirectedIO) {
			case 1:
				results = append(results, fmt.Sprintf("%s has been redirected", redirectedIO[0]))
			case 2:
				results = append(results, fmt.Sprintf("%s have been redirected", strings.Join(redirectedIO, " and ")))
			case 3:
				results = append(results, "stderr, stdin, and stdout have been redirected")
			}
		}
		if !ec.adminPermitted {
			results = append(results, fmt.Sprintf("The environment variable %s evaluates as false", adminPermissionVar))
		}
	}
	return results
}
