package subcommands

import (
	"flag"
	"mp3/internal/files"
)

type CommandProcessor interface {
	name() string
	Exec([]string)
}

type subcommandInitializer struct {
	name              string
	defaultSubCommand bool
	initializer       func(*flag.FlagSet) CommandProcessor
}

func ProcessCommand(args []string) (cmd CommandProcessor, callingArgs []string) {
	var initializers []subcommandInitializer
	initializers = append(initializers, subcommandInitializer{name: "ls", defaultSubCommand: true, initializer: newLs})
	initializers = append(initializers, subcommandInitializer{name: "check", defaultSubCommand: false, initializer: newCheck})
	initializers = append(initializers, subcommandInitializer{name: "repair", defaultSubCommand: false, initializer: newRepair})
	processorMap := make(map[string]CommandProcessor)
	for _, subcommandInitializer := range initializers {
		fSet := flag.NewFlagSet(subcommandInitializer.name, flag.ExitOnError)
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

// processArgs parses command line arguments for a CommandProcessor; it does not
// return if there are errors in the input arguments
func (c *CommonCommandFlags) processArgs(args []string) {
	// ignore the error return, as all FlagSets are initialized with ExitOnError
	_ = c.fs.Parse(args)
}
