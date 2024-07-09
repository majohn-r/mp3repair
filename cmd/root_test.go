/*
Copyright © 2021 Marc Johnson (marc.johnson27591@gmail.com)
*/
package cmd_test

import (
	"fmt"
	"mp3repair/cmd"
	"os"
	"reflect"
	"testing"
	"time"

	cmdtoolkit "github.com/majohn-r/cmd-toolkit"
	"github.com/majohn-r/output"
)

func TestExecute(t *testing.T) {
	cmd.InitGlobals()
	originalArgs := os.Args
	originalExit := cmd.Exit
	originalBus := cmd.Bus
	defer func() {
		os.Args = originalArgs
		cmd.Exit = originalExit
		cmd.Bus = originalBus
	}()
	tests := map[string]struct {
		args []string
	}{
		"good": {args: nil},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			os.Args = tt.args
			got := -1
			cmd.Exit = func(code int) {
				got = code
			}
			cmd.Execute()
			if got != 0 {
				t.Errorf("Execute: got exit code %d, wanted 0", got)
			}
		})
	}
}

type happyCommand struct{}

func (h happyCommand) SetArgs(_ []string) {}
func (h happyCommand) Execute() error     { return nil }

type sadCommand struct{}

func (s sadCommand) SetArgs(_ []string) {}
func (s sadCommand) Execute() error     { return fmt.Errorf("sad") }

type panickyCommand struct{}

func (p panickyCommand) SetArgs(_ []string) {}
func (p panickyCommand) Execute() error     { panic("oh dear") }

func TestRunMain(t *testing.T) {
	originalArgs := os.Args
	originalSince := cmd.Since
	originalVersion := cmd.Version
	originalCreation := cmd.Creation
	originalCachedGoVersion := cmd.CachedGoVersion
	originalCachedBuildDependencies := cmd.CachedBuildDependencies
	originalMP3repairElevationControl := cmd.MP3RepairElevationControl
	defer func() {
		cmd.Since = originalSince
		os.Args = originalArgs
		cmd.Version = originalVersion
		cmd.Creation = originalCreation
		cmd.CachedGoVersion = originalCachedGoVersion
		cmd.CachedBuildDependencies = originalCachedBuildDependencies
		cmd.MP3RepairElevationControl = originalMP3repairElevationControl
	}()
	cmd.MP3RepairElevationControl = testingElevationControl{
		logFields: map[string]any{
			"elevated":             true,
			"admin_permission":     true,
			"stderr_redirected":    false,
			"stdin_redirected":     false,
			"stdout_redirected":    false,
			"environment_variable": cmd.ElevatedPrivilegesPermissionVar,
		},
	}
	type args struct {
		cmd   cmd.CommandExecutor
		start time.Time
	}
	tests := map[string]struct {
		args
		cmdline      []string
		appVersion   string
		timestamp    string
		goVersion    string
		dependencies []string
		output.WantedRecording
	}{
		"happy": {
			args:         args{cmd: happyCommand{}, start: time.Now()},
			cmdline:      []string{"happyApp", "arg1", "arg2"},
			appVersion:   "0.1.2",
			timestamp:    "2021-11-28T12:01:02Z05:00",
			goVersion:    "1.22.x",
			dependencies: []string{"foo v1.1.1", "bar v1.2.2"},
			WantedRecording: output.WantedRecording{
				Log: "level='info'" +
					" args='[arg1 arg2]'" +
					" dependencies='[foo v1.1.1 bar v1.2.2]'" +
					" goVersion='1.22.x'" +
					" timeStamp='2021-11-28T12:01:02Z05:00'" +
					" version='0.1.2'" +
					" msg='execution starts'\n" +
					"level='info'" +
					" admin_permission='true'" +
					" elevated='true'" +
					" environment_variable='MP3REPAIR_RUNS_AS_ADMIN'" +
					" stderr_redirected='false'" +
					" stdin_redirected='false'" +
					" stdout_redirected='false'" +
					" msg='elevation state'\n" +
					"level='info'" +
					" duration='0s'" +
					" exitCode='0'" +
					" msg='execution ends'\n",
			},
		},
		"sad": {
			args:         args{cmd: sadCommand{}, start: time.Now()},
			appVersion:   "0.2.3",
			timestamp:    "2021-11-29T13:02:03Z05:00",
			cmdline:      []string{"sadApp", "arg1a", "arg2a"},
			goVersion:    "1.22.x",
			dependencies: []string{"foo v1.1.2", "bar v1.2.3"},
			WantedRecording: output.WantedRecording{
				Error: "" +
					"\"mp3repair\" version 0.2.3, created at 2021-11-29T13:02:03Z05:00, failed.\n",
				Log: "level='info'" +
					" args='[arg1a arg2a]'" +
					" dependencies='[foo v1.1.2 bar v1.2.3]'" +
					" goVersion='1.22.x'" +
					" timeStamp='2021-11-29T13:02:03Z05:00'" +
					" version='0.2.3'" +
					" msg='execution starts'\n" +
					"level='info'" +
					" admin_permission='true'" +
					" elevated='true'" +
					" environment_variable='MP3REPAIR_RUNS_AS_ADMIN'" +
					" stderr_redirected='false'" +
					" stdin_redirected='false'" +
					" stdout_redirected='false'" +
					" msg='elevation state'\n" +
					"level='info'" +
					" duration='0s'" +
					" exitCode='1'" +
					" msg='execution ends'\n",
			},
		},
		"panicky": {
			args:         args{cmd: panickyCommand{}, start: time.Now()},
			appVersion:   "0.2.3",
			timestamp:    "2021-11-29T13:02:03Z05:00",
			cmdline:      []string{"sadApp", "arg1a", "arg2a"},
			goVersion:    "1.22.x",
			dependencies: []string{"foo v1.1.2", "bar v1.2.3"},
			WantedRecording: output.WantedRecording{
				Error: "" +
					"A runtime error occurred: \"oh dear\".\n",
				Log: "level='info'" +
					" args='[arg1a arg2a]'" +
					" dependencies='[foo v1.1.2 bar v1.2.3]'" +
					" goVersion='1.22.x'" +
					" timeStamp='2021-11-29T13:02:03Z05:00'" +
					" version='0.2.3'" +
					" msg='execution starts'\n" +
					"level='info'" +
					" admin_permission='true'" +
					" elevated='true'" +
					" environment_variable='MP3REPAIR_RUNS_AS_ADMIN'" +
					" stderr_redirected='false'" +
					" stdin_redirected='false'" +
					" stdout_redirected='false'" +
					" msg='elevation state'\n" +
					"level='error'" +
					" error='oh dear'" +
					" msg='Panic recovered'\n",
			},
		},
	}
	for name, tt := range tests {
		cmd.Since = func(_ time.Time) time.Duration {
			return 0
		}
		t.Run(name, func(t *testing.T) {
			os.Args = tt.cmdline
			cmd.Version = tt.appVersion
			cmd.Creation = tt.timestamp
			cmd.CachedGoVersion = tt.goVersion
			cmd.CachedBuildDependencies = tt.dependencies
			o := output.NewRecorder()
			cmd.RunMain(o, tt.args.cmd, tt.args.start)
			o.Report(t, "RunMain()", tt.WantedRecording)
		})
	}
}

