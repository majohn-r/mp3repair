package subcommands

import (
	"flag"
	"fmt"
	"log"
	"mp3/internal/files"
)

type check struct {
	fs                        *flag.FlagSet
	checkEmptyFolders         *bool
	checkGapsInTrackNumbering *bool
	checkIntegrity            *bool
	albumRegex                *string
	artistRegex               *string
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
		albumRegex:                fSet.String("albums", ".*", "regular expression of albums to repair"),
		artistRegex:               fSet.String("artists", ".*", "regular epxression of artists to repair"),
	}
}

func (c *check) Exec(args []string) {
	err := c.fs.Parse(args)
	switch err {
	case nil:
		c.runSubcommand()
	default:
		fmt.Printf("%v\n", err)
	}
}

func (c *check) runSubcommand() {
	log.Printf("%s: empty: %t, gaps: %t, integrity: %t\n", c.Name(), *c.checkEmptyFolders, *c.checkGapsInTrackNumbering, *c.checkIntegrity)
	log.Printf("search %s for files with extension %s for artists '%s' and albums '%s'", *c.topDirectory, *c.fileExtension, *c.artistRegex, *c.albumRegex)
}
