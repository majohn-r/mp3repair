package subcommands

import (
	"flag"
	"fmt"
	"io"
	"mp3/internal"
	"mp3/internal/files"
	"os"
	"path/filepath"
	"sort"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type repair struct {
	n      string
	dryRun *bool
	sf     *files.SearchFlags
}

func (r *repair) name() string {
	return r.n
}

func newRepair(v *viper.Viper, fSet *flag.FlagSet) CommandProcessor {
	return newRepairSubCommand(v, fSet)
}

const (
	dryRunFlag    = "dryRun"
	defaultDryRun = false
)

func newRepairSubCommand(v *viper.Viper, fSet *flag.FlagSet) *repair {
	subViper := internal.SafeSubViper(v, "repair")
	return &repair{
		n: fSet.Name(),
		dryRun: fSet.Bool(dryRunFlag,
			internal.GetBoolDefault(subViper, dryRunFlag, defaultDryRun),
			"if true, output what would have repaired, but make no repairs"),
		sf: files.NewSearchFlags(v, fSet),
	}
}

func (r *repair) Exec(w io.Writer, args []string) {
	if s := r.sf.ProcessArgs(os.Stderr, args); s != nil {
		r.runSubcommand(w, s)
	}
}

func (r *repair) logFields() logrus.Fields {
	return logrus.Fields{internal.LOG_COMMAND_NAME: r.name(), dryRunFlag: *r.dryRun}
}

func (r *repair) runSubcommand(w io.Writer, s *files.Search) {
	logrus.WithFields(r.logFields()).Info(internal.LOG_EXECUTING_COMMAND)
	artists := s.LoadData()
	files.UpdateTracks(artists, files.RawReadTags)
	tracksWithConflicts := findConflictedTracks(artists)
	if len(tracksWithConflicts) == 0 {
		fmt.Fprintln(w, noProblemsFound)
	} else {
		if *r.dryRun {
			reportTracks(w, tracksWithConflicts)
		} else {
			r.createBackups(w, tracksWithConflicts)
			r.fixTracks(w, tracksWithConflicts)
		}
	}
}

func findConflictedTracks(artists []*files.Artist) []*files.Track {
	var t []*files.Track
	for _, artist := range artists {
		for _, album := range artist.Albums() {
			for _, track := range album.Tracks() {
				if state := track.AnalyzeIssues(); state.HasTaggingConflicts() {
					t = append(t, track)
				}
			}
		}
	}
	sort.Sort(files.Tracks(t))
	return t
}

const noProblemsFound = "No repairable track defects found"

func reportTracks(w io.Writer, tracks []*files.Track) {
	lastArtistName := ""
	lastAlbumName := ""
	for _, t := range tracks {
		albumName := t.ContainingAlbum.Name()
		artistName := t.ContainingAlbum.RecordingArtistName()
		if lastArtistName != artistName {
			fmt.Fprintf(w, "%q\n", artistName)
			lastArtistName = artistName
			lastAlbumName = ""
		}
		if albumName != lastAlbumName {
			fmt.Fprintf(w, "    %q\n", albumName)
			lastAlbumName = albumName
		}
		s := t.AnalyzeIssues()
		fmt.Fprintf(w, "        %2d %q need to fix%s%s%s%s\n",
			t.TrackNumber, t.Name(),
			reportProblem(s.HasNumberingConflict(), " track numbering;"),
			reportProblem(s.HasTrackNameConflict(), " track name;"),
			reportProblem(s.HasAlbumNameConflict(), " album name;"),
			reportProblem(s.HasArtistNameConflict(), " artist name;"))
	}
}

func reportProblem(b bool, problem string) (s string) {
	if b {
		s = problem
	}
	return
}

func (r *repair) fixTracks(w io.Writer, tracks []*files.Track) {
	for _, t := range tracks {
		if err := t.EditTags(); err != nil {
			fmt.Fprintf(w, "An error occurred fixing track %q\n", t)
			logrus.WithFields(logrus.Fields{
				internal.LOG_EXECUTING_COMMAND: r.name(),
				internal.LOG_PATH:              t.String(),
				internal.LOG_ERROR:             err,
			}).Warn("attempt to edit track failed")
		} else {
			fmt.Fprintf(w, "%q fixed\n", t)
		}
	}
}

func (r *repair) createBackups(w io.Writer, tracks []*files.Track) {
	albumPaths := getAlbumPaths(tracks)
	r.makeBackupDirectories(w, albumPaths)
	r.backupTracks(w, tracks)
}

func (r *repair) backupTracks(w io.Writer, tracks []*files.Track) {
	for _, track := range tracks {
		r.backupTrack(w, track)
	}
}

func (r *repair) backupTrack(w io.Writer, t *files.Track) {
	backupDir := t.BackupDirectory()
	destinationPath := filepath.Join(backupDir, fmt.Sprintf("%d.mp3", t.TrackNumber))
	if internal.DirExists(backupDir) && !internal.PlainFileExists(destinationPath) {
		if err := t.Copy(destinationPath); err != nil {
			fmt.Fprintf(w, "The track %q cannot be backed up.\n", t)
			logrus.WithFields(logrus.Fields{
				internal.LOG_COMMAND_NAME: r.name(),
				"source":                  t.String,
				"destination":             destinationPath,
				internal.LOG_ERROR:        err,
			}).Info("error backing up file")
		} else {
			fmt.Fprintf(w, "The track %q has been backed up to %q.\n", t, destinationPath)
		}
	}
}

func (r *repair) makeBackupDirectories(w io.Writer, paths []string) {
	for _, path := range paths {
		newPath := filepath.Join(path, files.BackupDirName)
		if !internal.DirExists(newPath) {
			if err := internal.Mkdir(newPath); err != nil {
				fmt.Fprintf(w, internal.USER_CANNOT_CREATE_DIRECTORY, newPath, err)
				logrus.WithFields(logrus.Fields{
					internal.LOG_COMMAND_NAME: r.name(),
					internal.LOG_DIRECTORY:    newPath,
					internal.LOG_ERROR:        err,
				}).Info(internal.LOG_CANNOT_CREATE_DIRECTORY)
			}
		}
	}
}

func getAlbumPaths(tracks []*files.Track) []string {
	albumPaths := map[string]bool{}
	for _, t := range tracks {
		albumPaths[t.AlbumPath()] = true
	}
	var result []string
	for path := range albumPaths {
		result = append(result, path)
	}
	sort.Strings(result)
	return result
}
