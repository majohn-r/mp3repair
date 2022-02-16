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
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCmd, gotCallingArgs := ProcessCommand(tt.args.args)
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
