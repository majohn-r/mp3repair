package cmd_test

import (
	"fmt"
	"mp3repair/cmd"
	"path/filepath"
	"reflect"
	"testing"

	cmd_toolkit "github.com/majohn-r/cmd-toolkit"
	"github.com/majohn-r/output"
)

type testFlag struct {
	value     any
	valueKind cmd.ValueType
	changed   bool
}

type testFlagProducer struct {
	flags map[string]testFlag
}

func (tfp testFlagProducer) Changed(name string) bool {
	if flag, found := tfp.flags[name]; found {
		return flag.changed
	} else {
		return false
	}
}

func (tfp testFlagProducer) GetBool(name string) (b bool, flagErr error) {
	if flag, found := tfp.flags[name]; found {
		if flag.valueKind == cmd.BoolType {
			if value, ok := flag.value.(bool); ok {
				b = value
			} else {
				flagErr = fmt.Errorf(
					"code error: value for %q name is supposed to be bool, but it isn't",
					name)
			}
		} else {
			flagErr = fmt.Errorf("flag %q is not marked boolean", name)
		}
	} else {
		flagErr = fmt.Errorf("flag %q does not exist", name)
	}
	return
}

func (tfp testFlagProducer) GetInt(name string) (i int, flagErr error) {
	if flag, found := tfp.flags[name]; found {
		if flag.valueKind == cmd.IntType {
			if value, ok := flag.value.(int); ok {
				i = value
			} else {
				flagErr = fmt.Errorf(
					"code error: value for %q name is supposed to be int, but it isn't",
					name)
			}
		} else {
			flagErr = fmt.Errorf("flag %q is not marked int", name)
		}
	} else {
		flagErr = fmt.Errorf("flag %q does not exist", name)
	}
	return
}

func (tfp testFlagProducer) GetString(name string) (s string, flagErr error) {
	if flag, found := tfp.flags[name]; found {
		if flag.valueKind == cmd.StringType {
			if value, ok := flag.value.(string); ok {
				s = value
			} else {
				flagErr = fmt.Errorf(
					"code error: value for %q name is supposed to be string, but it isn't",
					name)
			}
		} else {
			flagErr = fmt.Errorf("flag %q is not marked string", name)
		}
	} else {
		flagErr = fmt.Errorf("flag %q does not exist", name)
	}
	return
}

func TestReadFlags(t *testing.T) {
	type args struct {
		producer cmd.FlagProducer
		defs     *cmd.SectionFlags
	}
	tests := map[string]struct {
		args
		want  map[string]*cmd.FlagValue
		want1 []string
	}{
		"mix of weird flags in producer": {
			args: args{
				producer: testFlagProducer{flags: map[string]testFlag{
					"misidentifiedBool":   {value: true, valueKind: cmd.IntType},
					"misidentifiedInt":    {value: 6, valueKind: cmd.StringType},
					"misidentifiedString": {value: "foo", valueKind: cmd.BoolType},
					"bool":                {value: true, valueKind: cmd.BoolType},
					"int":                 {value: 6, valueKind: cmd.IntType},
					"string":              {value: "foo", valueKind: cmd.StringType},
				}},
				defs: &cmd.SectionFlags{
					SectionName: "whatever",
					Details: map[string]*cmd.FlagDetails{
						"misidentifiedBool":   {ExpectedType: cmd.BoolType},
						"misidentifiedInt":    {ExpectedType: cmd.IntType},
						"misidentifiedString": {ExpectedType: cmd.StringType},
						"bool":                {ExpectedType: cmd.BoolType},
						"int":                 {ExpectedType: cmd.IntType},
						"string":              {ExpectedType: cmd.StringType},
						"unexpected":          {ExpectedType: cmd.UnspecifiedType},
					},
				},
			},
			want: map[string]*cmd.FlagValue{
				"bool":   {UserSet: false, Value: true},
				"int":    {UserSet: false, Value: 6},
				"string": {UserSet: false, Value: "foo"},
			},
			want1: []string{
				"flag \"misidentifiedBool\" is not marked boolean",
				"flag \"misidentifiedInt\" is not marked int",
				"flag \"misidentifiedString\" is not marked string",
				"unknown type for flag --unexpected",
			},
		},
		"normal run": {
			args: args{
				producer: testFlagProducer{flags: map[string]testFlag{
					"specifiedBool": {
						value:     true,
						valueKind: cmd.BoolType,
						changed:   true,
					},
					"specifiedInt": {value: 6, valueKind: cmd.IntType, changed: true},
					"specifiedString": {
						value:     "foo",
						valueKind: cmd.StringType,
						changed:   true,
					},
					"unspecifiedBool":   {value: true, valueKind: cmd.BoolType},
					"unspecifiedInt":    {value: 6, valueKind: cmd.IntType},
					"unspecifiedString": {value: "foo", valueKind: cmd.StringType},
				}},
				defs: &cmd.SectionFlags{
					SectionName: "whatever",
					Details: map[string]*cmd.FlagDetails{
						"specifiedBool":     {ExpectedType: cmd.BoolType},
						"specifiedInt":      {ExpectedType: cmd.IntType},
						"specifiedString":   {ExpectedType: cmd.StringType},
						"unspecifiedBool":   {ExpectedType: cmd.BoolType},
						"unspecifiedInt":    {ExpectedType: cmd.IntType},
						"unspecifiedString": {ExpectedType: cmd.StringType},
					},
				},
			},
			want: map[string]*cmd.FlagValue{
				"specifiedBool":     {UserSet: true, Value: true},
				"specifiedInt":      {UserSet: true, Value: 6},
				"specifiedString":   {UserSet: true, Value: "foo"},
				"unspecifiedBool":   {UserSet: false, Value: true},
				"unspecifiedInt":    {UserSet: false, Value: 6},
				"unspecifiedString": {UserSet: false, Value: "foo"},
			},
		},
		"code error - missing details": {
			args: args{
				producer: nil,
				defs: &cmd.SectionFlags{
					Details: map[string]*cmd.FlagDetails{"no deets": nil},
				},
			},
			want:  map[string]*cmd.FlagValue{},
			want1: []string{"no details for flag \"no deets\""},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, got1 := cmd.ReadFlags(tt.args.producer, tt.args.defs)
			if len(got) != len(tt.want) {
				t.Errorf("ReadFlags() got = %d entries, want %d", len(got), len(tt.want))
			} else {
				for k, v := range got {
					if *v != *tt.want[k] {
						t.Errorf("ReadFlags() got[%s] = %v , want %v", k, v, tt.want[k])
					}
				}
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReadFlags() got = %v, want %v", got, tt.want)
			}
			if len(got1) != len(tt.want1) {
				t.Errorf("ReadFlags() got1 = %d errors, want %d errors", len(got1),
					len(tt.want1))
			} else {
				for i, e := range got1 {
					if g := e.Error(); g != tt.want1[i] {
						t.Errorf("ReadFlags() got1[%d] = %q, want %q errors", i, e,
							tt.want1[i])
					}
				}
			}
		})
	}
}

