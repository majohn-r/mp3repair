package commands

import (
	"flag"
	"mp3/internal"
	"os"
	"path/filepath"
	"testing"

	"github.com/majohn-r/output"
)

var (
	fExportFlag = false
	tExportFlag = true
)

func Test_createFile(t *testing.T) {
	fnName := "createFile()"
	testDir := "createFile"
	if err := internal.Mkdir(testDir); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, testDir, err)
	}
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, testDir)
	}()
	unwritable := filepath.Join(testDir, "unwritable.txt")
	if err := internal.Mkdir(unwritable); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, unwritable, err)
	}
	type args struct {
		fileName string
		content  []byte
	}
	tests := []struct {
		name string
		args
		want bool
		output.WantedRecording
	}{
		{
			name: "cannot be written",
			args: args{
				fileName: unwritable,
				content:  []byte{1, 2, 3, 4, 5},
			},
			WantedRecording: output.WantedRecording{
				Error: "The file \"createFile\\\\unwritable.txt\" cannot be created: open createFile\\unwritable.txt: is a directory.\n",
				Log:   "level='error' command='export' error='open createFile\\unwritable.txt: is a directory' fileName='createFile\\unwritable.txt' msg='cannot create file'\n",
			},
		},
		{
			name: "can be written",
			args: args{
				fileName: filepath.Join(testDir, "write this"),
				content:  []byte{1, 2, 3, 4, 5},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := output.NewRecorder()
			if got := createFile(o, tt.args.fileName, tt.args.content); got != tt.want {
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

func Test_export_overwriteFile(t *testing.T) {
	fnName := "export.overwriteFile()"
	testDir := "overwriteFile"
	if err := internal.Mkdir(testDir); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, testDir, err)
	}
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, testDir)
	}()
	existingFile1 := filepath.Join(testDir, "existing1")
	if err := internal.CreateFileForTestingWithContent(testDir, "existing1", []byte{0, 1, 2, 3}); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, existingFile1, err)
	}
	existingFile1Backup := filepath.Join(testDir, "existing1-backup")
	if err := internal.Mkdir(existingFile1Backup); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, existingFile1Backup, err)
	}
	existingFile2 := filepath.Join(testDir, "existing2")
	if err := internal.CreateFileForTestingWithContent(testDir, "existing2", []byte{0, 1, 2, 3}); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, existingFile2, err)
	}
	if err := internal.CreateFileForTestingWithContent(testDir, "existing2-backup", []byte{0, 1, 2, 3, 4}); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, existingFile2+"-backup", err)
	}
	type args struct {
		fileName string
		content  []byte
	}
	tests := []struct {
		name string
		ex   *export
		args
		wantOk bool
		output.WantedRecording
	}{
		{
			name: "not permitted",
			ex:   &export{overwrite: &fExportFlag},
			args: args{fileName: "foo"},
			WantedRecording: output.WantedRecording{
				Error: "The file \"foo\" exists; set the overwrite flag to true if you want it overwritten.\n",
				Log:   "level='error' -overwrite='false' fileName='foo' msg='overwrite is not permitted'\n",
			},
		},
		{
			name: "permitted but cannot backup existing file",
			ex:   &export{overwrite: &tExportFlag},
			args: args{
				fileName: existingFile1,
			},
			WantedRecording: output.WantedRecording{
				Error: "The file \"overwriteFile\\\\existing1\" cannot be renamed to \"overwriteFile\\\\existing1-backup\": rename overwriteFile\\existing1 overwriteFile\\existing1-backup: Access is denied.\n",
				Log:   "level='error' backup='overwriteFile\\existing1-backup' error='rename overwriteFile\\existing1 overwriteFile\\existing1-backup: Access is denied.' original='overwriteFile\\existing1' msg='rename failed'\n",
			},
		},
		{
			name: "permitted and successful",
			ex:   &export{overwrite: &tExportFlag},
			args: args{
				fileName: existingFile2,
				content:  []byte{2, 3, 4, 5, 6},
			},
			wantOk: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := output.NewRecorder()
			if gotOk := tt.ex.overwriteFile(o, tt.args.fileName, tt.args.content); gotOk != tt.wantOk {
				t.Errorf("%s = %v, want %v", fnName, gotOk, tt.wantOk)
			}
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
		})
	}
}

