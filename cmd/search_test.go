package cmd_test

import (
	"io/fs"
	"mp3repair/cmd"
	"mp3repair/internal/files"
	"path/filepath"
	"reflect"
	"regexp"
	"testing"
	"time"

	cmdtoolkit "github.com/majohn-r/cmd-toolkit"
	"github.com/majohn-r/output"
	"github.com/spf13/afero"
)

func TestEvaluateFilter(t *testing.T) {
	tests := map[string]struct {
		filtering   cmd.FilterFlag
		want        cmd.EvaluatedFilter
		wantFilter  *regexp.Regexp
		wantOk      bool
		wantRegexOk bool
		output.WantedRecording
	}{
		"missing flag": {
			filtering: cmd.FilterFlag{
				Values:             map[string]*cmd.CommandFlag[any]{},
				FlagName:           "albumFilter",
				FlagRepresentation: "--albumFilter",
			},
			want: cmd.EvaluatedFilter{RegexOk: true},
			WantedRecording: output.WantedRecording{
				Error: "An internal error occurred: flag \"albumFilter\" is not found.\n",
				Log: "level='error'" +
					" error='flag not found'" +
					" flag='albumFilter'" +
					" msg='internal error'\n",
			},
		},
		"bad regex, user-supplied": {
			filtering: cmd.FilterFlag{
				Values: map[string]*cmd.CommandFlag[any]{
					"albumFilter": {UserSet: true, Value: "[9-0]"},
				},
				FlagName:           "albumFilter",
				FlagRepresentation: "--albumFilter",
			},
			want: cmd.EvaluatedFilter{},
			WantedRecording: output.WantedRecording{
				Error: "The --albumFilter value \"[9-0]\" cannot be used.\n" +
					"Why?\n" +
					"The value of --albumFilter that you specified is not a valid regular" +
					" expression: error parsing regexp: invalid character class range:" +
					" `9-0`.\n" +
					"What to do:\n" +
					"Either try a different setting, or omit setting --albumFilter and" +
					" try the default value.\n",
				Log: "level='error'" +
					" --albumFilter='[9-0]'" +
					" error='error parsing regexp: invalid character class range: `9-0`'" +
					" user-set='true'" +
					" msg='the filter cannot be parsed as a regular expression'\n",
			},
		},
		"bad regex, as configured": {
			filtering: cmd.FilterFlag{
				Values: map[string]*cmd.CommandFlag[any]{
					"albumFilter": {UserSet: false, Value: "[9-0]"},
				},
				FlagName:           "albumFilter",
				FlagRepresentation: "--albumFilter",
			},
			want: cmd.EvaluatedFilter{},
			WantedRecording: output.WantedRecording{
				Error: "The --albumFilter value \"[9-0]\" cannot be used.\n" +
					"Why?\n" +
					"The configured default value of --albumFilter is not a valid" +
					" regular expression: error parsing regexp: invalid character class" +
					" range: `9-0`.\n" +
					"What to do:\n" +
					"Either edit the defaults.yaml file containing the settings, or" +
					" explicitly set --albumFilter to a better value.\n",
				Log: "level='error'" +
					" --albumFilter='[9-0]'" +
					" error='error parsing regexp: invalid character class range: `9-0`'" +
					" user-set='false'" +
					" msg='the filter cannot be parsed as a regular expression'\n",
			},
		},
		"good regex": {
			filtering: cmd.FilterFlag{
				Values: map[string]*cmd.CommandFlag[any]{
					"albumFilter": {UserSet: true, Value: `\d`},
				},
				FlagName:           "albumFilter",
				FlagRepresentation: "--albumFilter",
			},
			want: cmd.EvaluatedFilter{
				Regex:    regexp.MustCompile(`\d`),
				FilterOk: true,
				RegexOk:  true,
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			got := cmd.EvaluateFilter(o, tt.filtering)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EvaluateFilter() got = %v, want %v", got, tt.want)
			}
			o.Report(t, "EvaluateFilter()", tt.WantedRecording)
		})
	}
}