type testFlagDatum struct {
	shorthand string
	value     any
	usage     string
}

type testFlagConsumer struct {
	flags map[string]*testFlagDatum
}

func (t *testFlagConsumer) String(name, value, usage string) *string {
	t.flags[name] = &testFlagDatum{value: value, usage: usage}
	return nil
}

func (t *testFlagConsumer) StringP(name, shorthand, value, usage string) *string {
	t.flags[name] = &testFlagDatum{shorthand: shorthand, value: value, usage: usage}
	return nil
}

func (t *testFlagConsumer) Bool(name string, value bool, usage string) *bool {
	t.flags[name] = &testFlagDatum{value: value, usage: usage}
	return nil
}

func (t *testFlagConsumer) BoolP(name, shorthand string, value bool, usage string) *bool {
	t.flags[name] = &testFlagDatum{shorthand: shorthand, value: value, usage: usage}
	return nil
}

func (t *testFlagConsumer) Int(name string, value int, usage string) *int {
	t.flags[name] = &testFlagDatum{value: value, usage: usage}
	return nil
}

func (t *testFlagConsumer) IntP(name, shorthand string, value int, usage string) *int {
	t.flags[name] = &testFlagDatum{shorthand: shorthand, value: value, usage: usage}
	return nil
}

type testConfigSource struct{ accept bool }

func (tcs testConfigSource) BoolDefault(name string, b bool) (bool, error) {
	if tcs.accept {
		return b, nil
	} else {
		return false, fmt.Errorf("rejecting default for %s", name)
	}
}

func (tcs testConfigSource) IntDefault(name string, i *cmd_toolkit.IntBounds) (int, error) {
	if tcs.accept {
		return 0, nil
	} else {
		return 0, fmt.Errorf("rejecting default for %s", name)
	}
}

func (tcs testConfigSource) StringDefault(name, s string) (string, error) {
	if tcs.accept {
		return s, nil
	} else {
		return "", fmt.Errorf("rejecting default for %s", name)
	}
}