func Test_export_writeDefaults(t *testing.T) {
	fnName := "export.writeDefaults()"
	savedAppData := internal.SaveEnvVarForTesting("APPDATA")
	testDir := "writeDefaults"
	if err := internal.Mkdir(testDir); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, testDir, err)
	}
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, testDir)
		savedAppData.RestoreForTesting()
	}()
	testDir2 := filepath.Join(testDir, "2")
	if err := internal.Mkdir(testDir2); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, testDir2, err)
	}
	occupiedMp3Path := filepath.Join(testDir2, "mp3")
	if err := internal.Mkdir(occupiedMp3Path); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, occupiedMp3Path, err)
	}
	if err := internal.CreateFileForTesting(occupiedMp3Path, internal.DefaultConfigFileName); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, filepath.Join(occupiedMp3Path, internal.DefaultConfigFileName), err)
	}
	type args struct {
		content []byte
	}
	tests := []struct {
		name         string
		ex           *export
		appDataValue *internal.SavedEnvVar
		args
		wantOk bool
		output.WantedRecording
	}{
		{
			name:         "no location",
			appDataValue: &internal.SavedEnvVar{Name: "APPDATA"},
			WantedRecording: output.WantedRecording{
				Log: "level='info' environmentVariable='APPDATA' msg='not set'\n",
			},
		},
		{
			name:         "no valid location",
			appDataValue: &internal.SavedEnvVar{Name: "APPDATA", Value: "no such dir", Set: true},
			WantedRecording: output.WantedRecording{
				Error: "The directory \"no such dir\\\\mp3\" cannot be created: mkdir no such dir\\mp3: The system cannot find the path specified.\n",
				Log:   "level='error' command='export' directory='no such dir\\mp3' error='mkdir no such dir\\mp3: The system cannot find the path specified.' msg='cannot create directory'\n",
			},
		},
		{
			name:         "valid location, not pre-existing",
			appDataValue: &internal.SavedEnvVar{Name: "APPDATA", Value: testDir, Set: true},
			args:         args{content: []byte{1, 2, 3, 4}},
			wantOk:       true,
		},
		{
			name:         "valid location, pre-existing",
			appDataValue: &internal.SavedEnvVar{Name: "APPDATA", Value: testDir2, Set: true},
			ex:           &export{overwrite: &tExportFlag},
			args:         args{content: []byte{1, 2, 3, 4}},
			wantOk:       true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.appDataValue.RestoreForTesting()
			o := output.NewRecorder()
			if gotOk := tt.ex.writeDefaults(o, tt.args.content); gotOk != tt.wantOk {
				t.Errorf("%s = %v, want %v", fnName, gotOk, tt.wantOk)
			}
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
		})
	}
}

