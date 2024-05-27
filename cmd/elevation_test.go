package cmd

import (
	"reflect"
	"testing"

	"github.com/majohn-r/output"
	"golang.org/x/sys/windows"
)

func Test_EnvironmentPermits(t *testing.T) {
	originalLookupEnv := LookupEnv
	defer func() {
		LookupEnv = originalLookupEnv
	}()
	tests := map[string]struct {
		value string
		found bool
		want  bool
	}{
		"no such environment variable": {
			value: "",
			found: false,
			want:  true,
		},
		"variable exists but badly set": {
			value: "foo?",
			found: true,
			want:  true,
		},
		"variable exists and is true": {
			value: "true",
			found: true,
			want:  true,
		},
		"variable exists and is false": {
			value: "false",
			found: true,
			want:  false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			LookupEnv = func(_ string) (string, bool) {
				return tt.value, tt.found
			}
			if got := environmentPermits(); got != tt.want {
				t.Errorf("environmentPermits() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_ProcessIsElevated(t *testing.T) {
	originalGetCurrentProcessToken := GetCurrentProcessToken
	originalIsElevatedFunc := IsElevated
	defer func() {
		GetCurrentProcessToken = originalGetCurrentProcessToken
		IsElevated = originalIsElevatedFunc
	}()
	GetCurrentProcessToken = func() (t windows.Token) {
		return
	}
	tests := map[string]struct {
		want bool
	}{
		"no":  {want: false},
		"yes": {want: true},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			IsElevated = func(_ windows.Token) bool { return tt.want }
			if got := processIsElevated(); got != tt.want {
				t.Errorf("processIsElevated() = %v, want %v", got, tt.want)
			}
		})
	}
}

var (
	ec000               *ElevationControl
	ec001               *ElevationControl
	ec010               *ElevationControl
	ec011               *ElevationControl
	ec100               *ElevationControl
	ec101               *ElevationControl
	ec110               *ElevationControl
	ec111               *ElevationControl
	examplesInitialized = false
)

func initExamples() {
	if !examplesInitialized {
		ec000 = NewElevationControl()
		ec000.adminPermitted = false
		ec000.elevated = false
		ec000.stderrRedirected = false
		ec000.stdinRedirected = false
		ec000.stdoutRedirected = false

		ec001 = NewElevationControl()
		ec001.adminPermitted = false
		ec001.elevated = false
		ec001.stderrRedirected = true
		ec001.stdinRedirected = true
		ec001.stdoutRedirected = true

		ec010 = NewElevationControl()
		ec010.adminPermitted = false
		ec010.elevated = true
		ec010.stderrRedirected = false
		ec010.stdinRedirected = false
		ec010.stdoutRedirected = false

		ec011 = NewElevationControl()
		ec011.adminPermitted = false
		ec011.elevated = true
		ec011.stderrRedirected = true
		ec011.stdinRedirected = true
		ec011.stdoutRedirected = true

		ec100 = NewElevationControl()
		ec100.adminPermitted = true
		ec100.elevated = false
		ec100.stderrRedirected = false
		ec100.stdinRedirected = false
		ec100.stdoutRedirected = false

		ec101 = NewElevationControl()
		ec101.adminPermitted = true
		ec101.elevated = false
		ec101.stderrRedirected = true
		ec101.stdinRedirected = true
		ec101.stdoutRedirected = true

		ec110 = NewElevationControl()
		ec110.adminPermitted = true
		ec110.elevated = true
		ec110.stderrRedirected = false
		ec110.stdinRedirected = false
		ec110.stdoutRedirected = false

		ec111 = NewElevationControl()
		ec111.adminPermitted = true
		ec111.elevated = true
		ec111.stderrRedirected = true
		ec111.stdinRedirected = true
		ec111.stdoutRedirected = true

		examplesInitialized = true
	}
}

func TestElevationGoverner_CanElevate(t *testing.T) {
	initExamples()
	tests := map[string]struct {
		ec   *ElevationControl
		want bool
	}{
		"000": {
			ec:   ec000,
			want: false,
		},
		"001": {
			ec:   ec001,
			want: false,
		},
		"010": {
			ec:   ec010,
			want: false,
		},
		"011": {
			ec:   ec011,
			want: false,
		},
		"100": {
			ec:   ec100,
			want: true,
		},
		"101": {
			ec:   ec101,
			want: false,
		},
		"110": {
			ec:   ec110,
			want: false,
		},
		"111": {
			ec:   ec111,
			want: false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.ec.canElevate(); got != tt.want {
				t.Errorf("ElevationGoverner.canElevate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestElevationControl_ConfigureExit(t *testing.T) {
	initExamples()
	originalExit := Exit
	originalScanf := Scanf
	defer func() {
		Exit = originalExit
		Scanf = originalScanf
	}()
	var scanfCalled bool
	Scanf = func(_ string, _ ...any) (int, error) {
		scanfCalled = true
		return 0, nil
	}
	var exitCalled bool
	Exit = func(_ int) {
		exitCalled = true
	}
	tests := map[string]struct {
		ec              *ElevationControl
		wantExitCalled  bool
		wantScanfCalled bool
	}{
		"not elevated": {
			ec:              ec000,
			wantExitCalled:  true,
			wantScanfCalled: false,
		},
		"elevated": {
			ec:              ec010,
			wantExitCalled:  true,
			wantScanfCalled: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			exitCalled = false
			scanfCalled = false
			Exit = func(_ int) {
				exitCalled = true
			}
			tt.ec.ConfigureExit()
			Exit(0)
			if got := exitCalled; got != tt.wantExitCalled {
				t.Errorf("ElevationControl.ConfigureExit exit called %t, want %t", got, tt.wantExitCalled)
			}
			if got := scanfCalled; got != tt.wantScanfCalled {
				t.Errorf("ElevationControl.ConfigureExit scanf called %t, want %t", got, tt.wantScanfCalled)
			}
		})
	}
}

func Test_mergeArguments(t *testing.T) {
	tests := map[string]struct {
		args []string
		want string
	}{
		"nil": {
			args: nil,
			want: "",
		},
		"empty": {
			args: []string{},
			want: "",
		},
		"one arg": {
			args: []string{"mp3repair"},
			want: "",
		},
		"two args": {
			args: []string{"mp3repair", "list"},
			want: "list",
		},
		"multiple args": {
			args: []string{"mp3repair", "list", "-t", "--byTitle"},
			want: "list -t --byTitle",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := mergeArguments(tt.args); got != tt.want {
				t.Errorf("mergeArguments() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_RunElevated(t *testing.T) {
	originalShellExecute := ShellExecute
	defer func() {
		ShellExecute = originalShellExecute
	}()
	var executed bool
	ShellExecute = func(_ windows.Handle, _, _, _, _ *uint16, _ int32) error {
		executed = true
		return nil
	}
	tests := map[string]struct {
		wantExecuted bool
	}{
		"expected": {wantExecuted: true},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			executed = false
			runElevated()
			if got := executed; got != tt.wantExecuted {
				t.Errorf("runElevated() got %t want %t", got, tt.wantExecuted)
			}
		})
	}
}

func TestElevationControl_WillRunElevated(t *testing.T) {
	initExamples()
	originalShellExecute := ShellExecute
	defer func() {
		ShellExecute = originalShellExecute
	}()
	ShellExecute = func(_ windows.Handle, _, _, _, _ *uint16, _ int32) error {
		return nil
	}
	tests := map[string]struct {
		ec   *ElevationControl
		want bool
	}{
		"000": {ec: ec000, want: false},
		"001": {ec: ec001, want: false},
		"010": {ec: ec010, want: false},
		"011": {ec: ec011, want: false},
		"100": {ec: ec100, want: true},
		"101": {ec: ec101, want: false},
		"110": {ec: ec110, want: false},
		"111": {ec: ec111, want: false},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.ec.WillRunElevated(); got != tt.want {
				t.Errorf("ElevationControl.WillRunElevated() = %t, want %t", got, tt.want)
			}
		})
	}
}

func TestElevationControl_Status(t *testing.T) {
	initExamples()
	tests := map[string]struct {
		ec      *ElevationControl
		appName string
		want    []string
	}{
		"000": {
			ec:      ec000,
			appName: "myApp",
			want: []string{
				"myApp is not running with elevated privileges",
				"The environment variable MP3REPAIR_RUNS_AS_ADMIN evaluates as false",
			},
		},
		"001": {
			ec:      ec001,
			appName: "myApp",
			want: []string{
				"myApp is not running with elevated privileges",
				"stderr, stdin, and stdout have been redirected",
				"The environment variable MP3REPAIR_RUNS_AS_ADMIN evaluates as false",
			},
		},
		"010": {
			ec:      ec010,
			appName: "myApp",
			want: []string{
				"myApp is running with elevated privileges",
			},
		},
		"011": {
			ec:      ec011,
			appName: "myApp",
			want: []string{
				"myApp is running with elevated privileges",
			},
		},
		"100": {
			ec:      ec100,
			appName: "myApp",
			want: []string{
				"myApp is not running with elevated privileges",
			},
		},
		"101": {
			ec:      ec101,
			appName: "myApp",
			want: []string{
				"myApp is not running with elevated privileges",
				"stderr, stdin, and stdout have been redirected",
			},
		},
		"110": {
			ec:      ec110,
			appName: "myApp",
			want: []string{
				"myApp is running with elevated privileges",
			},
		},
		"111": {
			ec:      ec111,
			appName: "myApp",
			want: []string{
				"myApp is running with elevated privileges",
			},
		},
		"stderr redirected": {
			ec: &ElevationControl{
				elevated:         false,
				adminPermitted:   true,
				stderrRedirected: true,
				stdinRedirected:  false,
				stdoutRedirected: false,
			},
			appName: "myApp",
			want: []string{
				"myApp is not running with elevated privileges",
				"stderr has been redirected",
			},
		},
		"stdin redirected": {
			ec: &ElevationControl{
				elevated:         false,
				adminPermitted:   true,
				stderrRedirected: false,
				stdinRedirected:  true,
				stdoutRedirected: false,
			},
			appName: "myApp",
			want: []string{
				"myApp is not running with elevated privileges",
				"stdin has been redirected",
			},
		},
		"stderr and stdin redirected": {
			ec: &ElevationControl{
				elevated:         false,
				adminPermitted:   true,
				stderrRedirected: true,
				stdinRedirected:  true,
				stdoutRedirected: false,
			},
			appName: "myApp",
			want: []string{
				"myApp is not running with elevated privileges",
				"stderr and stdin have been redirected",
			},
		},
		"stdout redirected": {
			ec: &ElevationControl{
				elevated:         false,
				adminPermitted:   true,
				stderrRedirected: false,
				stdinRedirected:  false,
				stdoutRedirected: true,
			},
			appName: "myApp",
			want: []string{
				"myApp is not running with elevated privileges",
				"stdout has been redirected",
			},
		},
		"stderr and stdout redirected": {
			ec: &ElevationControl{
				elevated:         false,
				adminPermitted:   true,
				stderrRedirected: true,
				stdinRedirected:  false,
				stdoutRedirected: true,
			},
			appName: "myApp",
			want: []string{
				"myApp is not running with elevated privileges",
				"stderr and stdout have been redirected",
			},
		},
		"stdin and stdout redirected": {
			ec: &ElevationControl{
				elevated:         false,
				adminPermitted:   true,
				stderrRedirected: false,
				stdinRedirected:  true,
				stdoutRedirected: true,
			},
			appName: "myApp",
			want: []string{
				"myApp is not running with elevated privileges",
				"stdin and stdout have been redirected",
			},
		},
		"stderr, stdin, and stdout redirected": {
			ec: &ElevationControl{
				elevated:         false,
				adminPermitted:   true,
				stderrRedirected: true,
				stdinRedirected:  true,
				stdoutRedirected: true,
			},
			appName: "myApp",
			want: []string{
				"myApp is not running with elevated privileges",
				"stderr, stdin, and stdout have been redirected",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.ec.Status(tt.appName); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ElevationControl.Status() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestElevationControl_Log(t *testing.T) {
	tests := map[string]struct {
		ec *ElevationControl
		output.WantedRecording
	}{
		"00000": {
			ec: &ElevationControl{
				elevated:         false,
				adminPermitted:   false,
				stderrRedirected: false,
				stdinRedirected:  false,
				stdoutRedirected: false,
			},
			WantedRecording: output.WantedRecording{
				Log: "" +
					"level='info'" +
					" admin_permission='false'" +
					" elevated='false'" +
					" stderr_redirected='false'" +
					" stdin_redirected='false'" +
					" stdout_redirected='false'" +
					" msg='elevation state'\n",
			},
		},
		"00001": {
			ec: &ElevationControl{
				elevated:         false,
				adminPermitted:   false,
				stderrRedirected: false,
				stdinRedirected:  false,
				stdoutRedirected: true,
			},
			WantedRecording: output.WantedRecording{
				Log: "" +
					"level='info'" +
					" admin_permission='false'" +
					" elevated='false'" +
					" stderr_redirected='false'" +
					" stdin_redirected='false'" +
					" stdout_redirected='true'" +
					" msg='elevation state'\n",
			},
		},
		"00010": {
			ec: &ElevationControl{
				elevated:         false,
				adminPermitted:   false,
				stderrRedirected: false,
				stdinRedirected:  true,
				stdoutRedirected: false,
			},
			WantedRecording: output.WantedRecording{
				Log: "" +
					"level='info'" +
					" admin_permission='false'" +
					" elevated='false'" +
					" stderr_redirected='false'" +
					" stdin_redirected='true'" +
					" stdout_redirected='false'" +
					" msg='elevation state'\n",
			},
		},
		"00011": {
			ec: &ElevationControl{
				elevated:         false,
				adminPermitted:   false,
				stderrRedirected: false,
				stdinRedirected:  true,
				stdoutRedirected: true,
			},
			WantedRecording: output.WantedRecording{
				Log: "" +
					"level='info'" +
					" admin_permission='false'" +
					" elevated='false'" +
					" stderr_redirected='false'" +
					" stdin_redirected='true'" +
					" stdout_redirected='true'" +
					" msg='elevation state'\n",
			},
		},
		"00100": {
			ec: &ElevationControl{
				elevated:         false,
				adminPermitted:   false,
				stderrRedirected: true,
				stdinRedirected:  false,
				stdoutRedirected: false,
			},
			WantedRecording: output.WantedRecording{
				Log: "" +
					"level='info'" +
					" admin_permission='false'" +
					" elevated='false'" +
					" stderr_redirected='true'" +
					" stdin_redirected='false'" +
					" stdout_redirected='false'" +
					" msg='elevation state'\n",
			},
		},
		"00101": {
			ec: &ElevationControl{
				elevated:         false,
				adminPermitted:   false,
				stderrRedirected: true,
				stdinRedirected:  false,
				stdoutRedirected: true,
			},
			WantedRecording: output.WantedRecording{
				Log: "" +
					"level='info'" +
					" admin_permission='false'" +
					" elevated='false'" +
					" stderr_redirected='true'" +
					" stdin_redirected='false'" +
					" stdout_redirected='true'" +
					" msg='elevation state'\n",
			},
		},
		"00110": {
			ec: &ElevationControl{
				elevated:         false,
				adminPermitted:   false,
				stderrRedirected: true,
				stdinRedirected:  true,
				stdoutRedirected: false,
			},
			WantedRecording: output.WantedRecording{
				Log: "" +
					"level='info'" +
					" admin_permission='false'" +
					" elevated='false'" +
					" stderr_redirected='true'" +
					" stdin_redirected='true'" +
					" stdout_redirected='false'" +
					" msg='elevation state'\n",
			},
		},
		"00111": {
			ec: &ElevationControl{
				elevated:         false,
				adminPermitted:   false,
				stderrRedirected: true,
				stdinRedirected:  true,
				stdoutRedirected: true,
			},
			WantedRecording: output.WantedRecording{
				Log: "" +
					"level='info'" +
					" admin_permission='false'" +
					" elevated='false'" +
					" stderr_redirected='true'" +
					" stdin_redirected='true'" +
					" stdout_redirected='true'" +
					" msg='elevation state'\n",
			},
		},
		"01000": {
			ec: &ElevationControl{
				elevated:         false,
				adminPermitted:   true,
				stderrRedirected: false,
				stdinRedirected:  false,
				stdoutRedirected: false,
			},
			WantedRecording: output.WantedRecording{
				Log: "" +
					"level='info'" +
					" admin_permission='true'" +
					" elevated='false'" +
					" stderr_redirected='false'" +
					" stdin_redirected='false'" +
					" stdout_redirected='false'" +
					" msg='elevation state'\n",
			},
		},
		"01001": {
			ec: &ElevationControl{
				elevated:         false,
				adminPermitted:   true,
				stderrRedirected: false,
				stdinRedirected:  false,
				stdoutRedirected: true,
			},
			WantedRecording: output.WantedRecording{
				Log: "" +
					"level='info'" +
					" admin_permission='true'" +
					" elevated='false'" +
					" stderr_redirected='false'" +
					" stdin_redirected='false'" +
					" stdout_redirected='true'" +
					" msg='elevation state'\n",
			},
		},
		"01010": {
			ec: &ElevationControl{
				elevated:         false,
				adminPermitted:   true,
				stderrRedirected: false,
				stdinRedirected:  true,
				stdoutRedirected: false,
			},
			WantedRecording: output.WantedRecording{
				Log: "" +
					"level='info'" +
					" admin_permission='true'" +
					" elevated='false'" +
					" stderr_redirected='false'" +
					" stdin_redirected='true'" +
					" stdout_redirected='false'" +
					" msg='elevation state'\n",
			},
		},
		"01011": {
			ec: &ElevationControl{
				elevated:         false,
				adminPermitted:   true,
				stderrRedirected: false,
				stdinRedirected:  true,
				stdoutRedirected: true,
			},
			WantedRecording: output.WantedRecording{
				Log: "" +
					"level='info'" +
					" admin_permission='true'" +
					" elevated='false'" +
					" stderr_redirected='false'" +
					" stdin_redirected='true'" +
					" stdout_redirected='true'" +
					" msg='elevation state'\n",
			},
		},
		"01100": {
			ec: &ElevationControl{
				elevated:         false,
				adminPermitted:   true,
				stderrRedirected: true,
				stdinRedirected:  false,
				stdoutRedirected: false,
			},
			WantedRecording: output.WantedRecording{
				Log: "" +
					"level='info'" +
					" admin_permission='true'" +
					" elevated='false'" +
					" stderr_redirected='true'" +
					" stdin_redirected='false'" +
					" stdout_redirected='false'" +
					" msg='elevation state'\n",
			},
		},
		"01101": {
			ec: &ElevationControl{
				elevated:         false,
				adminPermitted:   true,
				stderrRedirected: true,
				stdinRedirected:  false,
				stdoutRedirected: true,
			},
			WantedRecording: output.WantedRecording{
				Log: "" +
					"level='info'" +
					" admin_permission='true'" +
					" elevated='false'" +
					" stderr_redirected='true'" +
					" stdin_redirected='false'" +
					" stdout_redirected='true'" +
					" msg='elevation state'\n",
			},
		},
		"01110": {
			ec: &ElevationControl{
				elevated:         false,
				adminPermitted:   true,
				stderrRedirected: true,
				stdinRedirected:  true,
				stdoutRedirected: false,
			},
			WantedRecording: output.WantedRecording{
				Log: "" +
					"level='info'" +
					" admin_permission='true'" +
					" elevated='false'" +
					" stderr_redirected='true'" +
					" stdin_redirected='true'" +
					" stdout_redirected='false'" +
					" msg='elevation state'\n",
			},
		},
		"01111": {
			ec: &ElevationControl{
				elevated:         false,
				adminPermitted:   true,
				stderrRedirected: true,
				stdinRedirected:  true,
				stdoutRedirected: true,
			},
			WantedRecording: output.WantedRecording{
				Log: "" +
					"level='info'" +
					" admin_permission='true'" +
					" elevated='false'" +
					" stderr_redirected='true'" +
					" stdin_redirected='true'" +
					" stdout_redirected='true'" +
					" msg='elevation state'\n",
			},
		},
		"10000": {
			ec: &ElevationControl{
				elevated:         true,
				adminPermitted:   false,
				stderrRedirected: false,
				stdinRedirected:  false,
				stdoutRedirected: false,
			},
			WantedRecording: output.WantedRecording{
				Log: "" +
					"level='info'" +
					" admin_permission='false'" +
					" elevated='true'" +
					" stderr_redirected='false'" +
					" stdin_redirected='false'" +
					" stdout_redirected='false'" +
					" msg='elevation state'\n",
			},
		},
		"10001": {
			ec: &ElevationControl{
				elevated:         true,
				adminPermitted:   false,
				stderrRedirected: false,
				stdinRedirected:  false,
				stdoutRedirected: true,
			},
			WantedRecording: output.WantedRecording{
				Log: "" +
					"level='info'" +
					" admin_permission='false'" +
					" elevated='true'" +
					" stderr_redirected='false'" +
					" stdin_redirected='false'" +
					" stdout_redirected='true'" +
					" msg='elevation state'\n",
			},
		},
		"10010": {
			ec: &ElevationControl{
				elevated:         true,
				adminPermitted:   false,
				stderrRedirected: false,
				stdinRedirected:  true,
				stdoutRedirected: false,
			},
			WantedRecording: output.WantedRecording{
				Log: "" +
					"level='info'" +
					" admin_permission='false'" +
					" elevated='true'" +
					" stderr_redirected='false'" +
					" stdin_redirected='true'" +
					" stdout_redirected='false'" +
					" msg='elevation state'\n",
			},
		},
		"10011": {
			ec: &ElevationControl{
				elevated:         true,
				adminPermitted:   false,
				stderrRedirected: false,
				stdinRedirected:  true,
				stdoutRedirected: true,
			},
			WantedRecording: output.WantedRecording{
				Log: "" +
					"level='info'" +
					" admin_permission='false'" +
					" elevated='true'" +
					" stderr_redirected='false'" +
					" stdin_redirected='true'" +
					" stdout_redirected='true'" +
					" msg='elevation state'\n",
			},
		},
		"10100": {
			ec: &ElevationControl{
				elevated:         true,
				adminPermitted:   false,
				stderrRedirected: true,
				stdinRedirected:  false,
				stdoutRedirected: false,
			},
			WantedRecording: output.WantedRecording{
				Log: "" +
					"level='info'" +
					" admin_permission='false'" +
					" elevated='true'" +
					" stderr_redirected='true'" +
					" stdin_redirected='false'" +
					" stdout_redirected='false'" +
					" msg='elevation state'\n",
			},
		},
		"10101": {
			ec: &ElevationControl{
				elevated:         true,
				adminPermitted:   false,
				stderrRedirected: true,
				stdinRedirected:  false,
				stdoutRedirected: true,
			},
			WantedRecording: output.WantedRecording{
				Log: "" +
					"level='info'" +
					" admin_permission='false'" +
					" elevated='true'" +
					" stderr_redirected='true'" +
					" stdin_redirected='false'" +
					" stdout_redirected='true'" +
					" msg='elevation state'\n",
			},
		},
		"10110": {
			ec: &ElevationControl{
				elevated:         true,
				adminPermitted:   false,
				stderrRedirected: true,
				stdinRedirected:  true,
				stdoutRedirected: false,
			},
			WantedRecording: output.WantedRecording{
				Log: "" +
					"level='info'" +
					" admin_permission='false'" +
					" elevated='true'" +
					" stderr_redirected='true'" +
					" stdin_redirected='true'" +
					" stdout_redirected='false'" +
					" msg='elevation state'\n",
			},
		},
		"10111": {
			ec: &ElevationControl{
				elevated:         true,
				adminPermitted:   false,
				stderrRedirected: true,
				stdinRedirected:  true,
				stdoutRedirected: true,
			},
			WantedRecording: output.WantedRecording{
				Log: "" +
					"level='info'" +
					" admin_permission='false'" +
					" elevated='true'" +
					" stderr_redirected='true'" +
					" stdin_redirected='true'" +
					" stdout_redirected='true'" +
					" msg='elevation state'\n",
			},
		},
		"11000": {
			ec: &ElevationControl{
				elevated:         true,
				adminPermitted:   true,
				stderrRedirected: false,
				stdinRedirected:  false,
				stdoutRedirected: false,
			},
			WantedRecording: output.WantedRecording{
				Log: "" +
					"level='info'" +
					" admin_permission='true'" +
					" elevated='true'" +
					" stderr_redirected='false'" +
					" stdin_redirected='false'" +
					" stdout_redirected='false'" +
					" msg='elevation state'\n",
			},
		},
		"11001": {
			ec: &ElevationControl{
				elevated:         true,
				adminPermitted:   true,
				stderrRedirected: false,
				stdinRedirected:  false,
				stdoutRedirected: true,
			},
			WantedRecording: output.WantedRecording{
				Log: "" +
					"level='info'" +
					" admin_permission='true'" +
					" elevated='true'" +
					" stderr_redirected='false'" +
					" stdin_redirected='false'" +
					" stdout_redirected='true'" +
					" msg='elevation state'\n",
			},
		},
		"11010": {
			ec: &ElevationControl{
				elevated:         true,
				adminPermitted:   true,
				stderrRedirected: false,
				stdinRedirected:  true,
				stdoutRedirected: false,
			},
			WantedRecording: output.WantedRecording{
				Log: "" +
					"level='info'" +
					" admin_permission='true'" +
					" elevated='true'" +
					" stderr_redirected='false'" +
					" stdin_redirected='true'" +
					" stdout_redirected='false'" +
					" msg='elevation state'\n",
			},
		},
		"11011": {
			ec: &ElevationControl{
				elevated:         true,
				adminPermitted:   true,
				stderrRedirected: false,
				stdinRedirected:  true,
				stdoutRedirected: true,
			},
			WantedRecording: output.WantedRecording{
				Log: "" +
					"level='info'" +
					" admin_permission='true'" +
					" elevated='true'" +
					" stderr_redirected='false'" +
					" stdin_redirected='true'" +
					" stdout_redirected='true'" +
					" msg='elevation state'\n",
			},
		},
		"11100": {
			ec: &ElevationControl{
				elevated:         true,
				adminPermitted:   true,
				stderrRedirected: true,
				stdinRedirected:  false,
				stdoutRedirected: false,
			},
			WantedRecording: output.WantedRecording{
				Log: "" +
					"level='info'" +
					" admin_permission='true'" +
					" elevated='true'" +
					" stderr_redirected='true'" +
					" stdin_redirected='false'" +
					" stdout_redirected='false'" +
					" msg='elevation state'\n",
			},
		},
		"11101": {
			ec: &ElevationControl{
				elevated:         true,
				adminPermitted:   true,
				stderrRedirected: true,
				stdinRedirected:  false,
				stdoutRedirected: true,
			},
			WantedRecording: output.WantedRecording{
				Log: "" +
					"level='info'" +
					" admin_permission='true'" +
					" elevated='true'" +
					" stderr_redirected='true'" +
					" stdin_redirected='false'" +
					" stdout_redirected='true'" +
					" msg='elevation state'\n",
			},
		},
		"11110": {
			ec: &ElevationControl{
				elevated:         true,
				adminPermitted:   true,
				stderrRedirected: true,
				stdinRedirected:  true,
				stdoutRedirected: false,
			},
			WantedRecording: output.WantedRecording{
				Log: "" +
					"level='info'" +
					" admin_permission='true'" +
					" elevated='true'" +
					" stderr_redirected='true'" +
					" stdin_redirected='true'" +
					" stdout_redirected='false'" +
					" msg='elevation state'\n",
			},
		},
		"11111": {
			ec: &ElevationControl{
				elevated:         true,
				adminPermitted:   true,
				stderrRedirected: true,
				stdinRedirected:  true,
				stdoutRedirected: true,
			},
			WantedRecording: output.WantedRecording{
				Log: "" +
					"level='info'" +
					" admin_permission='true'" +
					" elevated='true'" +
					" stderr_redirected='true'" +
					" stdin_redirected='true'" +
					" stdout_redirected='true'" +
					" msg='elevation state'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			tt.ec.Log(o, output.Info)
			o.Report(t, "ElevationControl.Log()", tt.WantedRecording)
		})
	}
}
