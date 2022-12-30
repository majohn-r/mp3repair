package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/majohn-r/output"
)

func Test_Configuration_SubConfiguration(t *testing.T) {
	fnName := "Configuration.SubConfiguration()"
	savedState := SaveEnvVarForTesting(appDataVar)
	os.Setenv(appDataVar, SecureAbsolutePathForTesting("."))
	oldAppPath := ApplicationPath()
	defer func() {
		savedState.RestoreForTesting()
		SetApplicationPathForTesting(oldAppPath)
	}()
	InitApplicationPath(output.NewNilBus())
	if err := CreateDefaultYamlFileForTesting(); err != nil {
		t.Errorf("%s error creating defaults.yaml: %v", fnName, err)
	}
	defer func() {
		DestroyDirectoryForTesting(fnName, "./mp3")
	}()
	testConfiguration, _ := ReadConfigurationFile(output.NewNilBus())
	type args struct {
		key string
	}
	tests := []struct {
		name string
		c    *Configuration
		args
		want *Configuration
	}{
		{name: "no configuration", c: &Configuration{}, args: args{}, want: EmptyConfiguration()},
		{name: "commons", c: testConfiguration, args: args{key: "common"}, want: testConfiguration.cMap["common"]},
		{name: "list", c: testConfiguration, args: args{key: "list"}, want: testConfiguration.cMap["list"]},
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
	oldAppPath := ApplicationPath()
	InitApplicationPath(output.NewNilBus())
	if err := CreateDefaultYamlFileForTesting(); err != nil {
		t.Errorf("%s error creating defaults.yaml: %v", fnName, err)
	}
	saved := SaveEnvVarForTesting("FOO")
	defer func() {
		savedState.RestoreForTesting()
		SetApplicationPathForTesting(oldAppPath)
		DestroyDirectoryForTesting(fnName, "./mp3")
		saved.RestoreForTesting()
	}()
	os.Unsetenv("FOO")
	testConfiguration, _ := ReadConfigurationFile(output.NewNilBus())
	type args struct {
		key          string
		defaultValue bool
	}
	tests := []struct {
		name string
		c    *Configuration
		args
		wantB bool
		wantE bool
	}{
		{
			name:  "string '1' to bool",
			c:     &Configuration{sMap: map[string]string{"foo": "1"}},
			args:  args{key: "foo", defaultValue: false},
			wantB: true,
		},
		{
			name:  "string '0' to bool",
			c:     &Configuration{sMap: map[string]string{"foo": "0"}},
			args:  args{key: "foo", defaultValue: true},
			wantB: false,
		},
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
			c:     testConfiguration.cMap["list"],
			args:  args{key: "includeTracks", defaultValue: false},
			wantB: true,
		},
		{
			name:  "boolean value default true",
			c:     testConfiguration.cMap["list"],
			args:  args{key: "includeTracks", defaultValue: true},
			wantB: true,
		},
		{
			name:  "boolean (string) value not parseable",
			c:     testConfiguration.cMap["unused"],
			args:  args{key: "value", defaultValue: true},
			wantE: true,
		},
		{
			name:  "boolean true from numeric",
			c:     &Configuration{iMap: map[string]int{"myKey": 1}},
			args:  args{key: "myKey", defaultValue: false},
			wantB: true,
		},
		{
			name:  "boolean false from numeric",
			c:     &Configuration{iMap: map[string]int{"myKey": 0}},
			args:  args{key: "myKey", defaultValue: true},
			wantB: false,
		},
		{
			name:  "boolean true from invalid numeric",
			c:     &Configuration{iMap: map[string]int{"myKey": 10}},
			args:  args{key: "myKey", defaultValue: false},
			wantE: true,
		},
		{
			name:  "boolean true from string 't'",
			c:     &Configuration{sMap: map[string]string{"myKey": "t"}},
			args:  args{key: "myKey", defaultValue: false},
			wantB: true,
		},
		{
			name:  "boolean true from string 'T'",
			c:     &Configuration{sMap: map[string]string{"myKey": "T"}},
			args:  args{key: "myKey", defaultValue: false},
			wantB: true,
		},
		{
			name:  "boolean true from string 'true'",
			c:     &Configuration{sMap: map[string]string{"myKey": "true"}},
			args:  args{key: "myKey", defaultValue: false},
			wantB: true,
		},
		{
			name:  "boolean true from string 'TRUE'",
			c:     &Configuration{sMap: map[string]string{"myKey": "TRUE"}},
			args:  args{key: "myKey", defaultValue: false},
			wantB: true,
		},
		{
			name:  "boolean true from string 'True'",
			c:     &Configuration{sMap: map[string]string{"myKey": "True"}},
			args:  args{key: "myKey", defaultValue: false},
			wantB: true,
		},
		{
			name:  "boolean true from string 'f'",
			c:     &Configuration{sMap: map[string]string{"myKey": "f"}},
			args:  args{key: "myKey", defaultValue: true},
			wantB: false,
		},
		{
			name:  "boolean true from string 'F'",
			c:     &Configuration{sMap: map[string]string{"myKey": "F"}},
			args:  args{key: "myKey", defaultValue: true},
			wantB: false,
		},
		{
			name:  "boolean true from string 'false'",
			c:     &Configuration{sMap: map[string]string{"myKey": "false"}},
			args:  args{key: "myKey", defaultValue: true},
			wantB: false,
		},
		{
			name:  "boolean true from string 'FALSE'",
			c:     &Configuration{sMap: map[string]string{"myKey": "FALSE"}},
			args:  args{key: "myKey", defaultValue: true},
			wantB: false,
		},
		{
			name:  "boolean true from string 'False'",
			c:     &Configuration{sMap: map[string]string{"myKey": "False"}},
			args:  args{key: "myKey", defaultValue: true},
			wantB: false,
		},
		{
			name:  "boolean from string with missing reference",
			c:     &Configuration{sMap: map[string]string{"key": "%FOO%"}},
			args:  args{key: "key"},
			wantE: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotB, e := tt.c.BoolDefault(tt.args.key, tt.args.defaultValue)
			gotE := e != nil
			if gotE != tt.wantE {
				t.Errorf("%s gotE %t, want %t", fnName, gotE, tt.wantE)
			}
			if !gotE && gotB != tt.wantB {
				t.Errorf("%s = %t, want %t", fnName, gotB, tt.wantB)
			}
		})
	}
}

