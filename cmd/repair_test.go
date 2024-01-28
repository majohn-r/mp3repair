/*
Copyright © 2021 Marc Johnson (marc.johnson27591@gmail.com)
*/
package cmd_test

import (
	"fmt"
	"mp3/cmd"
	"mp3/internal/files"
	"reflect"
	"regexp"
	"testing"

	"github.com/bogem/id3v2/v2"
	cmd_toolkit "github.com/majohn-r/cmd-toolkit"
	"github.com/majohn-r/output"
	"github.com/spf13/cobra"
)

func TestProcessRepairFlags(t *testing.T) {
	tests := map[string]struct {
		values map[string]*cmd.FlagValue
		want   *cmd.RepairSettings
		want1  bool
		output.WantedRecording
	}{
		"bad value": {
			values: map[string]*cmd.FlagValue{},
			want:   &cmd.RepairSettings{},
			want1:  false,
			WantedRecording: output.WantedRecording{
				Error: "An internal error occurred: flag \"dryRun\" is not found.\n",
				Log: "" +
					"level='error'" +
					" error='flag not found'" +
					" flag='dryRun'" +
					" msg='internal error'\n",
			},
		},
		"good value": {
			values: map[string]*cmd.FlagValue{"dryRun": {Value: true}},
			want:   &cmd.RepairSettings{DryRun: true},
			want1:  true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			got, got1 := cmd.ProcessRepairFlags(o, tt.values)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ProcessRepairFlags() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("ProcessRepairFlags() got1 = %v, want %v", got1, tt.want1)
			}
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("ProcessRepairFlags() %s", issue)
				}
			}
		})
	}
}

func TestEnsureBackupDirectoryExists(t *testing.T) {
	oldDirExists := cmd.DirExists
	oldMkDir := cmd.MkDir
	defer func() {
		cmd.DirExists = oldDirExists
		cmd.MkDir = oldMkDir
	}()
	album := &files.Album{}
	if albums := generateAlbums(1, 5); len(albums) > 0 {
		album = albums[0]
	}
	tests := map[string]struct {
		cAl        *cmd.CheckedAlbum
		dirExists  func(s string) bool
		mkDir      func(s string) error
		wantPath   string
		wantExists bool
		output.WantedRecording
	}{
		"dir already exists": {
			cAl:        cmd.NewCheckedAlbum(album),
			dirExists:  func(_ string) bool { return true },
			wantPath:   album.BackupDirectory(),
			wantExists: true,
		},
		"dir does not exist but can be created": {
			cAl:        cmd.NewCheckedAlbum(album),
			dirExists:  func(_ string) bool { return false },
			mkDir:      func(_ string) error { return nil },
			wantPath:   album.BackupDirectory(),
			wantExists: true,
		},
		"dir does not exist and cannot be created": {
			cAl:        cmd.NewCheckedAlbum(album),
			dirExists:  func(_ string) bool { return false },
			mkDir:      func(_ string) error { return fmt.Errorf("plain file exists") },
			wantPath:   album.BackupDirectory(),
			wantExists: false,
			WantedRecording: output.WantedRecording{
				Error: "" +
					"The directory \"Music\\\\my artist\\\\my album 00\\\\pre-repair-backup\" cannot be created: plain file exists.\n" +
					"The track files in the directory \"Music\\\\my artist\\\\my album 00\" will not be repaired.\n",
				Log: "" +
					"level='error'" +
					" command='repair'" +
					" directory='Music\\my artist\\my album 00\\pre-repair-backup'" +
					" error='plain file exists'" +
					" msg='cannot create directory'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			cmd.DirExists = tt.dirExists
			cmd.MkDir = tt.mkDir
			o := output.NewRecorder()
			gotPath, gotExists := cmd.EnsureBackupDirectoryExists(o, tt.cAl)
			if gotPath != tt.wantPath {
				t.Errorf("EnsureBackupDirectoryExists() gotPath = %v, want %v", gotPath, tt.wantPath)
			}
			if gotExists != tt.wantExists {
				t.Errorf("EnsureBackupDirectoryExists() gotExists = %v, want %v", gotExists, tt.wantExists)
			}
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("EnsureBackupDirectoryExists() %s", issue)
				}
			}
		})
	}
}

func TestAttemptCopy(t *testing.T) {
	oldPlainFileExists := cmd.PlainFileExists
	oldCopyFile := cmd.CopyFile
	defer func() {
		cmd.PlainFileExists = oldPlainFileExists
		cmd.CopyFile = oldCopyFile
	}()
	track := &files.Track{}
	if tracks := generateTracks(1); len(tracks) > 0 {
		track = tracks[0]
	}
	type args struct {
		t    *files.Track
		path string
	}
	tests := map[string]struct {
		plainFileExists func(path string) bool
		copyFile        func(src, dest string) error
		args
		wantBackedUp bool
		output.WantedRecording
	}{
		"backup already exists": {
			plainFileExists: func(_ string) bool { return true },
			args:            args{t: track, path: "backupDir"},
			wantBackedUp:    false,
			WantedRecording: output.WantedRecording{
				Error: "" +
					"The backup file for track file \"Music\\\\my artist\\\\my album 00\\\\1 my track 001.mp3\", \"backupDir\\\\1.mp3\", already exists.\n" +
					"The track file \"Music\\\\my artist\\\\my album 00\\\\1 my track 001.mp3\" will not be repaired.\n",
				Log: "" +
					"level='error'" +
					" command='repair'" +
					" file='backupDir\\1.mp3'" +
					" msg='file already exists'\n",
			},
		},
		"backup does not exist but copy fails": {
			plainFileExists: func(_ string) bool { return false },
			copyFile:        func(_, _ string) error { return fmt.Errorf("dir by that name exists") },
			args:            args{t: track, path: "backupDir"},
			wantBackedUp:    false,
			WantedRecording: output.WantedRecording{
				Error: "" +
					"The track file \"Music\\\\my artist\\\\my album 00\\\\1 my track 001.mp3\" could not be backed up due to error dir by that name exists.\n" +
					"The track file \"Music\\\\my artist\\\\my album 00\\\\1 my track 001.mp3\" will not be repaired.\n",
				Log: "" +
					"level='error'" +
					" command='repair'" +
					" destination='backupDir\\1.mp3' error='dir by that name exists'" +
					" source='Music\\my artist\\my album 00\\1 my track 001.mp3'" +
					" msg='error copying file'\n",
			},
		},
		"successful backup": {
			plainFileExists: func(_ string) bool { return false },
			copyFile:        func(_, _ string) error { return nil },
			args:            args{t: track, path: "backupDir"},
			wantBackedUp:    true,
			WantedRecording: output.WantedRecording{
				Console: "The track file \"Music\\\\my artist\\\\my album 00\\\\1 my track 001.mp3\" has been backed up to \"backupDir\\\\1.mp3\".\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			cmd.PlainFileExists = tt.plainFileExists
			cmd.CopyFile = tt.copyFile
			o := output.NewRecorder()
			if gotBackedUp := cmd.AttemptCopy(o, tt.args.t, tt.args.path); gotBackedUp != tt.wantBackedUp {
				t.Errorf("AttemptCopy() = %v, want %v", gotBackedUp, tt.wantBackedUp)
			}
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("AttemptCopy() %s", issue)
				}
			}
		})
	}
}

