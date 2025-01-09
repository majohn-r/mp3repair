package cmd

import (
	"errors"
	"fmt"
	"os"
	"runtime/debug"
	"sync"
	"time"

	cmdtoolkit "github.com/majohn-r/cmd-toolkit"
	"github.com/majohn-r/output"
	"github.com/spf13/cobra"
)

const (
	// ElevatedPrivilegesPermissionVar is the name of an environment variable a user can set to
	// govern whether mp3repair should run with elevated privileges
	ElevatedPrivilegesPermissionVar = "MP3REPAIR_RUNS_AS_ADMIN"
	// DefaultElevatedPrivilegesPermission is the setting to use if ElevatedPrivilegesPermissionVar
	// is not set or is set to a non-boolean value
	DefaultElevatedPrivilegesPermission = true
)

var (
	rootCmd = &cobra.Command{
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

` + appName + ` ` + listCommand + ` -lrt

Then check for problems in the track metadata:

` + appName + ` ` + checkCommand + ` ` + checkFilesFlag + `

If problems were found, repair the mp3 files:

` + appName + ` ` + repairCommandName + `
The ` + repairCommandName + ` command creates backup files for each track it rewrites. After
listening to the files that have been repaired (verifying that the repair
process did not corrupt the audio), clean up those backups:

` + appName + ` ` + postRepairCommandName + `

After repairing the mp3 files, the Windows Media Player library may be out of
sync with the changes. While the library will eventually catch up, accelerate
the process:

` + appName + ` ` + resetLibraryCommandName,
	}
	bus            = output.NewNilBus()
	internalConfig = cmdtoolkit.EmptyConfiguration()
	busGetter      = getBus
	initLock       = &sync.RWMutex{}
	initialized    = false
)

func getBus() output.Bus {
	initGlobals()
	return bus
}

func getConfiguration() *cmdtoolkit.Configuration {
	initGlobals()
	return internalConfig
}

func initGlobals() {
	initLock.Lock()
	defer initLock.Unlock()
	if !initialized {
		bus = newDefaultBus(cmdtoolkit.ProductionLogger)
		configOk := false
		if initLogging(bus, appName) && initApplicationPath(bus, appName) {
			internalConfig, configOk = readDefaultsConfigFile(bus)
		}
		if !configOk {
			Exit(1)
		}
		bus.Log(output.Info, "process information", map[string]any{
			"process_id":        getPid(),
			"parent_process_id": getPpid(),
		})
		initialized = true
	}
}

func cookCommandLineArguments(o output.Bus, inputArgs []string) []string {
	args := make([]string, 0, len(inputArgs))
	if len(inputArgs) <= 1 {
		return args
	}
	for _, arg := range inputArgs[1:] {
		cookedArg, dereferenceErr := dereferenceEnvVar(arg)
		if dereferenceErr != nil {
			o.ErrorPrintf(
				"An error was found in processing argument %q: %s.\n",
				arg,
				cmdtoolkit.ErrorToString(dereferenceErr),
			)
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

type commandExecutor interface {
	SetArgs(a []string)
	Execute() error
}

// Execute adds all child commands to the root command and sets flags
// appropriately. This is called by main.main(). It only needs to happen once to
// the rootCmd.
func Execute() {
	start := time.Now()
	o := getBus()
	exitCode := runMain(o, rootCmd, start)
	Exit(exitCode)
}

func runMain(o output.Bus, cmd commandExecutor, start time.Time) int {
	defer func() {
		if r := recover(); r != nil {
			o.ErrorPrintf("A runtime error occurred: %q.\n", r)
			o.Log(output.Error, "Panic recovered", map[string]any{"error": r})
		}
	}()
	bi := getBuildData(debug.ReadBuildInfo)
	cachedGoVersion = bi.GoVersion()
	cachedBuildDependencies = bi.Dependencies()
	cookedArgs := cookCommandLineArguments(o, os.Args)
	o.Log(output.Info, "execution starts", map[string]any{
		"version":       version,
		"mainVersion":   bi.MainVersion(),
		"buildSettings": bi.Settings(),
		"timeStamp":     creation,
		"goVersion":     cachedGoVersion,
		"dependencies":  cachedBuildDependencies,
		"args":          cookedArgs,
		"defaults":      string(cmdtoolkit.WritableDefaults()),
	})
	mp3repairElevationControl.Log(o, output.Info)
	cmd.SetArgs(cookedArgs)
	err := cmd.Execute()
	exitCode := obtainExitCode(err)
	o.Log(output.Info, "execution ends", map[string]any{
		"duration": since(start),
		"exitCode": exitCode,
	})
	if exitCode != 0 {
		o.ErrorPrintf("%q version %s, created at %s, failed.\n", appName, version, creation)
	}
	return exitCode
}

func obtainExitCode(err error) int {
	switch {
	case err == nil:
		return 0
	default:
		var exitError *cmdtoolkit.ExitError
		if errors.As(err, &exitError) {
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
	rootCmd.SetErr(o.ErrorWriter())
	rootCmd.SetOut(o.ConsoleWriter())
	rootCmd.CompletionOptions.HiddenDefaultCmd = true
}
