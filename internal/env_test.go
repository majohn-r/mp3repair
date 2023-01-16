package internal

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func Test_findReferences(t *testing.T) {
	const fnName = "findReferences()"
	type args struct {
		s string
	}
	tests := map[string]struct {
		args
		want []string
	}{
		"no references":      {args: args{s: ".mp3"}},
		"lots of references": {args: args{s: "$PATH/$SUBPATH/%FILENAME%.%EXTENSION%"}, want: []string{"$SUBPATH", "$PATH", "%FILENAME%", "%EXTENSION%"}},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := findReferences(tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func TestInterpretEnvVarReferences(t *testing.T) {
	const fnName = "InterpretEnvVarReferences()"
	originalExtension := os.Getenv("EXTENSION")
	originalFileName := os.Getenv("FILENAME")
	originalPath := os.Getenv("PATH")
	originalSubPath := os.Getenv("SUBPATH")
	originalVarX := os.Getenv("VARX")
	originalVarY := os.Getenv("VARY")
	newExtension := "mp3"
	newFileName := "track"
	newPath := "/c/Users/MyUser"
	newSubPath := "Music"
	os.Setenv("EXTENSION", newExtension)
	os.Setenv("FILENAME", newFileName)
	os.Setenv("PATH", newPath)
	os.Setenv("SUBPATH", newSubPath)
	os.Unsetenv("VARX")
	os.Unsetenv("VARY")
	defer func() {
		os.Setenv("EXTENSION", originalExtension)
		os.Setenv("FILENAME", originalFileName)
		os.Setenv("PATH", originalPath)
		os.Setenv("SUBPATH", originalSubPath)
		os.Setenv("VARX", originalVarX)
		os.Setenv("VARY", originalVarY)
	}()
	type args struct {
		s string
	}
	tests := map[string]struct {
		args
		want    string
		wantErr bool
	}{
		"no references":      {args: args{s: "no references"}, want: "no references"},
		"lots of references": {args: args{s: "$PATH/$SUBPATH/%FILENAME%.%EXTENSION%"}, want: "/c/Users/MyUser/Music/track.mp3"},
		"missing references": {args: args{s: "$VARX + %VARY%"}, wantErr: true},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, gotErr := InterpretEnvVarReferences(tt.args.s)
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("%s gotErr %v wantErr %t", fnName, gotErr, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("%s = %q, want %q", fnName, got, tt.want)
			}
		})
	}
}

func TestCreateAppSpecificPath(t *testing.T) {
	const fnName = "CreateAppSpecificPath()"
	type args struct {
		topDir string
	}
	tests := map[string]struct {
		args
		want string
	}{"simple test": {args: args{topDir: "top"}, want: filepath.Join("top", AppName)}}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := CreateAppSpecificPath(tt.args.topDir); got != tt.want {
				t.Errorf("%s = %q, want %q", fnName, got, tt.want)
			}
		})
	}
}