func TestProcessUpdateResult(t *testing.T) {
	oldMarkDirty := cmd.MarkDirty
	defer func() {
		cmd.MarkDirty = oldMarkDirty
	}()
	var markedDirty bool
	cmd.MarkDirty = func(o output.Bus) {
		markedDirty = true
	}
	track := &files.Track{}
	if tracks := generateTracks(1); len(tracks) > 0 {
		track = tracks[0]
	}
	type args struct {
		t   *files.Track
		err []error
	}
	tests := map[string]struct {
		args
		wantDirty bool
		output.WantedRecording
	}{
		"success": {
			args:      args{t: track, err: nil},
			wantDirty: true,
			WantedRecording: output.WantedRecording{
				Console: "\"Music\\\\my artist\\\\my album 00\\\\1 my track 001.mp3\" repaired.\n",
			},
		},
		"single failure": {
			args:      args{t: track, err: []error{fmt.Errorf("file locked")}},
			wantDirty: false,
			WantedRecording: output.WantedRecording{
				Error: "An error occurred repairing track \"Music\\\\my artist\\\\my album 00\\\\1 my track 001.mp3\".\n",
				Log: "" +
					"level='error'" +
					" command='repair'" +
					" directory='Music\\my artist\\my album 00'" +
					" error='[\"file locked\"]'" +
					" fileName='1 my track 001.mp3'" +
					" msg='cannot edit track'\n",
			},
		},
		"multiple failures": {
			args: args{
				t:   track,
				err: []error{fmt.Errorf("file locked"), fmt.Errorf("syntax error")},
			},
			wantDirty: false,
			WantedRecording: output.WantedRecording{
				Error: "An error occurred repairing track \"Music\\\\my artist\\\\my album 00\\\\1 my track 001.mp3\".\n",
				Log: "" +
					"level='error'" +
					" command='repair'" +
					" directory='Music\\my artist\\my album 00'" +
					" error='[\"file locked\", \"syntax error\"]'" +
					" fileName='1 my track 001.mp3'" +
					" msg='cannot edit track'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			markedDirty = false
			cmd.ProcessUpdateResult(o, tt.args.t, tt.args.err)
			if got := markedDirty; got != tt.wantDirty {
				t.Errorf("ProcessUpdateResult() got %t want %t", got, tt.wantDirty)
			}
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("ProcessUpdateResult() %s", issue)
				}
			}
		})
	}
}

