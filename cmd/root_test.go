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
	originalExitFunction := cmd.ExitFunction
	originalBus := cmd.Bus
	defer func() {
		os.Args = originalArgs
		cmd.ExitFunction = originalExitFunction
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
			cmd.ExitFunction = func(code int) {
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
	oldArgs := os.Args
	oldDurationCalc := cmd.DurationCalc
	oldExitFunction := cmd.ExitFunction
	oldVersion := cmd.Version
	oldCreation := cmd.Creation
	defer func() {
		cmd.DurationCalc = oldDurationCalc
		cmd.ExitFunction = oldExitFunction
		os.Args = oldArgs
		cmd.Version = oldVersion
		cmd.Creation = oldCreation
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
		cmd.DurationCalc = func(_ time.Time) time.Duration {
			return 0
		}
		var capturedExitCode int
		cmd.ExitFunction = func(code int) {
			capturedExitCode = code
		}
		t.Run(name, func(t *testing.T) {
			capturedExitCode = -1
			os.Args = tt.cmdline
			cmd.Version = tt.appVersion
			cmd.Creation = tt.timestamp
			cmd.GoVersionFunc = func() string {
				return tt.goVersion
			}
			cmd.BuildDependenciesFunc = func() []string {
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
	oldDereferenceEnvVarFunc := cmd.DereferenceEnvVarFunc
	defer func() {
		cmd.DereferenceEnvVarFunc = oldDereferenceEnvVarFunc
	}()
	tests := map[string]struct {
		inputArgs []string
		derefFunc func(string) (string, error)
		want      []string
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
			derefFunc: func(s string) (string, error) {
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
			cmd.DereferenceEnvVarFunc = tt.derefFunc
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
	oldExitFunction := cmd.ExitFunction
	oldNewBusFunc := cmd.NewBusFunc
	oldAppNameSetFunc := cmd.AppNameSetFunc
	oldLogInitFunc := cmd.LogInitFunc
	oldAppPathInitFunc := cmd.AppPathInitFunc
	oldReadConfigFileFunc := cmd.ReadConfigFileFunc
	oldInitBuildDataFunc := cmd.InitBuildDataFunc
	oldFlagIndicatorSetFunc := cmd.FlagIndicatorSetFunc
	oldVersion := cmd.Version
	oldCreation := cmd.Creation
	oldInitialized := cmd.Initialized
	oldBus := cmd.Bus
	oldConfig := cmd.InternalConfig
	defer func() {
		cmd.ExitFunction = oldExitFunction
		cmd.NewBusFunc = oldNewBusFunc
		cmd.AppNameSetFunc = oldAppNameSetFunc
		cmd.LogInitFunc = oldLogInitFunc
		cmd.AppPathInitFunc = oldAppPathInitFunc
		cmd.ReadConfigFileFunc = oldReadConfigFileFunc
		cmd.InitBuildDataFunc = oldInitBuildDataFunc
		cmd.FlagIndicatorSetFunc = oldFlagIndicatorSetFunc
		cmd.Version = oldVersion
		cmd.Creation = oldCreation
		cmd.Initialized = oldInitialized
		cmd.Bus = oldBus
		cmd.InternalConfig = oldConfig
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
		initialize           bool
		exitFunc             func(int)
		wantExitFuncCalled   bool
		wantExitValue        int
		busGetter            func(output.Logger) output.Bus
		appNameSetter        func(string) error
		logInitializer       func(output.Bus) bool
		appPathInitializer   func(output.Bus) bool
		configFileReader     func(output.Bus) (*cmd_toolkit.Configuration, bool)
		wantConfig           *cmd_toolkit.Configuration
		buildDataInitializer func(string, string)
		wantCreation         string
		wantVersion          string
		flagIndicatorSetter  func(string)
		wantFlagIndicator    string
		versionVal           string
		creationVal          string
		wantAppName          string
		output.WantedRecording
	}{
		"already initialized": {
			initialize:    true,
			wantConfig:    cmd_toolkit.EmptyConfiguration(),
			wantExitValue: defaultExitCode,
		},
		"app name set error": {
			initialize: false,
			busGetter: func(output.Logger) output.Bus {
				return o
			},
			appNameSetter: func(string) error {
				return fmt.Errorf("app name could not be set")
			},
			logInitializer: func(_ output.Bus) bool { return false },
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
			busGetter: func(output.Logger) output.Bus {
				return o
			},
			appNameSetter: func(s string) error {
				appNameRecorded = s
				return nil
			},
			logInitializer: func(output.Bus) bool {
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
			busGetter: func(output.Logger) output.Bus {
				return o
			},
			appNameSetter: func(s string) error {
				appNameRecorded = s
				return nil
			},
			logInitializer: func(output.Bus) bool {
				return true
			},
			appPathInitializer: func(output.Bus) bool {
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
			busGetter: func(output.Logger) output.Bus {
				return o
			},
			appNameSetter: func(s string) error {
				appNameRecorded = s
				return nil
			},
			logInitializer: func(output.Bus) bool {
				return true
			},
			appPathInitializer: func(output.Bus) bool {
				return true
			},
			configFileReader: func(output.Bus) (*cmd_toolkit.Configuration, bool) {
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
			busGetter: func(output.Logger) output.Bus {
				return o
			},
			appNameSetter: func(s string) error {
				appNameRecorded = s
				return nil
			},
			logInitializer: func(output.Bus) bool {
				return true
			},
			appPathInitializer: func(output.Bus) bool {
				return true
			},
			configFileReader: func(output.Bus) (*cmd_toolkit.Configuration, bool) {
				return cmd_toolkit.EmptyConfiguration(), true
			},
			creationVal:  "created today",
			wantCreation: "",
			versionVal:   "v0.1.1",
			wantVersion:  "",
			buildDataInitializer: func(v string, c string) {
				versionRecorded = v
				creationRecorded = c
			},
			flagIndicatorSetter: func(s string) { flagIndicatorRecorded = s },
			wantFlagIndicator:   "",
			wantConfig:          cmd_toolkit.EmptyConfiguration(),
			wantExitFuncCalled:  false,
			wantExitValue:       defaultExitCode,
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
			cmd.ExitFunction = tt.exitFunc
			cmd.NewBusFunc = tt.busGetter
			cmd.AppNameSetFunc = tt.appNameSetter
			cmd.LogInitFunc = tt.logInitializer
			cmd.AppPathInitFunc = tt.appPathInitializer
			cmd.ReadConfigFileFunc = tt.configFileReader
			cmd.InitBuildDataFunc = tt.buildDataInitializer
			cmd.FlagIndicatorSetFunc = tt.flagIndicatorSetter
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
