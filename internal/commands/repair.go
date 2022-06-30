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

func newRepair(c *internal.Configuration, fSet *flag.FlagSet) CommandProcessor {
	return newRepairCommand(c, fSet)
}

const (
	dryRunFlag      = "dryRun"
	defaultDryRun   = false
	fkDestination   = "destination"
	fkDryRunFlag    = "-" + dryRunFlag
	fkSource        = "source"
	noProblemsFound = "No repairable track defects found"
)

func newRepairCommand(c *internal.Configuration, fSet *flag.FlagSet) *repair {
	name := fSet.Name()
	configuration := c.SubConfiguration(name)
	return &repair{
		n: name,
		dryRun: fSet.Bool(dryRunFlag,
			configuration.BoolDefault(dryRunFlag, defaultDryRun),
			"if true, output what would have repaired, but make no repairs"),
		sf: files.NewSearchFlags(c, fSet),
	}
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
	o.LogWriter().Log(internal.INFO, internal.LI_EXECUTING_COMMAND, r.logFields())
	artists, ok := s.LoadData(o.ErrorWriter())
	if ok {
		files.UpdateTracks(artists, files.RawReadTags)
		tracksWithConflicts := findConflictedTracks(artists)
		if len(tracksWithConflicts) == 0 {
			fmt.Fprintln(o.ConsoleWriter(), noProblemsFound)
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
			fmt.Fprintf(o.ConsoleWriter(), "%q\n", artistName)
			lastArtistName = artistName
			lastAlbumName = ""
		}
		if albumName != lastAlbumName {
			fmt.Fprintf(o.ConsoleWriter(), "    %q\n", albumName)
			lastAlbumName = albumName
		}
		s := t.AnalyzeIssues()
		fmt.Fprintf(o.ConsoleWriter(), "        %2d %q need to fix%s%s%s%s\n",
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
		if err := t.EditTags(); err != nil {
			fmt.Fprintf(o.ErrorWriter(), "An error occurred fixing track %q\n", t)
			o.LogWriter().Log(internal.WARN, internal.LW_CANNOT_EDIT_TRACK, map[string]interface{}{
				internal.LI_EXECUTING_COMMAND: r.name(),
				internal.FK_DIRECTORY:         t.Directory(),
				internal.FK_FILE_NAME:         t.FileName(),
				internal.FK_ERROR:             err,
			})
		} else {
			fmt.Fprintf(o.ConsoleWriter(), "%q fixed\n", t)
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
			fmt.Fprintf(o.ErrorWriter(), "The track %q cannot be backed up.\n", t)
			o.LogWriter().Log(internal.WARN, internal.LW_CANNOT_COPY_FILE, map[string]interface{}{
				fkCommandName:     r.name(),
				fkSource:          t.Path(),
				fkDestination:     destinationPath,
				internal.FK_ERROR: err,
			})
		} else {
			fmt.Fprintf(o.ConsoleWriter(), "The track %q has been backed up to %q.\n", t, destinationPath)
		}
	}
}

func (r *repair) makeBackupDirectories(o internal.OutputBus, paths []string) {
	for _, path := range paths {
		newPath := files.CreateBackupPath(path)
		if !internal.DirExists(newPath) {
			if err := internal.Mkdir(newPath); err != nil {
				fmt.Fprintf(o.ErrorWriter(), internal.USER_CANNOT_CREATE_DIRECTORY, newPath, err)
				o.LogWriter().Log(internal.WARN, internal.LW_CANNOT_CREATE_DIRECTORY, map[string]interface{}{
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
