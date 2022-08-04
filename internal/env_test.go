package internal

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func Test_findReferences(t *testing.T) {
	fnName := "findReferences()"
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
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func TestInterpretEnvVarReferences(t *testing.T) {
	fnName := "InterpretEnvVarReferences()"
	originalExtension := os.Getenv("EXTENSION")
	originalFileName := os.Getenv("FILENAME")
	originalPath := os.Getenv("PATH")
	originalSubPath := os.Getenv("SUBPATH")
	originalVarX := os.Getenv("VARX")
	originalVarY := os.Getenv("VARY")
	defer func() {
		os.Setenv("EXTENSION", originalExtension)
		os.Setenv("FILENAME", originalFileName)
		os.Setenv("PATH", originalPath)
		os.Setenv("SUBPATH", originalSubPath)
		os.Setenv("VARX", originalVarX)
		os.Setenv("VARY", originalVarY)
	}()
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
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args
		want    string
		wantErr bool
	}{
		{name: "no references", args: args{s: "no references"}, want: "no references"},
		{
			name: "lots of references",
			args: args{s: "$PATH/$SUBPATH/%FILENAME%.%EXTENSION%"},
			want: "/c/Users/MyUser/Music/track.mp3",
		},
		{
			name:    "missing references",
			args:    args{s: "$VARX + %VARY%"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
	fnName := "CreateAppSpecificPath()"
	type args struct {
		topDir string
	}
	tests := []struct {
		name string
		args args
		want string
	}{{name: "simple test", args: args{topDir: "top"}, want: filepath.Join("top", AppName)}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CreateAppSpecificPath(tt.args.topDir); got != tt.want {
				t.Errorf("%s = %q, want %q", fnName, got, tt.want)
			}
		})
	}
}
