package commands

import (
	"flag"
	"mp3/internal"
	"mp3/internal/files"
	"os"
	"path/filepath"
	"testing"
)

func makePostRepairCommandForTesting() *postrepair {
	pr, _ := newPostRepairCommand(
		internal.NewOutputDeviceForTesting(),
		internal.EmptyConfiguration(),
		flag.NewFlagSet("postRepair", flag.ContinueOnError))
	return pr
}

func Test_postrepair_Exec(t *testing.T) {
	fnName := "postrepair.Exec()"
	topDirName := "postRepairExec"
	topDir2Name := "postRepairExec2"
	if err := internal.Mkdir(topDirName); err != nil {
		t.Errorf("%s error creating directory %q: %v", fnName, topDirName, err)
	}
	if err := internal.Mkdir(topDir2Name); err != nil {
		t.Errorf("%s error creating directory %q: %v", fnName, topDir2Name, err)
	}
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, topDirName)
		internal.DestroyDirectoryForTesting(fnName, topDir2Name)
	}()
	if err := internal.PopulateTopDirForTesting(topDirName); err != nil {
		t.Errorf("%s error populating directory %q: %v", fnName, topDirName, err)
	}
	artistDir := "the artist"
	artistPath := filepath.Join(topDir2Name, artistDir)
	if err := internal.Mkdir(artistPath); err != nil {
		t.Errorf("%s error creating directory %q: %v", fnName, artistPath, err)
	}
	artist := files.NewArtist(artistDir, artistPath)
	albumDir := "the album"
	albumPath := filepath.Join(artistPath, albumDir)
	if err := internal.Mkdir(albumPath); err != nil {
		t.Errorf("%s error creating directory %q: %v", fnName, albumPath, err)
	}
	album := files.NewAlbum(albumDir, artist, albumPath)
	if err := internal.CreateFileForTesting(albumPath, "01 the track.mp3"); err != nil {
		t.Errorf("%s error creating file in album directory %q: %v", fnName, "01 the track.mp3", err)
	}
	backupDirectory := album.BackupDirectory()
	if err := internal.Mkdir(backupDirectory); err != nil {
		t.Errorf("%s error creating directory %q: %v", fnName, backupDirectory, err)
	}
	if err := internal.CreateFileForTesting(backupDirectory, "1.mp3"); err != nil {
		t.Errorf("%s error creating file in backup directory %q: %v", fnName, "1.mp3", err)
	}
	type args struct {
		args []string
	}
	tests := []struct {
		name string
		p    *postrepair
		args args
		internal.WantedOutput
	}{
		{
			name: "handle bad common arguments",
			p:    makePostRepairCommandForTesting(),
			args: args{args: []string{"-topDir", "non-existent directory"}},
			WantedOutput: internal.WantedOutput{
				WantErrorOutput: "The -topDir value you specified, \"non-existent directory\", cannot be read: CreateFile non-existent directory: The system cannot find the file specified.\n",
				WantLogOutput:   "level='warn' -topDir='non-existent directory' error='CreateFile non-existent directory: The system cannot find the file specified.' msg='cannot read directory'\n",
			},
		},
		{
			name: "handle normal processing with nothing to do",
			p:    makePostRepairCommandForTesting(),
			args: args{args: []string{"-topDir", topDirName}},
			WantedOutput: internal.WantedOutput{
				WantConsoleOutput: "There are no backup directories to delete.\n",
				WantLogOutput: "level='info' command='postRepair' msg='executing command'\n" +
					"level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='postRepairExec' msg='reading filtered music files'\n",
			},
		},
		{
			name: "handle normal processing",
			p:    makePostRepairCommandForTesting(),
			args: args{args: []string{"-topDir", topDir2Name}},
			WantedOutput: internal.WantedOutput{
				WantConsoleOutput: "The backup directory for artist \"the artist\" album \"the album\" has been deleted.\n",
				WantLogOutput: "level='info' command='postRepair' msg='executing command'\n" +
					"level='info' -albumFilter='.*' -artistFilter='.*' -ext='.mp3' -topDir='postRepairExec2' msg='reading filtered music files'\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := internal.NewOutputDeviceForTesting()
			tt.p.Exec(o, tt.args.args)
			if issues, ok := o.CheckOutput(tt.WantedOutput); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
		})
	}
}

