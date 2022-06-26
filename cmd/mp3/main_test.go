package main

import (
	"fmt"
	"mp3/internal"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func Test_run(t *testing.T) {
	savedState1 := internal.SaveEnvVarForTesting("APPDATA")
	savedState2 := internal.SaveEnvVarForTesting("HOMEPATH")
	if err := internal.Mkdir("Music"); err != nil {
		t.Errorf("error creating Music: %v", err)
	}
	defer func() {
		internal.DestroyDirectoryForTesting("run()", "Music")
		savedState1.RestoreForTesting()
		savedState2.RestoreForTesting()
	}()
	thisDir := internal.SecureAbsolutePathForTesting(".")
	os.Setenv("APPDATA", thisDir)
	os.Setenv("HOMEPATH", thisDir)
	if err := internal.Mkdir("Music/myArtist"); err != nil {
		t.Errorf("error creating Music/myArtist: %v", err)
	}
	if err := internal.Mkdir("Music/myArtist/myAlbum"); err != nil {
		t.Errorf("error creating Music/myArtist/myAlbum: %v", err)
	}
	if err := internal.CreateFileForTestingWithContent("Music/myArtist/myAlbum", "01 myTrack.mp3", "no real content"); err != nil {
		t.Errorf("error creating Music/myArtist/myAlbum/01 myTrack.mp3: %v", err)
	}
	type args struct {
		cmdlineArgs []string
	}
	tests := []struct {
		name            string
		args            args
		wantReturnValue int
		wantConsole     string
		wantErr         string
		wantLogPrefix   string
		wantLogSuffix   string
	}{
		{
			name:            "failure",
			args:            args{cmdlineArgs: []string{"./mp3", "foo"}},
			wantReturnValue: 1,
			wantErr:         "There is no command named \"foo\"; valid commands include [check ls postRepair repair].\n",
			wantLogPrefix: "level='info' args='[./mp3 foo]' timeStamp='' version='unknown version!' msg='execution starts'\n" +
				fmt.Sprintf("level='info' directory='%s' fileName='defaults.yaml' msg='file does not exist'\n", filepath.Join(thisDir, internal.AppName)) +
				"level='warn' command='foo' msg='unrecognized command'\n"+
				"level='info' duration='",
			wantLogSuffix: "' exitCode='1' msg='execution ends'\n",
		},
		{
			name:            "success",
			args:            args{cmdlineArgs: []string{"./mp3"}},
			wantReturnValue: 0,
			wantLogPrefix: "level='info' args='[./mp3]' timeStamp='' version='unknown version!' msg='execution starts'\n" +
				fmt.Sprintf("level='info' directory='%s' fileName='defaults.yaml' msg='file does not exist'\n", filepath.Join(thisDir, internal.AppName)) +
				"level='info' duration='",
			wantLogSuffix: "' exitCode='0' msg='execution ends'\n",
			wantConsole:   "Artist: myArtist\n  Album: myAlbum\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := internal.NewOutputDeviceForTesting()
			if gotReturnValue := run(o, tt.args.cmdlineArgs); gotReturnValue != tt.wantReturnValue {
				t.Errorf("run() = %v, want %v", gotReturnValue, tt.wantReturnValue)
			}
			if gotErr := o.Stderr(); gotErr != tt.wantErr {
				t.Errorf("run() error output = %v, want %v", gotErr, tt.wantErr)
			}
			if gotOut := o.Stdout(); gotOut != tt.wantConsole {
				t.Errorf("run() console output = %v, want %v", gotOut, tt.wantConsole)
			}
			gotLog := o.LogOutput()
			if !strings.HasPrefix(gotLog, tt.wantLogPrefix) {
				t.Errorf("run() log output %q does not start with %q", gotLog, tt.wantLogPrefix)
			}
			if !strings.HasSuffix(gotLog, tt.wantLogSuffix) {
				t.Errorf("run() log output %q does not end with %q", gotLog, tt.wantLogSuffix)
			}
		})
	}
}

func Test_report(t *testing.T) {
	creation = time.Now().Format(time.RFC3339)
	version = "test"
	type args struct {
		returnValue int
	}
	tests := []struct {
		name    string
		args    args
		wantOut string
		wantErr string
		wantLog string
	}{
		{name: "success", args: args{returnValue: 0}, wantErr: ""},
		{name: "failure", args: args{returnValue: 1}, wantErr: fmt.Sprintf(statusFormat, "mp3", version, creation)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := internal.NewOutputDeviceForTesting()
			report(o, tt.args.returnValue)
			if gotOut := o.Stdout(); gotOut != tt.wantOut {
				t.Errorf("report() console output = %v, want %v", gotOut, tt.wantOut)
			}
			if gotErr := o.Stderr(); gotErr != tt.wantErr {
				t.Errorf("report() error output = %v, want %v", gotErr, tt.wantErr)
			}
			if gotLog := o.LogOutput(); gotLog != tt.wantLog {
				t.Errorf("report() log output = %v, want %v", gotLog, tt.wantLog)
			}
		})
	}
}
