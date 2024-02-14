/*
Copyright © 2021 Marc Johnson (marc.johnson27591@gmail.com)
*/
package cmd

import (
	"fmt"
	"mp3/internal/files"
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
		Run: RepairRun,
	}
	RepairFlags = NewSectionFlags().WithSectionName("repair").WithFlags(
		map[string]*FlagDetails{
			"dryRun": NewFlagDetails().WithUsage(
				"output what would have been repaired, but make no repairs",
			).WithExpectedType(BoolType).WithDefaultValue(false),
		},
	)
)

func RepairRun(cmd *cobra.Command, _ []string) {
	status := ProgramError
	o := getBus()
	producer := cmd.Flags()
	values, eSlice := ReadFlags(producer, RepairFlags)
	searchSettings, searchFlagsOk := EvaluateSearchFlags(o, producer)
	if ProcessFlagErrors(o, eSlice) && searchFlagsOk {
		if rs, ok := ProcessRepairFlags(o, values); ok {
			details := map[string]any{repairDryRunFlag: rs.dryRun}
			for k, v := range searchSettings.Values() {
				details[k] = v
			}
			LogCommandStart(o, repairCommandName, details)
			allArtists, loaded := searchSettings.Load(o)
			status = rs.ProcessArtists(o, allArtists, loaded, searchSettings)
		}
	}
	Exit(status)
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
	loaded bool, ss *SearchSettings) int {
	status := UserError
	if loaded {
		if filteredArtists, filtered := ss.Filter(o, allArtists); filtered {
			status = rs.RepairArtists(o, filteredArtists)
		}
	}
	return status
}

func (rs *RepairSettings) RepairArtists(o output.Bus, artists []*files.Artist) int {
	status := Success
	ReadMetadata(o, artists) // read all track metadata
	concernedArtists := PrepareConcernedArtists(artists)
	count := FindConflictedTracks(concernedArtists)
	if rs.dryRun {
		ReportRepairsNeeded(o, concernedArtists)
	} else {
		if count == 0 {
			nothingToDo(o)
		} else {
			status = BackupAndFix(o, concernedArtists)
		}
	}
	return status
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
	artistNames := []string{}
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

func BackupAndFix(o output.Bus, concernedArtists []*ConcernedArtist) int {
	status := Success
	for _, cAr := range concernedArtists {
		if cAr.IsConcerned() {
			for _, cAl := range cAr.albums {
				if cAl.IsConcerned() {
					if path, exists := EnsureBackupDirectoryExists(o, cAl); exists {
						for _, cT := range cAl.tracks {
							if cT.IsConcerned() {
								t := cT.backing
								if AttemptCopy(o, t, path) {
									err := t.UpdateMetadata()
									if state := ProcessUpdateResult(o, t,
										err); state == SystemError {
										status = SystemError
									}
								} else {
									status = SystemError
								}
							}
						}
					} else {
						status = SystemError
					}
				}
			}
		}
	}
	return status
}

func ProcessUpdateResult(o output.Bus, t *files.Track, err []error) int {
	status := Success
	if len(err) == 0 {
		o.WriteConsole("%q repaired.\n", t)
		MarkDirty(o)
	} else {
		o.WriteCanonicalError("An error occurred repairing track %q", t)
		errorStrings := []string{}
		for _, e := range err {
			errorStrings = append(errorStrings, fmt.Sprintf("%q", e.Error()))
		}
		o.Log(output.Error, "cannot edit track", map[string]any{
			"command":   repairCommandName,
			"directory": t.Directory(),
			"fileName":  t.FileName(),
			"error":     fmt.Sprintf("[%s]", strings.Join(errorStrings, ", ")),
		})
		status = SystemError
	}
	return status
}

func AttemptCopy(o output.Bus, t *files.Track, path string) (backedUp bool) {
	backupFile := filepath.Join(path, fmt.Sprintf("%d.mp3", t.Number()))
	if PlainFileExists(backupFile) {
		o.WriteCanonicalError("The backup file for track file %q, %q, already exists", t,
			backupFile)
		o.Log(output.Error, "file already exists", map[string]any{
			"command": repairCommandName,
			"file":    backupFile,
		})
	} else {
		if err := CopyFile(t.Path(), backupFile); err == nil {
			o.WriteCanonicalConsole("The track file %q has been backed up to %q", t,
				backupFile)
			backedUp = true
		} else {
			o.WriteCanonicalError(
				"The track file %q could not be backed up due to error %v", t, err)
			o.Log(output.Error, "error copying file", map[string]any{
				"command":     repairCommandName,
				"source":      t.Path(),
				"destination": backupFile,
				"error":       err,
			})
		}
	}
	if !backedUp {
		o.WriteCanonicalError("The track file %q will not be repaired", t)
	}
	return
}

func EnsureBackupDirectoryExists(o output.Bus, cAl *ConcernedAlbum) (path string, exists bool) {
	path = cAl.backing.BackupDirectory()
	exists = true
	if !DirExists(path) {
		if err := Mkdir(path); err != nil {
			exists = false
			o.WriteCanonicalError("The directory %q cannot be created: %v", path, err)
			o.WriteCanonicalError(
				"The track files in the directory %q will not be repaired",
				cAl.backing.Path())
			o.Log(output.Error, "cannot create directory", map[string]any{
				"command":   repairCommandName,
				"directory": path,
				"error":     err,
			})
		}
	}
	return
}

func ProcessRepairFlags(o output.Bus, values map[string]*FlagValue) (*RepairSettings, bool) {
	rs := &RepairSettings{}
	ok := true // optimistic
	var err error
	if rs.dryRun, _, err = GetBool(o, values, repairDryRun); err != nil {
		ok = false
	}
	return rs, ok
}

func init() {
	RootCmd.AddCommand(RepairCmd)
	addDefaults(RepairFlags)
	o := getBus()
	c := getConfiguration()
	AddFlags(o, c, RepairCmd.Flags(), RepairFlags, SearchFlags)
}
