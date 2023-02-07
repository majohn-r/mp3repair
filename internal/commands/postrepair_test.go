package commands

import (
	"flag"
	"mp3/internal"
	"mp3/internal/files"
	"os"
	"path/filepath"
	"testing"

	tools "github.com/majohn-r/cmd-toolkit"
	"github.com/majohn-r/output"
)

func makePostRepairCommandForTesting() *postrepair {
	pr, _ := newPostRepairCommand(output.NewNilBus(), tools.EmptyConfiguration(), flag.NewFlagSet("postRepair", flag.ContinueOnError))
	return pr
}

func Test_postrepair_Exec(t *testing.T) {
	const fnName = "postrepair.Exec()"
	topDirName := "postRepairExec"
	topDir2Name := "postRepairExec2"
	if err := tools.Mkdir(topDirName); err != nil {
		t.Errorf("%s error creating directory %q: %v", fnName, topDirName, err)
	}
	if err := tools.Mkdir(topDir2Name); err != nil {
		t.Errorf("%s error creating directory %q: %v", fnName, topDir2Name, err)
	}
	savedHome := tools.NewEnvVarMemento("HOMEPATH")
	os.Setenv("HOMEPATH", "C:\\Users\\The User")
	if err := internal.PopulateTopDirForTesting(topDirName); err != nil {
		t.Errorf("%s error populating directory %q: %v", fnName, topDirName, err)
	}
	artistDir := "the artist"
	artistPath := filepath.Join(topDir2Name, artistDir)
	if err := tools.Mkdir(artistPath); err != nil {
		t.Errorf("%s error creating directory %q: %v", fnName, artistPath, err)
	}
	artist := files.NewArtist(artistDir, artistPath)
	albumDir := "the album"
	albumPath := filepath.Join(artistPath, albumDir)
	if err := tools.Mkdir(albumPath); err != nil {
		t.Errorf("%s error creating directory %q: %v", fnName, albumPath, err)
	}
	album := files.NewAlbum(albumDir, artist, albumPath)
	if err := internal.CreateFileForTesting(albumPath, "01 the track.mp3"); err != nil {
		t.Errorf("%s error creating file in album directory %q: %v", fnName, "01 the track.mp3", err)
	}
	backupDirectory := album.BackupDirectory()
	if err := tools.Mkdir(backupDirectory); err != nil {
		t.Errorf("%s error creating directory %q: %v", fnName, backupDirectory, err)
	}
	if err := internal.CreateFileForTesting(backupDirectory, "1.mp3"); err != nil {
		t.Errorf("%s error creating file in backup directory %q: %v", fnName, "1.mp3", err)
	}
	defer func() {
		savedHome.Restore()
		internal.DestroyDirectoryForTesting(fnName, topDirName)
		internal.DestroyDirectoryForTesting(fnName, topDir2Name)
	}()
	type args struct {
		args []string
	}
	tests := map[string]struct {
		p *postrepair
		args
		output.WantedRecording
	}{
		"help": {
			p:    makePostRepairCommandForTesting(),
			args: args{args: []string{"--help"}},
			WantedRecording: output.WantedRecording{
				Error: "Usage of postRepair:\n" +
					"  -albumFilter regular expression\n" +
					"    \tregular expression specifying which albums to select (default \".*\")\n" +
					"  -artistFilter regular expression\n" +
					"    \tregular expression specifying which artists to select (default \".*\")\n" +
					"  -ext extension\n" +
					"    \textension identifying music files (default \".mp3\")\n" +
					"  -topDir directory\n" +
					"    \ttop directory specifying where to find music files (default \"C:\\\\Users\\\\The User\\\\Music\")\n",
				Log: "level='error' arguments='[--help]' msg='flag: help requested'\n",
			},
		},
		"handle bad common arguments": {
			p:    makePostRepairCommandForTesting(),
			args: args{args: []string{"-topDir", "non-existent directory"}},
			WantedRecording: output.WantedRecording{
				Error: "The -topDir value you specified, \"non-existent directory\", cannot be read: CreateFile non-existent directory: The system cannot find the file specified.\n",
				Log:   "level='error' directory='non-existent directory' error='CreateFile non-existent directory: The system cannot find the file specified.' msg='cannot read directory'\n",
			},
		},
		"handle normal processing with nothing to do": {
			p:    makePostRepairCommandForTesting(),
			args: args{args: []string{"-topDir", topDirName}},
			WantedRecording: output.WantedRecording{
				Console: "There are no backup directories to delete.\n",
				Log: "level='info' command='postRepair' msg='executing command'\n" +
					"level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='postRepairExec' msg='reading filtered music files'\n",
			},
		},
		"handle normal processing": {
			p:    makePostRepairCommandForTesting(),
			args: args{args: []string{"-topDir", topDir2Name}},
			WantedRecording: output.WantedRecording{
				Console: "The backup directory for artist \"the artist\" album \"the album\" has been deleted.\n",
				Log: "level='info' command='postRepair' msg='executing command'\n" +
					"level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='postRepairExec2' msg='reading filtered music files'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			tt.p.Exec(o, tt.args.args)
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
		})
	}
}

