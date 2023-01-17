package files

import (
	"bytes"
	"flag"
	"mp3/internal"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"testing"

	"github.com/majohn-r/output"
)

func Test_NewSearchFlags(t *testing.T) {
	const fnName = "NewSearchFlags()"
	savedAppData := internal.SaveEnvVarForTesting("APPDATA")
	os.Setenv("APPDATA", internal.SecureAbsolutePathForTesting("."))
	oldAppPath := internal.ApplicationPath()
	o := output.NewNilBus()
	internal.InitApplicationPath(o)
	savedFoo := internal.SaveEnvVarForTesting("FOO")
	os.Unsetenv("FOO")
	savedHomePath := internal.SaveEnvVarForTesting("HOMEPATH")
	os.Setenv("HOMEPATH", ".")
	if err := internal.CreateDefaultYamlFileForTesting(); err != nil {
		t.Errorf("%s error creating defaults.yaml: %v", fnName, err)
	}
	defaultConfig, _ := internal.ReadConfigurationFile(o)
	defer func() {
		savedAppData.RestoreForTesting()
		internal.SetApplicationPathForTesting(oldAppPath)
		savedFoo.RestoreForTesting()
		internal.DestroyDirectoryForTesting(fnName, "./mp3")
		savedHomePath.RestoreForTesting()
	}()
	type args struct {
		c *internal.Configuration
	}
	tests := map[string]struct {
		args
		wantOk          bool
		wantTopDir      string
		wantExtension   string
		wantAlbumRegex  string
		wantArtistRegex string
		output.WantedRecording
	}{
		"default":   {args: args{c: internal.EmptyConfiguration()}, wantTopDir: ".\\Music", wantExtension: ".mp3", wantAlbumRegex: ".*", wantArtistRegex: ".*", wantOk: true},
		"overrides": {args: args{c: defaultConfig}, wantTopDir: ".", wantExtension: ".mpeg", wantAlbumRegex: "^.*$", wantArtistRegex: "^.*$", wantOk: true},
		"bad default topDir": {
			args: args{c: internal.NewConfiguration(output.NewNilBus(), map[string]any{"common": map[string]any{"topDir": "$FOO"}})},
			WantedRecording: output.WantedRecording{
				Error: "The configuration file \"defaults.yaml\" contains an invalid value for \"common\": invalid value \"$FOO\" for flag -topDir: missing environment variables: [FOO].\n",
				Log:   "level='error' error='invalid value \"$FOO\" for flag -topDir: missing environment variables: [FOO]' section='common' msg='invalid content in configuration file'\n",
			},
		},
		"bad default extension": {
			args: args{c: internal.NewConfiguration(output.NewNilBus(), map[string]any{"common": map[string]any{"ext": "$FOO"}})},
			WantedRecording: output.WantedRecording{
				Error: "The configuration file \"defaults.yaml\" contains an invalid value for \"common\": invalid value \"$FOO\" for flag -ext: missing environment variables: [FOO].\n",
				Log:   "level='error' error='invalid value \"$FOO\" for flag -ext: missing environment variables: [FOO]' section='common' msg='invalid content in configuration file'\n",
			},
		},
		"bad default album filter": {
			args: args{c: internal.NewConfiguration(output.NewNilBus(), map[string]any{"common": map[string]any{"albumFilter": "$FOO"}})},
			WantedRecording: output.WantedRecording{
				Error: "The configuration file \"defaults.yaml\" contains an invalid value for \"common\": invalid value \"$FOO\" for flag -albumFilter: missing environment variables: [FOO].\n",
				Log:   "level='error' error='invalid value \"$FOO\" for flag -albumFilter: missing environment variables: [FOO]' section='common' msg='invalid content in configuration file'\n",
			},
		},
		"bad default artist filter": {
			args: args{c: internal.NewConfiguration(output.NewNilBus(), map[string]any{"common": map[string]any{"artistFilter": "$FOO"}})},
			WantedRecording: output.WantedRecording{
				Error: "The configuration file \"defaults.yaml\" contains an invalid value for \"common\": invalid value \"$FOO\" for flag -artistFilter: missing environment variables: [FOO].\n",
				Log:   "level='error' error='invalid value \"$FOO\" for flag -artistFilter: missing environment variables: [FOO]' section='common' msg='invalid content in configuration file'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
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
						t.Errorf("%s %q got top directory %q want %q", fnName, name, *got.topDirectory, tt.wantTopDir)
					}
					if *got.fileExtension != tt.wantExtension {
						t.Errorf("%s %q got extension %q want %q", fnName, name, *got.fileExtension, tt.wantExtension)
					}
					if *got.albumRegex != tt.wantAlbumRegex {
						t.Errorf("%s %q got album regex %q want %q", fnName, name, *got.albumRegex, tt.wantAlbumRegex)
					}
					if *got.artistRegex != tt.wantArtistRegex {
						t.Errorf("%s %q got artist regex %q want %q", fnName, name, *got.artistRegex, tt.wantArtistRegex)
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
	const fnName = "validateRegexp()"
	type args struct {
		value string
		name  string
	}
	tests := map[string]struct {
		args
		wantFilter *regexp.Regexp
		wantOk     bool
		output.WantedRecording
	}{
		"valid filter with regex": {args: args{value: "^.*$", name: "artist"}, wantFilter: regexp.MustCompile("^.*$"), wantOk: true},
		"valid simple filter":     {args: args{value: "Beatles", name: "artist"}, wantFilter: regexp.MustCompile("Beatles"), wantOk: true},
		"invalid filter": {
			args: args{value: "disc[", name: "album"},
			WantedRecording: output.WantedRecording{
				Error: "The album filter value you specified, \"disc[\", cannot be used: error parsing regexp: missing closing ]: `[`.\n",
				Log:   "level='error' album='disc[' error='error parsing regexp: missing closing ]: `[`' msg='the filter cannot be parsed as a regular expression'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			gotFilter, gotOk := validateRegexp(o, tt.args.value, tt.args.name)
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
	const fnName = "validateSearchParameters()"
	type args struct {
		dir     string
		ext     string
		albums  string
		artists string
	}
	tests := map[string]struct {
		args
		wantAlbumsFilter  *regexp.Regexp
		wantArtistsFilter *regexp.Regexp
		wantOk            bool
		output.WantedRecording
	}{
		"valid input": {
			args:              args{dir: ".", ext: ".mp3", albums: ".*", artists: ".*"},
			wantAlbumsFilter:  regexp.MustCompile(".*"),
			wantArtistsFilter: regexp.MustCompile(".*"),
			wantOk:            true,
		},
		"bad extension 1": {
			args:              args{dir: ".", ext: "mp3", albums: ".*", artists: ".*"},
			wantAlbumsFilter:  regexp.MustCompile(".*"),
			wantArtistsFilter: regexp.MustCompile(".*"),
			WantedRecording: output.WantedRecording{
				Error: "The -ext value you specified, \"mp3\", must contain exactly one '.' and '.' must be the first character.\n",
				Log:   "level='error' -ext='mp3' msg='the file extension must begin with '.' and contain no other '.' characters'\n",
			},
		},
		"bad extension 2": {
			args:              args{dir: ".", ext: ".m.p3", albums: ".*", artists: ".*"},
			wantAlbumsFilter:  regexp.MustCompile(".*"),
			wantArtistsFilter: regexp.MustCompile(".*"),
			WantedRecording: output.WantedRecording{
				Error: "The -ext value you specified, \".m.p3\", must contain exactly one '.' and '.' must be the first character.\n",
				Log:   "level='error' -ext='.m.p3' msg='the file extension must begin with '.' and contain no other '.' characters'\n",
			},
		},
		"bad extension 3": {
			args:              args{dir: ".", ext: ".mp[3", albums: ".*", artists: ".*"},
			wantAlbumsFilter:  regexp.MustCompile(".*"),
			wantArtistsFilter: regexp.MustCompile(".*"),
			WantedRecording: output.WantedRecording{
				Error: "The -ext value you specified, \".mp[3\", cannot be used for file matching: error parsing regexp: missing closing ]: `[3$`.\n",
				Log:   "level='error' -ext='.mp[3' error='error parsing regexp: missing closing ]: `[3$`' msg='the file extension cannot be parsed as a regular expression'\n",
			},
		},
		"bad album filter 1": {
			args:              args{dir: ".", ext: ".mp3", albums: ".[*", artists: ".*"},
			wantArtistsFilter: regexp.MustCompile(".*"),
			WantedRecording: output.WantedRecording{
				Error: "The -albumFilter filter value you specified, \".[*\", cannot be used: error parsing regexp: missing closing ]: `[*`.\n",
				Log:   "level='error' -albumFilter='.[*' error='error parsing regexp: missing closing ]: `[*`' msg='the filter cannot be parsed as a regular expression'\n",
			},
		},
		"bad album filter 2": {
			args:             args{dir: ".", ext: ".mp3", albums: ".*", artists: ".[*"},
			wantAlbumsFilter: regexp.MustCompile(".*"),
			WantedRecording: output.WantedRecording{
				Error: "The -artistFilter filter value you specified, \".[*\", cannot be used: error parsing regexp: missing closing ]: `[*`.\n",
				Log:   "level='error' -artistFilter='.[*' error='error parsing regexp: missing closing ]: `[*`' msg='the filter cannot be parsed as a regular expression'\n",
			},
		},
		"non-existent directory": {
			args:              args{dir: "no such directory", ext: ".mp3", albums: ".*", artists: ".*"},
			wantAlbumsFilter:  regexp.MustCompile(".*"),
			wantArtistsFilter: regexp.MustCompile(".*"),
			WantedRecording: output.WantedRecording{
				Error: "The -topDir value you specified, \"no such directory\", cannot be read: CreateFile no such directory: The system cannot find the file specified.\n",
				Log:   "level='error' directory='no such directory' error='CreateFile no such directory: The system cannot find the file specified.' msg='cannot read directory'\n",
			},
		},
		"directory is not a directory": {
			args:              args{dir: "utilities_test.go", ext: ".mp3", albums: ".*", artists: ".*"},
			wantAlbumsFilter:  regexp.MustCompile(".*"),
			wantArtistsFilter: regexp.MustCompile(".*"),
			WantedRecording: output.WantedRecording{
				Error: "The -topDir value you specified, \"utilities_test.go\", is not a directory.\n",
				Log:   "level='error' -topDir='utilities_test.go' msg='the file is not a directory'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			sf := &SearchFlags{topDirectory: &tt.args.dir, fileExtension: &tt.args.ext, albumRegex: &tt.args.albums, artistRegex: &tt.args.artists}
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
	const fnName = "SearchFlags.validateTopLevelDirectory()"
	thisDir := "."
	notAFile := "no such file"
	notADir := "searchflags_test.go"
	tests := map[string]struct {
		sf   *SearchFlags
		want bool
		output.WantedRecording
	}{
		"is directory": {sf: &SearchFlags{topDirectory: &thisDir}, want: true},
		"non-existent directory": {
			sf: &SearchFlags{topDirectory: &notAFile},
			WantedRecording: output.WantedRecording{
				Error: "The -topDir value you specified, \"no such file\", cannot be read: CreateFile no such file: The system cannot find the file specified.\n",
				Log:   "level='error' directory='no such file' error='CreateFile no such file: The system cannot find the file specified.' msg='cannot read directory'\n",
			},
		},
		"file that is not a directory": {
			sf: &SearchFlags{topDirectory: &notADir},
			WantedRecording: output.WantedRecording{
				Error: "The -topDir value you specified, \"searchflags_test.go\", is not a directory.\n",
				Log:   "level='error' -topDir='searchflags_test.go' msg='the file is not a directory'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
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
	const fnName = "SearchFlags.validateExtension()"
	originalRegex := trackNameRegex
	defaultExtension := defaultFileExtension
	missingLeadDot := "mp3"
	multipleDots := ".m.p3"
	badChar := ".m[p3"
	defer func() {
		trackNameRegex = originalRegex
	}()
	tests := map[string]struct {
		sf        *SearchFlags
		wantValid bool
		output.WantedRecording
	}{
		"valid extension": {sf: &SearchFlags{fileExtension: &defaultExtension}, wantValid: true},
		"extension does not start with '.'": {
			sf: &SearchFlags{fileExtension: &missingLeadDot},
			WantedRecording: output.WantedRecording{
				Error: "The -ext value you specified, \"mp3\", must contain exactly one '.' and '.' must be the first character.\n",
				Log:   "level='error' -ext='mp3' msg='the file extension must begin with '.' and contain no other '.' characters'\n",
			},
		},
		"extension contains multiple '.'": {
			sf: &SearchFlags{fileExtension: &multipleDots},
			WantedRecording: output.WantedRecording{
				Error: "The -ext value you specified, \".m.p3\", must contain exactly one '.' and '.' must be the first character.\n",
				Log:   "level='error' -ext='.m.p3' msg='the file extension must begin with '.' and contain no other '.' characters'\n",
			},
		},
		"extension contains invalid characters": {
			sf: &SearchFlags{fileExtension: &badChar},
			WantedRecording: output.WantedRecording{
				Error: "The -ext value you specified, \".m[p3\", cannot be used for file matching: error parsing regexp: missing closing ]: `[p3$`.\n",
				Log:   "level='error' -ext='.m[p3' error='error parsing regexp: missing closing ]: `[p3$`' msg='the file extension cannot be parsed as a regular expression'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
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
	const fnName = "SearchFlags.ProcessArgs()"
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
	tests := map[string]struct {
		sf *SearchFlags
		args
		wantS  *Search
		wantOk bool
		output.WantedRecording
	}{
		"good arguments": {sf: flags, args: args{args: nil}, wantS: s, wantOk: true},
		"request help": {
			sf:              flags,
			args:            args{args: []string{"-help"}},
			WantedRecording: output.WantedRecording{Error: usage, Log: "level='error' arguments='[-help]' msg='flag: help requested'\n"},
		},
		"request invalid argument": {
			sf:   flags,
			args: args{args: []string{"-foo"}},
			WantedRecording: output.WantedRecording{
				Error: "flag provided but not defined: -foo\n" + usage,
				Log:   "level='error' arguments='[-foo]' msg='flag provided but not defined: -foo'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
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
	const fnName = "SearchDefaults()"
	tests := map[string]struct {
		want    string
		wantMap map[string]any
	}{
		"single use case": {
			want: "common",
			wantMap: map[string]any{
				"albumFilter":  ".*",
				"artistFilter": ".*",
				"ext":          ".mp3",
				"topDir":       filepath.Join("%HOMEPATH%", "Music"),
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, gotMap := SearchDefaults()
			if got != tt.want {
				t.Errorf("%s got = %v, want %v", fnName, got, tt.want)
			}
			if !reflect.DeepEqual(gotMap, tt.wantMap) {
				t.Errorf("%s got1 = %v, want %v", fnName, gotMap, tt.wantMap)
			}
		})
	}
}
