package commands

import (
	"bytes"
	"flag"
	"fmt"
	"mp3/internal"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestProcessCommand(t *testing.T) {
	if err := internal.CreateDefaultYamlFileForTesting(); err != nil {
		t.Errorf("error creating defaults.yaml: %v", err)
	}
	fnName := "ProcessCommand()"
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, "./mp3")
	}()
	if err := internal.Mkdir("./mp3/mp3"); err != nil {
		t.Errorf("error creating defective ./mp3/mp3: %v", err)
	}
	if err := internal.Mkdir("./mp3/mp3/defaults.yaml"); err != nil {
		t.Errorf("error creating defective defaults.yaml: %v", err)
	}
	badDir := "badData"
	if err := internal.Mkdir(badDir); err != nil {
		t.Errorf("error creating bad data directory: %v", err)
	}
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, badDir)
	}()
	badMp3Dir := filepath.Join(badDir, "mp3")
	if err := internal.Mkdir(badMp3Dir); err != nil {
		t.Errorf("error creating bad data mp3 directory: %v", err)
	}
	if err := internal.CreateFileForTestingWithContent(badMp3Dir, "defaults.yaml", "command:\n    default: list\n"); err != nil {
		t.Errorf("error creating bad data defaults.yaml: %v", err)
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
		args  args
		want  CommandProcessor
		want1 []string
		want2 bool
		wantW string
	}{
		{
			name:  "problematic default command",
			state: &internal.SavedEnvVar{Name: "APPDATA", Value: internal.SecureAbsolutePathForTesting(badDir), Set: true},
			wantW: "The configuration file specifies \"list\" as the default command. There is no such command.\n",
		},
		{
			name:  "problematic input",
			state: &internal.SavedEnvVar{Name: "APPDATA", Value: internal.SecureAbsolutePathForTesting("./mp3"), Set: true},
		},
		{
			name:  "call ls",
			state: &internal.SavedEnvVar{Name: "APPDATA", Value: normalDir, Set: true},
			args:  args{args: []string{"mp3.exe", "ls", "-track=true"}},
			want:  newLs(internal.EmptyConfiguration(), flag.NewFlagSet("ls", flag.ExitOnError)),
			want1: []string{"-track=true"},
			want2: true,
		},
		{
			name:  "call check",
			state: &internal.SavedEnvVar{Name: "APPDATA", Value: normalDir, Set: true},
			args:  args{args: []string{"mp3.exe", "check", "-integrity=false"}},
			want:  newCheck(internal.EmptyConfiguration(), flag.NewFlagSet("check", flag.ExitOnError)),
			want1: []string{"-integrity=false"},
			want2: true,
		},
		{
			name:  "call repair",
			state: &internal.SavedEnvVar{Name: "APPDATA", Value: normalDir, Set: true},
			args:  args{args: []string{"mp3.exe", "repair"}},
			want:  newRepair(internal.EmptyConfiguration(), flag.NewFlagSet("repair", flag.ExitOnError)),
			want1: []string{},
			want2: true,
		},
		{
			name:  "call postRepair",
			state: &internal.SavedEnvVar{Name: "APPDATA", Value: normalDir, Set: true},
			args:  args{args: []string{"mp3.exe", "postRepair"}},
			want:  newPostRepair(internal.EmptyConfiguration(), flag.NewFlagSet("postRepair", flag.ExitOnError)),
			want1: []string{},
			want2: true,
		},
		{
			name:  "call default command",
			state: &internal.SavedEnvVar{Name: "APPDATA", Value: normalDir, Set: true},
			args:  args{args: []string{"mp3.exe"}},
			want:  newLs(internal.EmptyConfiguration(), flag.NewFlagSet("ls", flag.ExitOnError)),
			want1: []string{"ls"},
			want2: true,
		},
		{
			name:  "call invalid command",
			state: &internal.SavedEnvVar{Name: "APPDATA", Value: normalDir, Set: true},
			args:  args{args: []string{"mp3.exe", "no such command"}},
			wantW: "There is no command named \"no such command\"; valid commands include [check ls postRepair repair].\n",
		},
		{
			name:  "pass arguments to default subcommand",
			state: &internal.SavedEnvVar{Name: "APPDATA", Value: normalDir, Set: true},
			args:  args{args: []string{"mp3.exe", "-album", "-artist", "-track"}},
			want:  newLs(internal.EmptyConfiguration(), flag.NewFlagSet("ls", flag.ExitOnError)),
			want1: []string{"-album", "-artist", "-track"},
			want2: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.state.RestoreForTesting()
			w := &bytes.Buffer{}
			got, got1, got2 := ProcessCommand(w, tt.args.args)
			if got == nil {
				if tt.want != nil {
					t.Errorf("%s got = %v, want %v", fnName, got, tt.want)
				}
			} else {
				if tt.want == nil {
					t.Errorf("%s got = %v, want %v", fnName, got, tt.want)
				} else {
					if got.name() != tt.want.name() {
						t.Errorf("%s got name = %v, want name %v", fnName, got.name(), tt.want.name())
					}
				}
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("%s got1 = %v, want %v", fnName, got1, tt.want1)
			}
			if got2 != tt.want2 {
				t.Errorf("%s got2 = %v, want %v", fnName, got2, tt.want2)
			}
			if gotW := w.String(); gotW != tt.wantW {
				t.Errorf("%s gotW = %v, want %v", fnName, gotW, tt.wantW)
			}
		})
	}
}

