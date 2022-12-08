package commands

import (
	"reflect"
	"strings"
	"testing"

	"github.com/majohn-r/output"
)

func Test_finalYear(t *testing.T) {
	fnName := "finalYear()"
	type args struct {
		timestamp string
	}
	tests := []struct {
		name string
		args
		want int
		output.WantedRecording
	}{
		{
			name: "normal",
			args: args{timestamp: "2022-08-09T12:32:21-04:00"},
			want: 2022,
		},
		{
			name: "weird time",
			args: args{timestamp: "in the year 2525"},
			want: 2021,
			WantedRecording: output.WantedRecording{
				Error: "The build time \"in the year 2525\" cannot be parsed: parsing time \"in the year 2525\" as \"2006-01-02T15:04:05Z07:00\": cannot parse \"in the year 2525\" as \"2006\".\n",
				Log:   "level='error' error='parsing time \"in the year 2525\" as \"2006-01-02T15:04:05Z07:00\": cannot parse \"in the year 2525\" as \"2006\"' value='in the year 2525' msg='parse error'\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
	fnName := "formatCopyright()"
	type args struct {
		firstYear int
		lastYear  int
	}
	tests := []struct {
		name string
		args
		want string
	}{
		{
			name: "older than first year",
			args: args{
				firstYear: 2021,
				lastYear:  2020,
			},
			want: "Copyright © 2021 Marc Johnson",
		},
		{
			name: "same as first year",
			args: args{
				firstYear: 2021,
				lastYear:  2021,
			},
			want: "Copyright © 2021 Marc Johnson",
		},
		{
			name: "newer than first year",
			args: args{
				firstYear: 2021,
				lastYear:  2025,
			},
			want: "Copyright © 2021-2025 Marc Johnson",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatCopyright(tt.args.firstYear, tt.args.lastYear); got != tt.want {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func Test_formatBuildData(t *testing.T) {
	fnName := "formatBuildData()"
	type args struct {
		bD *BuildData
	}
	tests := []struct {
		name string
		args
		want []string
	}{
		{
			name: "success",
			args: args{
				bD: &BuildData{
					GoVersion: "go1.x",
					Dependencies: []string{
						"github.com/bogem/id3v2/v2 v2.1.2",
						"github.com/lestrrat-go/strftime v1.0.6",
					},
				},
			},
			want: []string{
				" - Go version: go1.x",
				" - Dependency: github.com/bogem/id3v2/v2 v2.1.2",
				" - Dependency: github.com/lestrrat-go/strftime v1.0.6",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatBuildData(tt.args.bD); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func Test_reportAbout(t *testing.T) {
	fnName := "reportAbout()"
	type args struct {
		data []string
	}
	tests := []struct {
		name string
		args
		output.WantedRecording
	}{
		{
			name: "normal",
			args: args{data: []string{
				"blah blah blah",
				formatCopyright(2021, 2025),
				"build data",
				" - foo/bar/baz v1.2.3",
			}},
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := output.NewRecorder()
			reportAbout(o, tt.args.data)
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
		})
	}
}

func Test_aboutCmd_Exec(t *testing.T) {
	fnName := "aboutCmd.Exec()"
	AboutBuildData = &BuildData{}
	type args struct {
		o    output.Bus
		args []string
	}
	tests := []struct {
		name string
		v    *aboutCmd
		args
		wantOk bool
	}{
		{
			name:   "for sake of completeness",
			v:      &aboutCmd{},
			args:   args{o: output.NewNilBus()},
			wantOk: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotOk := tt.v.Exec(tt.args.o, tt.args.args); gotOk != tt.wantOk {
				t.Errorf("%s = %v, want %v", fnName, gotOk, tt.wantOk)
			}
		})
	}
}

func Test_translateTimestamp(t *testing.T) {
	fnName := "translateTimestamp()"
	type args struct {
		t string
	}
	tests := []struct {
		name string
		args
		want string
	}{
		{
			name: "good time",
			args: args{t: "2022-08-10T13:29:57-04:00"},
			want: "Wednesday, August 10 2022, 13:29:57 EDT",
		},
		{
			name: "badly formatted time",
			args: args{t: "today is Monday!"},
			want: "today is Monday!",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := translateTimestamp(tt.args.t); !strings.HasPrefix(got, tt.want) {
				t.Errorf("%s = %q, want to start with %q", fnName, got, tt.want)
			}
		})
	}
}
