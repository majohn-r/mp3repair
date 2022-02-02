package subcommands

import (
	"flag"
	"fmt"
	"mp3/internal/files"
)

type check struct {
	fs                        *flag.FlagSet
	checkEmptyFolders         *bool
	checkGapsInTrackNumbering *bool
	checkIntegrity            *bool
	topDirectory              *string
	fileExtension             *string
}

func (c *check) Name() string {
	return c.fs.Name()
}

func NewCheckCommand() *check {
	defaultTopDir, _ := files.DefaultDirectory()
	fSet := flag.NewFlagSet("check", flag.ExitOnError)
	return &check{
		fs:                        fSet,
		checkEmptyFolders:         fSet.Bool("empty", true, "check for empty artist and album folders"),
		checkGapsInTrackNumbering: fSet.Bool("gaps", true, "check for gaps in track numbers"),
		checkIntegrity:            fSet.Bool("integrity", true, "check for disagreement between the file system and audio file metadata"),
		topDirectory:              fSet.String("topDir", defaultTopDir, "top directory in which to look for music files"),
		fileExtension:             fSet.String("ext", files.DefaultFileExtension, "extension for music files"),
	}
}

func (c *check) Exec(args []string) {
	err := c.fs.Parse(args)
	if err == nil {
		c.runSubcommand()
	} else {
		fmt.Printf("%v\n", err)
	}
}

func (c *check) runSubcommand() {
	fmt.Printf("%s: empty: %t, gaps: %t, integrity: %t, top directory: %s, extension: %s\n", c.Name(), *c.checkEmptyFolders, *c.checkGapsInTrackNumbering, *c.checkIntegrity, *c.topDirectory, *c.fileExtension)
}
