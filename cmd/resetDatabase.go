package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"time"

	cmdtoolkit "github.com/majohn-r/cmd-toolkit"
	"github.com/majohn-r/output"
	"github.com/spf13/cobra"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)
/*
Copyright © 2026 Marc Johnson (marc.johnson27591@gmail.com)
*/

const (
	resetDatabaseCommandName             = "resetDatabase"
	resetDatabaseTimeout                 = "timeout"
	resetDatabaseTimeoutFlag             = "--" + resetDatabaseTimeout
	resetDatabaseTimeoutAbbr             = "t"
	resetDatabaseForce                   = "force"
	resetDatabaseForceFlag               = "--" + resetDatabaseForce
	resetDatabaseForceAbbr               = "f"
	resetDatabaseIgnoreServiceErrors     = "ignoreServiceErrors"
	resetDatabaseIgnoreServiceErrorsAbbr = "i"
	resetDatabaseIgnoreServiceErrorsFlag = "--" + resetDatabaseIgnoreServiceErrors
	minTimeout                           = 1
	defaultTimeout                       = 10
	maxTimeout                           = 60
	windowsMediaPlayerSharingService     = "WMPNetworkSVC"
	metadataFileExtension                = ".wmdb"
)

var (
	resetDatabaseCmd = &cobra.Command{
		Use: "" + resetDatabaseCommandName +
			" [" + resetDatabaseTimeoutFlag + " seconds]" +
			" [" + resetDatabaseForceFlag + "]" +
			" [" + resetDatabaseIgnoreServiceErrorsFlag + "]",
		DisableFlagsInUseLine: true,
		Short:                 "Resets the Windows Media Player database",
		Long: fmt.Sprintf("%q", resetDatabaseCommandName) + ` resets the Windows Media Player database

The changes made by the '` + rewriteCommandName + `' command make the mp3 files inconsistent with the
Windows Media Player database which organizes the files into albums and artists. This command
resets that database, which it accomplishes by deleting the database files.

Prior to deleting the files, the ` + resetDatabaseCommandName + ` command attempts to stop the Windows
Media Player service, which allows Windows Media Player to share its database with a network. If
there is such an active service, this command will need to be run as administrator. If, for
whatever reasons, the service cannot be stopped, using the` +
			"\n" + resetDatabaseIgnoreServiceErrorsFlag + ` flag allows the database files to be deleted, if possible.

This command does nothing if it determines that the ` + rewriteCommandName + ` command has not made any
changes, unless the ` + resetDatabaseForceFlag + ` flag is set.`,
		RunE: resetDatabaseRun,
	}
	timeoutBounds      = cmdtoolkit.NewIntBounds(minTimeout, defaultTimeout, maxTimeout)
	resetDatabaseFlags = &cmdtoolkit.FlagSet{
		Name: resetDatabaseCommandName,
		Details: map[string]*cmdtoolkit.FlagDetails{
			resetDatabaseTimeout: {
				AbbreviatedName: resetDatabaseTimeoutAbbr,
				Usage: fmt.Sprintf("timeout in seconds (minimum %d, maximum %d) for stopping the "+
					"media player service", minTimeout, maxTimeout),
				ExpectedType: cmdtoolkit.IntType,
				DefaultValue: timeoutBounds,
			},
			resetDatabaseForce: {
				AbbreviatedName: resetDatabaseForceAbbr,
				Usage:           "if set, force a database reset",
				ExpectedType:    cmdtoolkit.BoolType,
				DefaultValue:    false,
			},
			resetDatabaseIgnoreServiceErrors: {
				AbbreviatedName: resetDatabaseIgnoreServiceErrorsAbbr,
				Usage: "if set, ignore service errors and delete the Windows Media Player service's " +
					"metadata files",
				ExpectedType: cmdtoolkit.BoolType,
				DefaultValue: false,
			},
		},
	}
	stateToStatus = map[svc.State]string{
		svc.Stopped:         "stopped",
		svc.StartPending:    "start pending",
		svc.StopPending:     "stop pending",
		svc.Running:         "running",
		svc.ContinuePending: "continue pending",
		svc.PausePending:    "pause pending",
		svc.Paused:          "paused",
	}
)

func resetDatabaseRun(cmd *cobra.Command, _ []string) error {
	exitError := cmdtoolkit.NewExitSystemError(resetDatabaseCommandName)
	o := getBus()
	values, eSlice := cmdtoolkit.ReadFlags(cmd.Flags(), resetDatabaseFlags)
	if cmdtoolkit.ProcessFlagErrors(o, eSlice) {
		flags, flagsOk := processResetDatabaseFlags(o, values)
		if flagsOk {
			exitError = flags.resetService(o)
		}
	}
	return cmdtoolkit.ToErrorInterface(exitError)
}

