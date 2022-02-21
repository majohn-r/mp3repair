package subcommands

import (
	"flag"
	"reflect"
	"testing"
)

func TestProcessCommand(t *testing.T) {
	type args struct {
		args []string
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
			args:            args{args: []string{"mp3.exe", "ls", "-track=true"}},
			wantCmd:         newLs(flag.NewFlagSet("ls", flag.ExitOnError)),
			wantCallingArgs: []string{"-track=true"},
		},
		{
			name:            "call check",
			args:            args{args: []string{"mp3.exe", "check", "-integrity=false"}},
			wantCmd:         newCheck(flag.NewFlagSet("check", flag.ExitOnError)),
			wantCallingArgs: []string{"-integrity=false"},
		},
		{
			name:            "call repair",
			args:            args{args: []string{"mp3.exe", "repair", "-target=metadata"}},
			wantCmd:         newRepair(flag.NewFlagSet("repair", flag.ExitOnError)),
			wantCallingArgs: []string{"-target=metadata"},
		},
		{
			name:            "call default command",
			args:            args{args: []string{"mp3.exe"}},
			wantCmd:         newLs(flag.NewFlagSet("ls", flag.ExitOnError)),
			wantCallingArgs: []string{"ls"},
		},
		{
			name:            "call invalid command",
			args:            args{args: []string{"mp3.exe", "no such command"}},
			wantCmd:         nil,
			wantCallingArgs: nil,
			wantErr:         noSuchSubcommandError("no such command", []string{"check", "ls", "repair"}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCmd, gotCallingArgs, gotErr := ProcessCommand(tt.args.args)
			if !equalErrors(gotErr, tt.wantErr) {
				t.Errorf("ProcessCommand() gotErr = %v, want %v", gotErr, tt.wantErr)
			}
			if gotCmd == nil {
				if tt.wantCmd != nil {
					t.Errorf("ProcessCommand() gotCmd = %v, want %v", gotCmd, tt.wantCmd)
				}
			} else {
				if tt.wantCmd == nil {
					t.Errorf("ProcessCommand() gotCmd = %v, want %v", gotCmd, tt.wantCmd)
				} else {
					if gotCmd.name() != tt.wantCmd.name() {
						t.Errorf("ProcessCommand() gotCmd name = %v, want name %v", gotCmd.name(), tt.wantCmd.name())
					}
				}
			}
			if !reflect.DeepEqual(gotCallingArgs, tt.wantCallingArgs) {
				t.Errorf("ProcessCommand() gotCallingArgs = %v, want %v", gotCallingArgs, tt.wantCallingArgs)
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
	type args struct {
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
			name: "no initializers",
			args: args{
				initializers: nil,
			},
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
			_, _, gotErr := selectSubCommand(tt.args.initializers, tt.args.args)
			if !equalErrors(gotErr, tt.wantErr) {
				t.Errorf("selectSubCommand() gotErr = %v, want %v", gotErr, tt.wantErr)
			}
		})
	}
}