func TestFlagDetails_AddFlag(t *testing.T) {
	type args struct {
		c     cmd.ConfigSource
		flags *testFlagConsumer
		flag  cmd.Flag
	}
	tests := map[string]struct {
		f *cmd.FlagDetails
		args
		checkDetails  bool
		wantShorthand string
		wantValue     any
		wantUsage     string
		output.WantedRecording
	}{
		"rejected string": {
			f: &cmd.FlagDetails{
				Usage:        "a useful string",
				ExpectedType: cmd.StringType,
				DefaultValue: "",
			},
			args: args{
				c:     testConfigSource{accept: false},
				flags: &testFlagConsumer{},
				flag:  cmd.Flag{Section: "section", Name: "flag"},
			},
			WantedRecording: output.WantedRecording{
				Error: "The configuration file \"defaults.yaml\" contains an invalid value" +
					" for \"section\": rejecting default for flag.\n",
				Log: "level='error'" +
					" error='rejecting default for flag'" +
					" section='section'" +
					" msg='invalid content in configuration file'\n",
			},
		},
		"accepted string": {
			f: &cmd.FlagDetails{
				Usage:        "a useful string",
				ExpectedType: cmd.StringType,
				DefaultValue: "",
			},
			args: args{
				c:     testConfigSource{accept: true},
				flags: &testFlagConsumer{flags: map[string]*testFlagDatum{}},
				flag:  cmd.Flag{Section: "section", Name: "flag"},
			},
			checkDetails: true,
			wantValue:    "",
			wantUsage:    "a useful string (default \"\")",
		},
		"accepted string with shorthand": {
			f: &cmd.FlagDetails{
				AbbreviatedName: "f",
				Usage:           "a useful string",
				ExpectedType:    cmd.StringType,
				DefaultValue:    "",
			},
			args: args{
				c:     testConfigSource{accept: true},
				flags: &testFlagConsumer{flags: map[string]*testFlagDatum{}},
				flag:  cmd.Flag{Section: "section", Name: "flag"},
			},
			checkDetails:  true,
			wantShorthand: "f",
			wantValue:     "",
			wantUsage:     "a useful string (default \"\")",
		},
		"malformed string": {
			f: &cmd.FlagDetails{
				Usage:        "a useful string",
				ExpectedType: cmd.StringType,
				DefaultValue: true,
			},
			args: args{
				c:     testConfigSource{accept: true},
				flags: &testFlagConsumer{},
				flag:  cmd.Flag{Section: "section", Name: "flag"},
			},
			WantedRecording: output.WantedRecording{
				Error: "An internal error occurred: the type of flag \"flag\"'s value," +
					" 'true', is 'bool', but 'string' was expected.\n",
				Log: "level='error'" +
					" actual='bool'" +
					" error='default value mistyped'" +
					" expected='string'" +
					" flag='flag'" +
					" value='true'" +
					" msg='internal error'\n",
			},
		},
		"rejected int": {
			f: &cmd.FlagDetails{
				Usage:        "a useful int",
				ExpectedType: cmd.IntType,
				DefaultValue: cmd_toolkit.NewIntBounds(0, 1, 2),
			},
			args: args{
				c:     testConfigSource{accept: false},
				flags: &testFlagConsumer{},
				flag:  cmd.Flag{Section: "section", Name: "flag"},
			},
			WantedRecording: output.WantedRecording{
				Error: "The configuration file \"defaults.yaml\" contains an invalid value" +
					" for \"section\": rejecting default for flag.\n",
				Log: "level='error'" +
					" error='rejecting default for flag'" +
					" section='section'" +
					" msg='invalid content in configuration file'\n",
			},
		},
		"accepted int": {
			f: &cmd.FlagDetails{
				Usage:        "a useful int",
				ExpectedType: cmd.IntType,
				DefaultValue: cmd_toolkit.NewIntBounds(0, 1, 2),
			},
			args: args{
				c:     testConfigSource{accept: true},
				flags: &testFlagConsumer{flags: map[string]*testFlagDatum{}},
				flag:  cmd.Flag{Section: "section", Name: "flag"},
			},
			checkDetails: true,
			wantValue:    0,
			wantUsage:    "a useful int (default 0)",
		},
		"accepted int with shorthand": {
			f: &cmd.FlagDetails{
				AbbreviatedName: "i",
				Usage:           "a useful int",
				ExpectedType:    cmd.IntType,
				DefaultValue:    cmd_toolkit.NewIntBounds(0, 1, 2),
			},
			args: args{
				c:     testConfigSource{accept: true},
				flags: &testFlagConsumer{flags: map[string]*testFlagDatum{}},
				flag:  cmd.Flag{Section: "section", Name: "flag"},
			},
			checkDetails:  true,
			wantShorthand: "i",
			wantValue:     0,
			wantUsage:     "a useful int (default 0)",
		},
		"malformed int": {
			f: &cmd.FlagDetails{
				Usage:        "a useful int",
				ExpectedType: cmd.IntType,
				DefaultValue: true,
			},
			args: args{
				c:     testConfigSource{accept: true},
				flags: &testFlagConsumer{},
				flag:  cmd.Flag{Section: "section", Name: "flag"},
			},
			WantedRecording: output.WantedRecording{
				Error: "An internal error occurred: the type of flag \"flag\"'s value," +
					" 'true', is 'bool', but '*cmd_toolkit.IntBounds' was expected.\n",
				Log: "level='error'" +
					" actual='bool'" +
					" error='default value mistyped'" +
					" expected='*cmd_toolkit.IntBounds'" +
					" flag='flag'" +
					" value='true'" +
					" msg='internal error'\n",
			},
		},
		"rejected bool": {
			f: &cmd.FlagDetails{
				Usage:        "a useful bool",
				ExpectedType: cmd.BoolType,
				DefaultValue: false,
			},
			args: args{
				c:     testConfigSource{accept: false},
				flags: &testFlagConsumer{},
				flag:  cmd.Flag{Section: "section", Name: "flag"},
			},
			WantedRecording: output.WantedRecording{
				Error: "The configuration file \"defaults.yaml\" contains an invalid value" +
					" for \"section\": rejecting default for flag.\n",
				Log: "level='error'" +
					" error='rejecting default for flag'" +
					" section='section'" +
					" msg='invalid content in configuration file'\n",
			},
		},
		"accepted bool": {
			f: &cmd.FlagDetails{
				Usage:        "a useful bool",
				ExpectedType: cmd.BoolType,
				DefaultValue: false,
			},
			args: args{
				c:     testConfigSource{accept: true},
				flags: &testFlagConsumer{flags: map[string]*testFlagDatum{}},
				flag:  cmd.Flag{Section: "section", Name: "flag"},
			},
			checkDetails: true,
			wantValue:    false,
			wantUsage:    "a useful bool (default false)",
		},
		"accepted bool with shorthand": {
			f: &cmd.FlagDetails{
				AbbreviatedName: "b",
				Usage:           "a useful bool",
				ExpectedType:    cmd.BoolType,
				DefaultValue:    false,
			},
			args: args{
				c:     testConfigSource{accept: true},
				flags: &testFlagConsumer{flags: map[string]*testFlagDatum{}},
				flag:  cmd.Flag{Section: "section", Name: "flag"},
			},
			checkDetails:  true,
			wantShorthand: "b",
			wantValue:     false,
			wantUsage:     "a useful bool (default false)",
		},
		"malformed bool": {
			f: &cmd.FlagDetails{
				Usage:        "a useful bool",
				ExpectedType: cmd.BoolType,
				DefaultValue: cmd_toolkit.NewIntBounds(0, 1, 2),
			},
			args: args{
				c:     testConfigSource{accept: true},
				flags: &testFlagConsumer{},
				flag:  cmd.Flag{Section: "section", Name: "flag"},
			},
			WantedRecording: output.WantedRecording{
				Error: "An internal error occurred: the type of flag \"flag\"'s value," +
					" '&{0 1 2}', is '*cmd_toolkit.IntBounds', but 'bool' was expected.\n",
				Log: "level='error'" +
					" actual='*cmd_toolkit.IntBounds'" +
					" error='default value mistyped'" +
					" expected='bool'" +
					" flag='flag'" +
					" value='&{0 1 2}'" +
					" msg='internal error'\n",
			},
		},
		"malformed details": {
			f: &cmd.FlagDetails{
				Usage:        "useless",
				ExpectedType: cmd.UnspecifiedType,
				DefaultValue: true,
			},
			args: args{
				c:     testConfigSource{accept: true},
				flags: &testFlagConsumer{},
				flag:  cmd.Flag{Section: "section", Name: "flag"},
			},
			WantedRecording: output.WantedRecording{
				Error: "An internal error occurred: unspecified flag type; section" +
					" \"section\", flag \"flag\".\n",
				Log: "level='error'" +
					" default-type='bool'" +
					" default='true'" +
					" error='unspecified flag type'" +
					" flag='flag'" +
					" section='section'" +
					" specified-type='0'" +
					" msg='internal error'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			tt.f.AddFlag(o, tt.args.c, tt.args.flags, tt.args.flag)
			if tt.checkDetails {
				details := tt.args.flags.flags[tt.args.flag.Name]
				if got := details.shorthand; got != tt.wantShorthand {
					t.Errorf("FlagDetails.AddFlag() saved shorthand got = %q, want %q", got, tt.wantShorthand)
				}
				if got := details.value; got != tt.wantValue {
					t.Errorf("FlagDetails.AddFlag() saved value got = %v, want %v", got, tt.wantValue)
				}
				if got := details.usage; got != tt.wantUsage {
					t.Errorf("FlagDetails.AddFlag() saved usage got = %q, want %q", got, tt.wantUsage)
				}
			}
			o.Report(t, "FlagDetails.AddFlag()", tt.WantedRecording)
		})
	}
}