func TestBackupAndFix(t *testing.T) {
	checkedArtists := cmd.PrepareCheckedArtists(generateArtists(2, 3, 4))
	for _, cAr := range checkedArtists {
		for _, cAl := range cAr.Albums() {
			for _, cT := range cAl.Tracks() {
				cT.AddIssue(cmd.CheckConflictIssue, "artist field does not match artist name")
			}
		}
	}
	oldDirExists := cmd.DirExists
	oldPlainFileExists := cmd.PlainFileExists
	oldCopyFile := cmd.CopyFile
	defer func() {
		cmd.DirExists = oldDirExists
		cmd.PlainFileExists = oldPlainFileExists
		cmd.CopyFile = oldCopyFile
	}()
	cmd.DirExists = func(_ string) bool { return true }
	cmd.PlainFileExists = func(_ string) bool { return false }
	cmd.CopyFile = func(_, _ string) error { return nil }
	tests := map[string]struct {
		checkedArtists []*cmd.CheckedArtist
		output.WantedRecording
	}{
		"basic test": {
			checkedArtists: checkedArtists,
			WantedRecording: output.WantedRecording{
				Console: "" +
					"The track file \"Music\\\\my artist\\\\my album 00\\\\1 my track 001.mp3\" has been backed up to \"Music\\\\my artist\\\\my album 00\\\\pre-repair-backup\\\\1.mp3\".\n" +
					"The track file \"Music\\\\my artist\\\\my album 00\\\\2 my track 002.mp3\" has been backed up to \"Music\\\\my artist\\\\my album 00\\\\pre-repair-backup\\\\2.mp3\".\n" +
					"The track file \"Music\\\\my artist\\\\my album 00\\\\3 my track 003.mp3\" has been backed up to \"Music\\\\my artist\\\\my album 00\\\\pre-repair-backup\\\\3.mp3\".\n" +
					"The track file \"Music\\\\my artist\\\\my album 00\\\\4 my track 004.mp3\" has been backed up to \"Music\\\\my artist\\\\my album 00\\\\pre-repair-backup\\\\4.mp3\".\n" +
					"The track file \"Music\\\\my artist\\\\my album 01\\\\1 my track 011.mp3\" has been backed up to \"Music\\\\my artist\\\\my album 01\\\\pre-repair-backup\\\\1.mp3\".\n" +
					"The track file \"Music\\\\my artist\\\\my album 01\\\\2 my track 012.mp3\" has been backed up to \"Music\\\\my artist\\\\my album 01\\\\pre-repair-backup\\\\2.mp3\".\n" +
					"The track file \"Music\\\\my artist\\\\my album 01\\\\3 my track 013.mp3\" has been backed up to \"Music\\\\my artist\\\\my album 01\\\\pre-repair-backup\\\\3.mp3\".\n" +
					"The track file \"Music\\\\my artist\\\\my album 01\\\\4 my track 014.mp3\" has been backed up to \"Music\\\\my artist\\\\my album 01\\\\pre-repair-backup\\\\4.mp3\".\n" +
					"The track file \"Music\\\\my artist\\\\my album 02\\\\1 my track 021.mp3\" has been backed up to \"Music\\\\my artist\\\\my album 02\\\\pre-repair-backup\\\\1.mp3\".\n" +
					"The track file \"Music\\\\my artist\\\\my album 02\\\\2 my track 022.mp3\" has been backed up to \"Music\\\\my artist\\\\my album 02\\\\pre-repair-backup\\\\2.mp3\".\n" +
					"The track file \"Music\\\\my artist\\\\my album 02\\\\3 my track 023.mp3\" has been backed up to \"Music\\\\my artist\\\\my album 02\\\\pre-repair-backup\\\\3.mp3\".\n" +
					"The track file \"Music\\\\my artist\\\\my album 02\\\\4 my track 024.mp3\" has been backed up to \"Music\\\\my artist\\\\my album 02\\\\pre-repair-backup\\\\4.mp3\".\n" +
					"The track file \"Music\\\\my artist\\\\my album 10\\\\1 my track 101.mp3\" has been backed up to \"Music\\\\my artist\\\\my album 10\\\\pre-repair-backup\\\\1.mp3\".\n" +
					"The track file \"Music\\\\my artist\\\\my album 10\\\\2 my track 102.mp3\" has been backed up to \"Music\\\\my artist\\\\my album 10\\\\pre-repair-backup\\\\2.mp3\".\n" +
					"The track file \"Music\\\\my artist\\\\my album 10\\\\3 my track 103.mp3\" has been backed up to \"Music\\\\my artist\\\\my album 10\\\\pre-repair-backup\\\\3.mp3\".\n" +
					"The track file \"Music\\\\my artist\\\\my album 10\\\\4 my track 104.mp3\" has been backed up to \"Music\\\\my artist\\\\my album 10\\\\pre-repair-backup\\\\4.mp3\".\n" +
					"The track file \"Music\\\\my artist\\\\my album 11\\\\1 my track 111.mp3\" has been backed up to \"Music\\\\my artist\\\\my album 11\\\\pre-repair-backup\\\\1.mp3\".\n" +
					"The track file \"Music\\\\my artist\\\\my album 11\\\\2 my track 112.mp3\" has been backed up to \"Music\\\\my artist\\\\my album 11\\\\pre-repair-backup\\\\2.mp3\".\n" +
					"The track file \"Music\\\\my artist\\\\my album 11\\\\3 my track 113.mp3\" has been backed up to \"Music\\\\my artist\\\\my album 11\\\\pre-repair-backup\\\\3.mp3\".\n" +
					"The track file \"Music\\\\my artist\\\\my album 11\\\\4 my track 114.mp3\" has been backed up to \"Music\\\\my artist\\\\my album 11\\\\pre-repair-backup\\\\4.mp3\".\n" +
					"The track file \"Music\\\\my artist\\\\my album 12\\\\1 my track 121.mp3\" has been backed up to \"Music\\\\my artist\\\\my album 12\\\\pre-repair-backup\\\\1.mp3\".\n" +
					"The track file \"Music\\\\my artist\\\\my album 12\\\\2 my track 122.mp3\" has been backed up to \"Music\\\\my artist\\\\my album 12\\\\pre-repair-backup\\\\2.mp3\".\n" +
					"The track file \"Music\\\\my artist\\\\my album 12\\\\3 my track 123.mp3\" has been backed up to \"Music\\\\my artist\\\\my album 12\\\\pre-repair-backup\\\\3.mp3\".\n" +
					"The track file \"Music\\\\my artist\\\\my album 12\\\\4 my track 124.mp3\" has been backed up to \"Music\\\\my artist\\\\my album 12\\\\pre-repair-backup\\\\4.mp3\".\n",
				Error: "" +
					"An error occurred repairing track \"Music\\\\my artist\\\\my album 00\\\\1 my track 001.mp3\".\n" +
					"An error occurred repairing track \"Music\\\\my artist\\\\my album 00\\\\2 my track 002.mp3\".\n" +
					"An error occurred repairing track \"Music\\\\my artist\\\\my album 00\\\\3 my track 003.mp3\".\n" +
					"An error occurred repairing track \"Music\\\\my artist\\\\my album 00\\\\4 my track 004.mp3\".\n" +
					"An error occurred repairing track \"Music\\\\my artist\\\\my album 01\\\\1 my track 011.mp3\".\n" +
					"An error occurred repairing track \"Music\\\\my artist\\\\my album 01\\\\2 my track 012.mp3\".\n" +
					"An error occurred repairing track \"Music\\\\my artist\\\\my album 01\\\\3 my track 013.mp3\".\n" +
					"An error occurred repairing track \"Music\\\\my artist\\\\my album 01\\\\4 my track 014.mp3\".\n" +
					"An error occurred repairing track \"Music\\\\my artist\\\\my album 02\\\\1 my track 021.mp3\".\n" +
					"An error occurred repairing track \"Music\\\\my artist\\\\my album 02\\\\2 my track 022.mp3\".\n" +
					"An error occurred repairing track \"Music\\\\my artist\\\\my album 02\\\\3 my track 023.mp3\".\n" +
					"An error occurred repairing track \"Music\\\\my artist\\\\my album 02\\\\4 my track 024.mp3\".\n" +
					"An error occurred repairing track \"Music\\\\my artist\\\\my album 10\\\\1 my track 101.mp3\".\n" +
					"An error occurred repairing track \"Music\\\\my artist\\\\my album 10\\\\2 my track 102.mp3\".\n" +
					"An error occurred repairing track \"Music\\\\my artist\\\\my album 10\\\\3 my track 103.mp3\".\n" +
					"An error occurred repairing track \"Music\\\\my artist\\\\my album 10\\\\4 my track 104.mp3\".\n" +
					"An error occurred repairing track \"Music\\\\my artist\\\\my album 11\\\\1 my track 111.mp3\".\n" +
					"An error occurred repairing track \"Music\\\\my artist\\\\my album 11\\\\2 my track 112.mp3\".\n" +
					"An error occurred repairing track \"Music\\\\my artist\\\\my album 11\\\\3 my track 113.mp3\".\n" +
					"An error occurred repairing track \"Music\\\\my artist\\\\my album 11\\\\4 my track 114.mp3\".\n" +
					"An error occurred repairing track \"Music\\\\my artist\\\\my album 12\\\\1 my track 121.mp3\".\n" +
					"An error occurred repairing track \"Music\\\\my artist\\\\my album 12\\\\2 my track 122.mp3\".\n" +
					"An error occurred repairing track \"Music\\\\my artist\\\\my album 12\\\\3 my track 123.mp3\".\n" +
					"An error occurred repairing track \"Music\\\\my artist\\\\my album 12\\\\4 my track 124.mp3\".\n",
				Log: "" +
					"level='error' command='repair' directory='Music\\my artist\\my album 00' error='[\"no edit required\"]' fileName='1 my track 001.mp3' msg='cannot edit track'\n" +
					"level='error' command='repair' directory='Music\\my artist\\my album 00' error='[\"no edit required\"]' fileName='2 my track 002.mp3' msg='cannot edit track'\n" +
					"level='error' command='repair' directory='Music\\my artist\\my album 00' error='[\"no edit required\"]' fileName='3 my track 003.mp3' msg='cannot edit track'\n" +
					"level='error' command='repair' directory='Music\\my artist\\my album 00' error='[\"no edit required\"]' fileName='4 my track 004.mp3' msg='cannot edit track'\n" +
					"level='error' command='repair' directory='Music\\my artist\\my album 01' error='[\"no edit required\"]' fileName='1 my track 011.mp3' msg='cannot edit track'\n" +
					"level='error' command='repair' directory='Music\\my artist\\my album 01' error='[\"no edit required\"]' fileName='2 my track 012.mp3' msg='cannot edit track'\n" +
					"level='error' command='repair' directory='Music\\my artist\\my album 01' error='[\"no edit required\"]' fileName='3 my track 013.mp3' msg='cannot edit track'\n" +
					"level='error' command='repair' directory='Music\\my artist\\my album 01' error='[\"no edit required\"]' fileName='4 my track 014.mp3' msg='cannot edit track'\n" +
					"level='error' command='repair' directory='Music\\my artist\\my album 02' error='[\"no edit required\"]' fileName='1 my track 021.mp3' msg='cannot edit track'\n" +
					"level='error' command='repair' directory='Music\\my artist\\my album 02' error='[\"no edit required\"]' fileName='2 my track 022.mp3' msg='cannot edit track'\n" +
					"level='error' command='repair' directory='Music\\my artist\\my album 02' error='[\"no edit required\"]' fileName='3 my track 023.mp3' msg='cannot edit track'\n" +
					"level='error' command='repair' directory='Music\\my artist\\my album 02' error='[\"no edit required\"]' fileName='4 my track 024.mp3' msg='cannot edit track'\n" +
					"level='error' command='repair' directory='Music\\my artist\\my album 10' error='[\"no edit required\"]' fileName='1 my track 101.mp3' msg='cannot edit track'\n" +
					"level='error' command='repair' directory='Music\\my artist\\my album 10' error='[\"no edit required\"]' fileName='2 my track 102.mp3' msg='cannot edit track'\n" +
					"level='error' command='repair' directory='Music\\my artist\\my album 10' error='[\"no edit required\"]' fileName='3 my track 103.mp3' msg='cannot edit track'\n" +
					"level='error' command='repair' directory='Music\\my artist\\my album 10' error='[\"no edit required\"]' fileName='4 my track 104.mp3' msg='cannot edit track'\n" +
					"level='error' command='repair' directory='Music\\my artist\\my album 11' error='[\"no edit required\"]' fileName='1 my track 111.mp3' msg='cannot edit track'\n" +
					"level='error' command='repair' directory='Music\\my artist\\my album 11' error='[\"no edit required\"]' fileName='2 my track 112.mp3' msg='cannot edit track'\n" +
					"level='error' command='repair' directory='Music\\my artist\\my album 11' error='[\"no edit required\"]' fileName='3 my track 113.mp3' msg='cannot edit track'\n" +
					"level='error' command='repair' directory='Music\\my artist\\my album 11' error='[\"no edit required\"]' fileName='4 my track 114.mp3' msg='cannot edit track'\n" +
					"level='error' command='repair' directory='Music\\my artist\\my album 12' error='[\"no edit required\"]' fileName='1 my track 121.mp3' msg='cannot edit track'\n" +
					"level='error' command='repair' directory='Music\\my artist\\my album 12' error='[\"no edit required\"]' fileName='2 my track 122.mp3' msg='cannot edit track'\n" +
					"level='error' command='repair' directory='Music\\my artist\\my album 12' error='[\"no edit required\"]' fileName='3 my track 123.mp3' msg='cannot edit track'\n" +
					"level='error' command='repair' directory='Music\\my artist\\my album 12' error='[\"no edit required\"]' fileName='4 my track 124.mp3' msg='cannot edit track'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			cmd.BackupAndFix(o, tt.checkedArtists)
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("BackupAndFix() %s", issue)
				}
			}
		})
	}
}

