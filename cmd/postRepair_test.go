/*
Copyright © 2021 Marc Johnson (marc.johnson27591@gmail.com)
*/
package cmd_test

import (
	"fmt"
	"mp3repair/cmd"
	"mp3repair/internal/files"
	"regexp"
	"testing"

	cmd_toolkit "github.com/majohn-r/cmd-toolkit"
	"github.com/majohn-r/output"
	"github.com/spf13/cobra"
)

func TestRemoveBackupDirectory(t *testing.T) {
	originalRemoveAll := cmd.RemoveAll
	defer func() {
		cmd.RemoveAll = originalRemoveAll
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
				Log: "level='error'" +
					" directory='my directory'" +
					" error='dir locked'" +
					" msg='cannot delete directory'\n",
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
			if differences, ok := o.Verify(tt.WantedRecording); !ok {
				for _, difference := range differences {
					t.Errorf("RemoveBackupDirectory() %s", difference)
				}
			}
		})
	}
}

func TestPostRepairWork(t *testing.T) {
	originalRemoveAll := cmd.RemoveAll
	originalDirExists := cmd.DirExists
	defer func() {
		cmd.RemoveAll = originalRemoveAll
		cmd.DirExists = originalDirExists
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
		"no artists": {
			args: args{
				ss:         cmd.NewSearchSettings(),
				allArtists: []*files.Artist{},
				loaded:     true,
			},
			WantedRecording: output.WantedRecording{
				Error: "" +
					"No music files remain after filtering.\n" +
					"Why?\n" +
					"After applying --artistFilter=<nil>, --albumFilter=<nil>, and" +
					" --trackFilter=<nil>, no files remained.\n" +
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
				ss: cmd.NewSearchSettings().WithArtistFilter(
					regexp.MustCompile(".*")).WithAlbumFilter(
					regexp.MustCompile(".*")).WithTrackFilter(regexp.MustCompile(".*")),
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
				ss: cmd.NewSearchSettings().WithArtistFilter(
					regexp.MustCompile(".*")).WithAlbumFilter(
					regexp.MustCompile(".*")).WithTrackFilter(regexp.MustCompile(".*")),
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
		"backups to remove, not all successfully": {
			dirExists: func(dir string) bool { return true },
			removeAll: func(dir string) error { return fmt.Errorf("nope") },
			args: args{
				ss: cmd.NewSearchSettings().WithArtistFilter(
					regexp.MustCompile(".*")).WithAlbumFilter(
					regexp.MustCompile(".*")).WithTrackFilter(regexp.MustCompile(".*")),
				allArtists: generateArtists(2, 3, 4),
				loaded:     true,
			},
			WantedRecording: output.WantedRecording{
				Console: "" +
					"Backup directories to delete: 6.\n" +
					"Backup directories deleted: 0.\n",
				Log: "" +
					"level='error'" +
					" directory='Music\\my artist\\my album 00\\pre-repair-backup'" +
					" error='nope'" +
					" msg='cannot delete directory'\n" +
					"level='error'" +
					" directory='Music\\my artist\\my album 01\\pre-repair-backup'" +
					" error='nope'" +
					" msg='cannot delete directory'\n" +
					"level='error'" +
					" directory='Music\\my artist\\my album 02\\pre-repair-backup'" +
					" error='nope'" +
					" msg='cannot delete directory'\n" +
					"level='error'" +
					" directory='Music\\my artist\\my album 10\\pre-repair-backup'" +
					" error='nope'" +
					" msg='cannot delete directory'\n" +
					"level='error'" +
					" directory='Music\\my artist\\my album 11\\pre-repair-backup'" +
					" error='nope'" +
					" msg='cannot delete directory'\n" +
					"level='error'" +
					" directory='Music\\my artist\\my album 12\\pre-repair-backup'" +
					" error='nope'" +
					" msg='cannot delete directory'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			cmd.RemoveAll = tt.removeAll
			cmd.DirExists = tt.dirExists
			o := output.NewRecorder()
			cmd.PostRepairWork(o, tt.args.ss, tt.args.allArtists, tt.args.loaded)
			if differences, ok := o.Verify(tt.WantedRecording); !ok {
				for _, difference := range differences {
					t.Errorf("PostRepairWork() %s", difference)
				}
			}
		})
	}
}

func TestPostRepairRun(t *testing.T) {
	cmd.InitGlobals()
	originalBus := cmd.Bus
	defer func() {
		cmd.Bus = originalBus
	}()
	command := &cobra.Command{}
	cmd.AddFlags(output.NewNilBus(), cmd_toolkit.EmptyConfiguration(), command.Flags(),
		safeSearchFlags)
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
					"Set --topDir to the path of a directory that contains artist" +
					" directories.\n",
				Log: "" +
					"level='info'" +
					" --albumFilter='.*'" +
					" --artistFilter='.*'" +
					" --extensions='[.mp3]'" +
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
			if differences, ok := o.Verify(tt.WantedRecording); !ok {
				for _, difference := range differences {
					t.Errorf("PostRepairRun() %s", difference)
				}
			}
		})
	}
}

