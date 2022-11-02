package commands

import (
	"flag"
	"fmt"
	"mp3/internal"
	"mp3/internal/output"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func makeCheck() CommandProcessor {
	cp, _ := newCheck(output.NewNilBus(), internal.EmptyConfiguration(), flag.NewFlagSet("check", flag.ExitOnError))
	return cp
}

func makeList() CommandProcessor {
	list, _ := newList(output.NewNilBus(), internal.EmptyConfiguration(), flag.NewFlagSet("list", flag.ExitOnError))
	return list
}

func makeRepair() CommandProcessor {
	r, _ := newRepair(output.NewNilBus(), internal.EmptyConfiguration(), flag.NewFlagSet("repair", flag.ExitOnError))
	return r
}

func makePostRepair() CommandProcessor {
	pr, _ := newPostRepair(output.NewNilBus(), internal.EmptyConfiguration(), flag.NewFlagSet("postRepair", flag.ExitOnError))
	return pr
}

func TestProcessCommand(t *testing.T) {
	fnName := "ProcessCommand()"
	if err := internal.CreateDefaultYamlFileForTesting(); err != nil {
		t.Errorf("%s error creating defaults.yaml: %v", fnName, err)
	}
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, "./mp3")
	}()
	if err := internal.Mkdir("./mp3/mp3"); err != nil {
		t.Errorf("%s error creating defective ./mp3/mp3: %v", fnName, err)
	}
	if err := internal.Mkdir("./mp3/mp3/defaults.yaml"); err != nil {
		t.Errorf("%s error creating defective defaults.yaml: %v", fnName, err)
	}
	badDir := "badData"
	if err := internal.Mkdir(badDir); err != nil {
		t.Errorf("%s error creating bad data directory: %v", fnName, err)
	}
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, badDir)
	}()
	badMp3Dir := filepath.Join(badDir, "mp3")
	if err := internal.Mkdir(badMp3Dir); err != nil {
		t.Errorf("%s error creating bad data mp3 directory: %v", fnName, err)
	}
	if err := internal.CreateFileForTestingWithContent(badMp3Dir, "defaults.yaml", []byte("command:\n    default: lister\n")); err != nil {
		t.Errorf("%s error creating bad data defaults.yaml: %v", fnName, err)
	}
	normalDir := internal.SecureAbsolutePathForTesting(".")
	savedState := internal.SaveEnvVarForTesting("APPDATA")
	defer func() {
		savedState.RestoreForTesting()
	}()
	type args struct {
		args []string
	}
	tests := []struct {
		name  string
		state *internal.SavedEnvVar
		args
		want  CommandProcessor
		want1 []string
		want2 bool
		output.WantedRecording
	}{
		{
			name:  "problematic default command",
			state: &internal.SavedEnvVar{Name: "APPDATA", Value: internal.SecureAbsolutePathForTesting(badDir), Set: true},
			WantedRecording: output.WantedRecording{
				Error: "The configuration file specifies \"lister\" as the default command. There is no such command.\n",
				Log: fmt.Sprintf("level='info' directory='%s' fileName='defaults.yaml' value='map[command:map[default:lister]]' msg='read configuration file'\n", internal.SecureAbsolutePathForTesting(badMp3Dir)) +
					"level='error' command='lister' msg='invalid default command'\n",
			},
		},
		{
			name:  "problematic input",
			state: &internal.SavedEnvVar{Name: "APPDATA", Value: internal.SecureAbsolutePathForTesting("./mp3"), Set: true},
			WantedRecording: output.WantedRecording{
				Error: fmt.Sprintf("The configuration file %q is a directory.\n", internal.SecureAbsolutePathForTesting(filepath.Join("./mp3", "mp3", "defaults.yaml"))),
				Log:   fmt.Sprintf("level='error' directory='%s' fileName='defaults.yaml' msg='file is a directory'\n", internal.SecureAbsolutePathForTesting(filepath.Join("./mp3", "mp3"))),
			},
		},
		{
			name:  "call list",
			state: &internal.SavedEnvVar{Name: "APPDATA", Value: normalDir, Set: true},
			args:  args{args: []string{"mp3.exe", "list", "-track=true"}},
			want:  makeList(),
			want1: []string{"-track=true"},
			want2: true,
			WantedRecording: output.WantedRecording{
				Error: "The key \"value\", with value '1.25', has an unexpected type float64.\n",
				Log: "level='error' key='value' type='float64' value='1.25' msg='unexpected value type'\n" +
					fmt.Sprintf("level='info' directory='%s' fileName='defaults.yaml' value='map[check:map[empty:true gaps:true integrity:false] common:map[albumFilter:^.*$ artistFilter:^.*$ ext:.mpeg topDir:.] list:map[annotate:true includeAlbums:false includeArtists:false includeTracks:true], map[sort:alpha] repair:map[dryRun:true] unused:map[value:1.25]]' msg='read configuration file'\n", internal.SecureAbsolutePathForTesting("mp3")),
			},
		},
		{
			name:  "call check",
			state: &internal.SavedEnvVar{Name: "APPDATA", Value: normalDir, Set: true},
			args:  args{args: []string{"mp3.exe", "check", "-integrity=false"}},
			want:  makeCheck(),
			want1: []string{"-integrity=false"},
			want2: true,
			WantedRecording: output.WantedRecording{
				Error: "The key \"value\", with value '1.25', has an unexpected type float64.\n",
				Log: "level='error' key='value' type='float64' value='1.25' msg='unexpected value type'\n" +
					fmt.Sprintf("level='info' directory='%s' fileName='defaults.yaml' value='map[check:map[empty:true gaps:true integrity:false] common:map[albumFilter:^.*$ artistFilter:^.*$ ext:.mpeg topDir:.] list:map[annotate:true includeAlbums:false includeArtists:false includeTracks:true], map[sort:alpha] repair:map[dryRun:true] unused:map[value:1.25]]' msg='read configuration file'\n", internal.SecureAbsolutePathForTesting("mp3")),
			},
		},
		{
			name:  "call repair",
			state: &internal.SavedEnvVar{Name: "APPDATA", Value: normalDir, Set: true},
			args:  args{args: []string{"mp3.exe", "repair"}},
			want:  makeRepair(),
			want1: []string{},
			want2: true,
			WantedRecording: output.WantedRecording{
				Error: "The key \"value\", with value '1.25', has an unexpected type float64.\n",
				Log: "level='error' key='value' type='float64' value='1.25' msg='unexpected value type'\n" +
					fmt.Sprintf("level='info' directory='%s' fileName='defaults.yaml' value='map[check:map[empty:true gaps:true integrity:false] common:map[albumFilter:^.*$ artistFilter:^.*$ ext:.mpeg topDir:.] list:map[annotate:true includeAlbums:false includeArtists:false includeTracks:true], map[sort:alpha] repair:map[dryRun:true] unused:map[value:1.25]]' msg='read configuration file'\n", internal.SecureAbsolutePathForTesting("mp3")),
			},
		},
		{
			name:  "call postRepair",
			state: &internal.SavedEnvVar{Name: "APPDATA", Value: normalDir, Set: true},
			args:  args{args: []string{"mp3.exe", "postRepair"}},
			want:  makePostRepair(),
			want1: []string{},
			want2: true,
			WantedRecording: output.WantedRecording{
				Error: "The key \"value\", with value '1.25', has an unexpected type float64.\n",
				Log: "level='error' key='value' type='float64' value='1.25' msg='unexpected value type'\n" +
					fmt.Sprintf("level='info' directory='%s' fileName='defaults.yaml' value='map[check:map[empty:true gaps:true integrity:false] common:map[albumFilter:^.*$ artistFilter:^.*$ ext:.mpeg topDir:.] list:map[annotate:true includeAlbums:false includeArtists:false includeTracks:true], map[sort:alpha] repair:map[dryRun:true] unused:map[value:1.25]]' msg='read configuration file'\n", internal.SecureAbsolutePathForTesting("mp3")),
			},
		},
		{
			name:  "call default command",
			state: &internal.SavedEnvVar{Name: "APPDATA", Value: normalDir, Set: true},
			args:  args{args: []string{"mp3.exe"}},
			want:  makeList(),
			want1: []string{"list"},
			want2: true,
			WantedRecording: output.WantedRecording{
				Error: "The key \"value\", with value '1.25', has an unexpected type float64.\n",
				Log: "level='error' key='value' type='float64' value='1.25' msg='unexpected value type'\n" +
					fmt.Sprintf("level='info' directory='%s' fileName='defaults.yaml' value='map[check:map[empty:true gaps:true integrity:false] common:map[albumFilter:^.*$ artistFilter:^.*$ ext:.mpeg topDir:.] list:map[annotate:true includeAlbums:false includeArtists:false includeTracks:true], map[sort:alpha] repair:map[dryRun:true] unused:map[value:1.25]]' msg='read configuration file'\n", internal.SecureAbsolutePathForTesting("mp3")),
			},
		},
		{
			name:  "call invalid command",
			state: &internal.SavedEnvVar{Name: "APPDATA", Value: normalDir, Set: true},
			args:  args{args: []string{"mp3.exe", "no such command"}},
			WantedRecording: output.WantedRecording{
				Error: "The key \"value\", with value '1.25', has an unexpected type float64.\n" +
					"There is no command named \"no such command\"; valid commands include [about check export list postRepair repair resetDatabase].\n",
				Log: "level='error' key='value' type='float64' value='1.25' msg='unexpected value type'\n" +
					fmt.Sprintf("level='info' directory='%s' fileName='defaults.yaml' value='map[check:map[empty:true gaps:true integrity:false] common:map[albumFilter:^.*$ artistFilter:^.*$ ext:.mpeg topDir:.] list:map[annotate:true includeAlbums:false includeArtists:false includeTracks:true], map[sort:alpha] repair:map[dryRun:true] unused:map[value:1.25]]' msg='read configuration file'\n", internal.SecureAbsolutePathForTesting("mp3")) +
					"level='error' command='no such command' msg='unrecognized command'\n",
			},
		},
		{
			name:  "pass arguments to default command",
			state: &internal.SavedEnvVar{Name: "APPDATA", Value: normalDir, Set: true},
			args:  args{args: []string{"mp3.exe", "-album", "-artist", "-track"}},
			want:  makeList(),
			want1: []string{"-album", "-artist", "-track"},
			want2: true,
			WantedRecording: output.WantedRecording{
				Error: "The key \"value\", with value '1.25', has an unexpected type float64.\n",
				Log: "level='error' key='value' type='float64' value='1.25' msg='unexpected value type'\n" +
					fmt.Sprintf("level='info' directory='%s' fileName='defaults.yaml' value='map[check:map[empty:true gaps:true integrity:false] common:map[albumFilter:^.*$ artistFilter:^.*$ ext:.mpeg topDir:.] list:map[annotate:true includeAlbums:false includeArtists:false includeTracks:true], map[sort:alpha] repair:map[dryRun:true] unused:map[value:1.25]]' msg='read configuration file'\n", internal.SecureAbsolutePathForTesting("mp3")),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.state.RestoreForTesting()
			o := output.NewRecorder()
			got, got1, got2 := ProcessCommand(o, tt.args.args)
			if got == nil {
				if tt.want != nil {
					t.Errorf("%s got = %v, want %v", fnName, got, tt.want)
				}
			} else {
				if tt.want == nil {
					t.Errorf("%s got = %v, want %v", fnName, got, tt.want)
				}
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("%s got1 = %v, want %v", fnName, got1, tt.want1)
			}
			if got2 != tt.want2 {
				t.Errorf("%s got2 = %v, want %v", fnName, got2, tt.want2)
			}
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
		})
	}
}

