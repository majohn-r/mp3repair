package cmd

import (
	"fmt"
	cmdtoolkit "github.com/majohn-r/cmd-toolkit"
	"mp3repair/internal/files"
	"path/filepath"
	"slices"
	"strings"

	"github.com/majohn-r/output"
	"github.com/spf13/cobra"
)

const (
	repairCommandName = "repair"
	repairDryRun      = "dryRun"
	repairDryRunFlag  = "--" + repairDryRun
)

var (
	repairCmd = &cobra.Command{
		Use: repairCommandName + " [" + repairDryRunFlag + "] " +
			searchUsage,
		DisableFlagsInUseLine: true,
		Short: "Repairs problems found by running '" + checkCommand + " " +
			checkFilesFlag + "'",
		Long: "" +
			fmt.Sprintf("%q repairs the problems found by running '%s %s'\n",
				repairCommandName, checkCommand, checkFilesFlag) +
			"\n" +
			"This command rewrites the mp3 files that the " + checkCommand +
			" command noted as having metadata\n" +
			"inconsistent with the file structure. Prior to rewriting an mp3 file, the " +
			repairCommandName + "\n" +
			"command creates a backup directory for the parent album and copies the" +
			" original mp3\n" +
			"file into that backup directory. Use the " + postRepairCommandName +
			" command to automatically delete\n" +
			"the backup folders.",
		Example: repairCommandName + " " + repairDryRunFlag + "\n" +
			"  Output what would be repaired, but does not perform the stated repairs",
		RunE: repairRun,
	}
	repairFlags = &cmdtoolkit.FlagSet{
		Name: repairCommandName,
		Details: map[string]*cmdtoolkit.FlagDetails{
			"dryRun": {
				Usage:        "output what would have been repaired, but make no repairs",
				ExpectedType: cmdtoolkit.BoolType,
				DefaultValue: false,
			},
		},
	}
)

func repairRun(cmd *cobra.Command, _ []string) error {
	exitError := cmdtoolkit.NewExitProgrammingError(repairCommandName)
	o := getBus()
	producer := cmd.Flags()
	values, eSlice := cmdtoolkit.ReadFlags(producer, repairFlags)
	searchSettings, searchFlagsOk := EvaluateSearchFlags(o, producer)
	if cmdtoolkit.ProcessFlagErrors(o, eSlice) && searchFlagsOk {
		if rs, flagsOk := processRepairFlags(o, values); flagsOk {
			details := map[string]any{repairDryRunFlag: rs.dryRun.Value}
			for k, v := range searchSettings.Values() {
				details[k] = v
			}
			logCommandStart(o, repairCommandName, details)
			allArtists := searchSettings.Load(o)
			exitError = rs.processArtists(o, allArtists, searchSettings)
		}
	}
	return cmdtoolkit.ToErrorInterface(exitError)
}

type repairSettings struct {
	dryRun cmdtoolkit.CommandFlag[bool]
}

func (rs *repairSettings) processArtists(o output.Bus, allArtists []*files.Artist, ss *SearchSettings) (e *cmdtoolkit.ExitError) {
	e = cmdtoolkit.NewExitUserError(repairCommandName)
	if len(allArtists) != 0 {
		if filteredArtists := ss.Filter(o, allArtists); len(filteredArtists) != 0 {
			e = rs.repairArtists(o, filteredArtists)
		}
	}
	return
}

func (rs *repairSettings) repairArtists(o output.Bus, artists []*files.Artist) *cmdtoolkit.ExitError {
	readMetadata(o, artists) // read all track metadata
	concernedArtists := createConcernedArtists(artists)
	count := findConflictedTracks(concernedArtists)
	if rs.dryRun.Value {
		reportRepairsNeeded(o, concernedArtists)
		return nil
	}
	if count == 0 {
		nothingToDo(o)
		return nil
	}
	return backupAndRepairTracks(o, concernedArtists)
}

func findConflictedTracks(concernedArtists []*concernedArtist) int {
	count := 0
	for _, cAr := range concernedArtists {
		for _, cAl := range cAr.albums() {
			for _, cT := range cAl.tracks() {
				state := cT.backing.ReconcileMetadata()
				if state.HasArtistNameConflict() {
					cT.addConcern(conflictConcern,
						"the artist name field does not match the name of the artist"+
							" directory")
				}
				if state.HasAlbumNameConflict() {
					cT.addConcern(conflictConcern,
						"the album name field does not match the name of the album"+
							" directory")
				}
				if state.HasGenreConflict() {
					cT.addConcern(conflictConcern,
						"the genre field does not match the other tracks in the album")
				}
				if state.HasMCDIConflict() {
					cT.addConcern(conflictConcern,
						"the music CD identifier field does not match the other tracks in"+
							" the album")
				}
				if state.HasNumberingConflict() {
					cT.addConcern(conflictConcern,
						"the track number field does not match the track's file name")
				}
				if state.HasTrackNameConflict() {
					cT.addConcern(conflictConcern,
						"the track name field does not match the track's file name")
				}
				if state.HasYearConflict() {
					cT.addConcern(conflictConcern,
						"the year field does not match the other tracks in the album")
				}
				if cT.isConcerned() {
					count++
				}
			}
		}
	}
	return count
}