type resetDatabaseSettings struct {
	force               cmdtoolkit.CommandFlag[bool]
	ignoreServiceErrors cmdtoolkit.CommandFlag[bool]
	timeout             cmdtoolkit.CommandFlag[int]
}

func (rDBSettings *resetDatabaseSettings) resetService(o output.Bus) (e *cmdtoolkit.ExitError) {
	if rDBSettings.force.Value || dirty(o) {
		stopped, e2 := rDBSettings.stopService(o)
		if e2 != nil {
			e = e2
		}
		e2 = rDBSettings.cleanUpMetadata(o, stopped)
		e = updateServiceStatus(e, e2)
		maybeClearDirty(o, e)
		return
	}
	e = cmdtoolkit.NewExitUserError(resetDatabaseCommandName)
	o.ErrorPrintf("The %q command has no work to perform.\n", resetDatabaseCommandName)
	o.ErrorPrintln("Why?")
	o.ErrorPrintf("The %q program has not made any changes to any mp3 files\n", applicationName)
	o.ErrorPrintln("since the last successful database reset.")
	o.ErrorPrintln("What to do:")
	o.ErrorPrintln("If you believe the Windows Media Player database needs to be reset, run this command")
	o.ErrorPrintf("again and use the %q flag.\n", resetDatabaseForceFlag)
	return
}

func updateServiceStatus(currentStatus, proposedStatus *cmdtoolkit.ExitError) *cmdtoolkit.ExitError {
	if currentStatus == nil {
		return proposedStatus
	}
	return currentStatus
}

func maybeClearDirty(o output.Bus, e *cmdtoolkit.ExitError) {
	if e == nil {
		clearDirty(o)
	}
}

type serviceManager interface {
	Disconnect() error
	OpenService(name string) (*mgr.Service, error)
	ListServices() ([]string, error)
}

type serviceRep interface {
	Close() error
	Control(c svc.Cmd) (svc.Status, error)
	Query() (svc.Status, error)
}

func openService(manager serviceManager, serviceName string) (serviceRep, error) {
	if manager == nil || reflect.ValueOf(manager).IsNil() {
		return nil, fmt.Errorf("nil manager")
	}
	return manager.OpenService(serviceName)
}

func (rDBSettings *resetDatabaseSettings) stopService(o output.Bus) (bool, *cmdtoolkit.ExitError) {
	if manager, connectErr := connect(); connectErr != nil {
		e := cmdtoolkit.NewExitSystemError(resetDatabaseCommandName)
		o.ErrorPrintf(
			"An attempt to connect with the service manager failed; error is %s.\n",
			cmdtoolkit.ErrorToString(connectErr),
		)
		outputSystemErrorCause(o)
		o.Log(output.Error, "service manager connect failed", map[string]any{"error": connectErr})
		return false, e
	} else {
		return rDBSettings.disableService(o, manager)
	}
}

func outputSystemErrorCause(o output.Bus) {
	if !processIsElevated() {
		o.ErrorPrintln("Why?")
		o.ErrorPrintln("This failure is likely to be due to lack of permissions.")
		o.ErrorPrintln("What to do:")
		o.ErrorPrintln("If you can, try running this command as an administrator.")
	}
}

func listServices(manager serviceManager) ([]string, error) {
	if manager == nil || reflect.ValueOf(manager).IsNil() {
		return nil, fmt.Errorf("nil manager")
	}
	return manager.ListServices()
}

func (rDBSettings *resetDatabaseSettings) disableService(o output.Bus, manager serviceManager) (ok bool,
	e *cmdtoolkit.ExitError) {
	service, serviceError := openService(manager, windowsMediaPlayerSharingService)
	if serviceError != nil {
		e = cmdtoolkit.NewExitSystemError(resetDatabaseCommandName)
		o.ErrorPrintf(
			"The service %q cannot be opened: %s.\n",
			windowsMediaPlayerSharingService,
			cmdtoolkit.ErrorToString(serviceError),
		)
		o.Log(output.Error, "service problem", map[string]any{
			"service": windowsMediaPlayerSharingService,
			"trigger": "OpenService",
			"error":   serviceError,
		})
		serviceList, listError := listServices(manager)
		switch listError {
		case nil:
			listAvailableServices(o, manager, serviceList)
		default:
			o.Log(output.Error, "service problem", map[string]any{
				"trigger": "ListServices",
				"error":   listError,
			})
		}
		disconnectManager(manager)
		return
	}
	ok, e = rDBSettings.stopFoundService(o, manager, service)
	return
}

