package cmd_test

import (
	"io/fs"
	"mp3/cmd"
	"mp3/internal/files"
	"path/filepath"
	"reflect"
	"regexp"
	"testing"

	"github.com/majohn-r/output"
)

func TestEvaluateFilter(t *testing.T) {
	type args struct {
		values     map[string]*cmd.FlagValue
		flagName   string
		nameAsFlag string
	}
	tests := map[string]struct {
		args
		wantFilter  *regexp.Regexp
		wantOk      bool
		wantRegexOk bool
		output.WantedRecording
	}{
		"missing flag": {
			args: args{
				values:     map[string]*cmd.FlagValue{},
				flagName:   "albumFilter",
				nameAsFlag: "--albumFilter",
			},
			wantRegexOk: true, // not a bad regex, right?
			WantedRecording: output.WantedRecording{
				Error: "An internal error occurred: flag \"albumFilter\" is not found.\n",
				Log: "level='error'" +
					" error='flag not found'" +
					" flag='albumFilter'" +
					" msg='internal error'\n",
			},
		},
		"bad regex, user-supplied": {
			args: args{
				values: map[string]*cmd.FlagValue{
					"albumFilter": cmd.NewFlagValue().WithExplicitlySet(
						true).WithValueType(cmd.StringType).WithValue("[9-0]"),
				},
				flagName:   "albumFilter",
				nameAsFlag: "--albumFilter",
			},
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
			args: args{
				values: map[string]*cmd.FlagValue{
					"albumFilter": cmd.NewFlagValue().WithExplicitlySet(
						false).WithValueType(cmd.StringType).WithValue("[9-0]"),
				},
				flagName:   "albumFilter",
				nameAsFlag: "--albumFilter",
			},
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
			args: args{
				values: map[string]*cmd.FlagValue{
					"albumFilter": cmd.NewFlagValue().WithExplicitlySet(
						true).WithValueType(cmd.StringType).WithValue(`\d`),
				},
				flagName:   "albumFilter",
				nameAsFlag: "--albumFilter",
			},
			wantFilter:  regexp.MustCompile(`\d`),
			wantOk:      true,
			wantRegexOk: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			gotFilter, gotOk, gotRegexOk := cmd.EvaluateFilter(o, tt.args.values,
				tt.args.flagName, tt.args.nameAsFlag)
			if !reflect.DeepEqual(gotFilter, tt.wantFilter) {
				t.Errorf("EvaluateFilter() gotFilter = %v, want %v", gotFilter,
					tt.wantFilter)
			}
			if gotOk != tt.wantOk {
				t.Errorf("EvaluateFilter() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
			if gotRegexOk != tt.wantRegexOk {
				t.Errorf("EvaluateFilter() gotRegexOk = %v, want %v", gotRegexOk,
					tt.wantRegexOk)
			}
			if differences, ok := o.Verify(tt.WantedRecording); !ok {
				for _, difference := range differences {
					t.Errorf("EvaluateFilter() %s", difference)
				}
			}
		})
	}
}

func TestEvaluateTopDir(t *testing.T) {
	tests := map[string]struct {
		values  map[string]*cmd.FlagValue
		wantDir string
		wantOk  bool
		output.WantedRecording
	}{
		"missing flag": {
			values: map[string]*cmd.FlagValue{},
			WantedRecording: output.WantedRecording{
				Error: "An internal error occurred: flag \"topDir\" is not found.\n",
				Log: "level='error'" +
					" error='flag not found'" +
					" flag='topDir'" +
					" msg='internal error'\n",
			},
		},
		"non-existent file, user set": {
			values: map[string]*cmd.FlagValue{
				"topDir": cmd.NewFlagValue().WithExplicitlySet(true).WithValueType(
					cmd.StringType).WithValue("no such directory"),
			},
			WantedRecording: output.WantedRecording{
				Error: "The --topDir value, \"no such directory\", cannot be used.\n" +
					"Why?\n" +
					"The value you specified is not a readable file.\n" +
					"What to do:\n" +
					"Specify a value that is a readable file.\n",
				Log: "level='error'" +
					" --topDir='no such directory'" +
					" error='CreateFile no such directory: The system cannot find the file" +
					" specified.'" +
					" user-set='true'" +
					" msg='invalid directory'\n",
			},
		},
		"non-existent file, as configured": {
			values: map[string]*cmd.FlagValue{
				"topDir": cmd.NewFlagValue().WithExplicitlySet(false).WithValueType(
					cmd.StringType).WithValue("no such directory"),
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
					" error='CreateFile no such directory: The system cannot find the file" +
					" specified.'" +
					" user-set='false'" +
					" msg='invalid directory'\n",
			},
		},
		"non-existent directory, user set": {
			values: map[string]*cmd.FlagValue{
				"topDir": cmd.NewFlagValue().WithExplicitlySet(true).WithValueType(
					cmd.StringType).WithValue("./commonFlags_test.go"),
			},
			WantedRecording: output.WantedRecording{
				Error: "The --topDir value, \"./commonFlags_test.go\", cannot be used.\n" +
					"Why?\n" +
					"The value you specified is not the name of a directory.\n" +
					"What to do:\n" +
					"Specify a value that is the name of a directory.\n",
				Log: "level='error'" +
					" --topDir='./commonFlags_test.go'" +
					" user-set='true'" +
					" msg='the file is not a directory'\n",
			},
		},
		"non-existent directory, as configured": {
			values: map[string]*cmd.FlagValue{
				"topDir": cmd.NewFlagValue().WithExplicitlySet(false).WithValueType(
					cmd.StringType).WithValue("./commonFlags_test.go"),
			},
			WantedRecording: output.WantedRecording{
				Error: "The --topDir value, \"./commonFlags_test.go\", cannot be used.\n" +
					"Why?\n" +
					"The currently configured value is not the name of a directory.\n" +
					"What to do:\n" +
					"Edit the configuration file or specify --topDir with a value that is" +
					" the name of a directory.\n",
				Log: "level='error'" +
					" --topDir='./commonFlags_test.go'" +
					" user-set='false'" +
					" msg='the file is not a directory'\n",
			},
		},
		"valid directory": {
			values: map[string]*cmd.FlagValue{
				"topDir": cmd.NewFlagValue().WithExplicitlySet(false).WithValueType(
					cmd.StringType).WithValue("."),
			},
			wantDir: ".",
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
			if differences, ok := o.Verify(tt.WantedRecording); !ok {
				for _, difference := range differences {
					t.Errorf("EvaluateTopDir() %s", difference)
				}
			}
		})
	}
}

