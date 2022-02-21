package subcommands

import (
	"errors"
	"flag"
	"fmt"
	"mp3/internal"
	"mp3/internal/files"
	"path/filepath"
	"sort"
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

func noSuchSubcommandError(commandName string, validNames []string) error {
	return fmt.Errorf("no subcommand named %q; valid subcommands include %v", commandName, validNames)
}

func internalErrorNoSubCommandInitializers() error {
	return errors.New("internal error: no subcommand initializers defined")
}

func internalErrorIncorrectNumberOfDefaultSubcommands(defaultInitializers int) error {
	return fmt.Errorf("internal error: only 1 subcommand should be designated as default; %d were found", defaultInitializers)
}

func ProcessCommand(args []string) (CommandProcessor, []string, error) {
	var initializers []subcommandInitializer
	initializers = append(initializers, subcommandInitializer{name: "ls", defaultSubCommand: true, initializer: newLs})
	initializers = append(initializers, subcommandInitializer{name: "check", defaultSubCommand: false, initializer: newCheck})
	initializers = append(initializers, subcommandInitializer{name: "repair", defaultSubCommand: false, initializer: newRepair})
	return selectSubCommand(initializers, args)
}

func selectSubCommand(initializers []subcommandInitializer, args []string) (cmd CommandProcessor, callingArgs []string, err error) {
	if len(initializers) == 0 {
		err = internalErrorNoSubCommandInitializers()
		return
	}
	var defaultInitializers int
	var defaultInitializerName string
	for _, initializer := range initializers {
		if initializer.defaultSubCommand {
			defaultInitializers++
			defaultInitializerName = initializer.name
		}
	}
	if defaultInitializers != 1 {
		err = internalErrorIncorrectNumberOfDefaultSubcommands(defaultInitializers)
		return
	}
	processorMap := make(map[string]CommandProcessor)
	for _, subcommandInitializer := range initializers {
		fSet := flag.NewFlagSet(subcommandInitializer.name, flag.ContinueOnError)
		processorMap[subcommandInitializer.name] = subcommandInitializer.initializer(fSet)
	}
	if len(args) < 2 {
		cmd = processorMap[defaultInitializerName]
		callingArgs = []string{defaultInitializerName}
		return
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

func defaultDirectory() string {
	return filepath.Join(internal.HomePath, "Music")
}

func newCommonCommandFlags(fSet *flag.FlagSet) *CommonCommandFlags {
	return &CommonCommandFlags{
		fs:            fSet,
		topDirectory:  fSet.String("topDir", defaultDirectory(), "top directory in which to look for music files"),
		fileExtension: fSet.String("ext", files.DefaultFileExtension, "extension for music files"),
		albumRegex:    fSet.String("albums", ".*", "regular expression of albums to select"),
		artistRegex:   fSet.String("artists", ".*", "regular expression of artists to select"),
	}
}

func (c *CommonCommandFlags) name() string {
	return c.fs.Name()
}

func (c *CommonCommandFlags) processArgs(args []string) (*files.DirectorySearchParams) {
	if err := c.fs.Parse(args); err != nil {
		fmt.Println(err)
		return nil
	}
	return files.NewDirectorySearchParams(*c.topDirectory, *c.fileExtension, *c.albumRegex, *c.artistRegex)
}
