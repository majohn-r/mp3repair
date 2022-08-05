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
	fnName := "run()"
	savedState1 := internal.SaveEnvVarForTesting("APPDATA")
	savedState2 := internal.SaveEnvVarForTesting("HOMEPATH")
	if err := internal.Mkdir("Music"); err != nil {
		t.Errorf("%s error creating Music: %v", fnName, err)
	}
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, "Music")
		savedState1.RestoreForTesting()
		savedState2.RestoreForTesting()
	}()
	thisDir := internal.SecureAbsolutePathForTesting(".")
	os.Setenv("APPDATA", thisDir)
	os.Setenv("HOMEPATH", thisDir)
	if err := internal.Mkdir("Music/myArtist"); err != nil {
		t.Errorf("%s error creating Music/myArtist: %v", fnName, err)
	}
	if err := internal.Mkdir("Music/myArtist/myAlbum"); err != nil {
		t.Errorf("%s error creating Music/myArtist/myAlbum: %v", fnName, err)
	}
	if err := internal.CreateFileForTestingWithContent("Music/myArtist/myAlbum", "01 myTrack.mp3", []byte("no real content")); err != nil {
		t.Errorf("%s error creating Music/myArtist/myAlbum/01 myTrack.mp3: %v", fnName, err)
	}
	type args struct {
		cmdlineArgs []string
	}
	tests := []struct {
		name              string
		args              args
		wantReturnValue   int
		wantConsoleOutput string
		wantErrorOutput   string
		wantLogPrefix     string
		wantLogSuffix     string
	}{
		// TODO [#85] need more tests: explicitly call ls, check, repair, postRepair
		{
			name:            "failure",
			args:            args{cmdlineArgs: []string{"./mp3", "foo"}},
			wantReturnValue: 1,
			wantErrorOutput: "There is no command named \"foo\"; valid commands include [check ls postRepair repair resetDatabase].\n",
			wantLogPrefix: "level='info' args='[./mp3 foo]' timeStamp='' version='unknown version!' msg='execution starts'\n" +
				fmt.Sprintf("level='info' directory='%s' fileName='defaults.yaml' msg='file does not exist'\n", filepath.Join(thisDir, internal.AppName)) +
				"level='error' command='foo' msg='unrecognized command'\n" +
				"level='info' duration='",
			wantLogSuffix: "' exitCode='1' msg='execution ends'\n",
		},
		{
			name:            "success",
			args:            args{cmdlineArgs: []string{"./mp3", "-topDir", "./Music"}},
			wantReturnValue: 0,
			wantLogPrefix: "level='info' args='[./mp3 -topDir ./Music]' timeStamp='' version='unknown version!' msg='execution starts'\n" +
				fmt.Sprintf("level='info' directory='%s' fileName='defaults.yaml' msg='file does not exist'\n", filepath.Join(thisDir, internal.AppName)) +
				"level='info' -annotate='false' -diagnostic='false' -includeAlbums='true' -includeArtists='true' -includeTracks='false' -sort='numeric' command='ls' msg='executing command'\n" +
				"level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='./Music' msg='reading filtered music files'\n" +
				"level='info' duration='",
			wantLogSuffix:     "' exitCode='0' msg='execution ends'\n",
			wantConsoleOutput: "Artist: myArtist\n  Album: myAlbum\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := internal.NewOutputDeviceForTesting()
			if gotReturnValue := run(o, tt.args.cmdlineArgs); gotReturnValue != tt.wantReturnValue {
				t.Errorf("%s = %d, want %d", fnName, gotReturnValue, tt.wantReturnValue)
			}
			if gotErrorOutput := o.ErrorOutput(); gotErrorOutput != tt.wantErrorOutput {
				t.Errorf("%s error output = %q, want %q", fnName, gotErrorOutput, tt.wantErrorOutput)
			}
			if gotConsoleOutput := o.ConsoleOutput(); gotConsoleOutput != tt.wantConsoleOutput {
				t.Errorf("%s console output = %q, want %q", fnName, gotConsoleOutput, tt.wantConsoleOutput)
			}
			gotLog := o.LogOutput()
			if !strings.HasPrefix(gotLog, tt.wantLogPrefix) {
				t.Errorf("%s log output %q does not start with %q", fnName, gotLog, tt.wantLogPrefix)
			}
			if !strings.HasSuffix(gotLog, tt.wantLogSuffix) {
				t.Errorf("%s log output %q does not end with %q", fnName, gotLog, tt.wantLogSuffix)
			}
		})
	}
}

func Test_report(t *testing.T) {
	fnName := "report()"
	creation = time.Now().Format(time.RFC3339)
	version = "1.2.3"
	type args struct {
		returnValue int
	}
	tests := []struct {
		name string
		args args
		internal.WantedOutput
	}{
		{name: "success", args: args{returnValue: 0}, WantedOutput: internal.WantedOutput{WantErrorOutput: ""}},
		{
			name: "failure",
			args: args{returnValue: 1},
			WantedOutput: internal.WantedOutput{
				WantErrorOutput: "\"mp3\" version 1.2.3, created at " + creation + ", failed.\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := internal.NewOutputDeviceForTesting()
			report(o, tt.args.returnValue)
			if issues, ok := o.CheckOutput(tt.WantedOutput); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
		})
	}
}

func Test_exec(t *testing.T) {
	fnName := "exec()"
	type args struct {
		logInit func(internal.OutputBus) bool
		cmdLine []string
	}
	tests := []struct {
		name            string
		args            args
		wantReturnValue int
	}{
		{
			name: "init logging fails",
			args: args{
				logInit: func(internal.OutputBus) bool {
					return false
				},
			},
			wantReturnValue: 1,
		},
		{
			name: "run fails",
			args: args{
				logInit: func(internal.OutputBus) bool {
					return true
				},
				cmdLine: []string{"mp3", "no-such-command"},
			},
			wantReturnValue: 1,
		},
		{
			name: "success",
			args: args{
				logInit: func(internal.OutputBus) bool {
					return true
				},
				cmdLine: []string{"mp3"},
			},
			wantReturnValue: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotReturnValue := exec(tt.args.logInit, tt.args.cmdLine); gotReturnValue != tt.wantReturnValue {
				t.Errorf("%s = %d, want %d", fnName, gotReturnValue, tt.wantReturnValue)
			}
		})
	}
}