func TestReportRepairsNeeded(t *testing.T) {
	dirty := cmd.PrepareCheckedArtists(generateArtists(2, 3, 4))
	for _, cAr := range dirty {
		for _, cAl := range cAr.Albums() {
			for _, cT := range cAl.Tracks() {
				cT.AddIssue(cmd.CheckConflictIssue, "artist field does not match artist name")
			}
		}
	}
	clean := cmd.PrepareCheckedArtists(generateArtists(2, 3, 4))
	tests := map[string]struct {
		checkedArtists []*cmd.CheckedArtist
		output.WantedRecording
	}{
		"clean": {
			checkedArtists: clean,
			WantedRecording: output.WantedRecording{
				Console: "No repairable track defects were found.\n",
			},
		},
		"dirty": {
			checkedArtists: dirty,
			WantedRecording: output.WantedRecording{
				Console: "" +
					"The following issues can be repaired:\n" +
					"Artist \"my artist 0\"\n" +
					"  Album \"my album 00\"\n" +
					"    Track \"my track 001\"\n" +
					"    * [metadata conflict] artist field does not match artist name\n" +
					"    Track \"my track 002\"\n" +
					"    * [metadata conflict] artist field does not match artist name\n" +
					"    Track \"my track 003\"\n" +
					"    * [metadata conflict] artist field does not match artist name\n" +
					"    Track \"my track 004\"\n" +
					"    * [metadata conflict] artist field does not match artist name\n" +
					"  Album \"my album 01\"\n" +
					"    Track \"my track 011\"\n" +
					"    * [metadata conflict] artist field does not match artist name\n" +
					"    Track \"my track 012\"\n" +
					"    * [metadata conflict] artist field does not match artist name\n" +
					"    Track \"my track 013\"\n" +
					"    * [metadata conflict] artist field does not match artist name\n" +
					"    Track \"my track 014\"\n" +
					"    * [metadata conflict] artist field does not match artist name\n" +
					"  Album \"my album 02\"\n" +
					"    Track \"my track 021\"\n" +
					"    * [metadata conflict] artist field does not match artist name\n" +
					"    Track \"my track 022\"\n" +
					"    * [metadata conflict] artist field does not match artist name\n" +
					"    Track \"my track 023\"\n" +
					"    * [metadata conflict] artist field does not match artist name\n" +
					"    Track \"my track 024\"\n" +
					"    * [metadata conflict] artist field does not match artist name\n" +
					"Artist \"my artist 1\"\n" +
					"  Album \"my album 10\"\n" +
					"    Track \"my track 101\"\n" +
					"    * [metadata conflict] artist field does not match artist name\n" +
					"    Track \"my track 102\"\n" +
					"    * [metadata conflict] artist field does not match artist name\n" +
					"    Track \"my track 103\"\n" +
					"    * [metadata conflict] artist field does not match artist name\n" +
					"    Track \"my track 104\"\n" +
					"    * [metadata conflict] artist field does not match artist name\n" +
					"  Album \"my album 11\"\n" +
					"    Track \"my track 111\"\n" +
					"    * [metadata conflict] artist field does not match artist name\n" +
					"    Track \"my track 112\"\n" +
					"    * [metadata conflict] artist field does not match artist name\n" +
					"    Track \"my track 113\"\n" +
					"    * [metadata conflict] artist field does not match artist name\n" +
					"    Track \"my track 114\"\n" +
					"    * [metadata conflict] artist field does not match artist name\n" +
					"  Album \"my album 12\"\n" +
					"    Track \"my track 121\"\n" +
					"    * [metadata conflict] artist field does not match artist name\n" +
					"    Track \"my track 122\"\n" +
					"    * [metadata conflict] artist field does not match artist name\n" +
					"    Track \"my track 123\"\n" +
					"    * [metadata conflict] artist field does not match artist name\n" +
					"    Track \"my track 124\"\n" +
					"    * [metadata conflict] artist field does not match artist name\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			cmd.ReportRepairsNeeded(o, tt.checkedArtists)
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("ReportRepairsNeeded() %s", issue)
				}
			}
		})
	}
}

