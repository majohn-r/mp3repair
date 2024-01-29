/*
Copyright © 2021 Marc Johnson (marc.johnson27591@gmail.com)
*/
package cmd

import (
	"mp3/internal/files"
	"sort"

	"github.com/majohn-r/output"
	"github.com/spf13/cobra"
)

const postRepairCommandName = "postRepair"

var (
	// postRepairCmd represents the postRepair command
	postRepairCmd = &cobra.Command{
		Use:                   postRepairCommandName + " " + searchUsage,
		DisableFlagsInUseLine: true,
		Short:                 "Delete the backup directories, and their contents, created by the " + repairCommandName + " command",
		Run:                   PostRepairRun,
	}
)

func PostRepairRun(cmd *cobra.Command, _ []string) {
	o := getBus()
	producer := cmd.Flags()
	ss, searchFlagsOk := EvaluateSearchFlags(o, producer)
	if searchFlagsOk {
		// do some work here!
		LogCommandStart(o, postRepairCommandName, map[string]any{
			SearchAlbumFilterFlag:  ss.AlbumFilter,
			SearchArtistFilterFlag: ss.ArtistFilter,
			SearchTrackFilterFlag:  ss.TrackFilter,
			SearchTopDirFlag:       ss.TopDirectory,
		})
		allArtists, loaded := ss.Load(o)
		PostRepairWork(o, ss, allArtists, loaded)
	}
}

func PostRepairWork(o output.Bus, ss *SearchSettings, allArtists []*files.Artist, loaded bool) {
	if loaded {
		if filteredArtists, filtered := ss.Filter(o, allArtists); filtered {
			dirs := []string{}
			for _, artist := range filteredArtists {
				for _, album := range artist.Albums() {
					dir := album.BackupDirectory()
					if DirExists(dir) {
						dirs = append(dirs, dir)
					}
				}
			}
			o.WriteCanonicalConsole("Backup directories to delete: %d", len(dirs))
			if len(dirs) > 0 {
				sort.Strings(dirs)
				dirsDeleted := 0
				for _, dir := range dirs {
					if RemoveBackupDirectory(o, dir) {
						dirsDeleted++
					}
				}
				o.WriteCanonicalConsole("Backup directories deleted: %d", dirsDeleted)
			}
		}
	}
}

func RemoveBackupDirectory(o output.Bus, dir string) bool {
	if err := RemoveAll(dir); err != nil {
		o.Log(output.Error, "cannot delete directory", map[string]any{
			"directory": dir,
			"error":     err,
		})
		return false
	}
	o.Log(output.Info, "directory deleted", map[string]any{"directory": dir})
	return true
}

func init() {
	rootCmd.AddCommand(postRepairCmd)
	bus := getBus()
	c := getConfiguration()
	AddFlags(bus, c, postRepairCmd.Flags(), SearchFlags, false)
}
