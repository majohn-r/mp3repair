package commands

import (
	"flag"
	"mp3/internal"
	"path/filepath"
	"testing"

	"github.com/majohn-r/output"
)

func Test_createFile(t *testing.T) {
	const fnName = "createFile()"
	testDir := "createFile"
	if err := internal.Mkdir(testDir); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, testDir, err)
	}
	unwritable := filepath.Join(testDir, "unwritable.txt")
	if err := internal.Mkdir(unwritable); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, unwritable, err)
	}
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, testDir)
	}()
	type args struct {
		f string
		b []byte
	}
	tests := map[string]struct {
		args
		want bool
		output.WantedRecording
	}{
		"cannot be written": {
			args: args{
				f: unwritable,
				b: []byte{1, 2, 3, 4, 5},
			},
			WantedRecording: output.WantedRecording{
				Error: "The file \"createFile\\\\unwritable.txt\" cannot be created: open createFile\\unwritable.txt: is a directory.\n",
				Log:   "level='error' command='export' error='open createFile\\unwritable.txt: is a directory' fileName='createFile\\unwritable.txt' msg='cannot create file'\n",
			},
		},
		"can be written": {
			args: args{
				f: filepath.Join(testDir, "write this"),
				b: []byte{1, 2, 3, 4, 5},
			},
			want: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			if got := createFile(o, tt.args.f, tt.args.b); got != tt.want {
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
	const fnName = "export.overwriteFile()"
	fFlag := false
	tFlag := true
	testDir := "overwriteFile"
	if err := internal.Mkdir(testDir); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, testDir, err)
	}
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
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, testDir)
	}()
	type args struct {
		f string
		b []byte
	}
	tests := map[string]struct {
		ex *export
		args
		want bool
		output.WantedRecording
	}{
		"not permitted": {
			ex:   &export{overwrite: &fFlag},
			args: args{f: "foo"},
			WantedRecording: output.WantedRecording{
				Error: "The file \"foo\" exists; set the overwrite flag to true if you want it overwritten.\n",
				Log:   "level='error' -overwrite='false' fileName='foo' msg='overwrite is not permitted'\n",
			},
		},
		"permitted but cannot backup existing file": {
			ex: &export{overwrite: &tFlag},
			args: args{
				f: existingFile1,
			},
			WantedRecording: output.WantedRecording{
				Error: "The file \"overwriteFile\\\\existing1\" cannot be renamed to \"overwriteFile\\\\existing1-backup\": rename overwriteFile\\existing1 overwriteFile\\existing1-backup: Access is denied.\n",
				Log:   "level='error' error='rename overwriteFile\\existing1 overwriteFile\\existing1-backup: Access is denied.' new='overwriteFile\\existing1-backup' old='overwriteFile\\existing1' msg='rename failed'\n",
			},
		},
		"permitted and successful": {
			ex: &export{overwrite: &tFlag},
			args: args{
				f: existingFile2,
				b: []byte{2, 3, 4, 5, 6},
			},
			want: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			if gotOk := tt.ex.overwriteFile(o, tt.args.f, tt.args.b); gotOk != tt.want {
				t.Errorf("%s = %v, want %v", fnName, gotOk, tt.want)
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
	const fnName = "export.writeDefaults()"
	oldAppPath := internal.ApplicationPath()
	testDir := "writeDefaults"
	if err := internal.Mkdir(testDir); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, testDir, err)
	}
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
	tFlag := true
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, testDir)
		internal.SetApplicationPathForTesting(oldAppPath)
	}()
	type args struct {
		b []byte
	}
	tests := map[string]struct {
		ex      *export
		appPath string
		args
		wantOk bool
		output.WantedRecording
	}{
		"valid location, not pre-existing": {
			appPath: testDir,
			args:    args{b: []byte{1, 2, 3, 4}},
			wantOk:  true,
		},
		"valid location, pre-existing": {
			appPath: occupiedMp3Path,
			ex:      &export{overwrite: &tFlag},
			args:    args{b: []byte{1, 2, 3, 4}},
			wantOk:  true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			internal.SetApplicationPathForTesting(tt.appPath)
			o := output.NewRecorder()
			if gotOk := tt.ex.writeDefaults(o, tt.args.b); gotOk != tt.wantOk {
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
	fFlag := false
	tFlag := true
	testDir := "exportDefaults"
	if err := internal.Mkdir(testDir); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, testDir, err)
	}
	oldAppPath := internal.SetApplicationPathForTesting(testDir)
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, testDir)
		internal.SetApplicationPathForTesting(oldAppPath)
	}()
	tests := map[string]struct {
		ex   *export
		want bool
		output.WantedRecording
	}{
		"without permission": {
			ex:   &export{defaults: &fFlag},
			want: true,
		},
		"with permission": {
			ex:   &export{defaults: &tFlag},
			want: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
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

func Test_defaultsContent(t *testing.T) {
	const fnName = "defaultsContent()"
	tests := map[string]struct {
		want string
	}{
		"usual test case": {
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
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := string(defaultsContent()); got != tt.want {
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
	testAppPath := "appPath"
	if err := internal.Mkdir(testAppPath); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, testAppPath, err)
	}
	oldAppPath := internal.SetApplicationPathForTesting(testAppPath)
	defer func() {
		internal.SetApplicationPathForTesting(oldAppPath)
		internal.DestroyDirectoryForTesting(fnName, testAppPath)
	}()
	type args struct {
		args []string
	}
	tests := map[string]struct {
		ex *export
		args
		want bool
		output.WantedRecording
	}{
		"need help": {
			ex:   makeExportForTesting(),
			args: args{args: []string{"--help"}},
			want: false,
			WantedRecording: output.WantedRecording{
				Error: "Usage of export:\n" +
					"  -defaults\n" +
					"    \twrite defaults.yaml (default false)\n" +
					"  -overwrite\n" +
					"    \toverwrite file if it exists (default false)\n",
				Log: "level='error' arguments='[--help]' msg='flag: help requested'\n",
			},
		},
		"nothing to do": {
			ex:   makeExportForTesting(),
			args: args{args: []string{}},
			want: false,
			WantedRecording: output.WantedRecording{
				Error: "You disabled all functionality for the command \"export\".\n",
				Log:   "level='error' -defaults='false' -overwrite='false' command='export' msg='the user disabled all functionality'\n",
			},
		},
		"work to do": {
			ex:   makeExportForTesting(),
			args: args{args: []string{"-defaults=true", "-overwrite=true"}},
			want: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			if gotOk := tt.ex.Exec(o, tt.args.args); gotOk != tt.want {
				t.Errorf("%s = %v, want %v", fnName, gotOk, tt.want)
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
	tests := map[string]struct {
		args
		want   *export
		wantOk bool
		output.WantedRecording
	}{
		"normal": {
			args:   args{c: internal.EmptyConfiguration(), fSet: flag.NewFlagSet(exportCommandName, flag.ContinueOnError)},
			want:   &export{},
			wantOk: true,
		},
		"abnormal": {
			args: args{
				c: internal.CreateConfiguration(output.NewNilBus(), map[string]any{
					exportCommandName: map[string]any{
						defaultsFlag:  "Beats me",
						overwriteFlag: 12,
					},
				}),
				fSet: flag.NewFlagSet(exportCommandName, flag.ContinueOnError),
			},
			want:   nil,
			wantOk: false,
			WantedRecording: output.WantedRecording{
				Error: "The configuration file \"defaults.yaml\" contains an invalid value for \"export\": invalid boolean value \"Beats me\" for -defaults: parse error.\n" +
					"The configuration file \"defaults.yaml\" contains an invalid value for \"export\": invalid boolean value \"12\" for -overwrite: parse error.\n",
				Log: "level='error' error='invalid boolean value \"Beats me\" for -defaults: parse error' section='export' msg='invalid content in configuration file'\n" +
					"level='error' error='invalid boolean value \"12\" for -overwrite: parse error' section='export' msg='invalid content in configuration file'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			got, gotOk := newExportCommand(o, tt.args.c, tt.args.fSet)
			if got == nil && tt.want != nil {
				t.Errorf("%s got = %v, want %v", fnName, got, tt.want)
			} else if got != nil && tt.want == nil {
				t.Errorf("%s got = %v, want %v", fnName, got, tt.want)
			}
			if gotOk != tt.wantOk {
				t.Errorf("%s got1 = %v, want %v", fnName, gotOk, tt.wantOk)
			}
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
		})
	}
}