func TestFindConflictedTracks(t *testing.T) {
	dirty := cmd.PrepareCheckedArtists(generateArtists(2, 3, 4))
	for _, cAr := range dirty {
		for _, cAl := range cAr.Albums() {
			for _, cT := range cAl.Tracks() {
				t := cT.Track()
				t.SetMetadata(&files.TrackMetadata{
					Album:                      []string{"", "some other album", "some other album"},
					Artist:                     []string{"", "some other artist", "some other artist"},
					Genre:                      []string{"", "pop emo", "pop emo"},
					MusicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{1, 2, 3}},
					Title:                      []string{"", "some other title", "some other title"},
					Track:                      []int{0, 99, 99},
					Year:                       []string{"", "2001", "2001"},
					CanonicalType:              files.ID3V1,
					ErrCause:                   []string{"", "", ""},
					CorrectedAlbum:             []string{"", "", ""},
					CorrectedArtist:            []string{"", "", ""},
					CorrectedGenre:             []string{"", "", ""},
					CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
					CorrectedTitle:             []string{"", "", ""},
					CorrectedTrack:             []int{0, 0, 0},
					CorrectedYear:              []string{"", "", ""},
					RequiresEdit:               []bool{false, false, false},
				})
			}
		}
	}
	clean := cmd.PrepareCheckedArtists(generateArtists(2, 3, 4))
	tests := map[string]struct {
		checkedArtists []*cmd.CheckedArtist
		want           int
	}{
		"clean": {checkedArtists: clean, want: 0},
		"dirty": {checkedArtists: dirty, want: 24},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := cmd.FindConflictedTracks(tt.checkedArtists); got != tt.want {
				t.Errorf("FindConflictedTracks() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRepairSettings_RepairArtists(t *testing.T) {
	oldMetadataReader := cmd.MetadataReader
	oldDirExists := cmd.DirExists
	oldPlainFileExists := cmd.PlainFileExists
	oldCopyFile := cmd.CopyFile
	oldMarkDirty := cmd.MarkDirty
	defer func() {
		cmd.MetadataReader = oldMetadataReader
		cmd.DirExists = oldDirExists
		cmd.PlainFileExists = oldPlainFileExists
		cmd.CopyFile = oldCopyFile
		cmd.MarkDirty = oldMarkDirty
	}()
	cmd.MetadataReader = func(_ output.Bus, _ []*files.Artist) {}
	cmd.DirExists = func(_ string) bool { return true }
	cmd.PlainFileExists = func(_ string) bool { return false }
	cmd.CopyFile = func(_, _ string) error { return nil }
	cmd.MarkDirty = func(_ output.Bus) {}
	dirty := generateArtists(2, 3, 4)
	for _, aR := range dirty {
		for _, aL := range aR.Albums() {
			for _, t := range aL.Tracks() {
				t.SetMetadata(&files.TrackMetadata{
					Album:                      []string{"", "", ""},
					Artist:                     []string{"", "", ""},
					Genre:                      []string{"", "", ""},
					MusicCDIdentifier:          id3v2.UnknownFrame{Body: []byte{1, 2, 3}},
					Title:                      []string{"", "", ""},
					Track:                      []int{0, 99, 99},
					Year:                       []string{"", "", ""},
					CanonicalType:              files.ID3V1,
					ErrCause:                   []string{"", "", ""},
					CorrectedAlbum:             []string{"", "", ""},
					CorrectedArtist:            []string{"", "", ""},
					CorrectedGenre:             []string{"", "", ""},
					CorrectedMusicCDIdentifier: id3v2.UnknownFrame{},
					CorrectedTitle:             []string{"", "", ""},
					CorrectedTrack:             []int{0, 0, 0},
					CorrectedYear:              []string{"", "", ""},
					RequiresEdit:               []bool{false, false, false},
				})
			}
		}
	}
	tests := map[string]struct {
		rs      *cmd.RepairSettings
		artists []*files.Artist
		output.WantedRecording
	}{
		"clean dry run": {
			rs:      &cmd.RepairSettings{DryRun: true},
			artists: generateArtists(2, 3, 4),
			WantedRecording: output.WantedRecording{
				Console: "No repairable track defects were found.\n",
			},
		},
		"dirty dry run": {
			rs:      &cmd.RepairSettings{DryRun: true},
			artists: dirty,
			WantedRecording: output.WantedRecording{
				Console: "" +
					"The following issues can be repaired:\n" +
					"Artist \"my artist 0\"\n" +
					"  Album \"my album 00\"\n" +
					"    Track \"my track 001\"\n" +
					"    * [metadata conflict] the album name field does not match the name of the album directory\n" +
					"    * [metadata conflict] the artist name field does not match the name of the artist directory\n" +
					"    * [metadata conflict] the music CD identifier field does not match the other tracks in the album\n" +
					"    * [metadata conflict] the track name field does not match the track's file name\n" +
					"    * [metadata conflict] the track number field does not match the track's file name\n" +
					"    Track \"my track 002\"\n" +
					"    * [metadata conflict] the album name field does not match the name of the album directory\n" +
					"    * [metadata conflict] the artist name field does not match the name of the artist directory\n" +
					"    * [metadata conflict] the music CD identifier field does not match the other tracks in the album\n" +
					"    * [metadata conflict] the track name field does not match the track's file name\n" +
					"    * [metadata conflict] the track number field does not match the track's file name\n" +
					"    Track \"my track 003\"\n" +
					"    * [metadata conflict] the album name field does not match the name of the album directory\n" +
					"    * [metadata conflict] the artist name field does not match the name of the artist directory\n" +
					"    * [metadata conflict] the music CD identifier field does not match the other tracks in the album\n" +
					"    * [metadata conflict] the track name field does not match the track's file name\n" +
					"    * [metadata conflict] the track number field does not match the track's file name\n" +
					"    Track \"my track 004\"\n" +
					"    * [metadata conflict] the album name field does not match the name of the album directory\n" +
					"    * [metadata conflict] the artist name field does not match the name of the artist directory\n" +
					"    * [metadata conflict] the music CD identifier field does not match the other tracks in the album\n" +
					"    * [metadata conflict] the track name field does not match the track's file name\n" +
					"    * [metadata conflict] the track number field does not match the track's file name\n" +
					"  Album \"my album 01\"\n" +
					"    Track \"my track 011\"\n" +
					"    * [metadata conflict] the album name field does not match the name of the album directory\n" +
					"    * [metadata conflict] the artist name field does not match the name of the artist directory\n" +
					"    * [metadata conflict] the music CD identifier field does not match the other tracks in the album\n" +
					"    * [metadata conflict] the track name field does not match the track's file name\n" +
					"    * [metadata conflict] the track number field does not match the track's file name\n" +
					"    Track \"my track 012\"\n" +
					"    * [metadata conflict] the album name field does not match the name of the album directory\n" +
					"    * [metadata conflict] the artist name field does not match the name of the artist directory\n" +
					"    * [metadata conflict] the music CD identifier field does not match the other tracks in the album\n" +
					"    * [metadata conflict] the track name field does not match the track's file name\n" +
					"    * [metadata conflict] the track number field does not match the track's file name\n" +
					"    Track \"my track 013\"\n" +
					"    * [metadata conflict] the album name field does not match the name of the album directory\n" +
					"    * [metadata conflict] the artist name field does not match the name of the artist directory\n" +
					"    * [metadata conflict] the music CD identifier field does not match the other tracks in the album\n" +
					"    * [metadata conflict] the track name field does not match the track's file name\n" +
					"    * [metadata conflict] the track number field does not match the track's file name\n" +
					"    Track \"my track 014\"\n" +
					"    * [metadata conflict] the album name field does not match the name of the album directory\n" +
					"    * [metadata conflict] the artist name field does not match the name of the artist directory\n" +
					"    * [metadata conflict] the music CD identifier field does not match the other tracks in the album\n" +
					"    * [metadata conflict] the track name field does not match the track's file name\n" +
					"    * [metadata conflict] the track number field does not match the track's file name\n" +
					"  Album \"my album 02\"\n" +
					"    Track \"my track 021\"\n" +
					"    * [metadata conflict] the album name field does not match the name of the album directory\n" +
					"    * [metadata conflict] the artist name field does not match the name of the artist directory\n" +
					"    * [metadata conflict] the music CD identifier field does not match the other tracks in the album\n" +
					"    * [metadata conflict] the track name field does not match the track's file name\n" +
					"    * [metadata conflict] the track number field does not match the track's file name\n" +
					"    Track \"my track 022\"\n" +
					"    * [metadata conflict] the album name field does not match the name of the album directory\n" +
					"    * [metadata conflict] the artist name field does not match the name of the artist directory\n" +
					"    * [metadata conflict] the music CD identifier field does not match the other tracks in the album\n" +
					"    * [metadata conflict] the track name field does not match the track's file name\n" +
					"    * [metadata conflict] the track number field does not match the track's file name\n" +
					"    Track \"my track 023\"\n" +
					"    * [metadata conflict] the album name field does not match the name of the album directory\n" +
					"    * [metadata conflict] the artist name field does not match the name of the artist directory\n" +
					"    * [metadata conflict] the music CD identifier field does not match the other tracks in the album\n" +
					"    * [metadata conflict] the track name field does not match the track's file name\n" +
					"    * [metadata conflict] the track number field does not match the track's file name\n" +
					"    Track \"my track 024\"\n" +
					"    * [metadata conflict] the album name field does not match the name of the album directory\n" +
					"    * [metadata conflict] the artist name field does not match the name of the artist directory\n" +
					"    * [metadata conflict] the music CD identifier field does not match the other tracks in the album\n" +
					"    * [metadata conflict] the track name field does not match the track's file name\n" +
					"    * [metadata conflict] the track number field does not match the track's file name\n" +
					"Artist \"my artist 1\"\n" +
					"  Album \"my album 10\"\n" +
					"    Track \"my track 101\"\n" +
					"    * [metadata conflict] the album name field does not match the name of the album directory\n" +
					"    * [metadata conflict] the artist name field does not match the name of the artist directory\n" +
					"    * [metadata conflict] the music CD identifier field does not match the other tracks in the album\n" +
					"    * [metadata conflict] the track name field does not match the track's file name\n" +
					"    * [metadata conflict] the track number field does not match the track's file name\n" +
					"    Track \"my track 102\"\n" +
					"    * [metadata conflict] the album name field does not match the name of the album directory\n" +
					"    * [metadata conflict] the artist name field does not match the name of the artist directory\n" +
					"    * [metadata conflict] the music CD identifier field does not match the other tracks in the album\n" +
					"    * [metadata conflict] the track name field does not match the track's file name\n" +
					"    * [metadata conflict] the track number field does not match the track's file name\n" +
					"    Track \"my track 103\"\n" +
					"    * [metadata conflict] the album name field does not match the name of the album directory\n" +
					"    * [metadata conflict] the artist name field does not match the name of the artist directory\n" +
					"    * [metadata conflict] the music CD identifier field does not match the other tracks in the album\n" +
					"    * [metadata conflict] the track name field does not match the track's file name\n" +
					"    * [metadata conflict] the track number field does not match the track's file name\n" +
					"    Track \"my track 104\"\n" +
					"    * [metadata conflict] the album name field does not match the name of the album directory\n" +
					"    * [metadata conflict] the artist name field does not match the name of the artist directory\n" +
					"    * [metadata conflict] the music CD identifier field does not match the other tracks in the album\n" +
					"    * [metadata conflict] the track name field does not match the track's file name\n" +
					"    * [metadata conflict] the track number field does not match the track's file name\n" +
					"  Album \"my album 11\"\n" +
					"    Track \"my track 111\"\n" +
					"    * [metadata conflict] the album name field does not match the name of the album directory\n" +
					"    * [metadata conflict] the artist name field does not match the name of the artist directory\n" +
					"    * [metadata conflict] the music CD identifier field does not match the other tracks in the album\n" +
					"    * [metadata conflict] the track name field does not match the track's file name\n" +
					"    * [metadata conflict] the track number field does not match the track's file name\n" +
					"    Track \"my track 112\"\n" +
					"    * [metadata conflict] the album name field does not match the name of the album directory\n" +
					"    * [metadata conflict] the artist name field does not match the name of the artist directory\n" +
					"    * [metadata conflict] the music CD identifier field does not match the other tracks in the album\n" +
					"    * [metadata conflict] the track name field does not match the track's file name\n" +
					"    * [metadata conflict] the track number field does not match the track's file name\n" +
					"    Track \"my track 113\"\n" +
					"    * [metadata conflict] the album name field does not match the name of the album directory\n" +
					"    * [metadata conflict] the artist name field does not match the name of the artist directory\n" +
					"    * [metadata conflict] the music CD identifier field does not match the other tracks in the album\n" +
					"    * [metadata conflict] the track name field does not match the track's file name\n" +
					"    * [metadata conflict] the track number field does not match the track's file name\n" +
					"    Track \"my track 114\"\n" +
					"    * [metadata conflict] the album name field does not match the name of the album directory\n" +
					"    * [metadata conflict] the artist name field does not match the name of the artist directory\n" +
					"    * [metadata conflict] the music CD identifier field does not match the other tracks in the album\n" +
					"    * [metadata conflict] the track name field does not match the track's file name\n" +
					"    * [metadata conflict] the track number field does not match the track's file name\n" +
					"  Album \"my album 12\"\n" +
					"    Track \"my track 121\"\n" +
					"    * [metadata conflict] the album name field does not match the name of the album directory\n" +
					"    * [metadata conflict] the artist name field does not match the name of the artist directory\n" +
					"    * [metadata conflict] the music CD identifier field does not match the other tracks in the album\n" +
					"    * [metadata conflict] the track name field does not match the track's file name\n" +
					"    * [metadata conflict] the track number field does not match the track's file name\n" +
					"    Track \"my track 122\"\n" +
					"    * [metadata conflict] the album name field does not match the name of the album directory\n" +
					"    * [metadata conflict] the artist name field does not match the name of the artist directory\n" +
					"    * [metadata conflict] the music CD identifier field does not match the other tracks in the album\n" +
					"    * [metadata conflict] the track name field does not match the track's file name\n" +
					"    * [metadata conflict] the track number field does not match the track's file name\n" +
					"    Track \"my track 123\"\n" +
					"    * [metadata conflict] the album name field does not match the name of the album directory\n" +
					"    * [metadata conflict] the artist name field does not match the name of the artist directory\n" +
					"    * [metadata conflict] the music CD identifier field does not match the other tracks in the album\n" +
					"    * [metadata conflict] the track name field does not match the track's file name\n" +
					"    * [metadata conflict] the track number field does not match the track's file name\n" +
					"    Track \"my track 124\"\n" +
					"    * [metadata conflict] the album name field does not match the name of the album directory\n" +
					"    * [metadata conflict] the artist name field does not match the name of the artist directory\n" +
					"    * [metadata conflict] the music CD identifier field does not match the other tracks in the album\n" +
					"    * [metadata conflict] the track name field does not match the track's file name\n" +
					"    * [metadata conflict] the track number field does not match the track's file name\n",
			},
		},
		"clean repair": {
			rs:      &cmd.RepairSettings{DryRun: false},
			artists: generateArtists(2, 3, 4),
			WantedRecording: output.WantedRecording{
				Console: "No repairable track defects were found.\n",
			},
		},
		"dirty repair": {
			rs:      &cmd.RepairSettings{DryRun: false},
			artists: dirty,
			WantedRecording: output.WantedRecording{
				Console: "" +
					"The track file \"Music\\\\my artist\\\\my album 00\\\\1 my track 001.mp3\" has been backed up to \"Music\\\\my artist\\\\my album 00\\\\pre-repair-backup\\\\1.mp3\".\n" +
					"The track file \"Music\\\\my artist\\\\my album 00\\\\2 my track 002.mp3\" has been backed up to \"Music\\\\my artist\\\\my album 00\\\\pre-repair-backup\\\\2.mp3\".\n" +
					"The track file \"Music\\\\my artist\\\\my album 00\\\\3 my track 003.mp3\" has been backed up to \"Music\\\\my artist\\\\my album 00\\\\pre-repair-backup\\\\3.mp3\".\n" +
					"The track file \"Music\\\\my artist\\\\my album 00\\\\4 my track 004.mp3\" has been backed up to \"Music\\\\my artist\\\\my album 00\\\\pre-repair-backup\\\\4.mp3\".\n" +
					"The track file \"Music\\\\my artist\\\\my album 01\\\\1 my track 011.mp3\" has been backed up to \"Music\\\\my artist\\\\my album 01\\\\pre-repair-backup\\\\1.mp3\".\n" +
					"The track file \"Music\\\\my artist\\\\my album 01\\\\2 my track 012.mp3\" has been backed up to \"Music\\\\my artist\\\\my album 01\\\\pre-repair-backup\\\\2.mp3\".\n" +
					"The track file \"Music\\\\my artist\\\\my album 01\\\\3 my track 013.mp3\" has been backed up to \"Music\\\\my artist\\\\my album 01\\\\pre-repair-backup\\\\3.mp3\".\n" +
					"The track file \"Music\\\\my artist\\\\my album 01\\\\4 my track 014.mp3\" has been backed up to \"Music\\\\my artist\\\\my album 01\\\\pre-repair-backup\\\\4.mp3\".\n" +
					"The track file \"Music\\\\my artist\\\\my album 02\\\\1 my track 021.mp3\" has been backed up to \"Music\\\\my artist\\\\my album 02\\\\pre-repair-backup\\\\1.mp3\".\n" +
					"The track file \"Music\\\\my artist\\\\my album 02\\\\2 my track 022.mp3\" has been backed up to \"Music\\\\my artist\\\\my album 02\\\\pre-repair-backup\\\\2.mp3\".\n" +
					"The track file \"Music\\\\my artist\\\\my album 02\\\\3 my track 023.mp3\" has been backed up to \"Music\\\\my artist\\\\my album 02\\\\pre-repair-backup\\\\3.mp3\".\n" +
					"The track file \"Music\\\\my artist\\\\my album 02\\\\4 my track 024.mp3\" has been backed up to \"Music\\\\my artist\\\\my album 02\\\\pre-repair-backup\\\\4.mp3\".\n" +
					"The track file \"Music\\\\my artist\\\\my album 10\\\\1 my track 101.mp3\" has been backed up to \"Music\\\\my artist\\\\my album 10\\\\pre-repair-backup\\\\1.mp3\".\n" +
					"The track file \"Music\\\\my artist\\\\my album 10\\\\2 my track 102.mp3\" has been backed up to \"Music\\\\my artist\\\\my album 10\\\\pre-repair-backup\\\\2.mp3\".\n" +
					"The track file \"Music\\\\my artist\\\\my album 10\\\\3 my track 103.mp3\" has been backed up to \"Music\\\\my artist\\\\my album 10\\\\pre-repair-backup\\\\3.mp3\".\n" +
					"The track file \"Music\\\\my artist\\\\my album 10\\\\4 my track 104.mp3\" has been backed up to \"Music\\\\my artist\\\\my album 10\\\\pre-repair-backup\\\\4.mp3\".\n" +
					"The track file \"Music\\\\my artist\\\\my album 11\\\\1 my track 111.mp3\" has been backed up to \"Music\\\\my artist\\\\my album 11\\\\pre-repair-backup\\\\1.mp3\".\n" +
					"The track file \"Music\\\\my artist\\\\my album 11\\\\2 my track 112.mp3\" has been backed up to \"Music\\\\my artist\\\\my album 11\\\\pre-repair-backup\\\\2.mp3\".\n" +
					"The track file \"Music\\\\my artist\\\\my album 11\\\\3 my track 113.mp3\" has been backed up to \"Music\\\\my artist\\\\my album 11\\\\pre-repair-backup\\\\3.mp3\".\n" +
					"The track file \"Music\\\\my artist\\\\my album 11\\\\4 my track 114.mp3\" has been backed up to \"Music\\\\my artist\\\\my album 11\\\\pre-repair-backup\\\\4.mp3\".\n" +
					"The track file \"Music\\\\my artist\\\\my album 12\\\\1 my track 121.mp3\" has been backed up to \"Music\\\\my artist\\\\my album 12\\\\pre-repair-backup\\\\1.mp3\".\n" +
					"The track file \"Music\\\\my artist\\\\my album 12\\\\2 my track 122.mp3\" has been backed up to \"Music\\\\my artist\\\\my album 12\\\\pre-repair-backup\\\\2.mp3\".\n" +
					"The track file \"Music\\\\my artist\\\\my album 12\\\\3 my track 123.mp3\" has been backed up to \"Music\\\\my artist\\\\my album 12\\\\pre-repair-backup\\\\3.mp3\".\n" +
					"The track file \"Music\\\\my artist\\\\my album 12\\\\4 my track 124.mp3\" has been backed up to \"Music\\\\my artist\\\\my album 12\\\\pre-repair-backup\\\\4.mp3\".\n",
				Error: "" +
					"An error occurred repairing track \"Music\\\\my artist\\\\my album 00\\\\1 my track 001.mp3\".\n" +
					"An error occurred repairing track \"Music\\\\my artist\\\\my album 00\\\\2 my track 002.mp3\".\n" +
					"An error occurred repairing track \"Music\\\\my artist\\\\my album 00\\\\3 my track 003.mp3\".\n" +
					"An error occurred repairing track \"Music\\\\my artist\\\\my album 00\\\\4 my track 004.mp3\".\n" +
					"An error occurred repairing track \"Music\\\\my artist\\\\my album 01\\\\1 my track 011.mp3\".\n" +
					"An error occurred repairing track \"Music\\\\my artist\\\\my album 01\\\\2 my track 012.mp3\".\n" +
					"An error occurred repairing track \"Music\\\\my artist\\\\my album 01\\\\3 my track 013.mp3\".\n" +
					"An error occurred repairing track \"Music\\\\my artist\\\\my album 01\\\\4 my track 014.mp3\".\n" +
					"An error occurred repairing track \"Music\\\\my artist\\\\my album 02\\\\1 my track 021.mp3\".\n" +
					"An error occurred repairing track \"Music\\\\my artist\\\\my album 02\\\\2 my track 022.mp3\".\n" +
					"An error occurred repairing track \"Music\\\\my artist\\\\my album 02\\\\3 my track 023.mp3\".\n" +
					"An error occurred repairing track \"Music\\\\my artist\\\\my album 02\\\\4 my track 024.mp3\".\n" +
					"An error occurred repairing track \"Music\\\\my artist\\\\my album 10\\\\1 my track 101.mp3\".\n" +
					"An error occurred repairing track \"Music\\\\my artist\\\\my album 10\\\\2 my track 102.mp3\".\n" +
					"An error occurred repairing track \"Music\\\\my artist\\\\my album 10\\\\3 my track 103.mp3\".\n" +
					"An error occurred repairing track \"Music\\\\my artist\\\\my album 10\\\\4 my track 104.mp3\".\n" +
					"An error occurred repairing track \"Music\\\\my artist\\\\my album 11\\\\1 my track 111.mp3\".\n" +
					"An error occurred repairing track \"Music\\\\my artist\\\\my album 11\\\\2 my track 112.mp3\".\n" +
					"An error occurred repairing track \"Music\\\\my artist\\\\my album 11\\\\3 my track 113.mp3\".\n" +
					"An error occurred repairing track \"Music\\\\my artist\\\\my album 11\\\\4 my track 114.mp3\".\n" +
					"An error occurred repairing track \"Music\\\\my artist\\\\my album 12\\\\1 my track 121.mp3\".\n" +
					"An error occurred repairing track \"Music\\\\my artist\\\\my album 12\\\\2 my track 122.mp3\".\n" +
					"An error occurred repairing track \"Music\\\\my artist\\\\my album 12\\\\3 my track 123.mp3\".\n" +
					"An error occurred repairing track \"Music\\\\my artist\\\\my album 12\\\\4 my track 124.mp3\".\n",
				Log: "level='error' command='repair' directory='Music\\my artist\\my album 00' error='[\"open Music\\\\my artist\\\\my album 00\\\\1 my track 001.mp3: The system cannot find the path specified.\", \"open Music\\\\my artist\\\\my album 00\\\\1 my track 001.mp3: The system cannot find the path specified.\"]' fileName='1 my track 001.mp3' msg='cannot edit track'\n" +
					"level='error' command='repair' directory='Music\\my artist\\my album 00' error='[\"open Music\\\\my artist\\\\my album 00\\\\2 my track 002.mp3: The system cannot find the path specified.\", \"open Music\\\\my artist\\\\my album 00\\\\2 my track 002.mp3: The system cannot find the path specified.\"]' fileName='2 my track 002.mp3' msg='cannot edit track'\n" +
					"level='error' command='repair' directory='Music\\my artist\\my album 00' error='[\"open Music\\\\my artist\\\\my album 00\\\\3 my track 003.mp3: The system cannot find the path specified.\", \"open Music\\\\my artist\\\\my album 00\\\\3 my track 003.mp3: The system cannot find the path specified.\"]' fileName='3 my track 003.mp3' msg='cannot edit track'\n" +
					"level='error' command='repair' directory='Music\\my artist\\my album 00' error='[\"open Music\\\\my artist\\\\my album 00\\\\4 my track 004.mp3: The system cannot find the path specified.\", \"open Music\\\\my artist\\\\my album 00\\\\4 my track 004.mp3: The system cannot find the path specified.\"]' fileName='4 my track 004.mp3' msg='cannot edit track'\n" +
					"level='error' command='repair' directory='Music\\my artist\\my album 01' error='[\"open Music\\\\my artist\\\\my album 01\\\\1 my track 011.mp3: The system cannot find the path specified.\", \"open Music\\\\my artist\\\\my album 01\\\\1 my track 011.mp3: The system cannot find the path specified.\"]' fileName='1 my track 011.mp3' msg='cannot edit track'\n" +
					"level='error' command='repair' directory='Music\\my artist\\my album 01' error='[\"open Music\\\\my artist\\\\my album 01\\\\2 my track 012.mp3: The system cannot find the path specified.\", \"open Music\\\\my artist\\\\my album 01\\\\2 my track 012.mp3: The system cannot find the path specified.\"]' fileName='2 my track 012.mp3' msg='cannot edit track'\n" +
					"level='error' command='repair' directory='Music\\my artist\\my album 01' error='[\"open Music\\\\my artist\\\\my album 01\\\\3 my track 013.mp3: The system cannot find the path specified.\", \"open Music\\\\my artist\\\\my album 01\\\\3 my track 013.mp3: The system cannot find the path specified.\"]' fileName='3 my track 013.mp3' msg='cannot edit track'\n" +
					"level='error' command='repair' directory='Music\\my artist\\my album 01' error='[\"open Music\\\\my artist\\\\my album 01\\\\4 my track 014.mp3: The system cannot find the path specified.\", \"open Music\\\\my artist\\\\my album 01\\\\4 my track 014.mp3: The system cannot find the path specified.\"]' fileName='4 my track 014.mp3' msg='cannot edit track'\n" +
					"level='error' command='repair' directory='Music\\my artist\\my album 02' error='[\"open Music\\\\my artist\\\\my album 02\\\\1 my track 021.mp3: The system cannot find the path specified.\", \"open Music\\\\my artist\\\\my album 02\\\\1 my track 021.mp3: The system cannot find the path specified.\"]' fileName='1 my track 021.mp3' msg='cannot edit track'\n" +
					"level='error' command='repair' directory='Music\\my artist\\my album 02' error='[\"open Music\\\\my artist\\\\my album 02\\\\2 my track 022.mp3: The system cannot find the path specified.\", \"open Music\\\\my artist\\\\my album 02\\\\2 my track 022.mp3: The system cannot find the path specified.\"]' fileName='2 my track 022.mp3' msg='cannot edit track'\n" +
					"level='error' command='repair' directory='Music\\my artist\\my album 02' error='[\"open Music\\\\my artist\\\\my album 02\\\\3 my track 023.mp3: The system cannot find the path specified.\", \"open Music\\\\my artist\\\\my album 02\\\\3 my track 023.mp3: The system cannot find the path specified.\"]' fileName='3 my track 023.mp3' msg='cannot edit track'\n" +
					"level='error' command='repair' directory='Music\\my artist\\my album 02' error='[\"open Music\\\\my artist\\\\my album 02\\\\4 my track 024.mp3: The system cannot find the path specified.\", \"open Music\\\\my artist\\\\my album 02\\\\4 my track 024.mp3: The system cannot find the path specified.\"]' fileName='4 my track 024.mp3' msg='cannot edit track'\n" +
					"level='error' command='repair' directory='Music\\my artist\\my album 10' error='[\"open Music\\\\my artist\\\\my album 10\\\\1 my track 101.mp3: The system cannot find the path specified.\", \"open Music\\\\my artist\\\\my album 10\\\\1 my track 101.mp3: The system cannot find the path specified.\"]' fileName='1 my track 101.mp3' msg='cannot edit track'\n" +
					"level='error' command='repair' directory='Music\\my artist\\my album 10' error='[\"open Music\\\\my artist\\\\my album 10\\\\2 my track 102.mp3: The system cannot find the path specified.\", \"open Music\\\\my artist\\\\my album 10\\\\2 my track 102.mp3: The system cannot find the path specified.\"]' fileName='2 my track 102.mp3' msg='cannot edit track'\n" +
					"level='error' command='repair' directory='Music\\my artist\\my album 10' error='[\"open Music\\\\my artist\\\\my album 10\\\\3 my track 103.mp3: The system cannot find the path specified.\", \"open Music\\\\my artist\\\\my album 10\\\\3 my track 103.mp3: The system cannot find the path specified.\"]' fileName='3 my track 103.mp3' msg='cannot edit track'\n" +
					"level='error' command='repair' directory='Music\\my artist\\my album 10' error='[\"open Music\\\\my artist\\\\my album 10\\\\4 my track 104.mp3: The system cannot find the path specified.\", \"open Music\\\\my artist\\\\my album 10\\\\4 my track 104.mp3: The system cannot find the path specified.\"]' fileName='4 my track 104.mp3' msg='cannot edit track'\n" +
					"level='error' command='repair' directory='Music\\my artist\\my album 11' error='[\"open Music\\\\my artist\\\\my album 11\\\\1 my track 111.mp3: The system cannot find the path specified.\", \"open Music\\\\my artist\\\\my album 11\\\\1 my track 111.mp3: The system cannot find the path specified.\"]' fileName='1 my track 111.mp3' msg='cannot edit track'\n" +
					"level='error' command='repair' directory='Music\\my artist\\my album 11' error='[\"open Music\\\\my artist\\\\my album 11\\\\2 my track 112.mp3: The system cannot find the path specified.\", \"open Music\\\\my artist\\\\my album 11\\\\2 my track 112.mp3: The system cannot find the path specified.\"]' fileName='2 my track 112.mp3' msg='cannot edit track'\n" +
					"level='error' command='repair' directory='Music\\my artist\\my album 11' error='[\"open Music\\\\my artist\\\\my album 11\\\\3 my track 113.mp3: The system cannot find the path specified.\", \"open Music\\\\my artist\\\\my album 11\\\\3 my track 113.mp3: The system cannot find the path specified.\"]' fileName='3 my track 113.mp3' msg='cannot edit track'\n" +
					"level='error' command='repair' directory='Music\\my artist\\my album 11' error='[\"open Music\\\\my artist\\\\my album 11\\\\4 my track 114.mp3: The system cannot find the path specified.\", \"open Music\\\\my artist\\\\my album 11\\\\4 my track 114.mp3: The system cannot find the path specified.\"]' fileName='4 my track 114.mp3' msg='cannot edit track'\n" +
					"level='error' command='repair' directory='Music\\my artist\\my album 12' error='[\"open Music\\\\my artist\\\\my album 12\\\\1 my track 121.mp3: The system cannot find the path specified.\", \"open Music\\\\my artist\\\\my album 12\\\\1 my track 121.mp3: The system cannot find the path specified.\"]' fileName='1 my track 121.mp3' msg='cannot edit track'\n" +
					"level='error' command='repair' directory='Music\\my artist\\my album 12' error='[\"open Music\\\\my artist\\\\my album 12\\\\2 my track 122.mp3: The system cannot find the path specified.\", \"open Music\\\\my artist\\\\my album 12\\\\2 my track 122.mp3: The system cannot find the path specified.\"]' fileName='2 my track 122.mp3' msg='cannot edit track'\n" +
					"level='error' command='repair' directory='Music\\my artist\\my album 12' error='[\"open Music\\\\my artist\\\\my album 12\\\\3 my track 123.mp3: The system cannot find the path specified.\", \"open Music\\\\my artist\\\\my album 12\\\\3 my track 123.mp3: The system cannot find the path specified.\"]' fileName='3 my track 123.mp3' msg='cannot edit track'\n" +
					"level='error' command='repair' directory='Music\\my artist\\my album 12' error='[\"open Music\\\\my artist\\\\my album 12\\\\4 my track 124.mp3: The system cannot find the path specified.\", \"open Music\\\\my artist\\\\my album 12\\\\4 my track 124.mp3: The system cannot find the path specified.\"]' fileName='4 my track 124.mp3' msg='cannot edit track'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			tt.rs.RepairArtists(o, tt.artists)
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("RepairSettings.RepairArtists() %s", issue)
				}
			}
		})
	}
}

