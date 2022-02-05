package subcommands

import (
	"flag"
	"fmt"
	"log"
	"mp3/internal/files"
	"strings"
)

type repair struct {
	fs            *flag.FlagSet
	target        *string
	albumRegex    *string
	artistRegex   *string
	topDirectory  *string
	fileExtension *string
}

const (
	defaultRepairType string = "metadata"
	fsRepair string = "files"
)

func (r *repair) Name() string {
	return r.fs.Name()
}
func NewRepairCommand() *repair {
	defaultTopDir, _ := files.DefaultDirectory()
	fSet := flag.NewFlagSet("repair", flag.ExitOnError)
	return &repair{
		fs:            fSet,
		target:        fSet.String("target", defaultRepairType, fmt.Sprintf("either '%s' (make metadata agree with file system) or '%s' (make file system agree with metadata)", defaultRepairType, fsRepair)),
		albumRegex:    fSet.String("albums", ".*", "regular expression of albums to repair"),
		artistRegex:   fSet.String("artists", ".*", "regular epxression of artists to repair"),
		topDirectory:  fSet.String("topDir", defaultTopDir, "top directory in which to look for music files"),
		fileExtension: fSet.String("ext", files.DefaultFileExtension, "extension for music files"),
	}
}

func (r *repair) Exec(args []string) {
	err := r.fs.Parse(args)
	if err == nil {
		r.runSubcommand()
	} else {
		fmt.Printf("%v\n", err)
	}
}

func (r *repair) runSubcommand() {
	r.validateTarget()
	var output []string
	output = append(output, fmt.Sprintf(" target: %s", *r.target))
	output = append(output, fmt.Sprintf(" albums: %s", *r.albumRegex))
	output = append(output, fmt.Sprintf(" artists: %s", *r.artistRegex))
	log.Printf("%s:%s", r.Name(), strings.Join(output, ";"))
	log.Printf("search %s for files with extension %s", *r.topDirectory, *r.fileExtension)
}

func (r *repair) validateTarget() {
	switch *r.target{
	case defaultRepairType:
	case fsRepair:
	default:
		log.Printf("-target=%s is not valid\n", *r.target)
		s := defaultRepairType
		r.target = &s
	}
}