func Test_selectCommand(t *testing.T) {
	fnName := "selectCommand()"
	type args struct {
		c    *internal.Configuration
		i    []commandInitializer
		args []string
	}
	tests := []struct {
		name string
		args
		wantCmd         CommandProcessor
		wantCallingArgs []string
		wantOk          bool
		output.WantedRecording
	}{
		// only handling error cases here, success cases are handled by TestProcessCommand
		{
			name: "no initializers",
			args: args{},
			WantedRecording: output.WantedRecording{
				Error: "An internal error has occurred: no commands are defined!\n",
				Log:   "level='error' count='0' msg='incorrect number of commands'\n",
			},
		},
		{
			name: "no default initializers",
			args: args{i: []commandInitializer{{}}},
			WantedRecording: output.WantedRecording{
				Error: "An internal error has occurred: there are 0 default commands!\n",
				Log:   "level='error' count='0' msg='incorrect number of default commands'\n",
			},
		},
		{
			name: "too many default initializers",
			args: args{i: []commandInitializer{{defaultCommand: true}, {defaultCommand: true}}},
			WantedRecording: output.WantedRecording{
				Error: "An internal error has occurred: there are 2 default commands!\n",
				Log:   "level='error' count='2' msg='incorrect number of default commands'\n",
			},
		},
		{
			name: "unfortunate defaults",
			args: args{
				c: internal.CreateConfiguration(output.NewNilBus(), map[string]any{
					"list": map[string]any{
						"includeTracks": "no!!",
					},
				}),
				i: []commandInitializer{{
					name:           "list",
					defaultCommand: true,
					initializer:    newList,
				}},
			},
			WantedRecording: output.WantedRecording{
				Error: "The configuration file \"defaults.yaml\" contains an invalid value for \"list\": invalid boolean value \"no!!\" for -includeTracks: parse error.\n",
				Log:   "level='error' error='invalid boolean value \"no!!\" for -includeTracks: parse error' section='list' msg='invalid content in configuration file'\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := output.NewRecorder()
			gotCmd, gotCallingArgs, gotOk := selectCommand(o, tt.args.c, tt.args.i, tt.args.args)
			if !reflect.DeepEqual(gotCmd, tt.wantCmd) {
				t.Errorf("%s gotCmd = %v, want %v", fnName, gotCmd, tt.wantCmd)
			}
			if !reflect.DeepEqual(gotCallingArgs, tt.wantCallingArgs) {
				t.Errorf("%s gotCallingArgs = %v, want %v", fnName, gotCallingArgs, tt.wantCallingArgs)
			}
			if gotOk != tt.wantOk {
				t.Errorf("%s  gotOk = %v, want %v", fnName, gotOk, tt.wantOk)
			}
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
		})
	}
}