func TestRepairSettings_ProcessArtists(t *testing.T) {
	oldMetadataReader := cmd.MetadataReader
	defer func() {
		cmd.MetadataReader = oldMetadataReader
	}()
	cmd.MetadataReader = func(_ output.Bus, _ []*files.Artist) {}
	type args struct {
		allArtists []*files.Artist
		loaded     bool
		ss         *cmd.SearchSettings
	}
	tests := map[string]struct {
		rs *cmd.RepairSettings
		args
		output.WantedRecording
	}{
		"nothing to do": {
			rs:   &cmd.RepairSettings{DryRun: true},
			args: args{},
		},
		"clean artists": {
			rs: &cmd.RepairSettings{DryRun: true},
			args: args{
				allArtists: generateArtists(2, 3, 4),
				loaded:     true,
				ss: &cmd.SearchSettings{
					ArtistFilter: regexp.MustCompile(".*"),
					AlbumFilter:  regexp.MustCompile(".*"),
					TrackFilter:  regexp.MustCompile(".*"),
				},
			},
			WantedRecording: output.WantedRecording{
				Console: "No repairable track defects were found.\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			tt.rs.ProcessArtists(o, tt.args.allArtists, tt.args.loaded, tt.args.ss)
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("RepairSettings.ProcessArtists() %s", issue)
				}
			}
		})
	}
}

