/*
Copyright © 2021 Marc Johnson (marc.johnson27591@gmail.com)
*/
package cmd

import (
	"fmt"
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
	// RepairCmd represents the repair command
	RepairCmd = &cobra.Command{
		Use: repairCommandName + " [" + repairDryRunFlag + "] " +
			searchUsage,
		DisableFlagsInUseLine: true,
		Short: "Repairs problems found by running '" + CheckCommand + " " +
			CheckFilesFlag + "'",
		Long: "" +
			fmt.Sprintf("%q repairs the problems found by running '%s %s'\n",
				repairCommandName, CheckCommand, CheckFilesFlag) +
			"\n" +
			"This command rewrites the mp3 files that the " + CheckCommand +
			" command noted as having metadata\n" +
			"inconsistent with the file structure. Prior to rewriting an mp3 file, the " +
			repairCommandName + "\n" +
			"command creates a backup directory for the parent album and copies the" +
			" original mp3\n" +
			"file into that backup directory. Use the " + postRepairCommandName +
			" command to automatically delete\n" +
			"the backup folders.",
		RunE: RepairRun,
	}
	RepairFlags = NewSectionFlags().WithSectionName(repairCommandName).WithFlags(
		map[string]*FlagDetails{
			"dryRun": NewFlagDetails().WithUsage(
				"output what would have been repaired, but make no repairs",
			).WithExpectedType(BoolType).WithDefaultValue(false),
		},
	)
)

func RepairRun(cmd *cobra.Command, _ []string) error {
	exitError := NewExitProgrammingError(repairCommandName)
	o := getBus()
	producer := cmd.Flags()
	values, eSlice := ReadFlags(producer, RepairFlags)
	searchSettings, searchFlagsOk := EvaluateSearchFlags(o, producer)
	if ProcessFlagErrors(o, eSlice) && searchFlagsOk {
		if rs, flagsOk := ProcessRepairFlags(o, values); flagsOk {
			details := map[string]any{repairDryRunFlag: rs.dryRun}
			for k, v := range searchSettings.Values() {
				details[k] = v
			}
			LogCommandStart(o, repairCommandName, details)
			allArtists, loaded := searchSettings.Load(o)
			exitError = rs.ProcessArtists(o, allArtists, loaded, searchSettings)
		}
	}
	return ToErrorInterface(exitError)
}

type RepairSettings struct {
	dryRun bool
}

func NewRepairSettings() *RepairSettings {
	return &RepairSettings{}
}

func (rs *RepairSettings) WithDryRun(b bool) *RepairSettings {
	rs.dryRun = b
	return rs
}

func (rs *RepairSettings) ProcessArtists(o output.Bus, allArtists []*files.Artist,
	loaded bool, ss *SearchSettings) (e *ExitError) {
	e = NewExitUserError(repairCommandName)
	if loaded {
		if filteredArtists, filtered := ss.Filter(o, allArtists); filtered {
			e = rs.RepairArtists(o, filteredArtists)
		}
	}
	return
}

func (rs *RepairSettings) RepairArtists(o output.Bus, artists []*files.Artist) *ExitError {
	ReadMetadata(o, artists) // read all track metadata
	concernedArtists := PrepareConcernedArtists(artists)
	count := FindConflictedTracks(concernedArtists)
	if rs.dryRun {
		ReportRepairsNeeded(o, concernedArtists)
		return nil
	}
	if count == 0 {
		nothingToDo(o)
		return nil
	}
	return BackupAndFix(o, concernedArtists)
}

func FindConflictedTracks(concernedArtists []*ConcernedArtist) int {
	count := 0
	for _, cAr := range concernedArtists {
		for _, cAl := range cAr.Albums() {
			for _, cT := range cAl.Tracks() {
				state := cT.backing.ReconcileMetadata()
				if state.HasArtistNameConflict() {
					cT.AddConcern(ConflictConcern,
						"the artist name field does not match the name of the artist"+
							" directory")
				}
				if state.HasAlbumNameConflict() {
					cT.AddConcern(ConflictConcern,
						"the album name field does not match the name of the album"+
							" directory")
				}
				if state.HasGenreConflict() {
					cT.AddConcern(ConflictConcern,
						"the genre field does not match the other tracks in the album")
				}
				if state.HasMCDIConflict() {
					cT.AddConcern(ConflictConcern,
						"the music CD identifier field does not match the other tracks in"+
							" the album")
				}
				if state.HasNumberingConflict() {
					cT.AddConcern(ConflictConcern,
						"the track number field does not match the track's file name")
				}
				if state.HasTrackNameConflict() {
					cT.AddConcern(ConflictConcern,
						"the track name field does not match the track's file name")
				}
				if state.HasYearConflict() {
					cT.AddConcern(ConflictConcern,
						"the year field does not match the other tracks in the album")
				}
				if cT.IsConcerned() {
					count++
				}
			}
		}
	}
	return count
}

