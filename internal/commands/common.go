package commands

import (
	"flag"
	"mp3/internal"
	"sort"
	"strings"
)

const (
	checkCommand         = "check"
	fkCommandName        = "command"
	fkCount              = "count"
	lsCommand            = "ls"
	postRepairCommand    = "postRepair"
	repairCommand        = "repair"
	resetDatabaseCommand = "resetDatabase"
)

type CommandProcessor interface {
	name() string
	Exec(internal.OutputBus, []string) bool
}

type commandInitializer struct {
	name           string
	defaultCommand bool
	initializer    func(internal.OutputBus, *internal.Configuration, *flag.FlagSet) (CommandProcessor, bool)
}

// ProcessCommand selects which command to be run and returns the relevant
// CommandProcessor, command line arguments and ok status
func ProcessCommand(o internal.OutputBus, args []string) (cmd CommandProcessor, cmdArgs []string, ok bool) {
	var c *internal.Configuration
	if c, ok = internal.ReadConfigurationFile(o); !ok {
		return nil, nil, false
	}
	var defaultSettings map[string]bool
	if defaultSettings, ok = getDefaultSettings(o, c.SubConfiguration("command")); !ok {
		return nil, nil, false
	}
	var initializers []commandInitializer
	initializers = append(initializers, commandInitializer{
		name:           lsCommand,
		defaultCommand: defaultSettings[lsCommand],
		initializer:    newLs,
	})
	initializers = append(initializers, commandInitializer{
		name:           checkCommand,
		defaultCommand: defaultSettings[checkCommand],
		initializer:    newCheck,
	})
	initializers = append(initializers, commandInitializer{
		name:           repairCommand,
		defaultCommand: defaultSettings[repairCommand],
		initializer:    newRepair,
	})
	initializers = append(initializers, commandInitializer{
		name:           postRepairCommand,
		defaultCommand: defaultSettings[postRepairCommand],
		initializer:    newPostRepair,
	})
	initializers = append(initializers, commandInitializer{
		name:           resetDatabaseCommand,
		defaultCommand: defaultSettings[resetDatabaseCommand],
		initializer:    newResetDatabase,
	})
	cmd, cmdArgs, ok = selectCommand(o, c, initializers, args)
	return
}

func getDefaultSettings(o internal.OutputBus, c *internal.Configuration) (m map[string]bool, ok bool) {
	defaultCommand, ok := c.StringValue("default")
	if !ok { // no definition
		m = map[string]bool{
			checkCommand:         false,
			lsCommand:            true,
			postRepairCommand:    false,
			repairCommand:        false,
			resetDatabaseCommand: false,
		}
		ok = true
		return
	}
	m = map[string]bool{
		checkCommand:         defaultCommand == checkCommand,
		lsCommand:            defaultCommand == lsCommand,
		postRepairCommand:    defaultCommand == postRepairCommand,
		repairCommand:        defaultCommand == repairCommand,
		resetDatabaseCommand: defaultCommand == resetDatabaseCommand,
	}
	found := false
	for _, value := range m {
		if value {
			found = true
			break
		}
	}
	if !found {
		o.LogWriter().Warn(internal.LW_INVALID_DEFAULT_COMMAND, map[string]interface{}{
			fkCommandName: defaultCommand,
		})
		o.WriteError(internal.USER_INVALID_DEFAULT_COMMAND, defaultCommand)
		m = nil
		ok = false
		return
	}
	ok = true
	return
}

func selectCommand(o internal.OutputBus, c *internal.Configuration, i []commandInitializer, args []string) (cmd CommandProcessor, callingArgs []string, ok bool) {
	if len(i) == 0 {
		o.LogWriter().Error(internal.LE_COMMAND_COUNT, map[string]interface{}{
			fkCount: 0,
		})
		o.WriteError(internal.USER_NO_COMMANDS_DEFINED)
		return
	}
	var defaultInitializers int
	var defaultInitializerName string
	for _, initializer := range i {
		if initializer.defaultCommand {
			defaultInitializers++
			defaultInitializerName = initializer.name
		}
	}
	if defaultInitializers != 1 {
		o.LogWriter().Error(internal.LE_DEFAULT_COMMAND_COUNT, map[string]interface{}{
			fkCount: defaultInitializers,
		})
		o.WriteError(internal.USER_INCORRECT_NUMBER_OF_DEFAULT_COMMANDS_DEFINED, defaultInitializers)
		return
	}
	processorMap := make(map[string]CommandProcessor)
	allCommandsOk := true
	for _, commandInitializer := range i {
		fSet := flag.NewFlagSet(commandInitializer.name, flag.ContinueOnError)
		command, cOk := commandInitializer.initializer(o, c, fSet)
		if cOk {
			processorMap[commandInitializer.name] = command
		} else {
			allCommandsOk = false
		}
	}
	if !allCommandsOk {
		return
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
		o.LogWriter().Warn(internal.LW_UNRECOGNIZED_COMMAND, map[string]interface{}{
			fkCommandName: commandName,
		})
		var commandNames []string
		for _, initializer := range i {
			commandNames = append(commandNames, initializer.name)
		}
		sort.Strings(commandNames)
		o.WriteError(internal.USER_NO_SUCH_COMMAND, commandName, commandNames)
		return
	}
	callingArgs = args[2:]
	ok = true
	return
}

func reportBadDefault(o internal.OutputBus, section string, err error) {
	o.WriteError(internal.USER_CONFIGURATION_FILE_INVALID, internal.DefaultConfigFileName, section, err)
	o.LogWriter().Warn(internal.LW_INVALID_CONFIGURATION_DATA, map[string]interface{}{
		internal.FK_SECTION: section,
		internal.FK_ERROR:   err,
	})
}
