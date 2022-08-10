package internal

import (
	"flag"
	"os"
	"testing"
)

func TestProcessArgs(t *testing.T) {
	fnName := "ProcessArgs()"
	goodFlags := flag.NewFlagSet("cmd", flag.ContinueOnError)
	goodFlags.String("foo", "bar", "some silly flag")
	goodFlags2 := flag.NewFlagSet("cmd2", flag.ContinueOnError)
	goodFlags2.String("foo2", "bar2", "some sillier flag")
	goodFlags3 := flag.NewFlagSet("cmd3", flag.ContinueOnError)
	goodFlags3.String("foo3", "bar3", "some even sillier flag")
	saved := SaveEnvVarForTesting("NOSUCHVAR")
	defer func() {
		saved.RestoreForTesting()
	}()
	os.Unsetenv("NOSUCHVAR")
	type args struct {
		f    *flag.FlagSet
		args []string
	}
	tests := []struct {
		name string
		args
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
				WantErrorOutput: "flag provided but not defined: -foo\n" +
					"Usage of cmd2:\n" +
					"  -foo2 string\n" +
					"    \tsome sillier flag (default \"bar2\")\n",
				WantLogOutput: "level='error' arguments='[-foo=bar]' msg='flag provided but not defined: -foo'\n",
			},
		},
		{
			name: "bad references",
			args: args{f: goodFlags3, args: []string{"-foo3=$NOSUCHVAR"}},
			WantedOutput: WantedOutput{
				WantErrorOutput: "The value for argument \"-foo3=$NOSUCHVAR\" cannot be used: missing environment variables: [NOSUCHVAR].\n",
				WantLogOutput:   "level='error' error='missing environment variables: [NOSUCHVAR]' value='-foo3=$NOSUCHVAR' msg='argument cannot be used'\n",
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
