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
	fnName := "NewFileFlags()"
	if err := internal.CreateDefaultYamlFileForTesting(); err != nil {
		t.Errorf("error creating defaults.yaml: %v", err)
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
		name              string
		args              args
		wantFilter        *regexp.Regexp
		wantOk            bool
		wantConsoleOutput string
		wantErrorOutput   string
		wantLogOutput     string
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
			name:            "invalid filter",
			args:            args{pattern: "disc[", name: "album"},
			wantErrorOutput: "The album filter value you specified, \"disc[\", cannot be used: error parsing regexp: missing closing ]: `[`\n",
			wantLogOutput:   "level='warn' album='disc[' error='error parsing regexp: missing closing ]: `[`' msg='the filter cannot be parsed as a regular expression'\n",
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
			if gotConsoleOutput := o.ConsoleOutput(); gotConsoleOutput != tt.wantConsoleOutput {
				t.Errorf("%s console output = %q, want %q", fnName, gotConsoleOutput, tt.wantConsoleOutput)
			}
			if gotErrorOutput := o.ErrorOutput(); gotErrorOutput != tt.wantErrorOutput {
				t.Errorf("%s error output = %q, want %q", fnName, gotErrorOutput, tt.wantErrorOutput)
			}
			if gotLogOutput := o.LogOutput(); gotLogOutput != tt.wantLogOutput {
				t.Errorf("%s log output = %q, want %q", fnName, gotLogOutput, tt.wantLogOutput)
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
		wantConsoleOutput string
		wantErrorOutput   string
		wantLogOutput     string
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
			wantErrorOutput:   "The -ext value you specified, \"mp3\", must contain exactly one '.' and '.' must be the first character.\n",
			wantLogOutput:     "level='warn' -ext='mp3' msg='the file extension must begin with '.' and contain no other '.' characters'\n",
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
			wantErrorOutput:   "The -ext value you specified, \".m.p3\", must contain exactly one '.' and '.' must be the first character.\n",
			wantLogOutput:     "level='warn' -ext='.m.p3' msg='the file extension must begin with '.' and contain no other '.' characters'\n",
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
			wantErrorOutput:   "The -ext value you specified, \".mp[3\", cannot be used for file matching: error parsing regexp: missing closing ]: `[3$`.\n",
			wantLogOutput:     "level='warn' -ext='.mp[3' error='error parsing regexp: missing closing ]: `[3$`' msg='the file extension cannot be parsed as a regular expression'\n",
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
			wantErrorOutput:   "The -albumFilter filter value you specified, \".[*\", cannot be used: error parsing regexp: missing closing ]: `[*`\n",
			wantLogOutput:     "level='warn' -albumFilter='.[*' error='error parsing regexp: missing closing ]: `[*`' msg='the filter cannot be parsed as a regular expression'\n",
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
			wantErrorOutput:  "The -artistFilter filter value you specified, \".[*\", cannot be used: error parsing regexp: missing closing ]: `[*`\n",
			wantLogOutput:    "level='warn' -artistFilter='.[*' error='error parsing regexp: missing closing ]: `[*`' msg='the filter cannot be parsed as a regular expression'\n",
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
			wantErrorOutput:   "The -topDir value you specified, \"no such directory\", cannot be read: CreateFile no such directory: The system cannot find the file specified..\n",
			wantLogOutput:     "level='warn' -topDir='no such directory' error='CreateFile no such directory: The system cannot find the file specified.' msg='cannot read directory'\n",
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
			wantErrorOutput:   "The -topDir value you specified, \"utilities_test.go\", is not a directory.\n",
			wantLogOutput:     "level='warn' -topDir='utilities_test.go' msg='the file is not a directory'\n",
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
			if gotConsoleOutput := o.ConsoleOutput(); gotConsoleOutput != tt.wantConsoleOutput {
				t.Errorf("%s console output = %q, want %q", fnName, gotConsoleOutput, tt.wantConsoleOutput)
			}
			if gotErrorOutput := o.ErrorOutput(); gotErrorOutput != tt.wantErrorOutput {
				t.Errorf("%s error output = %q, want %q", fnName, gotErrorOutput, tt.wantErrorOutput)
			}
			if gotLogOutput := o.LogOutput(); gotLogOutput != tt.wantLogOutput {
				t.Errorf("%s log output = %q, want %q", fnName, gotLogOutput, tt.wantLogOutput)
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
		name              string
		sf                *SearchFlags
		want              bool
		wantConsoleOutput string
		wantErrorOutput   string
		wantLogOutput     string
	}{
		{name: "is directory", sf: &SearchFlags{topDirectory: &thisDir}, want: true},
		{
			name:            "non-existent directory",
			sf:              &SearchFlags{topDirectory: &notAFile},
			wantErrorOutput: "The -topDir value you specified, \"no such file\", cannot be read: CreateFile no such file: The system cannot find the file specified..\n",
			wantLogOutput:   "level='warn' -topDir='no such file' error='CreateFile no such file: The system cannot find the file specified.' msg='cannot read directory'\n",
		},
		{
			name:            "file that is not a directory",
			sf:              &SearchFlags{topDirectory: &notADir},
			wantErrorOutput: "The -topDir value you specified, \"searchflags_test.go\", is not a directory.\n",
			wantLogOutput:   "level='warn' -topDir='searchflags_test.go' msg='the file is not a directory'\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := internal.NewOutputDeviceForTesting()
			if got := tt.sf.validateTopLevelDirectory(o); got != tt.want {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
			if gotConsoleOutput := o.ConsoleOutput(); gotConsoleOutput != tt.wantConsoleOutput {
				t.Errorf("%s console output = %q, want %q", fnName, gotConsoleOutput, tt.wantConsoleOutput)
			}
			if gotErrorOutput := o.ErrorOutput(); gotErrorOutput != tt.wantErrorOutput {
				t.Errorf("%s error output = %q, want %q", fnName, gotErrorOutput, tt.wantErrorOutput)
			}
			if gotLogOutput := o.LogOutput(); gotLogOutput != tt.wantLogOutput {
				t.Errorf("%s log output = %q, want %q", fnName, gotLogOutput, tt.wantLogOutput)
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
		name              string
		sf                *SearchFlags
		wantValid         bool
		wantConsoleOutput string
		wantErrorOutput   string
		wantLogOutput     string
	}{
		{name: "valid extension", sf: &SearchFlags{fileExtension: &defaultExtension}, wantValid: true},
		{
			name:            "extension does not start with '.'",
			sf:              &SearchFlags{fileExtension: &missingLeadDot},
			wantErrorOutput: "The -ext value you specified, \"mp3\", must contain exactly one '.' and '.' must be the first character.\n",
			wantLogOutput:   "level='warn' -ext='mp3' msg='the file extension must begin with '.' and contain no other '.' characters'\n",
		},
		{
			name:            "extension contains multiple '.'",
			sf:              &SearchFlags{fileExtension: &multipleDots},
			wantErrorOutput: "The -ext value you specified, \".m.p3\", must contain exactly one '.' and '.' must be the first character.\n",
			wantLogOutput:   "level='warn' -ext='.m.p3' msg='the file extension must begin with '.' and contain no other '.' characters'\n",
		},
		{
			name:            "extension contains invalid characters",
			sf:              &SearchFlags{fileExtension: &badChar},
			wantErrorOutput: "The -ext value you specified, \".m[p3\", cannot be used for file matching: error parsing regexp: missing closing ]: `[p3$`.\n",
			wantLogOutput:   "level='warn' -ext='.m[p3' error='error parsing regexp: missing closing ]: `[p3$`' msg='the file extension cannot be parsed as a regular expression'\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := internal.NewOutputDeviceForTesting()
			if gotValid := tt.sf.validateExtension(o); gotValid != tt.wantValid {
				t.Errorf("%s = %v, want %v", fnName, gotValid, tt.wantValid)
			}
			if gotConsoleOutput := o.ConsoleOutput(); gotConsoleOutput != tt.wantConsoleOutput {
				t.Errorf("%s console output = %q, want %q", fnName, gotConsoleOutput, tt.wantConsoleOutput)
			}
			if gotErrorOutput := o.ErrorOutput(); gotErrorOutput != tt.wantErrorOutput {
				t.Errorf("%s error output = %q, want %q", fnName, gotErrorOutput, tt.wantErrorOutput)
			}
			if gotLogOutput := o.LogOutput(); gotLogOutput != tt.wantLogOutput {
				t.Errorf("%s log output = %q, want %q", fnName, gotLogOutput, tt.wantLogOutput)
			}
		})
	}
}

func TestSearchFlags_ProcessArgs(t *testing.T) {
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
		name              string
		sf                *SearchFlags
		args              args
		wantS             *Search
		wantOk            bool
		wantConsoleOutput string
		wantErrorOutput   string
		wantLogOutput     string
	}{
		{name: "good arguments", sf: flags, args: args{args: nil}, wantS: s, wantOk: true},
		{
			name:            "request help",
			sf:              flags,
			args:            args{args: []string{"-help"}},
			wantErrorOutput: usage,
			wantLogOutput:   "level='error' arguments='[-help]' msg='flag: help requested'\n",
		},
		{
			name:            "request invalid argument",
			sf:              flags,
			args:            args{args: []string{"-foo"}},
			wantErrorOutput: "flag provided but not defined: -foo\n" + usage,
			wantLogOutput:   "level='error' arguments='[-foo]' msg='flag provided but not defined: -foo'\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := internal.NewOutputDeviceForTesting()
			gotS, gotOk := tt.sf.ProcessArgs(o, tt.args.args)
			if !reflect.DeepEqual(gotS, tt.wantS) {
				t.Errorf("SearchFlags.ProcessArgs() gotS = %v, want %v", gotS, tt.wantS)
			}
			if gotOk != tt.wantOk {
				t.Errorf("SearchFlags.ProcessArgs() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
			if gotConsoleOutput := o.ConsoleOutput(); gotConsoleOutput != tt.wantConsoleOutput {
				t.Errorf("SearchFlags.ProcessArgs() gotConsoleOutput = %v, want %v", gotConsoleOutput, tt.wantConsoleOutput)
			}
			if gotErrorOutput := o.ErrorOutput(); gotErrorOutput != tt.wantErrorOutput {
				t.Errorf("SearchFlags.ProcessArgs() gotErrorOutput = %v, want %v", gotErrorOutput, tt.wantErrorOutput)
			}
			if gotLogOutput := o.LogOutput(); gotLogOutput != tt.wantLogOutput {
				t.Errorf("SearchFlags.ProcessArgs() gotLogOutput = %v, want %v", gotLogOutput, tt.wantLogOutput)
			}
		})
	}
}