func reportRepairsNeeded(o output.Bus, concernedArtists []*concernedArtist) {
	artistNames := make([]string, 0, len(concernedArtists))
	artistMap := map[string]*concernedArtist{}
	for _, cAr := range concernedArtists {
		name := cAr.name()
		artistNames = append(artistNames, name)
		artistMap[name] = cAr
	}
	slices.Sort(artistNames)
	headerPrinted := false
	for _, name := range artistNames {
		if cAr := artistMap[name]; cAr != nil {
			if cAr.isConcerned() {
				if !headerPrinted {
					o.WriteConsole("The following concerns can be repaired:\n")
					headerPrinted = true
				}
				cAr.toConsole(o)
			}
		}
	}
	if !headerPrinted {
		nothingToDo(o)
	}
}

func nothingToDo(o output.Bus) {
	o.WriteCanonicalConsole("No repairable track defects were found.")
}

func backupAndRepairTracks(o output.Bus, concernedArtists []*concernedArtist) *cmdtoolkit.ExitError {
	var e *cmdtoolkit.ExitError
	for _, cAr := range concernedArtists {
		if !cAr.isConcerned() {
			continue
		}
		for _, cAl := range cAr.concernedAlbums {
			if !cAl.isConcerned() {
				continue
			}
			path, exists := ensureTrackBackupDirectoryExists(o, cAl)
			if !exists {
				e = cmdtoolkit.NewExitSystemError(repairCommandName)
				continue
			}
			for _, cT := range cAl.concernedTracks {
				if !cT.isConcerned() {
					continue
				}
				t := cT.backing
				if !tryTrackBackup(o, t, path) {
					e = cmdtoolkit.NewExitSystemError(repairCommandName)
					continue
				}
				err := t.UpdateMetadata()
				if e2 := processTrackRepairResults(o, t, err); e2 != nil {
					e = e2
				}
			}
		}
	}
	return e
}

func processTrackRepairResults(o output.Bus, t *files.Track, updateErrs []error) *cmdtoolkit.ExitError {
	if len(updateErrs) != 0 {
		o.WriteCanonicalError("An error occurred repairing track %q", t)
		errorStrings := make([]string, 0, len(updateErrs))
		for _, e2 := range updateErrs {
			errorStrings = append(errorStrings, fmt.Sprintf("%q", e2.Error()))
		}
		o.Log(output.Error, "cannot edit track", map[string]any{
			"command":   repairCommandName,
			"directory": t.Directory(),
			"fileName":  t.FileName(),
			"error":     fmt.Sprintf("[%s]", strings.Join(errorStrings, ", ")),
		})
		return cmdtoolkit.NewExitSystemError(repairCommandName)
	}
	o.WriteConsole("%q repaired.\n", t)
	markDirty(o)
	return nil
}

func tryTrackBackup(o output.Bus, t *files.Track, path string) (backedUp bool) {
	backupFile := filepath.Join(path, fmt.Sprintf("%d.mp3", t.Number))
	switch {
	case plainFileExists(backupFile):
		o.WriteCanonicalError("The backup file for track file %q, %q, already exists", t,
			backupFile)
		o.Log(output.Error, "file already exists", map[string]any{
			"command": repairCommandName,
			"file":    backupFile,
		})
	default:
		copyErr := copyFile(t.FilePath, backupFile)
		switch copyErr {
		case nil:
			o.WriteCanonicalConsole("The track file %q has been backed up to %q", t,
				backupFile)
			backedUp = true
		default:
			o.WriteCanonicalError(
				"The track file %q could not be backed up due to error %v", t, copyErr)
			o.Log(output.Error, "error copying file", map[string]any{
				"command":     repairCommandName,
				"source":      t.FilePath,
				"destination": backupFile,
				"error":       copyErr,
			})
		}
	}
	if !backedUp {
		o.WriteCanonicalError("The track file %q will not be repaired", t)
	}
	return
}

func ensureTrackBackupDirectoryExists(o output.Bus, cAl *concernedAlbum) (path string, exists bool) {
	path = cAl.backing.BackupDirectory()
	exists = true
	if !dirExists(path) {
		if fileErr := mkdir(path); fileErr != nil {
			exists = false
			o.WriteCanonicalError("The directory %q cannot be created: %v", path, fileErr)
			o.WriteCanonicalError(
				"The track files in the directory %q will not be repaired",
				cAl.backing.FilePath)
			o.Log(output.Error, "cannot create directory", map[string]any{
				"command":   repairCommandName,
				"directory": path,
				"error":     fileErr,
			})
		}
	}
	return
}

func processRepairFlags(o output.Bus, values map[string]*cmdtoolkit.CommandFlag[any]) (*repairSettings, bool) {
	rs := &repairSettings{}
	flagsOk := true // optimistic
	var flagErr error
	if rs.dryRun, flagErr = cmdtoolkit.GetBool(o, values, repairDryRun); flagErr != nil {
		flagsOk = false
	}
	return rs, flagsOk
}

func init() {
	RootCmd.AddCommand(repairCmd)
	addDefaults(repairFlags)
	o := getBus()
	c := getConfiguration()
	cmdtoolkit.AddFlags(o, c, repairCmd.Flags(), repairFlags, SearchFlags)
}