func Test_Configuration_StringDefault(t *testing.T) {
	fnName := "Configuration.StringDefault()"
	savedState := SaveEnvVarForTesting(appDataVar)
	savedFoo := SaveEnvVarForTesting("FOO")
	savedBar := SaveEnvVarForTesting("BAR")
	os.Unsetenv("FOO")
	os.Unsetenv("BAR")
	os.Setenv(appDataVar, SecureAbsolutePathForTesting("."))
	oldAppPath := ApplicationPath()
	InitApplicationPath(output.NewNilBus())
	defer func() {
		savedFoo.RestoreForTesting()
		savedBar.RestoreForTesting()
		savedState.RestoreForTesting()
		SetApplicationPathForTesting(oldAppPath)
	}()
	if err := CreateDefaultYamlFileForTesting(); err != nil {
		t.Errorf("%s error creating defaults.yaml: %v", fnName, err)
	}
	defer func() {
		DestroyDirectoryForTesting(fnName, "./mp3")
	}()
	testConfiguration, _ := ReadConfigurationFile(output.NewNilBus())
	type args struct {
		key          string
		defaultValue string
	}
	tests := []struct {
		name string
		c    *Configuration
		args
		wantS   string
		wantErr bool
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
			c:     testConfiguration.cMap["list"],
			args:  args{key: "sort", defaultValue: "my default value"},
			wantS: "alpha",
		},
		{
			name:    "bad default",
			c:       &Configuration{sMap: map[string]string{"key": "$FOO"}},
			args:    args{key: "key"},
			wantErr: true,
		},
		{
			name:    "bad default",
			c:       EmptyConfiguration(),
			args:    args{key: "key", defaultValue: "boo%BAR%"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotS, gotErr := tt.c.StringDefault(tt.args.key, tt.args.defaultValue)
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("%s gotErr %v wantErr %t", fnName, gotErr, tt.wantErr)
			}
			if gotS != tt.wantS {
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
		name string
		args
		wantOk  bool
		wantErr bool
		output.WantedRecording
	}{
		{
			name:            "ordinary success",
			args:            args{path: "./configuration_test.go"},
			wantOk:          true,
			WantedRecording: output.WantedRecording{},
		},
		{
			name:    "look for dir!",
			args:    args{path: "."},
			wantErr: true,
			WantedRecording: output.WantedRecording{
				Error: "The configuration file \".\" is a directory.\n",
				Log:   "level='error' directory='.' fileName='.' msg='file is a directory'\n",
			},
		},
		{
			name: "non-existent file",
			args: args{path: "./no-such-file.txt"},
			WantedRecording: output.WantedRecording{
				Log: "level='info' directory='.' fileName='no-such-file.txt' msg='file does not exist'\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := output.NewRecorder()
			gotOk, err := verifyFileExists(o, tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("%s error = %v, wantErr %v", fnName, err, tt.wantErr)
				return
			}
			if gotOk != tt.wantOk {
				t.Errorf("%s = %v, want %v", fnName, gotOk, tt.wantOk)
			}
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
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
	oldAppPath := ApplicationPath()
	InitApplicationPath(output.NewNilBus())
	if err := CreateDefaultYamlFileForTesting(); err != nil {
		t.Errorf("%s error creating defaults.yaml: %v", fnName, err)
	}
	defer func() {
		savedState.RestoreForTesting()
		SetApplicationPathForTesting(oldAppPath)
		DestroyDirectoryForTesting(fnName, "./mp3")
	}()
	badDir := filepath.Join(".", "mp3", "fake")
	if err := Mkdir(badDir); err != nil {
		t.Errorf("%s error creating fake dir: %v", fnName, err)
	}
	badDir2 := filepath.Join(badDir, AppName)
	if err := Mkdir(badDir2); err != nil {
		t.Errorf("%s error creating fake dir mp3 folder: %v", fnName, err)
	}
	badFile := filepath.Join(badDir2, DefaultConfigFileName)
	if err := Mkdir(badFile); err != nil {
		t.Errorf("%s error creating defaults.yaml as a directory: %v", fnName, err)
	}
	yamlAsDir := SecureAbsolutePathForTesting(badFile)
	gibberishDir := filepath.Join(badDir2, AppName)
	if err := Mkdir(gibberishDir); err != nil {
		t.Errorf("%s error creating gibberish folder: %v", fnName, err)
	}
	if err := CreateFileForTestingWithContent(gibberishDir, DefaultConfigFileName, []byte("gibberish")); err != nil {
		t.Errorf("%s error creating gibberish defaults.yaml: %v", fnName, err)
	}
	tests := []struct {
		name   string
		state  *SavedEnvVar
		wantC  *Configuration
		wantOk bool
		output.WantedRecording
	}{
		{
			name:  "good",
			state: &SavedEnvVar{Name: appDataVar, Value: canonicalPath, Set: true},
			wantC: &Configuration{
				bMap: map[string]bool{},
				iMap: map[string]int{},
				sMap: map[string]string{},
				cMap: map[string]*Configuration{
					"check": {
						bMap: map[string]bool{"empty": true, "gaps": true, "integrity": false},
						iMap: map[string]int{},
						sMap: map[string]string{},
						cMap: map[string]*Configuration{},
					},
					"common": {
						bMap: map[string]bool{},
						iMap: map[string]int{},
						sMap: map[string]string{
							"albumFilter":  "^.*$",
							"artistFilter": "^.*$",
							"ext":          ".mpeg",
							"topDir":       ".",
						},
						cMap: map[string]*Configuration{},
					},
					"list": {
						bMap: map[string]bool{
							"annotate":       true,
							"includeAlbums":  false,
							"includeArtists": false,
							"includeTracks":  true,
						},
						iMap: map[string]int{},
						sMap: map[string]string{"sort": "alpha"},
						cMap: map[string]*Configuration{},
					},
					"repair": {
						bMap: map[string]bool{"dryRun": true},
						iMap: map[string]int{},
						sMap: map[string]string{},
						cMap: map[string]*Configuration{},
					},
					"unused": {
						bMap: map[string]bool{},
						iMap: map[string]int{},
						sMap: map[string]string{"value": "1.25"},
						cMap: map[string]*Configuration{}},
				},
			},
			wantOk: true,
			WantedRecording: output.WantedRecording{
				Error: "The key \"value\", with value '1.25', has an unexpected type float64.\n",
				Log: fmt.Sprintf("level='error' key='value' type='float64' value='1.25' msg='unexpected value type'\n"+
					"level='info' directory='%s' fileName='defaults.yaml' value='map[check:map[empty:true gaps:true integrity:false] common:map[albumFilter:^.*$ artistFilter:^.*$ ext:.mpeg topDir:.] list:map[annotate:true includeAlbums:false includeArtists:false includeTracks:true], map[sort:alpha] repair:map[dryRun:true] unused:map[value:1.25]]' msg='read configuration file'\n", mp3Path),
			},
		},
		{
			name:  "defaults.yaml is a directory",
			state: &SavedEnvVar{Name: appDataVar, Value: SecureAbsolutePathForTesting(badDir), Set: true},
			WantedRecording: output.WantedRecording{
				Error: fmt.Sprintf("The configuration file %q is a directory.\n", yamlAsDir),
				Log:   fmt.Sprintf("level='error' directory='%s' fileName='defaults.yaml' msg='file is a directory'\n", SecureAbsolutePathForTesting(badDir2)),
			},
		},
		{
			name:   "missing yaml",
			state:  &SavedEnvVar{Name: appDataVar, Value: SecureAbsolutePathForTesting(yamlAsDir), Set: true},
			wantC:  EmptyConfiguration(),
			wantOk: true,
			WantedRecording: output.WantedRecording{
				Log: fmt.Sprintf("level='info' directory='%s' fileName='defaults.yaml' msg='file does not exist'\n", SecureAbsolutePathForTesting(filepath.Join(yamlAsDir, AppName))),
			},
		},
		{
			name:  "malformed yaml",
			state: &SavedEnvVar{Name: appDataVar, Value: SecureAbsolutePathForTesting(badDir2), Set: true},
			WantedRecording: output.WantedRecording{
				Error: fmt.Sprintf(
					"The configuration file %q is not well-formed YAML: yaml: unmarshal errors:\n  line 1: cannot unmarshal !!str `gibberish` into map[string]interface {}.\n",
					SecureAbsolutePathForTesting(filepath.Join(gibberishDir, DefaultConfigFileName))),
				Log: fmt.Sprintf("level='error' directory='%s' error='yaml: unmarshal errors:\n  line 1: cannot unmarshal !!str `gibberish` into map[string]interface {}' fileName='defaults.yaml' msg='cannot unmarshal yaml content'\n", SecureAbsolutePathForTesting(gibberishDir)),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.state.RestoreForTesting()
			InitApplicationPath(output.NewNilBus())
			o := output.NewRecorder()
			gotC, gotOk := ReadConfigurationFile(o)
			if !reflect.DeepEqual(gotC, tt.wantC) {
				t.Errorf("%s gotC = %v, want %v", fnName, gotC, tt.wantC)
			}
			if gotOk != tt.wantOk {
				t.Errorf("%s gotOk = %v, want %v", fnName, gotOk, tt.wantOk)
			}
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
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
		name string
		c    *Configuration
		args
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

func TestConfiguration_String(t *testing.T) {
	fnName := "Configuration.String()"
	tests := []struct {
		name string
		c    *Configuration
		want string
	}{
		{
			name: "empty case",
			c:    EmptyConfiguration(),
		},
		{
			name: "busy case",
			c: &Configuration{
				bMap: map[string]bool{
					"f": false,
					"t": true,
				},
				iMap: map[string]int{
					"zero": 0,
					"one":  1,
				},
				sMap: map[string]string{
					"foo": "bar",
					"bar": "foo",
				},
				cMap: map[string]*Configuration{
					"empty": EmptyConfiguration(),
				},
			},
			want: "map[f:false t:true], map[one:1 zero:0], map[bar:foo foo:bar], map[empty:]",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.String(); got != tt.want {
				t.Errorf("%s = %q, want %q", fnName, got, tt.want)
			}
		})
	}
}

func TestNewIntBounds(t *testing.T) {
	fnName := "NewIntBounds()"
	type args struct {
		v1 int
		v2 int
		v3 int
	}
	tests := []struct {
		name string
		args
		want *IntBounds
	}{
		{
			name: "case 1",
			args: args{v1: 1, v2: 2, v3: 3},
			want: &IntBounds{minValue: 1, defaultValue: 2, maxValue: 3},
		},
		{
			name: "case 2",
			args: args{v1: 1, v2: 3, v3: 2},
			want: &IntBounds{minValue: 1, defaultValue: 2, maxValue: 3},
		},
		{
			name: "case 3",
			args: args{v1: 2, v2: 1, v3: 3},
			want: &IntBounds{minValue: 1, defaultValue: 2, maxValue: 3},
		},
		{
			name: "case 4",
			args: args{v1: 2, v2: 3, v3: 1},
			want: &IntBounds{minValue: 1, defaultValue: 2, maxValue: 3},
		},
		{
			name: "case 5",
			args: args{v1: 3, v2: 1, v3: 2},
			want: &IntBounds{minValue: 1, defaultValue: 2, maxValue: 3},
		},
		{
			name: "case 6",
			args: args{v1: 3, v2: 2, v3: 1},
			want: &IntBounds{minValue: 1, defaultValue: 2, maxValue: 3},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewIntBounds(tt.args.v1, tt.args.v2, tt.args.v3); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func TestConfiguration_IntDefault(t *testing.T) {
	fnName := "Configuration.IntDefault()"
	saved := SaveEnvVarForTesting("FOO")
	os.Unsetenv("FOO")
	defer func() {
		saved.RestoreForTesting()
	}()
	type args struct {
		key          string
		sortedBounds *IntBounds
	}
	tests := []struct {
		name string
		c    *Configuration
		args
		wantI   int
		wantErr bool
	}{
		{
			name:  "miss",
			c:     EmptyConfiguration(),
			args:  args{key: "k", sortedBounds: NewIntBounds(1, 2, 3)},
			wantI: 2,
		},
		{
			name:  "hit int value too low",
			c:     &Configuration{iMap: map[string]int{"k": -1}},
			args:  args{key: "k", sortedBounds: NewIntBounds(1, 2, 3)},
			wantI: 1,
		},
		{
			name:  "hit int value too high",
			c:     &Configuration{iMap: map[string]int{"k": 10}},
			args:  args{key: "k", sortedBounds: NewIntBounds(1, 2, 3)},
			wantI: 3,
		},
		{
			name:  "hit int value in the middle",
			c:     &Configuration{iMap: map[string]int{"k": 15}},
			args:  args{key: "k", sortedBounds: NewIntBounds(10, 20, 30)},
			wantI: 15,
		},
		{
			name:  "hit string value too low",
			c:     &Configuration{iMap: map[string]int{}, sMap: map[string]string{"k": "-1"}},
			args:  args{key: "k", sortedBounds: NewIntBounds(1, 2, 3)},
			wantI: 1,
		},
		{
			name:  "hit string value too high",
			c:     &Configuration{iMap: map[string]int{}, sMap: map[string]string{"k": "10"}},
			args:  args{key: "k", sortedBounds: NewIntBounds(1, 2, 3)},
			wantI: 3,
		},
		{
			name:  "hit string value in the middle",
			c:     &Configuration{iMap: map[string]int{}, sMap: map[string]string{"k": "15"}},
			args:  args{key: "k", sortedBounds: NewIntBounds(10, 20, 30)},
			wantI: 15,
		},
		{
			name:    "hit invalid string value",
			c:       &Configuration{iMap: map[string]int{}, sMap: map[string]string{"k": "foo"}},
			args:    args{key: "k", sortedBounds: NewIntBounds(10, 20, 30)},
			wantErr: true,
		},
		{
			name:    "hit bad references",
			c:       &Configuration{iMap: map[string]int{}, sMap: map[string]string{"k": "$FOO"}},
			args:    args{key: "k", sortedBounds: NewIntBounds(10, 20, 30)},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotI, err := tt.c.IntDefault(tt.args.key, tt.args.sortedBounds)
			gotErr := err != nil
			if gotErr != tt.wantErr {
				t.Errorf("%s gotErr %t wantErr %t", fnName, gotErr, tt.wantErr)
			}
			if !gotErr && gotI != tt.wantI {
				t.Errorf("%s = %d, want %d", fnName, gotI, tt.wantI)
			}
		})
	}
}

func Test_createConfiguration(t *testing.T) {
	fnName := "createConfiguration()"
	type args struct {
		data map[string]any
	}
	tests := []struct {
		name string
		args
		want *Configuration
		output.WantedRecording
	}{
		{
			name: "busy!",
			args: args{
				data: map[string]any{
					"boolValue":   true,
					"intValue":    1,
					"stringValue": "foo",
					"weirdValue":  1.2345,
					"mapValue":    map[string]any{},
				},
			},
			want: &Configuration{
				bMap: map[string]bool{"boolValue": true},
				iMap: map[string]int{"intValue": 1},
				sMap: map[string]string{"stringValue": "foo", "weirdValue": "1.2345"},
				cMap: map[string]*Configuration{"mapValue": EmptyConfiguration()},
			},
			WantedRecording: output.WantedRecording{
				Error: "The key \"weirdValue\", with value '1.2345', has an unexpected type float64.\n",
				Log:   "level='error' key='weirdValue' type='float64' value='1.2345' msg='unexpected value type'\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := output.NewRecorder()
			if got := CreateConfiguration(o, tt.args.data); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
		})
	}
}

func Test_readYaml(t *testing.T) {
	fnName := "readYaml()"
	type args struct {
		yfile []byte
	}
	tests := []struct {
		name string
		args
		wantData map[string]any
		wantErr  bool
	}{
		{
			name: "boolean truth",
			args: args{
				yfile: []byte("---\n" +
					"block:\n" +
					" b1: 1\n" +
					" b2: t\n" +
					" b3: true\n" +
					" b4: True\n" +
					" b5: TRUE\n"),
			},
			wantData: map[string]any{
				"block": map[string]any{
					"b1": 1,
					"b2": "t",
					"b3": true,
					"b4": true,
					"b5": true,
				},
			},
		},
		{
			name: "boolean falsehood",
			args: args{
				yfile: []byte("---\n" +
					"block:\n" +
					" b1: 0\n" +
					" b2: f\n" +
					" b3: false\n" +
					" b4: False\n" +
					" b5: FALSE\n"),
			},
			wantData: map[string]any{
				"block": map[string]any{
					"b1": 0,
					"b2": "f",
					"b3": false,
					"b4": false,
					"b5": false,
				},
			},
		},
		{
			name: "numerics",
			args: args{
				yfile: []byte("---\n" +
					"block:\n" +
					" b1: 100\n" +
					" b2: 0x64\n" +
					" b3: 0144\n"),
			},
			wantData: map[string]any{
				"block": map[string]any{
					"b1": 100,
					"b2": 100,
					"b3": 100,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotData, err := readYaml(tt.args.yfile)
			if (err != nil) != tt.wantErr {
				t.Errorf("%s error = %v, wantErr %v", fnName, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotData, tt.wantData) {
				t.Errorf("%s = %v, want %v", fnName, gotData, tt.wantData)
			}
		})
	}
}
