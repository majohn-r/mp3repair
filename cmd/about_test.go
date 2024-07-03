/*
Copyright © 2021 Marc Johnson (marc.johnson27591@gmail.com)
*/
package cmd_test

import (
	"mp3repair/cmd"
	"reflect"
	"runtime/debug"
	"testing"

	"github.com/majohn-r/output"
	"github.com/spf13/cobra"
	"golang.org/x/sys/windows"
)

func TestAboutRun(t *testing.T) {
	originalBusGetter := cmd.BusGetter
	originalLogCommandStart := cmd.LogCommandStart
	originalInterpretBuildData := cmd.InterpretBuildData
	originalLogPath := cmd.LogPath
	originalVersion := cmd.Version
	originalCreation := cmd.Creation
	originalApplicationPath := cmd.ApplicationPath
	originalPlainFileExists := cmd.PlainFileExists
	originalIsElevatedFunc := cmd.IsElevated
	defer func() {
		cmd.BusGetter = originalBusGetter
		cmd.LogCommandStart = originalLogCommandStart
		cmd.InterpretBuildData = originalInterpretBuildData
		cmd.LogPath = originalLogPath
		cmd.Version = originalVersion
		cmd.Creation = originalCreation
		cmd.ApplicationPath = originalApplicationPath
		cmd.PlainFileExists = originalPlainFileExists
		cmd.IsElevated = originalIsElevatedFunc
	}()
	cmd.InterpretBuildData = func(func() (*debug.BuildInfo, bool)) (string, []string) {
		return "go1.22.x", []string{
			"go.dependency.1 v1.2.3",
			"go.dependency.2 v1.3.4",
			"go.dependency.3 v0.1.2",
		}
	}
	cmd.LogPath = func() string {
		return "/my/files/tmp/logs/mp3repair"
	}
	cmd.Version = "0.40.0"
	cmd.Creation = "2024-02-24T13:14:05-05:00"
	cmd.ApplicationPath = func() string {
		return "/my/files/apppath"
	}
	cmd.PlainFileExists = func(_ string) bool { return true }
	cmd.IsElevated = func(_ windows.Token) bool {
		return true
	}
	type args struct {
		in0 *cobra.Command
		in1 []string
	}
	tests := map[string]struct {
		args
		output.WantedRecording
	}{
		"simple": {
			WantedRecording: output.WantedRecording{
				Console: "" +
					"+-------------------------------------------------------------------------------+\n" +
					"| mp3repair version 0.40.0, built on Saturday, February 24 2024, 13:14:05 -0500 |\n" +
					"| Copyright © 2021-2024 Marc Johnson                                            |\n" +
					"| Build Information                                                             |\n" +
					"|  - Go version: go1.22.x                                                       |\n" +
					"|  - Dependency: go.dependency.1 v1.2.3                                         |\n" +
					"|  - Dependency: go.dependency.2 v1.3.4                                         |\n" +
					"|  - Dependency: go.dependency.3 v0.1.2                                         |\n" +
					"| Log files are written to /my/files/tmp/logs/mp3repair                         |\n" +
					"| Configuration file \\my\\files\\apppath\\defaults.yaml exists                     |\n" +
					"| mp3repair is running with elevated privileges                                 |\n" +
					"+-------------------------------------------------------------------------------+\n",
				Log: "level='info' command='about' msg='executing command'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			cmd.BusGetter = func() output.Bus { return o }
			cmd.LogCommandStart = func(bus output.Bus, cmdName string, args map[string]any) {
				bus.Log(output.Info, "executing command", map[string]any{"command": "about"})
			}
			_ = cmd.AboutRun(tt.args.in0, tt.args.in1)
			o.Report(t, "AboutRun()", tt.WantedRecording)
		})
	}
}

func enableCommandRecording(o *output.Recorder, command *cobra.Command) {
	command.SetErr(o.ErrorWriter())
	command.SetOut(o.ConsoleWriter())
}

func TestAboutHelp(t *testing.T) {
	tests := map[string]struct {
		output.WantedRecording
	}{
		"good": {
			WantedRecording: output.WantedRecording{
				Console: "" +
					"\"about\" provides the following information about the mp3repair program:\n" +
					"\n" +
					"* The program version and build timestamp\n" +
					"* Copyright information\n" +
					"* Build information:\n" +
					"  * The version of go used to compile the code\n" +
					"  * A list of dependencies and their versions\n" +
					"* The directory where log files are written\n" +
					"* The full path of the application configuration file and whether it exists\n" +
					"\n" +
					"Usage:\n" +
					"  mp3repair about\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			command := cmd.AboutCmd
			enableCommandRecording(o, command)
			_ = command.Help()
			o.Report(t, "about Help()", tt.WantedRecording)
		})
	}
}

