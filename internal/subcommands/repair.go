package subcommands

import (
	"flag"
	"fmt"
	"io"
	"mp3/internal"
	"mp3/internal/files"
	"path/filepath"
	"sort"

	"github.com/sirupsen/logrus"
)

type repair struct {
	n      string
	dryRun *bool
	sf     *files.SearchFlags
}

func (r *repair) name() string {
	return r.n
}

func newRepair(c *internal.Configuration, fSet *flag.FlagSet) CommandProcessor {
	return newRepairSubCommand(c, fSet)
}

const (
	dryRunFlag    = "dryRun"
	defaultDryRun = false
)

func newRepairSubCommand(c *internal.Configuration, fSet *flag.FlagSet) *repair {
	configuration := c.SubConfiguration("repair")
	return &repair{
		n: fSet.Name(),
		dryRun: fSet.Bool(dryRunFlag,
			configuration.BoolDefault(dryRunFlag, defaultDryRun),
			"if true, output what would have repaired, but make no repairs"),
		sf: files.NewSearchFlags(c, fSet),
	}
}

// TODO: rewrite unit test
func (r *repair) Exec(wOut io.Writer, wErr io.Writer, args []string) (ok bool) {
	if s := r.sf.ProcessArgs(wErr, args); s != nil {
		r.runSubcommand(wOut, s)
		ok = true
	}
	return
}

func (r *repair) logFields() logrus.Fields {
	return logrus.Fields{
		internal.FK_COMMAND_NAME: r.name(),
		internal.FK_DRY_RUN_FLAG: *r.dryRun,
	}
}

func (r *repair) runSubcommand(w io.Writer, s *files.Search) {
	logrus.WithFields(r.logFields()).Info(internal.LI_EXECUTING_COMMAND)
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
		albumName := t.AlbumName()
		artistName := t.RecordingArtist()
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
			t.Number(), t.Name(),
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
			// TODO: should be a 2nd writer - stderr
			fmt.Fprintf(w, "An error occurred fixing track %q\n", t)
			logrus.WithFields(logrus.Fields{
				internal.LI_EXECUTING_COMMAND: r.name(),
				internal.FK_DIRECTORY:         t.Directory(),
				internal.FK_FILE_NAME:         t.FileName(),
				internal.FK_ERROR:             err,
			}).Warn(internal.LW_CANNOT_EDIT_TRACK)
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

// TODO: need 2nd writer for errors
func (r *repair) backupTrack(w io.Writer, t *files.Track) {
	backupDir := t.BackupDirectory()
	destinationPath := filepath.Join(backupDir, fmt.Sprintf("%d.mp3", t.Number()))
	if internal.DirExists(backupDir) && !internal.PlainFileExists(destinationPath) {
		if err := t.Copy(destinationPath); err != nil {
			// TODO: use 2nd writer
			fmt.Fprintf(w, "The track %q cannot be backed up.\n", t)
			logrus.WithFields(logrus.Fields{
				internal.FK_COMMAND_NAME: r.name(),
				internal.FK_SOURCE:       t.Path(),
				internal.FK_DESTINATION:  destinationPath,
				internal.FK_ERROR:        err,
			}).Warn(internal.LW_CANNOT_COPY_FILE)
		} else {
			fmt.Fprintf(w, "The track %q has been backed up to %q.\n", t, destinationPath)
		}
	}
}

// TODO: w should be an error output
func (r *repair) makeBackupDirectories(w io.Writer, paths []string) {
	for _, path := range paths {
		newPath := files.CreateBackupPath(path)
		if !internal.DirExists(newPath) {
			if err := internal.Mkdir(newPath); err != nil {
				fmt.Fprintf(w, internal.USER_CANNOT_CREATE_DIRECTORY, newPath, err)
				logrus.WithFields(logrus.Fields{
					internal.FK_COMMAND_NAME: r.name(),
					internal.FK_DIRECTORY:    newPath,
					internal.FK_ERROR:        err,
				}).Warn(internal.LW_CANNOT_CREATE_DIRECTORY)
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
