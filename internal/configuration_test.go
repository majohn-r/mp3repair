package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func Test_Configuration_SubConfiguration(t *testing.T) {
	fnName := "Configuration.SubConfiguration()"
	savedState := SaveEnvVarForTesting(appDataVar)
	os.Setenv(appDataVar, SecureAbsolutePathForTesting("."))
	defer func() {
		savedState.RestoreForTesting()
	}()
	if err := CreateDefaultYamlFileForTesting(); err != nil {
		t.Errorf("%s error creating defaults.yaml: %v", fnName, err)
	}
	defer func() {
		DestroyDirectoryForTesting(fnName, "./mp3")
	}()
	testConfiguration, _ := ReadConfigurationFile(NewOutputDeviceForTesting())
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
	fnName := "Configuration.BoolDefault()"
	savedState := SaveEnvVarForTesting(appDataVar)
	os.Setenv(appDataVar, SecureAbsolutePathForTesting("."))
	defer func() {
		savedState.RestoreForTesting()
	}()
	if err := CreateDefaultYamlFileForTesting(); err != nil {
		t.Errorf("%s error creating defaults.yaml: %v", fnName, err)
	}
	defer func() {
		DestroyDirectoryForTesting(fnName, "./mp3")
	}()
	testConfiguration, _ := ReadConfigurationFile(NewOutputDeviceForTesting())
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
	fnName := "Configuration.StringDefault()"
	savedState := SaveEnvVarForTesting(appDataVar)
	os.Setenv(appDataVar, SecureAbsolutePathForTesting("."))
	defer func() {
		savedState.RestoreForTesting()
	}()
	if err := CreateDefaultYamlFileForTesting(); err != nil {
		t.Errorf("%s error creating defaults.yaml: %v", fnName, err)
	}
	defer func() {
		DestroyDirectoryForTesting(fnName, "./mp3")
	}()
	testConfiguration, _ := ReadConfigurationFile(NewOutputDeviceForTesting())
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

func Test_verifyFileExists(t *testing.T) {
	fnName := "verifyFileExists()"
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		wantOk  bool
		wantErr bool
		WantedOutput
	}{
		{
			name:         "ordinary success",
			args:         args{path: "./configuration_test.go"},
			wantOk:       true,
			WantedOutput: WantedOutput{},
		},
		{
			name:    "look for dir!",
			args:    args{path: "."},
			wantErr: true,
			WantedOutput: WantedOutput{
				WantErrorOutput: "The configuration file \".\" is a directory.\n",
				WantLogOutput:   "level='error' directory='.' fileName='.' msg='file is a directory'\n",
			},
		},
		{
			name: "non-existent file",
			args: args{path: "./no-such-file.txt"},
			WantedOutput: WantedOutput{
				WantLogOutput: "level='info' directory='.' fileName='no-such-file.txt' msg='file does not exist'\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := NewOutputDeviceForTesting()
			gotOk, err := verifyFileExists(o, tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("%s error = %v, wantErr %v", fnName, err, tt.wantErr)
				return
			}
			if gotOk != tt.wantOk {
				t.Errorf("%s = %v, want %v", fnName, gotOk, tt.wantOk)
			}
			if issues, ok := o.CheckOutput(tt.WantedOutput); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
		})
	}
}

