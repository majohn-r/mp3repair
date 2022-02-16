package subcommands

import (
	"flag"

	"github.com/sirupsen/logrus"
)

type check struct {
	checkEmptyFolders         *bool
	checkGapsInTrackNumbering *bool
	checkIntegrity            *bool
	commons                   *CommonCommandFlags
}

func (c *check) name() string {
	return c.commons.name()
}

func newCheck(fSet *flag.FlagSet) CommandProcessor {
	return &check{
		checkEmptyFolders:         fSet.Bool("empty", true, "check for empty artist and album folders"),
		checkGapsInTrackNumbering: fSet.Bool("gaps", true, "check for gaps in track numbers"),
		checkIntegrity:            fSet.Bool("integrity", true, "check for disagreement between the file system and audio file metadata"),
		commons:                   newCommonCommandFlags(fSet),
	}
}

func (c *check) Exec(args []string) {
	c.commons.processArgs(args)
	c.runSubcommand()
}

func (c *check) runSubcommand() {
	logrus.WithFields(logrus.Fields{
		"subcommandName":    c.name(),
		"checkEmptyFolders": *c.checkEmptyFolders,
		"checkTrackGaps":    *c.checkGapsInTrackNumbering,
		"checkIntegrity":    *c.checkIntegrity,
	}).Info("subcommand")
}