func TestCookCommandLineArguments(t *testing.T) {
	originalDereferenceEnvVar := cmd.DereferenceEnvVar
	defer func() {
		cmd.DereferenceEnvVar = originalDereferenceEnvVar
	}()
	tests := map[string]struct {
		inputArgs         []string
		dereferenceEnvVar func(string) (string, error)
		want              []string
		output.WantedRecording
	}{
		"nil args": {
			inputArgs: nil,
			want:      []string{},
		},
		"no args": {
			inputArgs: []string{},
			want:      []string{},
		},
		"only 1 arg": {
			inputArgs: []string{"app_Name"},
			want:      []string{},
		},
		"multiple args with problems": {
			inputArgs: []string{"app_Name", "%arg%", "foo", "bar"},
			dereferenceEnvVar: func(s string) (string, error) {
				if s == "%arg%" {
					return "", fmt.Errorf("dereference service dead")
				} else {
					return s, nil
				}
			},
			want: []string{"foo", "bar"},
			WantedRecording: output.WantedRecording{
				Error: "An error was found in processing argument \"%arg%\": dereference" +
					" service dead.\n",
				Log: "level='error'" +
					" argument='%arg%'" +
					" error='dereference service dead'" +
					" msg='Invalid argument value'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			cmd.DereferenceEnvVar = tt.dereferenceEnvVar
			got := cmd.CookCommandLineArguments(o, tt.inputArgs)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CookCommandLineArguments() = %v, want %v", got, tt.want)
			}
			o.Report(t, "CookCommandLineArguments()", tt.WantedRecording)
		})
	}
}