func TestAddFlags(t *testing.T) {
	originalSearchFlags := cmd.SearchFlags
	defer func() {
		cmd.SearchFlags = originalSearchFlags
	}()
	type args struct {
		flags           *testFlagConsumer
		defs            *cmd.SectionFlags
		includeSearches bool
	}
	tests := map[string]struct {
		args
		replaceSearchFlags *cmd.SectionFlags
		doReplacement      bool
		wantNames          []string
		output.WantedRecording
	}{
		"empty details with searches": {
			args: args{
				flags:           &testFlagConsumer{flags: map[string]*testFlagDatum{}},
				defs:            &cmd.SectionFlags{},
				includeSearches: true,
			},
			wantNames: []string{
				"albumFilter", "artistFilter", "topDir", "trackFilter", "extensions",
			},
		},
		"empty details without searches": {
			args: args{
				flags:           &testFlagConsumer{flags: map[string]*testFlagDatum{}},
				defs:            &cmd.SectionFlags{},
				includeSearches: false,
			},
			wantNames: []string{},
		},
		"empty details with bad searches": {
			args: args{
				flags:           &testFlagConsumer{flags: map[string]*testFlagDatum{}},
				defs:            &cmd.SectionFlags{},
				includeSearches: true,
			},
			replaceSearchFlags: &cmd.SectionFlags{
				SectionName: "common",
				Details: map[string]*cmd.FlagDetails{
					"albumFilter": {
						Usage:        "regular expression specifying which albums to select",
						ExpectedType: cmd.StringType,
						DefaultValue: ".*",
					},
					"artistFilter": {
						Usage:        "regular expression specifying which artists to select",
						ExpectedType: cmd.StringType,
						DefaultValue: ".*",
					},
					"topDir": {
						Usage:        "top directory specifying where to find mp3 files",
						ExpectedType: cmd.BoolType,
						DefaultValue: filepath.Join("%HOMEPATH%", "Music"),
					},
				},
			},
			doReplacement: true,
			wantNames:     []string{"albumFilter", "artistFilter"},
			WantedRecording: output.WantedRecording{
				Error: "An internal error occurred: the type of flag \"topDir\"'s value," +
					" '%HOMEPATH%\\Music', is 'string', but 'bool' was expected.\n",
				Log: "level='error'" +
					" actual='string'" +
					" error='default value mistyped'" +
					" expected='bool'" +
					" flag='topDir'" +
					" value='%HOMEPATH%\\Music'" +
					" msg='internal error'\n",
			},
		},
		"good details with good searches": {
			args: args{
				flags: &testFlagConsumer{flags: map[string]*testFlagDatum{}},
				defs: &cmd.SectionFlags{
					SectionName: "mySection",
					Details: map[string]*cmd.FlagDetails{
						"myFlag": {
							ExpectedType: cmd.BoolType,
							DefaultValue: false,
						},
					},
				},
				includeSearches: true,
			},
			wantNames: []string{
				"myFlag",
				"albumFilter",
				"artistFilter",
				"topDir",
				"trackFilter",
				"extensions",
			},
		},
		"good details without searches": {
			args: args{
				flags: &testFlagConsumer{flags: map[string]*testFlagDatum{}},
				defs: &cmd.SectionFlags{
					SectionName: "mySection",
					Details: map[string]*cmd.FlagDetails{
						"myFlag": {
							ExpectedType: cmd.BoolType,
							DefaultValue: false,
						},
					},
				},
				includeSearches: false,
			},
			wantNames: []string{"myFlag"},
		},
		"good details with bad searches": {
			args: args{
				flags: &testFlagConsumer{flags: map[string]*testFlagDatum{}},
				defs: &cmd.SectionFlags{
					SectionName: "mySection",
					Details: map[string]*cmd.FlagDetails{
						"myFlag": {
							ExpectedType: cmd.BoolType,
							DefaultValue: false,
						},
					},
				},
				includeSearches: true,
			},
			replaceSearchFlags: &cmd.SectionFlags{
				SectionName: "common",
				Details: map[string]*cmd.FlagDetails{
					"albumFilter": {
						Usage:        "regular expression specifying which albums to select",
						ExpectedType: cmd.StringType,
						DefaultValue: ".*",
					},
					"artistFilter": {
						Usage:        "regular expression specifying which artists to select",
						ExpectedType: cmd.StringType,
						DefaultValue: ".*",
					},
					"topDir": {
						Usage:        "top directory specifying where to find mp3 files",
						ExpectedType: cmd.BoolType,
						DefaultValue: filepath.Join("%HOMEPATH%", "Music"),
					},
				},
			},
			doReplacement: true,
			wantNames:     []string{"myFlag", "albumFilter", "artistFilter"},
			WantedRecording: output.WantedRecording{
				Error: "An internal error occurred: the type of flag \"topDir\"'s value," +
					" '%HOMEPATH%\\Music', is 'string', but 'bool' was expected.\n",
				Log: "level='error'" +
					" actual='string'" +
					" error='default value mistyped'" +
					" expected='bool'" +
					" flag='topDir'" +
					" value='%HOMEPATH%\\Music'" +
					" msg='internal error'\n",
			},
		},
		"bad details with good searches": {
			args: args{
				flags: &testFlagConsumer{flags: map[string]*testFlagDatum{}},
				defs: &cmd.SectionFlags{
					SectionName: "mySection",
					Details: map[string]*cmd.FlagDetails{
						"myFlag": {
							ExpectedType: cmd.BoolType,
							DefaultValue: false,
						},
						"myBadFlag": {
							ExpectedType: cmd.IntType,
							DefaultValue: false,
						},
						"myBetterFlag": {
							ExpectedType: cmd.BoolType,
							DefaultValue: true,
						},
						"myWorseFlag": {
							ExpectedType: cmd.BoolType,
							DefaultValue: "nope",
						},
					},
				},
				includeSearches: true,
			},
			wantNames: []string{
				"myFlag",
				"myBetterFlag",
				"albumFilter",
				"artistFilter",
				"topDir",
				"trackFilter",
				"extensions",
			},
			WantedRecording: output.WantedRecording{
				Error: "An internal error occurred: the type of flag \"myBadFlag\"'s value," +
					" 'false', is 'bool', but '*cmd_toolkit.IntBounds' was expected.\n" +
					"An internal error occurred: the type of flag \"myWorseFlag\"'s value," +
					" 'nope', is 'string', but 'bool' was expected.\n",
				Log: "level='error'" +
					" actual='bool'" +
					" error='default value mistyped'" +
					" expected='*cmd_toolkit.IntBounds'" +
					" flag='myBadFlag'" +
					" value='false'" +
					" msg='internal error'\n" +
					"level='error'" +
					" actual='string'" +
					" error='default value mistyped'" +
					" expected='bool'" +
					" flag='myWorseFlag'" +
					" value='nope'" +
					" msg='internal error'\n",
			},
		},
		"bad details without searches": {
			args: args{
				flags: &testFlagConsumer{flags: map[string]*testFlagDatum{}},
				defs: &cmd.SectionFlags{
					SectionName: "mySection",
					Details: map[string]*cmd.FlagDetails{
						"myFlag": {
							ExpectedType: cmd.BoolType,
							DefaultValue: false,
						},
						"myBadFlag": {
							ExpectedType: cmd.IntType,
							DefaultValue: false,
						},
						"myBetterFlag": {
							ExpectedType: cmd.BoolType,
							DefaultValue: true,
						},
						"myWorseFlag": {
							ExpectedType: cmd.BoolType,
							DefaultValue: "nope",
						},
					},
				},
				includeSearches: false,
			},
			wantNames: []string{"myFlag", "myBetterFlag"},
			WantedRecording: output.WantedRecording{
				Error: "An internal error occurred: the type of flag \"myBadFlag\"'s value," +
					" 'false', is 'bool', but '*cmd_toolkit.IntBounds' was expected.\n" +
					"An internal error occurred: the type of flag \"myWorseFlag\"'s value," +
					" 'nope', is 'string', but 'bool' was expected.\n",
				Log: "level='error'" +
					" actual='bool'" +
					" error='default value mistyped'" +
					" expected='*cmd_toolkit.IntBounds'" +
					" flag='myBadFlag'" +
					" value='false'" +
					" msg='internal error'\n" +
					"level='error'" +
					" actual='string'" +
					" error='default value mistyped'" +
					" expected='bool'" +
					" flag='myWorseFlag'" +
					" value='nope'" +
					" msg='internal error'\n",
			},
		},
		"bad details with bad searches": {
			args: args{
				flags: &testFlagConsumer{flags: map[string]*testFlagDatum{}},
				defs: &cmd.SectionFlags{
					SectionName: "mySection",
					Details: map[string]*cmd.FlagDetails{
						"myFlag": {
							ExpectedType: cmd.BoolType,
							DefaultValue: false,
						},
						"myBadFlag": {
							ExpectedType: cmd.IntType,
							DefaultValue: false,
						},
						"myBetterFlag": {
							ExpectedType: cmd.BoolType,
							DefaultValue: true,
						},
						"myWorseFlag": {
							ExpectedType: cmd.BoolType,
							DefaultValue: "nope",
						},
						"myAbsentDetails": nil,
					},
				},
				includeSearches: true,
			},
			replaceSearchFlags: &cmd.SectionFlags{
				SectionName: "common",
				Details: map[string]*cmd.FlagDetails{
					"albumFilter": {
						Usage:        "regular expression specifying which albums to select",
						ExpectedType: cmd.StringType,
						DefaultValue: ".*",
					},
					"artistFilter": {
						Usage:        "regular expression specifying which artists to select",
						ExpectedType: cmd.StringType,
						DefaultValue: ".*",
					},
					"topDir": {
						Usage:        "top directory specifying where to find mp3 files",
						ExpectedType: cmd.BoolType,
						DefaultValue: filepath.Join("%HOMEPATH%", "Music"),
					},
				},
			},
			doReplacement: true,
			wantNames:     []string{"myFlag", "myBetterFlag", "albumFilter", "artistFilter"},
			WantedRecording: output.WantedRecording{
				Error: "" +
					"An internal error occurred: there are no details for flag" +
					" \"myAbsentDetails\".\n" +
					"An internal error occurred: the type of flag \"myBadFlag\"'s value," +
					" 'false', is 'bool', but '*cmd_toolkit.IntBounds' was expected.\n" +
					"An internal error occurred: the type of flag \"myWorseFlag\"'s value," +
					" 'nope', is 'string', but 'bool' was expected.\n" +
					"An internal error occurred: the type of flag \"topDir\"'s value," +
					" '%HOMEPATH%\\Music', is 'string', but 'bool' was expected.\n",
				Log: "level='error'" +
					" error='no details present'" +
					" flag='myAbsentDetails'" +
					" section='mySection'" +
					" msg='internal error'\n" +
					"level='error'" +
					" actual='bool'" +
					" error='default value mistyped'" +
					" expected='*cmd_toolkit.IntBounds'" +
					" flag='myBadFlag'" +
					" value='false'" +
					" msg='internal error'\n" +
					"level='error'" +
					" actual='string'" +
					" error='default value mistyped'" +
					" expected='bool'" +
					" flag='myWorseFlag'" +
					" value='nope'" +
					" msg='internal error'\n" +
					"level='error'" +
					" actual='string'" +
					" error='default value mistyped'" +
					" expected='bool'" +
					" flag='topDir'" +
					" value='%HOMEPATH%\\Music'" +
					" msg='internal error'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			c := cmd_toolkit.EmptyConfiguration()
			if tt.doReplacement {
				cmd.SearchFlags = tt.replaceSearchFlags
			} else {
				cmd.SearchFlags = originalSearchFlags
			}
			if tt.args.includeSearches {
				cmd.AddFlags(o, c, tt.args.flags, tt.args.defs, cmd.SearchFlags)
			} else {
				cmd.AddFlags(o, c, tt.args.flags, tt.args.defs)
			}
			for _, name := range tt.wantNames {
				if _, found := tt.args.flags.flags[name]; !found {
					t.Errorf("AddFlags() did not register %q", name)
				}
			}
			if got := len(tt.args.flags.flags); got != len(tt.wantNames) {
				t.Errorf("AddFlags() got %d registered flags, expected %d", got, len(tt.wantNames))
			}
			o.Report(t, "GetBool()", tt.WantedRecording)
		})
	}
}

