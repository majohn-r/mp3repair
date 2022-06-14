package internal

import (
	"path/filepath"
	"reflect"
	"testing"
)

func Test_SafeSubNode(t *testing.T) {
	if err := CreateDefaultYamlFileForTesting(); err != nil {
		t.Errorf("error creating defaults.yaml: %v", err)
	}
	fnName := "SafeSubNode()"
	defer func() {
		DestroyDirectoryForTesting(fnName, "./mp3")
	}()
	testN := ReadYaml("./mp3")
	type args struct {
		n   *Node
		key string
	}
	tests := []struct {
		name string
		args args
		want *Node
	}{
		{name: "no node", args: args{}},
		{name: "commons", args: args{n: testN, key: "common"}, want: testN.nMap["common"]},
		{name: "ls", args: args{n: testN, key: "ls"}, want: testN.nMap["ls"]},
		{name: "check", args: args{n: testN, key: "check"}, want: testN.nMap["check"]},
		{name: "repair", args: args{n: testN, key: "repair"}, want: testN.nMap["repair"]},
		{name: "unknown key", args: args{n: testN, key: "unknown key"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SafeSubNode(tt.args.n, tt.args.key); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func Test_GetBoolDefault(t *testing.T) {
	if err := CreateDefaultYamlFileForTesting(); err != nil {
		t.Errorf("error creating defaults.yaml: %v", err)
	}
	fnName := "GetBoolDefault()"
	defer func() {
		DestroyDirectoryForTesting(fnName, "./mp3")
	}()
	testN := ReadYaml("./mp3")
	type args struct {
		n            *Node
		key          string
		defaultValue bool
	}
	tests := []struct {
		name  string
		args  args
		wantB bool
	}{
		{name: "nil node default false", args: args{defaultValue: false}, wantB: false},
		{name: "nil node default true", args: args{defaultValue: true}, wantB: true},
		{name: "undefined key default false", args: args{n: testN, key: "no key", defaultValue: false}, wantB: false},
		{name: "undefined key default true", args: args{n: testN, key: "no key", defaultValue: true}, wantB: true},
		{
			name:  "non-boolean value default false",
			args:  args{n: testN.nMap["common"], key: "albums", defaultValue: false},
			wantB: false,
		},
		{
			name:  "non-boolean value default true",
			args:  args{n: testN.nMap["common"], key: "albums", defaultValue: true},
			wantB: true,
		},
		{
			name:  "boolean value default false",
			args:  args{n: testN.nMap["ls"], key: "includeTracks", defaultValue: false},
			wantB: true,
		},
		{
			name:  "boolean value default true",
			args:  args{n: testN.nMap["ls"], key: "includeTracks", defaultValue: true},
			wantB: true,
		},
		{
			name:  "boolean (string) value default true",
			args:  args{n: testN.nMap["unused"], key: "value", defaultValue: true},
			wantB: true,
		},
		{
			name:  "boolean (string) value default false",
			args:  args{n: testN.nMap["unused"], key: "value", defaultValue: false},
			wantB: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotB := GetBoolDefault(tt.args.n, tt.args.key, tt.args.defaultValue); gotB != tt.wantB {
				t.Errorf("%s = %v, want %v", fnName, gotB, tt.wantB)
			}
		})
	}
}

func Test_GetStringDefault(t *testing.T) {
	if err := CreateDefaultYamlFileForTesting(); err != nil {
		t.Errorf("error creating defaults.yaml: %v", err)
	}
	fnName := "GetStringDefault()"
	defer func() {
		DestroyDirectoryForTesting(fnName, "./mp3")
	}()
	testN := ReadYaml("./mp3")
	type args struct {
		n            *Node
		key          string
		defaultValue string
	}
	tests := []struct {
		name  string
		args  args
		wantS string
	}{
		{name: "nil node", args: args{defaultValue: "my default value"}, wantS: "my default value"},
		{
			name:  "undefined key",
			args:  args{n: testN, key: "no key", defaultValue: "my default value"},
			wantS: "my default value",
		},
		{
			name:  "defined key",
			args:  args{n: testN.nMap["ls"], key: "sort", defaultValue: "my default value"},
			wantS: "alpha",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotS := GetStringDefault(tt.args.n, tt.args.key, tt.args.defaultValue); gotS != tt.wantS {
				t.Errorf("%s = %v, want %v", fnName, gotS, tt.wantS)
			}
		})
	}
}

func TestReadYaml(t *testing.T) {
	fnName := "ReadYaml()"
	if err := CreateDefaultYamlFileForTesting(); err != nil {
		t.Errorf("%s error creating defaults.yaml: %v", fnName, err)
	}
	defer func() {
		DestroyDirectoryForTesting(fnName, "./mp3")
	}()
	badDir := filepath.Join(".", "mp3", "badData")
	if err := Mkdir(badDir); err != nil {
		t.Errorf("%s error creating non-standard test directory %q: %v", fnName, badDir, err)
	}
	if err := CreateFileForTestingWithContent(badDir, DefaultConfigFileName, "gibberish"); err != nil {
		t.Errorf("%s error creating non-standard defaults.yaml: %v", fnName, err)
	}
	type args struct {
		path string
	}
	tests := []struct {
		name string
		args args
		want *Node
	}{
		{name: "bad", args: args{path: badDir}},
		{name: "good",
			args: args{path: "./mp3"},
			want: &Node{
				bMap: map[string]bool{},
				sMap: map[string]string{},
				nMap: map[string]*Node{
					"check": {
						bMap: map[string]bool{"empty": true, "gaps": true, "integrity": false},
						sMap: map[string]string{},
						nMap: map[string]*Node{},
					},
					"common": {
						bMap: map[string]bool{},
						sMap: map[string]string{
							"albumFilter":  "^.*$",
							"artistFilter": "^.*$",
							"ext":          ".mpeg",
							"topDir":       ".",
						},
						nMap: map[string]*Node{},
					},
					"ls": {
						bMap: map[string]bool{
							"annotate":       true,
							"includeAlbums":  false,
							"includeArtists": false,
							"includeTracks":  true,
						},
						sMap: map[string]string{"sort": "alpha"},
						nMap: map[string]*Node{},
					},
					"repair": {
						bMap: map[string]bool{"dryRun": true},
						sMap: map[string]string{},
						nMap: map[string]*Node{},
					},
					"unused": {
						bMap: map[string]bool{},
						sMap: map[string]string{"value": "1"},
						nMap: map[string]*Node{}},
				},
			},
		},
		{name: "error", args: args{path: "./non-existent-dir"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ReadYaml(tt.args.path)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReadYaml() = %v, want %v", got, tt.want)
			}
		})
	}
}