func TestRepairRun(t *testing.T) {
	cmd.InitGlobals()
	oldBus := cmd.Bus
	oldSearchFlags := cmd.SearchFlags
	defer func() {
		cmd.Bus = oldBus
		cmd.SearchFlags = oldSearchFlags
	}()
	cmd.SearchFlags = safeSearchFlags
	repairFlags := cmd.SectionFlags{
		SectionName: "repair",
		Flags: map[string]*cmd.FlagDetails{
			"dryRun": {
				Usage:        "output what would have been repaired, but make no repairs",
				ExpectedType: cmd.BoolType,
				DefaultValue: false,
			},
		},
	}
	command := &cobra.Command{}
	cmd.AddFlags(output.NewNilBus(), cmd_toolkit.EmptyConfiguration(), command.Flags(), repairFlags, true)
	tests := map[string]struct {
		cmd *cobra.Command
		in1 []string
		output.WantedRecording
	}{
		"basic": {
			cmd: command,
			WantedRecording: output.WantedRecording{
				Error: "" +
					"No music files could be found using the specified parameters.\n" +
					"Why?\n" +
					"There were no directories found in \".\" (the --topDir value).\n" +
					"What to do:\n" +
					"Set --topDir to the path of a directory that contains artist directories.\n",
				Log: "" +
					"level='info' --albumFilter='.*' --artistFilter='.*' --dryRun='false' --topDir='.' --trackFilter='.*' command='repair' msg='executing command'\n" +
					"level='error' --topDir='.' msg='cannot find any artist directories'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			cmd.Bus = o // cook getBus()
			cmd.RepairRun(tt.cmd, tt.in1)
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("RepairRun() %s", issue)
				}
			}
		})
	}
}
