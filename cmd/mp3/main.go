package main

import (
	"fmt"
	"mp3/internal"
	"mp3/internal/subcommands"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

func main() {
	initEnv()
	initLogging()
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

func initEnv() {
	if errors := internal.LookupEnvVars(); len(errors) > 0 {
		fmt.Println("1 or more environment variables unset")
		for _, e := range errors {
			fmt.Println(e)
		}
		os.Exit(1)
	}
}

func initLogging() {
	path := filepath.Join(internal.TmpFolder, "mp3", "logs")
	if err := os.MkdirAll(path, 0755); err != nil {
		fmt.Printf("cannot create path '%s': %v\n", path, err)
		os.Exit(1)
	}
	w := &lumberjack.Logger{
		Filename:   filepath.Join(path, "mp3.log"),
		MaxSize:    500, // megabytes
		MaxBackups: 30,
		MaxAge:     30, //days
	}
	log.SetOutput(w)
}
