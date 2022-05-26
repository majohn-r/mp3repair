package subcommands

import (
	"flag"
	"io"
	"mp3/internal"
	"mp3/internal/files"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type repair struct {
	n      string
	dryRun *bool
	sf     *files.SearchFlags
}

func (r *repair) name() string {
	return r.n
}

func newRepair(v *viper.Viper, fSet *flag.FlagSet) CommandProcessor {
	return newRepairSubCommand(v, fSet)
}

const (
	dryRunFlag    = "dryRun"
	defaultDryRun = false
)

func newRepairSubCommand(v *viper.Viper, fSet *flag.FlagSet) *repair {
	subViper := internal.SafeSubViper(v, "repair")
	return &repair{
		n: fSet.Name(),
		dryRun: fSet.Bool(dryRunFlag,
			internal.GetBoolDefault(subViper, dryRunFlag, defaultDryRun),
			"if true, output what would have repaired, but make no repairs"),
		sf: files.NewSearchFlags(v, fSet),
	}
}

func (r *repair) Exec(w io.Writer, args []string) {
	if s := r.sf.ProcessArgs(os.Stderr, args); s != nil {
		r.runSubcommand()
	}
}

func (r *repair) logFields() logrus.Fields {
	return logrus.Fields{internal.LOG_COMMAND_NAME: r.name(), dryRunFlag: *r.dryRun}
}

func (r *repair) runSubcommand() {
	logrus.WithFields(r.logFields()).Info(internal.LOG_EXECUTING_COMMAND)
	// TODO [#6] needs implementation! basically, run the integrity check (hmmm
	// ... maybe refactor that logic?) and fix the reported track-specific
	// problems found (in the case of dryRun = true, state what would be done
	// for each issue)
}
