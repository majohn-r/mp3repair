package main

import (
	"fmt"
	"mp3/internal/subcommands"
	"os"
)

func main() {
	sCmdMap := make(map[string]subcommands.CommandProcessor)
	lsCommand := subcommands.NewLsCommandProcessor()
	lsName := lsCommand.Name()
	sCmdMap[lsName] = lsCommand
	checkCommand := subcommands.NewCheckCommandProcessor()
	sCmdMap[checkCommand.Name()] = checkCommand
	repairCommand := subcommands.NewRepairCommandProcessor()
	sCmdMap[repairCommand.Name()] = repairCommand

	if len(os.Args) < 2 {
		lsCommand.Exec([]string{lsName})
	} else {
		commandName := os.Args[1]
		sCmd, found := sCmdMap[commandName]
		if !found {
			fmt.Printf("subcommand '%s' is not recognized\n", commandName)
			os.Exit(1)
		}
		sCmd.Exec(os.Args[2:])
	}
}
