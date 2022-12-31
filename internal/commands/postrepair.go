package commands

import (
	"flag"
	"mp3/internal"
	"mp3/internal/files"
	"os"
	"sort"

	"github.com/majohn-r/output"
)

func init() {
	addCommandData(postRepairCommandName, commandData{isDefault: false, init: newPostRepair})
}

const postRepairCommandName = "postRepair"

type postrepair struct {
	sf *files.SearchFlags
}

func newPostRepairCommand(o output.Bus, c *internal.Configuration, fSet *flag.FlagSet) (*postrepair, bool) {
	if sFlags, sFlagsOk := files.NewSearchFlags(o, c, fSet); sFlagsOk {
		return &postrepair{sf: sFlags}, true
	}
	return nil, false
}

func newPostRepair(o output.Bus, c *internal.Configuration, fSet *flag.FlagSet) (CommandProcessor, bool) {
	return newPostRepairCommand(o, c, fSet)
}

func (p *postrepair) Exec(o output.Bus, args []string) (ok bool) {
	if s, argsOk := p.sf.ProcessArgs(o, args); argsOk {
		ok = runCommand(o, s)
	}
	return
}

func logFields() map[string]any {
	return map[string]any{"command": postRepairCommandName}
}

func runCommand(o output.Bus, s *files.Search) (ok bool) {
	logStart(o, postRepairCommandName, logFields())
	if artists, artistsLoaded := s.LoadData(o); artistsLoaded {
		ok = true
		m := make(map[string]*files.Album)
		var dirs []string
		for _, aR := range artists {
			for _, aL := range aR.Albums() {
				dir := aL.BackupDirectory()
				if internal.DirExists(dir) {
					dirs = append(dirs, dir)
					m[dir] = aL
				}
			}
		}
		if len(dirs) == 0 {
			o.WriteCanonicalConsole("There are no backup directories to delete")
		} else {
			sort.Strings(dirs)
			for _, dir := range dirs {
				removeBackupDirectory(o, dir, m[dir])
			}
		}
	}
	return
}

func removeBackupDirectory(o output.Bus, dir string, aL *files.Album) {
	if err := os.RemoveAll(dir); err != nil {
		o.Log(output.Error, "cannot delete directory", map[string]any{
			"directory": dir,
			"error":     err,
		})
		o.WriteCanonicalError("The directory %q cannot be deleted: %v", dir, err)
	} else {
		o.WriteCanonicalConsole("The backup directory for artist %q album %q has been deleted\n", aL.RecordingArtistName(), aL.Name())
	}
}
