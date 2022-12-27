package commands

import (
	"flag"
	"mp3/internal"
	"sort"
	"strings"

	"github.com/majohn-r/output"
)

type commandData struct {
	isDefault bool
	init      func(output.Bus, *internal.Configuration, *flag.FlagSet) (CommandProcessor, bool)
}

var commandMap = map[string]commandData{}

func addCommandData(name string, d commandData) {
	commandMap[name] = d
}

// CommandProcessor defines the functions needed to run a command
type CommandProcessor interface {
	Exec(output.Bus, []string) bool
}

// ProcessCommand selects which command to be run and returns the relevant
// CommandProcessor, command line arguments and ok status
func ProcessCommand(o output.Bus, args []string) (cmd CommandProcessor, cmdArgs []string, ok bool) {
	var c *internal.Configuration
	if c, ok = internal.ReadConfigurationFile(o); !ok {
		return
	}
	var m map[string]bool
	if m, ok = defaultSettings(o, c.SubConfiguration("command")); !ok {
		return
	}
	for name, d := range commandMap {
		d.isDefault = m[name]
	}
	cmd, cmdArgs, ok = selectCommand(o, c, args)
	return
}

func defaultSettings(o output.Bus, c *internal.Configuration) (m map[string]bool, ok bool) {
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
		o.Log(output.Error, "invalid default command", map[string]any{"command": defaultCommand})
		o.WriteCanonicalError("The configuration file specifies %q as the default command. There is no such command", defaultCommand)
		m = nil
		ok = false
	case 1:
		ok = true
	default:
		sort.Strings(defaultCommands)
		o.WriteCanonicalError("Internal error: %d commands self-selected as default: %v; pick one!", len(defaultCommands), defaultCommands)
		m = nil
		ok = false
	}
	return
}

func selectCommand(o output.Bus, c *internal.Configuration, args []string) (cmd CommandProcessor, cmdArgs []string, ok bool) {
	if len(commandMap) == 0 {
		o.Log(output.Error, "incorrect number of commands", map[string]any{"count": 0})
		o.WriteCanonicalError("An internal error has occurred: no commands are defined!")
		return
	}
	var n int
	var defaultCmd string
	for name, cD := range commandMap {
		if cD.isDefault {
			n++
			defaultCmd = name
		}
	}
	if n != 1 {
		o.Log(output.Error, "incorrect number of default commands", map[string]any{"count": n})
		o.WriteCanonicalError("An internal error has occurred: there are %d default commands!", n)
		return
	}
	m := make(map[string]CommandProcessor)
	allCmdsOk := true
	for name, cD := range commandMap {
		fSet := flag.NewFlagSet(name, flag.ContinueOnError)
		cmd, cOk := cD.init(o, c, fSet)
		if cOk {
			m[name] = cmd
		} else {
			allCmdsOk = false
		}
	}
	if !allCmdsOk {
		return
	}
	if len(args) < 2 {
		// no arguments at all
		cmd = m[defaultCmd]
		cmdArgs = []string{defaultCmd}
		ok = true
		return
	}
	firstArg := args[1]
	if strings.HasPrefix(firstArg, "-") {
		// first argument is a flag
		cmd = m[defaultCmd]
		cmdArgs = args[1:]
		ok = true
		return
	}
	cmd, found := m[firstArg]
	if !found {
		cmd = nil
		cmdArgs = nil
		o.Log(output.Error, "unrecognized command", map[string]any{"command": firstArg})
		var commandNames []string
		for name := range commandMap {
			commandNames = append(commandNames, name)
		}
		sort.Strings(commandNames)
		o.WriteCanonicalError("There is no command named %q; valid commands include %v", firstArg, commandNames)
		return
	}
	cmdArgs = args[2:]
	ok = true
	return
}

func reportBadDefault(o output.Bus, section string, err error) {
	internal.ReportInvalidConfigurationData(o, section, err)
}

func logStart(o output.Bus, name string, m map[string]any) {
	m["command"] = name
	o.Log(output.Info, "executing command", m)
}

func reportDirectoryCreationFailure(o output.Bus, cmd, dir string, e error) {
	internal.WriteDirectoryCreationError(o, dir, e)
	o.Log(output.Error, "cannot create directory", map[string]any{
		"command":   cmd,
		"directory": dir,
		"error":     e,
	})
}

func reportFileCreationFailure(o output.Bus, cmd, file string, e error) {
	o.WriteCanonicalError("The file %q cannot be created: %v", file, e)
	o.Log(output.Error, "cannot create file", map[string]any{
		"command":  cmd,
		"fileName": file,
		"error":    e,
	})
}

func reportFileDeletionFailure(o output.Bus, file string, e error) {
	o.WriteCanonicalError("The file %q cannot be deleted: %v", file, e)
	internal.LogFileDeletionFailure(o, file, e)
}

func reportNothingToDo(o output.Bus, cmd string, fields map[string]any) {
	o.WriteCanonicalError("You disabled all functionality for the command %q", cmd)
	o.Log(output.Error, "the user disabled all functionality", fields)
}
