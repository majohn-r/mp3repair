package main

import (
	"fmt"
	"mp3/internal"
	"mp3/internal/subcommands"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

func main() {
	initEnv()
	initLogging()
	cmd, args := subcommands.ProcessCommand(os.Args)
	if cmd == nil {
		fmt.Printf("subcommand %q is not recognized\n", os.Args[1])
		os.Exit(1)
	} else {
		cmd.Exec(args)
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
	logrus.SetOutput(internal.ConfigureLogging(path))
	internal.CleanupLogFiles(path)
}
