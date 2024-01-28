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
	// repairCmd represents the repair command
	repairCmd = &cobra.Command{
		Use:                   repairCommandName + " [" + repairDryRunFlag + "] " + searchUsage,
		DisableFlagsInUseLine: true,
		Short:                 "Repairs problems found by running '" + CheckCommand + " " + CheckFilesFlag + "'",
		Long: "" +
			"Repairs problems found by running '" + CheckCommand + " " + CheckFilesFlag + "'\n" +
			"\n" +
			"This command rewrites the mp3 files that the " + CheckCommand + " command noted as having metadata\n" +
			"inconsistent with the file structure. Prior to rewriting an mp3 file, the " + repairCommandName + "\n" +
			"command creates a backup directory for the parent album and copies the original mp3\n" +
			"file into that backup directory. Use the " + postRepairCommandName + " command to automatically delete\n" +
			"the backup folders.",
		Run: RepairRun,
	}
	repairFlags = SectionFlags{
		SectionName: "repair",
		Flags: map[string]*FlagDetails{
			"dryRun": {
				Usage:        "output what would have been repaired, but make no repairs",
				ExpectedType: BoolType,
				DefaultValue: false,
			},
		},
	}
)

func RepairRun(cmd *cobra.Command, _ []string) {
	o := getBus()
	producer := cmd.Flags()
	values, eSlice := ReadFlags(producer, repairFlags)
	searchSettings, searchFlagsOk := EvaluateSearchFlags(o, producer)
	if ProcessFlagErrors(o, eSlice) && searchFlagsOk {
		if rs, ok := ProcessRepairFlags(o, values); ok {
			CommandStartLogger(o, repairCommandName, map[string]any{
				repairDryRunFlag:       rs.DryRun,
				SearchAlbumFilterFlag:  searchSettings.AlbumFilter,
				SearchArtistFilterFlag: searchSettings.ArtistFilter,
				SearchTrackFilterFlag:  searchSettings.TrackFilter,
				SearchTopDirFlag:       searchSettings.TopDirectory,
			})
			allArtists, loaded := searchSettings.Load(o)
			rs.ProcessArtists(o, allArtists, loaded, searchSettings)
		}
	}
}

type RepairSettings struct {
	DryRun bool
}

func (rs *RepairSettings) ProcessArtists(o output.Bus, allArtists []*files.Artist, loaded bool, ss *SearchSettings) {
	if loaded {
		if filteredArtists, filtered := ss.Filter(o, allArtists); filtered {
			rs.RepairArtists(o, filteredArtists)
		}
	}
}

func (rs *RepairSettings) RepairArtists(o output.Bus, artists []*files.Artist) {
	MetadataReader(o, artists) // read all track metadata
	checkedArtists := PrepareCheckedArtists(artists)
	count := FindConflictedTracks(checkedArtists)
	if rs.DryRun {
		ReportRepairsNeeded(o, checkedArtists)
	} else {
		if count == 0 {
			nothingToDo(o)
		} else {
			BackupAndFix(o, checkedArtists)
		}
	}
}

func FindConflictedTracks(checkedArtists []*CheckedArtist) int {
	count := 0
	for _, cAr := range checkedArtists {
		for _, cAl := range cAr.Albums() {
			for _, cT := range cAl.Tracks() {
				state := cT.backing.ReconcileMetadata()
				if state.HasArtistNameConflict() {
					cT.AddIssue(CheckConflictIssue, "the artist name field does not match the name of the artist directory")
				}
				if state.HasAlbumNameConflict() {
					cT.AddIssue(CheckConflictIssue, "the album name field does not match the name of the album directory")
				}
				if state.HasGenreConflict() {
					cT.AddIssue(CheckConflictIssue, "the genre field does not match the other tracks in the album")
				}
				if state.HasMCDIConflict() {
					cT.AddIssue(CheckConflictIssue, "the music CD identifier field does not match the other tracks in the album")
				}
				if state.HasNumberingConflict() {
					cT.AddIssue(CheckConflictIssue, "the track number field does not match the track's file name")
				}
				if state.HasTrackNameConflict() {
					cT.AddIssue(CheckConflictIssue, "the track name field does not match the track's file name")
				}
				if state.HasYearConflict() {
					cT.AddIssue(CheckConflictIssue, "the year field does not match the other tracks in the album")
				}
				if cT.HasIssues() {
					count++
				}
			}
		}
	}
	return count
}

func ReportRepairsNeeded(o output.Bus, checkedArtists []*CheckedArtist) {
	artistNames := []string{}
	artistMap := map[string]*CheckedArtist{}
	for _, cAr := range checkedArtists {
		name := cAr.name()
		artistNames = append(artistNames, name)
		artistMap[name] = cAr
	}
	slices.Sort(artistNames)
	headerPrinted := false
	for _, name := range artistNames {
		if cAr := artistMap[name]; cAr != nil {
			if cAr.HasIssues() {
				if !headerPrinted {
					o.WriteConsole("The following issues can be repaired:\n")
					headerPrinted = true
				}
				cAr.OutputIssues(o)
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

func BackupAndFix(o output.Bus, checkedArtists []*CheckedArtist) {
	for _, cAr := range checkedArtists {
		if cAr.HasIssues() {
			for _, cAl := range cAr.albums {
				if cAl.HasIssues() {
					if path, exists := EnsureBackupDirectoryExists(o, cAl); exists {
						for _, cT := range cAl.tracks {
							if cT.HasIssues() {
								t := cT.backing
								if AttemptCopy(o, t, path) {
									err := t.UpdateMetadata()
									ProcessUpdateResult(o, t, err)
								}
							}
						}
					}
				}
			}
		}
	}
}

func ProcessUpdateResult(o output.Bus, t *files.Track, err []error) {
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
	}
}

func AttemptCopy(o output.Bus, t *files.Track, path string) (backedUp bool) {
	backupFile := filepath.Join(path, fmt.Sprintf("%d.mp3", t.Number()))
	if PlainFileExists(backupFile) {
		o.WriteCanonicalError("The backup file for track file %q, %q, already exists", t, backupFile)
		o.Log(output.Error, "file already exists", map[string]any{
			"command": repairCommandName,
			"file":    backupFile,
		})
	} else {
		if err := CopyFile(t.Path(), backupFile); err == nil {
			o.WriteCanonicalConsole("The track file %q has been backed up to %q", t, backupFile)
			backedUp = true
		} else {
			o.WriteCanonicalError("The track file %q could not be backed up due to error %v", t, err)
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

func EnsureBackupDirectoryExists(o output.Bus, cAl *CheckedAlbum) (path string, exists bool) {
	path = cAl.backing.BackupDirectory()
	exists = true
	if !DirExists(path) {
		if err := MkDir(path); err != nil {
			exists = false
			o.WriteCanonicalError("The directory %q cannot be created: %v", path, err)
			o.WriteCanonicalError("The track files in the directory %q will not be repaired", cAl.backing.Path())
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
	if rs.DryRun, _, err = GetBool(o, values, repairDryRun); err != nil {
		ok = false
	}
	return rs, ok
}

func init() {
	rootCmd.AddCommand(repairCmd)
	addDefaults(repairFlags)
	o := getBus()
	c := getConfiguration()
	AddFlags(o, c, repairCmd.Flags(), repairFlags, true)
}
