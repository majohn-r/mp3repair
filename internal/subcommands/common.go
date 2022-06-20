package subcommands

import (
	"flag"
	"fmt"
	"io"
	"mp3/internal"
	"os"
	"sort"
	"strings"

	"github.com/sirupsen/logrus"
)

const (
	fkCommandName = "command"
	fkCount       = "count"
)

type CommandProcessor interface {
	name() string
	Exec(io.Writer, io.Writer, []string) bool
}

type subcommandInitializer struct {
	name              string
	defaultSubCommand bool
	initializer       func(*internal.Configuration, *flag.FlagSet) CommandProcessor
}

// ProcessCommand selects which command to be run and returns the relevant
// CommandProcessor, command line arguments and ok status
func ProcessCommand(w io.Writer, appDataPath string, args []string) (cmd CommandProcessor, cmdArgs []string, ok bool) {
	var c *internal.Configuration
	if c, ok = internal.ReadConfigurationFile(os.Stderr, internal.CreateAppSpecificPath(appDataPath)); !ok {
		return nil, nil, false
	}
	var initializers []subcommandInitializer
	lsSubCommand := subcommandInitializer{name: "ls", defaultSubCommand: true, initializer: newLs}
	checkSubCommand := subcommandInitializer{name: "check", initializer: newCheck}
	repairSubCommand := subcommandInitializer{name: "repair", initializer: newRepair}
	postRepairSubCommand := subcommandInitializer{name: "postRepair", initializer: newPostRepair}
	initializers = append(initializers, lsSubCommand, checkSubCommand, repairSubCommand, postRepairSubCommand)
	cmd, cmdArgs, ok = selectSubCommand(w, c, initializers, args)
	return
}

func selectSubCommand(w io.Writer, c *internal.Configuration, i []subcommandInitializer, args []string) (cmd CommandProcessor, callingArgs []string, ok bool) {
	if len(i) == 0 {
		logrus.WithFields(logrus.Fields{
			fkCount: 0,
		}).Error(internal.LE_COMMAND_COUNT)
		fmt.Fprint(w, internal.USER_NO_COMMANDS_DEFINED)
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
		logrus.WithFields(logrus.Fields{
			fkCount: defaultInitializers,
		}).Error(internal.LE_DEFAULT_COMMAND_COUNT)
		fmt.Fprintf(w, internal.USER_INCORRECT_NUMBER_OF_DEFAULT_COMMANDS_DEFINED, defaultInitializers)
		return
	}
	processorMap := make(map[string]CommandProcessor)
	for _, subcommandInitializer := range i {
		fSet := flag.NewFlagSet(subcommandInitializer.name, flag.ContinueOnError)
		processorMap[subcommandInitializer.name] = subcommandInitializer.initializer(c, fSet)
	}
	if len(args) < 2 {
		cmd = processorMap[defaultInitializerName]
		callingArgs = []string{defaultInitializerName}
		ok = true
		return
	}
	commandName := args[1]
	if strings.HasPrefix(commandName, "-") {
		cmd = processorMap[defaultInitializerName]
		callingArgs = args[1:]
		ok = true
		return
	}
	cmd, found := processorMap[commandName]
	if !found {
		cmd = nil
		callingArgs = nil
		logrus.WithFields(logrus.Fields{
			fkCommandName: commandName,
		}).Warn(internal.LW_UNRECOGNIZED_COMMAND)
		var subCommandNames []string
		for _, initializer := range i {
			subCommandNames = append(subCommandNames, initializer.name)
		}
		sort.Strings(subCommandNames)
		fmt.Fprintf(w, internal.USER_NO_SUCH_COMMAND, commandName, subCommandNames)
		return
	}
	callingArgs = args[2:]
	ok = true
	return
}
