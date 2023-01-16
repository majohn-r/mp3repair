package internal

import (
	"flag"
	"os"
	"testing"

	"github.com/majohn-r/output"
)

func TestProcessArgs(t *testing.T) {
	const fnName = "ProcessArgs()"
	goodFlags := flag.NewFlagSet("cmd", flag.ContinueOnError)
	goodFlags.String("foo", "bar", "some silly flag")
	goodFlags2 := flag.NewFlagSet("cmd2", flag.ContinueOnError)
	goodFlags2.String("foo2", "bar2", "some sillier flag")
	goodFlags3 := flag.NewFlagSet("cmd3", flag.ContinueOnError)
	goodFlags3.String("foo3", "bar3", "some even sillier flag")
	saved := SaveEnvVarForTesting("NOSUCHVAR")
	os.Unsetenv("NOSUCHVAR")
	defer func() {
		saved.RestoreForTesting()
	}()
	type args struct {
		f    *flag.FlagSet
		args []string
	}
	tests := map[string]struct {
		args
		wantOk bool
		output.WantedRecording
	}{
		"no errors": {args: args{f: goodFlags, args: []string{}}, wantOk: true},
		"errors": {
			args: args{f: goodFlags2, args: []string{"-foo=bar"}},
			WantedRecording: output.WantedRecording{
				Error: "flag provided but not defined: -foo\n" +
					"Usage of cmd2:\n" +
					"  -foo2 string\n" +
					"    \tsome sillier flag (default \"bar2\")\n",
				Log: "level='error' arguments='[-foo=bar]' msg='flag provided but not defined: -foo'\n",
			},
		},
		"bad references": {
			args: args{f: goodFlags3, args: []string{"-foo3=$NOSUCHVAR"}},
			WantedRecording: output.WantedRecording{
				Error: "The value for argument \"-foo3=$NOSUCHVAR\" cannot be used: missing environment variables: [NOSUCHVAR].\n",
				Log:   "level='error' error='missing environment variables: [NOSUCHVAR]' value='-foo3=$NOSUCHVAR' msg='argument cannot be used'\n",
			},
		},
		"bad references2": {
			args: args{f: goodFlags3, args: []string{"-foo3=%NOSUCHVAR%"}},
			WantedRecording: output.WantedRecording{
				Error: "The value for argument \"-foo3=%NOSUCHVAR%\" cannot be used: missing environment variables: [NOSUCHVAR].\n",
				Log:   "level='error' error='missing environment variables: [NOSUCHVAR]' value='-foo3=%NOSUCHVAR%' msg='argument cannot be used'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			if gotOk := ProcessArgs(o, tt.args.f, tt.args.args); gotOk != tt.wantOk {
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
