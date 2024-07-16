package cmd

import (
	cmdtoolkit "github.com/majohn-r/cmd-toolkit"
	"github.com/spf13/afero"
	"path/filepath"
	"reflect"
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

func Test_aboutRun(t *testing.T) {
	originalBusGetter := busGetter
	originalLogPath := logPath
	originalVersion := version
	originalCreation := creation
	originalApplicationPath := cmdtoolkit.SetApplicationPath("/my/files/apppath")
	originalMP3RepairElevationControl := mp3repairElevationControl
	fs := afero.NewOsFs()
	originalFileSystem := cmdtoolkit.AssignFileSystem(fs)
	originalCachedGoVersion := cachedGoVersion
	originalCachedBuildDependencies := cachedBuildDependencies
	defer func() {
		busGetter = originalBusGetter
		logPath = originalLogPath
		version = originalVersion
		creation = originalCreation
		cmdtoolkit.SetApplicationPath(originalApplicationPath)
		mp3repairElevationControl = originalMP3RepairElevationControl
		cmdtoolkit.AssignFileSystem(originalFileSystem)
		cachedGoVersion = originalCachedGoVersion
		cachedBuildDependencies = originalCachedBuildDependencies
	}()
	cachedGoVersion = "go1.22.x"
	cachedBuildDependencies = []string{
		"go.dependency.1 v1.2.3",
		"go.dependency.2 v1.3.4",
		"go.dependency.3 v0.1.2",
	}
	logPath = func() string {
		return "/my/files/tmp/logs/mp3repair"
	}
	version = "0.40.0"
	creation = "2024-02-24T13:14:05-05:00"
	_ = fs.MkdirAll(cmdtoolkit.ApplicationPath(), cmdtoolkit.StdDirPermissions)
	_ = afero.WriteFile(
		fs,
		filepath.Join(cmdtoolkit.ApplicationPath(), "defaults.yaml"),
		[]byte{},
		cmdtoolkit.StdFilePermissions,
	)
	type args struct {
		in0 *cobra.Command
		in1 []string
	}
	tests := map[string]struct {
		args
		output.WantedRecording
	}{
		"simple": {
			args: args{in0: aboutCmd},
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
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			busGetter = func() output.Bus { return o }
			mp3repairElevationControl = testingElevationControl{
				desiredStatus: []string{"mp3repair is running with elevated privileges"},
			}
			_ = aboutRun(tt.args.in0, tt.args.in1)
			o.Report(t, "aboutRun()", tt.WantedRecording)
		})
	}
}

func enableCommandRecording(o *output.Recorder, command *cobra.Command) {
	command.SetErr(o.ErrorWriter())
	command.SetOut(o.ConsoleWriter())
}

func Test_about_Help(t *testing.T) {
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
			command := aboutCmd
			enableCommandRecording(o, command)
			_ = command.Help()
			o.Report(t, "about Help()", tt.WantedRecording)
		})
	}
}

