package subcommands

import (
	"flag"
	"fmt"
	"mp3/internal/files"
	"sort"
)

type CommandProcessor interface {
	name() string
	Exec([]string) error
}

type subcommandInitializer struct {
	name              string
	defaultSubCommand bool
	initializer       func(*flag.FlagSet) CommandProcessor
}

func noSuchSubcommandError(commandName string, validNames []string) error {
	return fmt.Errorf("no subcommand named %q; valid subcommands include %v", commandName, validNames)
}

func ProcessCommand(args []string) (cmd CommandProcessor, callingArgs []string, err error) {
	var initializers []subcommandInitializer
	initializers = append(initializers, subcommandInitializer{name: "ls", defaultSubCommand: true, initializer: newLs})
	initializers = append(initializers, subcommandInitializer{name: "check", defaultSubCommand: false, initializer: newCheck})
	initializers = append(initializers, subcommandInitializer{name: "repair", defaultSubCommand: false, initializer: newRepair})
	processorMap := make(map[string]CommandProcessor)
	for _, subcommandInitializer := range initializers {
		fSet := flag.NewFlagSet(subcommandInitializer.name, flag.ContinueOnError)
		processorMap[subcommandInitializer.name] = subcommandInitializer.initializer(fSet)
	}
	if len(args) < 2 {
		for _, initializer := range initializers {
			if initializer.defaultSubCommand {
				cmd = processorMap[initializer.name]
				callingArgs = []string{initializer.name}
				return
			}
		}
		panic("no default subcommand defined!")
	}
	commandName := args[1]
	cmd, found := processorMap[commandName]
	if !found {
		cmd = nil
		callingArgs = nil
		var subCommandNames []string
		for _, initializer := range initializers {
			subCommandNames = append(subCommandNames, initializer.name)
		}
		sort.Strings(subCommandNames)
		err = noSuchSubcommandError(commandName, subCommandNames)
		return
	}
	callingArgs = args[2:]
	return
}

type CommonCommandFlags struct {
	fs            *flag.FlagSet
	topDirectory  *string
	fileExtension *string
	albumRegex    *string
	artistRegex   *string
}

func newCommonCommandFlags(fSet *flag.FlagSet) *CommonCommandFlags {
	return &CommonCommandFlags{
		fs:            fSet,
		topDirectory:  fSet.String("topDir", files.DefaultDirectory(), "top directory in which to look for music files"),
		fileExtension: fSet.String("ext", files.DefaultFileExtension, "extension for music files"),
		albumRegex:    fSet.String("albums", ".*", "regular expression of albums to select"),
		artistRegex:   fSet.String("artists", ".*", "regular expression of artists to select"),
	}
}

func (c *CommonCommandFlags) name() string {
	return c.fs.Name()
}

func (c *CommonCommandFlags) processArgs(args []string) (*files.DirectorySearchParams, error) {
	// ignore the error return, as all FlagSets are initialized with ExitOnError
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	params := files.NewDirectorySearchParams(*c.topDirectory, *c.fileExtension, *c.albumRegex, *c.artistRegex)
	return params, nil
}