func TestAcquireAboutData(t *testing.T) {
	originalInterpretBuildData := cmd.InterpretBuildData
	originalLogPath := cmd.LogPath
	originalVersion := cmd.Version
	originalCreation := cmd.Creation
	originalApplicationPath := cmd.ApplicationPath
	originalPlainFileExists := cmd.PlainFileExists
	originalGetCurrentProcessToken := cmd.GetCurrentProcessToken
	originalIsElevatedFunc := cmd.IsElevated
	originalIsTerminal := cmd.IsTerminal
	originalIsCygwinTerminal := cmd.IsCygwinTerminal
	originalLookupEnv := cmd.LookupEnv
	defer func() {
		cmd.InterpretBuildData = originalInterpretBuildData
		cmd.LogPath = originalLogPath
		cmd.Version = originalVersion
		cmd.Creation = originalCreation
		cmd.ApplicationPath = originalApplicationPath
		cmd.PlainFileExists = originalPlainFileExists
		cmd.GetCurrentProcessToken = originalGetCurrentProcessToken
		cmd.IsElevated = originalIsElevatedFunc
		cmd.IsTerminal = originalIsTerminal
		cmd.IsCygwinTerminal = originalIsCygwinTerminal
		cmd.LookupEnv = originalLookupEnv
	}()
	cmd.InterpretBuildData = func(func() (*debug.BuildInfo, bool)) (string, []string) {
		return "go1.22.x", []string{
			"go.dependency.1 v1.2.3",
			"go.dependency.2 v1.3.4",
			"go.dependency.3 v0.1.2",
		}
	}
	cmd.LogPath = func() string {
		return "/my/files/tmp/logs/mp3repair"
	}
	cmd.Version = "0.40.0"
	cmd.Creation = "2024-02-24T13:14:05-05:00"
	cmd.ApplicationPath = func() string {
		return "/my/files/apppath"
	}
	cmd.GetCurrentProcessToken = func() (t windows.Token) {
		return
	}
	tests := map[string]struct {
		plainFileExists      func(string) bool
		forceElevated        bool
		forceRedirection     bool
		forceAdminPermission bool
		want                 []string
	}{
		"with existing config file, elevated": {
			plainFileExists:      func(_ string) bool { return true },
			forceElevated:        true,
			forceRedirection:     false,
			forceAdminPermission: false,
			want: []string{
				"mp3repair version 0.40.0, built on Saturday, February 24 2024, 13:14:05 -0500",
				"Copyright © 2021-2024 Marc Johnson",
				"Build Information",
				" - Go version: go1.22.x",
				" - Dependency: go.dependency.1 v1.2.3",
				" - Dependency: go.dependency.2 v1.3.4",
				" - Dependency: go.dependency.3 v0.1.2",
				"Log files are written to /my/files/tmp/logs/mp3repair",
				"Configuration file \\my\\files\\apppath\\defaults.yaml exists",
				"mp3repair is running with elevated privileges",
			},
		},
		"without existing config file, not elevated, redirected, with admin permission": {
			plainFileExists:      func(_ string) bool { return false },
			forceElevated:        false,
			forceRedirection:     true,
			forceAdminPermission: true,
			want: []string{
				"mp3repair version 0.40.0, built on Saturday, February 24 2024, 13:14:05 -0500",
				"Copyright © 2021-2024 Marc Johnson",
				"Build Information",
				" - Go version: go1.22.x",
				" - Dependency: go.dependency.1 v1.2.3",
				" - Dependency: go.dependency.2 v1.3.4",
				" - Dependency: go.dependency.3 v0.1.2",
				"Log files are written to /my/files/tmp/logs/mp3repair",
				"Configuration file \\my\\files\\apppath\\defaults.yaml does not yet exist",
				"mp3repair is not running with elevated privileges",
				" - stderr, stdin, and stdout have been redirected",
			},
		},
		"without existing config file, not elevated, redirected, no admin permission": {
			plainFileExists:      func(_ string) bool { return false },
			forceElevated:        false,
			forceRedirection:     true,
			forceAdminPermission: false,
			want: []string{
				"mp3repair version 0.40.0, built on Saturday, February 24 2024, 13:14:05 -0500",
				"Copyright © 2021-2024 Marc Johnson",
				"Build Information",
				" - Go version: go1.22.x",
				" - Dependency: go.dependency.1 v1.2.3",
				" - Dependency: go.dependency.2 v1.3.4",
				" - Dependency: go.dependency.3 v0.1.2",
				"Log files are written to /my/files/tmp/logs/mp3repair",
				"Configuration file \\my\\files\\apppath\\defaults.yaml does not yet exist",
				"mp3repair is not running with elevated privileges",
				" - stderr, stdin, and stdout have been redirected",
				" - The environment variable MP3REPAIR_RUNS_AS_ADMIN evaluates as false",
			},
		},
		"without existing config file, not elevated, not redirected, no admin permission": {
			plainFileExists:      func(_ string) bool { return false },
			forceElevated:        false,
			forceRedirection:     false,
			forceAdminPermission: false,
			want: []string{
				"mp3repair version 0.40.0, built on Saturday, February 24 2024, 13:14:05 -0500",
				"Copyright © 2021-2024 Marc Johnson",
				"Build Information",
				" - Go version: go1.22.x",
				" - Dependency: go.dependency.1 v1.2.3",
				" - Dependency: go.dependency.2 v1.3.4",
				" - Dependency: go.dependency.3 v0.1.2",
				"Log files are written to /my/files/tmp/logs/mp3repair",
				"Configuration file \\my\\files\\apppath\\defaults.yaml does not yet exist",
				"mp3repair is not running with elevated privileges",
				" - The environment variable MP3REPAIR_RUNS_AS_ADMIN evaluates as false",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			cmd.PlainFileExists = tt.plainFileExists
			cmd.IsElevated = func(_ windows.Token) bool {
				return tt.forceElevated
			}
			cmd.IsTerminal = func(_ uintptr) bool {
				return !tt.forceRedirection
			}
			cmd.IsCygwinTerminal = func(_ uintptr) bool {
				return !tt.forceRedirection
			}
			if tt.forceAdminPermission {
				cmd.LookupEnv = func(_ string) (string, bool) {
					return "true", true
				}
			} else {
				cmd.LookupEnv = func(_ string) (string, bool) {
					return "false", true
				}
			}
			if got := cmd.AcquireAboutData(output.NewNilBus()); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AcquireAboutData() got %v, want %v", got, tt.want)
			}
		})
	}
}