func Test_acquireAboutData(t *testing.T) {
	originalLogPath := logPath
	originalVersion := version
	originalCreation := creation
	originalMP3RepairElevationControl := mp3repairElevationControl
	originalApplicationPath := cmdtoolkit.SetApplicationPath("/my/files/apppath")
	fs := afero.NewMemMapFs()
	originalFileSystem := cmdtoolkit.AssignFileSystem(fs)
	originalCachedGoVersion := cachedGoVersion
	originalCachedBuildDependencies := cachedBuildDependencies
	defer func() {
		logPath = originalLogPath
		version = originalVersion
		creation = originalCreation
		mp3repairElevationControl = originalMP3RepairElevationControl
		cmdtoolkit.AssignFileSystem(originalFileSystem)
		cmdtoolkit.SetApplicationPath(originalApplicationPath)
		cachedGoVersion = originalCachedGoVersion
		cachedBuildDependencies = originalCachedBuildDependencies
	}()
	cachedGoVersion = "go1.22.x"
	cachedBuildDependencies = []string{
		"go.dependency.1 v1.2.3",
		"go.dependency.2 v1.3.4",
		"go.dependency.3 v0.1.2",
	}
	logPath = func() string {
		return "/my/files/tmp/logs/mp3repair"
	}
	version = "0.40.0"
	creation = "2024-02-24T13:14:05-05:00"
	tests := map[string]struct {
		preTest              func()
		postTest             func()
		forceElevated        bool
		forceRedirection     bool
		forceAdminPermission bool
		want                 []string
	}{
		"with existing config file, elevated": {
			//plainFileExists:      func(_ string) bool { return true },
			preTest: func() {
				_ = fs.Mkdir(cmdtoolkit.ApplicationPath(), cmdtoolkit.StdDirPermissions)
				_ = afero.WriteFile(
					fs,
					filepath.Join(cmdtoolkit.ApplicationPath(), "defaults.yaml"),
					[]byte(""),
					cmdtoolkit.StdFilePermissions,
				)
			},
			postTest: func() {
				_ = fs.RemoveAll(cmdtoolkit.ApplicationPath())
			},
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
			preTest: func() {
				_ = fs.Mkdir(cmdtoolkit.ApplicationPath(), cmdtoolkit.StdDirPermissions)
			},
			postTest: func() {
				_ = fs.RemoveAll(cmdtoolkit.ApplicationPath())
			},
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
			preTest: func() {
				_ = fs.Mkdir(cmdtoolkit.ApplicationPath(), cmdtoolkit.StdDirPermissions)
			},
			postTest: func() {
				_ = fs.RemoveAll(cmdtoolkit.ApplicationPath())
			},
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
			preTest: func() {
				_ = fs.Mkdir(cmdtoolkit.ApplicationPath(), cmdtoolkit.StdDirPermissions)
			},
			postTest: func() {
				_ = fs.RemoveAll(cmdtoolkit.ApplicationPath())
			},
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
			tt.preTest()
			defer tt.postTest()
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
			mp3repairElevationControl = tec
			if got := acquireAboutData(output.NewNilBus()); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("acquireAboutData() got %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_interpretStyle(t *testing.T) {
	tests := map[string]struct {
		flag cmdtoolkit.CommandFlag[string]
		want cmdtoolkit.FlowerBoxStyle
	}{
		"lc ascii": {
			flag: cmdtoolkit.CommandFlag[string]{Value: "ascii"},
			want: cmdtoolkit.ASCIIFlowerBox,
		},
		"uc ascii": {
			flag: cmdtoolkit.CommandFlag[string]{Value: "ASCII"},
			want: cmdtoolkit.ASCIIFlowerBox,
		},
		"lc rounded": {
			flag: cmdtoolkit.CommandFlag[string]{Value: "rounded"},
			want: cmdtoolkit.CurvedFlowerBox,
		},
		"uc rounded": {
			flag: cmdtoolkit.CommandFlag[string]{Value: "ROUNDED"},
			want: cmdtoolkit.CurvedFlowerBox,
		},
		"lc light": {
			flag: cmdtoolkit.CommandFlag[string]{Value: "light"},
			want: cmdtoolkit.LightLinedFlowerBox,
		},
		"uc light": {
			flag: cmdtoolkit.CommandFlag[string]{Value: "LIGHT"},
			want: cmdtoolkit.LightLinedFlowerBox,
		},
		"lc heavy": {
			flag: cmdtoolkit.CommandFlag[string]{Value: "heavy"},
			want: cmdtoolkit.HeavyLinedFlowerBox,
		},
		"uc heavy": {
			flag: cmdtoolkit.CommandFlag[string]{Value: "HEAVY"},
			want: cmdtoolkit.HeavyLinedFlowerBox,
		},
		"lc double": {
			flag: cmdtoolkit.CommandFlag[string]{Value: "double"},
			want: cmdtoolkit.DoubleLinedFlowerBox,
		},
		"uc double": {
			flag: cmdtoolkit.CommandFlag[string]{Value: "DOUBLE"},
			want: cmdtoolkit.DoubleLinedFlowerBox,
		},
		"empty": {
			flag: cmdtoolkit.CommandFlag[string]{Value: ""},
			want: cmdtoolkit.CurvedFlowerBox,
		},
		"garbage": {
			flag: cmdtoolkit.CommandFlag[string]{Value: "abc"},
			want: cmdtoolkit.CurvedFlowerBox,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := interpretStyle(tt.flag); got != tt.want {
				t.Errorf("interpretStyle() = %v, want %v", got, tt.want)
			}
		})
	}
}