func TestGetBool(t *testing.T) {
	type args struct {
		results  map[string]*cmd.FlagValue
		flagName string
	}
	tests := map[string]struct {
		args
		want    cmd.BoolValue
		wantErr bool
		output.WantedRecording
	}{
		"no results": {
			args:    args{},
			wantErr: true,
			WantedRecording: output.WantedRecording{
				Error: "An internal error occurred: no flag values exist.\n",
				Log: "level='error'" +
					" error='no results to extract flag values from'" +
					" msg='internal error'\n",
			},
		},
		"no such flag": {
			args:    args{results: map[string]*cmd.FlagValue{}, flagName: "myFlag"},
			wantErr: true,
			WantedRecording: output.WantedRecording{
				Error: "An internal error occurred: flag \"myFlag\" is not found.\n",
				Log: "level='error'" +
					" error='flag not found'" +
					" flag='myFlag'" +
					" msg='internal error'\n",
			},
		},
		"flag has no data": {
			args: args{
				results:  map[string]*cmd.FlagValue{"myFlag": nil},
				flagName: "myFlag"},
			wantErr: true,
			WantedRecording: output.WantedRecording{
				Error: "An internal error occurred: flag \"myFlag\" has no data.\n",
				Log: "level='error'" +
					" error='no data associated with flag'" +
					" flag='myFlag'" +
					" msg='internal error'\n",
			},
		},
		"flag not bool": {
			args: args{
				results:  map[string]*cmd.FlagValue{"myFlag": {Value: 1}},
				flagName: "myFlag"},
			wantErr: true,
			WantedRecording: output.WantedRecording{
				Error: "An internal error occurred: flag \"myFlag\" is not boolean (1).\n",
				Log: "level='error'" +
					" error='flag value not boolean'" +
					" flag='myFlag'" +
					" value='1'" +
					" msg='internal error'\n",
			},
		},
		"good boolean": {
			args: args{
				results:  map[string]*cmd.FlagValue{"myFlag": {Value: true, UserSet: true}},
				flagName: "myFlag"},
			want:    cmd.BoolValue{Value: true, UserSet: true},
			wantErr: false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			got, gotErr := cmd.GetBool(o, tt.args.results, tt.args.flagName)
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("GetBool() error = %v, wantErr %v", gotErr, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetBool() gotVal = %v, want %v", got, tt.want)
			}
			o.Report(t, "AddFlags()", tt.WantedRecording)
		})
	}
}

