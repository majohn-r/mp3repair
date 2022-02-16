package subcommands

import (
	"flag"
	"fmt"

	"github.com/sirupsen/logrus"
)

type repair struct {
	target  *string
	dryRun  *bool
	commons *CommonCommandFlags
}

const (
	defaultRepairType string = "metadata"
	fsRepair          string = "files"
)

func (r *repair) name() string {
	return r.commons.name()
}
func newRepair(fSet *flag.FlagSet) CommandProcessor {
	return &repair{
		target:  fSet.String("target", defaultRepairType, fmt.Sprintf("either '%s' (make metadata agree with file system) or '%s' (make file system agree with metadata)", defaultRepairType, fsRepair)),
		dryRun:  fSet.Bool("dryRun", false, "if true, output what would have repaired, but make no repairs"),
		commons: newCommonCommandFlags(fSet),
	}
}

func (r *repair) Exec(args []string) {
	r.commons.processArgs(args)
	r.runSubcommand()
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
		logrus.Infof("search %s for files with extension %s for artists '%s' and albums '%s'", *r.commons.topDirectory, *r.commons.fileExtension, *r.commons.artistRegex, *r.commons.albumRegex)
	}
}

func (r *repair) validateTarget() {
	switch *r.target {
	case defaultRepairType, fsRepair:
	default:
		fmt.Printf("-target=%s is not valid\n", *r.target)
		s := defaultRepairType
		r.target = &s
	}
}
