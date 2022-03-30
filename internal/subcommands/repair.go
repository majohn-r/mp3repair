package subcommands

import (
	"flag"
	"fmt"
	"io"
	"mp3/internal"
	"mp3/internal/files"
	"os"

	"github.com/sirupsen/logrus"
)

type repair struct {
	n      string
	target *string
	dryRun *bool
	sf     *files.SearchFlags
}

const (
	defaultRepairType string = "metadata"
	fsRepair          string = "files"
)

func (r *repair) name() string {
	return r.n
}
func newRepair(fSet *flag.FlagSet) CommandProcessor {
	return &repair{
		n:      fSet.Name(),
		target: fSet.String("target", defaultRepairType, fmt.Sprintf("either '%s' (make metadata agree with file system) or '%s' (make file system agree with metadata)", defaultRepairType, fsRepair)),
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
	logTargetFlag string = "target"
)

func (r *repair) logFields() logrus.Fields {
	return logrus.Fields{internal.LOG_COMMAND_NAME: r.name(), logDryRunFlag: *r.dryRun, logTargetFlag: *r.target}
}

func (r *repair) runSubcommand() {
	r.validateTarget()
	logrus.WithFields(r.logFields()).Info(internal.LOG_EXECUTING_COMMAND)
	switch *r.dryRun {
	case true:
		logrus.Info("dry run only")
	case false:
		// TODO: replace with call to get files and perform the specified repairs
		// logrus.Infof("search %s for files with extension %s for artists '%s' and albums '%s'", *r.ff.topDirectory, *r.ff.fileExtension, *r.ff.artistRegex, *r.ff.albumRegex)
	}
}

func (r *repair) validateTarget() {
	switch *r.target {
	case defaultRepairType, fsRepair:
	default:
		fmt.Fprintf(os.Stderr, internal.USER_UNRECOGNIZED_VALUE, "-target", *r.target)
		logrus.WithFields(logrus.Fields{
			internal.LOG_COMMAND_NAME: r.name(),
			internal.LOG_FLAG:         "-target",
			internal.LOG_VALUE:        *r.target,
		}).Warn(internal.LOG_INVALID_FLAG_SETTING)
		s := defaultRepairType
		r.target = &s
	}
}