func TestEvaluateTopDir(t *testing.T) {
	originalFileSystem := cmdtoolkit.AssignFileSystem(afero.NewMemMapFs())
	defer func() {
		cmdtoolkit.AssignFileSystem(originalFileSystem)
	}()
	goodDir := "music"
	badDir := filepath.Join(goodDir, "moreMusic")
	_ = cmdtoolkit.FileSystem().Mkdir(goodDir, cmdtoolkit.StdDirPermissions)
	_ = afero.WriteFile(cmdtoolkit.FileSystem(), badDir, []byte("data"), cmdtoolkit.StdFilePermissions)
	tests := map[string]struct {
		values  map[string]*cmd.CommandFlag[any]
		wantDir string
		wantOk  bool
		output.WantedRecording
	}{
		"missing flag": {
			values: map[string]*cmd.CommandFlag[any]{},
			WantedRecording: output.WantedRecording{
				Error: "An internal error occurred: flag \"topDir\" is not found.\n",
				Log: "level='error'" +
					" error='flag not found'" +
					" flag='topDir'" +
					" msg='internal error'\n",
			},
		},
		"non-existent file, user set": {
			values: map[string]*cmd.CommandFlag[any]{
				"topDir": {UserSet: true, Value: "no such directory"},
			},
			WantedRecording: output.WantedRecording{
				Error: "The --topDir value, \"no such directory\", cannot be used.\n" +
					"Why?\n" +
					"The value you specified is not a readable file.\n" +
					"What to do:\n" +
					"Specify a value that is a readable file.\n",
				Log: "level='error'" +
					" --topDir='no such directory'" +
					" error='open no such directory: file does not exist'" +
					" user-set='true'" +
					" msg='invalid directory'\n",
			},
		},
		"non-existent file, as configured": {
			values: map[string]*cmd.CommandFlag[any]{
				"topDir": {UserSet: false, Value: "no such directory"},
			},
			WantedRecording: output.WantedRecording{
				Error: "The --topDir value, \"no such directory\", cannot be used.\n" +
					"Why?\n" +
					"The currently configured value is not a readable file.\n" +
					"What to do:\n" +
					"Edit the configuration file or specify --topDir with a value that is" +
					" a readable file.\n",
				Log: "level='error'" +
					" --topDir='no such directory'" +
					" error='open no such directory: file does not exist'" +
					" user-set='false'" +
					" msg='invalid directory'\n",
			},
		},
		"non-existent directory, user set": {
			values: map[string]*cmd.CommandFlag[any]{
				"topDir": {UserSet: true, Value: badDir},
			},
			WantedRecording: output.WantedRecording{
				Error: "The --topDir value, \"music\\\\moreMusic\", cannot be used.\n" +
					"Why?\n" +
					"The value you specified is not the name of a directory.\n" +
					"What to do:\n" +
					"Specify a value that is the name of a directory.\n",
				Log: "level='error'" +
					" --topDir='" + badDir + "'" +
					" user-set='true'" +
					" msg='the file is not a directory'\n",
			},
		},
		"non-existent directory, as configured": {
			values: map[string]*cmd.CommandFlag[any]{
				"topDir": {UserSet: false, Value: badDir},
			},
			WantedRecording: output.WantedRecording{
				Error: "The --topDir value, \"music\\\\moreMusic\", cannot be used.\n" +
					"Why?\n" +
					"The currently configured value is not the name of a directory.\n" +
					"What to do:\n" +
					"Edit the configuration file or specify --topDir with a value that is" +
					" the name of a directory.\n",
				Log: "level='error'" +
					" --topDir='" + badDir + "'" +
					" user-set='false'" +
					" msg='the file is not a directory'\n",
			},
		},
		"valid directory": {
			values: map[string]*cmd.CommandFlag[any]{
				"topDir": {UserSet: false, Value: goodDir},
			},
			wantDir: goodDir,
			wantOk:  true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			gotDir, gotOk := cmd.EvaluateTopDir(o, tt.values)
			if gotDir != tt.wantDir {
				t.Errorf("EvaluateTopDir() gotDir = %v, want %v", gotDir, tt.wantDir)
			}
			if gotOk != tt.wantOk {
				t.Errorf("EvaluateTopDir() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
			o.Report(t, "EvaluateTopDir()", tt.WantedRecording)
		})
	}
}

func TestProcessSearchFlags(t *testing.T) {
	tests := map[string]struct {
		values       map[string]*cmd.CommandFlag[any]
		wantSettings *cmd.SearchSettings
		wantOk       bool
		output.WantedRecording
	}{
		"no data": {
			values:       map[string]*cmd.CommandFlag[any]{},
			wantSettings: &cmd.SearchSettings{},
			WantedRecording: output.WantedRecording{
				Error: "An internal error occurred: flag \"albumFilter\" is not found.\n" +
					"An internal error occurred: flag \"artistFilter\" is not found.\n" +
					"An internal error occurred: flag \"trackFilter\" is not found.\n" +
					"An internal error occurred: flag \"topDir\" is not found.\n" +
					"An internal error occurred: flag \"extensions\" is not found.\n",
				Log: "level='error'" +
					" error='flag not found'" +
					" flag='albumFilter'" +
					" msg='internal error'\n" +
					"level='error'" +
					" error='flag not found'" +
					" flag='artistFilter'" +
					" msg='internal error'\n" +
					"level='error'" +
					" error='flag not found'" +
					" flag='trackFilter'" +
					" msg='internal error'\n" +
					"level='error'" +
					" error='flag not found'" +
					" flag='topDir'" +
					" msg='internal error'\n" +
					"level='error'" +
					" error='flag not found'" +
					" flag='extensions'" +
					" msg='internal error'\n",
			},
		},
		"bad data": {
			values: map[string]*cmd.CommandFlag[any]{
				"albumFilter":  {Value: "[2"},
				"artistFilter": {Value: "[1-0]"},
				"trackFilter":  {Value: "0++"},
				"topDir":       {Value: "no such dir"},
				"extensions":   {Value: "foo,bar"},
			},
			wantSettings: &cmd.SearchSettings{},
			WantedRecording: output.WantedRecording{
				Error: "The --albumFilter value \"[2\" cannot be used.\n" +
					"Why?\n" +
					"The configured default value of --albumFilter is not a valid regular" +
					" expression: error parsing regexp: missing closing ]: `[2`.\n" +
					"What to do:\n" +
					"Either edit the defaults.yaml file containing the settings, or" +
					" explicitly set --albumFilter to a better value.\n" +
					"The --artistFilter value \"[1-0]\" cannot be used.\n" +
					"Why?\nThe configured default value of --artistFilter is not a valid" +
					" regular expression: error parsing regexp: invalid character class range: `1-0`.\n" +
					"What to do:\n" +
					"Either edit the defaults.yaml file containing the settings, or" +
					" explicitly set --artistFilter to a better value.\n" +
					"The --trackFilter value \"0++\" cannot be used.\n" +
					"Why?\n" +
					"The configured default value of --trackFilter is not a valid regular" +
					" expression: error parsing regexp: invalid nested repetition operator: `++`.\n" +
					"What to do:\n" +
					"Either edit the defaults.yaml file containing the settings, or" +
					" explicitly set --trackFilter to a better value.\n" +
					"Here are some common errors in filter expressions and what to do:\n" +
					"Character class problems\n" +
					"Character classes are sets of 1 or more characters, enclosed in square" +
					" brackets: []\n" +
					"A common error is to forget the final ] bracket.\n" +
					"Character classes can include a range of characters, like this: [a-z]," +
					" which means\n" +
					"any character between a and z. Order is important - one might think" +
					" that [z-a] would\n" +
					"mean the same thing, but it doesn't; z comes after a. Do an internet" +
					" search for ASCII\n" +
					"table; that's the expected order for ranges of characters. And that" +
					" means [A-z] means\n" +
					"any letter, and [a-Z] is an error.\n" +
					"Repetition problems\n" +
					"The characters '+' and '*' specify repetition: a+ means \"exactly" +
					" one a\" and a* means\n" +
					"\"0 or more a's\". You can also put a count in curly braces - a{2}" +
					" means \"exactly two a's\".\n" +
					"Repetition can only be used once for a character or character class." +
					" 'a++', 'a+*',\n" +
					"and so on, are not allowed.\n" +
					"For more (too much, often, you are warned) information, do a web" +
					" search for\n" +
					"\"golang regexp\".\n" +
					"The --topDir value, \"no such dir\", cannot be used.\n" +
					"Why?\n" +
					"The currently configured value is not a readable file.\n" +
					"What to do:\n" +
					"Edit the configuration file or specify --topDir with a value that" +
					" is a readable file.\n" +
					"The extension \"foo\" cannot be used.\n" +
					"The extension \"bar\" cannot be used.\n" +
					"Why?\n" +
					"Extensions must be at least two characters long and begin with '.'.\n" +
					"What to do:\n" +
					"Provide appropriate extensions.\n",
				Log: "level='error'" +
					" --albumFilter='[2'" +
					" error='error parsing regexp: missing closing ]: `[2`'" +
					" user-set='false'" +
					" msg='the filter cannot be parsed as a regular expression'\n" +
					"level='error'" +
					" --artistFilter='[1-0]'" +
					" error='error parsing regexp: invalid character class range: `1-0`'" +
					" user-set='false'" +
					" msg='the filter cannot be parsed as a regular expression'\n" +
					"level='error'" +
					" --trackFilter='0++'" +
					" error='error parsing regexp: invalid nested repetition operator:" +
					" `++`'" +
					" user-set='false'" +
					" msg='the filter cannot be parsed as a regular expression'\n" +
					"level='error'" +
					" --topDir='no such dir'" +
					" error='CreateFile no such dir: The system cannot find the file" +
					" specified.'" +
					" user-set='false'" +
					" msg='invalid directory'\n" +
					"level='error'" +
					" --extensions='foo,bar'" +
					" rejected='[foo bar]'" +
					" msg='invalid file extensions'\n",
			},
		},
		"good data": {
			values: map[string]*cmd.CommandFlag[any]{
				"albumFilter":  {Value: "[23]"},
				"artistFilter": {Value: "[0-7]"},
				"trackFilter":  {Value: "0+"},
				"topDir":       {Value: "."},
				"extensions":   {Value: ".mp3"},
			},
			wantSettings: &cmd.SearchSettings{
				ArtistFilter:   regexp.MustCompile("[0-7]"),
				AlbumFilter:    regexp.MustCompile("[23]"),
				TrackFilter:    regexp.MustCompile("0+"),
				FileExtensions: []string{".mp3"},
				TopDirectory:   ".",
			},
			wantOk: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			gotSettings, gotOk := cmd.ProcessSearchFlags(o, tt.values)
			if !reflect.DeepEqual(gotSettings, tt.wantSettings) {
				t.Errorf("ProcessSearchFlags() gotSettings = %v, want %v", gotSettings, tt.wantSettings)
			}
			if gotOk != tt.wantOk {
				t.Errorf("ProcessSearchFlags() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
			o.Report(t, "ProcessSearchFlags()", tt.WantedRecording)
		})
	}
}

func TestEvaluateSearchFlags(t *testing.T) {
	tests := map[string]struct {
		producer     cmd.FlagProducer
		wantSettings *cmd.SearchSettings
		wantOk       bool
		output.WantedRecording
	}{
		"errors": {
			producer:     testFlagProducer{},
			wantSettings: &cmd.SearchSettings{},
			WantedRecording: output.WantedRecording{
				Error: "An internal error occurred: flag \"albumFilter\" does not exist.\n" +
					"An internal error occurred: flag \"artistFilter\" does not exist.\n" +
					"An internal error occurred: flag \"extensions\" does not exist.\n" +
					"An internal error occurred: flag \"topDir\" does not exist.\n" +
					"An internal error occurred: flag \"trackFilter\" does not exist.\n",
				Log: "level='error'" +
					" error='flag \"albumFilter\" does not exist'" +
					" msg='internal error'\n" +
					"level='error'" +
					" error='flag \"artistFilter\" does not exist'" +
					" msg='internal error'\n" +
					"level='error'" +
					" error='flag \"extensions\" does not exist'" +
					" msg='internal error'\n" +
					"level='error'" +
					" error='flag \"topDir\" does not exist'" +
					" msg='internal error'\n" +
					"level='error'" +
					" error='flag \"trackFilter\" does not exist'" +
					" msg='internal error'\n",
			},
		},
		"good data": {
			producer: testFlagProducer{
				flags: map[string]testFlag{
					"albumFilter":  {value: "\\d+", valueKind: cmd.StringType},
					"artistFilter": {value: "Beatles", valueKind: cmd.StringType},
					"trackFilter":  {value: "Sadie", valueKind: cmd.StringType},
					"topDir":       {value: ".", valueKind: cmd.StringType},
					"extensions":   {value: ".mp3", valueKind: cmd.StringType},
				},
			},
			wantSettings: &cmd.SearchSettings{
				ArtistFilter:   regexp.MustCompile("Beatles"),
				AlbumFilter:    regexp.MustCompile(`\d+`),
				TrackFilter:    regexp.MustCompile("Sadie"),
				FileExtensions: []string{".mp3"},
				TopDirectory:   ".",
			},
			wantOk: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			gotSettings, gotOk := cmd.EvaluateSearchFlags(o, tt.producer)
			if !reflect.DeepEqual(gotSettings, tt.wantSettings) {
				t.Errorf("EvaluateSearchFlags() gotSettings = %v, want %v", gotSettings, tt.wantSettings)
			}
			if gotOk != tt.wantOk {
				t.Errorf("EvaluateSearchFlags() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
			o.Report(t, "EvaluateSearchFlags()", tt.WantedRecording)
		})
	}
}

func TestEvaluateFileExtensions(t *testing.T) {
	tests := map[string]struct {
		values map[string]*cmd.CommandFlag[any]
		want   []string
		want1  bool
		output.WantedRecording
	}{
		"no data": {
			values: map[string]*cmd.CommandFlag[any]{},
			want:   []string{},
			want1:  false,
			WantedRecording: output.WantedRecording{
				Error: "An internal error occurred: flag \"extensions\" is not found.\n",
				Log: "level='error'" +
					" error='flag not found'" +
					" flag='extensions'" +
					" msg='internal error'\n",
			},
		},
		"one extension": {
			values: map[string]*cmd.CommandFlag[any]{"extensions": {Value: ".mp3"}},
			want:   []string{".mp3"},
			want1:  true,
		},
		"two extensions": {
			values: map[string]*cmd.CommandFlag[any]{"extensions": {Value: ".mp3,.mPThree"}},
			want:   []string{".mp3", ".mPThree"},
			want1:  true,
		},
		"bad extensions": {
			values: map[string]*cmd.CommandFlag[any]{"extensions": {Value: ".mp3,,foo,."}},
			want:   []string{".mp3"},
			want1:  false,
			WantedRecording: output.WantedRecording{
				Error: "The extension \"\" cannot be used.\n" +
					"The extension \"foo\" cannot be used.\n" +
					"The extension \".\" cannot be used.\n" +
					"Why?\n" +
					"Extensions must be at least two characters long and begin with '.'.\n" +
					"What to do:\n" +
					"Provide appropriate extensions.\n",
				Log: "level='error'" +
					" --extensions='.mp3,,foo,.'" +
					" rejected='[ foo .]'" +
					" msg='invalid file extensions'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			got, got1 := cmd.EvaluateFileExtensions(o, tt.values)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EvaluateFileExtensions() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("EvaluateFileExtensions() got1 = %v, want %v", got1, tt.want1)
			}
			o.Report(t, "EvaluateFileExtensions()", tt.WantedRecording)
		})
	}
}

type testFile struct {
	name  string
	mode  fs.FileMode
	files []*testFile
}

func (tf *testFile) Name() string {
	return tf.name
}

func (tf *testFile) IsDir() bool {
	return tf.mode.IsDir()
}

func (tf *testFile) Mode() fs.FileMode {
	return tf.mode
}

func (tf *testFile) ModTime() time.Time {
	return time.Now()
}

func (tf *testFile) Size() int64 {
	return 0
}

func (tf *testFile) Sys() any {
	return nil
}

func newTestFile(name string, contents []*testFile) *testFile {
	fm := fs.ModeDir
	if len(contents) == 0 {
		fm = 0
	}
	return &testFile{name: name, files: contents, mode: fm}
}

func TestSearchSettingsLoad(t *testing.T) {
	originalReadDirectory := cmd.ReadDirectory
	defer func() {
		cmd.ReadDirectory = originalReadDirectory
	}()
	album1Content1 := newTestFile("subfolder", []*testFile{newTestFile("foo", nil)})
	album1Content2 := newTestFile("cover.jpg", nil)
	album1Content3 := newTestFile("1 lovely music.mp3", nil)
	album1 := newTestFile("album", []*testFile{album1Content1, album1Content2, album1Content3})
	album2 := newTestFile("not an album", nil)
	artist1 := newTestFile("artist", []*testFile{album1, album2})
	artist2 := newTestFile("not an artist", nil)
	topDir := newTestFile("music", []*testFile{artist1, artist2})
	testFiles := map[string]*testFile{
		topDir.name:                                           topDir,
		filepath.Join(topDir.name, artist1.name):              artist1,
		filepath.Join(topDir.name, artist2.name):              artist2,
		filepath.Join(topDir.name, artist1.name, album1.name): album1,
		filepath.Join(topDir.name, artist1.name, album2.name): album2,
		filepath.Join(topDir.name, artist1.name, album1.name,
			album1Content1.name): album1Content1,
		filepath.Join(topDir.name, artist1.name, album1.name,
			album1Content2.name): album1Content2,
		filepath.Join(topDir.name, artist1.name, album1.name,
			album1Content3.name): album1Content3,
	}
	testArtist := files.NewArtistFromFile(artist1, topDir.name)
	testAlbum := files.NewAlbumFromFile(album1, testArtist)
	testArtist.AddAlbum(testAlbum)
	testTrack := files.TrackMaker{
		Album:      testAlbum,
		FileName:   album1Content3.name,
		SimpleName: "lovely music",
		Number:     1,
	}.NewTrack()
	testAlbum.AddTrack(testTrack)
	cmd.ReadDirectory = func(_ output.Bus, dir string) ([]fs.FileInfo, bool) {
		if tf, found := testFiles[dir]; found {
			var entries []fs.FileInfo
			for _, f := range tf.files {
				entries = append(entries, f)
			}
			return entries, true
		}
		return []fs.FileInfo{}, false
	}
	tests := map[string]struct {
		ss   *cmd.SearchSettings
		want []*files.Artist
		output.WantedRecording
	}{
		"topDir read error": {
			ss:   &cmd.SearchSettings{TopDirectory: "td"},
			want: []*files.Artist{},
			WantedRecording: output.WantedRecording{
				Error: "No mp3 files could be found using the specified parameters.\n" +
					"Why?\n" +
					"There were no directories found in \"td\" (the --topDir value).\n" +
					"What to do:\n" +
					"Set --topDir to the path of a directory that contains artist" +
					" directories.\n",
				Log: "level='error'" +
					" --topDir='td'" +
					" msg='cannot find any artist directories'\n",
			},
		},
		"good read": {
			ss: &cmd.SearchSettings{
				FileExtensions: []string{".mp3"},
				TopDirectory:   "music",
			},
			want: []*files.Artist{testArtist},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			got := tt.ss.Load(o)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SearchSettings.Load() got = %v, want %v", got, tt.want)
			}
			o.Report(t, "SearchSettings.Load()", tt.WantedRecording)
		})
	}
}

