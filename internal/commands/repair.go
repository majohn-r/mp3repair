package commands

import (
	"flag"
	"fmt"
	"mp3/internal"
	"mp3/internal/files"
	"mp3/internal/output"
	"path/filepath"
	"sort"
)

func init() {
	addCommandData(repairCommandName, commandData{isDefault: false, initFunction: newRepair})
	addDefaultMapping(repairCommandName, map[string]any{
		dryRunFlag: defaultDryRun,
	})
}

type repair struct {
	dryRun *bool
	sf     *files.SearchFlags
}

func newRepair(o output.Bus, c *internal.Configuration, fSet *flag.FlagSet) (CommandProcessor, bool) {
	return newRepairCommand(o, c, fSet)
}

const (
	repairCommandName = "repair"

	dryRunFlag    = "dryRun"
	defaultDryRun = false

	fieldKeyDestination = "destination"
	fieldKeyDryRunFlag  = "-" + dryRunFlag
	fieldKeySource      = "source"

	noProblemsFound = "No repairable track defects found"
)

func newRepairCommand(o output.Bus, c *internal.Configuration, fSet *flag.FlagSet) (*repair, bool) {
	ok := true
	defDryRun, err := c.SubConfiguration(repairCommandName).BoolDefault(dryRunFlag, defaultDryRun)
	if err != nil {
		reportBadDefault(o, repairCommandName, err)
		ok = false
	}
	sFlags, sFlagsOk := files.NewSearchFlags(o, c, fSet)
	if sFlagsOk && ok {
		dryRunUsage := internal.DecorateBoolFlagUsage("output what would have been repaired, but make no repairs", defDryRun)
		return &repair{
			dryRun: fSet.Bool(dryRunFlag, defDryRun, dryRunUsage),
			sf:     sFlags,
		}, true
	}
	return nil, false
}

func (r *repair) Exec(o output.Bus, args []string) (ok bool) {
	if s, argsOk := r.sf.ProcessArgs(o, args); argsOk {
		ok = r.runCommand(o, s)
	}
	return
}

func (r *repair) logFields() map[string]any {
	return map[string]any{
		fieldKeyCommandName: repairCommandName,
		fieldKeyDryRunFlag:  *r.dryRun,
	}
}

func (r *repair) runCommand(o output.Bus, s *files.Search) (ok bool) {
	o.Log(output.Info, internal.LogInfoExecutingCommand, r.logFields())
	artists, ok := s.LoadData(o)
	if ok {
		files.ReadMetadata(o, artists)
		tracksWithConflicts := findConflictedTracks(artists)
		if len(tracksWithConflicts) == 0 {
			o.WriteCanonicalConsole(noProblemsFound)
		} else {
			if *r.dryRun {
				reportTracks(o, tracksWithConflicts)
			} else {
				createBackups(o, tracksWithConflicts)
				fixTracks(o, tracksWithConflicts)
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
				if state := track.ReconcileMetadata(); state.HasTaggingConflicts() {
					t = append(t, track)
				}
			}
		}
	}
	sort.Sort(files.Tracks(t))
	return t
}

func reportTracks(o output.Bus, tracks []*files.Track) {
	lastArtistName := ""
	lastAlbumName := ""
	for _, t := range tracks {
		albumName := t.AlbumName()
		artistName := t.RecordingArtist()
		if lastArtistName != artistName {
			o.WriteConsole("%q\n", artistName)
			lastArtistName = artistName
			lastAlbumName = ""
		}
		if albumName != lastAlbumName {
			o.WriteConsole("    %q\n", albumName)
			lastAlbumName = albumName
		}
		s := t.ReconcileMetadata()
		o.WriteConsole("        %2d %q need to repair%s%s%s%s\n",
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

func fixTracks(o output.Bus, tracks []*files.Track) {
	for _, t := range tracks {
		if err := t.EditTags(); len(err) != 0 {
			o.WriteCanonicalError(internal.UserErrorRepairingTrackFile, t)
			o.Log(output.Error, internal.LogErrorCannotEditTrack, map[string]any{
				internal.LogInfoExecutingCommand: repairCommandName,
				internal.FieldKeyDirectory:       t.Directory(),
				internal.FieldKeyFileName:        t.FileName(),
				internal.FieldKeyError:           err,
			})
		} else {
			o.WriteConsole("%q repaired.\n", t)
			MarkDirty(o)
		}
	}
}

func createBackups(o output.Bus, tracks []*files.Track) {
	albumPaths := getAlbumPaths(tracks)
	makeBackupDirectories(o, albumPaths)
	backupTracks(o, tracks)
}

func backupTracks(o output.Bus, tracks []*files.Track) {
	for _, track := range tracks {
		backupTrack(o, track)
	}
}

func backupTrack(o output.Bus, t *files.Track) {
	backupDir := t.BackupDirectory()
	destinationPath := filepath.Join(backupDir, fmt.Sprintf("%d.mp3", t.Number()))
	if internal.DirExists(backupDir) && !internal.PlainFileExists(destinationPath) {
		if err := t.Copy(destinationPath); err != nil {
			o.WriteCanonicalError(internal.UserErrorCreatingBackupFile, t)
			o.Log(output.Error, internal.LogErrorCannotCopyFile, map[string]any{
				fieldKeyCommandName:    repairCommandName,
				fieldKeySource:         t.Path(),
				fieldKeyDestination:    destinationPath,
				internal.FieldKeyError: err,
			})
		} else {
			o.WriteCanonicalConsole("The track %q has been backed up to %q", t, destinationPath)
		}
	}
}

func makeBackupDirectories(o output.Bus, paths []string) {
	for _, path := range paths {
		newPath := files.CreateBackupPath(path)
		if !internal.DirExists(newPath) {
			if err := internal.Mkdir(newPath); err != nil {
				o.WriteCanonicalError(internal.UserCannotCreateDirectory, newPath, err)
				o.Log(output.Error, internal.LogErrorCannotCreateDirectory, map[string]any{
					fieldKeyCommandName:        repairCommandName,
					internal.FieldKeyDirectory: newPath,
					internal.FieldKeyError:     err,
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
