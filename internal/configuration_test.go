package internal

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func Test_Configuration_SubConfiguration(t *testing.T) {
	savedState := SaveEnvVarForTesting(appDataVar)
	os.Setenv(appDataVar, SecureAbsolutePathForTesting("."))
	defer func() {
		savedState.RestoreForTesting()
	}()
	if err := CreateDefaultYamlFileForTesting(); err != nil {
		t.Errorf("error creating defaults.yaml: %v", err)
	}
	fnName := "Configuration.SubConfiguration()"
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
	savedState := SaveEnvVarForTesting(appDataVar)
	os.Setenv(appDataVar, SecureAbsolutePathForTesting("."))
	defer func() {
		savedState.RestoreForTesting()
	}()
	if err := CreateDefaultYamlFileForTesting(); err != nil {
		t.Errorf("error creating defaults.yaml: %v", err)
	}
	fnName := "Configuration.BoolDefault()"
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
	savedState := SaveEnvVarForTesting(appDataVar)
	os.Setenv(appDataVar, SecureAbsolutePathForTesting("."))
	defer func() {
		savedState.RestoreForTesting()
	}()
	if err := CreateDefaultYamlFileForTesting(); err != nil {
		t.Errorf("error creating defaults.yaml: %v", err)
	}
	fnName := "Configuration.StringDefault()"
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
	type args struct {
		path string
	}
	tests := []struct {
		name     string
		args     args
		wantOk   bool
		wantWErr string
		wantErr  bool
	}{
		{
			name:     "ordinary success",
			args:     args{path: "./configuration_test.go"},
			wantOk:   true,
			wantWErr: "",
			wantErr:  false,
		},
		{
			name:     "look for dir!",
			args:     args{path: "."},
			wantOk:   false,
			wantWErr: "The configuration file \".\" is a directory.\n",
			wantErr:  true,
		},
		{
			name:     "non-existent file",
			args:     args{path: "./no-such-file.txt"},
			wantOk:   false,
			wantWErr: "",
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wErr := &bytes.Buffer{}
			gotOk, err := verifyFileExists(wErr, tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("verifyFileExists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotOk != tt.wantOk {
				t.Errorf("verifyFileExists() = %v, want %v", gotOk, tt.wantOk)
			}
			if gotWErr := wErr.String(); gotWErr != tt.wantWErr {
				t.Errorf("verifyFileExists() = %v, want %v", gotWErr, tt.wantWErr)
			}
		})
	}
}

