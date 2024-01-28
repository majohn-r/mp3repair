/*
Copyright © 2021 Marc Johnson (marc.johnson27591@gmail.com)
*/
package cmd

import (
	"os"
	"sync"
	"time"

	cmd_toolkit "github.com/majohn-r/cmd-toolkit"
	"github.com/majohn-r/output"
	"github.com/spf13/cobra"
)

var (
	// rootCmd represents the base command when called without any subcommands
	rootCmd = &cobra.Command{
		Use:   "mp3",
		Short: "A repair program for mp3 files",
		Long: `A repair program for mp3 files.

Most mp3 files contain metadata as well as audio data, and many audio systems rely on
that metadata to associate the files with specific albums and artists, and to play
those files in order. Mismatches between the file names (tracks, albums, and artists)
and that metadata subvert a user's expectations derived from reading the file names and
the names of their containing directories - tracks are mysteriously associated with
non-existent albums, tracks play out of sequence, and so forth.

The mp3 program exists to find such problems and repair the mp3 files' metadata.`,
		Example: `The mp3 program might be used like this:

First, get a listing of the available mp3 files:

mp3 ` + ListCommand + `

Then check for problems in the track metadata:

mp3 ` + CheckCommand + ` ` + CheckFilesFlag + `

If problems were found, repair the mp3 files:

mp3 ` + repairCommandName + `

The repair command creates backup files for each track it rewrites. After
spot-checking files that have been repaired, clean up those backups:

mp3 ` + postRepairCommandName + `

After repairing the mp3 files, the Windows media player system may be out of
sync with the changes. While the system will eventually catch up, accelerate
the process:

mp3 ` + resetDBCommandName,
	}
	// safe values until properly initialized
	Bus            = output.NewNilBus()
	InternalConfig = cmd_toolkit.EmptyConfiguration()
	// internals ...
	BusGetter   = getBus
	initLock    = &sync.RWMutex{}
	Initialized = false
)

func getBus() output.Bus {
	InitGlobals()
	return Bus
}

func getConfiguration() *cmd_toolkit.Configuration {
	InitGlobals()
	return InternalConfig
}

func InitGlobals() {
	initLock.Lock()
	defer initLock.Unlock()
	if !Initialized {
		ok := false
		Bus = NewBusFunc(cmd_toolkit.ProductionLogger)
		if _, err := AppNameGetFunc(); err != nil {
			AppNameSetFunc("mp3")
		}
		if LogInitFunc(Bus) && AppPathInitFunc(Bus) {
			InternalConfig, ok = ReadConfigFileFunc(Bus)
		}
		if !ok {
			ExitFunction(1)
		}
		Initialized = true
	}
}

func CookCommandLineArguments(o output.Bus, inputArgs []string) []string {
	args := []string{}
	if len(inputArgs) > 1 {
		for _, arg := range inputArgs[1:] {
			if cookedArg, err := DereferenceEnvVarFunc(arg); err != nil {
				o.WriteCanonicalError("An error was found in processng argument %q: %v", arg, err)
				o.Log(output.Error, "Invalid argument value", map[string]any{
					"argument": arg,
					"error":    err,
				})
			} else {
				args = append(args, cookedArg)
			}
		}
	}
	return args
}

type CommandExecutor interface {
	SetArgs(a []string)
	Execute() error
}

// Execute adds all child commands to the root command and sets flags
// appropriately. This is called by main.main(). It only needs to happen once to
// the rootCmd.
func Execute() {
	start := time.Now()
	o := getBus()
	RunMain(o, rootCmd, start)
}

func RunMain(o output.Bus, cmd CommandExecutor, start time.Time) {
	cookedArgs := CookCommandLineArguments(o, os.Args)
	o.Log(output.Info, "execution starts", map[string]any{
		"version":      Version,
		"timeStamp":    Creation,
		"goVersion":    GoVersionFunc(),
		"dependencies": BuildDependenciesFunc(),
		"args":         cookedArgs,
	})
	cmd.SetArgs(cookedArgs)
	err := cmd.Execute()
	exitCode := 0
	if err != nil {
		exitCode = 1
	}
	o.Log(output.Info, "execution ends", map[string]any{
		"duration": DurationCalc(start),
		"exitCode": exitCode,
	})
	if exitCode != 0 {
		o.WriteCanonicalError("%q version %s, created at %s, failed", AppName, Version, Creation)
	}
	ExitFunction(exitCode)
}

func init() {
	o := getBus()
	rootCmd.SetErr(o.ErrorWriter())
	rootCmd.SetOut(o.ConsoleWriter())
}