func Test_getDefaultSettings(t *testing.T) {
	fnName := "getDefaultSettings()"
	const (
		aboutCommand         = "about"
		checkCommand         = "check"
		exportCommand        = "export"
		listCommand          = "list"
		postRepairCommand    = "postRepair"
		repairCommand        = "repair"
		resetDatabaseCommand = "resetDatabase"
	)
	topDir := filepath.Join(".", "defaultSettings")
	if err := internal.Mkdir(topDir); err != nil {
		t.Errorf("%s error creating defaultSettings directory: %v", fnName, err)
	}
	configDir := filepath.Join(topDir, internal.AppName)
	if err := internal.Mkdir(configDir); err != nil {
		t.Errorf("%s error creating mp3 directory: %v", fnName, err)
	}
	savedCommandMap := commandMap
	savedState := internal.SaveEnvVarForTesting("APPDATA")
	defer func() {
		savedState.RestoreForTesting()
		commandMap = savedCommandMap
		internal.DestroyDirectoryForTesting("getDefaultSettings", topDir)
	}()
	os.Setenv("APPDATA", internal.SecureAbsolutePathForTesting(topDir))
	tests := []struct {
		name           string
		includeDefault bool
		defaultValue   string
		cmds           map[string]commandData
		wantM          map[string]bool
		wantOk         bool
		output.WantedRecording
	}{
		{
			name: "too many values defined in code",
			cmds: map[string]commandData{
				listCommand:          {isDefault: true},
				checkCommand:         {isDefault: true},
				exportCommand:        {},
				repairCommand:        {},
				postRepairCommand:    {},
				resetDatabaseCommand: {},
				aboutCommand:         {},
			},
			WantedRecording: output.WantedRecording{
				Error: "Internal error: 2 commands self-selected as default: [check list]; pick one!\n",
			},
		},
		{
			name: "no value defined",
			wantM: map[string]bool{
				listCommand:          true,
				checkCommand:         false,
				exportCommand:        false,
				repairCommand:        false,
				postRepairCommand:    false,
				resetDatabaseCommand: false,
				aboutCommand:         false,
			},
			wantOk: true,
		},
		{
			name:           "good value",
			includeDefault: true,
			defaultValue:   "check",
			wantM: map[string]bool{
				listCommand:          false,
				checkCommand:         true,
				exportCommand:        false,
				repairCommand:        false,
				postRepairCommand:    false,
				resetDatabaseCommand: false,
				aboutCommand:         false,
			},
			wantOk: true,
		},
		{
			name:           "bad value",
			includeDefault: true,
			defaultValue:   "lister",
			WantedRecording: output.WantedRecording{
				Error: "The configuration file specifies \"lister\" as the default command. There is no such command.\n",
				Log:   "level='error' command='lister' msg='invalid default command'\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var content string
			if tt.includeDefault {
				content = fmt.Sprintf("command:\n    default: %s\n", tt.defaultValue)
			} else {
				content = "command:\n    nodefault: true\n"
			}
			if tt.cmds == nil {
				commandMap = savedCommandMap
			} else {
				commandMap = tt.cmds
			}
			os.Remove(filepath.Join(configDir, "defaults.yaml"))
			if err := internal.CreateFileForTestingWithContent(configDir, "defaults.yaml", []byte(content)); err != nil {
				t.Errorf("%s error creating defaults.yaml: %v", fnName, err)
			}
			var c *internal.Configuration
			var ok bool
			if c, ok = internal.ReadConfigurationFile(output.NewNilBus()); !ok {
				t.Errorf("%s error reading defaults.yaml %q", fnName, content)
			}
			o := output.NewRecorder()
			gotM, gotOk := getDefaultSettings(o, c.SubConfiguration("command"))
			if !reflect.DeepEqual(gotM, tt.wantM) {
				t.Errorf("%s gotM = %v, want %v", fnName, gotM, tt.wantM)
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