func disconnectManager(manager serviceManager) {
	if !reflect.ValueOf(manager).IsNil() && manager != nil {
		_ = manager.Disconnect()
	}
}

func listAvailableServices(o output.Bus, manager serviceManager, services []string) {
	o.ErrorPrintln("The following services are available:")
	if len(services) == 0 {
		o.ErrorPrintln("  - none -")
		return
	}
	slices.Sort(services)
	m := map[string][]string{}
	for _, serviceName := range services {
		service, serviceErr := openService(manager, serviceName)
		switch serviceErr {
		case nil:
			addServiceState(m, service, serviceName)
			closeService(service)
		default:
			e := serviceErr.Error()
			m[e] = append(m[e], serviceName)
		}
	}
	states := make([]string, 0, len(m))
	for k := range m {
		states = append(states, k)
	}
	slices.Sort(states)
	for _, state := range states {
		o.ErrorPrintf("  State %q:\n", state)
		for _, serviceName := range m[state] {
			o.ErrorPrintf("    %q\n", serviceName)
		}
	}
}

func addServiceState(m map[string][]string, s serviceRep, serviceName string) {
	status, queryErr := runQuery(s)
	switch queryErr {
	case nil:
		key := stateToStatus[status.State]
		m[key] = append(m[key], serviceName)
	default:
		e := queryErr.Error()
		m[e] = append(m[e], serviceName)
	}
}

func runQuery(s serviceRep) (svc.Status, error) {
	if reflect.ValueOf(s).IsNil() {
		return svc.Status{}, fmt.Errorf("no service")
	}
	return s.Query()
}

func closeService(s serviceRep) {
	if !reflect.ValueOf(s).IsNil() {
		_ = s.Close()
	}
}

func (rDBSettings *resetDatabaseSettings) stopFoundService(o output.Bus, manager serviceManager,
	service serviceRep) (ok bool, e *cmdtoolkit.ExitError) {
	defer func() {
		_ = manager.Disconnect()
		closeService(service)
	}()
	status, svcErr := runQuery(service)
	if svcErr != nil {
		e = cmdtoolkit.NewExitSystemError(resetDatabaseCommandName)
		o.ErrorPrintf(
			"An error occurred while trying to stop service %q: %s.\n",
			windowsMediaPlayerSharingService,
			cmdtoolkit.ErrorToString(svcErr),
		)
		reportServiceQueryError(o, svcErr)
		return
	}
	if status.State == svc.Stopped {
		reportServiceStopped(o)
		ok = true
		return
	}
	status, svcErr = service.Control(svc.Stop)
	if svcErr == nil {
		if status.State == svc.Stopped {
			reportServiceStopped(o)
			ok = true
			return
		}
		timeout := time.Now().Add(time.Duration(rDBSettings.timeout.Value) * time.Second)
		ok, e = rDBSettings.waitForStop(o, service, timeout, 100*time.Millisecond)
		return
	}
	e = cmdtoolkit.NewExitSystemError(resetDatabaseCommandName)
	o.ErrorPrintf(
		"The service %q cannot be stopped: %s.\n",
		windowsMediaPlayerSharingService,
		cmdtoolkit.ErrorToString(svcErr),
	)
	o.Log(output.Error, "service problem", map[string]any{
		"service": windowsMediaPlayerSharingService,
		"trigger": "Stop",
		"error":   svcErr,
	})
	return
}

func reportServiceQueryError(o output.Bus, svcErr error) {
	o.Log(output.Error, "service query error", map[string]any{
		"service": windowsMediaPlayerSharingService,
		"error":   svcErr,
	})
}

func reportServiceStopped(o output.Bus) {
	o.Log(
		output.Info,
		"service stopped",
		map[string]any{
			"service": windowsMediaPlayerSharingService,
		},
	)
}

func (rDBSettings *resetDatabaseSettings) waitForStop(o output.Bus, s serviceRep, expiration time.Time,
	checkInterval time.Duration) (bool, *cmdtoolkit.ExitError) {
	for {
		if expiration.Before(time.Now()) {
			o.ErrorPrintf(
				"The service %q could not be stopped within the %d second timeout.\n",
				windowsMediaPlayerSharingService,
				rDBSettings.timeout.Value,
			)
			o.Log(output.Error, "service problem", map[string]any{
				"service": windowsMediaPlayerSharingService,
				"trigger": "Stop",
				"error":   "timed out",
				"timeout": rDBSettings.timeout.Value,
			})
			return false, cmdtoolkit.NewExitSystemError(resetDatabaseCommandName)
		}
		time.Sleep(checkInterval)
		status, svcErr := runQuery(s)
		if svcErr != nil {
			o.ErrorPrintf(
				"An error occurred while attempting to stop the service %q: %s.\n",
				windowsMediaPlayerSharingService,
				cmdtoolkit.ErrorToString(svcErr),
			)
			reportServiceQueryError(o, svcErr)
			return false, cmdtoolkit.NewExitSystemError(resetDatabaseCommandName)
		}
		if status.State == svc.Stopped {
			reportServiceStopped(o)
			return true, nil
		}
	}
}

