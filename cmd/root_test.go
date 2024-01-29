/*
Copyright © 2021 Marc Johnson (marc.johnson27591@gmail.com)
*/
package cmd_test

import (
	"fmt"
	"mp3/cmd"
	"os"
	"reflect"
	"testing"
	"time"

	cmd_toolkit "github.com/majohn-r/cmd-toolkit"
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

func (h happyCommand) SetArgs(a []string) {}
func (h happyCommand) Execute() error     { return nil }

type sadCommand struct{}

func (s sadCommand) SetArgs(a []string) {}
func (s sadCommand) Execute() error     { return fmt.Errorf("sad") }

func TestRunMain(t *testing.T) {
	originalArgs := os.Args
	originalSince := cmd.Since
	originalExit := cmd.Exit
	originalVersion := cmd.Version
	originalCreation := cmd.Creation
	defer func() {
		cmd.Since = originalSince
		cmd.Exit = originalExit
		os.Args = originalArgs
		cmd.Version = originalVersion
		cmd.Creation = originalCreation
	}()
	type args struct {
		cmd   cmd.CommandExecutor
		start time.Time
	}
	tests := map[string]struct {
		args
		cmdline        []string
		appVersion     string
		timestamp      string
		goVersion      string
		dependencies   []string
		wantedExitCode int
		output.WantedRecording
	}{
		"happy": {
			args:           args{cmd: happyCommand{}, start: time.Now()},
			cmdline:        []string{"happyApp", "arg1", "arg2"},
			appVersion:     "0.1.2",
			timestamp:      "2021-11-28T12:01:02Z05:00",
			goVersion:      "1.21.x",
			dependencies:   []string{"foo v1.1.1", "bar v1.2.2"},
			wantedExitCode: 0,
			WantedRecording: output.WantedRecording{
				Log: "level='info' args='[arg1 arg2]' dependencies='[foo v1.1.1 bar v1.2.2]' goVersion='1.21.x' timeStamp='2021-11-28T12:01:02Z05:00' version='0.1.2' msg='execution starts'\n" +
					"level='info' duration='0s' exitCode='0' msg='execution ends'\n",
			},
		},
		"sad": {
			args:           args{cmd: sadCommand{}, start: time.Now()},
			appVersion:     "0.2.3",
			timestamp:      "2021-11-29T13:02:03Z05:00",
			cmdline:        []string{"sadApp", "arg1a", "arg2a"},
			goVersion:      "1.22.x",
			dependencies:   []string{"foo v1.1.2", "bar v1.2.3"},
			wantedExitCode: 1,
			WantedRecording: output.WantedRecording{
				Error: "\"mp3\" version 0.2.3, created at 2021-11-29T13:02:03Z05:00, failed.\n",
				Log: "level='info' args='[arg1a arg2a]' dependencies='[foo v1.1.2 bar v1.2.3]' goVersion='1.22.x' timeStamp='2021-11-29T13:02:03Z05:00' version='0.2.3' msg='execution starts'\n" +
					"level='info' duration='0s' exitCode='1' msg='execution ends'\n",
			},
		},
	}
	for name, tt := range tests {
		cmd.Since = func(_ time.Time) time.Duration {
			return 0
		}
		var capturedExitCode int
		cmd.Exit = func(code int) {
			capturedExitCode = code
		}
		t.Run(name, func(t *testing.T) {
			capturedExitCode = -1
			os.Args = tt.cmdline
			cmd.Version = tt.appVersion
			cmd.Creation = tt.timestamp
			cmd.GoVersion = func() string {
				return tt.goVersion
			}
			cmd.BuildDependencies = func() []string {
				return tt.dependencies
			}
			o := output.NewRecorder()
			cmd.RunMain(o, tt.args.cmd, tt.args.start)
			if capturedExitCode != tt.wantedExitCode {
				t.Errorf("RunMain() got %d want %d", capturedExitCode, tt.wantedExitCode)
			}
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("RunMain() %s", issue)
				}
			}
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
				Error: "An error was found in processng argument \"%arg%\": dereference service dead.\n",
				Log:   "level='error' argument='%arg%' error='dereference service dead' msg='Invalid argument value'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			cmd.DereferenceEnvVar = tt.dereferenceEnvVar
			if got := cmd.CookCommandLineArguments(o, tt.inputArgs); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CookCommandLineArguments() = %v, want %v", got, tt.want)
			}
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("CookCommandLineArguments() %s", issue)
				}
			}
		})
	}
}

