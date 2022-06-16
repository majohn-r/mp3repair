package subcommands

import (
	"flag"
	"fmt"
	"io"
	"mp3/internal"
	"mp3/internal/files"
	"os"
	"sort"

	"github.com/sirupsen/logrus"
)

type postrepair struct {
	n  string
	sf *files.SearchFlags
}

func (p *postrepair) name() string {
	return p.n
}

func newPostRepairSubCommand(c *internal.Configuration, fSet *flag.FlagSet) *postrepair {
	return &postrepair{n: fSet.Name(), sf: files.NewSearchFlags(c, fSet)}
}

func newPostRepair(c *internal.Configuration, fSet *flag.FlagSet) CommandProcessor {
	return newPostRepairSubCommand(c, fSet)
}

func (p *postrepair) Exec(w io.Writer, args []string) {
	if s := p.sf.ProcessArgs(os.Stderr, args); s != nil {
		p.runSubcommand(w, s)
	}
}

func (p *postrepair) logFields() logrus.Fields {
	return logrus.Fields{internal.FK_COMMAND_NAME: p.name()}
}

func (p *postrepair) runSubcommand(w io.Writer, s *files.Search) {
	logrus.WithFields(p.logFields()).Info(internal.LI_EXECUTING_COMMAND)
	artists := s.LoadData()
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
		fmt.Fprintln(w, "There are no backup directories to delete")
	} else {
		sort.Strings(backupDirectories)
		for _, backupDirectory := range backupDirectories {
			removeBackupDirectory(w, backupDirectory, backups[backupDirectory])
		}
	}
}

func removeBackupDirectory(w io.Writer, d string, a *files.Album) {
	if err := os.RemoveAll(d); err != nil {
		logrus.WithFields(logrus.Fields{
			internal.FK_DIRECTORY: d,
			internal.FK_ERROR:     err,
		}).Warn(internal.LW_CANNOT_DELETE_DIRECTORY)
		fmt.Fprintf(w, internal.USER_CANNOT_DELETE_DIRECTORY, d, err)
	} else {
		fmt.Fprintf(w, "The backup directory for artist %q album %q has been deleted\n",
			a.RecordingArtistName(), a.Name())
	}
}