func Test_selectSubCommand(t *testing.T) {
	type args struct {
		c    *internal.Configuration
		i    []subcommandInitializer
		args []string
	}
	tests := []struct {
		name            string
		args            args
		wantCmd         CommandProcessor
		wantCallingArgs []string
		wantOk          bool
		wantW           string
	}{
		// only handling error cases here, success cases are handled by TestProcessCommand
		{
			name:  "no initializers",
			args:  args{},
			wantW: "An internal error has occurred: no commands are defined!\n",
		},
		{
			name:  "no default initializers",
			args:  args{i: []subcommandInitializer{{}}},
			wantW: "An internal error has occurred: there are 0 default commands!\n",
		},
		{
			name:  "too many default initializers",
			args:  args{i: []subcommandInitializer{{defaultSubCommand: true}, {defaultSubCommand: true}}},
			wantW: "An internal error has occurred: there are 2 default commands!\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &bytes.Buffer{}
			gotCmd, gotCallingArgs, gotOk := selectSubCommand(w, tt.args.c, tt.args.i, tt.args.args)
			if !reflect.DeepEqual(gotCmd, tt.wantCmd) {
				t.Errorf("selectSubCommand() gotCmd = %v, want %v", gotCmd, tt.wantCmd)
			}
			if !reflect.DeepEqual(gotCallingArgs, tt.wantCallingArgs) {
				t.Errorf("selectSubCommand() gotCallingArgs = %v, want %v", gotCallingArgs, tt.wantCallingArgs)
			}
			if gotOk != tt.wantOk {
				t.Errorf("selectSubCommand() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
			if gotW := w.String(); gotW != tt.wantW {
				t.Errorf("selectSubCommand() gotW = %v, want %v", gotW, tt.wantW)
			}
		})
	}
}

func Test_getDefaultSettings(t *testing.T) {
	topDir := filepath.Join(".", "defaultSettings")
	if err := internal.Mkdir(topDir); err != nil {
		t.Errorf("getDefaultSettings error creating defaultSettings directory: %v", err)
	}
	configDir := filepath.Join(topDir, internal.AppName)
	if err := internal.Mkdir(configDir); err != nil {
		t.Errorf("getDefaultSettings error creating mp3 directory: %v", err)
	}
	savedState := internal.SaveEnvVarForTesting("APPDATA")
	defer func() {
		savedState.RestoreForTesting()
		internal.DestroyDirectoryForTesting("getDefaultSettings", topDir)
	}()
	os.Setenv("APPDATA", internal.SecureAbsolutePathForTesting(topDir))
	type args struct {
		c *internal.Configuration
	}
	tests := []struct {
		name           string
		includeDefault bool
		defaultValue   string
		args           args
		wantM          map[string]bool
		wantOk         bool
		wantWErr       string
	}{
		{
			name: "no value defined",
			wantM: map[string]bool{
				lsCommand:         true,
				checkCommand:      false,
				repairCommand:     false,
				postRepairCommand: false,
			},
			wantOk: true,
		},
		{
			name:           "good value",
			includeDefault: true,
			defaultValue:   "check",
			wantM: map[string]bool{
				lsCommand:         false,
				checkCommand:      true,
				repairCommand:     false,
				postRepairCommand: false,
			},
			wantOk: true,
		},
		{
			name:           "bad value",
			includeDefault: true,
			defaultValue:   "list",
			wantWErr:       fmt.Sprintf(internal.USER_INVALID_DEFAULT_COMMAND, "list"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wErr := &bytes.Buffer{}
			var content string
			if tt.includeDefault {
				content = fmt.Sprintf("command:\n    default: %s\n", tt.defaultValue)
			} else {
				content = "command:\n    nodefault: true\n"
			}
			os.Remove(filepath.Join(configDir, "defaults.yaml"))
			if err := internal.CreateFileForTestingWithContent(configDir, "defaults.yaml", content); err != nil {
				t.Errorf("getDefaultSettings() error creating defaults.yaml")
			}
			if c, ok := internal.ReadConfigurationFile(os.Stderr); !ok {
				t.Errorf("getDefaultSettings() error reading defaults.yaml %q", content)
			} else {
				tt.args.c = c.SubConfiguration("command")
			}
			gotM, gotOk := getDefaultSettings(wErr, tt.args.c)
			if !reflect.DeepEqual(gotM, tt.wantM) {
				t.Errorf("getDefaultSettings() gotM = %v, want %v", gotM, tt.wantM)
			}
			if gotOk != tt.wantOk {
				t.Errorf("getDefaultSettings() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
			if gotWErr := wErr.String(); gotWErr != tt.wantWErr {
				t.Errorf("getDefaultSettings() gotWErr = %v, want %v", gotWErr, tt.wantWErr)
			}
		})
	}
}
