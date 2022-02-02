package main

import (
	"fmt"
	"mp3/internal/subcommands"
	"os"
)

func main() {
	sCmdMap := make(map[string]subcommands.SubCommand)
	lsCommand := subcommands.NewLsCommand()
	lsName := lsCommand.Name()
	sCmdMap[lsName] = lsCommand
	checkCommand := subcommands.NewCheckCommand()
	sCmdMap[checkCommand.Name()] = checkCommand
	repairCommand := subcommands.NewRepairCommand()
	sCmdMap[repairCommand.Name()] = repairCommand

	if len(os.Args) < 2 {
		var args []string
		args = append(args, lsName)
		lsCommand.Exec(args)
	} else {
		sCmd, found := sCmdMap[os.Args[1]]
		if !found {
			fmt.Printf("%s: argument '%s' is not recognized\n", os.Args[0], os.Args[1])
		} else {
			sCmd.Exec(os.Args[2:])
		}
	}
}
