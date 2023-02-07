package commands

var defaultCommand string

func DeclareDefault(s string) {
	defaultCommand = s
}

func IsDefault(commandName string) bool {
	return commandName == defaultCommand
}