func metadataDirectory() string {
	return filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Local", "Microsoft", "Media Player")
}

func (rDBSettings *resetDatabaseSettings) cleanUpMetadata(o output.Bus, stopped bool) *cmdtoolkit.ExitError {
	if !stopped {
		if !rDBSettings.ignoreServiceErrors.Value {
			o.ErrorPrintln("Metadata files will not be deleted.")
			o.ErrorPrintln("Why?")
			o.ErrorPrintf(
				"The Windows Media Player sharing service %q could not be stopped, and %q is false.\n",
				windowsMediaPlayerSharingService,
				resetDatabaseIgnoreServiceErrorsFlag,
			)
			o.ErrorPrintln("What to do:")
			o.ErrorPrintf("Rerun this command with %q set to true.\n", resetDatabaseIgnoreServiceErrorsFlag)
			return cmdtoolkit.NewExitUserError(resetDatabaseCommandName)
		}
	}
	// either stopped or service errors are ignored
	dir := metadataDirectory()
	metadataFiles, filesOk := readDirectory(o, dir)
	if !filesOk {
		return nil
	}
	pathsToDelete := filterMetadataFiles(metadataFiles)
	if len(pathsToDelete) > 0 {
		return deleteMetadataFiles(o, pathsToDelete)
	}
	o.ConsolePrintf("No metadata files were found in %q.\n", dir)
	o.Log(output.Info, "no files found", map[string]any{
		"directory": dir,
		"extension": metadataFileExtension,
	})
	return nil
}

func filterMetadataFiles(entries []fs.FileInfo) []string {
	paths := make([]string, 0, len(entries))
	dir := metadataDirectory()
	for _, file := range entries {
		if strings.HasSuffix(file.Name(), metadataFileExtension) {
			path := filepath.Join(dir, file.Name())
			if plainFileExists(path) {
				paths = append(paths, path)
			}
		}
	}
	return paths
}

func deleteMetadataFiles(o output.Bus, paths []string) (e *cmdtoolkit.ExitError) {
	if len(paths) == 0 {
		return
	}
	var count int
	for _, path := range paths {
		fileErr := remove(path)
		switch {
		case fileErr != nil:
			cmdtoolkit.LogFileDeletionFailure(o, path, fileErr)
			e = cmdtoolkit.NewExitSystemError(resetDatabaseCommandName)
		default:
			count++
		}
	}
	o.ConsolePrintf(
		"%d out of %d metadata files have been deleted from %q.\n",
		count,
		len(paths),
		metadataDirectory(),
	)
	return
}

func processResetDatabaseFlags(
	o output.Bus,
	values map[string]*cmdtoolkit.CommandFlag[any],
) (*resetDatabaseSettings, bool) {
	var flagErr error
	result := &resetDatabaseSettings{}
	flagsOk := true // optimistic
	result.timeout, flagErr = cmdtoolkit.GetInt(o, values, resetDatabaseTimeout)
	if flagErr != nil {
		flagsOk = false
	} else {
		rawValue := result.timeout.Value
		result.timeout.Value = constrainBoundedValue(o, resetDatabaseTimeout, rawValue, timeoutBounds)
	}
	result.force, flagErr = cmdtoolkit.GetBool(o, values, resetDatabaseForce)
	if flagErr != nil {
		flagsOk = false
	}
	result.ignoreServiceErrors, flagErr = cmdtoolkit.GetBool(o, values, resetDatabaseIgnoreServiceErrors)
	if flagErr != nil {
		flagsOk = false
	}
	return result, flagsOk
}

func constrainBoundedValue(o output.Bus, flag string, rawValue int, bounds *cmdtoolkit.IntBounds) int {
	result := bounds.ConstrainedValue(rawValue)
	if result != rawValue {
		o.Log(output.Warning, "user-supplied value replaced", map[string]any{
			"flag":          flag,
			"providedValue": rawValue,
			"replacedBy":    result,
		})
	}
	return result
}

func init() {
	rootCmd.AddCommand(resetDatabaseCmd)
	cmdtoolkit.AddDefaults(resetDatabaseFlags)
	cmdtoolkit.AddFlags(getBus(), getConfiguration(), resetDatabaseCmd.Flags(), resetDatabaseFlags)
}
