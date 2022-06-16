package internal

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func Test_findReferences(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{name: "no references", args: args{s: ".mp3"}, want: nil},
		{
			name: "lots of references",
			args: args{s: "$PATH/$SUBPATH/%FILENAME%.%EXTENSION%"},
			want: []string{"$SUBPATH", "$PATH", "%FILENAME%", "%EXTENSION%"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := findReferences(tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("findReferences() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInterpretEnvVarReferences(t *testing.T) {
	originalExtension := os.Getenv("EXTENSION")
	originalFileName := os.Getenv("FILENAME")
	originalPath := os.Getenv("PATH")
	originalSubPath := os.Getenv("SUBPATH")
	defer func() {
		os.Setenv("EXTENSION", originalExtension)
		os.Setenv("FILENAME", originalFileName)
		os.Setenv("PATH", originalPath)
		os.Setenv("SUBPATH", originalSubPath)
	}()
	newExtension := "mp3"
	newFileName := "track"
	newPath := "/c/Users/MyUser"
	newSubPath := "Music"
	os.Setenv("EXTENSION", newExtension)
	os.Setenv("FILENAME", newFileName)
	os.Setenv("PATH", newPath)
	os.Setenv("SUBPATH", newSubPath)
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "no references", args: args{s: "no references"}, want: "no references"},
		{
			name: "lots of references",
			args: args{s: "$PATH/$SUBPATH/%FILENAME%.%EXTENSION%"},
			want: "/c/Users/MyUser/Music/track.mp3",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InterpretEnvVarReferences(tt.args.s); got != tt.want {
				t.Errorf("InterpretEnvVarReferences() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateAppSpecificPath(t *testing.T) {
	type args struct {
		topDir string
	}
	tests := []struct {
		name string
		args args
		want string
	}{{name: "simple test", args: args{topDir: "top"}, want: filepath.Join("top", appName)}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CreateAppSpecificPath(tt.args.topDir); got != tt.want {
				t.Errorf("CreateAppSpecificPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLookupEnvVars(t *testing.T) {
	type envState struct {
		varName  string
		varValue string
		varSet   bool
	}
	var savedStates []envState
	for _, name := range []string{"TMP", "TEMP", "APPDATA"} {
		value, set := os.LookupEnv(name)
		savedStates = append(savedStates, envState{varName: name, varValue: value, varSet: set})
	}
	var savedTmpFolder = TemporaryFileFolder()
	var savedAppDataPath = ApplicationDataPath()
	defer func() {
		for _, ss := range savedStates {
			if ss.varSet {
				os.Setenv(ss.varName, ss.varValue)
			} else {
				os.Unsetenv(ss.varName)
			}
		}
		tmpFolder = savedTmpFolder
		appDataPath = savedAppDataPath
	}()
	tests := []struct {
		name            string
		envs            []envState
		wantOk          bool
		wantTmpFolder   string
		wantAppDataPath string
		wantW           string
	}{
		{
			name: "expected use case",
			envs: []envState{
				{varName: "TMP", varValue: "/tmp", varSet: true},
				{varName: "TEMP", varValue: "/tmp2", varSet: true},
				{varName: "APPDATA", varValue: "/users/myUser/AppData/Roaming", varSet: true},
			},
			wantOk:          true,
			wantTmpFolder:   "/tmp",
			wantAppDataPath: "/users/myUser/AppData/Roaming",
			wantW:           "",
		},
		{
			name: "missing TMP",
			envs: []envState{
				{varName: "TMP"},
				{varName: "TEMP", varValue: "/tmp2", varSet: true},
				{varName: "APPDATA", varValue: "/users/myUser/AppData/Roaming", varSet: true},
			},
			wantOk:          true,
			wantTmpFolder:   "/tmp2",
			wantAppDataPath: "/users/myUser/AppData/Roaming",
			wantW:           "",
		},
		{
			name: "missing TMP and TEMP",
			envs: []envState{
				{varName: "TMP"},
				{varName: "TEMP"},
				{varName: "APPDATA", varValue: "/users/myUser/AppData/Roaming", varSet: true},
			},
			wantOk:          false,
			wantTmpFolder:   "",
			wantAppDataPath: "/users/myUser/AppData/Roaming",
			wantW:           "Neither the TMP nor TEMP environment variables are defined.\n",
		},
		{
			name: "missing appDataPath",
			envs: []envState{
				{varName: "TMP", varValue: "/tmp", varSet: true},
				{varName: "TEMP", varValue: "/tmp2", varSet: true},
				{varName: "APPDATA"},
			},
			wantOk:          false,
			wantTmpFolder:   "/tmp",
			wantAppDataPath: "",
			wantW:           "The APPDATA environment variable is not defined.\n",
		},
		{
			name: "missing all vars",
			envs: []envState{
				{varName: "TMP"},
				{varName: "TEMP"},
				{varName: "APPDATA"},
			},
			wantOk:          false,
			wantTmpFolder:   "",
			wantAppDataPath: "",
			wantW:           "Neither the TMP nor TEMP environment variables are defined.\nThe APPDATA environment variable is not defined.\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// clear initial state
			tmpFolder = ""
			appDataPath = ""
			for _, env := range tt.envs {
				if env.varSet {
					os.Setenv(env.varName, env.varValue)
				} else {
					os.Unsetenv(env.varName)
				}
			}
			w := &bytes.Buffer{}
			if gotOk := LookupEnvVars(w); gotOk != tt.wantOk {
				t.Errorf("LookupEnvVars() = %v, want %v", gotOk, tt.wantOk)
			}
			if gotW := w.String(); gotW != tt.wantW {
				t.Errorf("LookupEnvVars() = %v, want %v", gotW, tt.wantW)
			}
			if TemporaryFileFolder() != tt.wantTmpFolder {
				t.Errorf("LookupEnvVars() TmpFolder = %v, want %v", TemporaryFileFolder(), tt.wantTmpFolder)
			}
			if ApplicationDataPath() != tt.wantAppDataPath {
				t.Errorf("LookupEnvVars() AppDataPath = %v, want %v", ApplicationDataPath(), tt.wantAppDataPath)
			}
		})
	}
}
