package files

import (
	"bytes"
	"flag"
	"mp3/internal"
	"mp3/internal/output"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"testing"
)

func Test_NewSearchFlags(t *testing.T) {
	fnName := "NewSearchFlags()"
	savedAppData := internal.SaveEnvVarForTesting("APPDATA")
	os.Setenv("APPDATA", internal.SecureAbsolutePathForTesting("."))
	savedFoo := internal.SaveEnvVarForTesting("FOO")
	os.Unsetenv("FOO")
	defer func() {
		savedAppData.RestoreForTesting()
		savedFoo.RestoreForTesting()
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
	defaultConfig, _ := internal.ReadConfigurationFile(output.NewNilBus())
	type args struct {
		c *internal.Configuration
	}
	tests := []struct {
		name string
		args
		wantOk          bool
		wantTopDir      string
		wantExtension   string
		wantAlbumRegex  string
		wantArtistRegex string
		output.WantedRecording
	}{
		{
			name:            "default",
			args:            args{c: internal.EmptyConfiguration()},
			wantTopDir:      ".\\Music",
			wantExtension:   ".mp3",
			wantAlbumRegex:  ".*",
			wantArtistRegex: ".*",
			wantOk:          true,
		},
		{
			name:            "overrides",
			args:            args{c: defaultConfig},
			wantTopDir:      ".",
			wantExtension:   ".mpeg",
			wantAlbumRegex:  "^.*$",
			wantArtistRegex: "^.*$",
			wantOk:          true,
		},
		{
			name: "bad default topDir",
			args: args{
				c: internal.CreateConfiguration(output.NewNilBus(), map[string]any{
					"common": map[string]any{
						"topDir": "$FOO",
					},
				}),
			},
			WantedRecording: output.WantedRecording{
				Error: "The configuration file \"defaults.yaml\" contains an invalid value for \"common\": invalid value \"$FOO\" for flag -topDir: missing environment variables: [FOO].\n",
				Log:   "level='error' error='invalid value \"$FOO\" for flag -topDir: missing environment variables: [FOO]' section='common' msg='invalid content in configuration file'\n",
			},
		},
		{
			name: "bad default extension",
			args: args{
				c: internal.CreateConfiguration(output.NewNilBus(), map[string]any{
					"common": map[string]any{
						"ext": "$FOO",
					},
				}),
			},
			WantedRecording: output.WantedRecording{
				Error: "The configuration file \"defaults.yaml\" contains an invalid value for \"common\": invalid value \"$FOO\" for flag -ext: missing environment variables: [FOO].\n",
				Log:   "level='error' error='invalid value \"$FOO\" for flag -ext: missing environment variables: [FOO]' section='common' msg='invalid content in configuration file'\n",
			},
		},
		{
			name: "bad default album filter",
			args: args{
				c: internal.CreateConfiguration(output.NewNilBus(), map[string]any{
					"common": map[string]any{
						"albumFilter": "$FOO",
					},
				}),
			},
			WantedRecording: output.WantedRecording{
				Error: "The configuration file \"defaults.yaml\" contains an invalid value for \"common\": invalid value \"$FOO\" for flag -albumFilter: missing environment variables: [FOO].\n",
				Log:   "level='error' error='invalid value \"$FOO\" for flag -albumFilter: missing environment variables: [FOO]' section='common' msg='invalid content in configuration file'\n",
			},
		},
		{
			name: "bad default artist filter",
			args: args{
				c: internal.CreateConfiguration(output.NewNilBus(), map[string]any{
					"common": map[string]any{
						"artistFilter": "$FOO",
					},
				}),
			},
			WantedRecording: output.WantedRecording{
				Error: "The configuration file \"defaults.yaml\" contains an invalid value for \"common\": invalid value \"$FOO\" for flag -artistFilter: missing environment variables: [FOO].\n",
				Log:   "level='error' error='invalid value \"$FOO\" for flag -artistFilter: missing environment variables: [FOO]' section='common' msg='invalid content in configuration file'\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := output.NewRecorder()
			got, gotOk := NewSearchFlags(o, tt.args.c, flag.NewFlagSet("test", flag.ContinueOnError))
			if gotOk != tt.wantOk {
				t.Errorf("%s gotOk %t wantOK %t", fnName, gotOk, tt.wantOk)
			}
			if got != nil {
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
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
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
		name string
		args
		wantFilter *regexp.Regexp
		wantOk     bool
		output.WantedRecording
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
			WantedRecording: output.WantedRecording{
				Error: "The album filter value you specified, \"disc[\", cannot be used: error parsing regexp: missing closing ]: `[`.\n",
				Log:   "level='error' album='disc[' error='error parsing regexp: missing closing ]: `[`' msg='the filter cannot be parsed as a regular expression'\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := output.NewRecorder()
			gotFilter, gotOk := validateRegexp(o, tt.args.pattern, tt.args.name)
			if tt.wantOk && !reflect.DeepEqual(gotFilter, tt.wantFilter) {
				t.Errorf("%s gotFilter = %v, want %v", fnName, gotFilter, tt.wantFilter)
			}
			if gotOk != tt.wantOk {
				t.Errorf("%s gotOk = %v, want %v", fnName, gotOk, tt.wantOk)
			}
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
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
		name string
		args
		wantAlbumsFilter  *regexp.Regexp
		wantArtistsFilter *regexp.Regexp
		wantOk            bool
		output.WantedRecording
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
			WantedRecording: output.WantedRecording{
				Error: "The -ext value you specified, \"mp3\", must contain exactly one '.' and '.' must be the first character.\n",
				Log:   "level='error' -ext='mp3' msg='the file extension must begin with '.' and contain no other '.' characters'\n",
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
			WantedRecording: output.WantedRecording{
				Error: "The -ext value you specified, \".m.p3\", must contain exactly one '.' and '.' must be the first character.\n",
				Log:   "level='error' -ext='.m.p3' msg='the file extension must begin with '.' and contain no other '.' characters'\n",
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
			WantedRecording: output.WantedRecording{
				Error: "The -ext value you specified, \".mp[3\", cannot be used for file matching: error parsing regexp: missing closing ]: `[3$`.\n",
				Log:   "level='error' -ext='.mp[3' error='error parsing regexp: missing closing ]: `[3$`' msg='the file extension cannot be parsed as a regular expression'\n",
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
			WantedRecording: output.WantedRecording{
				Error: "The -albumFilter filter value you specified, \".[*\", cannot be used: error parsing regexp: missing closing ]: `[*`.\n",
				Log:   "level='error' -albumFilter='.[*' error='error parsing regexp: missing closing ]: `[*`' msg='the filter cannot be parsed as a regular expression'\n",
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
			WantedRecording: output.WantedRecording{
				Error: "The -artistFilter filter value you specified, \".[*\", cannot be used: error parsing regexp: missing closing ]: `[*`.\n",
				Log:   "level='error' -artistFilter='.[*' error='error parsing regexp: missing closing ]: `[*`' msg='the filter cannot be parsed as a regular expression'\n",
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
			WantedRecording: output.WantedRecording{
				Error: "The -topDir value you specified, \"no such directory\", cannot be read: CreateFile no such directory: The system cannot find the file specified.\n",
				Log:   "level='error' -topDir='no such directory' error='CreateFile no such directory: The system cannot find the file specified.' msg='cannot read directory'\n",
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
			WantedRecording: output.WantedRecording{
				Error: "The -topDir value you specified, \"utilities_test.go\", is not a directory.\n",
				Log:   "level='error' -topDir='utilities_test.go' msg='the file is not a directory'\n",
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
			o := output.NewRecorder()
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
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
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
		output.WantedRecording
	}{
		{name: "is directory", sf: &SearchFlags{topDirectory: &thisDir}, want: true},
		{
			name: "non-existent directory",
			sf:   &SearchFlags{topDirectory: &notAFile},
			WantedRecording: output.WantedRecording{
				Error: "The -topDir value you specified, \"no such file\", cannot be read: CreateFile no such file: The system cannot find the file specified.\n",
				Log:   "level='error' -topDir='no such file' error='CreateFile no such file: The system cannot find the file specified.' msg='cannot read directory'\n",
			},
		},
		{
			name: "file that is not a directory",
			sf:   &SearchFlags{topDirectory: &notADir},
			WantedRecording: output.WantedRecording{
				Error: "The -topDir value you specified, \"searchflags_test.go\", is not a directory.\n",
				Log:   "level='error' -topDir='searchflags_test.go' msg='the file is not a directory'\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := output.NewRecorder()
			if got := tt.sf.validateTopLevelDirectory(o); got != tt.want {
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
		output.WantedRecording
	}{
		{name: "valid extension", sf: &SearchFlags{fileExtension: &defaultExtension}, wantValid: true},
		{
			name: "extension does not start with '.'",
			sf:   &SearchFlags{fileExtension: &missingLeadDot},
			WantedRecording: output.WantedRecording{
				Error: "The -ext value you specified, \"mp3\", must contain exactly one '.' and '.' must be the first character.\n",
				Log:   "level='error' -ext='mp3' msg='the file extension must begin with '.' and contain no other '.' characters'\n",
			},
		},
		{
			name: "extension contains multiple '.'",
			sf:   &SearchFlags{fileExtension: &multipleDots},
			WantedRecording: output.WantedRecording{
				Error: "The -ext value you specified, \".m.p3\", must contain exactly one '.' and '.' must be the first character.\n",
				Log:   "level='error' -ext='.m.p3' msg='the file extension must begin with '.' and contain no other '.' characters'\n",
			},
		},
		{
			name: "extension contains invalid characters",
			sf:   &SearchFlags{fileExtension: &badChar},
			WantedRecording: output.WantedRecording{
				Error: "The -ext value you specified, \".m[p3\", cannot be used for file matching: error parsing regexp: missing closing ]: `[p3$`.\n",
				Log:   "level='error' -ext='.m[p3' error='error parsing regexp: missing closing ]: `[p3$`' msg='the file extension cannot be parsed as a regular expression'\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := output.NewRecorder()
			if gotValid := tt.sf.validateExtension(o); gotValid != tt.wantValid {
				t.Errorf("%s = %v, want %v", fnName, gotValid, tt.wantValid)
			}
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
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
		name string
		sf   *SearchFlags
		args
		wantS  *Search
		wantOk bool
		output.WantedRecording
	}{
		{name: "good arguments", sf: flags, args: args{args: nil}, wantS: s, wantOk: true},
		{
			name: "request help",
			sf:   flags,
			args: args{args: []string{"-help"}},
			WantedRecording: output.WantedRecording{
				Error: usage,
				Log:   "level='error' arguments='[-help]' msg='flag: help requested'\n",
			},
		},
		{
			name: "request invalid argument",
			sf:   flags,
			args: args{args: []string{"-foo"}},
			WantedRecording: output.WantedRecording{
				Error: "flag provided but not defined: -foo\n" + usage,
				Log:   "level='error' arguments='[-foo]' msg='flag provided but not defined: -foo'\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := output.NewRecorder()
			gotS, gotOk := tt.sf.ProcessArgs(o, tt.args.args)
			if !reflect.DeepEqual(gotS, tt.wantS) {
				t.Errorf("%s gotS = %v, want %v", fnName, gotS, tt.wantS)
			}
			if gotOk != tt.wantOk {
				t.Errorf("%s gotOk = %v, want %v", fnName, gotOk, tt.wantOk)
			}
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
		})
	}
}

func TestSearchDefaults(t *testing.T) {
	fnName := "SearchDefaults()"
	tests := []struct {
		name  string
		want  string
		want1 map[string]any
	}{
		{
			name: "single use case",
			want: "common",
			want1: map[string]any{
				"albumFilter":  ".*",
				"artistFilter": ".*",
				"ext":          ".mp3",
				"topDir":       filepath.Join("%HOMEPATH%", "Music"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := SearchDefaults()
			if got != tt.want {
				t.Errorf("%s got = %v, want %v", fnName, got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("%s got1 = %v, want %v", fnName, got1, tt.want1)
			}
		})
	}
}