func Test_removeBackupDirectory(t *testing.T) {
	fnName := "removeBackupDirectory()"
	topDirName := "removeBackup"
	if err := internal.Mkdir(topDirName); err != nil {
		t.Errorf("%s error creating directory %q: %v", fnName, topDirName, err)
	}
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, topDirName)
	}()
	artistDir := "the artist"
	artistPath := filepath.Join(topDirName, artistDir)
	if err := internal.Mkdir(artistPath); err != nil {
		t.Errorf("%s error creating directory %q: %v", fnName, artistPath, err)
	}
	artist := files.NewArtist(artistDir, artistPath)
	albumDir := "the album"
	albumPath := filepath.Join(artistPath, albumDir)
	if err := internal.Mkdir(albumPath); err != nil {
		t.Errorf("%s error creating directory %q: %v", fnName, albumPath, err)
	}
	album := files.NewAlbum(albumDir, artist, albumPath)
	backupDirectory := album.BackupDirectory()
	if err := internal.Mkdir(backupDirectory); err != nil {
		t.Errorf("%s error creating directory %q: %v", fnName, backupDirectory, err)
	}
	if err := internal.CreateFileForTesting(backupDirectory, "1.mp3"); err != nil {
		t.Errorf("%s error creating file in backup directory %q: %v", fnName, "1.mp3", err)
	}
	type args struct {
		d string
		a *files.Album
	}
	tests := []struct {
		name string
		args args
		internal.WantedOutput
	}{
		{
			name: "error case",
			args: args{d: "dir/.", a: nil},
			WantedOutput: internal.WantedOutput{
				WantErrorOutput: "The directory \"dir/.\" cannot be deleted: RemoveAll dir/.: invalid argument.\n",
				WantLogOutput:   "level='warn' directory='dir/.' error='RemoveAll dir/.: invalid argument' msg='cannot delete directory'\n",
			},
		},
		{
			name: "normal case",
			args: args{d: backupDirectory, a: album},
			WantedOutput: internal.WantedOutput{
				WantConsoleOutput: "The backup directory for artist \"the artist\" album \"the album\" has been deleted.\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := internal.NewOutputDeviceForTesting()
			removeBackupDirectory(o, tt.args.d, tt.args.a)
			if issues, ok := o.CheckOutput(tt.WantedOutput); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
		})
	}
}

func Test_newPostRepairCommand(t *testing.T) {
	fnName := "newPostRepairCommand()"
	savedFoo := internal.SaveEnvVarForTesting("FOO")
	os.Unsetenv("FOO")
	defer func() {
		savedFoo.RestoreForTesting()
	}()
	type args struct {
		c    *internal.Configuration
		fSet *flag.FlagSet
	}
	tests := []struct {
		name string
		args
		wantPostRepair bool
		wantOk         bool
		internal.WantedOutput
	}{
		{
			name: "success",
			args: args{
				c:    internal.EmptyConfiguration(),
				fSet: flag.NewFlagSet("postRepair", flag.ContinueOnError),
			},
			wantPostRepair: true,
			wantOk:         true,
		},
		{
			name: "failure",
			args: args{
				c: internal.CreateConfiguration(internal.NewOutputDeviceForTesting(), map[string]interface{}{
					"common": map[string]interface{}{
						"topDir": "%FOO%",
					},
				}),
				fSet: flag.NewFlagSet("postRepair", flag.ContinueOnError),
			},
			WantedOutput: internal.WantedOutput{
				WantErrorOutput: "The configuration file \"defaults.yaml\" contains an invalid value for \"common\": invalid value \"%FOO%\" for flag -topDir: missing environment variables: [FOO].\n",
				WantLogOutput:   "level='warn' error='invalid value \"%FOO%\" for flag -topDir: missing environment variables: [FOO]' section='common' msg='invalid content in configuration file'\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := internal.NewOutputDeviceForTesting()
			got, gotOk := newPostRepairCommand(o, tt.args.c, tt.args.fSet)
			if (got != nil) != tt.wantPostRepair {
				t.Errorf("%s got = %v, want %v", fnName, got, tt.wantPostRepair)
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
