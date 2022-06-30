package commands

import (
	"flag"
	"mp3/internal"
	"mp3/internal/files"
	"path/filepath"
	"testing"
)

func Test_postrepair_Exec(t *testing.T) {
	fnName := "postrepair.Exec()"
	topDirName := "postRepairExec"
	topDir2Name := "postRepairExec2"
	if err := internal.Mkdir(topDirName); err != nil {
		t.Errorf("%s error creating directory %q", fnName, topDirName)
	}
	if err := internal.Mkdir(topDir2Name); err != nil {
		t.Errorf("%s error creating directory %q", fnName, topDir2Name)
	}
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, topDirName)
		internal.DestroyDirectoryForTesting(fnName, topDir2Name)
	}()
	if err := internal.PopulateTopDirForTesting(topDirName); err != nil {
		t.Errorf("%s error populating directory %q", fnName, topDirName)
	}
	artistDir := "the artist"
	artistPath := filepath.Join(topDir2Name, artistDir)
	if err := internal.Mkdir(artistPath); err != nil {
		t.Errorf("%s error creating directory %q", fnName, artistPath)
	}
	artist := files.NewArtist(artistDir, artistPath)
	albumDir := "the album"
	albumPath := filepath.Join(artistPath, albumDir)
	if err := internal.Mkdir(albumPath); err != nil {
		t.Errorf("%s error creating directory %q", fnName, albumPath)
	}
	album := files.NewAlbum(albumDir, artist, albumPath)
	if err := internal.CreateFileForTesting(albumPath, "01 the track.mp3"); err != nil {
		t.Errorf("%s error creating file in album directory %q", fnName, "01 the track.mp3")
	}
	backupDirectory := album.BackupDirectory()
	if err := internal.Mkdir(backupDirectory); err != nil {
		t.Errorf("%s error creating directory %q", fnName, backupDirectory)
	}
	if err := internal.CreateFileForTesting(backupDirectory, "1.mp3"); err != nil {
		t.Errorf("%s error creating file in backup directory %q", fnName, "1.mp3")
	}
	type args struct {
		args []string
	}
	tests := []struct {
		name              string
		p                 *postrepair
		args              args
		wantConsoleOutput string
		wantErrorOutput   string
		wantLogOutput     string
	}{
		{
			name: "handle bad common arguments",
			p: newPostRepairCommand(
				internal.EmptyConfiguration(), flag.NewFlagSet("postRepair", flag.ContinueOnError)),
			args: args{args: []string{"-topDir", "non-existent directory"}},
		},
		{
			name: "handle normal processing with nothing to do",
			p: newPostRepairCommand(
				internal.EmptyConfiguration(), flag.NewFlagSet("postRepair", flag.ContinueOnError)),
			args:              args{args: []string{"-topDir", topDirName}},
			wantConsoleOutput: "There are no backup directories to delete\n",
			wantLogOutput:     "level='info' command='postRepair' msg='executing command'\n",
		},
		{
			name: "handle normal processing",
			p: newPostRepairCommand(
				internal.EmptyConfiguration(), flag.NewFlagSet("postRepair", flag.ContinueOnError)),
			args:              args{args: []string{"-topDir", topDir2Name}},
			wantConsoleOutput: "The backup directory for artist \"the artist\" album \"the album\" has been deleted\n",
			wantLogOutput:     "level='info' command='postRepair' msg='executing command'\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := internal.NewOutputDeviceForTesting()
			tt.p.Exec(o, tt.args.args)
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

func Test_removeBackupDirectory(t *testing.T) {
	fnName := "removeBackupDirectory()"
	topDirName := "removeBackup"
	if err := internal.Mkdir(topDirName); err != nil {
		t.Errorf("%s error creating directory %q", fnName, topDirName)
	}
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, topDirName)
	}()
	artistDir := "the artist"
	artistPath := filepath.Join(topDirName, artistDir)
	if err := internal.Mkdir(artistPath); err != nil {
		t.Errorf("%s error creating directory %q", fnName, artistPath)
	}
	artist := files.NewArtist(artistDir, artistPath)
	albumDir := "the album"
	albumPath := filepath.Join(artistPath, albumDir)
	if err := internal.Mkdir(albumPath); err != nil {
		t.Errorf("%s error creating directory %q", fnName, albumPath)
	}
	album := files.NewAlbum(albumDir, artist, albumPath)
	backupDirectory := album.BackupDirectory()
	if err := internal.Mkdir(backupDirectory); err != nil {
		t.Errorf("%s error creating directory %q", fnName, backupDirectory)
	}
	if err := internal.CreateFileForTesting(backupDirectory, "1.mp3"); err != nil {
		t.Errorf("%s error creating file in backup directory %q", fnName, "1.mp3")
	}
	type args struct {
		d string
		a *files.Album
	}
	tests := []struct {
		name              string
		args              args
		wantConsoleOutput string
		wantErrorOutput   string
		wantLogOutput     string
	}{
		{
			name:            "error case",
			args:            args{d: "dir/.", a: nil},
			wantErrorOutput: "The directory \"dir/.\" cannot be deleted: RemoveAll dir/.: invalid argument.\n",
			wantLogOutput:   "level='warn' directory='dir/.' error='RemoveAll dir/.: invalid argument' msg='cannot delete directory'\n",
		},
		{
			name:              "normal case",
			args:              args{d: backupDirectory, a: album},
			wantConsoleOutput: "The backup directory for artist \"the artist\" album \"the album\" has been deleted\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := internal.NewOutputDeviceForTesting()
			removeBackupDirectory(o, tt.args.d, tt.args.a)
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
