package main

import (
	"fmt"
	"mp3/internal"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"testing"
	"time"

	"github.com/majohn-r/output"
)

func Test_run(t *testing.T) {
	const fnName = "run()"
	savedAppData := internal.SaveEnvVarForTesting("APPDATA")
	savedHomePath := internal.SaveEnvVarForTesting("HOMEPATH")
	if err := internal.Mkdir("Music"); err != nil {
		t.Errorf("%s error creating Music: %v", fnName, err)
	}
	if !internal.InitApplicationPath(output.NewNilBus()) {
		t.Errorf("%s error creating initializing application path", fnName)
	}
	oldAppPath := internal.ApplicationPath()
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, "Music")
		internal.DestroyDirectoryForTesting(fnName, "mp3")
		savedAppData.RestoreForTesting()
		savedHomePath.RestoreForTesting()
		internal.SetApplicationPathForTesting(oldAppPath)
	}()
	here := internal.SecureAbsolutePathForTesting(".")
	os.Setenv("APPDATA", here)
	os.Setenv("HOMEPATH", here)
	internal.InitApplicationPath(output.NewNilBus())
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
	tests := map[string]struct {
		args
		wantReturnValue int
		Console         string
		Error           string
		wantLogPrefix   string
		wantLogSuffix   string
	}{
		// TODO [#85] need more tests: explicitly call list, check, repair,
		// postRepair
		"failure": {
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
				fmt.Sprintf("level='info' directory='%s' fileName='defaults.yaml' msg='file does not exist'\n", filepath.Join(here, internal.AppName)) +
				"level='error' command='foo' msg='unrecognized command'\n" +
				"level='info' duration='",
			wantLogSuffix: "' exitCode='1' msg='execution ends'\n",
		},
		"success": {
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
				fmt.Sprintf("level='info' directory='%s' fileName='defaults.yaml' msg='file does not exist'\n", filepath.Join(here, internal.AppName)) +
				"level='info' -annotate='false' -details='false' -diagnostic='false' -includeAlbums='true' -includeArtists='true' -includeTracks='false' -sort='numeric' command='list' msg='executing command'\n" +
				"level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='./Music' msg='reading filtered music files'\n" +
				"level='info' duration='",
			wantLogSuffix: "' exitCode='0' msg='execution ends'\n",
			Console:       "Artist: myArtist\n  Album: myAlbum\n",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
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
	const fnName = "report()"
	creation = time.Now().Format(time.RFC3339)
	version = "1.2.3"
	type args struct {
		exitCode int
	}
	tests := map[string]struct {
		args
		output.WantedRecording
	}{
		"success": {args: args{exitCode: 0}, WantedRecording: output.WantedRecording{Error: ""}},
		"failure": {
			args: args{exitCode: 1},
			WantedRecording: output.WantedRecording{
				Error: "\"mp3\" version 1.2.3, created at " + creation + ", failed.\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			report(o, tt.args.exitCode)
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
		})
	}
}

func Test_exec(t *testing.T) {
	const fnName = "exec()"
	savedHomePath := internal.SaveEnvVarForTesting("HOMEPATH")
	savedAppData := internal.SaveEnvVarForTesting("APPDATA")
	if err := internal.Mkdir("Music"); err != nil {
		t.Errorf("%s error creating Music: %v", fnName, err)
	}
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, "Music")
		internal.DestroyDirectoryForTesting(fnName, "mp3")
		savedHomePath.RestoreForTesting()
		savedAppData.RestoreForTesting()
	}()
	here := internal.SecureAbsolutePathForTesting(".")
	if err := internal.Mkdir("Music/myArtist"); err != nil {
		t.Errorf("%s error creating Music/myArtist: %v", fnName, err)
	}
	if err := internal.Mkdir("Music/myArtist/myAlbum"); err != nil {
		t.Errorf("%s error creating Music/myArtist/myAlbum: %v", fnName, err)
	}
	if err := internal.CreateFileForTestingWithContent("Music/myArtist/myAlbum", "01 myTrack.mp3", []byte("no real content")); err != nil {
		t.Errorf("%s error creating Music/myArtist/myAlbum/01 myTrack.mp3: %v", fnName, err)
	}
	os.Setenv("HOMEPATH", here)
	type args struct {
		logInit func(output.Bus) bool
		cmdLine []string
	}
	tests := map[string]struct {
		args
		appData string
		want    int
	}{
		"init logging fails": {
			args: args{
				logInit: func(output.Bus) bool {
					return false
				},
			},
			appData: here,
			want:    1,
		},
		"acquisition of application path ": {
			args: args{
				logInit: func(output.Bus) bool {
					return true
				},
			},
			appData: "no such directory",
			want:    1,
		},
		"run fails": {
			args: args{
				logInit: func(output.Bus) bool {
					return true
				},
				cmdLine: []string{"mp3", "no-such-command"},
			},
			appData: here,
			want:    1,
		},
		"success": {
			args: args{
				logInit: func(output.Bus) bool {
					return true
				},
				cmdLine: []string{"mp3"},
			},
			appData: here,
			want:    0,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			os.Setenv("APPDATA", tt.appData)
			if got := exec(tt.args.logInit, tt.args.cmdLine); got != tt.want {
				t.Errorf("%s = %d, want %d", fnName, got, tt.want)
			}
		})
	}
}
