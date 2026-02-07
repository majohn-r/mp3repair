package cmd

import (
	"fmt"
	"mp3repair/internal/files"
	"path/filepath"
	"slices"
	"strings"

	cmdtoolkit "github.com/majohn-r/cmd-toolkit"

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
		Use:                   repairCommandName + " [" + repairDryRunFlag + "] " + searchUsage + " " + ioUsage,
		DisableFlagsInUseLine: true,
		Short:                 "Repairs problems found by running '" + scanCommand + " " + scanFilesFlag + "'",
		Long: "" +
			fmt.Sprintf("%q repairs the problems found by running '%s %s'\n",
				repairCommandName, scanCommand, scanFilesFlag) +
			"\n" +
			"This command rewrites the mp3 files that the " + scanCommand + " command noted as having metadata\n" +
			"inconsistent with the file structure. Prior to rewriting an mp3 file, the " + repairCommandName + "\n" +
			"command creates a backup directory for the parent album and copies the" + " original mp3\n" +
			"file into that backup directory. Use the " + postRepairCommandName + " command to automatically delete\n" +
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
	ss, searchFlagsOk := evaluateSearchFlags(o, producer)
	ios, ioFlagsOk := evaluateIOFlags(o, producer)
	if cmdtoolkit.ProcessFlagErrors(o, eSlice) && searchFlagsOk && ioFlagsOk {
		if rs, flagsOk := processRepairFlags(o, values); flagsOk {
			exitError = rs.processArtists(o, ss.load(o), ss, ios)
		}
	}
	return cmdtoolkit.ToErrorInterface(exitError)
}

type repairSettings struct {
	dryRun cmdtoolkit.CommandFlag[bool]
}

func (rs *repairSettings) processArtists(
	o output.Bus,
	allArtists []*files.Artist,
	ss *searchSettings,
	ios *ioSettings,
) (e *cmdtoolkit.ExitError) {
	e = cmdtoolkit.NewExitUserError(repairCommandName)
	if len(allArtists) != 0 {
		if filteredArtists := ss.filter(o, allArtists); len(filteredArtists) != 0 {
			e = rs.repairArtists(o, filteredArtists, ios)
		}
	}
	return
}

func (rs *repairSettings) repairArtists(o output.Bus, artists []*files.Artist, ios *ioSettings) *cmdtoolkit.ExitError {
	// read all track metadata
	readMetadata(o, artists, ios.openFileLimit)
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
				// this awkward declaration keeps helpful (sic) IDEs from declaring that
				// MetadataState need not be public
				var state files.MetadataState
				state = cT.backing.ReconcileMetadata()
				if state.HasArtistNameConflict() {
					cT.addConcern(conflictConcern,
						"the artist name field does not match the name of the artist directory")
				}
				if state.HasAlbumNameConflict() {
					cT.addConcern(conflictConcern,
						"the album name field does not match the name of the album directory")
				}
				if state.HasGenreConflict() {
					cT.addConcern(conflictConcern,
						"the genre field does not match the other tracks in the album")
				}
				if state.HasMCDIConflict() {
					cT.addConcern(conflictConcern,
						"the music CD identifier field does not match the other tracks in the album")
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
					o.ConsolePrintln("The following concerns can be repaired:")
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
	o.ConsolePrintln("No repairable track defects were found.")
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
		o.ErrorPrintf("An error occurred repairing track %q.\n", t)
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
	o.ConsolePrintf("%q repaired.\n", t)
	markDirty(o)
	return nil
}

func tryTrackBackup(o output.Bus, t *files.Track, path string) (backedUp bool) {
	backupFile := filepath.Join(path, fmt.Sprintf("%d.mp3", t.Number()))
	switch {
	case plainFileExists(backupFile):
		backedUp = true
		var status string
		if modTime, err := modificationTime(backupFile); err == nil {
			status = modTime.Format("2006-01-02 15:04:05")
		} else {
			status = fmt.Sprintf("error getting modification time: %v", err)
		}
		o.Log(output.Info, "file already exists", map[string]any{
			"command": repairCommandName,
			"file":    backupFile,
			"modTime": status,
		})
	default:
		copyErr := copyFile(t.Path(), backupFile)
		switch copyErr {
		case nil:
			o.ConsolePrintf("The track file %q has been backed up to %q.\n", t, backupFile)
			backedUp = true
		default:
			o.ErrorPrintf(
				"The track file %q could not be backed up due to error %s.\n",
				t,
				cmdtoolkit.ErrorToString(copyErr),
			)
			o.Log(output.Error, "error copying file", map[string]any{
				"command":     repairCommandName,
				"source":      t.Path(),
				"destination": backupFile,
				"error":       copyErr,
			})
		}
	}
	if !backedUp {
		o.ErrorPrintf("The track file %q will not be repaired.\n", t)
	}
	return
}

func ensureTrackBackupDirectoryExists(o output.Bus, cAl *concernedAlbum) (path string, exists bool) {
	path = cAl.backing.BackupDirectory()
	exists = true
	if !dirExists(path) {
		if fileErr := mkdir(path); fileErr != nil {
			exists = false
			o.ErrorPrintf("The directory %q cannot be created: %s.\n", path, cmdtoolkit.ErrorToString(fileErr))
			o.ErrorPrintf("The track files in the directory %q will not be repaired.\n", cAl.backing.Directory())
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
	rootCmd.AddCommand(repairCmd)
	cmdtoolkit.AddDefaults(repairFlags)
	cmdtoolkit.AddFlags(getBus(), getConfiguration(), repairCmd.Flags(), repairFlags, searchFlags, ioFlags)
}