func ReportRepairsNeeded(o output.Bus, concernedArtists []*ConcernedArtist) {
	artistNames := make([]string, 0, len(concernedArtists))
	artistMap := map[string]*ConcernedArtist{}
	for _, cAr := range concernedArtists {
		name := cAr.name()
		artistNames = append(artistNames, name)
		artistMap[name] = cAr
	}
	slices.Sort(artistNames)
	headerPrinted := false
	for _, name := range artistNames {
		if cAr := artistMap[name]; cAr != nil {
			if cAr.IsConcerned() {
				if !headerPrinted {
					o.WriteConsole("The following concerns can be repaired:\n")
					headerPrinted = true
				}
				cAr.ToConsole(o)
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

// TODO: Better name: BackupAndRepairTracks
func BackupAndFix(o output.Bus, concernedArtists []*ConcernedArtist) *ExitError {
	var e *ExitError
	for _, cAr := range concernedArtists {
		if !cAr.IsConcerned() {
			continue
		}
		for _, cAl := range cAr.albums {
			if !cAl.IsConcerned() {
				continue
			}
			path, exists := EnsureBackupDirectoryExists(o, cAl)
			if !exists {
				e = NewExitSystemError(repairCommandName)
				continue
			}
			for _, cT := range cAl.tracks {
				if !cT.IsConcerned() {
					continue
				}
				t := cT.backing
				if !AttemptCopy(o, t, path) {
					e = NewExitSystemError(repairCommandName)
					continue
				}
				err := t.UpdateMetadata()
				if e2 := ProcessUpdateResult(o, t, err); e2 != nil {
					e = e2
				}
			}
		}
	}
	return e
}

// TODO: Better name: ProcessTrackRepairResults
func ProcessUpdateResult(o output.Bus, t *files.Track, updateErrs []error) *ExitError {
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
		return NewExitSystemError(repairCommandName)
	}
	o.WriteConsole("%q repaired.\n", t)
	MarkDirty(o)
	return nil
}

// TODO: better name: TryTrackBackup
func AttemptCopy(o output.Bus, t *files.Track, path string) (backedUp bool) {
	backupFile := filepath.Join(path, fmt.Sprintf("%d.mp3", t.Number()))
	switch {
	case PlainFileExists(backupFile):
		o.WriteCanonicalError("The backup file for track file %q, %q, already exists", t,
			backupFile)
		o.Log(output.Error, "file already exists", map[string]any{
			"command": repairCommandName,
			"file":    backupFile,
		})
	default:
		copyErr := CopyFile(t.Path(), backupFile)
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
				"source":      t.Path(),
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

// TODO: Better name: EnsureTrackBackupDirectoryExists
func EnsureBackupDirectoryExists(o output.Bus, cAl *ConcernedAlbum) (path string, exists bool) {
	path = cAl.backing.BackupDirectory()
	exists = true
	if !DirExists(path) {
		if fileErr := Mkdir(path); fileErr != nil {
			exists = false
			o.WriteCanonicalError("The directory %q cannot be created: %v", path, fileErr)
			o.WriteCanonicalError(
				"The track files in the directory %q will not be repaired",
				cAl.backing.Path())
			o.Log(output.Error, "cannot create directory", map[string]any{
				"command":   repairCommandName,
				"directory": path,
				"error":     fileErr,
			})
		}
	}
	return
}

func ProcessRepairFlags(o output.Bus, values map[string]*FlagValue) (*RepairSettings, bool) {
	rs := &RepairSettings{}
	flagsOk := true // optimistic
	var flagErr error
	if rs.dryRun, _, flagErr = GetBool(o, values, repairDryRun); flagErr != nil {
		flagsOk = false
	}
	return rs, flagsOk
}

func init() {
	RootCmd.AddCommand(RepairCmd)
	addDefaults(RepairFlags)
	o := getBus()
	c := getConfiguration()
	AddFlags(o, c, RepairCmd.Flags(), RepairFlags, SearchFlags)
}
