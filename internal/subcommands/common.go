package subcommands

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"mp3/internal"
	"sort"
)

type CommandProcessor interface {
	name() string
	Exec(io.Writer, []string)
}

type subcommandInitializer struct {
	name              string
	defaultSubCommand bool
	initializer       func(*flag.FlagSet) CommandProcessor
}

func noSuchSubcommandError(commandName string, validNames []string) error {
	return fmt.Errorf(internal.LOG_NO_SUCH_COMMAND, commandName, validNames)
}

func internalErrorNoSubCommandInitializers() error {
	return errors.New(internal.LOG_NO_DEFAULT_COMMAND_DEFINED)
}

func internalErrorIncorrectNumberOfDefaultSubcommands(defaultInitializers int) error {
	return fmt.Errorf(internal.LOG_TOO_MANY_DEFAULT_COMMANDS_DEFINED, defaultInitializers)
}

func ProcessCommand(args []string) (CommandProcessor, []string, error) {
	var initializers []subcommandInitializer
	lsSubCommand := subcommandInitializer{name: "ls", defaultSubCommand: true, initializer: newLs}
	checkSubCommand := subcommandInitializer{name: "check", defaultSubCommand: false, initializer: newCheck}
	repairSubCommand := subcommandInitializer{name: "repair", defaultSubCommand: false, initializer: newRepair}
	initializers = append(initializers, lsSubCommand, checkSubCommand, repairSubCommand)
	return selectSubCommand(initializers, args)
}

func selectSubCommand(i []subcommandInitializer, args []string) (cmd CommandProcessor, callingArgs []string, err error) {
	if len(i) == 0 {
		err = internalErrorNoSubCommandInitializers()
		return
	}
	var defaultInitializers int
	var defaultInitializerName string
	for _, initializer := range i {
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
	for _, subcommandInitializer := range i {
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
		for _, initializer := range i {
			subCommandNames = append(subCommandNames, initializer.name)
		}
		sort.Strings(subCommandNames)
		err = noSuchSubcommandError(commandName, subCommandNames)
		return
	}
	callingArgs = args[2:]
	return
}
