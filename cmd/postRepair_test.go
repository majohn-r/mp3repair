/*
Copyright © 2021 Marc Johnson (marc.johnson27591@gmail.com)
*/
package cmd_test

import (
	"fmt"
	"mp3/cmd"
	"mp3/internal/files"
	"regexp"
	"testing"

	cmd_toolkit "github.com/majohn-r/cmd-toolkit"
	"github.com/majohn-r/output"
	"github.com/spf13/cobra"
)

func TestRemoveBackupDirectory(t *testing.T) {
	oldRemoveAll := cmd.RemoveAll
	defer func() {
		cmd.RemoveAll = oldRemoveAll
	}()
	tests := map[string]struct {
		removeAll func(dir string) error
		dir       string
		want      bool
		output.WantedRecording
	}{
		"failure": {
			removeAll: func(dir string) error { return fmt.Errorf("dir locked") },
			dir:       "my directory",
			want:      false,
			WantedRecording: output.WantedRecording{
				Log: "level='error' directory='my directory' error='dir locked' msg='cannot delete directory'\n",
			},
		},
		"success": {
			removeAll: func(dir string) error { return nil },
			dir:       "my other directory",
			want:      true,
			WantedRecording: output.WantedRecording{
				Log: "" +
					"level='info'" +
					" directory='my other directory'" +
					" msg='directory deleted'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			cmd.RemoveAll = tt.removeAll
			o := output.NewRecorder()
			if got := cmd.RemoveBackupDirectory(o, tt.dir); got != tt.want {
				t.Errorf("RemoveBackupDirectory() = %v, want %v", got, tt.want)
			}
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("RemoveBackupDirectory() %s", issue)
				}
			}
		})
	}
}

func TestPostRepairWork(t *testing.T) {
	oldRemoveAll := cmd.RemoveAll
	oldDirExists := cmd.DirExists
	defer func() {
		cmd.RemoveAll = oldRemoveAll
		cmd.DirExists = oldDirExists
	}()
	type args struct {
		ss         *cmd.SearchSettings
		allArtists []*files.Artist
		loaded     bool
	}
	tests := map[string]struct {
		removeAll func(dir string) error
		dirExists func(dir string) bool
		args
		output.WantedRecording
	}{
		"no load": {args: args{}},
		"no artists": {args: args{
			ss:         &cmd.SearchSettings{},
			allArtists: []*files.Artist{},
			loaded:     true,
		},
			WantedRecording: output.WantedRecording{
				Error: "" +
					"No music files remain after filtering.\n" +
					"Why?\n" +
					"After applying --artistFilter=<nil>, --albumFilter=<nil>, and --trackFilter=<nil>, no files remained.\n" +
					"What to do:\n" +
					"Use less restrictive filter settings.\n",
				Log: "" +
					"level='error'" +
					" --albumFilter='<nil>'" +
					" --artistFilter='<nil>'" +
					" --trackFilter='<nil>'" +
					" msg='no files remain after filtering'\n",
			},
		},
		"artists with no work to do": {
			dirExists: func(dir string) bool { return false },
			args: args{
				ss: &cmd.SearchSettings{
					ArtistFilter: regexp.MustCompile(".*"),
					AlbumFilter:  regexp.MustCompile(".*"),
					TrackFilter:  regexp.MustCompile(".*"),
				},
				allArtists: generateArtists(2, 3, 4),
				loaded:     true,
			},
			WantedRecording: output.WantedRecording{
				Console: "Backup directories to delete: 0.\n",
			},
		},
		"backups to remove": {
			dirExists: func(dir string) bool { return true },
			removeAll: func(dir string) error { return nil },
			args: args{
				ss: &cmd.SearchSettings{
					ArtistFilter: regexp.MustCompile(".*"),
					AlbumFilter:  regexp.MustCompile(".*"),
					TrackFilter:  regexp.MustCompile(".*"),
				},
				allArtists: generateArtists(2, 3, 4),
				loaded:     true,
			},
			WantedRecording: output.WantedRecording{
				Console: "" +
					"Backup directories to delete: 6.\n" +
					"Backup directories deleted: 6.\n",
				Log: "" +
					"level='info'" +
					" directory='Music\\my artist\\my album 00\\pre-repair-backup'" +
					" msg='directory deleted'\n" +
					"level='info'" +
					" directory='Music\\my artist\\my album 01\\pre-repair-backup'" +
					" msg='directory deleted'\n" +
					"level='info'" +
					" directory='Music\\my artist\\my album 02\\pre-repair-backup'" +
					" msg='directory deleted'\n" +
					"level='info'" +
					" directory='Music\\my artist\\my album 10\\pre-repair-backup'" +
					" msg='directory deleted'\n" +
					"level='info'" +
					" directory='Music\\my artist\\my album 11\\pre-repair-backup'" +
					" msg='directory deleted'\n" +
					"level='info'" +
					" directory='Music\\my artist\\my album 12\\pre-repair-backup'" +
					" msg='directory deleted'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			cmd.RemoveAll = tt.removeAll
			cmd.DirExists = tt.dirExists
			o := output.NewRecorder()
			cmd.PostRepairWork(o, tt.args.ss, tt.args.allArtists, tt.args.loaded)
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("PostRepairWork() %s", issue)
				}
			}
		})
	}
}

func TestPostRepairRun(t *testing.T) {
	cmd.InitGlobals()
	oldBus := cmd.Bus
	defer func() {
		cmd.Bus = oldBus
	}()
	command := &cobra.Command{}
	cmd.AddFlags(output.NewNilBus(), cmd_toolkit.EmptyConfiguration(), command.Flags(), safeSearchFlags, false)
	type args struct {
		cmd *cobra.Command
		in1 []string
	}
	tests := map[string]struct {
		args
		output.WantedRecording
	}{
		"typical": {
			args: args{cmd: command},
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
					" --topDir='.'" +
					" --trackFilter='.*'" +
					" command='postRepair'" +
					" msg='executing command'\n" +
					"level='error'" +
					" --topDir='.'" +
					" msg='cannot find any artist directories'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			cmd.Bus = o // cook getBus()
			cmd.PostRepairRun(tt.args.cmd, tt.args.in1)
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("PostRepairRun() %s", issue)
				}
			}
		})
	}
}
