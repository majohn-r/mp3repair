package main

import (
	"fmt"
	"mp3/internal"
	"mp3/internal/subcommands"
	"os"
	"path/filepath"
	"time"

	rotatelogs "github.com/lestrrat/go-file-rotatelogs"
	log "github.com/sirupsen/logrus"
)

const (
	day   time.Duration = time.Hour * 24
	month time.Duration = day * 30
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
	path := filepath.Join(internal.TmpFolder, "mp3","logs")
	if err := os.MkdirAll(path, 0755); err != nil {
		fmt.Printf("cannot create path '%s': %v\n", path, err)
		os.Exit(1)
	}
	writer, err := rotatelogs.New(
		filepath.Join(path, "%Y-%m-%d.log"),
		rotatelogs.WithLinkName(filepath.Join(path,"latest")),
		rotatelogs.WithMaxAge(month),
		rotatelogs.WithRotationTime(day),
	)
	if err != nil {
		fmt.Printf("failed to initialize logging: %v\n", err)
		os.Exit(1)
	}
	log.SetOutput(writer)
}
