package subcommands

import (
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
	type args struct {
		appDataPath string
		args        []string
	}
	tests := []struct {
		name            string
		args            args
		wantCmd         CommandProcessor
		wantCallingArgs []string
		wantErr         error
	}{
		{
			name:            "call ls",
			args:            args{appDataPath: ".", args: []string{"mp3.exe", "ls", "-track=true"}},
			wantCmd:         newLs(internal.EmptyConfiguration(), flag.NewFlagSet("ls", flag.ExitOnError)),
			wantCallingArgs: []string{"-track=true"},
		},
		{
			name:            "call check",
			args:            args{appDataPath: ".", args: []string{"mp3.exe", "check", "-integrity=false"}},
			wantCmd:         newCheck(internal.EmptyConfiguration(), flag.NewFlagSet("check", flag.ExitOnError)),
			wantCallingArgs: []string{"-integrity=false"},
		},
		{
			name:            "call repair",
			args:            args{appDataPath: ".", args: []string{"mp3.exe", "repair", "-target=metadata"}},
			wantCmd:         newRepair(internal.EmptyConfiguration(), flag.NewFlagSet("repair", flag.ExitOnError)),
			wantCallingArgs: []string{"-target=metadata"},
		},
		{
			name:            "call default command",
			args:            args{appDataPath: ".", args: []string{"mp3.exe"}},
			wantCmd:         newLs(internal.EmptyConfiguration(), flag.NewFlagSet("ls", flag.ExitOnError)),
			wantCallingArgs: []string{"ls"},
		},
		{
			name:    "call invalid command",
			args:    args{appDataPath: ".", args: []string{"mp3.exe", "no such command"}},
			wantErr: noSuchSubcommandError("no such command", []string{"check", "ls", "repair"}),
		},
		{
			name:            "[#38] pass arguments to default subcommand",
			args:            args{appDataPath: ".", args: []string{"mp3.exe", "-album", "-artist", "-track"}},
			wantCmd:         newLs(internal.EmptyConfiguration(), flag.NewFlagSet("ls", flag.ExitOnError)),
			wantCallingArgs: []string{"-album", "-artist", "-track"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCmd, gotCallingArgs, gotErr := ProcessCommand(tt.args.appDataPath, tt.args.args)
			if !equalErrors(gotErr, tt.wantErr) {
				t.Errorf("%s gotErr = %v, want %v", fnName, gotErr, tt.wantErr)
			}
			if gotCmd == nil {
				if tt.wantCmd != nil {
					t.Errorf("%s gotCmd = %v, want %v", fnName, gotCmd, tt.wantCmd)
				}
			} else {
				if tt.wantCmd == nil {
					t.Errorf("%s gotCmd = %v, want %v", fnName, gotCmd, tt.wantCmd)
				} else {
					if gotCmd.name() != tt.wantCmd.name() {
						t.Errorf("%s gotCmd name = %v, want name %v", fnName, gotCmd.name(), tt.wantCmd.name())
					}
				}
			}
			if !reflect.DeepEqual(gotCallingArgs, tt.wantCallingArgs) {
				t.Errorf("%s gotCallingArgs = %v, want %v", fnName, gotCallingArgs, tt.wantCallingArgs)
			}
		})
	}
}

func equalErrors(gotErr error, wantErr error) bool {
	if gotErr == nil {
		return wantErr == nil
	}
	if wantErr == nil {
		return false
	}
	return gotErr.Error() == wantErr.Error()
}

func Test_selectSubCommand(t *testing.T) {
	fnName := "selectSubCommand()"
	type args struct {
		c            *internal.Configuration
		initializers []subcommandInitializer
		args         []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		// only handling error cases here, success cases are handled by TestProcessCommand
		{
			name:    "no initializers",
			args:    args{},
			wantErr: internalErrorNoSubCommandInitializers(),
		},
		{
			name:    "no default initializers",
			args:    args{initializers: []subcommandInitializer{{}}},
			wantErr: internalErrorIncorrectNumberOfDefaultSubcommands(0),
		},
		{
			name:    "no default initializers",
			args:    args{initializers: []subcommandInitializer{{defaultSubCommand: true}, {defaultSubCommand: true}}},
			wantErr: internalErrorIncorrectNumberOfDefaultSubcommands(2),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, gotErr := selectSubCommand(tt.args.c, tt.args.initializers, tt.args.args)
			if !equalErrors(gotErr, tt.wantErr) {
				t.Errorf("%s gotErr = %v, want %v", fnName, gotErr, tt.wantErr)
			}
		})
	}
}
