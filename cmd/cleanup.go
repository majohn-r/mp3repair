package cmd

import (
	"fmt"
	"mp3repair/internal/files"
	"sort"

	cmdtoolkit "github.com/majohn-r/cmd-toolkit"

	"github.com/majohn-r/output"
	"github.com/spf13/cobra"
)

const cleanupCommandName = "cleanup"

var (
	cleanupCmd = &cobra.Command{
		Use:                   cleanupCommandName + " " + searchUsage,
		DisableFlagsInUseLine: true,
		Short: "Deletes the backup directories, and their contents, created" +
			" by the " + rewriteCommandName + " command",
		Long: fmt.Sprintf(
			"%q deletes the backup directories (and their contents) created by the %q command",
			cleanupCommandName, rewriteCommandName),
		RunE: cleanupRun,
	}
)

func cleanupRun(cmd *cobra.Command, _ []string) error {
	exitError := cmdtoolkit.NewExitProgrammingError(cleanupCommandName)
	o := getBus()
	producer := cmd.Flags()
	ss, searchFlagsOk := evaluateSearchFlags(o, producer)
	if searchFlagsOk {
		// do some work here!
		exitError = cleanupWork(o, ss, ss.load(o))
	}
	return cmdtoolkit.ToErrorInterface(exitError)
}

func cleanupWork(o output.Bus, ss *searchSettings, allArtists []*files.Artist) (e *cmdtoolkit.ExitError) {
	e = cmdtoolkit.NewExitUserError(cleanupCommandName)
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
						e = cmdtoolkit.NewExitSystemError(cleanupCommandName)
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
	rootCmd.AddCommand(cleanupCmd)
	cmdtoolkit.AddFlags(getBus(), getConfiguration(), cleanupCmd.Flags(), searchFlags)
}