func TestGetInt(t *testing.T) {
	type args struct {
		results  map[string]*cmd.FlagValue
		flagName string
	}
	tests := map[string]struct {
		args
		want    cmd.IntValue
		wantErr bool
		output.WantedRecording
	}{
		"no results": {
			args:    args{},
			wantErr: true,
			WantedRecording: output.WantedRecording{
				Error: "An internal error occurred: no flag values exist.\n",
				Log: "level='error'" +
					" error='no results to extract flag values from'" +
					" msg='internal error'\n",
			},
		},
		"no such flag": {
			args:    args{results: map[string]*cmd.FlagValue{}, flagName: "myFlag"},
			wantErr: true,
			WantedRecording: output.WantedRecording{
				Error: "An internal error occurred: flag \"myFlag\" is not found.\n",
				Log: "level='error'" +
					" error='flag not found'" +
					" flag='myFlag'" +
					" msg='internal error'\n",
			},
		},
		"flag has no data": {
			args: args{
				results:  map[string]*cmd.FlagValue{"myFlag": nil},
				flagName: "myFlag"},
			wantErr: true,
			WantedRecording: output.WantedRecording{
				Error: "An internal error occurred: flag \"myFlag\" has no data.\n",
				Log: "level='error'" +
					" error='no data associated with flag'" +
					" flag='myFlag'" +
					" msg='internal error'\n",
			},
		},
		"flag not int": {
			args: args{
				results:  map[string]*cmd.FlagValue{"myFlag": {Value: false}},
				flagName: "myFlag"},
			wantErr: true,
			WantedRecording: output.WantedRecording{
				Error: "An internal error occurred: flag \"myFlag\" is not an integer (false).\n",
				Log: "level='error'" +
					" error='flag value not int'" +
					" flag='myFlag'" +
					" value='false'" +
					" msg='internal error'\n",
			},
		},
		"good int": {
			args: args{
				results:  map[string]*cmd.FlagValue{"myFlag": {Value: 15, UserSet: true}},
				flagName: "myFlag"},
			want:    cmd.IntValue{Value: 15, UserSet: true},
			wantErr: false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			got, gotErr := cmd.GetInt(o, tt.args.results, tt.args.flagName)
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("GetInt() error = %v, wantErr %v", gotErr, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetInt() gotVal = %v, want %v", got, tt.want)
			}
			o.Report(t, "GetInt()", tt.WantedRecording)
		})
	}
}