func Test_InitGlobals(t *testing.T) {
	originalExit := cmd.Exit
	originalNewDefaultBus := cmd.NewDefaultBus
	originalSetAppName := cmd.SetAppName
	originalInitLogging := cmd.InitLogging
	originalInitApplicationPath := cmd.InitApplicationPath
	originalReadConfigurationFile := cmd.ReadConfigurationFile
	originalInitBuildData := cmd.InitBuildData
	originalSetFlagIndicator := cmd.SetFlagIndicator
	originalVersion := cmd.Version
	originalCreation := cmd.Creation
	originalInitialized := cmd.Initialized
	originalBus := cmd.Bus
	originalInternalConfig := cmd.InternalConfig
	defer func() {
		cmd.Exit = originalExit
		cmd.NewDefaultBus = originalNewDefaultBus
		cmd.SetAppName = originalSetAppName
		cmd.InitLogging = originalInitLogging
		cmd.InitApplicationPath = originalInitApplicationPath
		cmd.ReadConfigurationFile = originalReadConfigurationFile
		cmd.InitBuildData = originalInitBuildData
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
	defaultAppName := ""
	defaultCreation := ""
	defaultVersion := ""
	defaultFlagIndicator := ""
	ExitFunctionCalled := defaultExitFunctionCalled
	exitCodeRecorded := defaultExitCode
	appNameRecorded := defaultAppName
	creationRecorded := defaultCreation
	versionRecorded := defaultVersion
	flagIndicatorRecorded := defaultFlagIndicator
	tests := map[string]struct {
		initialize            bool
		exitFunc              func(int)
		wantExitFuncCalled    bool
		wantExitValue         int
		newDefaultBus         func(output.Logger) output.Bus
		setAppName            func(string) error
		initLogging           func(output.Bus) bool
		initApplicationPath   func(output.Bus) bool
		readConfigurationFile func(output.Bus) (*cmd_toolkit.Configuration, bool)
		wantConfig            *cmd_toolkit.Configuration
		initBuildData         func(string, string)
		wantCreation          string
		wantVersion           string
		setFlagIndicator      func(string)
		wantFlagIndicator     string
		versionVal            string
		creationVal           string
		wantAppName           string
		output.WantedRecording
	}{
		"already initialized": {
			initialize:    true,
			wantConfig:    cmd_toolkit.EmptyConfiguration(),
			wantExitValue: defaultExitCode,
		},
		"app name set error": {
			initialize: false,
			newDefaultBus: func(output.Logger) output.Bus {
				return o
			},
			setAppName: func(string) error {
				return fmt.Errorf("app name could not be set")
			},
			initLogging: func(_ output.Bus) bool { return false },
			exitFunc: func(c int) {
				exitCodeRecorded = c
				ExitFunctionCalled = true
			},
			wantConfig:         cmd_toolkit.EmptyConfiguration(),
			wantExitFuncCalled: true,
			wantExitValue:      1,
		},
		"log initialization failure": {
			initialize: false,
			newDefaultBus: func(output.Logger) output.Bus {
				return o
			},
			setAppName: func(s string) error {
				appNameRecorded = s
				return nil
			},
			initLogging: func(output.Bus) bool {
				return false
			},
			exitFunc: func(c int) {
				exitCodeRecorded = c
				ExitFunctionCalled = true
			},
			wantConfig:         cmd_toolkit.EmptyConfiguration(),
			wantExitFuncCalled: true,
			wantExitValue:      1,
		},
		"app path initialization failure": {
			initialize: false,
			newDefaultBus: func(output.Logger) output.Bus {
				return o
			},
			setAppName: func(s string) error {
				appNameRecorded = s
				return nil
			},
			initLogging: func(output.Bus) bool {
				return true
			},
			initApplicationPath: func(output.Bus) bool {
				return false
			},
			exitFunc: func(c int) {
				exitCodeRecorded = c
				ExitFunctionCalled = true
			},
			wantConfig:         cmd_toolkit.EmptyConfiguration(),
			wantExitFuncCalled: true,
			wantExitValue:      1,
		},
		"config file read failed": {
			initialize: false,
			newDefaultBus: func(output.Logger) output.Bus {
				return o
			},
			setAppName: func(s string) error {
				appNameRecorded = s
				return nil
			},
			initLogging: func(output.Bus) bool {
				return true
			},
			initApplicationPath: func(output.Bus) bool {
				return true
			},
			readConfigurationFile: func(output.Bus) (*cmd_toolkit.Configuration, bool) {
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
			setAppName: func(s string) error {
				appNameRecorded = s
				return nil
			},
			initLogging: func(output.Bus) bool {
				return true
			},
			initApplicationPath: func(output.Bus) bool {
				return true
			},
			readConfigurationFile: func(output.Bus) (*cmd_toolkit.Configuration, bool) {
				return cmd_toolkit.EmptyConfiguration(), true
			},
			creationVal:  "created today",
			wantCreation: "",
			versionVal:   "v0.1.1",
			wantVersion:  "",
			initBuildData: func(v string, c string) {
				versionRecorded = v
				creationRecorded = c
			},
			setFlagIndicator:   func(s string) { flagIndicatorRecorded = s },
			wantFlagIndicator:  "",
			wantConfig:         cmd_toolkit.EmptyConfiguration(),
			wantExitFuncCalled: false,
			wantExitValue:      defaultExitCode,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o = output.NewRecorder()
			cmd.InternalConfig = cmd_toolkit.EmptyConfiguration()
			ExitFunctionCalled = defaultExitFunctionCalled
			exitCodeRecorded = defaultExitCode
			appNameRecorded = defaultAppName
			creationRecorded = defaultCreation
			versionRecorded = defaultVersion
			flagIndicatorRecorded = defaultFlagIndicator
			cmd.Initialized = tt.initialize
			cmd.Exit = tt.exitFunc
			cmd.NewDefaultBus = tt.newDefaultBus
			cmd.SetAppName = tt.setAppName
			cmd.InitLogging = tt.initLogging
			cmd.InitApplicationPath = tt.initApplicationPath
			cmd.ReadConfigurationFile = tt.readConfigurationFile
			cmd.InitBuildData = tt.initBuildData
			cmd.SetFlagIndicator = tt.setFlagIndicator
			cmd.Creation = tt.creationVal
			cmd.Version = tt.versionVal
			cmd.InitGlobals()
			if got := appNameRecorded; got != tt.wantAppName {
				t.Errorf("InitGlobals appNameRecorded got %s want %s", got, tt.wantAppName)
			}
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
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("InitGlobals() %s", issue)
				}
			}
		})
	}
}

