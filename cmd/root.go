/*
Copyright © 2021 Marc Johnson (marc.johnson27591@gmail.com)
*/
package cmd

import (
	"fmt"
	"os"
	"sync"
	"time"

	cmd_toolkit "github.com/majohn-r/cmd-toolkit"
	"github.com/majohn-r/output"
	"github.com/spf13/cobra"
)

var (
	// RootCmd represents the base command when called without any subcommands
	RootCmd = &cobra.Command{
		SilenceErrors: true,
		Use:           appName,
		Short:         fmt.Sprintf("%q is a repair program for mp3 files", appName),
		Long: fmt.Sprintf("%q is a repair program for mp3 files.\n"+
			"\n"+
			"Most mp3 files, particularly if ripped from CDs, contain metadata as well as\n"+
			"audio data, and many audio systems use that metadata to associate the files\n"+
			"with specific albums and artists, and to play those files in order. Mismatches\n"+
			"between that metadata and the names of the mp3 files and the names of the\n"+
			"directories containing them (the album and artist directories) subvert the\n"+
			"user's expectations derived from reading those file and directory names -\n"+
			"tracks are mysteriously associated with non-existent albums, tracks play out\n"+
			"of sequence, and so forth.\n"+
			"\n"+
			"The %q program exists to find and repair such problems.", appName, appName),
		Example: `The ` + appName + ` program might be used like this:

First, get a listing of the available mp3 files:

` + appName + ` ` + ListCommand + ` -lrt

Then check for problems in the track metadata:

` + appName + ` ` + CheckCommand + ` ` + CheckFilesFlag + `

If problems were found, repair the mp3 files:

` + appName + ` ` + repairCommandName + `
The ` + repairCommandName + ` command creates backup files for each track it rewrites. After
listening to the files that have been repaired (verifying that the repair
process did not corrupt the audio), clean up those backups:

` + appName + ` ` + postRepairCommandName + `

After repairing the mp3 files, the Windows media player system may be out of
sync with the changes. While the system will eventually catch up, accelerate
the process:

` + appName + ` ` + resetDBCommandName,
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
		Bus = NewDefaultBus(cmd_toolkit.ProductionLogger)
		if _, appNameInitErr := cmd_toolkit.AppName(); appNameInitErr != nil {
			SetAppName(appName)
		}
		configOk := false
		if InitLogging(Bus) && InitApplicationPath(Bus) {
			InternalConfig, configOk = ReadConfigurationFile(Bus)
		}
		if !configOk {
			Exit(1)
		}
		Initialized = true
	}
}

func CookCommandLineArguments(o output.Bus, inputArgs []string) []string {
	args := make([]string, 0, len(inputArgs))
	if len(inputArgs) <= 1 {
		return args
	}
	for _, arg := range inputArgs[1:] {
		cookedArg, dereferenceErr := DereferenceEnvVar(arg)
		if dereferenceErr != nil {
			o.WriteCanonicalError("An error was found in processng argument %q: %v",
				arg, dereferenceErr)
			o.Log(output.Error, "Invalid argument value", map[string]any{
				"argument": arg,
				"error":    dereferenceErr,
			})
			continue
		}
		args = append(args, cookedArg)
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
	exitCode := RunMain(o, RootCmd, start)
	Exit(exitCode)
}

func RunMain(o output.Bus, cmd CommandExecutor, start time.Time) int {
	defer func() {
		if r := recover(); r != nil {
			o.WriteCanonicalError("A runtime error occurred: %q", r)
			o.Log(output.Error, "Panic recovered", map[string]any{"error": r})
		}
	}()
	cookedArgs := CookCommandLineArguments(o, os.Args)
	o.Log(output.Info, "execution starts", map[string]any{
		"version":      Version,
		"timeStamp":    Creation,
		"goVersion":    GoVersion(),
		"dependencies": BuildDependencies(),
		"args":         cookedArgs,
	})
	NewElevationControl().Log(o, output.Info)
	cmd.SetArgs(cookedArgs)
	err := cmd.Execute()
	exitCode := ObtainExitCode(err)
	o.Log(output.Info, "execution ends", map[string]any{
		"duration": Since(start),
		"exitCode": exitCode,
	})
	if exitCode != 0 {
		o.WriteCanonicalError("%q version %s, created at %s, failed", appName, Version,
			Creation)
	}
	return exitCode
}

func ObtainExitCode(err error) int {
	switch {
	case err == nil:
		return 0
	default:
		if exitError, ok := err.(*ExitError); ok {
			if exitError == nil {
				return 0
			}
			return exitError.Status()
		}
		return 1
	}
}

func init() {
	o := getBus()
	RootCmd.SetErr(o.ErrorWriter())
	RootCmd.SetOut(o.ConsoleWriter())
	RootCmd.CompletionOptions.HiddenDefaultCmd = true
}
