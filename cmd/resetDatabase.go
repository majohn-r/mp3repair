package cmd

import (
	"fmt"
	"io/fs"
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

const (
	resetDBCommandName             = "resetDatabase"
	resetDBTimeout                 = "timeout"
	resetDBTimeoutFlag             = "--" + resetDBTimeout
	resetDBTimeoutAbbr             = "t"
	resetDBService                 = "service"
	resetDBServiceFlag             = "--" + resetDBService
	resetDBMetadataDir             = "metadataDir"
	resetDBMetadataDirFlag         = "--" + resetDBMetadataDir
	resetDBExtension               = "extension"
	resetDBExtensionFlag           = "--" + resetDBExtension
	resetDBForce                   = "force"
	resetDBForceFlag               = "--" + resetDBForce
	resetDBForceAbbr               = "f"
	resetDBIgnoreServiceErrors     = "ignoreServiceErrors"
	resetDBIgnoreServiceErrorsAbbr = "i"
	resetDBIgnoreServiceErrorsFlag = "--" + resetDBIgnoreServiceErrors
	minTimeout                     = 1
	defaultTimeout                 = 10
	maxTimeout                     = 60
)

var (
	resetDatabaseCmd = &cobra.Command{
		Use: "" + resetDBCommandName +
			" [" + resetDBTimeoutFlag + " seconds]" +
			" [" + resetDBServiceFlag + " name]" +
			" [" + resetDBMetadataDirFlag + " dir]" +
			" [" + resetDBExtensionFlag + " string]" +
			" [" + resetDBForceFlag + "]" +
			" [" + resetDBIgnoreServiceErrorsFlag + "]",
		DisableFlagsInUseLine: true,
		Short:                 "Resets the Windows music database",
		Long: fmt.Sprintf("%q", resetDBCommandName) + ` resets the Windows music database

The changes made by the '` + repairCommandName + `' command make the mp3 files inconsistent with the
database Windows uses to organize the files into albums and artists. This command
resets that database, which it accomplishes by deleting the database files.

Prior to deleting the files, the ` + resetDBCommandName + ` command attempts to stop the Windows
media player service. If there is such an active service, this command will need to be
run as administrator. If, for whatever reasons, the service cannot be stopped, using the` +
			"\n" + resetDBIgnoreServiceErrorsFlag + ` flag allows the database files to be deleted, if possible.

This command does nothing if it determines that the ` + repairCommandName + ` command has not made any
changes, unless the ` + resetDBForceFlag + ` flag is set.`,
		RunE: resetDBRun,
	}
	resetDatabaseFlags = &cmdtoolkit.FlagSet{
		Name: resetDBCommandName,
		Details: map[string]*cmdtoolkit.FlagDetails{
			resetDBTimeout: {
				AbbreviatedName: resetDBTimeoutAbbr,
				Usage:           fmt.Sprintf("timeout in seconds (minimum %d, maximum %d) for stopping the media player service", minTimeout, maxTimeout),
				ExpectedType:    cmdtoolkit.IntType,
				DefaultValue:    cmdtoolkit.NewIntBounds(minTimeout, defaultTimeout, maxTimeout),
			},
			resetDBService: {
				Usage:        "name of the media player service",
				ExpectedType: cmdtoolkit.StringType,
				DefaultValue: "WMPNetworkSVC",
			},
			resetDBMetadataDir: {
				Usage:        "directory where the media player service metadata files are stored",
				ExpectedType: cmdtoolkit.StringType,
				DefaultValue: filepath.Join("%USERPROFILE%", "AppData", "Local", "Microsoft", "Media Player"),
			},
			resetDBExtension: {
				Usage:        "extension for metadata files",
				ExpectedType: cmdtoolkit.StringType,
				DefaultValue: ".wmdb",
			},
			resetDBForce: {
				AbbreviatedName: resetDBForceAbbr,
				Usage:           "if set, force a database reset",
				ExpectedType:    cmdtoolkit.BoolType,
				DefaultValue:    false,
			},
			resetDBIgnoreServiceErrors: {
				AbbreviatedName: resetDBIgnoreServiceErrorsAbbr,
				Usage:           "if set, ignore service errors and delete the media player service metadata files",
				ExpectedType:    cmdtoolkit.BoolType,
				DefaultValue:    false,
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

func resetDBRun(cmd *cobra.Command, _ []string) error {
	exitError := cmdtoolkit.NewExitSystemError(resetDBCommandName)
	o := getBus()
	values, eSlice := cmdtoolkit.ReadFlags(cmd.Flags(), resetDatabaseFlags)
	if cmdtoolkit.ProcessFlagErrors(o, eSlice) {
		flags, flagsOk := processResetDBFlags(o, values)
		if flagsOk {
			exitError = flags.resetService(o)
		}
	}
	return cmdtoolkit.ToErrorInterface(exitError)
}

type resetDBSettings struct {
	extension           cmdtoolkit.CommandFlag[string]
	force               cmdtoolkit.CommandFlag[bool]
	ignoreServiceErrors cmdtoolkit.CommandFlag[bool]
	metadataDir         cmdtoolkit.CommandFlag[string]
	service             cmdtoolkit.CommandFlag[string]
	timeout             cmdtoolkit.CommandFlag[int]
}

func (rDBSettings *resetDBSettings) resetService(o output.Bus) (e *cmdtoolkit.ExitError) {
	if rDBSettings.force.Value || dirty() {
		stopped, e2 := rDBSettings.stopService(o)
		if e2 != nil {
			e = e2
		}
		e2 = rDBSettings.cleanUpMetadata(o, stopped)
		e = updateServiceStatus(e, e2)
		maybeClearDirty(o, e)
		return
	}
	e = cmdtoolkit.NewExitUserError(resetDBCommandName)
	o.WriteCanonicalError("The %q command has no work to perform.", resetDBCommandName)
	o.WriteCanonicalError("Why?")
	o.WriteCanonicalError("The %q program has not made any changes to any mp3 files\n"+
		"since the last successful database reset.", appName)
	o.WriteError("What to do:\n")
	o.WriteCanonicalError("If you believe the Windows database needs to be reset, run"+
		" this command\nagain and use the %q flag.", resetDBForceFlag)
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

func (rDBSettings *resetDBSettings) stopService(o output.Bus) (bool, *cmdtoolkit.ExitError) {
	if manager, connectErr := connect(); connectErr != nil {
		e := cmdtoolkit.NewExitSystemError(resetDBCommandName)
		o.WriteCanonicalError("An attempt to connect with the service manager failed; error"+
			" is '%v'", connectErr)
		outputSystemErrorCause(o)
		o.Log(output.Error, "service manager connect failed", map[string]any{"error": connectErr})
		return false, e
	} else {
		return rDBSettings.disableService(o, manager)
	}
}

func outputSystemErrorCause(o output.Bus) {
	if !processIsElevated() {
		o.WriteCanonicalError("Why?\nThis failure is likely to be due to lack of permissions")
		o.WriteCanonicalError("What to do:\n" +
			"If you can, try running this command as an administrator.")
	}
}

func listServices(manager serviceManager) ([]string, error) {
	if manager == nil || reflect.ValueOf(manager).IsNil() {
		return nil, fmt.Errorf("nil manager")
	}
	return manager.ListServices()
}

func (rDBSettings *resetDBSettings) disableService(o output.Bus, manager serviceManager) (ok bool,
	e *cmdtoolkit.ExitError) {
	service, serviceError := openService(manager, rDBSettings.service.Value)
	if serviceError != nil {
		e = cmdtoolkit.NewExitSystemError(resetDBCommandName)
		o.WriteCanonicalError("The service %q cannot be opened: %v", rDBSettings.service.Value,
			serviceError)
		o.Log(output.Error, "service problem", map[string]any{
			"service": rDBSettings.service.Value,
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
	o.WriteError("The following services are available:\n")
	if len(services) == 0 {
		o.WriteError("  - none -\n")
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
		o.WriteError("  State %q:\n", state)
		for _, serviceName := range m[state] {
			o.WriteError("    %q\n", serviceName)
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

func (rDBSettings *resetDBSettings) stopFoundService(o output.Bus, manager serviceManager,
	service serviceRep) (ok bool, e *cmdtoolkit.ExitError) {
	defer func() {
		_ = manager.Disconnect()
		closeService(service)
	}()
	status, svcErr := runQuery(service)
	if svcErr != nil {
		e = cmdtoolkit.NewExitSystemError(resetDBCommandName)
		o.WriteCanonicalError("An error occurred while trying to stop service %q: %v",
			rDBSettings.service.Value, svcErr)
		rDBSettings.reportServiceQueryError(o, svcErr)
		return
	}
	if status.State == svc.Stopped {
		rDBSettings.reportServiceStopped(o)
		ok = true
		return
	}
	status, svcErr = service.Control(svc.Stop)
	if svcErr == nil {
		if status.State == svc.Stopped {
			rDBSettings.reportServiceStopped(o)
			ok = true
			return
		}
		timeout := time.Now().Add(time.Duration(rDBSettings.timeout.Value) * time.Second)
		ok, e = rDBSettings.waitForStop(o, service, timeout, 100*time.Millisecond)
		return
	}
	e = cmdtoolkit.NewExitSystemError(resetDBCommandName)
	o.WriteCanonicalError("The service %q cannot be stopped: %v", rDBSettings.service.Value, svcErr)
	o.Log(output.Error, "service problem", map[string]any{
		"service": rDBSettings.service.Value,
		"trigger": "Stop",
		"error":   svcErr,
	})
	return
}

func (rDBSettings *resetDBSettings) reportServiceQueryError(o output.Bus, svcErr error) {
	o.Log(output.Error, "service query error", map[string]any{
		"service": rDBSettings.service.Value,
		"error":   svcErr,
	})
}

func (rDBSettings *resetDBSettings) reportServiceStopped(o output.Bus) {
	o.Log(output.Info, "service stopped", map[string]any{"service": rDBSettings.service.Value})
}

func (rDBSettings *resetDBSettings) waitForStop(o output.Bus, s serviceRep, expiration time.Time,
	checkInterval time.Duration) (bool, *cmdtoolkit.ExitError) {
	for {
		if expiration.Before(time.Now()) {
			o.WriteCanonicalError(
				"The service %q could not be stopped within the %d second timeout",
				rDBSettings.service.Value, rDBSettings.timeout.Value)
			o.Log(output.Error, "service problem", map[string]any{
				"service": rDBSettings.service.Value,
				"trigger": "Stop",
				"error":   "timed out",
				"timeout": rDBSettings.timeout.Value,
			})
			return false, cmdtoolkit.NewExitSystemError(resetDBCommandName)
		}
		time.Sleep(checkInterval)
		status, svcErr := runQuery(s)
		if svcErr != nil {
			o.WriteCanonicalError(
				"An error occurred while attempting to stop the service %q: %v",
				rDBSettings.service.Value, svcErr)
			rDBSettings.reportServiceQueryError(o, svcErr)
			return false, cmdtoolkit.NewExitSystemError(resetDBCommandName)
		}
		if status.State == svc.Stopped {
			rDBSettings.reportServiceStopped(o)
			return true, nil
		}
	}
}

func (rDBSettings *resetDBSettings) cleanUpMetadata(o output.Bus, stopped bool) *cmdtoolkit.ExitError {
	if !stopped {
		if !rDBSettings.ignoreServiceErrors.Value {
			o.WriteCanonicalError("Metadata files will not be deleted")
			o.WriteCanonicalError(
				"Why?\nThe music service %q could not be stopped, and %q is false",
				rDBSettings.service.Value, resetDBIgnoreServiceErrorsFlag)
			o.WriteCanonicalError("What to do:\nRerun this command with %q set to true",
				resetDBIgnoreServiceErrorsFlag)
			return cmdtoolkit.NewExitUserError(resetDBCommandName)
		}
	}
	// either stopped or service errors are ignored
	metadataFiles, filesOk := readDirectory(o, rDBSettings.metadataDir.Value)
	if !filesOk {
		return nil
	}
	pathsToDelete := rDBSettings.filterMetadataFiles(metadataFiles)
	if len(pathsToDelete) > 0 {
		return rDBSettings.deleteMetadataFiles(o, pathsToDelete)
	}
	o.WriteCanonicalConsole("No metadata files were found in %q", rDBSettings.metadataDir.Value)
	o.Log(output.Info, "no files found", map[string]any{
		"directory": rDBSettings.metadataDir.Value,
		"extension": rDBSettings.extension.Value,
	})
	return nil
}

func (rDBSettings *resetDBSettings) filterMetadataFiles(entries []fs.FileInfo) []string {
	paths := make([]string, 0, len(entries))
	for _, file := range entries {
		if strings.HasSuffix(file.Name(), rDBSettings.extension.Value) {
			path := filepath.Join(rDBSettings.metadataDir.Value, file.Name())
			if plainFileExists(path) {
				paths = append(paths, path)
			}
		}
	}
	return paths
}

func (rDBSettings *resetDBSettings) deleteMetadataFiles(o output.Bus, paths []string) (e *cmdtoolkit.ExitError) {
	if len(paths) == 0 {
		return
	}
	var count int
	for _, path := range paths {
		fileErr := remove(path)
		switch {
		case fileErr != nil:
			cmdtoolkit.LogFileDeletionFailure(o, path, fileErr)
			e = cmdtoolkit.NewExitSystemError(resetDBCommandName)
		default:
			count++
		}
	}
	o.WriteCanonicalConsole(
		"%d out of %d metadata files have been deleted from %q", count, len(paths),
		rDBSettings.metadataDir.Value)
	return
}

func processResetDBFlags(o output.Bus, values map[string]*cmdtoolkit.CommandFlag[any]) (*resetDBSettings, bool) {
	var flagErr error
	result := &resetDBSettings{}
	flagsOk := true // optimistic
	result.timeout, flagErr = cmdtoolkit.GetInt(o, values, resetDBTimeout)
	if flagErr != nil {
		flagsOk = false
	}
	result.service, flagErr = cmdtoolkit.GetString(o, values, resetDBService)
	if flagErr != nil {
		flagsOk = false
	}
	result.metadataDir, flagErr = cmdtoolkit.GetString(o, values, resetDBMetadataDir)
	if flagErr != nil {
		flagsOk = false
	}
	result.extension, flagErr = cmdtoolkit.GetString(o, values, resetDBExtension)
	if flagErr != nil {
		flagsOk = false
	}
	result.force, flagErr = cmdtoolkit.GetBool(o, values, resetDBForce)
	if flagErr != nil {
		flagsOk = false
	}
	result.ignoreServiceErrors, flagErr = cmdtoolkit.GetBool(o, values, resetDBIgnoreServiceErrors)
	if flagErr != nil {
		flagsOk = false
	}
	return result, flagsOk
}

func init() {
	rootCmd.AddCommand(resetDatabaseCmd)
	cmdtoolkit.AddDefaults(resetDatabaseFlags)
	cmdtoolkit.AddFlags(getBus(), getConfiguration(), resetDatabaseCmd.Flags(), resetDatabaseFlags)
}
