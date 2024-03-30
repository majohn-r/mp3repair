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
	if flag, ok := tfp.flags[name]; ok {
		return flag.changed
	} else {
		return false
	}
}

func (tfp testFlagProducer) GetBool(name string) (b bool, err error) {
	if flag, ok := tfp.flags[name]; ok {
		if flag.valueKind == cmd.BoolType {
			if value, ok := flag.value.(bool); ok {
				b = value
			} else {
				err = fmt.Errorf(
					"code error: value for %q name is supposed to be bool, but it isn't",
					name)
			}
		} else {
			err = fmt.Errorf("flag %q is not marked boolean", name)
		}
	} else {
		err = fmt.Errorf("flag %q does not exist", name)
	}
	return
}

func (tfp testFlagProducer) GetInt(name string) (i int, err error) {
	if flag, ok := tfp.flags[name]; ok {
		if flag.valueKind == cmd.IntType {
			if value, ok := flag.value.(int); ok {
				i = value
			} else {
				err = fmt.Errorf(
					"code error: value for %q name is supposed to be int, but it isn't",
					name)
			}
		} else {
			err = fmt.Errorf("flag %q is not marked int", name)
		}
	} else {
		err = fmt.Errorf("flag %q does not exist", name)
	}
	return
}

func (tfp testFlagProducer) GetString(name string) (s string, err error) {
	if flag, ok := tfp.flags[name]; ok {
		if flag.valueKind == cmd.StringType {
			if value, ok := flag.value.(string); ok {
				s = value
			} else {
				err = fmt.Errorf(
					"code error: value for %q name is supposed to be string, but it isn't",
					name)
			}
		} else {
			err = fmt.Errorf("flag %q is not marked string", name)
		}
	} else {
		err = fmt.Errorf("flag %q does not exist", name)
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
				defs: cmd.NewSectionFlags().WithSectionName("whatever").WithFlags(
					map[string]*cmd.FlagDetails{
						"misidentifiedBool": cmd.NewFlagDetails().WithExpectedType(
							cmd.BoolType),
						"misidentifiedInt": cmd.NewFlagDetails().WithExpectedType(
							cmd.IntType),
						"misidentifiedString": cmd.NewFlagDetails().WithExpectedType(
							cmd.StringType),
						"bool": cmd.NewFlagDetails().WithExpectedType(
							cmd.BoolType),
						"int": cmd.NewFlagDetails().WithExpectedType(
							cmd.IntType),
						"string": cmd.NewFlagDetails().WithExpectedType(
							cmd.StringType),
						"unexpected": cmd.NewFlagDetails().WithExpectedType(
							cmd.UnspecifiedType),
					},
				),
			},
			want: map[string]*cmd.FlagValue{
				"bool": cmd.NewFlagValue().WithExplicitlySet(false).WithValueType(
					cmd.BoolType).WithValue(true),
				"int": cmd.NewFlagValue().WithExplicitlySet(false).WithValueType(
					cmd.IntType).WithValue(6),
				"string": cmd.NewFlagValue().WithExplicitlySet(false).WithValueType(
					cmd.StringType).WithValue("foo"),
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
				defs: cmd.NewSectionFlags().WithSectionName("whatever").WithFlags(
					map[string]*cmd.FlagDetails{
						"specifiedBool": cmd.NewFlagDetails().WithExpectedType(
							cmd.BoolType),
						"specifiedInt": cmd.NewFlagDetails().WithExpectedType(
							cmd.IntType),
						"specifiedString": cmd.NewFlagDetails().WithExpectedType(
							cmd.StringType),
						"unspecifiedBool": cmd.NewFlagDetails().WithExpectedType(
							cmd.BoolType),
						"unspecifiedInt": cmd.NewFlagDetails().WithExpectedType(
							cmd.IntType),
						"unspecifiedString": cmd.NewFlagDetails().WithExpectedType(
							cmd.StringType),
					},
				),
			},
			want: map[string]*cmd.FlagValue{
				"specifiedBool": cmd.NewFlagValue().WithExplicitlySet(
					true).WithValueType(cmd.BoolType).WithValue(true),
				"specifiedInt": cmd.NewFlagValue().WithExplicitlySet(
					true).WithValueType(cmd.IntType).WithValue(6),
				"specifiedString": cmd.NewFlagValue().WithExplicitlySet(
					true).WithValueType(cmd.StringType).WithValue("foo"),
				"unspecifiedBool": cmd.NewFlagValue().WithExplicitlySet(
					false).WithValueType(cmd.BoolType).WithValue(true),
				"unspecifiedInt": cmd.NewFlagValue().WithExplicitlySet(
					false).WithValueType(cmd.IntType).WithValue(6),
				"unspecifiedString": cmd.NewFlagValue().WithExplicitlySet(
					false).WithValueType(cmd.StringType).WithValue("foo"),
			},
		},
		"code error - missing details": {
			args: args{
				producer: nil,
				defs: cmd.NewSectionFlags().WithFlags(map[string]*cmd.FlagDetails{
					"no deets": nil,
				}),
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
		c           cmd.ConfigSource
		flags       *testFlagConsumer
		sectionName string
		flagName    string
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
			f: cmd.NewFlagDetails().WithUsage("a useful string").WithExpectedType(
				cmd.StringType).WithDefaultValue(""),
			args: args{
				c:           testConfigSource{accept: false},
				flags:       &testFlagConsumer{},
				sectionName: "section",
				flagName:    "flag",
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
			f: cmd.NewFlagDetails().WithUsage("a useful string").WithExpectedType(
				cmd.StringType).WithDefaultValue(""),
			args: args{
				c:           testConfigSource{accept: true},
				flags:       &testFlagConsumer{flags: map[string]*testFlagDatum{}},
				sectionName: "section",
				flagName:    "flag",
			},
			checkDetails: true,
			wantValue:    "",
			wantUsage:    "a useful string (default \"\")",
		},
		"accepted string with shorthand": {
			f: cmd.NewFlagDetails().WithAbbreviatedName("f").WithUsage(
				"a useful string").WithExpectedType(cmd.StringType).WithDefaultValue(""),
			args: args{
				c:           testConfigSource{accept: true},
				flags:       &testFlagConsumer{flags: map[string]*testFlagDatum{}},
				sectionName: "section",
				flagName:    "flag",
			},
			checkDetails:  true,
			wantShorthand: "f",
			wantValue:     "",
			wantUsage:     "a useful string (default \"\")",
		},
		"malformed string": {
			f: cmd.NewFlagDetails().WithUsage("a useful string").WithExpectedType(
				cmd.StringType).WithDefaultValue(true),
			args: args{
				c:           testConfigSource{accept: true},
				flags:       &testFlagConsumer{},
				sectionName: "section",
				flagName:    "flag",
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
			f: cmd.NewFlagDetails().WithUsage("a useful int").WithExpectedType(
				cmd.IntType).WithDefaultValue(cmd_toolkit.NewIntBounds(0, 1, 2)),
			args: args{
				c:           testConfigSource{accept: false},
				flags:       &testFlagConsumer{},
				sectionName: "section",
				flagName:    "flag",
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
			f: cmd.NewFlagDetails().WithUsage("a useful int").WithExpectedType(
				cmd.IntType).WithDefaultValue(cmd_toolkit.NewIntBounds(0, 1, 2)),
			args: args{
				c:           testConfigSource{accept: true},
				flags:       &testFlagConsumer{flags: map[string]*testFlagDatum{}},
				sectionName: "section",
				flagName:    "flag",
			},
			checkDetails: true,
			wantValue:    0,
			wantUsage:    "a useful int (default 0)",
		},
		"accepted int with shorthand": {
			f: cmd.NewFlagDetails().WithAbbreviatedName("i").WithUsage(
				"a useful int").WithExpectedType(cmd.IntType).WithDefaultValue(
				cmd_toolkit.NewIntBounds(0, 1, 2)),
			args: args{
				c:           testConfigSource{accept: true},
				flags:       &testFlagConsumer{flags: map[string]*testFlagDatum{}},
				sectionName: "section",
				flagName:    "flag",
			},
			checkDetails:  true,
			wantShorthand: "i",
			wantValue:     0,
			wantUsage:     "a useful int (default 0)",
		},
		"malformed int": {
			f: cmd.NewFlagDetails().WithUsage("a useful int").WithExpectedType(
				cmd.IntType).WithDefaultValue(true),
			args: args{
				c:           testConfigSource{accept: true},
				flags:       &testFlagConsumer{},
				sectionName: "section",
				flagName:    "flag",
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
			f: cmd.NewFlagDetails().WithUsage("a useful bool").WithExpectedType(
				cmd.BoolType).WithDefaultValue(false),
			args: args{
				c:           testConfigSource{accept: false},
				flags:       &testFlagConsumer{},
				sectionName: "section",
				flagName:    "flag",
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
			f: cmd.NewFlagDetails().WithUsage("a useful bool").WithExpectedType(
				cmd.BoolType).WithDefaultValue(false),
			args: args{
				c:           testConfigSource{accept: true},
				flags:       &testFlagConsumer{flags: map[string]*testFlagDatum{}},
				sectionName: "section",
				flagName:    "flag",
			},
			checkDetails: true,
			wantValue:    false,
			wantUsage:    "a useful bool (default false)",
		},
		"accepted bool with shorthand": {
			f: cmd.NewFlagDetails().WithAbbreviatedName("b").WithUsage(
				"a useful bool").WithExpectedType(cmd.BoolType).WithDefaultValue(false),
			args: args{
				c:           testConfigSource{accept: true},
				flags:       &testFlagConsumer{flags: map[string]*testFlagDatum{}},
				sectionName: "section",
				flagName:    "flag",
			},
			checkDetails:  true,
			wantShorthand: "b",
			wantValue:     false,
			wantUsage:     "a useful bool (default false)",
		},
		"malformed bool": {
			f: cmd.NewFlagDetails().WithUsage("a useful bool").WithExpectedType(
				cmd.BoolType).WithDefaultValue(cmd_toolkit.NewIntBounds(0, 1, 2)),
			args: args{
				c:           testConfigSource{accept: true},
				flags:       &testFlagConsumer{},
				sectionName: "section",
				flagName:    "flag",
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
			f: cmd.NewFlagDetails().WithUsage("useless").WithExpectedType(
				cmd.UnspecifiedType).WithDefaultValue(true),
			args: args{
				c:           testConfigSource{accept: true},
				flags:       &testFlagConsumer{},
				sectionName: "section",
				flagName:    "flag",
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
			tt.f.AddFlag(o, tt.args.c, tt.args.flags, tt.args.sectionName, tt.args.flagName)
			if tt.checkDetails {
				details := tt.args.flags.flags[tt.args.flagName]
				if got := details.shorthand; got != tt.wantShorthand {
					t.Errorf("FlagDetails.AddFlag() saved shorthand got = %q, want %q", got,
						tt.wantShorthand)
				}
				if got := details.value; got != tt.wantValue {
					t.Errorf("FlagDetails.AddFlag() saved value got = %v, want %v", got,
						tt.wantValue)
				}
				if got := details.usage; got != tt.wantUsage {
					t.Errorf("FlagDetails.AddFlag() saved usage got = %q, want %q", got,
						tt.wantUsage)
				}
			}
			if differences, ok := o.Verify(tt.WantedRecording); !ok {
				for _, difference := range differences {
					t.Errorf("FlagDetails.AddFlag() %s", difference)
				}
			}
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
				defs:            cmd.NewSectionFlags(),
				includeSearches: true,
			},
			wantNames: []string{
				"albumFilter", "artistFilter", "topDir", "trackFilter", "extensions",
			},
		},
		"empty details without searches": {
			args: args{
				flags:           &testFlagConsumer{flags: map[string]*testFlagDatum{}},
				defs:            cmd.NewSectionFlags(),
				includeSearches: false,
			},
			wantNames: []string{},
		},
		"empty details with bad searches": {
			args: args{
				flags:           &testFlagConsumer{flags: map[string]*testFlagDatum{}},
				defs:            cmd.NewSectionFlags(),
				includeSearches: true,
			},
			replaceSearchFlags: cmd.NewSectionFlags().WithSectionName("common").WithFlags(
				map[string]*cmd.FlagDetails{
					"albumFilter": cmd.NewFlagDetails().WithUsage(
						"regular expression specifying which albums to select",
					).WithExpectedType(cmd.StringType).WithDefaultValue(".*"),
					"artistFilter": cmd.NewFlagDetails().WithUsage(
						"regular expression specifying which artists to select",
					).WithExpectedType(cmd.StringType).WithDefaultValue(".*"),
					"topDir": cmd.NewFlagDetails().WithUsage(
						"top directory specifying where to find mp3 files",
					).WithExpectedType(cmd.BoolType).WithDefaultValue(
						filepath.Join("%HOMEPATH%", "Music")),
				},
			),
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
				defs: cmd.NewSectionFlags().WithSectionName("mySection").WithFlags(
					map[string]*cmd.FlagDetails{
						"myFlag": cmd.NewFlagDetails().WithExpectedType(
							cmd.BoolType).WithDefaultValue(false),
					},
				),
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
				defs: cmd.NewSectionFlags().WithSectionName("mySection").WithFlags(
					map[string]*cmd.FlagDetails{
						"myFlag": cmd.NewFlagDetails().WithExpectedType(
							cmd.BoolType).WithDefaultValue(false),
					},
				),
				includeSearches: false,
			},
			wantNames: []string{"myFlag"},
		},
		"good details with bad searches": {
			args: args{
				flags: &testFlagConsumer{flags: map[string]*testFlagDatum{}},
				defs: cmd.NewSectionFlags().WithSectionName("mySection").WithFlags(
					map[string]*cmd.FlagDetails{
						"myFlag": cmd.NewFlagDetails().WithExpectedType(
							cmd.BoolType).WithDefaultValue(false),
					},
				),
				includeSearches: true,
			},
			replaceSearchFlags: cmd.NewSectionFlags().WithSectionName("common").WithFlags(
				map[string]*cmd.FlagDetails{
					"albumFilter": cmd.NewFlagDetails().WithUsage(
						"regular expression specifying which albums to select",
					).WithExpectedType(cmd.StringType).WithDefaultValue(".*"),
					"artistFilter": cmd.NewFlagDetails().WithUsage(
						"regular expression specifying which artists to select",
					).WithExpectedType(cmd.StringType).WithDefaultValue(".*"),
					"topDir": cmd.NewFlagDetails().WithUsage(
						"top directory specifying where to find mp3 files",
					).WithExpectedType(cmd.BoolType).WithDefaultValue(
						filepath.Join("%HOMEPATH%", "Music")),
				},
			),
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
				defs: cmd.NewSectionFlags().WithSectionName("mySection").WithFlags(
					map[string]*cmd.FlagDetails{
						"myFlag": cmd.NewFlagDetails().WithExpectedType(
							cmd.BoolType).WithDefaultValue(false),
						"myBadFlag": cmd.NewFlagDetails().WithExpectedType(
							cmd.IntType).WithDefaultValue(false),
						"myBetterFlag": cmd.NewFlagDetails().WithExpectedType(
							cmd.BoolType).WithDefaultValue(true),
						"myWorseFlag": cmd.NewFlagDetails().WithExpectedType(
							cmd.BoolType).WithDefaultValue("nope"),
					},
				),
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
				defs: cmd.NewSectionFlags().WithSectionName("mySection").WithFlags(
					map[string]*cmd.FlagDetails{
						"myFlag": cmd.NewFlagDetails().WithExpectedType(
							cmd.BoolType).WithDefaultValue(false),
						"myBadFlag": cmd.NewFlagDetails().WithExpectedType(
							cmd.IntType).WithDefaultValue(false),
						"myBetterFlag": cmd.NewFlagDetails().WithExpectedType(
							cmd.BoolType).WithDefaultValue(true),
						"myWorseFlag": cmd.NewFlagDetails().WithExpectedType(
							cmd.BoolType).WithDefaultValue("nope"),
					},
				),
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
				defs: cmd.NewSectionFlags().WithSectionName("mySection").WithFlags(
					map[string]*cmd.FlagDetails{
						"myFlag": cmd.NewFlagDetails().WithExpectedType(
							cmd.BoolType).WithDefaultValue(false),
						"myBadFlag": cmd.NewFlagDetails().WithExpectedType(
							cmd.IntType).WithDefaultValue(false),
						"myBetterFlag": cmd.NewFlagDetails().WithExpectedType(
							cmd.BoolType).WithDefaultValue(true),
						"myWorseFlag": cmd.NewFlagDetails().WithExpectedType(
							cmd.BoolType).WithDefaultValue("nope"),
						"myAbsentDetails": nil,
					},
				),
				includeSearches: true,
			},
			replaceSearchFlags: cmd.NewSectionFlags().WithSectionName("common").WithFlags(
				map[string]*cmd.FlagDetails{
					"albumFilter": cmd.NewFlagDetails().WithUsage(
						"regular expression specifying which albums to select",
					).WithExpectedType(cmd.StringType).WithDefaultValue(".*"),
					"artistFilter": cmd.NewFlagDetails().WithUsage(
						"regular expression specifying which artists to select",
					).WithExpectedType(cmd.StringType).WithDefaultValue(".*"),
					"topDir": cmd.NewFlagDetails().WithUsage(
						"top directory specifying where to find mp3 files",
					).WithExpectedType(cmd.BoolType).WithDefaultValue(
						filepath.Join("%HOMEPATH%", "Music")),
				},
			),
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
				if _, ok := tt.args.flags.flags[name]; !ok {
					t.Errorf("AddFlags() did not register %q", name)
				}
			}
			if got := len(tt.args.flags.flags); got != len(tt.wantNames) {
				t.Errorf("AddFlags() got %d registered flags, expected %d", got,
					len(tt.wantNames))
			}
			if differences, ok := o.Verify(tt.WantedRecording); !ok {
				for _, difference := range differences {
					t.Errorf("GetBool() %s", difference)
				}
			}
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
		wantVal     bool
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
		"flag not bool": {
			args: args{
				results:  map[string]*cmd.FlagValue{"myFlag": cmd.NewFlagValue().WithValue(1)},
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
				results: map[string]*cmd.FlagValue{"myFlag": cmd.NewFlagValue().WithValue(
					true).WithExplicitlySet(true)},
				flagName: "myFlag"},
			wantErr:     false,
			wantVal:     true,
			wantUserSet: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			gotVal, gotUserSet, err := cmd.GetBool(o, tt.args.results, tt.args.flagName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetBool() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotVal != tt.wantVal {
				t.Errorf("GetBool() gotVal = %v, want %v", gotVal, tt.wantVal)
			}
			if gotUserSet != tt.wantUserSet {
				t.Errorf("GetBool() gotUserSet = %v, want %v", gotUserSet, tt.wantUserSet)
			}
			if differences, ok := o.Verify(tt.WantedRecording); !ok {
				for _, difference := range differences {
					t.Errorf("AddFlags() %s", difference)
				}
			}
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
		wantVal     int
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
		"flag not int": {
			args: args{
				results: map[string]*cmd.FlagValue{"myFlag": cmd.NewFlagValue().WithValue(
					false)},
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
				results: map[string]*cmd.FlagValue{"myFlag": cmd.NewFlagValue().WithValue(
					15).WithExplicitlySet(true)},
				flagName: "myFlag"},
			wantErr:     false,
			wantVal:     15,
			wantUserSet: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			gotVal, gotUserSet, err := cmd.GetInt(o, tt.args.results, tt.args.flagName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetInt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotVal != tt.wantVal {
				t.Errorf("GetInt() gotVal = %v, want %v", gotVal, tt.wantVal)
			}
			if gotUserSet != tt.wantUserSet {
				t.Errorf("GetInt() gotUserSet = %v, want %v", gotUserSet, tt.wantUserSet)
			}
			if differences, ok := o.Verify(tt.WantedRecording); !ok {
				for _, difference := range differences {
					t.Errorf("GetInt() %s", difference)
				}
			}
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
		wantVal     string
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
				results: map[string]*cmd.FlagValue{"myFlag": cmd.NewFlagValue().WithValue(
					false)},
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
				results: map[string]*cmd.FlagValue{"myFlag": cmd.NewFlagValue().WithValue(
					"foo").WithExplicitlySet(true)},
				flagName: "myFlag"},
			wantErr:     false,
			wantVal:     "foo",
			wantUserSet: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			gotVal, gotUserSet, err := cmd.GetString(o, tt.args.results, tt.args.flagName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotVal != tt.wantVal {
				t.Errorf("GetString() gotVal = %v, want %v", gotVal, tt.wantVal)
			}
			if gotUserSet != tt.wantUserSet {
				t.Errorf("GetString() gotUserSet = %v, want %v", gotUserSet, tt.wantUserSet)
			}
			if differences, ok := o.Verify(tt.WantedRecording); !ok {
				for _, difference := range differences {
					t.Errorf("GetString() %s", difference)
				}
			}
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
			if differences, ok := o.Verify(tt.WantedRecording); !ok {
				for _, difference := range differences {
					t.Errorf("ProcessFlagErrors() %s", difference)
				}
			}
		})
	}
}
