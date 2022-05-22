package internal

import (
	"reflect"
	"testing"

	"github.com/spf13/viper"
)

func Test_ReadDefaultsYaml(t *testing.T) {
	fnName := "ReadDefaultsYaml()"
	if err := CreateDefaultYamlFile(); err != nil {
		t.Errorf("%s error creating defaults.yaml: %v", fnName, err)
	}
	defer func() {
		DestroyDirectoryForTesting(fnName, "./mp3")
	}()
	type args struct {
		path string
	}
	tests := []struct {
		name  string
		args  args
		wantV *viper.Viper
	}{
		{name: "good", args: args{path: "./mp3"}, wantV: viper.New()},
		{name: "error", args: args{path: "./non-existent-dir"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotV := ReadDefaultsYaml(tt.args.path); !okViper(gotV, tt.wantV) {
				t.Errorf("%s = %v, want %v", fnName, gotV, tt.wantV)
			}
		})
	}
}

func okViper(got, want *viper.Viper) bool {
	if got == nil {
		return want == nil
	}
	return want != nil
}

func Test_SafeSubViper(t *testing.T) {
	if err := CreateDefaultYamlFile(); err != nil {
		t.Errorf("error creating defaults.yaml: %v", err)
	}
	fnName := "SafeSubViper()"
	defer func() {
		DestroyDirectoryForTesting(fnName, "./mp3")
	}()
	testV := ReadDefaultsYaml("./mp3")
	type args struct {
		v   *viper.Viper
		key string
	}
	tests := []struct {
		name string
		args args
		want *viper.Viper
	}{
		{name: "no viper", args: args{}},
		{name: "commons", args: args{v: testV, key: "common"}, want: testV.Sub("common")},
		{name: "ls", args: args{v: testV, key: "ls"}, want: testV.Sub("ls")},
		{name: "check", args: args{v: testV, key: "check"}, want: testV.Sub("check")},
		{name: "repair", args: args{v: testV, key: "repair"}, want: testV.Sub("repair")},
		{name: "unknown key", args: args{v: testV, key: "unknown key"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SafeSubViper(tt.args.v, tt.args.key); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func Test_GetBoolDefault(t *testing.T) {
	if err := CreateDefaultYamlFile(); err != nil {
		t.Errorf("error creating defaults.yaml: %v", err)
	}
	fnName := "GetBoolDefault()"
	defer func() {
		DestroyDirectoryForTesting(fnName, "./mp3")
	}()
	testV := ReadDefaultsYaml("./mp3")
	type args struct {
		v            *viper.Viper
		key          string
		defaultValue bool
	}
	tests := []struct {
		name  string
		args  args
		wantB bool
	}{
		{name: "nil viper default false", args: args{defaultValue: false}, wantB: false},
		{name: "nil viper default true", args: args{defaultValue: true}, wantB: true},
		{name: "undefined key default false", args: args{v: testV, key: "no key", defaultValue: false}, wantB: false},
		{name: "undefined key default true", args: args{v: testV, key: "no key", defaultValue: true}, wantB: true},
		{
			name:  "non-boolean value default false",
			args:  args{v: testV.Sub("common"), key: "albums", defaultValue: false},
			wantB: false,
		},
		{
			name:  "non-boolean value default true",
			args:  args{v: testV.Sub("common"), key: "albums", defaultValue: true},
			wantB: true,
		},
		{
			name:  "boolean value default false",
			args:  args{v: testV.Sub("ls"), key: "track", defaultValue: false},
			wantB: true,
		},
		{
			name:  "boolean value default true",
			args:  args{v: testV.Sub("ls"), key: "track", defaultValue: true},
			wantB: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotB := GetBoolDefault(tt.args.v, tt.args.key, tt.args.defaultValue); gotB != tt.wantB {
				t.Errorf("%s = %v, want %v", fnName, gotB, tt.wantB)
			}
		})
	}
}

func Test_GetStringDefault(t *testing.T) {
	if err := CreateDefaultYamlFile(); err != nil {
		t.Errorf("error creating defaults.yaml: %v", err)
	}
	fnName := "GetStringDefault()"
	defer func() {
		DestroyDirectoryForTesting(fnName, "./mp3")
	}()
	testV := ReadDefaultsYaml("./mp3")
	type args struct {
		v            *viper.Viper
		key          string
		defaultValue string
	}
	tests := []struct {
		name  string
		args  args
		wantS string
	}{
		{name: "nil viper", args: args{defaultValue: "my default value"}, wantS: "my default value"},
		{
			name:  "undefined key",
			args:  args{v: testV, key: "no key", defaultValue: "my default value"},
			wantS: "my default value",
		},
		{
			name:  "defined key",
			args:  args{v: testV.Sub("ls"), key: "sort", defaultValue: "my default value"},
			wantS: "alpha",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotS := GetStringDefault(tt.args.v, tt.args.key, tt.args.defaultValue); gotS != tt.wantS {
				t.Errorf("%s = %v, want %v", fnName, gotS, tt.wantS)
			}
		})
	}
}