func TestProcessSearchFlags(t *testing.T) {
	tests := map[string]struct {
		values       map[string]*cmd.FlagValue
		wantSettings *cmd.SearchSettings
		wantOk       bool
		output.WantedRecording
	}{
		"no data": {
			values:       map[string]*cmd.FlagValue{},
			wantSettings: cmd.NewSearchSettings(),
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
			values: map[string]*cmd.FlagValue{
				"albumFilter": cmd.NewFlagValue().WithValueType(cmd.StringType).WithValue(
					"[2"),
				"artistFilter": cmd.NewFlagValue().WithValueType(cmd.StringType).WithValue(
					"[1-0]"),
				"trackFilter": cmd.NewFlagValue().WithValueType(cmd.StringType).WithValue(
					"0++"),
				"topDir": cmd.NewFlagValue().WithValueType(cmd.StringType).WithValue(
					"no such dir"),
				"extensions": cmd.NewFlagValue().WithValueType(cmd.StringType).WithValue(
					"foo,bar"),
			},
			wantSettings: cmd.NewSearchSettings(),
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
			values: map[string]*cmd.FlagValue{
				"albumFilter": cmd.NewFlagValue().WithValueType(
					cmd.StringType).WithValue("[23]"),
				"artistFilter": cmd.NewFlagValue().WithValueType(
					cmd.StringType).WithValue("[0-7]"),
				"trackFilter": cmd.NewFlagValue().WithValueType(
					cmd.StringType).WithValue("0+"),
				"topDir": cmd.NewFlagValue().WithValueType(
					cmd.StringType).WithValue("."),
				"extensions": cmd.NewFlagValue().WithValueType(
					cmd.StringType).WithValue(".mp3"),
			},
			wantSettings: cmd.NewSearchSettings().WithAlbumFilter(
				regexp.MustCompile("[23]")).WithArtistFilter(
				regexp.MustCompile("[0-7]")).WithTrackFilter(
				regexp.MustCompile("0+")).WithTopDirectory(".").WithFileExtensions(
				[]string{".mp3"}),
			wantOk: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			gotSettings, gotOk := cmd.ProcessSearchFlags(o, tt.values)
			if !reflect.DeepEqual(gotSettings, tt.wantSettings) {
				t.Errorf("ProcessSearchFlags() gotSettings = %v, want %v", gotSettings,
					tt.wantSettings)
			}
			if gotOk != tt.wantOk {
				t.Errorf("ProcessSearchFlags() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
			if differences, ok := o.Verify(tt.WantedRecording); !ok {
				for _, difference := range differences {
					t.Errorf("ProcessSearchFlags() %s", difference)
				}
			}
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
			wantSettings: cmd.NewSearchSettings(),
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
			wantSettings: cmd.NewSearchSettings().WithAlbumFilter(
				regexp.MustCompile(`\d+`)).WithArtistFilter(
				regexp.MustCompile("Beatles")).WithTrackFilter(
				regexp.MustCompile("Sadie")).WithTopDirectory(".").WithFileExtensions(
				[]string{".mp3"}),
			wantOk: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			gotSettings, gotOk := cmd.EvaluateSearchFlags(o, tt.producer)
			if !reflect.DeepEqual(gotSettings, tt.wantSettings) {
				t.Errorf("EvaluateSearchFlags() gotSettings = %v, want %v", gotSettings,
					tt.wantSettings)
			}
			if gotOk != tt.wantOk {
				t.Errorf("EvaluateSearchFlags() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
			if differences, ok := o.Verify(tt.WantedRecording); !ok {
				for _, difference := range differences {
					t.Errorf("EvaluateSearchFlags() %s", difference)
				}
			}
		})
	}
}

func TestEvaluateFileExtensions(t *testing.T) {
	tests := map[string]struct {
		values map[string]*cmd.FlagValue
		want   []string
		want1  bool
		output.WantedRecording
	}{
		"no data": {
			values: map[string]*cmd.FlagValue{},
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
			values: map[string]*cmd.FlagValue{
				"extensions": cmd.NewFlagValue().WithValueType(cmd.StringType).WithValue(
					".mp3"),
			},
			want:  []string{".mp3"},
			want1: true,
		},
		"two extensions": {
			values: map[string]*cmd.FlagValue{
				"extensions": cmd.NewFlagValue().WithValueType(cmd.StringType).WithValue(
					".mp3,.mpthree"),
			},
			want:  []string{".mp3", ".mpthree"},
			want1: true,
		},
		"bad extensions": {
			values: map[string]*cmd.FlagValue{
				"extensions": cmd.NewFlagValue().WithValueType(cmd.StringType).WithValue(
					".mp3,,foo,."),
			},
			want:  []string{".mp3"},
			want1: false,
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
			if differences, ok := o.Verify(tt.WantedRecording); !ok {
				for _, difference := range differences {
					t.Errorf("EvaluateFileExtensions() %s", difference)
				}
			}
		})
	}
}

type testFile struct {
	name  string
	files []*testFile
}

func (tf *testFile) Name() string {
	return tf.name
}

func (tf *testFile) IsDir() bool {
	return len(tf.files) > 0
}

func (tf *testFile) Type() fs.FileMode {
	if tf.IsDir() {
		return fs.ModeDir
	}
	return 0
}

func (tf *testFile) Info() (fs.FileInfo, error) {
	return nil, nil
}

func newTestFile(name string, content []*testFile) *testFile {
	return &testFile{name: name, files: content}
}

func TestSearchSettingsLoad(t *testing.T) {
	originalReadDirectory := cmd.ReadDirectory
	defer func() {
		cmd.ReadDirectory = originalReadDirectory
	}()
	album1Content1 := newTestFile("subfolder", []*testFile{newTestFile("foo", nil)})
	album1Content2 := newTestFile("cover.jpg", nil)
	album1Content3 := newTestFile("1 lovely music.mp3", nil)
	album1 := newTestFile("album", []*testFile{album1Content1, album1Content2,
		album1Content3})
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
	testTrack := files.NewTrack(testAlbum, album1Content3.name, "lovely music", 1)
	testAlbum.AddTrack(testTrack)
	cmd.ReadDirectory = func(_ output.Bus, dir string) ([]fs.DirEntry, bool) {
		if tf, ok := testFiles[dir]; ok {
			entries := []fs.DirEntry{}
			for _, f := range tf.files {
				entries = append(entries, f)
			}
			return entries, true
		}
		return []fs.DirEntry{}, false
	}
	tests := map[string]struct {
		ss    *cmd.SearchSettings
		want  []*files.Artist
		want1 bool
		output.WantedRecording
	}{
		"topDir read error": {
			ss:    cmd.NewSearchSettings().WithTopDirectory("td"),
			want:  []*files.Artist{},
			want1: false,
			WantedRecording: output.WantedRecording{
				Error: "No music files could be found using the specified parameters.\n" +
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
			ss: cmd.NewSearchSettings().WithTopDirectory("music").WithFileExtensions(
				[]string{".mp3"}),
			want:  []*files.Artist{testArtist},
			want1: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			got, got1 := tt.ss.Load(o)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SearchSettings.Load() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("SearchSettings.Load() got1 = %v, want %v", got1, tt.want1)
			}
			if differences, ok := o.Verify(tt.WantedRecording); !ok {
				for _, difference := range differences {
					t.Errorf("SearchSettings.Load() %s", difference)
				}
			}
		})
	}
}

func TestSearchSettingsFilter(t *testing.T) {
	artist1 := files.NewArtist("A", filepath.Join("music", "A"))
	albumA1 := files.NewAlbum("A1", artist1, filepath.Join(artist1.Path(), "A1"))
	trackA11 := files.NewTrack(albumA1, "1 A11.mp3", "A11", 1)
	albumA1.AddTrack(trackA11)
	trackA12 := files.NewTrack(albumA1, "2 B12.mp3", "B12", 1)
	albumA1.AddTrack(trackA12)
	artist1.AddAlbum(albumA1)
	albumA2 := files.NewAlbum("B1", artist1, filepath.Join(artist1.Path(), "B1"))
	trackA21 := files.NewTrack(albumA2, "1 A21.mp3", "A21", 1)
	albumA2.AddTrack(trackA21)
	trackA22 := files.NewTrack(albumA2, "2 B22.mp3", "B22", 1)
	albumA2.AddTrack(trackA22)
	artist1.AddAlbum(albumA2)
	// add empty album
	artist1.AddAlbum(files.NewAlbum("A2", artist1, filepath.Join(artist1.Path(), "A2")))
	artist2 := files.NewArtist("B", filepath.Join("music", "B"))
	albumB1 := files.NewAlbum("B1", artist2, filepath.Join(artist2.Path(), "B1"))
	trackB11 := files.NewTrack(albumB1, "1 A11a.mp3", "A11a", 1)
	albumB1.AddTrack(trackB11)
	trackB12 := files.NewTrack(albumB1, "2 B12a.mp3", "B12a", 1)
	albumB1.AddTrack(trackB12)
	artist2.AddAlbum(albumB1)
	albumB2 := files.NewAlbum("B2", artist2, filepath.Join(artist2.Path(), "B2"))
	trackB21 := files.NewTrack(albumB2, "1 A21a.mp3", "A21a", 1)
	albumB2.AddTrack(trackB21)
	trackB22 := files.NewTrack(albumB2, "2 B22a.mp3", "B22a", 1)
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
		want1           bool
		output.WantedRecording
	}{
		"nothing to filter": {
			ss: cmd.NewSearchSettings().WithArtistFilter(
				regexp.MustCompile(".*")).WithAlbumFilter(
				regexp.MustCompile(".*")).WithTrackFilter(regexp.MustCompile(".*")),
			originalArtists: []*files.Artist{},
			want:            []*files.Artist{},
			want1:           false,
			WantedRecording: output.WantedRecording{
				Error: "No music files remain after filtering.\n" +
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
			ss: cmd.NewSearchSettings().WithArtistFilter(
				regexp.MustCompile("^$")).WithAlbumFilter(
				regexp.MustCompile("^$")).WithTrackFilter(regexp.MustCompile("^$")),
			originalArtists: []*files.Artist{artist1, artist2, artist3},
			want:            []*files.Artist{},
			want1:           false,
			WantedRecording: output.WantedRecording{
				Error: "No music files remain after filtering.\n" +
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
			ss: cmd.NewSearchSettings().WithArtistFilter(
				regexp.MustCompile("^A")).WithAlbumFilter(
				regexp.MustCompile("^A")).WithTrackFilter(regexp.MustCompile("^A")),
			originalArtists: []*files.Artist{artist1, artist2, artist3},
			want:            []*files.Artist{filteredArtist1},
			want1:           true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			got, got1 := tt.ss.Filter(o, tt.originalArtists)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SearchSettings.Filter() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("SearchSettings.Filter() got1 = %v, want %v", got1, tt.want1)
			}
			if differences, ok := o.Verify(tt.WantedRecording); !ok {
				for _, difference := range differences {
					t.Errorf("SearchSettings.Filter() %s", difference)
				}
			}
		})
	}
}
