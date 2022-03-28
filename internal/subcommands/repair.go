package subcommands

import (
	"flag"
	"fmt"
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

func (r *repair) Exec(args []string) {
	if params := r.sf.ProcessArgs(os.Stderr, args); params != nil {
		r.runSubcommand()
	}
}

func (r *repair) runSubcommand() {
	r.validateTarget()
	logrus.WithFields(logrus.Fields{
		"subcommandName": r.name(),
		"target":         *r.target,
	}).Info("subcommand")
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
		fmt.Fprintf(os.Stderr, "-target=%s is not valid\n", *r.target)
		logrus.WithFields(logrus.Fields{"subcommand": r.name(), "setting": "-target", "value": *r.target}).Warn("unexpected setting")
		s := defaultRepairType
		r.target = &s
	}
}