func TestPostRepairHelp(t *testing.T) {
	commandUnderTest := cloneCommand(cmd.PostRepairCmd)
	cmd.AddFlags(output.NewNilBus(), cmd_toolkit.EmptyConfiguration(),
		commandUnderTest.Flags(), safeSearchFlags)
	tests := map[string]struct {
		output.WantedRecording
	}{
		"good": {
			WantedRecording: output.WantedRecording{
				Console: "" +
					"\"postRepair\" deletes the backup directories (and their contents)" +
					" created by the \"repair\" command\n" +
					"\n" +
					"Usage:\n" +
					"  postRepair [--albumFilter regex] [--artistFilter regex]" +
					" [--trackFilter regex] [--topDir dir] [--extensions extensions]\n" +
					"\n" +
					"Flags:\n" +
					"      --albumFilter string    regular expression specifying which" +
					" albums to select (default \".*\")\n" +
					"      --artistFilter string   regular expression specifying which" +
					" artists to select (default \".*\")\n" +
					"      --extensions string     comma-delimited list of file extensions" +
					" used by mp3 files (default \".mp3\")\n" +
					"      --topDir string         top directory specifying where to find" +
					" mp3 files (default \".\")\n" +
					"      --trackFilter string    regular expression specifying which" +
					" tracks to select (default \".*\")\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			command := commandUnderTest
			enableCommandRecording(o, command)
			command.Help()
			if differences, ok := o.Verify(tt.WantedRecording); !ok {
				for _, difference := range differences {
					t.Errorf("postRepair Help() %s", difference)
				}
			}
		})
	}
}

func TestPostRepairUsage(t *testing.T) {
	commandUnderTest := cloneCommand(cmd.PostRepairCmd)
	cmd.AddFlags(output.NewNilBus(), cmd_toolkit.EmptyConfiguration(),
		commandUnderTest.Flags(), safeSearchFlags)
	tests := map[string]struct {
		output.WantedRecording
	}{
		"good": {
			WantedRecording: output.WantedRecording{
				Console: "" +
					"Usage:\n" +
					"  postRepair [--albumFilter regex] [--artistFilter regex]" +
					" [--trackFilter regex] [--topDir dir] [--extensions extensions]\n" +
					"\n" +
					"Flags:\n" +
					"      --albumFilter string    regular expression specifying which" +
					" albums to select (default \".*\")\n" +
					"      --artistFilter string   regular expression specifying which" +
					" artists to select (default \".*\")\n" +
					"      --extensions string     comma-delimited list of file extensions" +
					" used by mp3 files (default \".mp3\")\n" +
					"      --topDir string         top directory specifying where to find" +
					" mp3 files (default \".\")\n" +
					"      --trackFilter string    regular expression specifying which" +
					" tracks to select (default \".*\")\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			command := commandUnderTest
			enableCommandRecording(o, command)
			command.Usage()
			if differences, ok := o.Verify(tt.WantedRecording); !ok {
				for _, difference := range differences {
					t.Errorf("postRepair Usage() %s", difference)
				}
			}
		})
	}
}