func TestSearchSettingsFilter(t *testing.T) {
	artist1 := files.NewArtist("A", filepath.Join("music", "A"))
	albumA1 := files.AlbumMaker{
		Title:  "A1",
		Artist: artist1,
		Path:   filepath.Join(artist1.FilePath, "A1"),
	}.NewAlbum()
	trackA11 := files.TrackMaker{
		Album:      albumA1,
		FileName:   "1 A11.mp3",
		SimpleName: "A11",
		Number:     1,
	}.NewTrack()
	albumA1.AddTrack(trackA11)
	trackA12 := files.TrackMaker{
		Album:      albumA1,
		FileName:   "2 B12.mp3",
		SimpleName: "B12",
		Number:     1,
	}.NewTrack()
	albumA1.AddTrack(trackA12)
	artist1.AddAlbum(albumA1)
	albumA2 := files.AlbumMaker{
		Title:  "B1",
		Artist: artist1,
		Path:   filepath.Join(artist1.FilePath, "B1"),
	}.NewAlbum()
	trackA21 := files.TrackMaker{
		Album:      albumA2,
		FileName:   "1 A21.mp3",
		SimpleName: "A21",
		Number:     1,
	}.NewTrack()
	albumA2.AddTrack(trackA21)
	trackA22 := files.TrackMaker{
		Album:      albumA2,
		FileName:   "2 B22.mp3",
		SimpleName: "B22",
		Number:     1,
	}.NewTrack()
	albumA2.AddTrack(trackA22)
	artist1.AddAlbum(albumA2)
	// add empty album
	artist1.AddAlbum(files.AlbumMaker{
		Title:  "A2",
		Artist: artist1,
		Path:   filepath.Join(artist1.FilePath, "A2"),
	}.NewAlbum())
	artist2 := files.NewArtist("B", filepath.Join("music", "B"))
	albumB1 := files.AlbumMaker{
		Title:  "B1",
		Artist: artist2,
		Path:   filepath.Join(artist2.FilePath, "B1"),
	}.NewAlbum()
	trackB11 := files.TrackMaker{
		Album:      albumB1,
		FileName:   "1 A11a.mp3",
		SimpleName: "A11a",
		Number:     1,
	}.NewTrack()
	albumB1.AddTrack(trackB11)
	trackB12 := files.TrackMaker{
		Album:      albumB1,
		FileName:   "2 B12a.mp3",
		SimpleName: "B12a",
		Number:     1,
	}.NewTrack()
	albumB1.AddTrack(trackB12)
	artist2.AddAlbum(albumB1)
	albumB2 := files.AlbumMaker{
		Title:  "B2",
		Artist: artist2,
		Path:   filepath.Join(artist2.FilePath, "B2"),
	}.NewAlbum()
	trackB21 := files.TrackMaker{
		Album:      albumB2,
		FileName:   "1 A21a.mp3",
		SimpleName: "A21a",
		Number:     1,
	}.NewTrack()
	albumB2.AddTrack(trackB21)
	trackB22 := files.TrackMaker{
		Album:      albumB2,
		FileName:   "2 B22a.mp3",
		SimpleName: "B22a",
		Number:     1,
	}.NewTrack()
	albumB2.AddTrack(trackB22)
	artist2.AddAlbum(albumB2)
	// create empty artist
	artist3 := files.NewArtist("AA", filepath.Join("music", "AA"))
	filteredArtist1 := artist1.Copy()
	filteredAlbumA1 := albumA1.Copy(filteredArtist1, false)
	filteredTrackA11 := trackA11.Copy(filteredAlbumA1)
	filteredAlbumA1.AddTrack(filteredTrackA11)
	filteredArtist1.AddAlbum(filteredAlbumA1)
	tests := map[string]struct {
		ss              *cmd.SearchSettings
		originalArtists []*files.Artist
		want            []*files.Artist
		output.WantedRecording
	}{
		"nothing to filter": {
			ss: &cmd.SearchSettings{
				ArtistFilter: regexp.MustCompile(".*"),
				AlbumFilter:  regexp.MustCompile(".*"),
				TrackFilter:  regexp.MustCompile(".*"),
			},
			originalArtists: []*files.Artist{},
			want:            []*files.Artist{},
			WantedRecording: output.WantedRecording{
				Error: "No mp3 files remain after filtering.\n" +
					"Why?\n" +
					"After applying --artistFilter=\".*\", --albumFilter=\".*\", and" +
					" --trackFilter=\".*\", no files remained.\n" +
					"What to do:\n" +
					"Use less restrictive filter settings.\n",
				Log: "level='error'" +
					" --albumFilter='.*'" +
					" --artistFilter='.*'" +
					" --trackFilter='.*'" +
					" msg='no files remain after filtering'\n",
			},
		},
		"filter out everything": {
			ss: &cmd.SearchSettings{
				ArtistFilter: regexp.MustCompile("^$"),
				AlbumFilter:  regexp.MustCompile("^$"),
				TrackFilter:  regexp.MustCompile("^$"),
			},
			originalArtists: []*files.Artist{artist1, artist2, artist3},
			want:            []*files.Artist{},
			WantedRecording: output.WantedRecording{
				Error: "No mp3 files remain after filtering.\n" +
					"Why?\n" +
					"After applying --artistFilter=\"^$\", --albumFilter=\"^$\", and" +
					" --trackFilter=\"^$\", no files remained.\n" +
					"What to do:\n" +
					"Use less restrictive filter settings.\n",
				Log: "level='error'" +
					" --albumFilter='^$'" +
					" --artistFilter='^$'" +
					" --trackFilter='^$'" +
					" msg='no files remain after filtering'\n",
			},
		},
		"filter out selectively": {
			ss: &cmd.SearchSettings{
				ArtistFilter: regexp.MustCompile("^A"),
				AlbumFilter:  regexp.MustCompile("^A"),
				TrackFilter:  regexp.MustCompile("^A"),
			},
			originalArtists: []*files.Artist{artist1, artist2, artist3},
			want:            []*files.Artist{filteredArtist1},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			got := tt.ss.Filter(o, tt.originalArtists)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SearchSettings.Filter() got = %v, want %v", got, tt.want)
			}
			o.Report(t, "SearchSettings.Filter()", tt.WantedRecording)
		})
	}
}
