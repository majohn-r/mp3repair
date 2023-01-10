package commands

import (
	"reflect"
	"runtime/debug"
	"strings"
	"testing"

	"github.com/majohn-r/output"
)

func Test_finalYear(t *testing.T) {
	const fnName = "finalYear()"
	type args struct {
		timestamp string
	}
	tests := map[string]struct {
		args
		want int
		output.WantedRecording
	}{
		"normal": {args: args{timestamp: "2022-08-09T12:32:21-04:00"}, want: 2022},
		"weird time": {
			args: args{timestamp: "in the year 2525"},
			want: 2021,
			WantedRecording: output.WantedRecording{
				Error: "The build time \"in the year 2525\" cannot be parsed: parsing time \"in the year 2525\" as \"2006-01-02T15:04:05Z07:00\": cannot parse \"in the year 2525\" as \"2006\".\n",
				Log:   "level='error' error='parsing time \"in the year 2525\" as \"2006-01-02T15:04:05Z07:00\": cannot parse \"in the year 2525\" as \"2006\"' value='in the year 2525' msg='parse error'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			if got := finalYear(o, tt.args.timestamp); got != tt.want {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
		})
	}
}

func Test_formatCopyright(t *testing.T) {
	const fnName = "formatCopyright()"
	type args struct {
		firstYear int
		lastYear  int
	}
	tests := map[string]struct {
		args
		want string
	}{
		"older than first year": {
			args: args{firstYear: 2021, lastYear: 2020},
			want: "Copyright © 2021 Marc Johnson",
		},
		"same as first year": {
			args: args{firstYear: 2021, lastYear: 2021},
			want: "Copyright © 2021 Marc Johnson",
		},
		"newer than first year": {
			args: args{firstYear: 2021, lastYear: 2025},
			want: "Copyright © 2021-2025 Marc Johnson",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := formatCopyright(tt.args.firstYear, tt.args.lastYear); got != tt.want {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func Test_formatBuildData(t *testing.T) {
	const fnName = "formatBuildData()"
	originalGoVersion := goVersion
	originalDependencies := buildDependencies
	defer func() {
		goVersion = originalGoVersion
		buildDependencies = originalDependencies
	}()
	tests := map[string]struct {
		version      string
		dependencies []string
		want         []string
	}{
		"success": {
			version:      "go1.x",
			dependencies: []string{"github.com/bogem/id3v2/v2 v2.1.2", "github.com/lestrrat-go/strftime v1.0.6"},
			want: []string{
				"Build Information",
				" - Go version: go1.x",
				" - Dependency: github.com/bogem/id3v2/v2 v2.1.2",
				" - Dependency: github.com/lestrrat-go/strftime v1.0.6",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			goVersion = tt.version
			buildDependencies = tt.dependencies
			if got := formatBuildData(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func Test_reportAbout(t *testing.T) {
	const fnName = "reportAbout()"
	type args struct {
		lines []string
	}
	tests := map[string]struct {
		args
		output.WantedRecording
	}{
		"normal": {
			args: args{lines: []string{"blah blah blah", formatCopyright(2021, 2025), "build data", " - foo/bar/baz v1.2.3"}},
			WantedRecording: output.WantedRecording{
				Console: "+------------------------------------+\n" +
					"| blah blah blah                     |\n" +
					"| Copyright © 2021-2025 Marc Johnson |\n" +
					"| build data                         |\n" +
					"|  - foo/bar/baz v1.2.3              |\n" +
					"+------------------------------------+\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			reportAbout(o, tt.args.lines)
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
		})
	}
}

func Test_aboutCmd_Exec(t *testing.T) {
	const fnName = "aboutCmd.Exec()"
	type args struct {
		o    output.Bus
		args []string
	}
	tests := map[string]struct {
		a *aboutCmd
		args
		wantOk bool
	}{
		"for sake of completeness": {a: &aboutCmd{}, args: args{o: output.NewNilBus()}, wantOk: true},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if gotOk := tt.a.Exec(tt.args.o, tt.args.args); gotOk != tt.wantOk {
				t.Errorf("%s = %v, want %v", fnName, gotOk, tt.wantOk)
			}
		})
	}
}

func Test_translateTimestamp(t *testing.T) {
	const fnName = "translateTimestamp()"
	type args struct {
		t string
	}
	tests := map[string]struct {
		args
		want string
	}{
		"good time":            {args: args{t: "2022-08-10T13:29:57-04:00"}, want: "Wednesday, August 10 2022, 13:29:57"},
		"badly formatted time": {args: args{t: "today is Monday!"}, want: "today is Monday!"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := translateTimestamp(tt.args.t); !strings.HasPrefix(got, tt.want) {
				t.Errorf("%s = %q, want to start with %q", fnName, got, tt.want)
			}
		})
	}
}

func TestInitBuildData(t *testing.T) {
	const fnName = "InitBuildData()"
	type args struct {
		f        func() (*debug.BuildInfo, bool)
		version  string
		creation string
	}
	tests := map[string]struct {
		args
		wantGoVersion         string
		wantBuildDependencies []string
	}{
		"happy path": {
			args: args{
				f: func() (*debug.BuildInfo, bool) {
					return &debug.BuildInfo{
						GoVersion: "go1.x",
						Deps: []*debug.Module{
							{Path: "blah/foo", Version: "v1.1.1"},
							{Path: "foo/blah/v2", Version: "v2.2.2"},
						},
					}, true
				},
				version:  "0.1.1",
				creation: "today",
			},
			wantGoVersion:         "go1.x",
			wantBuildDependencies: []string{"blah/foo v1.1.1", "foo/blah/v2 v2.2.2"},
		},
		"unhappy path": {
			args: args{
				f: func() (*debug.BuildInfo, bool) {
					return nil, false
				},
				version:  "unknown",
				creation: "tomorrow",
			},
			wantGoVersion: "unknown",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			originalGoVersion := goVersion
			originalDependencies := buildDependencies
			goVersion = ""
			buildDependencies = nil
			InitBuildData(tt.args.f, tt.args.version, tt.args.creation)
			if appVersion != tt.args.version {
				t.Errorf("%s = %v, appVersion %v", fnName, appVersion, tt.args.version)
			}
			if buildTimestamp != tt.args.creation {
				t.Errorf("%s = %v, buildTimestamp %v", fnName, buildTimestamp, tt.args.creation)
			}
			if got := GoVersion(); got != tt.wantGoVersion {
				t.Errorf("%s = %v, wantGoVersion %v", fnName, got, tt.wantGoVersion)
			}
			if got := BuildDependencies(); !reflect.DeepEqual(got, tt.wantBuildDependencies) {
				t.Errorf("%s = %v, wantBuildDependencies %v", fnName, got, tt.wantBuildDependencies)
			}
			goVersion = originalGoVersion
			buildDependencies = originalDependencies
		})
	}
}
