/*
Copyright © 2026 Marc Johnson (marc.johnson27591@gmail.com)
*/
package cmd

import (
	"fmt"
	"os"
	"reflect"
	"runtime/debug"
	"testing"
	"time"

	cmdtoolkit "github.com/majohn-r/cmd-toolkit"
	"github.com/majohn-r/output"
)

func Test_execute(t *testing.T) {
	initGlobals()
	originalArgs := os.Args
	originalExit := Exit
	originalBus := bus
	defer func() {
		os.Args = originalArgs
		Exit = originalExit
		bus = originalBus
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
			Exit = func(code int) {
				got = code
			}
			Execute()
			if got != 0 {
				t.Errorf("execute: got exit code %d, wanted 0", got)
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

type localBuildInfo struct {
	goVersion    string
	mainVersion  string
	dependencies []string
	settings     []string
}

func (lib *localBuildInfo) GoVersion() string {
	return lib.goVersion
}

func (lib *localBuildInfo) MainVersion() string {
	return lib.mainVersion
}

func (lib *localBuildInfo) Dependencies() []string {
	return lib.dependencies
}

func (lib *localBuildInfo) Settings() []string {
	return lib.settings
}

func Test_runMain(t *testing.T) {
	originalArgs := os.Args
	originalSince := since
	originalVersion := version
	originalCreation := creation
	originalCachedGoVersion := cachedGoVersion
	originalCachedBuildDependencies := cachedBuildDependencies
	originalGetBuildData := getBuildData
	originalMP3repairElevationControl := mp3repairElevationControl
	originalApplicationName := applicationName
	defer func() {
		since = originalSince
		os.Args = originalArgs
		version = originalVersion
		creation = originalCreation
		cachedGoVersion = originalCachedGoVersion
		cachedBuildDependencies = originalCachedBuildDependencies
		mp3repairElevationControl = originalMP3repairElevationControl
		getBuildData = originalGetBuildData
		applicationName = originalApplicationName
	}()
	applicationName = "mp3repair"
	mp3repairElevationControl = testingElevationControl{
		logFields: map[string]any{
			"elevated":             true,
			"admin_permission":     true,
			"stderr_redirected":    false,
			"stdin_redirected":     false,
			"stdout_redirected":    false,
			"environment_variable": ElevatedPrivilegesPermissionVar,
		},
	}
	const startLog = "level='info'" +
		" args='[arg1 arg2]'" +
		" buildSettings='[-ldflags: -X main.version=0.45.0 cmd: gcc git: 2.3.4]'" +
		" defaults='" +
		"about:\n" +
		"    style: rounded\n" +
		"export:\n" +
		"    defaults: false\n" +
		"    overwrite: false\n" +
		"io:\n" +
		"    maxOpenFiles: 1000\n" +
		"list:\n" +
		"    albums: false\n" +
		"    annotate: false\n" +
		"    artists: false\n" +
		"    byNumber: false\n" +
		"    byTitle: false\n" +
		"    diagnostic: false\n" +
		"    tracks: false\n" +
		"resetDatabase:\n" +
		"    force: false\n" +
		"    ignoreServiceErrors: false\n" +
		"    timeout: 10\n" +
		"rewrite:\n" +
		"    dryRun: false\n" +
		"scan:\n" +
		"    empty: false\n" +
		"    files: false\n" +
		"    numbering: false\n" +
		"search:\n" +
		"    albumFilter: .*\n" +
		"    artistFilter: .*\n" +
		"    extensions: .mp3\n" +
		"    trackFilter: .*\n'" +
		" dependencies='[foo v1.1.1 bar v1.2.2]'" +
		" goVersion='1.22.x'" +
		" mainVersion='0.45.0'" +
		" timeStamp='2021-11-28T12:01:02Z05:00'" +
		" version='0.1.2'" +
		" msg='execution starts'\n"
	type args struct {
		cmd   commandExecutor
		start time.Time
	}
	tests := map[string]struct {
		args
		cmdline      []string
		appVersion   string
		timestamp    string
		goVersion    string
		mainVersion  string
		dependencies []string
		settings     []string
		output.WantedRecording
	}{
		"happy": {
			args:         args{cmd: happyCommand{}, start: time.Now()},
			cmdline:      []string{"happyApp", "arg1", "arg2"},
			appVersion:   "0.1.2",
			timestamp:    "2021-11-28T12:01:02Z05:00",
			goVersion:    "1.22.x",
			mainVersion:  "0.45.0",
			settings:     []string{"-ldflags: -X main.version=0.45.0", "cmd: gcc", "git: 2.3.4"},
			dependencies: []string{"foo v1.1.1", "bar v1.2.2"},
			WantedRecording: output.WantedRecording{
				Log: startLog +
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
			appVersion:   "0.1.2",
			timestamp:    "2021-11-28T12:01:02Z05:00",
			cmdline:      []string{"sadApp", "arg1", "arg2"},
			goVersion:    "1.22.x",
			mainVersion:  "0.45.0",
			settings:     []string{"-ldflags: -X main.version=0.45.0", "cmd: gcc", "git: 2.3.4"},
			dependencies: []string{"foo v1.1.1", "bar v1.2.2"},
			WantedRecording: output.WantedRecording{
				Error: "" +
					"\"mp3repair\" version 0.1.2, created at 2021-11-28T12:01:02Z05:00, failed.\n",
				Log: startLog +
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
			appVersion:   "0.1.2",
			timestamp:    "2021-11-28T12:01:02Z05:00",
			cmdline:      []string{"sadApp", "arg1", "arg2"},
			goVersion:    "1.22.x",
			mainVersion:  "0.45.0",
			settings:     []string{"-ldflags: -X main.version=0.45.0", "cmd: gcc", "git: 2.3.4"},
			dependencies: []string{"foo v1.1.1", "bar v1.2.2"},
			WantedRecording: output.WantedRecording{
				Error: "" +
					"A runtime error occurred: \"oh dear\".\n",
				Log: startLog +
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
		since = func(_ time.Time) time.Duration {
			return 0
		}
		t.Run(name, func(t *testing.T) {
			os.Args = tt.cmdline
			version = tt.appVersion
			creation = tt.timestamp
			getBuildData = func(buildInfoReader func() (*debug.BuildInfo, bool)) cmdtoolkit.BuildInformation {
				return &localBuildInfo{
					goVersion:    tt.goVersion,
					dependencies: tt.dependencies,
					mainVersion:  tt.mainVersion,
					settings:     tt.settings,
				}
			}
			o := output.NewRecorder()
			runMain(o, tt.args.cmd, tt.args.start)
			o.Report(t, "runMain()", tt.WantedRecording)
		})
	}
}

func Test_cookCommandLineArguments(t *testing.T) {
	originalDereferenceEnvVar := dereferenceEnvVar
	defer func() {
		dereferenceEnvVar = originalDereferenceEnvVar
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
				}
				return s, nil
			},
			want: []string{"foo", "bar"},
			WantedRecording: output.WantedRecording{
				Error: "An error was found in processing argument \"%arg%\": 'dereference service dead'.\n",
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
			dereferenceEnvVar = tt.dereferenceEnvVar
			got := cookCommandLineArguments(o, tt.inputArgs)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("cookCommandLineArguments() = %v, want %v", got, tt.want)
			}
			o.Report(t, "cookCommandLineArguments()", tt.WantedRecording)
		})
	}
}

func Test_initGlobals(t *testing.T) {
	originalExit := Exit
	originalNewDefaultBus := newDefaultBus
	originalInitLogging := initLogging
	originalInitApplicationPath := initApplicationPath
	originalReadConfigurationFile := readDefaultsConfigFile
	originalVersion := version
	originalCreation := creation
	originalInitialized := initialized
	originalBus := bus
	originalInternalConfig := internalConfig
	originalGetPid := getPid
	originalGetPpid := getPpid
	defer func() {
		Exit = originalExit
		newDefaultBus = originalNewDefaultBus
		initLogging = originalInitLogging
		initApplicationPath = originalInitApplicationPath
		readDefaultsConfigFile = originalReadConfigurationFile
		version = originalVersion
		creation = originalCreation
		initialized = originalInitialized
		bus = originalBus
		internalConfig = originalInternalConfig
		getPid = originalGetPid
		getPpid = originalGetPpid
	}()
	getPid = func() int { return 12345 }
	getPpid = func() int { return 67890 }
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
			WantedRecording: output.WantedRecording{
				Log: "" +
					"level='info'" +
					" parent_process_id='67890'" +
					" process_id='12345'" +
					" msg='process information'\n",
			},
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
			WantedRecording: output.WantedRecording{
				Log: "" +
					"level='info'" +
					" parent_process_id='67890'" +
					" process_id='12345'" +
					" msg='process information'\n",
			},
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
			WantedRecording: output.WantedRecording{
				Log: "" +
					"level='info'" +
					" parent_process_id='67890'" +
					" process_id='12345'" +
					" msg='process information'\n",
			},
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
			WantedRecording: output.WantedRecording{
				Log: "" +
					"level='info'" +
					" parent_process_id='67890'" +
					" process_id='12345'" +
					" msg='process information'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o = output.NewRecorder()
			internalConfig = cmdtoolkit.EmptyConfiguration()
			ExitFunctionCalled = defaultExitFunctionCalled
			exitCodeRecorded = defaultExitCode
			creationRecorded = defaultCreation
			versionRecorded = defaultVersion
			flagIndicatorRecorded = defaultFlagIndicator
			initialized = tt.initialize
			Exit = tt.exitFunc
			newDefaultBus = tt.newDefaultBus
			initLogging = tt.initLogging
			initApplicationPath = tt.initApplicationPath
			readDefaultsConfigFile = tt.readConfigurationFile

			creation = tt.creationVal
			version = tt.versionVal
			initGlobals()
			if got := internalConfig; !reflect.DeepEqual(got, tt.wantConfig) {
				t.Errorf("initGlobals: _c got %v want %v", got, tt.wantConfig)
			}
			if got := ExitFunctionCalled; got != tt.wantExitFuncCalled {
				t.Errorf("initGlobals: exit called got %t want %t", got, tt.wantExitFuncCalled)
			}
			if got := exitCodeRecorded; got != tt.wantExitValue {
				t.Errorf("initGlobals: exit code got %d want %d", got, tt.wantExitValue)
			}
			if got := creationRecorded; got != tt.wantCreation {
				t.Errorf("initGlobals: creation got %q want %q", got, tt.wantCreation)
			}
			if got := versionRecorded; got != tt.wantVersion {
				t.Errorf("initGlobals: version got %q want %q", got, tt.wantVersion)
			}
			if got := flagIndicatorRecorded; got != tt.wantFlagIndicator {
				t.Errorf("initGlobals: flag indicator got %q want %q", got, tt.wantFlagIndicator)
			}
			o.Report(t, "initGlobals()", tt.WantedRecording)
		})
	}
}

func Test_root_Usage(t *testing.T) {
	originalApplicationName := applicationName
	defer func() { applicationName = originalApplicationName }()
	applicationName = "mp3repair"
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
					"Then scan for problems in the track metadata:\n" +
					"\n" +
					"mp3repair scan --files\n" +
					"\n" +
					"If problems were found, rewrite the mp3 files:\n" +
					"\n" +
					"mp3repair rewrite\n" +
					"The rewrite command creates backup files for each track it rewrites. After\n" +
					"listening to the files that have been rewritten (verifying that the rewrite\n" +
					"process did not corrupt the audio), clean up those backups:\n" +
					"\n" +
					"mp3repair cleanup\n" +
					"\n" +
					"After rewriting the mp3 files, the Windows Media Player database may be out of\n" +
					"sync with the changes. While the database will eventually catch up, accelerate\n" +
					"the process:\n" +
					"\n" +
					"mp3repair resetDatabase\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			command := cloneCommand(rootCmd)
			command.Example = genExample(applicationName)
			enableCommandRecording(o, command)
			_ = command.Usage()
			o.Report(t, "root Usage()", tt.WantedRecording)
		})
	}
}

func Test_obtainExitCode(t *testing.T) {
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
			if got := obtainExitCode(tt.err); got != tt.want {
				t.Errorf("obtainExitCode() = %v, want %v", got, tt.want)
			}
		})
	}
}
