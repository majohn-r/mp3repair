package internal

import (
	"path/filepath"
	"reflect"
	"testing"
)

func Test_Configuration_SubConfiguration(t *testing.T) {
	if err := CreateDefaultYamlFileForTesting(); err != nil {
		t.Errorf("error creating defaults.yaml: %v", err)
	}
	fnName := "Configuration_SubConfiguration()"
	defer func() {
		DestroyDirectoryForTesting(fnName, "./mp3")
	}()
	testConfiguration := ReadConfigurationFile("./mp3")
	type args struct {
		key string
	}
	tests := []struct {
		name string
		c    *Configuration
		args args
		want *Configuration
	}{
		{name: "no configuration", c: &Configuration{}, args: args{}, want: EmptyConfiguration()},
		{name: "commons", c: testConfiguration, args: args{key: "common"}, want: testConfiguration.cMap["common"]},
		{name: "ls", c: testConfiguration, args: args{key: "ls"}, want: testConfiguration.cMap["ls"]},
		{name: "check", c: testConfiguration, args: args{key: "check"}, want: testConfiguration.cMap["check"]},
		{name: "repair", c: testConfiguration, args: args{key: "repair"}, want: testConfiguration.cMap["repair"]},
		{name: "unknown key", c: testConfiguration, args: args{key: "unknown key"}, want: EmptyConfiguration()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.SubConfiguration(tt.args.key); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func Test_Configuration_BoolDefault(t *testing.T) {
	if err := CreateDefaultYamlFileForTesting(); err != nil {
		t.Errorf("error creating defaults.yaml: %v", err)
	}
	fnName := "Configuration_BoolDefault()"
	defer func() {
		DestroyDirectoryForTesting(fnName, "./mp3")
	}()
	testConfiguration := ReadConfigurationFile("./mp3")
	type args struct {
		key          string
		defaultValue bool
	}
	tests := []struct {
		name  string
		c     *Configuration
		args  args
		wantB bool
	}{
		{
			name:  "empty configuration default false",
			c:     EmptyConfiguration(),
			args:  args{defaultValue: false},
			wantB: false,
		},
		{
			name:  "empty configuration default true",
			c:     EmptyConfiguration(),
			args:  args{defaultValue: true},
			wantB: true,
		},
		{
			name:  "undefined key default false",
			c:     testConfiguration,
			args:  args{key: "no key", defaultValue: false},
			wantB: false,
		},
		{
			name:  "undefined key default true",
			c:     testConfiguration,
			args:  args{key: "no key", defaultValue: true},
			wantB: true,
		},
		{
			name:  "non-boolean value default false",
			c:     testConfiguration.cMap["common"],
			args:  args{key: "albums", defaultValue: false},
			wantB: false,
		},
		{
			name:  "non-boolean value default true",
			c:     testConfiguration.cMap["common"],
			args:  args{key: "albums", defaultValue: true},
			wantB: true,
		},
		{
			name:  "boolean value default false",
			c:     testConfiguration.cMap["ls"],
			args:  args{key: "includeTracks", defaultValue: false},
			wantB: true,
		},
		{
			name:  "boolean value default true",
			c:     testConfiguration.cMap["ls"],
			args:  args{key: "includeTracks", defaultValue: true},
			wantB: true,
		},
		{
			name:  "boolean (string) value default true",
			c:     testConfiguration.cMap["unused"],
			args:  args{key: "value", defaultValue: true},
			wantB: true,
		},
		{
			name:  "boolean (string) value default false",
			c:     testConfiguration.cMap["unused"],
			args:  args{key: "value", defaultValue: false},
			wantB: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotB := tt.c.BoolDefault(tt.args.key, tt.args.defaultValue); gotB != tt.wantB {
				t.Errorf("%s = %v, want %v", fnName, gotB, tt.wantB)
			}
		})
	}
}

func Test_Configuration_StringDefault(t *testing.T) {
	if err := CreateDefaultYamlFileForTesting(); err != nil {
		t.Errorf("error creating defaults.yaml: %v", err)
	}
	fnName := "Configuration_StringDefault()"
	defer func() {
		DestroyDirectoryForTesting(fnName, "./mp3")
	}()
	testConfiguration := ReadConfigurationFile("./mp3")
	type args struct {
		key          string
		defaultValue string
	}
	tests := []struct {
		name  string
		c     *Configuration
		args  args
		wantS string
	}{
		{
			name:  "empty configuration",
			c:     EmptyConfiguration(),
			args:  args{defaultValue: "my default value"},
			wantS: "my default value"},
		{
			name:  "undefined key",
			c:     testConfiguration,
			args:  args{key: "no key", defaultValue: "my default value"},
			wantS: "my default value",
		},
		{
			name:  "defined key",
			c:     testConfiguration.cMap["ls"],
			args:  args{key: "sort", defaultValue: "my default value"},
			wantS: "alpha",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotS := tt.c.StringDefault(tt.args.key, tt.args.defaultValue); gotS != tt.wantS {
				t.Errorf("%s = %v, want %v", fnName, gotS, tt.wantS)
			}
		})
	}
}

func TestReadConfigurationFile(t *testing.T) {
	fnName := "ReadConfigurationFile()"
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
	if err := CreateFileForTestingWithContent(badDir, defaultConfigFileName, "gibberish"); err != nil {
		t.Errorf("%s error creating non-standard defaults.yaml: %v", fnName, err)
	}
	type args struct {
		path string
	}
	tests := []struct {
		name string
		args args
		want *Configuration
	}{
		{name: "bad", args: args{path: badDir}, want: EmptyConfiguration()},
		{name: "good",
			args: args{path: "./mp3"},
			want: &Configuration{
				bMap: map[string]bool{},
				sMap: map[string]string{},
				cMap: map[string]*Configuration{
					"check": {
						bMap: map[string]bool{"empty": true, "gaps": true, "integrity": false},
						sMap: map[string]string{},
						cMap: map[string]*Configuration{},
					},
					"common": {
						bMap: map[string]bool{},
						sMap: map[string]string{
							"albumFilter":  "^.*$",
							"artistFilter": "^.*$",
							"ext":          ".mpeg",
							"topDir":       ".",
						},
						cMap: map[string]*Configuration{},
					},
					"ls": {
						bMap: map[string]bool{
							"annotate":       true,
							"includeAlbums":  false,
							"includeArtists": false,
							"includeTracks":  true,
						},
						sMap: map[string]string{"sort": "alpha"},
						cMap: map[string]*Configuration{},
					},
					"repair": {
						bMap: map[string]bool{"dryRun": true},
						sMap: map[string]string{},
						cMap: map[string]*Configuration{},
					},
					"unused": {
						bMap: map[string]bool{},
						sMap: map[string]string{"value": "1"},
						cMap: map[string]*Configuration{}},
				},
			},
		},
		{name: "error", args: args{path: "./non-existent-dir"}, want: EmptyConfiguration()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ReadConfigurationFile(tt.args.path)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReadYaml() = %v, want %v", got, tt.want)
			}
		})
	}
}
