package files

import (
	"bytes"
	"flag"
	"reflect"
	"regexp"
	"testing"
)

func TestFileFlags_processArgs(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	flags := &FileFlags{
		f:             fs,
		topDirectory:  fs.String("topDir", ".", "top directory in which to look for music files"),
		fileExtension: fs.String("ext", DefaultFileExtension, "extension for music files"),
		albumRegex:    fs.String("albums", ".*", "regular expression of albums to select"),
		artistRegex:   fs.String("artists", ".*", "regular expression of artists to select"),
	}
	s := &Search{
		topDirectory:    ".",
		targetExtension: DefaultFileExtension,
		albumFilter:     regexp.MustCompile(".*"),
		artistFilter:    regexp.MustCompile(".*"),
	}
	savedWriter := fs.Output()
	usageWriter := &bytes.Buffer{}
	fs.SetOutput(usageWriter)
	fs.Usage()
	usage := usageWriter.String()
	fs.SetOutput(savedWriter)
	type args struct {
		args []string
	}
	tests := []struct {
		name       string
		ff         *FileFlags
		args       args
		want       *Search
		wantWriter string
	}{
		{name: "good arguments", ff: flags, args: args{args: nil}, want: s, wantWriter: ""},
		{name: "request help", ff: flags, args: args{args: []string{"-help"}}, want: nil, wantWriter: usage},
		{name: "request invalid argument", ff: flags, args: args{args: []string{"-foo"}}, want: nil, wantWriter: "flag provided but not defined: -foo\n" + usage},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer := &bytes.Buffer{}
			if got := tt.ff.ProcessArgs(writer, tt.args.args); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FileFlags.processArgs() = %v, want %v", got, tt.want)
			}
			if gotWriter := writer.String(); gotWriter != tt.wantWriter {
				t.Errorf("FileFlags.processArgs() = %v, want %v", gotWriter, tt.wantWriter)
			}
		})
	}
}

func Test_NewFileFlags(t *testing.T) {
	type args struct {
		fSet *flag.FlagSet
	}
	tests := []struct {
		name string
		args args
	}{
		{name: "default", args: args{fSet: flag.NewFlagSet("test", flag.ContinueOnError)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewFileFlags(tt.args.fSet); got == nil {
				t.Errorf("newFileFlags() = %v", got)
			}
		})
	}
}

