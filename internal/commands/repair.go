package commands

import (
	"flag"
	"fmt"
	"mp3/internal"
	"mp3/internal/files"
	"path/filepath"
	"sort"
)

type repair struct {
	n      string
	dryRun *bool
	sf     *files.SearchFlags
}

func (r *repair) name() string {
	return r.n
}

func newRepair(o internal.OutputBus, c *internal.Configuration, fSet *flag.FlagSet) (CommandProcessor, bool) {
	return newRepairCommand(o, c, fSet)
}

const (
	dryRunFlag      = "dryRun"
	defaultDryRun   = false
	fkDestination   = "destination"
	fkDryRunFlag    = "-" + dryRunFlag
	fkSource        = "source"
	noProblemsFound = "No repairable track defects found"
)

func newRepairCommand(o internal.OutputBus, c *internal.Configuration, fSet *flag.FlagSet) (*repair, bool) {
	name := fSet.Name()
	ok := true
	defDryRun, err := c.SubConfiguration(name).BoolDefault(dryRunFlag, defaultDryRun)
	if err != nil {
		reportBadDefault(o, name, err)
		ok = false
	}
	sFlags, sFlagsOk := files.NewSearchFlags(o, c, fSet)
	if sFlagsOk && ok {
		return &repair{
			n:      name,
			dryRun: fSet.Bool(dryRunFlag, defDryRun, "if true, output what would have repaired, but make no repairs"),
			sf:     sFlags,
		}, true
	}
	return nil, false
}

func (r *repair) Exec(o internal.OutputBus, args []string) (ok bool) {
	if s, argsOk := r.sf.ProcessArgs(o, args); argsOk {
		ok = r.runCommand(o, s)
	}
	return
}

func (r *repair) logFields() map[string]interface{} {
	return map[string]interface{}{
		fkCommandName: r.name(),
		fkDryRunFlag:  *r.dryRun,
	}
}

func (r *repair) runCommand(o internal.OutputBus, s *files.Search) (ok bool) {
	o.LogWriter().Info(internal.LI_EXECUTING_COMMAND, r.logFields())
	artists, ok := s.LoadData(o)
	if ok {
		files.UpdateTracks(o, artists, files.RawReadID3V2Tag)
		tracksWithConflicts := findConflictedTracks(artists)
		if len(tracksWithConflicts) == 0 {
			o.WriteConsole(true, noProblemsFound)
		} else {
			if *r.dryRun {
				reportTracks(o, tracksWithConflicts)
			} else {
				r.createBackups(o, tracksWithConflicts)
				r.fixTracks(o, tracksWithConflicts)
			}
		}
	}
	return
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

func reportTracks(o internal.OutputBus, tracks []*files.Track) {
	lastArtistName := ""
	lastAlbumName := ""
	for _, t := range tracks {
		albumName := t.AlbumName()
		artistName := t.RecordingArtist()
		if lastArtistName != artistName {
			o.WriteConsole(false, "%q\n", artistName)
			lastArtistName = artistName
			lastAlbumName = ""
		}
		if albumName != lastAlbumName {
			o.WriteConsole(false, "    %q\n", albumName)
			lastAlbumName = albumName
		}
		s := t.AnalyzeIssues()
		o.WriteConsole(false, "        %2d %q need to repair%s%s%s%s\n",
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

func (r *repair) fixTracks(o internal.OutputBus, tracks []*files.Track) {
	for _, t := range tracks {
		if err := t.EditID3V2Tag(); err != nil {
			o.WriteError(internal.USER_ERROR_REPAIRING_TRACK_FILE, t)
			o.LogWriter().Error(internal.LE_CANNOT_EDIT_TRACK, map[string]interface{}{
				internal.LI_EXECUTING_COMMAND: r.name(),
				internal.FK_DIRECTORY:         t.Directory(),
				internal.FK_FILE_NAME:         t.FileName(),
				internal.FK_ERROR:             err,
			})
		} else {
			o.WriteConsole(false, "%q repaired.\n", t)
		}
	}
}

func (r *repair) createBackups(o internal.OutputBus, tracks []*files.Track) {
	albumPaths := getAlbumPaths(tracks)
	r.makeBackupDirectories(o, albumPaths)
	r.backupTracks(o, tracks)
}

func (r *repair) backupTracks(o internal.OutputBus, tracks []*files.Track) {
	for _, track := range tracks {
		r.backupTrack(o, track)
	}
}

func (r *repair) backupTrack(o internal.OutputBus, t *files.Track) {
	backupDir := t.BackupDirectory()
	destinationPath := filepath.Join(backupDir, fmt.Sprintf("%d.mp3", t.Number()))
	if internal.DirExists(backupDir) && !internal.PlainFileExists(destinationPath) {
		if err := t.Copy(destinationPath); err != nil {
			o.WriteError(internal.USER_ERROR_CREATING_BACKUP_FILE, t)
			o.LogWriter().Error(internal.LE_CANNOT_COPY_FILE, map[string]interface{}{
				fkCommandName:     r.name(),
				fkSource:          t.Path(),
				fkDestination:     destinationPath,
				internal.FK_ERROR: err,
			})
		} else {
			o.WriteConsole(true, "The track %q has been backed up to %q", t, destinationPath)
		}
	}
}

func (r *repair) makeBackupDirectories(o internal.OutputBus, paths []string) {
	for _, path := range paths {
		newPath := files.CreateBackupPath(path)
		if !internal.DirExists(newPath) {
			if err := internal.Mkdir(newPath); err != nil {
				o.WriteError(internal.USER_CANNOT_CREATE_DIRECTORY, newPath, err)
				o.LogWriter().Error(internal.LE_CANNOT_CREATE_DIRECTORY, map[string]interface{}{
					fkCommandName:         r.name(),
					internal.FK_DIRECTORY: newPath,
					internal.FK_ERROR:     err,
				})
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
