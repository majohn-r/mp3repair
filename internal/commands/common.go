package commands

import (
	"flag"
	"mp3/internal"
	"sort"
	"strings"

	"github.com/majohn-r/output"
)

type commandData struct {
	isDefault    bool
	initFunction func(output.Bus, *internal.Configuration, *flag.FlagSet) (CommandProcessor, bool)
}

var commandMap = map[string]commandData{}

func addCommandData(name string, d commandData) {
	commandMap[name] = d
}

const (
	fieldKeyCommandName = "command"
	fieldKeyCount       = "count"
)

// CommandProcessor defines the functions needed to run a command
type CommandProcessor interface {
	Exec(output.Bus, []string) bool
}

type commandInitializer struct {
	name           string
	defaultCommand bool
	initializer    func(output.Bus, *internal.Configuration, *flag.FlagSet) (CommandProcessor, bool)
}

// ProcessCommand selects which command to be run and returns the relevant
// CommandProcessor, command line arguments and ok status
func ProcessCommand(o output.Bus, args []string) (cmd CommandProcessor, cmdArgs []string, ok bool) {
	var c *internal.Configuration
	if c, ok = internal.ReadConfigurationFile(o); !ok {
		return nil, nil, false
	}
	var defaultSettings map[string]bool
	if defaultSettings, ok = getDefaultSettings(o, c.SubConfiguration("command")); !ok {
		return nil, nil, false
	}
	var initializers []commandInitializer
	for name, d := range commandMap {
		initializers = append(initializers, commandInitializer{
			name:           name,
			defaultCommand: defaultSettings[name],
			initializer:    d.initFunction,
		})
	}
	cmd, cmdArgs, ok = selectCommand(o, c, initializers, args)
	return
}

func getDefaultSettings(o output.Bus, c *internal.Configuration) (m map[string]bool, ok bool) {
	m = map[string]bool{}
	defaultCommand, ok := c.StringValue("default")
	if !ok { // no definition
		for name, d := range commandMap {
			m[name] = d.isDefault
		}
	} else {
		for name := range commandMap {
			m[name] = defaultCommand == name
		}
	}
	var defaultCommands []string
	for k, value := range m {
		if value {
			defaultCommands = append(defaultCommands, k)
		}
	}
	switch len(defaultCommands) {
	case 0:
		o.Log(output.Error, internal.LogErrorInvalidDefaultCommand, map[string]any{fieldKeyCommandName: defaultCommand})
		o.WriteCanonicalError(internal.UserInvalidDefaultCommand, defaultCommand)
		m = nil
		ok = false
	case 1:
		ok = true
	default:
		// not using a well-defined constant: this is a developer error.
		sort.Strings(defaultCommands)
		o.WriteCanonicalError("Internal error: %d commands self-selected as default: %v; pick one!", len(defaultCommands), defaultCommands)
		m = nil
		ok = false
	}
	return
}

func selectCommand(o output.Bus, c *internal.Configuration, i []commandInitializer, args []string) (cmd CommandProcessor, callingArgs []string, ok bool) {
	if len(i) == 0 {
		o.Log(output.Error, internal.LogErrorCommandCount, map[string]any{fieldKeyCount: 0})
		o.WriteCanonicalError(internal.UserNoCommandsDefined)
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
		o.Log(output.Error, internal.LogErrorDefaultCommandCount, map[string]any{fieldKeyCount: defaultInitializers})
		o.WriteCanonicalError(internal.UserIncorrectNumberOfDefaultCommandsDefined, defaultInitializers)
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
		o.Log(output.Error, internal.LogErrorUnrecognizedCommand, map[string]any{fieldKeyCommandName: commandName})
		var commandNames []string
		for _, initializer := range i {
			commandNames = append(commandNames, initializer.name)
		}
		sort.Strings(commandNames)
		o.WriteCanonicalError(internal.UserNoSuchCommand, commandName, commandNames)
		return
	}
	callingArgs = args[2:]
	ok = true
	return
}

func reportBadDefault(o output.Bus, section string, err error) {
	o.WriteCanonicalError(internal.UserConfigurationFileInvalid, internal.DefaultConfigFileName, section, err)
	o.Log(output.Error, internal.LogErrorInvalidConfigurationData, map[string]any{
		internal.FieldKeySection: section,
		internal.FieldKeyError:   err,
	})
}
