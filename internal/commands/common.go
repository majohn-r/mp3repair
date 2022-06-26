package commands

import (
	"flag"
	"fmt"
	"io"
	"mp3/internal"
	"sort"
	"strings"

	"github.com/sirupsen/logrus"
)

const (
	checkCommand      = "check"
	fkCommandName     = "command"
	fkCount           = "count"
	lsCommand         = "ls"
	postRepairCommand = "postRepair"
	repairCommand     = "repair"
)

type CommandProcessor interface {
	name() string
	Exec(internal.OutputBus, []string) bool
}

type subcommandInitializer struct {
	name              string
	defaultSubCommand bool
	initializer       func(*internal.Configuration, *flag.FlagSet) CommandProcessor
}

// ProcessCommand selects which command to be run and returns the relevant
// CommandProcessor, command line arguments and ok status
func ProcessCommand(o internal.OutputBus, args []string) (cmd CommandProcessor, cmdArgs []string, ok bool) {
	var c *internal.Configuration
	if c, ok = internal.ReadConfigurationFile(o); !ok {
		return nil, nil, false
	}
	var defaultSettings map[string]bool
	// TODO: [#77] replace o.ErrorWriter() with o
	if defaultSettings, ok = getDefaultSettings(o.ErrorWriter(), c.SubConfiguration("command")); !ok {
		return nil, nil, false
	}
	var initializers []subcommandInitializer
	lsSubCommand := subcommandInitializer{
		name:              lsCommand,
		defaultSubCommand: defaultSettings[lsCommand],
		initializer:       newLs,
	}
	checkSubCommand := subcommandInitializer{
		name:              checkCommand,
		defaultSubCommand: defaultSettings[checkCommand],
		initializer:       newCheck,
	}
	repairSubCommand := subcommandInitializer{
		name:              repairCommand,
		defaultSubCommand: defaultSettings[repairCommand],
		initializer:       newRepair,
	}
	postRepairSubCommand := subcommandInitializer{
		name:              postRepairCommand,
		defaultSubCommand: defaultSettings[postRepairCommand],
		initializer:       newPostRepair,
	}
	initializers = append(initializers, lsSubCommand, checkSubCommand, repairSubCommand, postRepairSubCommand)
	// TODO: [#77] replace o.ErrorWriter() with o
	cmd, cmdArgs, ok = selectSubCommand(o.ErrorWriter(), c, initializers, args)
	return
}

func getDefaultSettings(wErr io.Writer, c *internal.Configuration) (m map[string]bool, ok bool) {
	defaultCommand, ok := c.StringValue("default")
	if !ok { // no definition
		m = map[string]bool{
			checkCommand:      false,
			lsCommand:         true,
			postRepairCommand: false,
			repairCommand:     false,
		}
		ok = true
		return
	}
	m = map[string]bool{
		checkCommand:      defaultCommand == checkCommand,
		lsCommand:         defaultCommand == lsCommand,
		postRepairCommand: defaultCommand == postRepairCommand,
		repairCommand:     defaultCommand == repairCommand,
	}
	found := false
	for _, value := range m {
		if value {
			found = true
			break
		}
	}
	if !found {
		logrus.WithFields(logrus.Fields{
			fkCommandName: defaultCommand,
		}).Warn(internal.LW_INVALID_DEFAULT_COMMAND)
		fmt.Fprintf(wErr, internal.USER_INVALID_DEFAULT_COMMAND, defaultCommand)
		m = nil
		ok = false
		return
	}
	ok = true
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
