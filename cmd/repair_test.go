/*
Copyright © 2021 Marc Johnson (marc.johnson27591@gmail.com)
*/
package cmd

import (
	"fmt"
	"mp3repair/internal/files"
	"reflect"
	"regexp"
	"testing"

	cmdtoolkit "github.com/majohn-r/cmd-toolkit"
	"github.com/majohn-r/output"
	"github.com/spf13/cobra"
)

func Test_processRepairFlags(t *testing.T) {
	tests := map[string]struct {
		values map[string]*cmdtoolkit.CommandFlag[any]
		want   *repairSettings
		want1  bool
		output.WantedRecording
	}{
		"bad value": {
			values: map[string]*cmdtoolkit.CommandFlag[any]{},
			want:   &repairSettings{},
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
			values: map[string]*cmdtoolkit.CommandFlag[any]{"dryRun": {Value: true}},
			want:   &repairSettings{dryRun: cmdtoolkit.CommandFlag[bool]{Value: true}},
			want1:  true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			got, got1 := processRepairFlags(o, tt.values)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("processRepairFlags() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("processRepairFlags() got1 = %v, want %v", got1, tt.want1)
			}
			o.Report(t, "processRepairFlags()", tt.WantedRecording)
		})
	}
}

func Test_ensureTrackBackupDirectoryExists(t *testing.T) {
	originalDirExists := dirExists
	originalMkdir := mkdir
	defer func() {
		dirExists = originalDirExists
		mkdir = originalMkdir
	}()
	album := &files.Album{}
	if albums := generateAlbums(1, 5); len(albums) > 0 {
		album = albums[0]
	}
	tests := map[string]struct {
		cAl        *concernedAlbum
		dirExists  func(s string) bool
		mkdir      func(s string) error
		wantPath   string
		wantExists bool
		output.WantedRecording
	}{
		"dir already exists": {
			cAl:        newConcernedAlbum(album),
			dirExists:  func(_ string) bool { return true },
			wantPath:   album.BackupDirectory(),
			wantExists: true,
		},
		"dir does not exist but can be created": {
			cAl:        newConcernedAlbum(album),
			dirExists:  func(_ string) bool { return false },
			mkdir:      func(_ string) error { return nil },
			wantPath:   album.BackupDirectory(),
			wantExists: true,
		},
		"dir does not exist and cannot be created": {
			cAl:        newConcernedAlbum(album),
			dirExists:  func(_ string) bool { return false },
			mkdir:      func(_ string) error { return fmt.Errorf("plain file exists") },
			wantPath:   album.BackupDirectory(),
			wantExists: false,
			WantedRecording: output.WantedRecording{
				Error: "" +
					"The directory" +
					" \"Music\\\\my artist\\\\my album 00\\\\pre-repair-backup\"" +
					" cannot be created: plain file exists.\n" +
					"The track files in the directory" +
					" \"Music\\\\my artist\\\\my album 00\" will not be repaired.\n",
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
			dirExists = tt.dirExists
			mkdir = tt.mkdir
			o := output.NewRecorder()
			gotPath, gotExists := ensureTrackBackupDirectoryExists(o, tt.cAl)
			if gotPath != tt.wantPath {
				t.Errorf("ensureTrackBackupDirectoryExists() gotPath = %v, want %v", gotPath, tt.wantPath)
			}
			if gotExists != tt.wantExists {
				t.Errorf("ensureTrackBackupDirectoryExists() gotExists = %v, want %v", gotExists, tt.wantExists)
			}
			o.Report(t, "ensureTrackBackupDirectoryExists()", tt.WantedRecording)
		})
	}
}

func Test_tryTrackBackup(t *testing.T) {
	originalPlainFileExists := plainFileExists
	originalCopyFile := copyFile
	defer func() {
		plainFileExists = originalPlainFileExists
		copyFile = originalCopyFile
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
		copyFile        func(src, destination string) error
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
					"The backup file for track file" +
					" \"Music\\\\my artist\\\\my album 00\\\\1 my track 001.mp3\"," +
					" \"backupDir\\\\1.mp3\", already exists.\n" +
					"The track file " +
					"\"Music\\\\my artist\\\\my album 00\\\\1 my track 001.mp3\"" +
					" will not be repaired.\n",
				Log: "" +
					"level='error'" +
					" command='repair'" +
					" file='backupDir\\1.mp3'" +
					" msg='file already exists'\n",
			},
		},
		"backup does not exist but copy fails": {
			plainFileExists: func(_ string) bool { return false },
			copyFile: func(_, _ string) error {
				return fmt.Errorf("dir by that name exists")
			},
			args:         args{t: track, path: "backupDir"},
			wantBackedUp: false,
			WantedRecording: output.WantedRecording{
				Error: "" +
					"The track file" +
					" \"Music\\\\my artist\\\\my album 00\\\\1 my track 001.mp3\"" +
					" could not be backed up due to error dir by that name exists.\n" +
					"The track file" +
					" \"Music\\\\my artist\\\\my album 00\\\\1 my track 001.mp3\"" +
					" will not be repaired.\n",
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
				Console: "The track file" +
					" \"Music\\\\my artist\\\\my album 00\\\\1 my track 001.mp3\"" +
					" has been backed up to \"backupDir\\\\1.mp3\".\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			plainFileExists = tt.plainFileExists
			copyFile = tt.copyFile
			o := output.NewRecorder()
			gotBackedUp := tryTrackBackup(o, tt.args.t, tt.args.path)
			if gotBackedUp != tt.wantBackedUp {
				t.Errorf("tryTrackBackup() = %v, want %v", gotBackedUp, tt.wantBackedUp)
			}
			o.Report(t, "tryTrackBackup()", tt.WantedRecording)
		})
	}
}

