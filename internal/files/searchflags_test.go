package files

import (
	"bytes"
	"flag"
	"mp3/internal"
	"os"
	"reflect"
	"regexp"
	"testing"

	"github.com/spf13/viper"
)

func TestFileFlags_processArgs(t *testing.T) {
	fnName := "FileFlags.processArgs()"
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	flags := &SearchFlags{
		f:             fs,
		topDirectory:  fs.String("topDir", ".", "top directory in which to look for music files"),
		fileExtension: fs.String("ext", defaultFileExtension, "extension for music files"),
		albumRegex:    fs.String("albums", ".*", "regular expression of albums to select"),
		artistRegex:   fs.String("artists", ".*", "regular expression of artists to select"),
	}
	s := &Search{
		topDirectory:    ".",
		targetExtension: defaultFileExtension,
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
		sf         *SearchFlags
		args       args
		want       *Search
		wantWriter string
	}{
		{name: "good arguments", sf: flags, args: args{args: nil}, want: s, wantWriter: ""},
		{name: "request help", sf: flags, args: args{args: []string{"-help"}}, want: nil, wantWriter: usage},
		{name: "request invalid argument", sf: flags, args: args{args: []string{"-foo"}}, want: nil, wantWriter: "flag provided but not defined: -foo\n" + usage},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer := &bytes.Buffer{}
			if got := tt.sf.ProcessArgs(writer, tt.args.args); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
			if gotWriter := writer.String(); gotWriter != tt.wantWriter {
				t.Errorf("%s = %v, want %v", fnName, gotWriter, tt.wantWriter)
			}
		})
	}
}

func Test_NewFileFlags(t *testing.T) {
	oldHomePath := os.Getenv("HOMEPATH")
	defer func() {
		os.Setenv("HOMEPATH", oldHomePath)
	}()
	os.Setenv("HOMEPATH", ".")
	fnName := "NewFileFlags()"
	if err := internal.CreateDefaultYamlFileForTesting(); err != nil {
		t.Errorf("error creating defaults.yaml: %v", err)
	}
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, "./mp3")
	}()
	type args struct {
		v *viper.Viper
	}
	tests := []struct {
		name            string
		args            args
		wantTopDir      string
		wantExtension   string
		wantAlbumRegex  string
		wantArtistRegex string
	}{
		{
			name:            "default",
			args:            args{},
			wantTopDir:      "./Music",
			wantExtension:   ".mp3",
			wantAlbumRegex:  ".*",
			wantArtistRegex: ".*",
		},
		{
			name:            "overrides",
			args:            args{v: internal.ReadDefaultsYaml("./mp3")},
			wantTopDir:      ".",
			wantExtension:   ".mpeg",
			wantAlbumRegex:  "^.*$",
			wantArtistRegex: "^.*$",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewSearchFlags(tt.args.v, flag.NewFlagSet("test", flag.ContinueOnError)); got == nil {
				t.Errorf("%s = %v", fnName, got)
			} else {
				if err := got.f.Parse([]string{}); err != nil {
					t.Errorf("%s error parsing flags: %v", fnName, err)
				} else {
					if *got.topDirectory != tt.wantTopDir {
						t.Errorf("%s %s got top directory %q want %q", fnName, tt.name, *got.topDirectory, tt.wantTopDir)
					}
					if *got.fileExtension != tt.wantExtension {
						t.Errorf("%s %s got extension %q want %q", fnName, tt.name, *got.fileExtension, tt.wantExtension)
					}
					if *got.albumRegex != tt.wantAlbumRegex {
						t.Errorf("%s %s got album regex %q want %q", fnName, tt.name, *got.albumRegex, tt.wantAlbumRegex)
					}
					if *got.artistRegex != tt.wantArtistRegex {
						t.Errorf("%s %s got artist regex %q want %q", fnName, tt.name, *got.artistRegex, tt.wantArtistRegex)
					}
				}
			}
		})
	}
}

func Test_validateRegexp(t *testing.T) {
	fnName := "validateRegexp()"
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
				t.Errorf("%s gotFilter = %v, want %v", fnName, gotFilter, tt.wantFilter)
			}
			if gotBadRegex != tt.wantBadRegex {
				t.Errorf("%s gotBadRegex = %v, want %v", fnName, gotBadRegex, tt.wantBadRegex)
			}
		})
	}
}

func Test_validateSearchParameters(t *testing.T) {
	fnName := "validateSearchParameters()"
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
			sf := &SearchFlags{
				topDirectory:  &tt.args.dir,
				fileExtension: &tt.args.ext,
				albumRegex:    &tt.args.albums,
				artistRegex:   &tt.args.artists,
			}
			gotAlbumsFilter, gotArtistsFilter, gotProblemsExist := sf.validate()
			if !tt.wantProblemsExist {
				if !reflect.DeepEqual(gotAlbumsFilter, tt.wantAlbumsFilter) {
					t.Errorf("%s gotAlbumsFilter = %v, want %v", fnName, gotAlbumsFilter, tt.wantAlbumsFilter)
				}
				if !reflect.DeepEqual(gotArtistsFilter, tt.wantArtistsFilter) {
					t.Errorf("%s gotArtistsFilter = %v, want %v", fnName, gotArtistsFilter, tt.wantArtistsFilter)
				}
			}
			if gotProblemsExist != tt.wantProblemsExist {
				t.Errorf("%s gotProblemsExist = %v, want %v", fnName, gotProblemsExist, tt.wantProblemsExist)
			}
		})
	}
}

func TestSearchFlags_validateTopLevelDirectory(t *testing.T) {
	fnName := "SearchFlags.validateTopLevelDirectory()"
	thisDir := "."
	notAFile := "no such file"
	notADir := "searchflags_test.go"
	tests := []struct {
		name string
		sf   *SearchFlags
		want bool
	}{
		{name: "is directory", sf: &SearchFlags{topDirectory: &thisDir}, want: true},
		{name: "non-existent directory", sf: &SearchFlags{topDirectory: &notAFile}, want: false},
		{name: "file that is not a directory", sf: &SearchFlags{topDirectory: &notADir}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.sf.validateTopLevelDirectory(); got != tt.want {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func TestSearchFlags_validateExtension(t *testing.T) {
	originalRegex := trackNameRegex
	defer func() {
		trackNameRegex = originalRegex
	}()
	fnName := "SearchFlags.validateExtension()"
	defaultExtension := defaultFileExtension
	missingLeadDot := "mp3"
	multipleDots := ".m.p3"
	badChar := ".m[p3"
	tests := []struct {
		name      string
		sf        *SearchFlags
		wantValid bool
	}{
		{name: "valid extension", sf: &SearchFlags{fileExtension: &defaultExtension}, wantValid: true},
		{name: "extension does not start with '.'", sf: &SearchFlags{fileExtension: &missingLeadDot}, wantValid: false},
		{name: "extension contains multiple '.'", sf: &SearchFlags{fileExtension: &multipleDots}, wantValid: false},
		{name: "extension contains invalid characters", sf: &SearchFlags{fileExtension: &badChar}, wantValid: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotValid := tt.sf.validateExtension(); gotValid != tt.wantValid {
				t.Errorf("%s = %v, want %v", fnName, gotValid, tt.wantValid)
			}
		})
	}
}
