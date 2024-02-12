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
			want:   cmd.NewRepairSettings(),
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
			values: map[string]*cmd.FlagValue{"dryRun": cmd.NewFlagValue().WithValue(true)},
			want:   cmd.NewRepairSettings().WithDryRun(true),
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
	originalDirExists := cmd.DirExists
	originalMkdir := cmd.Mkdir
	defer func() {
		cmd.DirExists = originalDirExists
		cmd.Mkdir = originalMkdir
	}()
	album := &files.Album{}
	if albums := generateAlbums(1, 5); len(albums) > 0 {
		album = albums[0]
	}
	tests := map[string]struct {
		cAl        *cmd.CheckedAlbum
		dirExists  func(s string) bool
		mkdir      func(s string) error
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
			mkdir:      func(_ string) error { return nil },
			wantPath:   album.BackupDirectory(),
			wantExists: true,
		},
		"dir does not exist and cannot be created": {
			cAl:        cmd.NewCheckedAlbum(album),
			dirExists:  func(_ string) bool { return false },
			mkdir:      func(_ string) error { return fmt.Errorf("plain file exists") },
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
			cmd.Mkdir = tt.mkdir
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
	originalPlainFileExists := cmd.PlainFileExists
	originalCopyFile := cmd.CopyFile
	defer func() {
		cmd.PlainFileExists = originalPlainFileExists
		cmd.CopyFile = originalCopyFile
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
	originalMarkDirty := cmd.MarkDirty
	defer func() {
		cmd.MarkDirty = originalMarkDirty
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
		wantDirty  bool
		wantStatus int
		output.WantedRecording
	}{
		"success": {
			args:       args{t: track, err: nil},
			wantDirty:  true,
			wantStatus: cmd.Success,
			WantedRecording: output.WantedRecording{
				Console: "\"Music\\\\my artist\\\\my album 00\\\\1 my track 001.mp3\" repaired.\n",
			},
		},
		"single failure": {
			args:       args{t: track, err: []error{fmt.Errorf("file locked")}},
			wantDirty:  false,
			wantStatus: cmd.SystemError,
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
			wantDirty:  false,
			wantStatus: cmd.SystemError,
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
			if got := cmd.ProcessUpdateResult(o, tt.args.t, tt.args.err); got != tt.wantStatus {
				t.Errorf("ProcessUpdateResult() got %d want %d", got, tt.wantStatus)
			}
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
	originalDirExists := cmd.DirExists
	originalPlainFileExists := cmd.PlainFileExists
	originalCopyFile := cmd.CopyFile
	defer func() {
		cmd.DirExists = originalDirExists
		cmd.PlainFileExists = originalPlainFileExists
		cmd.CopyFile = originalCopyFile
	}()
	tests := map[string]struct {
		dirExists       func(string) bool
		plainFileExists func(string) bool
		copyFile        func(string, string) error
		checkedArtists  []*cmd.CheckedArtist
		wantStatus      int
		output.WantedRecording
	}{
		"basic test": {
			dirExists:       func(_ string) bool { return true },
			plainFileExists: func(_ string) bool { return false },
			copyFile:        func(_, _ string) error { return nil },
			checkedArtists:  checkedArtists,
			wantStatus:      cmd.SystemError,
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
		"basic test2": {
			dirExists:       func(_ string) bool { return false },
			plainFileExists: func(_ string) bool { return false },
			copyFile:        func(_, _ string) error { return nil },
			checkedArtists:  checkedArtists,
			wantStatus:      cmd.SystemError,
			WantedRecording: output.WantedRecording{
				Error: "" +
					"The directory \"Music\\\\my artist\\\\my album 00\\\\pre-repair-backup\" cannot be created: mkdir Music\\my artist\\my album 00\\pre-repair-backup: The system cannot find the path specified.\n" +
					"The track files in the directory \"Music\\\\my artist\\\\my album 00\" will not be repaired.\n" +
					"The directory \"Music\\\\my artist\\\\my album 01\\\\pre-repair-backup\" cannot be created: mkdir Music\\my artist\\my album 01\\pre-repair-backup: The system cannot find the path specified.\n" +
					"The track files in the directory \"Music\\\\my artist\\\\my album 01\" will not be repaired.\n" +
					"The directory \"Music\\\\my artist\\\\my album 02\\\\pre-repair-backup\" cannot be created: mkdir Music\\my artist\\my album 02\\pre-repair-backup: The system cannot find the path specified.\n" +
					"The track files in the directory \"Music\\\\my artist\\\\my album 02\" will not be repaired.\n" +
					"The directory \"Music\\\\my artist\\\\my album 10\\\\pre-repair-backup\" cannot be created: mkdir Music\\my artist\\my album 10\\pre-repair-backup: The system cannot find the path specified.\n" +
					"The track files in the directory \"Music\\\\my artist\\\\my album 10\" will not be repaired.\n" +
					"The directory \"Music\\\\my artist\\\\my album 11\\\\pre-repair-backup\" cannot be created: mkdir Music\\my artist\\my album 11\\pre-repair-backup: The system cannot find the path specified.\n" +
					"The track files in the directory \"Music\\\\my artist\\\\my album 11\" will not be repaired.\n" +
					"The directory \"Music\\\\my artist\\\\my album 12\\\\pre-repair-backup\" cannot be created: mkdir Music\\my artist\\my album 12\\pre-repair-backup: The system cannot find the path specified.\n" +
					"The track files in the directory \"Music\\\\my artist\\\\my album 12\" will not be repaired.\n",
				Log: "" +
					"level='error' command='repair' directory='Music\\my artist\\my album 00\\pre-repair-backup' error='mkdir Music\\my artist\\my album 00\\pre-repair-backup: The system cannot find the path specified.' msg='cannot create directory'\n" +
					"level='error' command='repair' directory='Music\\my artist\\my album 01\\pre-repair-backup' error='mkdir Music\\my artist\\my album 01\\pre-repair-backup: The system cannot find the path specified.' msg='cannot create directory'\n" +
					"level='error' command='repair' directory='Music\\my artist\\my album 02\\pre-repair-backup' error='mkdir Music\\my artist\\my album 02\\pre-repair-backup: The system cannot find the path specified.' msg='cannot create directory'\n" +
					"level='error' command='repair' directory='Music\\my artist\\my album 10\\pre-repair-backup' error='mkdir Music\\my artist\\my album 10\\pre-repair-backup: The system cannot find the path specified.' msg='cannot create directory'\n" +
					"level='error' command='repair' directory='Music\\my artist\\my album 11\\pre-repair-backup' error='mkdir Music\\my artist\\my album 11\\pre-repair-backup: The system cannot find the path specified.' msg='cannot create directory'\n" +
					"level='error' command='repair' directory='Music\\my artist\\my album 12\\pre-repair-backup' error='mkdir Music\\my artist\\my album 12\\pre-repair-backup: The system cannot find the path specified.' msg='cannot create directory'\n",
			},
		},
		"basic test3": {
			dirExists:       func(_ string) bool { return true },
			plainFileExists: func(_ string) bool { return false },
			copyFile:        func(_, _ string) error { return fmt.Errorf("oops") },
			checkedArtists:  checkedArtists,
			wantStatus:      cmd.SystemError,
			WantedRecording: output.WantedRecording{
				Error: "" +
					"The track file \"Music\\\\my artist\\\\my album 00\\\\1 my track 001.mp3\" could not be backed up due to error oops.\n" +
					"The track file \"Music\\\\my artist\\\\my album 00\\\\1 my track 001.mp3\" will not be repaired.\n" +
					"The track file \"Music\\\\my artist\\\\my album 00\\\\2 my track 002.mp3\" could not be backed up due to error oops.\n" +
					"The track file \"Music\\\\my artist\\\\my album 00\\\\2 my track 002.mp3\" will not be repaired.\n" +
					"The track file \"Music\\\\my artist\\\\my album 00\\\\3 my track 003.mp3\" could not be backed up due to error oops.\n" +
					"The track file \"Music\\\\my artist\\\\my album 00\\\\3 my track 003.mp3\" will not be repaired.\n" +
					"The track file \"Music\\\\my artist\\\\my album 00\\\\4 my track 004.mp3\" could not be backed up due to error oops.\n" +
					"The track file \"Music\\\\my artist\\\\my album 00\\\\4 my track 004.mp3\" will not be repaired.\n" +
					"The track file \"Music\\\\my artist\\\\my album 01\\\\1 my track 011.mp3\" could not be backed up due to error oops.\n" +
					"The track file \"Music\\\\my artist\\\\my album 01\\\\1 my track 011.mp3\" will not be repaired.\n" +
					"The track file \"Music\\\\my artist\\\\my album 01\\\\2 my track 012.mp3\" could not be backed up due to error oops.\n" +
					"The track file \"Music\\\\my artist\\\\my album 01\\\\2 my track 012.mp3\" will not be repaired.\n" +
					"The track file \"Music\\\\my artist\\\\my album 01\\\\3 my track 013.mp3\" could not be backed up due to error oops.\n" +
					"The track file \"Music\\\\my artist\\\\my album 01\\\\3 my track 013.mp3\" will not be repaired.\n" +
					"The track file \"Music\\\\my artist\\\\my album 01\\\\4 my track 014.mp3\" could not be backed up due to error oops.\n" +
					"The track file \"Music\\\\my artist\\\\my album 01\\\\4 my track 014.mp3\" will not be repaired.\n" +
					"The track file \"Music\\\\my artist\\\\my album 02\\\\1 my track 021.mp3\" could not be backed up due to error oops.\n" +
					"The track file \"Music\\\\my artist\\\\my album 02\\\\1 my track 021.mp3\" will not be repaired.\n" +
					"The track file \"Music\\\\my artist\\\\my album 02\\\\2 my track 022.mp3\" could not be backed up due to error oops.\n" +
					"The track file \"Music\\\\my artist\\\\my album 02\\\\2 my track 022.mp3\" will not be repaired.\n" +
					"The track file \"Music\\\\my artist\\\\my album 02\\\\3 my track 023.mp3\" could not be backed up due to error oops.\n" +
					"The track file \"Music\\\\my artist\\\\my album 02\\\\3 my track 023.mp3\" will not be repaired.\n" +
					"The track file \"Music\\\\my artist\\\\my album 02\\\\4 my track 024.mp3\" could not be backed up due to error oops.\n" +
					"The track file \"Music\\\\my artist\\\\my album 02\\\\4 my track 024.mp3\" will not be repaired.\n" +
					"The track file \"Music\\\\my artist\\\\my album 10\\\\1 my track 101.mp3\" could not be backed up due to error oops.\n" +
					"The track file \"Music\\\\my artist\\\\my album 10\\\\1 my track 101.mp3\" will not be repaired.\n" +
					"The track file \"Music\\\\my artist\\\\my album 10\\\\2 my track 102.mp3\" could not be backed up due to error oops.\n" +
					"The track file \"Music\\\\my artist\\\\my album 10\\\\2 my track 102.mp3\" will not be repaired.\n" +
					"The track file \"Music\\\\my artist\\\\my album 10\\\\3 my track 103.mp3\" could not be backed up due to error oops.\n" +
					"The track file \"Music\\\\my artist\\\\my album 10\\\\3 my track 103.mp3\" will not be repaired.\n" +
					"The track file \"Music\\\\my artist\\\\my album 10\\\\4 my track 104.mp3\" could not be backed up due to error oops.\n" +
					"The track file \"Music\\\\my artist\\\\my album 10\\\\4 my track 104.mp3\" will not be repaired.\n" +
					"The track file \"Music\\\\my artist\\\\my album 11\\\\1 my track 111.mp3\" could not be backed up due to error oops.\n" +
					"The track file \"Music\\\\my artist\\\\my album 11\\\\1 my track 111.mp3\" will not be repaired.\n" +
					"The track file \"Music\\\\my artist\\\\my album 11\\\\2 my track 112.mp3\" could not be backed up due to error oops.\n" +
					"The track file \"Music\\\\my artist\\\\my album 11\\\\2 my track 112.mp3\" will not be repaired.\n" +
					"The track file \"Music\\\\my artist\\\\my album 11\\\\3 my track 113.mp3\" could not be backed up due to error oops.\n" +
					"The track file \"Music\\\\my artist\\\\my album 11\\\\3 my track 113.mp3\" will not be repaired.\n" +
					"The track file \"Music\\\\my artist\\\\my album 11\\\\4 my track 114.mp3\" could not be backed up due to error oops.\n" +
					"The track file \"Music\\\\my artist\\\\my album 11\\\\4 my track 114.mp3\" will not be repaired.\n" +
					"The track file \"Music\\\\my artist\\\\my album 12\\\\1 my track 121.mp3\" could not be backed up due to error oops.\n" +
					"The track file \"Music\\\\my artist\\\\my album 12\\\\1 my track 121.mp3\" will not be repaired.\n" +
					"The track file \"Music\\\\my artist\\\\my album 12\\\\2 my track 122.mp3\" could not be backed up due to error oops.\n" +
					"The track file \"Music\\\\my artist\\\\my album 12\\\\2 my track 122.mp3\" will not be repaired.\n" +
					"The track file \"Music\\\\my artist\\\\my album 12\\\\3 my track 123.mp3\" could not be backed up due to error oops.\n" +
					"The track file \"Music\\\\my artist\\\\my album 12\\\\3 my track 123.mp3\" will not be repaired.\n" +
					"The track file \"Music\\\\my artist\\\\my album 12\\\\4 my track 124.mp3\" could not be backed up due to error oops.\n" +
					"The track file \"Music\\\\my artist\\\\my album 12\\\\4 my track 124.mp3\" will not be repaired.\n",
				Log: "" +
					"level='error' command='repair' destination='Music\\my artist\\my album 00\\pre-repair-backup\\1.mp3' error='oops' source='Music\\my artist\\my album 00\\1 my track 001.mp3' msg='error copying file'\n" +
					"level='error' command='repair' destination='Music\\my artist\\my album 00\\pre-repair-backup\\2.mp3' error='oops' source='Music\\my artist\\my album 00\\2 my track 002.mp3' msg='error copying file'\n" +
					"level='error' command='repair' destination='Music\\my artist\\my album 00\\pre-repair-backup\\3.mp3' error='oops' source='Music\\my artist\\my album 00\\3 my track 003.mp3' msg='error copying file'\n" +
					"level='error' command='repair' destination='Music\\my artist\\my album 00\\pre-repair-backup\\4.mp3' error='oops' source='Music\\my artist\\my album 00\\4 my track 004.mp3' msg='error copying file'\n" +
					"level='error' command='repair' destination='Music\\my artist\\my album 01\\pre-repair-backup\\1.mp3' error='oops' source='Music\\my artist\\my album 01\\1 my track 011.mp3' msg='error copying file'\n" +
					"level='error' command='repair' destination='Music\\my artist\\my album 01\\pre-repair-backup\\2.mp3' error='oops' source='Music\\my artist\\my album 01\\2 my track 012.mp3' msg='error copying file'\n" +
					"level='error' command='repair' destination='Music\\my artist\\my album 01\\pre-repair-backup\\3.mp3' error='oops' source='Music\\my artist\\my album 01\\3 my track 013.mp3' msg='error copying file'\n" +
					"level='error' command='repair' destination='Music\\my artist\\my album 01\\pre-repair-backup\\4.mp3' error='oops' source='Music\\my artist\\my album 01\\4 my track 014.mp3' msg='error copying file'\n" +
					"level='error' command='repair' destination='Music\\my artist\\my album 02\\pre-repair-backup\\1.mp3' error='oops' source='Music\\my artist\\my album 02\\1 my track 021.mp3' msg='error copying file'\n" +
					"level='error' command='repair' destination='Music\\my artist\\my album 02\\pre-repair-backup\\2.mp3' error='oops' source='Music\\my artist\\my album 02\\2 my track 022.mp3' msg='error copying file'\n" +
					"level='error' command='repair' destination='Music\\my artist\\my album 02\\pre-repair-backup\\3.mp3' error='oops' source='Music\\my artist\\my album 02\\3 my track 023.mp3' msg='error copying file'\n" +
					"level='error' command='repair' destination='Music\\my artist\\my album 02\\pre-repair-backup\\4.mp3' error='oops' source='Music\\my artist\\my album 02\\4 my track 024.mp3' msg='error copying file'\n" +
					"level='error' command='repair' destination='Music\\my artist\\my album 10\\pre-repair-backup\\1.mp3' error='oops' source='Music\\my artist\\my album 10\\1 my track 101.mp3' msg='error copying file'\n" +
					"level='error' command='repair' destination='Music\\my artist\\my album 10\\pre-repair-backup\\2.mp3' error='oops' source='Music\\my artist\\my album 10\\2 my track 102.mp3' msg='error copying file'\n" +
					"level='error' command='repair' destination='Music\\my artist\\my album 10\\pre-repair-backup\\3.mp3' error='oops' source='Music\\my artist\\my album 10\\3 my track 103.mp3' msg='error copying file'\n" +
					"level='error' command='repair' destination='Music\\my artist\\my album 10\\pre-repair-backup\\4.mp3' error='oops' source='Music\\my artist\\my album 10\\4 my track 104.mp3' msg='error copying file'\n" +
					"level='error' command='repair' destination='Music\\my artist\\my album 11\\pre-repair-backup\\1.mp3' error='oops' source='Music\\my artist\\my album 11\\1 my track 111.mp3' msg='error copying file'\n" +
					"level='error' command='repair' destination='Music\\my artist\\my album 11\\pre-repair-backup\\2.mp3' error='oops' source='Music\\my artist\\my album 11\\2 my track 112.mp3' msg='error copying file'\n" +
					"level='error' command='repair' destination='Music\\my artist\\my album 11\\pre-repair-backup\\3.mp3' error='oops' source='Music\\my artist\\my album 11\\3 my track 113.mp3' msg='error copying file'\n" +
					"level='error' command='repair' destination='Music\\my artist\\my album 11\\pre-repair-backup\\4.mp3' error='oops' source='Music\\my artist\\my album 11\\4 my track 114.mp3' msg='error copying file'\n" +
					"level='error' command='repair' destination='Music\\my artist\\my album 12\\pre-repair-backup\\1.mp3' error='oops' source='Music\\my artist\\my album 12\\1 my track 121.mp3' msg='error copying file'\n" +
					"level='error' command='repair' destination='Music\\my artist\\my album 12\\pre-repair-backup\\2.mp3' error='oops' source='Music\\my artist\\my album 12\\2 my track 122.mp3' msg='error copying file'\n" +
					"level='error' command='repair' destination='Music\\my artist\\my album 12\\pre-repair-backup\\3.mp3' error='oops' source='Music\\my artist\\my album 12\\3 my track 123.mp3' msg='error copying file'\n" +
					"level='error' command='repair' destination='Music\\my artist\\my album 12\\pre-repair-backup\\4.mp3' error='oops' source='Music\\my artist\\my album 12\\4 my track 124.mp3' msg='error copying file'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			cmd.DirExists = tt.dirExists
			cmd.PlainFileExists = tt.plainFileExists
			cmd.CopyFile = tt.copyFile
			o := output.NewRecorder()
			if got := cmd.BackupAndFix(o, tt.checkedArtists); got != tt.wantStatus {
				t.Errorf("BackupAndFix() got %d want %d", got, tt.wantStatus)
			}
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
				t.SetMetadata(files.NewTrackMetadata().WithAlbumNames(
					[]string{"", "some other album", "some other album"}).WithArtistNames(
					[]string{"", "some other artist", "some other artist"}).WithGenres(
					[]string{"", "pop emo", "pop emo"}).WithMusicCDIdentifier(
					[]byte{1, 2, 3}).WithTrackNames(
					[]string{"", "some other title", "some other title"}).WithTrackNumbers(
					[]int{0, 99, 99}).WithYears(
					[]string{"", "2001", "2001"}).WithPrimarySource(files.ID3V1))
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
	originalReadMetadata := cmd.ReadMetadata
	originalDirExists := cmd.DirExists
	originalPlainFileExists := cmd.PlainFileExists
	originalCopyFile := cmd.CopyFile
	originalMarkDirty := cmd.MarkDirty
	defer func() {
		cmd.ReadMetadata = originalReadMetadata
		cmd.DirExists = originalDirExists
		cmd.PlainFileExists = originalPlainFileExists
		cmd.CopyFile = originalCopyFile
		cmd.MarkDirty = originalMarkDirty
	}()
	cmd.ReadMetadata = func(_ output.Bus, _ []*files.Artist) {}
	cmd.DirExists = func(_ string) bool { return true }
	cmd.PlainFileExists = func(_ string) bool { return false }
	cmd.CopyFile = func(_, _ string) error { return nil }
	cmd.MarkDirty = func(_ output.Bus) {}
	dirty := generateArtists(2, 3, 4)
	for _, aR := range dirty {
		for _, aL := range aR.Albums() {
			for _, t := range aL.Tracks() {
				t.SetMetadata(files.NewTrackMetadata().WithMusicCDIdentifier(
					[]byte{1, 2, 3}).WithTrackNumbers([]int{0, 99, 99}).WithPrimarySource(
					files.ID3V1))
			}
		}
	}
	tests := map[string]struct {
		rs         *cmd.RepairSettings
		artists    []*files.Artist
		wantStatus int
		output.WantedRecording
	}{
		"clean dry run": {
			rs:         cmd.NewRepairSettings().WithDryRun(true),
			artists:    generateArtists(2, 3, 4),
			wantStatus: cmd.Success,
			WantedRecording: output.WantedRecording{
				Console: "No repairable track defects were found.\n",
			},
		},
		"dirty dry run": {
			rs:         cmd.NewRepairSettings().WithDryRun(true),
			artists:    dirty,
			wantStatus: cmd.Success,
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
			rs:         cmd.NewRepairSettings().WithDryRun(false),
			artists:    generateArtists(2, 3, 4),
			wantStatus: cmd.Success,
			WantedRecording: output.WantedRecording{
				Console: "No repairable track defects were found.\n",
			},
		},
		"dirty repair": {
			rs:         cmd.NewRepairSettings().WithDryRun(false),
			artists:    dirty,
			wantStatus: cmd.SystemError,
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
			if got := tt.rs.RepairArtists(o, tt.artists); got != tt.wantStatus {
				t.Errorf("RepairSettings.RepairArtists() got %d want %d", got, tt.wantStatus)
			}
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("RepairSettings.RepairArtists() %s", issue)
				}
			}
		})
	}
}

func TestRepairSettings_ProcessArtists(t *testing.T) {
	originalReadMetadata := cmd.ReadMetadata
	defer func() {
		cmd.ReadMetadata = originalReadMetadata
	}()
	cmd.ReadMetadata = func(_ output.Bus, _ []*files.Artist) {}
	type args struct {
		allArtists []*files.Artist
		loaded     bool
		ss         *cmd.SearchSettings
	}
	tests := map[string]struct {
		rs *cmd.RepairSettings
		args
		wantStatus int
		output.WantedRecording
	}{
		"nothing to do": {
			rs:         cmd.NewRepairSettings().WithDryRun(true),
			args:       args{},
			wantStatus: cmd.UserError,
		},
		"clean artists": {
			rs: cmd.NewRepairSettings().WithDryRun(true),
			args: args{
				allArtists: generateArtists(2, 3, 4),
				loaded:     true,
				ss:         cmd.NewSearchSettings().WithArtistFilter(regexp.MustCompile(".*")).WithAlbumFilter(regexp.MustCompile(".*")).WithTrackFilter(regexp.MustCompile(".*")),
			},
			wantStatus: cmd.Success,
			WantedRecording: output.WantedRecording{
				Console: "No repairable track defects were found.\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			if got := tt.rs.ProcessArtists(o, tt.args.allArtists, tt.args.loaded, tt.args.ss); got != tt.wantStatus {
				t.Errorf("RepairSettings.ProcessArtists() got %d want %d", got, tt.wantStatus)
			}
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
	originalBus := cmd.Bus
	originalSearchFlags := cmd.SearchFlags
	originalExit := cmd.Exit
	defer func() {
		cmd.Bus = originalBus
		cmd.SearchFlags = originalSearchFlags
		cmd.Exit = originalExit
	}()
	var exitCode int
	var exitCalled bool
	cmd.Exit = func(code int) {
		exitCode = code
		exitCalled = true
	}
	cmd.SearchFlags = safeSearchFlags
	repairFlags := cmd.NewSectionFlags().WithSectionName("repair").WithFlags(
		map[string]*cmd.FlagDetails{
			"dryRun": cmd.NewFlagDetails().WithUsage("output what would have been repaired, but make no repairs").WithExpectedType(cmd.BoolType).WithDefaultValue(false),
		},
	)
	command := &cobra.Command{}
	cmd.AddFlags(output.NewNilBus(), cmd_toolkit.EmptyConfiguration(), command.Flags(), repairFlags, cmd.SearchFlags)
	tests := map[string]struct {
		cmd            *cobra.Command
		in1            []string
		wantExitCode   int
		wantExitCalled bool
		output.WantedRecording
	}{
		"basic": {
			cmd:            command,
			wantExitCode:   cmd.UserError,
			wantExitCalled: true,
			WantedRecording: output.WantedRecording{
				Error: "" +
					"No music files could be found using the specified parameters.\n" +
					"Why?\n" +
					"There were no directories found in \".\" (the --topDir value).\n" +
					"What to do:\n" +
					"Set --topDir to the path of a directory that contains artist directories.\n",
				Log: "" +
					"level='info'" +
					" --albumFilter='.*'" +
					" --artistFilter='.*'" +
					" --dryRun='false'" +
					" --extensions='[.mp3]'" +
					" --topDir='.'" +
					" --trackFilter='.*'" +
					" command='repair'" +
					" msg='executing command'\n" +
					"level='error'" +
					" --topDir='.'" +
					" msg='cannot find any artist directories'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			exitCode = -1
			exitCalled = false
			o := output.NewRecorder()
			cmd.Bus = o // cook getBus()
			cmd.RepairRun(tt.cmd, tt.in1)
			if got := exitCode; got != tt.wantExitCode {
				t.Errorf("RepairRun() got %d want %d", got, tt.wantExitCode)
			}
			if got := exitCalled; got != tt.wantExitCalled {
				t.Errorf("RepairRun() got %t want %t", got, tt.wantExitCalled)
			}
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("RepairRun() %s", issue)
				}
			}
		})
	}
}

func TestRepairHelp(t *testing.T) {
	originalSearchFlags := cmd.SearchFlags
	defer func() {
		cmd.SearchFlags = originalSearchFlags
	}()
	cmd.SearchFlags = safeSearchFlags
	commandUnderTest := cloneCommand(cmd.RepairCmd)
	cmd.AddFlags(output.NewNilBus(), cmd_toolkit.EmptyConfiguration(), commandUnderTest.Flags(), cmd.RepairFlags, cmd.SearchFlags)
	tests := map[string]struct {
		output.WantedRecording
	}{
		"good": {
			WantedRecording: output.WantedRecording{
				Console: "" +
					"\"repair\" repairs the problems found by running 'check --files'\n" +
					"\n" +
					"This command rewrites the mp3 files that the check command noted as having metadata\n" +
					"inconsistent with the file structure. Prior to rewriting an mp3 file, the repair\n" +
					"command creates a backup directory for the parent album and copies the original mp3\n" +
					"file into that backup directory. Use the postRepair command to automatically delete\n" +
					"the backup folders.\n" +
					"\n" +
					"Usage:\n" +
					"  repair [--dryRun] [--albumFilter regex] [--artistFilter regex] [--trackFilter regex] [--topDir dir] [--extensions extensions]\n" +
					"\n" +
					"Flags:\n" +
					"      --albumFilter string    regular expression specifying which albums to select (default \".*\")\n" +
					"      --artistFilter string   regular expression specifying which artists to select (default \".*\")\n" +
					"      --dryRun                output what would have been repaired, but make no repairs (default false)\n" +
					"      --extensions string     comma-delimited list of file extensions used by mp3 files (default \".mp3\")\n" +
					"      --topDir string         top directory specifying where to find mp3 files (default \".\")\n" +
					"      --trackFilter string    regular expression specifying which tracks to select (default \".*\")\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			command := commandUnderTest
			enableCommandRecording(o, command)
			command.Help()
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("repair Help() %s", issue)
				}
			}
		})
	}
}
