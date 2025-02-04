package cmd

import (
	"fmt"
	cmdtoolkit "github.com/majohn-r/cmd-toolkit"
	"mp3repair/internal/files"
	"sort"

	"github.com/majohn-r/output"
	"github.com/spf13/cobra"
)

const postRepairCommandName = "postRepair"

var (
	postRepairCmd = &cobra.Command{
		Use:                   postRepairCommandName + " " + searchUsage,
		DisableFlagsInUseLine: true,
		Short: "Deletes the backup directories, and their contents, created" +
			" by the " + repairCommandName + " command",
		Long: fmt.Sprintf(
			"%q deletes the backup directories (and their contents) created by the %q command",
			postRepairCommandName, repairCommandName),
		RunE: postRepairRun,
	}
)

func postRepairRun(cmd *cobra.Command, _ []string) error {
	exitError := cmdtoolkit.NewExitProgrammingError(postRepairCommandName)
	o := getBus()
	producer := cmd.Flags()
	ss, searchFlagsOk := evaluateSearchFlags(o, producer)
	if searchFlagsOk {
		// do some work here!
		exitError = postRepairWork(o, ss, ss.load(o))
	}
	return cmdtoolkit.ToErrorInterface(exitError)
}

func postRepairWork(o output.Bus, ss *searchSettings, allArtists []*files.Artist) (e *cmdtoolkit.ExitError) {
	e = cmdtoolkit.NewExitUserError(postRepairCommandName)
	if len(allArtists) != 0 {
		if filteredArtists := ss.filter(o, allArtists); len(filteredArtists) != 0 {
			e = nil
			dirCount := 0
			for _, artist := range filteredArtists {
				dirCount += len(artist.Albums())
			}
			dirs := make([]string, 0, dirCount)
			for _, artist := range filteredArtists {
				for _, album := range artist.Albums() {
					dir := album.BackupDirectory()
					if dirExists(dir) {
						dirs = append(dirs, dir)
					}
				}
			}
			o.ConsolePrintf("Backup directories to delete: %d.\n", len(dirs))
			if len(dirs) > 0 {
				sort.Strings(dirs)
				dirsDeleted := 0
				for _, dir := range dirs {
					switch removeTrackBackupDirectory(o, dir) {
					case true:
						dirsDeleted++
					default:
						e = cmdtoolkit.NewExitSystemError(postRepairCommandName)
					}
				}
				o.ConsolePrintf("Backup directories deleted: %d.\n", dirsDeleted)
			}
		}
	}
	return
}

func removeTrackBackupDirectory(o output.Bus, dir string) bool {
	if fileErr := removeAll(dir); fileErr != nil {
		o.Log(output.Error, "cannot delete directory", map[string]any{
			"directory": dir,
			"error":     fileErr,
		})
		return false
	}
	o.Log(output.Info, "directory deleted", map[string]any{"directory": dir})
	return true
}

func init() {
	rootCmd.AddCommand(postRepairCmd)
	cmdtoolkit.AddFlags(getBus(), getConfiguration(), postRepairCmd.Flags(), searchFlags)
}