func TestReadConfigurationFile(t *testing.T) {
	savedState := SaveEnvVarForTesting(appDataVar)
	canonicalPath := SecureAbsolutePathForTesting(".")
	mp3Path := SecureAbsolutePathForTesting("mp3")
	os.Setenv(appDataVar, canonicalPath)
	defer func() {
		savedState.RestoreForTesting()
	}()
	fnName := "ReadConfigurationFile()"
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
		name    string
		state   *SavedEnvVar
		wantC   *Configuration
		wantOk  bool
		wantOut string
		wantErr string
		wantLog string
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
			wantLog: fmt.Sprintf("level='warn' key='value' type='int' value='1' msg='unexpected value type'\n"+
				"level='info' directory='%s' fileName='defaults.yaml' value='map[check:map[empty:true gaps:true integrity:false] common:map[albumFilter:^.*$ artistFilter:^.*$ ext:.mpeg topDir:.] ls:map[annotate:true includeAlbums:false includeArtists:false includeTracks:true], map[sort:alpha] repair:map[dryRun:true] unused:map[value:1]]' msg='read configuration file'\n", mp3Path),
		},
		{
			name:    "APPDATA not set",
			state:   &SavedEnvVar{Name: appDataVar},
			wantC:   EmptyConfiguration(),
			wantOk:  true,
			wantLog: "level='info' environment variable='APPDATA' msg='not set'\n",
		},
		{
			name:    "defaults.yaml is a directory",
			state:   &SavedEnvVar{Name: appDataVar, Value: SecureAbsolutePathForTesting(badDir), Set: true},
			wantErr: fmt.Sprintf("The configuration file %q is a directory.\n", yamlAsDir),
		},
		{
			name:   "missing yaml",
			state:  &SavedEnvVar{Name: appDataVar, Value: SecureAbsolutePathForTesting(yamlAsDir), Set: true},
			wantC:  EmptyConfiguration(),
			wantOk: true,
		},
		{
			name:  "malformed yaml",
			state: &SavedEnvVar{Name: appDataVar, Value: SecureAbsolutePathForTesting(badDir2), Set: true},
			wantErr: fmt.Sprintf(
				"The configuration file %q is not well-formed YAML: yaml: unmarshal errors:\n  line 1: cannot unmarshal !!str `gibberish` into map[string]interface {}\n",
				SecureAbsolutePathForTesting(filepath.Join(gibberishDir, defaultConfigFileName))),
			wantLog: fmt.Sprintf("level='warn' directory='%s' error='yaml: unmarshal errors:\n  line 1: cannot unmarshal !!str `gibberish` into map[string]interface {}' fileName='defaults.yaml' msg='cannot unmarshal yaml content'\n", SecureAbsolutePathForTesting(gibberishDir)),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.state.RestoreForTesting()
			o := NewOutputDeviceForTesting()
			gotC, gotOk := ReadConfigurationFile(o)
			if !reflect.DeepEqual(gotC, tt.wantC) {
				t.Errorf("ReadConfigurationFile() gotC = %v, want %v", gotC, tt.wantC)
			}
			if gotOk != tt.wantOk {
				t.Errorf("ReadConfigurationFile() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
			if gotOut := o.Stdout(); gotOut != tt.wantOut {
				t.Errorf("ReadConfigurationFile() console output = %v, want %v", gotOut, tt.wantOut)
			}
			if gotErr := o.Stderr(); gotErr != tt.wantErr {
				t.Errorf("ReadConfigurationFile() error output = %v, want %v", gotErr, tt.wantErr)
			}
			if gotLog := o.LogOutput(); gotLog != tt.wantLog {
				t.Errorf("ReadConfigurationFile() log output = %v, want %v", gotLog, tt.wantLog)
			}
		})
	}
}

func Test_appData(t *testing.T) {
	savedState := SaveEnvVarForTesting(appDataVar)
	os.Setenv(appDataVar, SecureAbsolutePathForTesting("."))
	defer func() {
		savedState.RestoreForTesting()
	}()
	tests := []struct {
		name    string
		state   *SavedEnvVar
		want    string
		want1   bool
		wantOut string
		wantErr string
		wantLog string
	}{
		{
			name:  "value is set",
			state: &SavedEnvVar{Name: appDataVar, Value: "appData!", Set: true},
			want:  "appData!",
			want1: true,
		},
		{
			name:    "value is not set",
			state:   &SavedEnvVar{Name: appDataVar},
			wantLog: "level='info' environment variable='APPDATA' msg='not set'\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.state.RestoreForTesting()
			o := NewOutputDeviceForTesting()
			got, got1 := appData(o)
			if got != tt.want {
				t.Errorf("appData() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("appData() got1 = %v, want %v", got1, tt.want1)
			}
			if gotOut := o.Stdout(); gotOut != tt.wantOut {
				t.Errorf("appData() console output = %q, want %q", gotOut, tt.wantOut)
			}
			if gotErr := o.Stderr(); gotErr != tt.wantErr {
				t.Errorf("appData() error output = %q, want %q", gotErr, tt.wantErr)
			}
			if gotLog := o.LogOutput(); gotLog != tt.wantLog {
				t.Errorf("appData() log output = %q, want %q", gotLog, tt.wantLog)
			}
		})
	}
}

func TestConfiguration_StringValue(t *testing.T) {
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
				t.Errorf("Configuration.StringValue() gotValue = %v, want %v", gotValue, tt.wantValue)
			}
			if gotOk != tt.wantOk {
				t.Errorf("Configuration.StringValue() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}