func TestGetString(t *testing.T) {
	type args struct {
		results  map[string]*cmd.FlagValue
		flagName string
	}
	tests := map[string]struct {
		args
		want        cmd.StringValue
		wantUserSet bool
		wantErr     bool
		output.WantedRecording
	}{
		"no results": {
			args:    args{},
			wantErr: true,
			WantedRecording: output.WantedRecording{
				Error: "An internal error occurred: no flag values exist.\n",
				Log: "level='error'" +
					" error='no results to extract flag values from'" +
					" msg='internal error'\n",
			},
		},
		"no such flag": {
			args:    args{results: map[string]*cmd.FlagValue{}, flagName: "myFlag"},
			wantErr: true,
			WantedRecording: output.WantedRecording{
				Error: "An internal error occurred: flag \"myFlag\" is not found.\n",
				Log: "level='error'" +
					" error='flag not found'" +
					" flag='myFlag'" +
					" msg='internal error'\n",
			},
		},
		"flag has no data": {
			args: args{
				results:  map[string]*cmd.FlagValue{"myFlag": nil},
				flagName: "myFlag"},
			wantErr: true,
			WantedRecording: output.WantedRecording{
				Error: "An internal error occurred: flag \"myFlag\" has no data.\n",
				Log: "level='error'" +
					" error='no data associated with flag'" +
					" flag='myFlag'" +
					" msg='internal error'\n",
			},
		},
		"flag not string": {
			args: args{
				results:  map[string]*cmd.FlagValue{"myFlag": {Value: false}},
				flagName: "myFlag"},
			wantErr: true,
			WantedRecording: output.WantedRecording{
				Error: "An internal error occurred: flag \"myFlag\" is not a string (false).\n",
				Log: "level='error'" +
					" error='flag value not string'" +
					" flag='myFlag'" +
					" value='false'" +
					" msg='internal error'\n",
			},
		},
		"good string": {
			args: args{
				results:  map[string]*cmd.FlagValue{"myFlag": {Value: "foo", UserSet: true}},
				flagName: "myFlag"},
			wantErr: false,
			want:    cmd.StringValue{Value: "foo", UserSet: true},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			got, gotErr := cmd.GetString(o, tt.args.results, tt.args.flagName)
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("GetString() error = %v, wantErr %v", gotErr, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetString() gotVal = %v, want %v", got, tt.want)
			}
			o.Report(t, "GetString()", tt.WantedRecording)
		})
	}
}

func TestProcessFlagErrors(t *testing.T) {
	type args struct {
		eSlice []error
	}
	tests := map[string]struct {
		args
		wantOk bool
		output.WantedRecording
	}{
		"nil errors":   {args: args{eSlice: nil}, wantOk: true},
		"empty errors": {args: args{eSlice: []error{}}, wantOk: true},
		"errors": {
			args:   args{eSlice: []error{fmt.Errorf("generic flag error")}},
			wantOk: false,
			WantedRecording: output.WantedRecording{
				Error: "An internal error occurred: generic flag error.\n",
				Log: "level='error'" +
					" error='generic flag error'" +
					" msg='internal error'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			if gotOk := cmd.ProcessFlagErrors(o, tt.args.eSlice); gotOk != tt.wantOk {
				t.Errorf("ProcessFlagErrors() = %v, want %v", gotOk, tt.wantOk)
			}
			o.Report(t, "ProcessFlagErrors()", tt.WantedRecording)
		})
	}
}