func Test_processTrackRepairResults(t *testing.T) {
	originalMarkDirty := markDirty
	defer func() {
		markDirty = originalMarkDirty
	}()
	var markedDirty bool
	markDirty = func(o output.Bus) {
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
		wantStatus *cmdtoolkit.ExitError
		output.WantedRecording
	}{
		"success": {
			args:       args{t: track, err: nil},
			wantDirty:  true,
			wantStatus: nil,
			WantedRecording: output.WantedRecording{
				Console: "\"Music\\\\my artist\\\\my album 00\\\\1 my track 001.mp3\"" +
					" repaired.\n",
			},
		},
		"single failure": {
			args:       args{t: track, err: []error{fmt.Errorf("file locked")}},
			wantDirty:  false,
			wantStatus: cmdtoolkit.NewExitSystemError("repair"),
			WantedRecording: output.WantedRecording{
				Error: "An error occurred repairing track" +
					" \"Music\\\\my artist\\\\my album 00\\\\1 my track 001.mp3\".\n",
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
			wantStatus: cmdtoolkit.NewExitSystemError("repair"),
			WantedRecording: output.WantedRecording{
				Error: "An error occurred repairing track" +
					" \"Music\\\\my artist\\\\my album 00\\\\1 my track 001.mp3\".\n",
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
			if got := processTrackRepairResults(o, tt.args.t, tt.args.err); !compareExitErrors(got, tt.wantStatus) {
				t.Errorf("processTrackRepairResults() got %s want %s", got, tt.wantStatus)
			}
			if got := markedDirty; got != tt.wantDirty {
				t.Errorf("processTrackRepairResults() got %t want %t", got, tt.wantDirty)
			}
			o.Report(t, "processTrackRepairResults()", tt.WantedRecording)
		})
	}
}

func Test_backupAndRepairTracks(t *testing.T) {
	concernedArtists := createConcernedArtists(generateArtists(2, 3, 4, nil))
	skipArtist := true
	skipAlbum := true
	skipTrack := true
	for _, cAr := range concernedArtists {
		if skipArtist {
			skipArtist = false
			continue
		}
		for _, cAl := range cAr.albums() {
			if skipAlbum {
				skipAlbum = false
				continue
			}
			for _, cT := range cAl.tracks() {
				if skipTrack {
					skipTrack = false
					continue
				}
				cT.addConcern(conflictConcern,
					"artist field does not match artist name")
			}
		}
	}
	originalDirExists := dirExists
	originalPlainFileExists := plainFileExists
	originalCopyFile := copyFile
	defer func() {
		dirExists = originalDirExists
		plainFileExists = originalPlainFileExists
		copyFile = originalCopyFile
	}()
	tests := map[string]struct {
		dirExists        func(string) bool
		plainFileExists  func(string) bool
		copyFile         func(string, string) error
		concernedArtists []*concernedArtist
		wantStatus       *cmdtoolkit.ExitError
		output.WantedRecording
	}{
		"basic test": {
			dirExists:        func(_ string) bool { return true },
			plainFileExists:  func(_ string) bool { return false },
			copyFile:         func(_, _ string) error { return nil },
			concernedArtists: concernedArtists,
			wantStatus:       cmdtoolkit.NewExitSystemError("repair"),
			WantedRecording: output.WantedRecording{
				Console: "" +
					"The track file" +
					" \"Music\\\\my artist\\\\my album 11\\\\2 my track 112.mp3\"" +
					" has been backed up to" +
					" \"Music\\\\my artist\\\\my album 11\\\\pre-repair-backup\\\\2.mp3\".\n" +
					"The track file" +
					" \"Music\\\\my artist\\\\my album 11\\\\3 my track 113.mp3\"" +
					" has been backed up to" +
					" \"Music\\\\my artist\\\\my album 11\\\\pre-repair-backup\\\\3.mp3\".\n" +
					"The track file" +
					" \"Music\\\\my artist\\\\my album 11\\\\4 my track 114.mp3\"" +
					" has been backed up to" +
					" \"Music\\\\my artist\\\\my album 11\\\\pre-repair-backup\\\\4.mp3\".\n" +
					"The track file" +
					" \"Music\\\\my artist\\\\my album 12\\\\1 my track 121.mp3\"" +
					" has been backed up to" +
					" \"Music\\\\my artist\\\\my album 12\\\\pre-repair-backup\\\\1.mp3\".\n" +
					"The track file" +
					" \"Music\\\\my artist\\\\my album 12\\\\2 my track 122.mp3\"" +
					" has been backed up to" +
					" \"Music\\\\my artist\\\\my album 12\\\\pre-repair-backup\\\\2.mp3\".\n" +
					"The track file" +
					" \"Music\\\\my artist\\\\my album 12\\\\3 my track 123.mp3\"" +
					" has been backed up to" +
					" \"Music\\\\my artist\\\\my album 12\\\\pre-repair-backup\\\\3.mp3\".\n" +
					"The track file" +
					" \"Music\\\\my artist\\\\my album 12\\\\4 my track 124.mp3\"" +
					" has been backed up to" +
					" \"Music\\\\my artist\\\\my album 12\\\\pre-repair-backup\\\\4.mp3\".\n",
				Error: "" +
					"An error occurred repairing track" +
					" \"Music\\\\my artist\\\\my album 11\\\\2 my track 112.mp3\".\n" +
					"An error occurred repairing track" +
					" \"Music\\\\my artist\\\\my album 11\\\\3 my track 113.mp3\".\n" +
					"An error occurred repairing track" +
					" \"Music\\\\my artist\\\\my album 11\\\\4 my track 114.mp3\".\n" +
					"An error occurred repairing track" +
					" \"Music\\\\my artist\\\\my album 12\\\\1 my track 121.mp3\".\n" +
					"An error occurred repairing track" +
					" \"Music\\\\my artist\\\\my album 12\\\\2 my track 122.mp3\".\n" +
					"An error occurred repairing track" +
					" \"Music\\\\my artist\\\\my album 12\\\\3 my track 123.mp3\".\n" +
					"An error occurred repairing track" +
					" \"Music\\\\my artist\\\\my album 12\\\\4 my track 124.mp3\".\n",
				Log: "" +
					"level='error'" +
					" command='repair'" +
					" directory='Music\\my artist\\my album 11'" +
					" error='[\"no edit required\"]'" +
					" fileName='2 my track 112.mp3'" +
					" msg='cannot edit track'\n" +
					"level='error'" +
					" command='repair'" +
					" directory='Music\\my artist\\my album 11'" +
					" error='[\"no edit required\"]'" +
					" fileName='3 my track 113.mp3'" +
					" msg='cannot edit track'\n" +
					"level='error'" +
					" command='repair'" +
					" directory='Music\\my artist\\my album 11'" +
					" error='[\"no edit required\"]'" +
					" fileName='4 my track 114.mp3'" +
					" msg='cannot edit track'\n" +
					"level='error'" +
					" command='repair'" +
					" directory='Music\\my artist\\my album 12'" +
					" error='[\"no edit required\"]'" +
					" fileName='1 my track 121.mp3'" +
					" msg='cannot edit track'\n" +
					"level='error'" +
					" command='repair'" +
					" directory='Music\\my artist\\my album 12'" +
					" error='[\"no edit required\"]'" +
					" fileName='2 my track 122.mp3'" +
					" msg='cannot edit track'\n" +
					"level='error'" +
					" command='repair'" +
					" directory='Music\\my artist\\my album 12'" +
					" error='[\"no edit required\"]'" +
					" fileName='3 my track 123.mp3'" +
					" msg='cannot edit track'\n" +
					"level='error'" +
					" command='repair'" +
					" directory='Music\\my artist\\my album 12'" +
					" error='[\"no edit required\"]'" +
					" fileName='4 my track 124.mp3'" +
					" msg='cannot edit track'\n",
			},
		},
		"basic test2": {
			dirExists:        func(_ string) bool { return false },
			plainFileExists:  func(_ string) bool { return false },
			copyFile:         func(_, _ string) error { return nil },
			concernedArtists: concernedArtists,
			wantStatus:       cmdtoolkit.NewExitSystemError("repair"),
			WantedRecording: output.WantedRecording{
				Error: "" +
					"The directory" +
					" \"Music\\\\my artist\\\\my album 11\\\\pre-repair-backup\"" +
					" cannot be created: parent directory is not a directory.\n" +
					"The track files in the directory" +
					" \"Music\\\\my artist\\\\my album 11\" will not be repaired.\n" +
					"The directory" +
					" \"Music\\\\my artist\\\\my album 12\\\\pre-repair-backup\"" +
					" cannot be created: parent directory is not a directory.\n" +
					"The track files in the directory" +
					" \"Music\\\\my artist\\\\my album 12\" will not be repaired.\n",
				Log: "" +
					"level='error'" +
					" command='repair'" +
					" directory='Music\\my artist\\my album 11\\pre-repair-backup'" +
					" error='parent directory is not a directory'" +
					" msg='cannot create directory'\n" +
					"level='error'" +
					" command='repair'" +
					" directory='Music\\my artist\\my album 12\\pre-repair-backup'" +
					" error='parent directory is not a directory'" +
					" msg='cannot create directory'\n",
			},
		},
		"basic test3": {
			dirExists:        func(_ string) bool { return true },
			plainFileExists:  func(_ string) bool { return false },
			copyFile:         func(_, _ string) error { return fmt.Errorf("oops") },
			concernedArtists: concernedArtists,
			wantStatus:       cmdtoolkit.NewExitSystemError("repair"),
			WantedRecording: output.WantedRecording{
				Error: "" +
					"The track file" +
					" \"Music\\\\my artist\\\\my album 11\\\\2 my track 112.mp3\"" +
					" could not be backed up due to error oops.\n" +
					"The track file" +
					" \"Music\\\\my artist\\\\my album 11\\\\2 my track 112.mp3\"" +
					" will not be repaired.\n" +
					"The track file" +
					" \"Music\\\\my artist\\\\my album 11\\\\3 my track 113.mp3\"" +
					" could not be backed up due to error oops.\n" +
					"The track file" +
					" \"Music\\\\my artist\\\\my album 11\\\\3 my track 113.mp3\"" +
					" will not be repaired.\n" +
					"The track file" +
					" \"Music\\\\my artist\\\\my album 11\\\\4 my track 114.mp3\"" +
					" could not be backed up due to error oops.\n" +
					"The track file" +
					" \"Music\\\\my artist\\\\my album 11\\\\4 my track 114.mp3\"" +
					" will not be repaired.\n" +
					"The track file" +
					" \"Music\\\\my artist\\\\my album 12\\\\1 my track 121.mp3\"" +
					" could not be backed up due to error oops.\n" +
					"The track file" +
					" \"Music\\\\my artist\\\\my album 12\\\\1 my track 121.mp3\"" +
					" will not be repaired.\n" +
					"The track file" +
					" \"Music\\\\my artist\\\\my album 12\\\\2 my track 122.mp3\"" +
					" could not be backed up due to error oops.\n" +
					"The track file" +
					" \"Music\\\\my artist\\\\my album 12\\\\2 my track 122.mp3\"" +
					" will not be repaired.\n" +
					"The track file" +
					" \"Music\\\\my artist\\\\my album 12\\\\3 my track 123.mp3\"" +
					" could not be backed up due to error oops.\n" +
					"The track file" +
					" \"Music\\\\my artist\\\\my album 12\\\\3 my track 123.mp3\"" +
					" will not be repaired.\n" +
					"The track file" +
					" \"Music\\\\my artist\\\\my album 12\\\\4 my track 124.mp3\"" +
					" could not be backed up due to error oops.\n" +
					"The track file" +
					" \"Music\\\\my artist\\\\my album 12\\\\4 my track 124.mp3\"" +
					" will not be repaired.\n",
				Log: "" +
					"level='error'" +
					" command='repair'" +
					" destination='Music\\my artist\\my album 11\\pre-repair-backup\\2.mp3'" +
					" error='oops'" +
					" source='Music\\my artist\\my album 11\\2 my track 112.mp3'" +
					" msg='error copying file'\n" +
					"level='error'" +
					" command='repair'" +
					" destination='Music\\my artist\\my album 11\\pre-repair-backup\\3.mp3'" +
					" error='oops'" +
					" source='Music\\my artist\\my album 11\\3 my track 113.mp3'" +
					" msg='error copying file'\n" +
					"level='error'" +
					" command='repair'" +
					" destination='Music\\my artist\\my album 11\\pre-repair-backup\\4.mp3'" +
					" error='oops'" +
					" source='Music\\my artist\\my album 11\\4 my track 114.mp3'" +
					" msg='error copying file'\n" +
					"level='error'" +
					" command='repair'" +
					" destination='Music\\my artist\\my album 12\\pre-repair-backup\\1.mp3'" +
					" error='oops'" +
					" source='Music\\my artist\\my album 12\\1 my track 121.mp3'" +
					" msg='error copying file'\n" +
					"level='error'" +
					" command='repair'" +
					" destination='Music\\my artist\\my album 12\\pre-repair-backup\\2.mp3'" +
					" error='oops'" +
					" source='Music\\my artist\\my album 12\\2 my track 122.mp3'" +
					" msg='error copying file'\n" +
					"level='error'" +
					" command='repair'" +
					" destination='Music\\my artist\\my album 12\\pre-repair-backup\\3.mp3'" +
					" error='oops'" +
					" source='Music\\my artist\\my album 12\\3 my track 123.mp3'" +
					" msg='error copying file'\n" +
					"level='error'" +
					" command='repair'" +
					" destination='Music\\my artist\\my album 12\\pre-repair-backup\\4.mp3'" +
					" error='oops'" +
					" source='Music\\my artist\\my album 12\\4 my track 124.mp3'" +
					" msg='error copying file'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			dirExists = tt.dirExists
			plainFileExists = tt.plainFileExists
			copyFile = tt.copyFile
			o := output.NewRecorder()
			if got := backupAndRepairTracks(o, tt.concernedArtists); !compareExitErrors(got, tt.wantStatus) {
				t.Errorf("backupAndRepairTracks() got %s want %s", got, tt.wantStatus)
			}
			o.Report(t, "backupAndRepairTracks()", tt.WantedRecording)
		})
	}
}

func Test_reportRepairsNeeded(t *testing.T) {
	dirty := createConcernedArtists(generateArtists(2, 3, 4, nil))
	for _, cAr := range dirty {
		for _, cAl := range cAr.albums() {
			for _, cT := range cAl.tracks() {
				cT.addConcern(conflictConcern, "artist field does not match artist name")
			}
		}
	}
	clean := createConcernedArtists(generateArtists(2, 3, 4, nil))
	tests := map[string]struct {
		concernedArtists []*concernedArtist
		output.WantedRecording
	}{
		"clean": {
			concernedArtists: clean,
			WantedRecording: output.WantedRecording{
				Console: "No repairable track defects were found.\n",
			},
		},
		"dirty": {
			concernedArtists: dirty,
			WantedRecording: output.WantedRecording{
				Console: "" +
					"The following concerns can be repaired:\n" +
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
			reportRepairsNeeded(o, tt.concernedArtists)
			o.Report(t, "reportRepairsNeeded()", tt.WantedRecording)
		})
	}
}

func Test_findConflictedTracks(t *testing.T) {
	tm := files.NewTrackMetadata()
	for _, src := range []files.SourceType{files.ID3V1, files.ID3V2} {
		tm.SetArtistName(src, "some other artist")
		tm.SetAlbumName(src, "some other album")
		tm.SetAlbumGenre(src, "pop emo")
		tm.SetAlbumYear(src, "2001")
		tm.SetTrackName(src, "some other title")
		tm.SetTrackNumber(src, 99)
	}
	tm.SetCDIdentifier([]byte{1, 2, 3})
	tm.SetCanonicalSource(files.ID3V1)
	dirty := createConcernedArtists(generateArtists(2, 3, 4, tm))
	clean := createConcernedArtists(generateArtists(2, 3, 4, nil))
	tests := map[string]struct {
		concernedArtists []*concernedArtist
		want             int
	}{
		"clean": {concernedArtists: clean, want: 0},
		"dirty": {concernedArtists: dirty, want: 24},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := findConflictedTracks(tt.concernedArtists); got != tt.want {
				t.Errorf("findConflictedTracks() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_repairSettings_repairArtists(t *testing.T) {
	originalReadMetadata := readMetadata
	originalDirExists := dirExists
	originalPlainFileExists := plainFileExists
	originalCopyFile := copyFile
	originalMarkDirty := markDirty
	defer func() {
		readMetadata = originalReadMetadata
		dirExists = originalDirExists
		plainFileExists = originalPlainFileExists
		copyFile = originalCopyFile
		markDirty = originalMarkDirty
	}()
	readMetadata = func(_ output.Bus, _ []*files.Artist) {}
	dirExists = func(_ string) bool { return true }
	plainFileExists = func(_ string) bool { return false }
	copyFile = func(_, _ string) error { return nil }
	markDirty = func(_ output.Bus) {}
	tm := files.NewTrackMetadata()
	tm.SetTrackNumber(files.ID3V1, 99)
	tm.SetTrackNumber(files.ID3V2, 99)
	tm.SetCDIdentifier([]byte{1, 2, 3})
	tm.SetCanonicalSource(files.ID3V1)
	dirty := generateArtists(2, 3, 4, tm)
	tests := map[string]struct {
		rs         *repairSettings
		artists    []*files.Artist
		wantStatus *cmdtoolkit.ExitError
		output.WantedRecording
	}{
		"clean dry run": {
			rs:         &repairSettings{dryRun: cmdtoolkit.CommandFlag[bool]{Value: true}},
			artists:    generateArtists(2, 3, 4, nil),
			wantStatus: nil,
			WantedRecording: output.WantedRecording{
				Console: "No repairable track defects were found.\n",
			},
		},
		"dirty dry run": {
			rs:         &repairSettings{dryRun: cmdtoolkit.CommandFlag[bool]{Value: true}},
			artists:    dirty,
			wantStatus: nil,
			WantedRecording: output.WantedRecording{
				Console: "" +
					"The following concerns can be repaired:\n" +
					"Artist \"my artist 0\"\n" +
					"  Album \"my album 00\"\n" +
					"    Track \"my track 001\"\n" +
					"    * [metadata conflict]" +
					" the album name field does not match the name of the album directory\n" +
					"    * [metadata conflict]" +
					" the artist name field does not match the name of the artist directory\n" +
					"    * [metadata conflict]" +
					" the music CD identifier field does not match the other tracks in the album\n" +
					"    * [metadata conflict]" +
					" the track name field does not match the track's file name\n" +
					"    * [metadata conflict]" +
					" the track number field does not match the track's file name\n" +
					"    Track \"my track 002\"\n" +
					"    * [metadata conflict]" +
					" the album name field does not match the name of the album directory\n" +
					"    * [metadata conflict]" +
					" the artist name field does not match the name of the artist directory\n" +
					"    * [metadata conflict]" +
					" the music CD identifier field does not match the other tracks in the album\n" +
					"    * [metadata conflict]" +
					" the track name field does not match the track's file name\n" +
					"    * [metadata conflict]" +
					" the track number field does not match the track's file name\n" +
					"    Track \"my track 003\"\n" +
					"    * [metadata conflict]" +
					" the album name field does not match the name of the album directory\n" +
					"    * [metadata conflict]" +
					" the artist name field does not match the name of the artist directory\n" +
					"    * [metadata conflict]" +
					" the music CD identifier field does not match the other tracks in the album\n" +
					"    * [metadata conflict]" +
					" the track name field does not match the track's file name\n" +
					"    * [metadata conflict]" +
					" the track number field does not match the track's file name\n" +
					"    Track \"my track 004\"\n" +
					"    * [metadata conflict]" +
					" the album name field does not match the name of the album directory\n" +
					"    * [metadata conflict]" +
					" the artist name field does not match the name of the artist directory\n" +
					"    * [metadata conflict]" +
					" the music CD identifier field does not match the other tracks in the album\n" +
					"    * [metadata conflict]" +
					" the track name field does not match the track's file name\n" +
					"    * [metadata conflict]" +
					" the track number field does not match the track's file name\n" +
					"  Album \"my album 01\"\n" +
					"    Track \"my track 011\"\n" +
					"    * [metadata conflict]" +
					" the album name field does not match the name of the album directory\n" +
					"    * [metadata conflict]" +
					" the artist name field does not match the name of the artist directory\n" +
					"    * [metadata conflict]" +
					" the music CD identifier field does not match the other tracks in the album\n" +
					"    * [metadata conflict]" +
					" the track name field does not match the track's file name\n" +
					"    * [metadata conflict]" +
					" the track number field does not match the track's file name\n" +
					"    Track \"my track 012\"\n" +
					"    * [metadata conflict]" +
					" the album name field does not match the name of the album directory\n" +
					"    * [metadata conflict]" +
					" the artist name field does not match the name of the artist directory\n" +
					"    * [metadata conflict]" +
					" the music CD identifier field does not match the other tracks in the album\n" +
					"    * [metadata conflict]" +
					" the track name field does not match the track's file name\n" +
					"    * [metadata conflict]" +
					" the track number field does not match the track's file name\n" +
					"    Track \"my track 013\"\n" +
					"    * [metadata conflict]" +
					" the album name field does not match the name of the album directory\n" +
					"    * [metadata conflict]" +
					" the artist name field does not match the name of the artist directory\n" +
					"    * [metadata conflict]" +
					" the music CD identifier field does not match the other tracks in the album\n" +
					"    * [metadata conflict]" +
					" the track name field does not match the track's file name\n" +
					"    * [metadata conflict]" +
					" the track number field does not match the track's file name\n" +
					"    Track \"my track 014\"\n" +
					"    * [metadata conflict]" +
					" the album name field does not match the name of the album directory\n" +
					"    * [metadata conflict]" +
					" the artist name field does not match the name of the artist directory\n" +
					"    * [metadata conflict]" +
					" the music CD identifier field does not match the other tracks in the album\n" +
					"    * [metadata conflict]" +
					" the track name field does not match the track's file name\n" +
					"    * [metadata conflict]" +
					" the track number field does not match the track's file name\n" +
					"  Album \"my album 02\"\n" +
					"    Track \"my track 021\"\n" +
					"    * [metadata conflict]" +
					" the album name field does not match the name of the album directory\n" +
					"    * [metadata conflict]" +
					" the artist name field does not match the name of the artist directory\n" +
					"    * [metadata conflict]" +
					" the music CD identifier field does not match the other tracks in the album\n" +
					"    * [metadata conflict]" +
					" the track name field does not match the track's file name\n" +
					"    * [metadata conflict]" +
					" the track number field does not match the track's file name\n" +
					"    Track \"my track 022\"\n" +
					"    * [metadata conflict]" +
					" the album name field does not match the name of the album directory\n" +
					"    * [metadata conflict]" +
					" the artist name field does not match the name of the artist directory\n" +
					"    * [metadata conflict]" +
					" the music CD identifier field does not match the other tracks in the album\n" +
					"    * [metadata conflict]" +
					" the track name field does not match the track's file name\n" +
					"    * [metadata conflict]" +
					" the track number field does not match the track's file name\n" +
					"    Track \"my track 023\"\n" +
					"    * [metadata conflict]" +
					" the album name field does not match the name of the album directory\n" +
					"    * [metadata conflict]" +
					" the artist name field does not match the name of the artist directory\n" +
					"    * [metadata conflict]" +
					" the music CD identifier field does not match the other tracks in the album\n" +
					"    * [metadata conflict]" +
					" the track name field does not match the track's file name\n" +
					"    * [metadata conflict]" +
					" the track number field does not match the track's file name\n" +
					"    Track \"my track 024\"\n" +
					"    * [metadata conflict]" +
					" the album name field does not match the name of the album directory\n" +
					"    * [metadata conflict]" +
					" the artist name field does not match the name of the artist directory\n" +
					"    * [metadata conflict]" +
					" the music CD identifier field does not match the other tracks in the album\n" +
					"    * [metadata conflict]" +
					" the track name field does not match the track's file name\n" +
					"    * [metadata conflict]" +
					" the track number field does not match the track's file name\n" +
					"Artist \"my artist 1\"\n" +
					"  Album \"my album 10\"\n" +
					"    Track \"my track 101\"\n" +
					"    * [metadata conflict]" +
					" the album name field does not match the name of the album directory\n" +
					"    * [metadata conflict]" +
					" the artist name field does not match the name of the artist directory\n" +
					"    * [metadata conflict]" +
					" the music CD identifier field does not match the other tracks in the album\n" +
					"    * [metadata conflict]" +
					" the track name field does not match the track's file name\n" +
					"    * [metadata conflict]" +
					" the track number field does not match the track's file name\n" +
					"    Track \"my track 102\"\n" +
					"    * [metadata conflict]" +
					" the album name field does not match the name of the album directory\n" +
					"    * [metadata conflict]" +
					" the artist name field does not match the name of the artist directory\n" +
					"    * [metadata conflict]" +
					" the music CD identifier field does not match the other tracks in the album\n" +
					"    * [metadata conflict]" +
					" the track name field does not match the track's file name\n" +
					"    * [metadata conflict]" +
					" the track number field does not match the track's file name\n" +
					"    Track \"my track 103\"\n" +
					"    * [metadata conflict]" +
					" the album name field does not match the name of the album directory\n" +
					"    * [metadata conflict]" +
					" the artist name field does not match the name of the artist directory\n" +
					"    * [metadata conflict]" +
					" the music CD identifier field does not match the other tracks in the album\n" +
					"    * [metadata conflict]" +
					" the track name field does not match the track's file name\n" +
					"    * [metadata conflict]" +
					" the track number field does not match the track's file name\n" +
					"    Track \"my track 104\"\n" +
					"    * [metadata conflict]" +
					" the album name field does not match the name of the album directory\n" +
					"    * [metadata conflict]" +
					" the artist name field does not match the name of the artist directory\n" +
					"    * [metadata conflict]" +
					" the music CD identifier field does not match the other tracks in the album\n" +
					"    * [metadata conflict]" +
					" the track name field does not match the track's file name\n" +
					"    * [metadata conflict]" +
					" the track number field does not match the track's file name\n" +
					"  Album \"my album 11\"\n" +
					"    Track \"my track 111\"\n" +
					"    * [metadata conflict]" +
					" the album name field does not match the name of the album directory\n" +
					"    * [metadata conflict]" +
					" the artist name field does not match the name of the artist directory\n" +
					"    * [metadata conflict]" +
					" the music CD identifier field does not match the other tracks in the album\n" +
					"    * [metadata conflict]" +
					" the track name field does not match the track's file name\n" +
					"    * [metadata conflict]" +
					" the track number field does not match the track's file name\n" +
					"    Track \"my track 112\"\n" +
					"    * [metadata conflict]" +
					" the album name field does not match the name of the album directory\n" +
					"    * [metadata conflict]" +
					" the artist name field does not match the name of the artist directory\n" +
					"    * [metadata conflict]" +
					" the music CD identifier field does not match the other tracks in the album\n" +
					"    * [metadata conflict]" +
					" the track name field does not match the track's file name\n" +
					"    * [metadata conflict]" +
					" the track number field does not match the track's file name\n" +
					"    Track \"my track 113\"\n" +
					"    * [metadata conflict]" +
					" the album name field does not match the name of the album directory\n" +
					"    * [metadata conflict]" +
					" the artist name field does not match the name of the artist directory\n" +
					"    * [metadata conflict]" +
					" the music CD identifier field does not match the other tracks in the album\n" +
					"    * [metadata conflict]" +
					" the track name field does not match the track's file name\n" +
					"    * [metadata conflict]" +
					" the track number field does not match the track's file name\n" +
					"    Track \"my track 114\"\n" +
					"    * [metadata conflict]" +
					" the album name field does not match the name of the album directory\n" +
					"    * [metadata conflict]" +
					" the artist name field does not match the name of the artist directory\n" +
					"    * [metadata conflict]" +
					" the music CD identifier field does not match the other tracks in the album\n" +
					"    * [metadata conflict]" +
					" the track name field does not match the track's file name\n" +
					"    * [metadata conflict]" +
					" the track number field does not match the track's file name\n" +
					"  Album \"my album 12\"\n" +
					"    Track \"my track 121\"\n" +
					"    * [metadata conflict]" +
					" the album name field does not match the name of the album directory\n" +
					"    * [metadata conflict]" +
					" the artist name field does not match the name of the artist directory\n" +
					"    * [metadata conflict]" +
					" the music CD identifier field does not match the other tracks in the album\n" +
					"    * [metadata conflict]" +
					" the track name field does not match the track's file name\n" +
					"    * [metadata conflict]" +
					" the track number field does not match the track's file name\n" +
					"    Track \"my track 122\"\n" +
					"    * [metadata conflict]" +
					" the album name field does not match the name of the album directory\n" +
					"    * [metadata conflict]" +
					" the artist name field does not match the name of the artist directory\n" +
					"    * [metadata conflict]" +
					" the music CD identifier field does not match the other tracks in the album\n" +
					"    * [metadata conflict]" +
					" the track name field does not match the track's file name\n" +
					"    * [metadata conflict]" +
					" the track number field does not match the track's file name\n" +
					"    Track \"my track 123\"\n" +
					"    * [metadata conflict]" +
					" the album name field does not match the name of the album directory\n" +
					"    * [metadata conflict]" +
					" the artist name field does not match the name of the artist directory\n" +
					"    * [metadata conflict]" +
					" the music CD identifier field does not match the other tracks in the album\n" +
					"    * [metadata conflict]" +
					" the track name field does not match the track's file name\n" +
					"    * [metadata conflict]" +
					" the track number field does not match the track's file name\n" +
					"    Track \"my track 124\"\n" +
					"    * [metadata conflict]" +
					" the album name field does not match the name of the album directory\n" +
					"    * [metadata conflict]" +
					" the artist name field does not match the name of the artist directory\n" +
					"    * [metadata conflict]" +
					" the music CD identifier field does not match the other tracks in the album\n" +
					"    * [metadata conflict]" +
					" the track name field does not match the track's file name\n" +
					"    * [metadata conflict]" +
					" the track number field does not match the track's file name\n",
			},
		},
		"clean repair": {
			rs:         &repairSettings{dryRun: cmdtoolkit.CommandFlag[bool]{Value: false}},
			artists:    generateArtists(2, 3, 4, nil),
			wantStatus: nil,
			WantedRecording: output.WantedRecording{
				Console: "No repairable track defects were found.\n",
			},
		},
		"dirty repair": {
			rs:         &repairSettings{dryRun: cmdtoolkit.CommandFlag[bool]{Value: false}},
			artists:    dirty,
			wantStatus: cmdtoolkit.NewExitSystemError("repair"),
			WantedRecording: output.WantedRecording{
				Console: "" +
					"The track file" +
					" \"Music\\\\my artist\\\\my album 00\\\\1 my track 001.mp3\"" +
					" has been backed up to" +
					" \"Music\\\\my artist\\\\my album 00\\\\pre-repair-backup\\\\1.mp3\".\n" +
					"The track file" +
					" \"Music\\\\my artist\\\\my album 00\\\\2 my track 002.mp3\"" +
					" has been backed up to" +
					" \"Music\\\\my artist\\\\my album 00\\\\pre-repair-backup\\\\2.mp3\".\n" +
					"The track file" +
					" \"Music\\\\my artist\\\\my album 00\\\\3 my track 003.mp3\"" +
					" has been backed up to" +
					" \"Music\\\\my artist\\\\my album 00\\\\pre-repair-backup\\\\3.mp3\".\n" +
					"The track file" +
					" \"Music\\\\my artist\\\\my album 00\\\\4 my track 004.mp3\"" +
					" has been backed up to" +
					" \"Music\\\\my artist\\\\my album 00\\\\pre-repair-backup\\\\4.mp3\".\n" +
					"The track file" +
					" \"Music\\\\my artist\\\\my album 01\\\\1 my track 011.mp3\"" +
					" has been backed up to" +
					" \"Music\\\\my artist\\\\my album 01\\\\pre-repair-backup\\\\1.mp3\".\n" +
					"The track file" +
					" \"Music\\\\my artist\\\\my album 01\\\\2 my track 012.mp3\"" +
					" has been backed up to" +
					" \"Music\\\\my artist\\\\my album 01\\\\pre-repair-backup\\\\2.mp3\".\n" +
					"The track file" +
					" \"Music\\\\my artist\\\\my album 01\\\\3 my track 013.mp3\"" +
					" has been backed up to" +
					" \"Music\\\\my artist\\\\my album 01\\\\pre-repair-backup\\\\3.mp3\".\n" +
					"The track file" +
					" \"Music\\\\my artist\\\\my album 01\\\\4 my track 014.mp3\"" +
					" has been backed up to" +
					" \"Music\\\\my artist\\\\my album 01\\\\pre-repair-backup\\\\4.mp3\".\n" +
					"The track file" +
					" \"Music\\\\my artist\\\\my album 02\\\\1 my track 021.mp3\"" +
					" has been backed up to" +
					" \"Music\\\\my artist\\\\my album 02\\\\pre-repair-backup\\\\1.mp3\".\n" +
					"The track file" +
					" \"Music\\\\my artist\\\\my album 02\\\\2 my track 022.mp3\"" +
					" has been backed up to" +
					" \"Music\\\\my artist\\\\my album 02\\\\pre-repair-backup\\\\2.mp3\".\n" +
					"The track file" +
					" \"Music\\\\my artist\\\\my album 02\\\\3 my track 023.mp3\"" +
					" has been backed up to" +
					" \"Music\\\\my artist\\\\my album 02\\\\pre-repair-backup\\\\3.mp3\".\n" +
					"The track file" +
					" \"Music\\\\my artist\\\\my album 02\\\\4 my track 024.mp3\"" +
					" has been backed up to" +
					" \"Music\\\\my artist\\\\my album 02\\\\pre-repair-backup\\\\4.mp3\".\n" +
					"The track file" +
					" \"Music\\\\my artist\\\\my album 10\\\\1 my track 101.mp3\"" +
					" has been backed up to" +
					" \"Music\\\\my artist\\\\my album 10\\\\pre-repair-backup\\\\1.mp3\".\n" +
					"The track file" +
					" \"Music\\\\my artist\\\\my album 10\\\\2 my track 102.mp3\"" +
					" has been backed up to" +
					" \"Music\\\\my artist\\\\my album 10\\\\pre-repair-backup\\\\2.mp3\".\n" +
					"The track file" +
					" \"Music\\\\my artist\\\\my album 10\\\\3 my track 103.mp3\"" +
					" has been backed up to" +
					" \"Music\\\\my artist\\\\my album 10\\\\pre-repair-backup\\\\3.mp3\".\n" +
					"The track file" +
					" \"Music\\\\my artist\\\\my album 10\\\\4 my track 104.mp3\"" +
					" has been backed up to" +
					" \"Music\\\\my artist\\\\my album 10\\\\pre-repair-backup\\\\4.mp3\".\n" +
					"The track file" +
					" \"Music\\\\my artist\\\\my album 11\\\\1 my track 111.mp3\"" +
					" has been backed up to" +
					" \"Music\\\\my artist\\\\my album 11\\\\pre-repair-backup\\\\1.mp3\".\n" +
					"The track file" +
					" \"Music\\\\my artist\\\\my album 11\\\\2 my track 112.mp3\"" +
					" has been backed up to" +
					" \"Music\\\\my artist\\\\my album 11\\\\pre-repair-backup\\\\2.mp3\".\n" +
					"The track file" +
					" \"Music\\\\my artist\\\\my album 11\\\\3 my track 113.mp3\"" +
					" has been backed up to" +
					" \"Music\\\\my artist\\\\my album 11\\\\pre-repair-backup\\\\3.mp3\".\n" +
					"The track file" +
					" \"Music\\\\my artist\\\\my album 11\\\\4 my track 114.mp3\"" +
					" has been backed up to" +
					" \"Music\\\\my artist\\\\my album 11\\\\pre-repair-backup\\\\4.mp3\".\n" +
					"The track file" +
					" \"Music\\\\my artist\\\\my album 12\\\\1 my track 121.mp3\"" +
					" has been backed up to" +
					" \"Music\\\\my artist\\\\my album 12\\\\pre-repair-backup\\\\1.mp3\".\n" +
					"The track file" +
					" \"Music\\\\my artist\\\\my album 12\\\\2 my track 122.mp3\"" +
					" has been backed up to" +
					" \"Music\\\\my artist\\\\my album 12\\\\pre-repair-backup\\\\2.mp3\".\n" +
					"The track file" +
					" \"Music\\\\my artist\\\\my album 12\\\\3 my track 123.mp3\"" +
					" has been backed up to" +
					" \"Music\\\\my artist\\\\my album 12\\\\pre-repair-backup\\\\3.mp3\".\n" +
					"The track file" +
					" \"Music\\\\my artist\\\\my album 12\\\\4 my track 124.mp3\"" +
					" has been backed up to" +
					" \"Music\\\\my artist\\\\my album 12\\\\pre-repair-backup\\\\4.mp3\".\n",
				Error: "" +
					"An error occurred repairing track" +
					" \"Music\\\\my artist\\\\my album 00\\\\1 my track 001.mp3\".\n" +
					"An error occurred repairing track" +
					" \"Music\\\\my artist\\\\my album 00\\\\2 my track 002.mp3\".\n" +
					"An error occurred repairing track" +
					" \"Music\\\\my artist\\\\my album 00\\\\3 my track 003.mp3\".\n" +
					"An error occurred repairing track" +
					" \"Music\\\\my artist\\\\my album 00\\\\4 my track 004.mp3\".\n" +
					"An error occurred repairing track" +
					" \"Music\\\\my artist\\\\my album 01\\\\1 my track 011.mp3\".\n" +
					"An error occurred repairing track" +
					" \"Music\\\\my artist\\\\my album 01\\\\2 my track 012.mp3\".\n" +
					"An error occurred repairing track" +
					" \"Music\\\\my artist\\\\my album 01\\\\3 my track 013.mp3\".\n" +
					"An error occurred repairing track" +
					" \"Music\\\\my artist\\\\my album 01\\\\4 my track 014.mp3\".\n" +
					"An error occurred repairing track" +
					" \"Music\\\\my artist\\\\my album 02\\\\1 my track 021.mp3\".\n" +
					"An error occurred repairing track" +
					" \"Music\\\\my artist\\\\my album 02\\\\2 my track 022.mp3\".\n" +
					"An error occurred repairing track" +
					" \"Music\\\\my artist\\\\my album 02\\\\3 my track 023.mp3\".\n" +
					"An error occurred repairing track" +
					" \"Music\\\\my artist\\\\my album 02\\\\4 my track 024.mp3\".\n" +
					"An error occurred repairing track" +
					" \"Music\\\\my artist\\\\my album 10\\\\1 my track 101.mp3\".\n" +
					"An error occurred repairing track" +
					" \"Music\\\\my artist\\\\my album 10\\\\2 my track 102.mp3\".\n" +
					"An error occurred repairing track" +
					" \"Music\\\\my artist\\\\my album 10\\\\3 my track 103.mp3\".\n" +
					"An error occurred repairing track" +
					" \"Music\\\\my artist\\\\my album 10\\\\4 my track 104.mp3\".\n" +
					"An error occurred repairing track" +
					" \"Music\\\\my artist\\\\my album 11\\\\1 my track 111.mp3\".\n" +
					"An error occurred repairing track" +
					" \"Music\\\\my artist\\\\my album 11\\\\2 my track 112.mp3\".\n" +
					"An error occurred repairing track" +
					" \"Music\\\\my artist\\\\my album 11\\\\3 my track 113.mp3\".\n" +
					"An error occurred repairing track" +
					" \"Music\\\\my artist\\\\my album 11\\\\4 my track 114.mp3\".\n" +
					"An error occurred repairing track" +
					" \"Music\\\\my artist\\\\my album 12\\\\1 my track 121.mp3\".\n" +
					"An error occurred repairing track" +
					" \"Music\\\\my artist\\\\my album 12\\\\2 my track 122.mp3\".\n" +
					"An error occurred repairing track" +
					" \"Music\\\\my artist\\\\my album 12\\\\3 my track 123.mp3\".\n" +
					"An error occurred repairing track" +
					" \"Music\\\\my artist\\\\my album 12\\\\4 my track 124.mp3\".\n",
				Log: "level='error'" +
					" command='repair'" +
					" directory='Music\\my artist\\my album 00'" +
					" error='[\"open Music\\\\my artist\\\\my album 00\\\\1 my track" +
					" 001.mp3: The system cannot find the path specified.\"," +
					" \"open Music\\\\my artist\\\\my album 00\\\\1 my track 001.mp3: The" +
					" system cannot find the path specified.\"]'" +
					" fileName='1 my track 001.mp3'" +
					" msg='cannot edit track'\n" +
					"level='error'" +
					" command='repair'" +
					" directory='Music\\my artist\\my album 00'" +
					" error='[\"open Music\\\\my artist\\\\my album 00\\\\2 my track" +
					" 002.mp3: The system cannot find the path specified.\"," +
					" \"open Music\\\\my artist\\\\my album 00\\\\2 my track 002.mp3: The" +
					" system cannot find the path specified.\"]'" +
					" fileName='2 my track 002.mp3'" +
					" msg='cannot edit track'\n" +
					"level='error'" +
					" command='repair'" +
					" directory='Music\\my artist\\my album 00'" +
					" error='[\"open Music\\\\my artist\\\\my album 00\\\\3 my track" +
					" 003.mp3: The system cannot find the path specified.\"," +
					" \"open Music\\\\my artist\\\\my album 00\\\\3 my track 003.mp3: The" +
					" system cannot find the path specified.\"]'" +
					" fileName='3 my track 003.mp3'" +
					" msg='cannot edit track'\n" +
					"level='error'" +
					" command='repair'" +
					" directory='Music\\my artist\\my album 00'" +
					" error='[\"open Music\\\\my artist\\\\my album 00\\\\4 my track" +
					" 004.mp3: The system cannot find the path specified.\"," +
					" \"open Music\\\\my artist\\\\my album 00\\\\4 my track 004.mp3: The" +
					" system cannot find the path specified.\"]'" +
					" fileName='4 my track 004.mp3'" +
					" msg='cannot edit track'\n" +
					"level='error'" +
					" command='repair'" +
					" directory='Music\\my artist\\my album 01'" +
					" error='[\"open Music\\\\my artist\\\\my album 01\\\\1 my track" +
					" 011.mp3: The system cannot find the path specified.\"," +
					" \"open Music\\\\my artist\\\\my album 01\\\\1 my track 011.mp3: The" +
					" system cannot find the path specified.\"]'" +
					" fileName='1 my track 011.mp3'" +
					" msg='cannot edit track'\n" +
					"level='error'" +
					" command='repair'" +
					" directory='Music\\my artist\\my album 01'" +
					" error='[\"open Music\\\\my artist\\\\my album 01\\\\2 my track" +
					" 012.mp3: The system cannot find the path specified.\"," +
					" \"open Music\\\\my artist\\\\my album 01\\\\2 my track 012.mp3: The" +
					" system cannot find the path specified.\"]'" +
					" fileName='2 my track 012.mp3'" +
					" msg='cannot edit track'\n" +
					"level='error'" +
					" command='repair'" +
					" directory='Music\\my artist\\my album 01'" +
					" error='[\"open Music\\\\my artist\\\\my album 01\\\\3 my track" +
					" 013.mp3: The system cannot find the path specified.\"," +
					" \"open Music\\\\my artist\\\\my album 01\\\\3 my track 013.mp3: The" +
					" system cannot find the path specified.\"]'" +
					" fileName='3 my track 013.mp3'" +
					" msg='cannot edit track'\n" +
					"level='error'" +
					" command='repair'" +
					" directory='Music\\my artist\\my album 01'" +
					" error='[\"open Music\\\\my artist\\\\my album 01\\\\4 my track" +
					" 014.mp3: The system cannot find the path specified.\"," +
					" \"open Music\\\\my artist\\\\my album 01\\\\4 my track 014.mp3: The" +
					" system cannot find the path specified.\"]'" +
					" fileName='4 my track 014.mp3'" +
					" msg='cannot edit track'\n" +
					"level='error'" +
					" command='repair'" +
					" directory='Music\\my artist\\my album 02'" +
					" error='[\"open Music\\\\my artist\\\\my album 02\\\\1 my track" +
					" 021.mp3: The system cannot find the path specified.\"," +
					" \"open Music\\\\my artist\\\\my album 02\\\\1 my track 021.mp3: The" +
					" system cannot find the path specified.\"]'" +
					" fileName='1 my track 021.mp3'" +
					" msg='cannot edit track'\n" +
					"level='error'" +
					" command='repair'" +
					" directory='Music\\my artist\\my album 02'" +
					" error='[\"open Music\\\\my artist\\\\my album 02\\\\2 my track" +
					" 022.mp3: The system cannot find the path specified.\"," +
					" \"open Music\\\\my artist\\\\my album 02\\\\2 my track 022.mp3: The" +
					" system cannot find the path specified.\"]'" +
					" fileName='2 my track 022.mp3'" +
					" msg='cannot edit track'\n" +
					"level='error'" +
					" command='repair'" +
					" directory='Music\\my artist\\my album 02'" +
					" error='[\"open Music\\\\my artist\\\\my album 02\\\\3 my track" +
					" 023.mp3: The system cannot find the path specified.\"," +
					" \"open Music\\\\my artist\\\\my album 02\\\\3 my track 023.mp3: The" +
					" system cannot find the path specified.\"]'" +
					" fileName='3 my track 023.mp3'" +
					" msg='cannot edit track'\n" +
					"level='error'" +
					" command='repair'" +
					" directory='Music\\my artist\\my album 02'" +
					" error='[\"open Music\\\\my artist\\\\my album 02\\\\4 my track" +
					" 024.mp3: The system cannot find the path specified.\"," +
					" \"open Music\\\\my artist\\\\my album 02\\\\4 my track 024.mp3: The" +
					" system cannot find the path specified.\"]'" +
					" fileName='4 my track 024.mp3'" +
					" msg='cannot edit track'\n" +
					"level='error'" +
					" command='repair'" +
					" directory='Music\\my artist\\my album 10'" +
					" error='[\"open Music\\\\my artist\\\\my album 10\\\\1 my track" +
					" 101.mp3: The system cannot find the path specified.\"," +
					" \"open Music\\\\my artist\\\\my album 10\\\\1 my track 101.mp3: The" +
					" system cannot find the path specified.\"]'" +
					" fileName='1 my track 101.mp3'" +
					" msg='cannot edit track'\n" +
					"level='error'" +
					" command='repair'" +
					" directory='Music\\my artist\\my album 10'" +
					" error='[\"open Music\\\\my artist\\\\my album 10\\\\2 my track" +
					" 102.mp3: The system cannot find the path specified.\"," +
					" \"open Music\\\\my artist\\\\my album 10\\\\2 my track 102.mp3: The" +
					" system cannot find the path specified.\"]'" +
					" fileName='2 my track 102.mp3'" +
					" msg='cannot edit track'\n" +
					"level='error'" +
					" command='repair'" +
					" directory='Music\\my artist\\my album 10'" +
					" error='[\"open Music\\\\my artist\\\\my album 10\\\\3 my track" +
					" 103.mp3: The system cannot find the path specified.\"," +
					" \"open Music\\\\my artist\\\\my album 10\\\\3 my track 103.mp3: The" +
					" system cannot find the path specified.\"]'" +
					" fileName='3 my track 103.mp3'" +
					" msg='cannot edit track'\n" +
					"level='error'" +
					" command='repair'" +
					" directory='Music\\my artist\\my album 10'" +
					" error='[\"open Music\\\\my artist\\\\my album 10\\\\4 my track" +
					" 104.mp3: The system cannot find the path specified.\"," +
					" \"open Music\\\\my artist\\\\my album 10\\\\4 my track 104.mp3: The" +
					" system cannot find the path specified.\"]'" +
					" fileName='4 my track 104.mp3'" +
					" msg='cannot edit track'\n" +
					"level='error'" +
					" command='repair'" +
					" directory='Music\\my artist\\my album 11'" +
					" error='[\"open Music\\\\my artist\\\\my album 11\\\\1 my track" +
					" 111.mp3: The system cannot find the path specified.\"," +
					" \"open Music\\\\my artist\\\\my album 11\\\\1 my track 111.mp3: The" +
					" system cannot find the path specified.\"]'" +
					" fileName='1 my track 111.mp3'" +
					" msg='cannot edit track'\n" +
					"level='error'" +
					" command='repair'" +
					" directory='Music\\my artist\\my album 11'" +
					" error='[\"open Music\\\\my artist\\\\my album 11\\\\2 my track" +
					" 112.mp3: The system cannot find the path specified.\"," +
					" \"open Music\\\\my artist\\\\my album 11\\\\2 my track 112.mp3: The" +
					" system cannot find the path specified.\"]'" +
					" fileName='2 my track 112.mp3'" +
					" msg='cannot edit track'\n" +
					"level='error'" +
					" command='repair'" +
					" directory='Music\\my artist\\my album 11'" +
					" error='[\"open Music\\\\my artist\\\\my album 11\\\\3 my track" +
					" 113.mp3: The system cannot find the path specified.\"," +
					" \"open Music\\\\my artist\\\\my album 11\\\\3 my track 113.mp3: The" +
					" system cannot find the path specified.\"]'" +
					" fileName='3 my track 113.mp3'" +
					" msg='cannot edit track'\n" +
					"level='error'" +
					" command='repair'" +
					" directory='Music\\my artist\\my album 11'" +
					" error='[\"open Music\\\\my artist\\\\my album 11\\\\4 my track" +
					" 114.mp3: The system cannot find the path specified.\"," +
					" \"open Music\\\\my artist\\\\my album 11\\\\4 my track 114.mp3: The" +
					" system cannot find the path specified.\"]'" +
					" fileName='4 my track 114.mp3'" +
					" msg='cannot edit track'\n" +
					"level='error'" +
					" command='repair'" +
					" directory='Music\\my artist\\my album 12'" +
					" error='[\"open Music\\\\my artist\\\\my album 12\\\\1 my track" +
					" 121.mp3: The system cannot find the path specified.\"," +
					" \"open Music\\\\my artist\\\\my album 12\\\\1 my track 121.mp3: The" +
					" system cannot find the path specified.\"]'" +
					" fileName='1 my track 121.mp3'" +
					" msg='cannot edit track'\n" +
					"level='error'" +
					" command='repair'" +
					" directory='Music\\my artist\\my album 12'" +
					" error='[\"open Music\\\\my artist\\\\my album 12\\\\2 my track" +
					" 122.mp3: The system cannot find the path specified.\"," +
					" \"open Music\\\\my artist\\\\my album 12\\\\2 my track 122.mp3: The" +
					" system cannot find the path specified.\"]'" +
					" fileName='2 my track 122.mp3'" +
					" msg='cannot edit track'\n" +
					"level='error'" +
					" command='repair'" +
					" directory='Music\\my artist\\my album 12'" +
					" error='[\"open Music\\\\my artist\\\\my album 12\\\\3 my track" +
					" 123.mp3: The system cannot find the path specified.\"," +
					" \"open Music\\\\my artist\\\\my album 12\\\\3 my track 123.mp3: The" +
					" system cannot find the path specified.\"]'" +
					" fileName='3 my track 123.mp3'" +
					" msg='cannot edit track'\n" +
					"level='error'" +
					" command='repair'" +
					" directory='Music\\my artist\\my album 12'" +
					" error='[\"open Music\\\\my artist\\\\my album 12\\\\4 my track" +
					" 124.mp3: The system cannot find the path specified.\"," +
					" \"open Music\\\\my artist\\\\my album 12\\\\4 my track 124.mp3: The" +
					" system cannot find the path specified.\"]'" +
					" fileName='4 my track 124.mp3'" +
					" msg='cannot edit track'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			got := tt.rs.repairArtists(o, tt.artists)
			if !compareExitErrors(got, tt.wantStatus) {
				t.Errorf("repairSettings.repairArtists() got %s want %s", got, tt.wantStatus)
			}
			o.Report(t, "repairSettings.repairArtists()", tt.WantedRecording)
		})
	}
}

func Test_repairSettings_processArtists(t *testing.T) {
	originalReadMetadata := readMetadata
	defer func() {
		readMetadata = originalReadMetadata
	}()
	readMetadata = func(_ output.Bus, _ []*files.Artist) {}
	type args struct {
		allArtists []*files.Artist
		ss         *searchSettings
	}
	tests := map[string]struct {
		rs *repairSettings
		args
		wantStatus *cmdtoolkit.ExitError
		output.WantedRecording
	}{
		"nothing to do": {
			rs:         &repairSettings{dryRun: cmdtoolkit.CommandFlag[bool]{Value: true}},
			args:       args{},
			wantStatus: cmdtoolkit.NewExitUserError("repair"),
		},
		"clean artists": {
			rs: &repairSettings{dryRun: cmdtoolkit.CommandFlag[bool]{Value: true}},
			args: args{
				allArtists: generateArtists(2, 3, 4, nil),
				ss: &searchSettings{
					artistFilter: regexp.MustCompile(".*"),
					albumFilter:  regexp.MustCompile(".*"),
					trackFilter:  regexp.MustCompile(".*"),
				},
			},
			wantStatus: nil,
			WantedRecording: output.WantedRecording{
				Console: "No repairable track defects were found.\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			got := tt.rs.processArtists(o, tt.args.allArtists, tt.args.ss)
			if !compareExitErrors(got, tt.wantStatus) {
				t.Errorf("repairSettings.processArtists() got %s want %s", got, tt.wantStatus)
			}
			o.Report(t, "repairSettings.processArtists()", tt.WantedRecording)
		})
	}
}

func Test_repairRun(t *testing.T) {
	initGlobals()
	originalBus := bus
	originalSearchFlags := searchFlags
	defer func() {
		bus = originalBus
		searchFlags = originalSearchFlags
	}()
	searchFlags = safeSearchFlags
	repairFlags := &cmdtoolkit.FlagSet{
		Name: "repair",
		Details: map[string]*cmdtoolkit.FlagDetails{
			"dryRun": {
				Usage:        "output what would have been repaired, but make no" + " repairs",
				ExpectedType: cmdtoolkit.BoolType,
				DefaultValue: false,
			},
		},
	}
	command := &cobra.Command{}
	cmdtoolkit.AddFlags(output.NewNilBus(), cmdtoolkit.EmptyConfiguration(), command.Flags(),
		repairFlags, searchFlags)
	tests := map[string]struct {
		cmd *cobra.Command
		in1 []string
		output.WantedRecording
	}{
		"basic": {
			cmd: command,
			WantedRecording: output.WantedRecording{
				Error: "" +
					"No mp3 files could be found using the specified parameters.\n" +
					"Why?\n" +
					"There were no directories found in \".\" (the --topDir value).\n" +
					"What to do:\n" +
					"Set --topDir to the path of a directory that contains artist" +
					" directories.\n",
				Log: "" +
					"level='error'" +
					" --topDir='.'" +
					" msg='cannot find any artist directories'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			bus = o // cook getBus()
			_ = repairRun(tt.cmd, tt.in1)
			o.Report(t, "repairRun()", tt.WantedRecording)
		})
	}
}

func Test_repair_Help(t *testing.T) {
	originalSearchFlags := searchFlags
	defer func() {
		searchFlags = originalSearchFlags
	}()
	searchFlags = safeSearchFlags
	commandUnderTest := cloneCommand(repairCmd)
	cmdtoolkit.AddFlags(output.NewNilBus(), cmdtoolkit.EmptyConfiguration(),
		commandUnderTest.Flags(), repairFlags, searchFlags)
	tests := map[string]struct {
		output.WantedRecording
	}{
		"good": {
			WantedRecording: output.WantedRecording{
				Console: "" +
					"\"repair\" repairs the problems found by running 'check --files'\n" +
					"\n" +
					"This command rewrites the mp3 files that the check command noted as" +
					" having metadata\n" +
					"inconsistent with the file structure. Prior to rewriting an mp3 file," +
					" the repair\n" +
					"command creates a backup directory for the parent album and copies the" +
					" original mp3\n" +
					"file into that backup directory. Use the postRepair command to" +
					" automatically delete\n" +
					"the backup folders.\n" +
					"\n" +
					"Usage:\n" +
					"  repair [--dryRun] [--albumFilter regex] [--artistFilter regex]" +
					" [--trackFilter regex] [--topDir dir] [--extensions extensions]\n" +
					"\n" +
					"Examples:\n" +
					"repair --dryRun\n  Output what would be repaired, but does not perform the stated repairs\n" +
					"\n" +
					"Flags:\n" +
					"      --albumFilter string    " +
					"regular expression specifying which albums to select (default \".*\")\n" +
					"      --artistFilter string   " +
					"regular expression specifying which artists to select (default \".*\")\n" +
					"      --dryRun                " +
					"output what would have been repaired, but make no repairs (default false)\n" +
					"      --extensions string     " +
					"comma-delimited list of file extensions used by mp3 files (default \".mp3\")\n" +
					"      --topDir string         " +
					"top directory specifying where to find mp3 files (default \".\")\n" +
					"      --trackFilter string    " +
					"regular expression specifying which tracks to select (default \".*\")\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			command := commandUnderTest
			enableCommandRecording(o, command)
			_ = command.Help()
			o.Report(t, "repair Help()", tt.WantedRecording)
		})
	}
}
