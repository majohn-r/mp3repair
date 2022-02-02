package subcommands

import (
	"flag"
	"fmt"
	"mp3/internal/files"
)

type repair struct {
	fs            *flag.FlagSet
	target        *string
	topDirectory  *string
	fileExtension *string
}

func (r *repair) Name() string {
	return r.fs.Name()
}
func NewRepairCommand() *repair {
	defaultTopDir, _ := files.DefaultDirectory()
	fSet := flag.NewFlagSet("repair", flag.ExitOnError)
	return &repair{
		fs:            fSet,
		target:        fSet.String("target", "metadata", "either 'metadata' (make metadata agree with file system) or 'files' (make file system agree with metadata)"),
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
	fmt.Printf("%s: target: %s, top directory: %s, extension: %s\n", r.Name(), *r.target, *r.topDirectory, *r.fileExtension)
}