func TestReadConfigurationFile(t *testing.T) {
	fnName := "ReadConfigurationFile()"
	savedState := SaveEnvVarForTesting(appDataVar)
	canonicalPath := SecureAbsolutePathForTesting(".")
	mp3Path := SecureAbsolutePathForTesting("mp3")
	os.Setenv(appDataVar, canonicalPath)
	defer func() {
		savedState.RestoreForTesting()
	}()
	if err := CreateDefaultYamlFileForTesting(); err != nil {
		t.Errorf("%s error creating defaults.yaml: %v", fnName, err)
	}
	defer func() {
		DestroyDirectoryForTesting(fnName, "./mp3")
	}()
	badDir := filepath.Join("./mp3", "fake")
	if err := Mkdir(badDir); err != nil {
		t.Errorf("%s error creating fake dir: %v", fnName, err)
	}
	badDir2 := filepath.Join(badDir, AppName)
	if err := Mkdir(badDir2); err != nil {
		t.Errorf("%s error creating fake dir mp3 folder: %v", fnName, err)
	}
	badFile := filepath.Join(badDir2, defaultConfigFileName)
	if err := Mkdir(badFile); err != nil {
		t.Errorf("%s error creating defaults.yaml as a directory: %v", fnName, err)
	}
	yamlAsDir := SecureAbsolutePathForTesting(badFile)
	gibberishDir := filepath.Join(badDir2, AppName)
	if err := Mkdir(gibberishDir); err != nil {
		t.Errorf("%s error creating gibberish folder: %v", fnName, err)
	}
	if err := CreateFileForTestingWithContent(gibberishDir, defaultConfigFileName, "gibberish"); err != nil {
		t.Errorf("%s error creating gibberish defaults.yaml: %v", fnName, err)
	}
	tests := []struct {
		name   string
		state  *SavedEnvVar
		wantC  *Configuration
		wantOk bool
		WantedOutput
	}{
		{
			name:  "good",
			state: &SavedEnvVar{Name: appDataVar, Value: canonicalPath, Set: true},
			wantC: &Configuration{
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
			wantOk: true,
			WantedOutput: WantedOutput{
				WantLogOutput: fmt.Sprintf("level='warn' key='value' type='int' value='1' msg='unexpected value type'\n"+
					"level='info' directory='%s' fileName='defaults.yaml' value='map[check:map[empty:true gaps:true integrity:false] common:map[albumFilter:^.*$ artistFilter:^.*$ ext:.mpeg topDir:.] ls:map[annotate:true includeAlbums:false includeArtists:false includeTracks:true], map[sort:alpha] repair:map[dryRun:true] unused:map[value:1]]' msg='read configuration file'\n", mp3Path),
			},
		},
		{
			name:   "APPDATA not set",
			state:  &SavedEnvVar{Name: appDataVar},
			wantC:  EmptyConfiguration(),
			wantOk: true,
			WantedOutput: WantedOutput{
				WantLogOutput: "level='info' environment variable='APPDATA' msg='not set'\n",
			},
		},
		{
			name:  "defaults.yaml is a directory",
			state: &SavedEnvVar{Name: appDataVar, Value: SecureAbsolutePathForTesting(badDir), Set: true},
			WantedOutput: WantedOutput{
				WantErrorOutput: fmt.Sprintf("The configuration file %q is a directory.\n", yamlAsDir),
				WantLogOutput:   fmt.Sprintf("level='error' directory='%s' fileName='defaults.yaml' msg='file is a directory'\n", SecureAbsolutePathForTesting(badDir2)),
			},
		},
		{
			name:   "missing yaml",
			state:  &SavedEnvVar{Name: appDataVar, Value: SecureAbsolutePathForTesting(yamlAsDir), Set: true},
			wantC:  EmptyConfiguration(),
			wantOk: true,
			WantedOutput: WantedOutput{
				WantLogOutput: fmt.Sprintf("level='info' directory='%s' fileName='defaults.yaml' msg='file does not exist'\n", SecureAbsolutePathForTesting(filepath.Join(yamlAsDir, AppName))),
			},
		},
		{
			name:  "malformed yaml",
			state: &SavedEnvVar{Name: appDataVar, Value: SecureAbsolutePathForTesting(badDir2), Set: true},
			WantedOutput: WantedOutput{
				WantErrorOutput: fmt.Sprintf(
					"The configuration file %q is not well-formed YAML: yaml: unmarshal errors:\n  line 1: cannot unmarshal !!str `gibberish` into map[string]interface {}\n",
					SecureAbsolutePathForTesting(filepath.Join(gibberishDir, defaultConfigFileName))),
				WantLogOutput: fmt.Sprintf("level='warn' directory='%s' error='yaml: unmarshal errors:\n  line 1: cannot unmarshal !!str `gibberish` into map[string]interface {}' fileName='defaults.yaml' msg='cannot unmarshal yaml content'\n", SecureAbsolutePathForTesting(gibberishDir)),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.state.RestoreForTesting()
			o := NewOutputDeviceForTesting()
			gotC, gotOk := ReadConfigurationFile(o)
			if !reflect.DeepEqual(gotC, tt.wantC) {
				t.Errorf("%s gotC = %v, want %v", fnName, gotC, tt.wantC)
			}
			if gotOk != tt.wantOk {
				t.Errorf("%s gotOk = %v, want %v", fnName, gotOk, tt.wantOk)
			}
			if issues, ok := o.CheckOutput(tt.WantedOutput); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
		})
	}
}

func Test_appData(t *testing.T) {
	fnName := "appData()"
	savedState := SaveEnvVarForTesting(appDataVar)
	os.Setenv(appDataVar, SecureAbsolutePathForTesting("."))
	defer func() {
		savedState.RestoreForTesting()
	}()
	tests := []struct {
		name  string
		state *SavedEnvVar
		want  string
		want1 bool
		WantedOutput
	}{
		{
			name:         "value is set",
			state:        &SavedEnvVar{Name: appDataVar, Value: "appData!", Set: true},
			want:         "appData!",
			want1:        true,
			WantedOutput: WantedOutput{},
		},
		{
			name:  "value is not set",
			state: &SavedEnvVar{Name: appDataVar},
			WantedOutput: WantedOutput{
				WantLogOutput: "level='info' environment variable='APPDATA' msg='not set'\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.state.RestoreForTesting()
			o := NewOutputDeviceForTesting()
			got, got1 := appData(o)
			if got != tt.want {
				t.Errorf("%s got = %q, want %q", fnName, got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("%s got1 = %v, want %v", fnName, got1, tt.want1)
			}
			if issues, ok := o.CheckOutput(tt.WantedOutput); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
		})
	}
}

func TestConfiguration_StringValue(t *testing.T) {
	fnName := "Configuration.StringValue()"
	type args struct {
		key string
	}
	tests := []struct {
		name      string
		c         *Configuration
		args      args
		wantValue string
		wantOk    bool
	}{
		{
			name:      "found",
			c:         &Configuration{sMap: map[string]string{"key": "value"}},
			args:      args{key: "key"},
			wantValue: "value",
			wantOk:    true,
		},
		{
			name: "not found",
			c:    EmptyConfiguration(),
			args: args{key: "key"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotValue, gotOk := tt.c.StringValue(tt.args.key)
			if gotValue != tt.wantValue {
				t.Errorf("%s gotValue = %q, want %q", fnName, gotValue, tt.wantValue)
			}
			if gotOk != tt.wantOk {
				t.Errorf("%s gotOk = %v, want %v", fnName, gotOk, tt.wantOk)
			}
		})
	}
}
