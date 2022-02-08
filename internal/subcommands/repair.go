package subcommands

import (
	"flag"
	"fmt"
	"mp3/internal/files"

	log "github.com/sirupsen/logrus"
)

type repair struct {
	fs            *flag.FlagSet
	target        *string
	albumRegex    *string
	artistRegex   *string
	dryRun        *bool
	topDirectory  *string
	fileExtension *string
}

const (
	defaultRepairType string = "metadata"
	fsRepair          string = "files"
)

func (r *repair) Name() string {
	return r.fs.Name()
}
func NewRepairCommandProcessor() *repair {
	fSet := flag.NewFlagSet("repair", flag.ExitOnError)
	return &repair{
		fs:            fSet,
		target:        fSet.String("target", defaultRepairType, fmt.Sprintf("either '%s' (make metadata agree with file system) or '%s' (make file system agree with metadata)", defaultRepairType, fsRepair)),
		albumRegex:    fSet.String("albums", ".*", "regular expression of albums to repair"),
		artistRegex:   fSet.String("artists", ".*", "regular epxression of artists to repair"),
		dryRun:        fSet.Bool("dryRun", false, "if true, output what would have repaired, but make no repairs"),
		topDirectory:  fSet.String("topDir", files.DefaultDirectory(), "top directory in which to look for music files"),
		fileExtension: fSet.String("ext", files.DefaultFileExtension, "extension for music files"),
	}
}

func (r *repair) Exec(args []string) {
	err := r.fs.Parse(args)
	switch err {
	case nil:
		r.runSubcommand()
	default:
		log.Fatal(err)
	}
}

func (r *repair) runSubcommand() {
	r.validateTarget()
	log.Infof("%s %s", r.Name(), *r.target)
	switch *r.dryRun {
	case true:
		log.Info("dry run only")
	case false:
		log.Infof("search %s for files with extension %s for artists '%s' and albums '%s'", *r.topDirectory, *r.fileExtension, *r.artistRegex, *r.albumRegex)
	}
}

func (r *repair) validateTarget() {
	switch *r.target {
	case defaultRepairType, fsRepair:
	default:
		log.Warnf("-target=%s is not valid\n", *r.target)
		s := defaultRepairType
		r.target = &s
	}
}
