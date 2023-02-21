package commands

const defaultCommand = "list"

// IsDefault is called by the init function of each command, so it can state
// whether it's the default command - and without hardcoding that fact into one
// of the commands.
func IsDefault(commandName string) bool {
	return commandName == defaultCommand
}

// Load is meant to be called by main(), to load the commands package
func Load() {
	// does nothing
}
