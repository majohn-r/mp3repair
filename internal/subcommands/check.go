package subcommands

import (
	"flag"
	"fmt"
	"os"

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
		checkEmptyFolders:         fSet.Bool("empty", false, "check for empty artist and album folders"),
		checkGapsInTrackNumbering: fSet.Bool("gaps", false, "check for gaps in track numbers"),
		checkIntegrity:            fSet.Bool("integrity", true, "check for disagreement between the file system and audio file metadata"),
		commons:                   newCommonCommandFlags(fSet),
	}
}

func (c *check) Exec(args []string) {
	if params := c.commons.processArgs(os.Stderr, args); params != nil {
		c.runSubcommand()
	}
}

func (c *check) runSubcommand() {
	if !*c.checkEmptyFolders && !*c.checkGapsInTrackNumbering && !*c.checkIntegrity {
		fmt.Fprintf(os.Stderr, "%s: nothing to do!", c.name())
		logrus.WithFields(logrus.Fields{"subcommand name": c.name()}).Error("nothing to do")
	} else {
		logrus.WithFields(logrus.Fields{
			"subcommandName":    c.name(),
			"checkEmptyFolders": *c.checkEmptyFolders,
			"checkTrackGaps":    *c.checkGapsInTrackNumbering,
			"checkIntegrity":    *c.checkIntegrity,
		}).Info("subcommand")
	}
}
