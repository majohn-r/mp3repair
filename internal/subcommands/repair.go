package subcommands

import (
	"flag"
	"io"
	"mp3/internal"
	"mp3/internal/files"
	"os"

	"github.com/sirupsen/logrus"
)

type repair struct {
	n      string
	dryRun *bool
	sf     *files.SearchFlags
}

func (r *repair) name() string {
	return r.n
}

func newRepair(fSet *flag.FlagSet) CommandProcessor {
	return newRepairSubCommand(fSet)
}

func newRepairSubCommand(fSet *flag.FlagSet) *repair {
	return &repair{
		n:      fSet.Name(),
		dryRun: fSet.Bool("dryRun", false, "if true, output what would have repaired, but make no repairs"),
		sf:     files.NewSearchFlags(fSet),
	}
}

func (r *repair) Exec(w io.Writer, args []string) {
	if s := r.sf.ProcessArgs(os.Stderr, args); s != nil {
		r.runSubcommand()
	}
}

const (
	logDryRunFlag string = "dryRun"
)

func (r *repair) logFields() logrus.Fields {
	return logrus.Fields{internal.LOG_COMMAND_NAME: r.name(), logDryRunFlag: *r.dryRun}
}

func (r *repair) runSubcommand() {
	logrus.WithFields(r.logFields()).Info(internal.LOG_EXECUTING_COMMAND)
	switch *r.dryRun {
	case true:
		logrus.Info("dry run only")
	case false:
		// TODO: replace with call to get files and perform the specified repairs
		// logrus.Infof("search %s for files with extension %s for artists '%s' and albums '%s'", *r.ff.topDirectory, *r.ff.fileExtension, *r.ff.artistRegex, *r.ff.albumRegex)
	}
}
