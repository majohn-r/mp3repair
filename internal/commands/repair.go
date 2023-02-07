package commands

import (
	"flag"
	"fmt"
	"mp3/internal/files"
	"path/filepath"
	"sort"

	tools "github.com/majohn-r/cmd-toolkit"
	"github.com/majohn-r/output"
)

func init() {
	tools.AddCommandData(repairCommandName, &tools.CommandDescription{IsDefault: IsDefault(repairCommandName), Initializer: newRepair})
	addDefaultMapping(repairCommandName, map[string]any{
		dryRunFlag: defaultDryRun,
	})
}

type repair struct {
	dryRun *bool
	sf     *files.SearchFlags
}

func newRepair(o output.Bus, c *tools.Configuration, fSet *flag.FlagSet) (tools.CommandProcessor, bool) {
	return newRepairCommand(o, c, fSet)
}

const (
	repairCommandName = "repair"

	dryRunFlag    = "dryRun"
	defaultDryRun = false
)

func newRepairCommand(o output.Bus, c *tools.Configuration, fSet *flag.FlagSet) (*repair, bool) {
	if defDryRun, err := c.SubConfiguration(repairCommandName).BoolDefault(dryRunFlag, defaultDryRun); err != nil {
		tools.ReportInvalidConfigurationData(o, repairCommandName, err)
	} else {
		if sFlags, ok := files.NewSearchFlags(o, c, fSet); ok {
			dryRunUsage := tools.DecorateBoolFlagUsage("output what would have been repaired, but make no repairs", defDryRun)
			return &repair{dryRun: fSet.Bool(dryRunFlag, defDryRun, dryRunUsage), sf: sFlags}, true
		}
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
	return map[string]any{"command": repairCommandName, "-" + dryRunFlag: *r.dryRun}
}

func (r *repair) runCommand(o output.Bus, s *files.Search) (ok bool) {
	tools.LogCommandStart(o, repairCommandName, r.logFields())
	if artists, loaded := s.Load(o); loaded {
		ok = true
		files.ReadMetadata(o, artists)
		if tracks := findConflictedTracks(artists); len(tracks) == 0 {
			o.WriteCanonicalConsole("No repairable track defects found")
		} else {
			if *r.dryRun {
				reportTracks(o, tracks)
			} else {
				createBackups(o, tracks)
				fixTracks(o, tracks)
			}
		}
	}
	return
}

func findConflictedTracks(artists []*files.Artist) []*files.Track {
	var tracks []*files.Track
	for _, aR := range artists {
		for _, aL := range aR.Albums() {
			for _, t := range aL.Tracks() {
				if state := t.ReconcileMetadata(); state.HasConflicts() {
					tracks = append(tracks, t)
				}
			}
		}
	}
	sort.Sort(files.Tracks(tracks))
	return tracks
}

func reportTracks(o output.Bus, tracks []*files.Track) {
	lastArtist := ""
	lastAlbum := ""
	for _, t := range tracks {
		album := t.AlbumName()
		artist := t.RecordingArtist()
		if lastArtist != artist {
			o.WriteConsole("%q\n", artist)
			lastArtist = artist
			lastAlbum = ""
		}
		if album != lastAlbum {
			o.WriteConsole("    %q\n", album)
			lastAlbum = album
		}
		s := t.ReconcileMetadata()
		o.WriteConsole("        %2d %q need to repair%s%s%s%s\n", t.Number(), t.CommonName(),
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
		if err := t.UpdateMetadata(); len(err) != 0 {
			o.WriteCanonicalError("An error occurred repairing track %q", t)
			o.Log(output.Error, "cannot edit track", map[string]any{
				"command":   repairCommandName,
				"directory": t.Directory(),
				"fileName":  t.FileName(),
				"error":     err,
			})
		} else {
			o.WriteConsole("%q repaired.\n", t)
			markDirty(o, repairCommandName)
		}
	}
}

func createBackups(o output.Bus, tracks []*files.Track) {
	makeBackupDirectories(o, albumPaths(tracks))
	backupTracks(o, tracks)
}

func backupTracks(o output.Bus, tracks []*files.Track) {
	for _, track := range tracks {
		backupTrack(o, track)
	}
}

func backupTrack(o output.Bus, t *files.Track) {
	dir := t.BackupDirectory()
	backupFile := filepath.Join(dir, fmt.Sprintf("%d.mp3", t.Number()))
	if tools.DirExists(dir) && !tools.PlainFileExists(backupFile) {
		if err := t.CopyFile(backupFile); err != nil {
			o.WriteCanonicalError("The track %q cannot be backed up", t)
			o.Log(output.Error, "error copying file", map[string]any{
				"command":     repairCommandName,
				"source":      t.Path(),
				"destination": backupFile,
				"error":       err,
			})
		} else {
			o.WriteCanonicalConsole("The track %q has been backed up to %q", t, backupFile)
		}
	}
}

func makeBackupDirectories(o output.Bus, paths []string) {
	for _, path := range paths {
		newPath := files.CreateBackupPath(path)
		if !tools.DirExists(newPath) {
			if err := tools.Mkdir(newPath); err != nil {
				tools.ReportDirectoryCreationFailure(o, repairCommandName, newPath, err)
			}
		}
	}
}

func albumPaths(tracks []*files.Track) []string {
	m := map[string]bool{}
	for _, t := range tracks {
		m[t.AlbumPath()] = true
	}
	var paths []string
	for path := range m {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return paths
}
