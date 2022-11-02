package commands

import (
	"flag"
	"mp3/internal"
	"mp3/internal/files"
	"mp3/internal/output"
	"os"
	"sort"
)

func init() {
	addCommandData(postRepairCommandName, commandData{isDefault: false, initFunction: newPostRepair})
}

const postRepairCommandName = "postRepair"

type postrepair struct {
	sf *files.SearchFlags
}

func newPostRepairCommand(o output.Bus, c *internal.Configuration, fSet *flag.FlagSet) (*postrepair, bool) {
	sFlags, sFlagsOk := files.NewSearchFlags(o, c, fSet)
	if sFlagsOk {
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
	return map[string]any{fieldKeyCommandName: postRepairCommandName}
}

func runCommand(o output.Bus, s *files.Search) (ok bool) {
	o.Log(output.Info, internal.LogInfoExecutingCommand, logFields())
	artists, ok := s.LoadData(o)
	if ok {
		backups := make(map[string]*files.Album)
		var backupDirectories []string
		for _, artist := range artists {
			for _, album := range artist.Albums() {
				backupDirectory := album.BackupDirectory()
				if internal.DirExists(backupDirectory) {
					backupDirectories = append(backupDirectories, backupDirectory)
					backups[backupDirectory] = album
				}
			}
		}
		if len(backupDirectories) == 0 {
			o.WriteCanonicalConsole("There are no backup directories to delete")
		} else {
			sort.Strings(backupDirectories)
			for _, backupDirectory := range backupDirectories {
				removeBackupDirectory(o, backupDirectory, backups[backupDirectory])
			}
		}
	}
	return
}

func removeBackupDirectory(o output.Bus, d string, a *files.Album) {
	if err := os.RemoveAll(d); err != nil {
		o.Log(output.Error, internal.LogErrorCannotDeleteDirectory, map[string]any{
			internal.FieldKeyDirectory: d,
			internal.FieldKeyError:     err,
		})
		o.WriteCanonicalError(internal.UserCannotDeleteDirectory, d, err)
	} else {
		o.WriteCanonicalConsole("The backup directory for artist %q album %q has been deleted\n", a.RecordingArtistName(), a.Name())
	}
}