func Test_removeBackupDirectory(t *testing.T) {
	const fnName = "removeBackupDirectory()"
	topDirName := "removeBackup"
	if err := tools.Mkdir(topDirName); err != nil {
		t.Errorf("%s error creating directory %q: %v", fnName, topDirName, err)
	}
	artistDir := "the artist"
	artistPath := filepath.Join(topDirName, artistDir)
	if err := tools.Mkdir(artistPath); err != nil {
		t.Errorf("%s error creating directory %q: %v", fnName, artistPath, err)
	}
	artist := files.NewArtist(artistDir, artistPath)
	albumDir := "the album"
	albumPath := filepath.Join(artistPath, albumDir)
	if err := tools.Mkdir(albumPath); err != nil {
		t.Errorf("%s error creating directory %q: %v", fnName, albumPath, err)
	}
	album := files.NewAlbum(albumDir, artist, albumPath)
	backupDirectory := album.BackupDirectory()
	if err := tools.Mkdir(backupDirectory); err != nil {
		t.Errorf("%s error creating directory %q: %v", fnName, backupDirectory, err)
	}
	if err := internal.CreateFileForTesting(backupDirectory, "1.mp3"); err != nil {
		t.Errorf("%s error creating file in backup directory %q: %v", fnName, "1.mp3", err)
	}
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, topDirName)
	}()
	type args struct {
		dir string
		a   *files.Album
	}
	tests := map[string]struct {
		args
		output.WantedRecording
	}{
		"error case": {
			args: args{dir: "dir/.", a: nil},
			WantedRecording: output.WantedRecording{
				Error: "The directory \"dir/.\" cannot be deleted: RemoveAll dir/.: invalid argument.\n",
				Log:   "level='error' directory='dir/.' error='RemoveAll dir/.: invalid argument' msg='cannot delete directory'\n",
			},
		},
		"normal case": {
			args: args{dir: backupDirectory, a: album},
			WantedRecording: output.WantedRecording{
				Console: "The backup directory for artist \"the artist\" album \"the album\" has been deleted.\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			removeBackupDirectory(o, tt.args.dir, tt.args.a)
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
		})
	}
}

func Test_newPostRepairCommand(t *testing.T) {
	const fnName = "newPostRepairCommand()"
	savedFoo := tools.NewEnvVarMemento("FOO")
	os.Unsetenv("FOO")
	defer func() {
		savedFoo.Restore()
	}()
	type args struct {
		c    *tools.Configuration
		fSet *flag.FlagSet
	}
	tests := map[string]struct {
		args
		wantPostRepair bool
		wantOk         bool
		output.WantedRecording
	}{
		"success": {
			args: args{
				c:    tools.EmptyConfiguration(),
				fSet: flag.NewFlagSet("postRepair", flag.ContinueOnError),
			},
			wantPostRepair: true,
			wantOk:         true,
		},
		"failure": {
			args: args{
				c: tools.NewConfiguration(output.NewNilBus(), map[string]any{
					"common": map[string]any{
						"topDir": "%FOO%",
					},
				}),
				fSet: flag.NewFlagSet("postRepair", flag.ContinueOnError),
			},
			WantedRecording: output.WantedRecording{
				Error: "The configuration file \"defaults.yaml\" contains an invalid value for \"common\": invalid value \"%FOO%\" for flag -topDir: missing environment variables: [FOO].\n",
				Log:   "level='error' error='invalid value \"%FOO%\" for flag -topDir: missing environment variables: [FOO]' section='common' msg='invalid content in configuration file'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			got, gotOk := newPostRepairCommand(o, tt.args.c, tt.args.fSet)
			if (got != nil) != tt.wantPostRepair {
				t.Errorf("%s got = %v, want %v", fnName, got, tt.wantPostRepair)
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

func Test_newPostRepair(t *testing.T) {
	type args struct {
		c    *tools.Configuration
		fSet *flag.FlagSet
	}
	tests := map[string]struct {
		args
		want  tools.CommandProcessor
		want1 bool
		output.WantedRecording
	}{
		"basic": {
			args:  args{c: tools.EmptyConfiguration(), fSet: flag.NewFlagSet(postRepairCommandName, flag.ContinueOnError)},
			want:  makePostRepairCommandForTesting(),
			want1: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			got, got1 := newPostRepair(o, tt.args.c, tt.args.fSet)
			if _, ok := got.(*postrepair); !ok {
				t.Errorf("newPostRepair() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("newPostRepair() got1 = %v, want %v", got1, tt.want1)
			}
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("newPostRepair() %s", issue)
				}
			}
		})
	}
}