func Test_export_exportDefaults(t *testing.T) {
	const fnName = "export.exportDefaults()"
	savedAppData := internal.SaveEnvVarForTesting("APPDATA")
	testDir := "exportDefaults"
	if err := internal.Mkdir(testDir); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, testDir, err)
	}
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, testDir)
		savedAppData.RestoreForTesting()
	}()
	os.Setenv("APPDATA", testDir)
	tests := []struct {
		name string
		ex   *export
		want bool
		output.WantedRecording
	}{
		{
			name: "without permission",
			ex:   &export{defaults: &fExportFlag},
			want: true,
		},
		{
			name: "with permission",
			ex:   &export{defaults: &tExportFlag},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := output.NewRecorder()
			if got := tt.ex.exportDefaults(o); got != tt.want {
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

func Test_getDefaultsContent(t *testing.T) {
	const fnName = "getDefaultsContent()"
	tests := []struct {
		name string
		want string
	}{
		{
			name: "usual test case",
			want: "check:\n" +
				"    empty: false\n" +
				"    gaps: false\n" +
				"    integrity: true\n" +
				"command:\n" +
				"    default: list\n" +
				"common:\n" +
				"    albumFilter: .*\n" +
				"    artistFilter: .*\n" +
				"    ext: .mp3\n" +
				"    topDir: '%HOMEPATH%\\Music'\n" +
				"export:\n" +
				"    defaults: false\n" +
				"    overwrite: false\n" +
				"list:\n" +
				"    annotate: false\n" +
				"    details: false\n" +
				"    diagnostic: false\n" +
				"    includeAlbums: true\n" +
				"    includeArtists: true\n" +
				"    includeTracks: false\n" +
				"    sort: numeric\n" +
				"repair:\n" +
				"    dryRun: false\n" +
				"resetDatabase:\n" +
				"    extension: .wmdb\n" +
				"    metadata: '%USERPROFILE%\\AppData\\Local\\Microsoft\\Media Player'\n" +
				"    service: WMPNetworkSVC\n" +
				"    timeout: 10\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := string(getDefaultsContent()); got != tt.want {
				t.Errorf("%s = %q, want %q", fnName, got, tt.want)
			}
		})
	}
}

func makeExportForTesting() *export {
	e, _ := newExportCommand(output.NewNilBus(), internal.EmptyConfiguration(), flag.NewFlagSet(exportCommandName, flag.ContinueOnError))
	return e
}

func Test_export_Exec(t *testing.T) {
	const fnName = "export.Exec()"
	savedAppData := internal.SaveEnvVarForTesting("APPDATA")
	defer func() {
		savedAppData.RestoreForTesting()
	}()
	os.Unsetenv("APPDATA") // make writes impossible
	type args struct {
		args []string
	}
	tests := []struct {
		name string
		ex   *export
		args
		wantOk bool
		output.WantedRecording
	}{
		{
			name:   "need help",
			ex:     makeExportForTesting(),
			args:   args{args: []string{"--help"}},
			wantOk: false,
			WantedRecording: output.WantedRecording{
				Error: "Usage of export:\n" +
					"  -defaults\n" +
					"    \twrite defaults.yaml (default false)\n" +
					"  -overwrite\n" +
					"    \toverwrite file if it exists (default false)\n",
				Log: "level='error' arguments='[--help]' msg='flag: help requested'\n",
			},
		},
		{
			name:   "nothing to do",
			ex:     makeExportForTesting(),
			args:   args{args: []string{}},
			wantOk: false,
			WantedRecording: output.WantedRecording{
				Error: "You disabled all functionality for the command \"export\".\n",
				Log:   "level='error' -defaults='false' -overwrite='false' command='export' msg='the user disabled all functionality'\n",
			},
		},
		{
			name:   "try to do something, but fail",
			ex:     makeExportForTesting(),
			args:   args{args: []string{"-defaults"}},
			wantOk: false,
			WantedRecording: output.WantedRecording{
				Log: "level='info' environmentVariable='APPDATA' msg='not set'\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := output.NewRecorder()
			if gotOk := tt.ex.Exec(o, tt.args.args); gotOk != tt.wantOk {
				t.Errorf("%s = %v, want %v", fnName, gotOk, tt.wantOk)
			}
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
		})
	}
}

func Test_newExportCommand(t *testing.T) {
	const fnName = "newExportCommand()"
	type args struct {
		c    *internal.Configuration
		fSet *flag.FlagSet
	}
	tests := []struct {
		name string
		args
		want  *export
		want1 bool
		output.WantedRecording
	}{
		{
			name:  "normal",
			args:  args{c: internal.EmptyConfiguration(), fSet: flag.NewFlagSet(exportCommandName, flag.ContinueOnError)},
			want:  &export{},
			want1: true,
		},
		{
			name: "abnormal",
			args: args{
				c: internal.CreateConfiguration(output.NewNilBus(), map[string]any{
					exportCommandName: map[string]any{
						defaultsFlag:  "Beats me",
						overwriteFlag: 12,
					},
				}),
				fSet: flag.NewFlagSet(exportCommandName, flag.ContinueOnError),
			},
			want:  nil,
			want1: false,
			WantedRecording: output.WantedRecording{
				Error: "The configuration file \"defaults.yaml\" contains an invalid value for \"export\": invalid boolean value \"Beats me\" for -defaults: parse error.\n" +
					"The configuration file \"defaults.yaml\" contains an invalid value for \"export\": invalid boolean value \"12\" for -overwrite: parse error.\n",
				Log: "level='error' error='invalid boolean value \"Beats me\" for -defaults: parse error' section='export' msg='invalid content in configuration file'\n" +
					"level='error' error='invalid boolean value \"12\" for -overwrite: parse error' section='export' msg='invalid content in configuration file'\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := output.NewRecorder()
			got, got1 := newExportCommand(o, tt.args.c, tt.args.fSet)
			if got == nil && tt.want != nil {
				t.Errorf("%s got = %v, want %v", fnName, got, tt.want)
			} else if got != nil && tt.want == nil {
				t.Errorf("%s got = %v, want %v", fnName, got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("%s got1 = %v, want %v", fnName, got1, tt.want1)
			}
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
		})
	}
}
