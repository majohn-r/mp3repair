package cmd_test

import (
	"mp3repair/cmd"
	"reflect"
	"runtime/debug"
	"testing"

	"github.com/majohn-r/output"
	"github.com/spf13/cobra"
)

type testingElevationControl struct {
	desiredStatus []string
	logFields     map[string]any
}

func (tec testingElevationControl) Log(o output.Bus, level output.Level) {
	o.Log(level, "elevation state", tec.logFields)
}

func (tec testingElevationControl) Status(_ string) []string {
	return tec.desiredStatus
}

func (tec testingElevationControl) ConfigureExit(f func(int)) func(int) { return f }

func (tec testingElevationControl) WillRunElevated() bool { return false }

func TestAboutRun(t *testing.T) {
	originalBusGetter := cmd.BusGetter
	originalInterpretBuildData := cmd.InterpretBuildData
	originalLogPath := cmd.LogPath
	originalVersion := cmd.Version
	originalCreation := cmd.Creation
	originalApplicationPath := cmd.ApplicationPath
	originalPlainFileExists := cmd.PlainFileExists
	originalMP3RepairElevationControl := cmd.MP3RepairElevationControl
	defer func() {
		cmd.BusGetter = originalBusGetter
		cmd.InterpretBuildData = originalInterpretBuildData
		cmd.LogPath = originalLogPath
		cmd.Version = originalVersion
		cmd.Creation = originalCreation
		cmd.ApplicationPath = originalApplicationPath
		cmd.PlainFileExists = originalPlainFileExists
		cmd.MP3RepairElevationControl = originalMP3RepairElevationControl
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
	type args struct {
		in0 *cobra.Command
		in1 []string
	}
	tests := map[string]struct {
		args
		output.WantedRecording
	}{
		"simple": {
			args: args{in0: cmd.AboutCmd},
			WantedRecording: output.WantedRecording{
				Console: "" +
					"╭───────────────────────────────────────────────────────────────────────────────╮\n" +
					"│ mp3repair version 0.40.0, built on Saturday, February 24 2024, 13:14:05 -0500 │\n" +
					"│ Copyright © 2021-2024 Marc Johnson                                            │\n" +
					"│ Build Information                                                             │\n" +
					"│  - Go version: go1.22.x                                                       │\n" +
					"│  - Dependency: go.dependency.1 v1.2.3                                         │\n" +
					"│  - Dependency: go.dependency.2 v1.3.4                                         │\n" +
					"│  - Dependency: go.dependency.3 v0.1.2                                         │\n" +
					"│ Log files are written to /my/files/tmp/logs/mp3repair                         │\n" +
					"│ Configuration file \\my\\files\\apppath\\defaults.yaml exists                     │\n" +
					"│ mp3repair is running with elevated privileges                                 │\n" +
					"╰───────────────────────────────────────────────────────────────────────────────╯\n",
				Log: "level='info' command='about' style='1' msg='executing command'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			cmd.BusGetter = func() output.Bus { return o }
			cmd.MP3RepairElevationControl = testingElevationControl{
				desiredStatus: []string{"mp3repair is running with elevated privileges"},
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
					"• The program version and build timestamp\n" +
					"• Copyright information\n" +
					"• Build information:\n" +
					"  • The version of go used to compile the code\n" +
					"  • A list of dependencies and their versions\n" +
					"• The directory where log files are written\n" +
					"• The full path of the application configuration file and whether it exists\n" +
					"• Whether mp3repair is running with elevated privileges, and, if not, why not\n" +
					"\n" +
					"Usage:\n" +
					"  mp3repair about [--style name]\n" +
					"\n" +
					"Examples:\n" +
					"about --style name\n" +
					"  Write 'about' information in a box of the named style.\n" +
					"  Valid names are:\n" +
					"  ● ascii\n" +
					"    +------------+\n" +
					"    | output ... |\n" +
					"    +------------+\n" +
					"  ● rounded (default)\n" +
					"    ╭────────────╮\n" +
					"    │ output ... │\n" +
					"    ╰────────────╯\n" +
					"  ● light\n" +
					"    ┌────────────┐\n" +
					"    │ output ... │\n" +
					"    └────────────┘\n" +
					"  ● heavy\n" +
					"    ┏━━━━━━━━━━━━┓\n" +
					"    ┃ output ... ┃\n" +
					"    ┗━━━━━━━━━━━━┛\n" +
					"  ● double\n" +
					"    ╔════════════╗\n" +
					"    ║ output ... ║\n" +
					"    ╚════════════╝\n" +
					"\n" +
					"Flags:\n" +
					"      --style string   specify the output border style (default \"rounded\")\n",
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
	originalMP3RepairElevationControl := cmd.MP3RepairElevationControl
	defer func() {
		cmd.InterpretBuildData = originalInterpretBuildData
		cmd.LogPath = originalLogPath
		cmd.Version = originalVersion
		cmd.Creation = originalCreation
		cmd.ApplicationPath = originalApplicationPath
		cmd.PlainFileExists = originalPlainFileExists
		cmd.MP3RepairElevationControl = originalMP3RepairElevationControl
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
			tec := testingElevationControl{
				desiredStatus: []string{},
			}
			if tt.forceElevated {
				tec.desiredStatus = append(tec.desiredStatus, "mp3repair is running with elevated privileges")
			} else {
				tec.desiredStatus = append(tec.desiredStatus, "mp3repair is not running with elevated privileges")
				if tt.forceRedirection {
					tec.desiredStatus = append(tec.desiredStatus, "stderr, stdin, and stdout have been redirected")
				}
				if !tt.forceAdminPermission {
					tec.desiredStatus = append(tec.desiredStatus, "The environment variable MP3REPAIR_RUNS_AS_ADMIN evaluates as false")
				}
			}
			cmd.MP3RepairElevationControl = tec
			if got := cmd.AcquireAboutData(output.NewNilBus()); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AcquireAboutData() got %v, want %v", got, tt.want)
			}
		})
	}
}