func TestRootHelp(t *testing.T) {
	tests := map[string]struct {
		output.WantedRecording
		optional output.WantedRecording
	}{
		"good": {
			WantedRecording: output.WantedRecording{
				Console: "" +
					"Usage:\n" +
					"  mp3 [command]\n" +
					"\n" +
					"Examples:\n" +
					"The mp3 program might be used like this:\n" +
					"\n" +
					"First, get a listing of the available mp3 files:\n" +
					"\n" +
					"mp3 list -lrt\n" +
					"\n" +
					"Then check for problems in the track metadata:\n" +
					"\n" +
					"mp3 check --files\n" +
					"\n" +
					"If problems were found, repair the mp3 files:\n" +
					"\n" +
					"mp3 repair\n" +
					"\n" +
					"The repair command creates backup files for each track it rewrites. After\n" +
					"spot-checking files that have been repaired, clean up those backups:\n" +
					"\n" +
					"mp3 postRepair\n" +
					"\n" +
					"After repairing the mp3 files, the Windows media player system may be out of\n" +
					"sync with the changes. While the system will eventually catch up, accelerate\n" +
					"the process:\n" +
					"\n" +
					"mp3 resetDatabase\n" +
					"\n" +
					"Available Commands:\n" +
					"  about         Provides information about the mp3 program\n" +
					"  check         Runs checks on mp3 files and their directories and reports problems\n" +
					"  export        Exports default program configuration data\n" +
					"  list          Lists mp3 files and containing album and artist directories\n" +
					"  postRepair    Deletes the backup directories, and their contents, created by the repair command\n" +
					"  repair        Repairs problems found by running 'check --files'\n" +
					"  resetDatabase Resets the Windows music database\n" +
					"\n" +
					"Use \"mp3 [command] --help\" for more information about a command.\n",
			},
			optional: output.WantedRecording{
				Console: "" +
					"Usage:\n" +
					"  mp3 [command]\n" +
					"\n" +
					"Examples:\n" +
					"The mp3 program might be used like this:\n" +
					"\n" +
					"First, get a listing of the available mp3 files:\n" +
					"\n" +
					"mp3 list -lrt\n" +
					"\n" +
					"Then check for problems in the track metadata:\n" +
					"\n" +
					"mp3 check --files\n" +
					"\n" +
					"If problems were found, repair the mp3 files:\n" +
					"\n" +
					"mp3 repair\n" +
					"\n" +
					"The repair command creates backup files for each track it rewrites. After\n" +
					"spot-checking files that have been repaired, clean up those backups:\n" +
					"\n" +
					"mp3 postRepair\n" +
					"\n" +
					"After repairing the mp3 files, the Windows media player system may be out of\n" +
					"sync with the changes. While the system will eventually catch up, accelerate\n" +
					"the process:\n" +
					"\n" +
					"mp3 resetDatabase\n" +
					"\n" +
					"Available Commands:\n" +
					"  about         Provides information about the mp3 program\n" +
					"  check         Runs checks on mp3 files and their directories and reports problems\n" +
					"  export        Exports default program configuration data\n" +
					"  help          Help about any command\n" +
					"  list          Lists mp3 files and containing album and artist directories\n" +
					"  postRepair    Deletes the backup directories, and their contents, created by the repair command\n" +
					"  repair        Repairs problems found by running 'check --files'\n" +
					"  resetDatabase Resets the Windows music database\n" +
					"\n" +
					"Flags:\n  -h, --help   help for mp3\n" +
					"\n" +
					"Use \"mp3 [command] --help\" for more information about a command.\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			command := cmd.RootCmd
			enableCommandRecording(o, command)
			command.Usage()
			issues, ok := o.Verify(tt.WantedRecording)
			optionalIssues, optionalOk := o.Verify(tt.optional)
			if !ok && !optionalOk {
				for _, issue := range issues {
					t.Errorf("root Help() %s", issue)
				}
				for _, issue := range optionalIssues {
					t.Errorf("root Help() optional %s", issue)
				}
			}
		})
	}
}