func Test_InitGlobals(t *testing.T) {
	originalExit := cmd.Exit
	originalNewDefaultBus := cmd.NewDefaultBus
	originalInitLogging := cmd.InitLogging
	originalInitApplicationPath := cmd.InitApplicationPath
	originalReadConfigurationFile := cmd.ReadConfigurationFile
	originalSetFlagIndicator := cmd.SetFlagIndicator
	originalVersion := cmd.Version
	originalCreation := cmd.Creation
	originalInitialized := cmd.Initialized
	originalBus := cmd.Bus
	originalInternalConfig := cmd.InternalConfig
	defer func() {
		cmd.Exit = originalExit
		cmd.NewDefaultBus = originalNewDefaultBus
		cmd.InitLogging = originalInitLogging
		cmd.InitApplicationPath = originalInitApplicationPath
		cmd.ReadConfigurationFile = originalReadConfigurationFile
		cmd.SetFlagIndicator = originalSetFlagIndicator
		cmd.Version = originalVersion
		cmd.Creation = originalCreation
		cmd.Initialized = originalInitialized
		cmd.Bus = originalBus
		cmd.InternalConfig = originalInternalConfig
	}()
	o := output.NewRecorder()
	defaultExitFunctionCalled := false
	defaultExitCode := -1
	defaultCreation := ""
	defaultVersion := ""
	defaultFlagIndicator := ""
	ExitFunctionCalled := false
	exitCodeRecorded := defaultExitCode
	creationRecorded := defaultCreation
	versionRecorded := defaultVersion
	flagIndicatorRecorded := defaultFlagIndicator
	tests := map[string]struct {
		initialize            bool
		exitFunc              func(int)
		wantExitFuncCalled    bool
		wantExitValue         int
		newDefaultBus         func(output.Logger) output.Bus
		initLogging           func(output.Bus, string) bool
		initApplicationPath   func(output.Bus, string) bool
		readConfigurationFile func(output.Bus) (*cmdtoolkit.Configuration, bool)
		wantConfig            *cmdtoolkit.Configuration
		wantCreation          string
		wantVersion           string
		setFlagIndicator      func(string)
		wantFlagIndicator     string
		versionVal            string
		creationVal           string
		output.WantedRecording
	}{
		"already initialized": {
			initialize:    true,
			wantConfig:    cmdtoolkit.EmptyConfiguration(),
			wantExitValue: defaultExitCode,
		},
		"log initialization failure": {
			initialize: false,
			newDefaultBus: func(output.Logger) output.Bus {
				return o
			},
			initLogging: func(output.Bus, string) bool {
				return false
			},
			exitFunc: func(c int) {
				exitCodeRecorded = c
				ExitFunctionCalled = true
			},
			wantConfig:         cmdtoolkit.EmptyConfiguration(),
			wantExitFuncCalled: true,
			wantExitValue:      1,
		},
		"app path initialization failure": {
			initialize: false,
			newDefaultBus: func(output.Logger) output.Bus {
				return o
			},
			initLogging: func(output.Bus, string) bool {
				return true
			},
			initApplicationPath: func(output.Bus, string) bool {
				return false
			},
			exitFunc: func(c int) {
				exitCodeRecorded = c
				ExitFunctionCalled = true
			},
			wantConfig:         cmdtoolkit.EmptyConfiguration(),
			wantExitFuncCalled: true,
			wantExitValue:      1,
		},
		"config file read failed": {
			initialize: false,
			newDefaultBus: func(output.Logger) output.Bus {
				return o
			},
			initLogging: func(output.Bus, string) bool {
				return true
			},
			initApplicationPath: func(output.Bus, string) bool {
				return true
			},
			readConfigurationFile: func(output.Bus) (*cmdtoolkit.Configuration, bool) {
				return nil, false
			},
			exitFunc: func(c int) {
				exitCodeRecorded = c
				ExitFunctionCalled = true
			},
			wantConfig:         nil,
			wantExitFuncCalled: true,
			wantExitValue:      1,
		},
		"all is well": {
			initialize: false,
			newDefaultBus: func(output.Logger) output.Bus {
				return o
			},
			initLogging: func(output.Bus, string) bool {
				return true
			},
			initApplicationPath: func(output.Bus, string) bool {
				return true
			},
			readConfigurationFile: func(output.Bus) (*cmdtoolkit.Configuration, bool) {
				return cmdtoolkit.EmptyConfiguration(), true
			},
			creationVal:        "created today",
			wantCreation:       "",
			versionVal:         "v0.1.1",
			wantVersion:        "",
			setFlagIndicator:   func(s string) { flagIndicatorRecorded = s },
			wantFlagIndicator:  "",
			wantConfig:         cmdtoolkit.EmptyConfiguration(),
			wantExitFuncCalled: false,
			wantExitValue:      defaultExitCode,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o = output.NewRecorder()
			cmd.InternalConfig = cmdtoolkit.EmptyConfiguration()
			ExitFunctionCalled = defaultExitFunctionCalled
			exitCodeRecorded = defaultExitCode
			creationRecorded = defaultCreation
			versionRecorded = defaultVersion
			flagIndicatorRecorded = defaultFlagIndicator
			cmd.Initialized = tt.initialize
			cmd.Exit = tt.exitFunc
			cmd.NewDefaultBus = tt.newDefaultBus
			cmd.InitLogging = tt.initLogging
			cmd.InitApplicationPath = tt.initApplicationPath
			cmd.ReadConfigurationFile = tt.readConfigurationFile
			cmd.SetFlagIndicator = tt.setFlagIndicator
			cmd.Creation = tt.creationVal
			cmd.Version = tt.versionVal
			cmd.InitGlobals()
			if got := cmd.InternalConfig; !reflect.DeepEqual(got, tt.wantConfig) {
				t.Errorf("InitGlobals: _c got %v want %v", got, tt.wantConfig)
			}
			if got := ExitFunctionCalled; got != tt.wantExitFuncCalled {
				t.Errorf("InitGlobals: exit called got %t want %t", got, tt.wantExitFuncCalled)
			}
			if got := exitCodeRecorded; got != tt.wantExitValue {
				t.Errorf("InitGlobals: exit code got %d want %d", got, tt.wantExitValue)
			}
			if got := creationRecorded; got != tt.wantCreation {
				t.Errorf("InitGlobals: creation got %q want %q", got, tt.wantCreation)
			}
			if got := versionRecorded; got != tt.wantVersion {
				t.Errorf("InitGlobals: version got %q want %q", got, tt.wantVersion)
			}
			if got := flagIndicatorRecorded; got != tt.wantFlagIndicator {
				t.Errorf("InitGlobals: flag indicator got %q want %q", got, tt.wantFlagIndicator)
			}
			o.Report(t, "InitGlobals()", tt.WantedRecording)
		})
	}
}

