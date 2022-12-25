package main

import (
	"fmt"
	"mp3/internal"
	"mp3/internal/commands"
	"os"
	"path/filepath"
	"reflect"
	"runtime/debug"
	"strings"
	"testing"
	"time"

	"github.com/majohn-r/output"
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
		f           func() (*debug.BuildInfo, bool)
		cmdlineArgs []string
	}
	tests := []struct {
		name string
		args
		wantReturnValue int
		Console         string
		Error           string
		wantLogPrefix   string
		wantLogSuffix   string
	}{
		// TODO [#85] need more tests: explicitly call list, check, repair,
		// postRepair
		{
			name: "failure",
			args: args{
				f: func() (*debug.BuildInfo, bool) {
					return &debug.BuildInfo{
						GoVersion: "go1.x",
					}, true
				},
				cmdlineArgs: []string{"./mp3", "foo"},
			},
			wantReturnValue: 1,
			Error:           "There is no command named \"foo\"; valid commands include [about check export list postRepair repair resetDatabase].\n",
			wantLogPrefix: "level='info' args='[./mp3 foo]' dependencies='[]' goVersion='go1.x' timeStamp='' version='unknown version!' msg='execution starts'\n" +
				fmt.Sprintf("level='info' directory='%s' fileName='defaults.yaml' msg='file does not exist'\n", filepath.Join(thisDir, internal.AppName)) +
				"level='error' command='foo' msg='unrecognized command'\n" +
				"level='info' duration='",
			wantLogSuffix: "' exitCode='1' msg='execution ends'\n",
		},
		{
			name: "success",
			args: args{
				f: func() (*debug.BuildInfo, bool) {
					return &debug.BuildInfo{
						GoVersion: "go1.x",
					}, true
				},
				cmdlineArgs: []string{"./mp3", "-topDir", "./Music"},
			},
			wantReturnValue: 0,
			wantLogPrefix: "level='info' args='[./mp3 -topDir ./Music]' dependencies='[]' goVersion='go1.x' timeStamp='' version='unknown version!' msg='execution starts'\n" +
				fmt.Sprintf("level='info' directory='%s' fileName='defaults.yaml' msg='file does not exist'\n", filepath.Join(thisDir, internal.AppName)) +
				"level='info' -annotate='false' -details='false' -diagnostic='false' -includeAlbums='true' -includeArtists='true' -includeTracks='false' -sort='numeric' command='list' msg='executing command'\n" +
				"level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='./Music' msg='reading filtered music files'\n" +
				"level='info' duration='",
			wantLogSuffix: "' exitCode='0' msg='execution ends'\n",
			Console:       "Artist: myArtist\n  Album: myAlbum\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := output.NewRecorder()
			if gotReturnValue := run(o, tt.args.f, tt.args.cmdlineArgs); gotReturnValue != tt.wantReturnValue {
				t.Errorf("%s = %d, want %d", fnName, gotReturnValue, tt.wantReturnValue)
			}
			if gotErrorOutput := o.ErrorOutput(); gotErrorOutput != tt.Error {
				t.Errorf("%s error output = %q, want %q", fnName, gotErrorOutput, tt.Error)
			}
			if gotConsoleOutput := o.ConsoleOutput(); gotConsoleOutput != tt.Console {
				t.Errorf("%s console output = %q, want %q", fnName, gotConsoleOutput, tt.Console)
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
		args
		output.WantedRecording
	}{
		{name: "success", args: args{returnValue: 0}, WantedRecording: output.WantedRecording{Error: ""}},
		{
			name: "failure",
			args: args{returnValue: 1},
			WantedRecording: output.WantedRecording{
				Error: "\"mp3\" version 1.2.3, created at " + creation + ", failed.\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := output.NewRecorder()
			report(o, tt.args.returnValue)
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
		})
	}
}

func Test_exec(t *testing.T) {
	fnName := "exec()"
	savedVar := internal.SaveEnvVarForTesting("HOMEPATH")
	if err := internal.Mkdir("Music"); err != nil {
		t.Errorf("%s error creating Music: %v", fnName, err)
	}
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, "Music")
		savedVar.RestoreForTesting()
	}()
	thisDir := internal.SecureAbsolutePathForTesting(".")
	if err := internal.Mkdir("Music/myArtist"); err != nil {
		t.Errorf("%s error creating Music/myArtist: %v", fnName, err)
	}
	if err := internal.Mkdir("Music/myArtist/myAlbum"); err != nil {
		t.Errorf("%s error creating Music/myArtist/myAlbum: %v", fnName, err)
	}
	if err := internal.CreateFileForTestingWithContent("Music/myArtist/myAlbum", "01 myTrack.mp3", []byte("no real content")); err != nil {
		t.Errorf("%s error creating Music/myArtist/myAlbum/01 myTrack.mp3: %v", fnName, err)
	}
	os.Setenv("HOMEPATH", thisDir)
	type args struct {
		logInit func(output.Bus) bool
		cmdLine []string
	}
	tests := []struct {
		name string
		args
		wantReturnValue int
	}{
		{
			name: "init logging fails",
			args: args{
				logInit: func(output.Bus) bool {
					return false
				},
			},
			wantReturnValue: 1,
		},
		{
			name: "run fails",
			args: args{
				logInit: func(output.Bus) bool {
					return true
				},
				cmdLine: []string{"mp3", "no-such-command"},
			},
			wantReturnValue: 1,
		},
		{
			name: "success",
			args: args{
				logInit: func(output.Bus) bool {
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

func Test_createBuildData(t *testing.T) {
	fnName := "createBuildData()"
	type args struct {
		f func() (*debug.BuildInfo, bool)
	}
	tests := []struct {
		name string
		args
		want *commands.BuildData
	}{
		{
			name: "happy path",
			args: args{
				f: func() (*debug.BuildInfo, bool) {
					return &debug.BuildInfo{
						GoVersion: "go1.x",
						Deps: []*debug.Module{
							{
								Path:    "blah/foo",
								Version: "v1.1.1",
							},
							{
								Path:    "foo/blah/v2",
								Version: "v2.2.2",
							},
						},
					}, true
				},
			},
			want: &commands.BuildData{
				GoVersion: "go1.x",
				Dependencies: []string{
					"blah/foo v1.1.1",
					"foo/blah/v2 v2.2.2",
				},
			},
		},
		{
			name: "unhappy path",
			args: args{
				f: func() (*debug.BuildInfo, bool) {
					return nil, false
				},
			},
			want: &commands.BuildData{
				GoVersion: "unknown",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := createBuildData(tt.args.f); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}
