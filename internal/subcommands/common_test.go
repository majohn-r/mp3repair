package subcommands

import (
	"bytes"
	"flag"
	"mp3/internal"
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
	type args struct {
		appDataPath string
		args        []string
	}
	tests := []struct {
		name  string
		args  args
		want  CommandProcessor
		want1 []string
		want2 bool
		wantW string
	}{
		{
			name:  "error handling", args:  args{appDataPath: "./mp3", args: nil},
		},
		{
			name:  "call ls",
			args:  args{appDataPath: ".", args: []string{"mp3.exe", "ls", "-track=true"}},
			want:  newLs(internal.EmptyConfiguration(), flag.NewFlagSet("ls", flag.ExitOnError)),
			want1: []string{"-track=true"},
			want2: true,
		},
		{
			name:  "call check",
			args:  args{appDataPath: ".", args: []string{"mp3.exe", "check", "-integrity=false"}},
			want:  newCheck(internal.EmptyConfiguration(), flag.NewFlagSet("check", flag.ExitOnError)),
			want1: []string{"-integrity=false"},
			want2: true,
		},
		{
			name:  "call repair",
			args:  args{appDataPath: ".", args: []string{"mp3.exe", "repair"}},
			want:  newRepair(internal.EmptyConfiguration(), flag.NewFlagSet("repair", flag.ExitOnError)),
			want1: []string{},
			want2: true,
		},
		{
			name:  "call postRepair",
			args:  args{appDataPath: ".", args: []string{"mp3.exe", "postRepair"}},
			want:  newPostRepair(internal.EmptyConfiguration(), flag.NewFlagSet("postRepair", flag.ExitOnError)),
			want1: []string{},
			want2: true,
		},
		{
			name:  "call default command",
			args:  args{appDataPath: ".", args: []string{"mp3.exe"}},
			want:  newLs(internal.EmptyConfiguration(), flag.NewFlagSet("ls", flag.ExitOnError)),
			want1: []string{"ls"},
			want2: true,
		},
		{
			name:  "call invalid command",
			args:  args{appDataPath: ".", args: []string{"mp3.exe", "no such command"}},
			wantW: "There is no command named \"no such command\"; valid commands include [check ls postRepair repair].\n",
		},
		{
			name:  "pass arguments to default subcommand",
			args:  args{appDataPath: ".", args: []string{"mp3.exe", "-album", "-artist", "-track"}},
			want:  newLs(internal.EmptyConfiguration(), flag.NewFlagSet("ls", flag.ExitOnError)),
			want1: []string{"-album", "-artist", "-track"},
			want2: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &bytes.Buffer{}
			got, got1, got2 := ProcessCommand(w, tt.args.appDataPath, tt.args.args)
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