func Test_validateTopLevelDirectory(t *testing.T) {
	type args struct {
		dir string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "is directory",
			args: args{dir: "."},
			want: true,
		},
		{
			name: "non-existent directory",
			args: args{dir: "no such file"},
			want: false,
		},
		{
			name: "file that is not a directory",
			args: args{dir: "utilities_test.go"},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validateTopLevelDirectory(tt.args.dir); got != tt.want {
				t.Errorf("validateTopLevelDirectory() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_validateExtension(t *testing.T) {
	type args struct {
		ext string
	}
	tests := []struct {
		name      string
		args      args
		wantValid bool
	}{
		{
			name: "valid extension",
			args: args{
				ext: ".mp3",
			},
			wantValid: true,
		},
		{
			name: "extension does not start with '.'",
			args: args{
				ext: "mp3",
			},
			wantValid: false,
		},
		{
			name: "extension contains multiple '.'",
			args: args{
				ext: ".m.p3",
			},
			wantValid: false,
		},
		{
			name: "extension contains invalid characters",
			args: args{
				ext: ".m[p3",
			},
			wantValid: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotValid := validateExtension(tt.args.ext); gotValid != tt.wantValid {
				t.Errorf("validateExtension() = %v, want %v", gotValid, tt.wantValid)
			}
		})
	}
}

func Test_validateRegexp(t *testing.T) {
	type args struct {
		pattern string
		name    string
	}
	tests := []struct {
		name         string
		args         args
		wantFilter   *regexp.Regexp
		wantBadRegex bool
	}{
		{
			name: "valid filter with regex",
			args: args{
				pattern: "^.*$",
				name:    "artist",
			},
			wantFilter:   regexp.MustCompile("^.*$"),
			wantBadRegex: false,
		},
		{
			name: "valid simple filter",
			args: args{
				pattern: "Beatles",
				name:    "artist",
			},
			wantFilter:   regexp.MustCompile("Beatles"),
			wantBadRegex: false,
		},
		{
			name: "invalid filter",
			args: args{
				pattern: "disc[",
				name:    "album",
			},
			wantFilter:   nil,
			wantBadRegex: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFilter, gotBadRegex := validateRegexp(tt.args.pattern, tt.args.name)
			if !tt.wantBadRegex && !reflect.DeepEqual(gotFilter, tt.wantFilter) {
				t.Errorf("validateRegexp() gotFilter = %v, want %v", gotFilter, tt.wantFilter)
			}
			if gotBadRegex != tt.wantBadRegex {
				t.Errorf("validateRegexp() gotBadRegex = %v, want %v", gotBadRegex, tt.wantBadRegex)
			}
		})
	}
}

func Test_validateSearchParameters(t *testing.T) {
	type args struct {
		dir     string
		ext     string
		albums  string
		artists string
	}
	tests := []struct {
		name              string
		args              args
		wantAlbumsFilter  *regexp.Regexp
		wantArtistsFilter *regexp.Regexp
		wantProblemsExist bool
	}{
		{
			name: "valid input",
			args: args{
				dir:     ".",
				ext:     ".mp3",
				albums:  ".*",
				artists: ".*",
			},
			wantAlbumsFilter:  regexp.MustCompile(".*"),
			wantArtistsFilter: regexp.MustCompile(".*"),
			wantProblemsExist: false,
		},
		{
			name: "bad extension 1",
			args: args{
				dir:     ".",
				ext:     "mp3",
				albums:  ".*",
				artists: ".*",
			},
			wantProblemsExist: true,
		},
		{
			name: "bad extension 2",
			args: args{
				dir:     ".",
				ext:     ".m.p3",
				albums:  ".*",
				artists: ".*",
			},
			wantProblemsExist: true,
		},
		{
			name: "bad extension 3",
			args: args{
				dir:     ".",
				ext:     ".mp[3",
				albums:  ".*",
				artists: ".*",
			},
			wantProblemsExist: true,
		},
		{
			name: "bad album filter",
			args: args{
				dir:     ".",
				ext:     ".mp3",
				albums:  ".[*",
				artists: ".*",
			},
			wantArtistsFilter: regexp.MustCompile(".*"),
			wantProblemsExist: true,
		},
		{
			name: "bad album filter",
			args: args{
				dir:     ".",
				ext:     ".mp3",
				albums:  ".*",
				artists: ".[*",
			},
			wantAlbumsFilter:  regexp.MustCompile(".*"),
			wantProblemsExist: true,
		},
		{
			name: "non-existent directory",
			args: args{
				dir:     "no such directory",
				ext:     ".mp3",
				albums:  ".*",
				artists: ".*",
			},
			wantAlbumsFilter:  regexp.MustCompile(".*"),
			wantArtistsFilter: regexp.MustCompile(".*"),
			wantProblemsExist: true,
		},
		{
			name: "directory is not a directory",
			args: args{
				dir:     "utilities_test.go",
				ext:     ".mp3",
				albums:  ".*",
				artists: ".*",
			},
			wantAlbumsFilter:  regexp.MustCompile(".*"),
			wantArtistsFilter: regexp.MustCompile(".*"),
			wantProblemsExist: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ff := &FileFlags{
				topDirectory:  &tt.args.dir,
				fileExtension: &tt.args.ext,
				albumRegex:    &tt.args.albums,
				artistRegex:   &tt.args.artists,
			}
			gotAlbumsFilter, gotArtistsFilter, gotProblemsExist := ff.validateSearchParameters()
			if !tt.wantProblemsExist {
				if !reflect.DeepEqual(gotAlbumsFilter, tt.wantAlbumsFilter) {
					t.Errorf("validateSearchParameters() gotAlbumsFilter = %v, want %v", gotAlbumsFilter, tt.wantAlbumsFilter)
				}
				if !reflect.DeepEqual(gotArtistsFilter, tt.wantArtistsFilter) {
					t.Errorf("validateSearchParameters() gotArtistsFilter = %v, want %v", gotArtistsFilter, tt.wantArtistsFilter)
				}
			}
			if gotProblemsExist != tt.wantProblemsExist {
				t.Errorf("validateSearchParameters() gotProblemsExist = %v, want %v", gotProblemsExist, tt.wantProblemsExist)
			}
		})
	}
}
