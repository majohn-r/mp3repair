package files

import (
	"bytes"
	"flag"
	"mp3/internal"
	"os"
	"reflect"
	"regexp"
	"testing"
)

func Test_NewFileFlags(t *testing.T) {
	fnName := "NewFileFlags()"
	savedState := internal.SaveEnvVarForTesting("APPDATA")
	os.Setenv("APPDATA", internal.SecureAbsolutePathForTesting("."))
	defer func() {
		savedState.RestoreForTesting()
	}()
	oldHomePath := os.Getenv("HOMEPATH")
	defer func() {
		os.Setenv("HOMEPATH", oldHomePath)
	}()
	os.Setenv("HOMEPATH", ".")
	if err := internal.CreateDefaultYamlFileForTesting(); err != nil {
		t.Errorf("%s error creating defaults.yaml: %v", fnName, err)
	}
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, "./mp3")
	}()
	defaultConfig, _ := internal.ReadConfigurationFile(internal.NewOutputDeviceForTesting())
	type args struct {
		c *internal.Configuration
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
			args:            args{c: internal.EmptyConfiguration()},
			wantTopDir:      "./Music",
			wantExtension:   ".mp3",
			wantAlbumRegex:  ".*",
			wantArtistRegex: ".*",
		},
		{
			name:            "overrides",
			args:            args{c: defaultConfig},
			wantTopDir:      ".",
			wantExtension:   ".mpeg",
			wantAlbumRegex:  "^.*$",
			wantArtistRegex: "^.*$",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewSearchFlags(tt.args.c, flag.NewFlagSet("test", flag.ContinueOnError)); got == nil {
				t.Errorf("%s = %v", fnName, got)
			} else {
				if err := got.f.Parse([]string{}); err != nil {
					t.Errorf("%s error parsing flags: %v", fnName, err)
				} else {
					if *got.topDirectory != tt.wantTopDir {
						t.Errorf("%s %q got top directory %q want %q", fnName, tt.name, *got.topDirectory, tt.wantTopDir)
					}
					if *got.fileExtension != tt.wantExtension {
						t.Errorf("%s %q got extension %q want %q", fnName, tt.name, *got.fileExtension, tt.wantExtension)
					}
					if *got.albumRegex != tt.wantAlbumRegex {
						t.Errorf("%s %q got album regex %q want %q", fnName, tt.name, *got.albumRegex, tt.wantAlbumRegex)
					}
					if *got.artistRegex != tt.wantArtistRegex {
						t.Errorf("%s %q got artist regex %q want %q", fnName, tt.name, *got.artistRegex, tt.wantArtistRegex)
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
		name       string
		args       args
		wantFilter *regexp.Regexp
		wantOk     bool
		internal.WantedOutput
	}{
		{
			name: "valid filter with regex",
			args: args{
				pattern: "^.*$",
				name:    "artist",
			},
			wantFilter: regexp.MustCompile("^.*$"),
			wantOk:     true,
		},
		{
			name: "valid simple filter",
			args: args{
				pattern: "Beatles",
				name:    "artist",
			},
			wantFilter: regexp.MustCompile("Beatles"),
			wantOk:     true,
		},
		{
			name: "invalid filter",
			args: args{pattern: "disc[", name: "album"},
			WantedOutput: internal.WantedOutput{
				WantErrorOutput: "The album filter value you specified, \"disc[\", cannot be used: error parsing regexp: missing closing ]: `[`.\n",
				WantLogOutput:   "level='warn' album='disc[' error='error parsing regexp: missing closing ]: `[`' msg='the filter cannot be parsed as a regular expression'\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := internal.NewOutputDeviceForTesting()
			gotFilter, gotOk := validateRegexp(o, tt.args.pattern, tt.args.name)
			if tt.wantOk && !reflect.DeepEqual(gotFilter, tt.wantFilter) {
				t.Errorf("%s gotFilter = %v, want %v", fnName, gotFilter, tt.wantFilter)
			}
			if gotOk != tt.wantOk {
				t.Errorf("%s gotOk = %v, want %v", fnName, gotOk, tt.wantOk)
			}
			if issues, ok := o.CheckOutput(tt.WantedOutput); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
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
		wantOk            bool
		internal.WantedOutput
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
			wantOk:            true,
		},
		{
			name: "bad extension 1",
			args: args{
				dir:     ".",
				ext:     "mp3",
				albums:  ".*",
				artists: ".*",
			},
			wantAlbumsFilter:  regexp.MustCompile(".*"),
			wantArtistsFilter: regexp.MustCompile(".*"),
			WantedOutput: internal.WantedOutput{
				WantErrorOutput: "The -ext value you specified, \"mp3\", must contain exactly one '.' and '.' must be the first character.\n",
				WantLogOutput:   "level='warn' -ext='mp3' msg='the file extension must begin with '.' and contain no other '.' characters'\n",
			},
		},
		{
			name: "bad extension 2",
			args: args{
				dir:     ".",
				ext:     ".m.p3",
				albums:  ".*",
				artists: ".*",
			},
			wantAlbumsFilter:  regexp.MustCompile(".*"),
			wantArtistsFilter: regexp.MustCompile(".*"),
			WantedOutput: internal.WantedOutput{
				WantErrorOutput: "The -ext value you specified, \".m.p3\", must contain exactly one '.' and '.' must be the first character.\n",
				WantLogOutput:   "level='warn' -ext='.m.p3' msg='the file extension must begin with '.' and contain no other '.' characters'\n",
			},
		},
		{
			name: "bad extension 3",
			args: args{
				dir:     ".",
				ext:     ".mp[3",
				albums:  ".*",
				artists: ".*",
			},
			wantAlbumsFilter:  regexp.MustCompile(".*"),
			wantArtistsFilter: regexp.MustCompile(".*"),
			WantedOutput: internal.WantedOutput{
				WantErrorOutput: "The -ext value you specified, \".mp[3\", cannot be used for file matching: error parsing regexp: missing closing ]: `[3$`.\n",
				WantLogOutput:   "level='warn' -ext='.mp[3' error='error parsing regexp: missing closing ]: `[3$`' msg='the file extension cannot be parsed as a regular expression'\n",
			},
		},
		{
			name: "bad album filter 1",
			args: args{
				dir:     ".",
				ext:     ".mp3",
				albums:  ".[*",
				artists: ".*",
			},
			wantArtistsFilter: regexp.MustCompile(".*"),
			WantedOutput: internal.WantedOutput{
				WantErrorOutput: "The -albumFilter filter value you specified, \".[*\", cannot be used: error parsing regexp: missing closing ]: `[*`.\n",
				WantLogOutput:   "level='warn' -albumFilter='.[*' error='error parsing regexp: missing closing ]: `[*`' msg='the filter cannot be parsed as a regular expression'\n",
			},
		},
		{
			name: "bad album filter 2",
			args: args{
				dir:     ".",
				ext:     ".mp3",
				albums:  ".*",
				artists: ".[*",
			},
			wantAlbumsFilter: regexp.MustCompile(".*"),
			WantedOutput: internal.WantedOutput{
				WantErrorOutput: "The -artistFilter filter value you specified, \".[*\", cannot be used: error parsing regexp: missing closing ]: `[*`.\n",
				WantLogOutput:   "level='warn' -artistFilter='.[*' error='error parsing regexp: missing closing ]: `[*`' msg='the filter cannot be parsed as a regular expression'\n",
			},
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
			WantedOutput: internal.WantedOutput{
				WantErrorOutput: "The -topDir value you specified, \"no such directory\", cannot be read: CreateFile no such directory: The system cannot find the file specified..\n",
				WantLogOutput:   "level='warn' -topDir='no such directory' error='CreateFile no such directory: The system cannot find the file specified.' msg='cannot read directory'\n",
			},
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
			WantedOutput: internal.WantedOutput{
				WantErrorOutput: "The -topDir value you specified, \"utilities_test.go\", is not a directory.\n",
				WantLogOutput:   "level='warn' -topDir='utilities_test.go' msg='the file is not a directory'\n",
			},
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
			o := internal.NewOutputDeviceForTesting()
			gotAlbumsFilter, gotArtistsFilter, gotOk := sf.validate(o)
			if !tt.wantOk {
				if !reflect.DeepEqual(gotAlbumsFilter, tt.wantAlbumsFilter) {
					t.Errorf("%s gotAlbumsFilter = %v, want %v", fnName, gotAlbumsFilter, tt.wantAlbumsFilter)
				}
				if !reflect.DeepEqual(gotArtistsFilter, tt.wantArtistsFilter) {
					t.Errorf("%s gotArtistsFilter = %v, want %v", fnName, gotArtistsFilter, tt.wantArtistsFilter)
				}
			}
			if gotOk != tt.wantOk {
				t.Errorf("%s gotOk = %v, want %v", fnName, gotOk, tt.wantOk)
			}
			if issues, ok := o.CheckOutput(tt.WantedOutput); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
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
		internal.WantedOutput
	}{
		{name: "is directory", sf: &SearchFlags{topDirectory: &thisDir}, want: true},
		{
			name: "non-existent directory",
			sf:   &SearchFlags{topDirectory: &notAFile},
			WantedOutput: internal.WantedOutput{
				WantErrorOutput: "The -topDir value you specified, \"no such file\", cannot be read: CreateFile no such file: The system cannot find the file specified..\n",
				WantLogOutput:   "level='warn' -topDir='no such file' error='CreateFile no such file: The system cannot find the file specified.' msg='cannot read directory'\n",
			},
		},
		{
			name: "file that is not a directory",
			sf:   &SearchFlags{topDirectory: &notADir},
			WantedOutput: internal.WantedOutput{
				WantErrorOutput: "The -topDir value you specified, \"searchflags_test.go\", is not a directory.\n",
				WantLogOutput:   "level='warn' -topDir='searchflags_test.go' msg='the file is not a directory'\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := internal.NewOutputDeviceForTesting()
			if got := tt.sf.validateTopLevelDirectory(o); got != tt.want {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
			if issues, ok := o.CheckOutput(tt.WantedOutput); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
		})
	}
}

func TestSearchFlags_validateExtension(t *testing.T) {
	fnName := "SearchFlags.validateExtension()"
	originalRegex := trackNameRegex
	defer func() {
		trackNameRegex = originalRegex
	}()
	defaultExtension := defaultFileExtension
	missingLeadDot := "mp3"
	multipleDots := ".m.p3"
	badChar := ".m[p3"
	tests := []struct {
		name      string
		sf        *SearchFlags
		wantValid bool
		internal.WantedOutput
	}{
		{name: "valid extension", sf: &SearchFlags{fileExtension: &defaultExtension}, wantValid: true},
		{
			name: "extension does not start with '.'",
			sf:   &SearchFlags{fileExtension: &missingLeadDot},
			WantedOutput: internal.WantedOutput{
				WantErrorOutput: "The -ext value you specified, \"mp3\", must contain exactly one '.' and '.' must be the first character.\n",
				WantLogOutput:   "level='warn' -ext='mp3' msg='the file extension must begin with '.' and contain no other '.' characters'\n",
			},
		},
		{
			name: "extension contains multiple '.'",
			sf:   &SearchFlags{fileExtension: &multipleDots},
			WantedOutput: internal.WantedOutput{
				WantErrorOutput: "The -ext value you specified, \".m.p3\", must contain exactly one '.' and '.' must be the first character.\n",
				WantLogOutput:   "level='warn' -ext='.m.p3' msg='the file extension must begin with '.' and contain no other '.' characters'\n",
			},
		},
		{
			name: "extension contains invalid characters",
			sf:   &SearchFlags{fileExtension: &badChar},
			WantedOutput: internal.WantedOutput{
				WantErrorOutput: "The -ext value you specified, \".m[p3\", cannot be used for file matching: error parsing regexp: missing closing ]: `[p3$`.\n",
				WantLogOutput:   "level='warn' -ext='.m[p3' error='error parsing regexp: missing closing ]: `[p3$`' msg='the file extension cannot be parsed as a regular expression'\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := internal.NewOutputDeviceForTesting()
			if gotValid := tt.sf.validateExtension(o); gotValid != tt.wantValid {
				t.Errorf("%s = %v, want %v", fnName, gotValid, tt.wantValid)
			}
			if issues, ok := o.CheckOutput(tt.WantedOutput); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
		})
	}
}

func TestSearchFlags_ProcessArgs(t *testing.T) {
	fnName := "SearchFlags.ProcessArgs()"
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
		name   string
		sf     *SearchFlags
		args   args
		wantS  *Search
		wantOk bool
		internal.WantedOutput
	}{
		{name: "good arguments", sf: flags, args: args{args: nil}, wantS: s, wantOk: true},
		{
			name: "request help",
			sf:   flags,
			args: args{args: []string{"-help"}},
			WantedOutput: internal.WantedOutput{
				WantErrorOutput: usage,
				WantLogOutput:   "level='error' arguments='[-help]' msg='flag: help requested'\n",
			},
		},
		{
			name: "request invalid argument",
			sf:   flags,
			args: args{args: []string{"-foo"}},
			WantedOutput: internal.WantedOutput{
				WantErrorOutput: "flag provided but not defined: -foo\n" + usage,
				WantLogOutput:   "level='error' arguments='[-foo]' msg='flag provided but not defined: -foo'\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := internal.NewOutputDeviceForTesting()
			gotS, gotOk := tt.sf.ProcessArgs(o, tt.args.args)
			if !reflect.DeepEqual(gotS, tt.wantS) {
				t.Errorf("%s gotS = %v, want %v", fnName, gotS, tt.wantS)
			}
			if gotOk != tt.wantOk {
				t.Errorf("%s gotOk = %v, want %v", fnName, gotOk, tt.wantOk)
			}
			if issues, ok := o.CheckOutput(tt.WantedOutput); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
		})
	}
}