func TestRootUsage(t *testing.T) {
	tests := map[string]struct {
		output.WantedRecording
	}{
		"good": {
			WantedRecording: output.WantedRecording{
				Console: "" +
					"Usage:\n" +
					"\n" +
					"Examples:\n" +
					"The mp3repair program might be used like this:\n" +
					"\n" +
					"First, get a listing of the available mp3 files:\n" +
					"\n" +
					"mp3repair list -lrt\n" +
					"\n" +
					"Then check for problems in the track metadata:\n" +
					"\n" +
					"mp3repair check --files\n" +
					"\n" +
					"If problems were found, repair the mp3 files:\n" +
					"\n" +
					"mp3repair repair\n" +
					"The repair command creates backup files for each track it rewrites. After\n" +
					"listening to the files that have been repaired (verifying that the repair\n" +
					"process did not corrupt the audio), clean up those backups:\n" +
					"\n" +
					"mp3repair postRepair\n" +
					"\n" +
					"After repairing the mp3 files, the Windows media player system may be out of\n" +
					"sync with the changes. While the system will eventually catch up, accelerate\n" +
					"the process:\n" +
					"\n" +
					"mp3repair resetDatabase\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			command := cloneCommand(cmd.RootCmd)
			enableCommandRecording(o, command)
			_ = command.Usage()
			o.Report(t, "root Usage()", tt.WantedRecording)
		})
	}
}

func TestObtainExitCode(t *testing.T) {
	var nilExitError *cmdtoolkit.ExitError
	tests := map[string]struct {
		err  error
		want int
	}{
		"nil":               {err: nil, want: 0},
		"nil ExitError":     {err: nilExitError, want: 0},
		"user error":        {err: cmdtoolkit.NewExitUserError("command"), want: 1},
		"programming error": {err: cmdtoolkit.NewExitProgrammingError("command"), want: 2},
		"system error":      {err: cmdtoolkit.NewExitSystemError("command"), want: 3},
		"unexpected":        {err: fmt.Errorf("some error"), want: 1},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := cmd.ObtainExitCode(tt.err); got != tt.want {
				t.Errorf("ObtainExitCode() = %v, want %v", got, tt.want)
			}
		})
	}
}
