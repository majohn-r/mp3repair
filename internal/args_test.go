package internal

import (
	"flag"
	"testing"
)

func TestProcessArgs(t *testing.T) {
	fnName := "ProcessArgs()"
	goodFlags := flag.NewFlagSet("cmd", flag.ContinueOnError)
	goodFlags.String("foo", "bar", "some silly flag")
	goodFlags2 := flag.NewFlagSet("cmd2", flag.ContinueOnError)
	goodFlags2.String("foo2", "bar2", "some sillier flag")
	type args struct {
		f    *flag.FlagSet
		args []string
	}
	tests := []struct {
		name   string
		args   args
		wantOk bool
		WantedOutput
	}{
		{
			name:   "no errors",
			args:   args{f: goodFlags, args: []string{}},
			wantOk: true,
		},
		{
			name: "errors",
			args: args{f: goodFlags2, args: []string{"-foo=bar"}},
			WantedOutput: WantedOutput{
				WantErrorOutput: "flag provided but not defined: -foo\nUsage of cmd2:\n  -foo2 string\n    \tsome sillier flag (default \"bar2\")\n",
				WantLogOutput: "level='error' arguments='[-foo=bar]' msg='flag provided but not defined: -foo'\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := NewOutputDeviceForTesting()
			if gotOk := ProcessArgs(o, tt.args.f, tt.args.args); gotOk != tt.wantOk {
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
