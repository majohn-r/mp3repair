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
	// ResetDatabaseCmd represents the resetDatabase command
	ResetDatabaseCmd = &cobra.Command{
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
		RunE: ResetDBRun,
	}
	ResetDatabaseFlags = &cmdtoolkit.FlagSet{
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
	ProcessIsElevated = cmdtoolkit.ProcessIsElevated
	stateToStatus     = map[svc.State]string{
		svc.Stopped:         "stopped",
		svc.StartPending:    "start pending",
		svc.StopPending:     "stop pending",
		svc.Running:         "running",
		svc.ContinuePending: "continue pending",
		svc.PausePending:    "pause pending",
		svc.Paused:          "paused",
	}
)

func ResetDBRun(cmd *cobra.Command, _ []string) error {
	exitError := cmdtoolkit.NewExitSystemError(resetDBCommandName)
	o := getBus()
	values, eSlice := cmdtoolkit.ReadFlags(cmd.Flags(), ResetDatabaseFlags)
	if cmdtoolkit.ProcessFlagErrors(o, eSlice) {
		flags, flagsOk := ProcessResetDBFlags(o, values)
		if flagsOk {
			logCommandStart(o, resetDBCommandName, map[string]any{
				resetDBTimeoutFlag:             flags.Timeout.Value,
				resetDBServiceFlag:             flags.Service.Value,
				resetDBMetadataDirFlag:         flags.MetadataDir.Value,
				resetDBExtensionFlag:           flags.Extension.Value,
				resetDBForceFlag:               flags.Force.Value,
				resetDBIgnoreServiceErrorsFlag: flags.IgnoreServiceErrors.Value,
			})
			exitError = flags.ResetService(o)
		}
	}
	return cmdtoolkit.ToErrorInterface(exitError)
}

type ResetDBSettings struct {
	Extension           cmdtoolkit.CommandFlag[string]
	Force               cmdtoolkit.CommandFlag[bool]
	IgnoreServiceErrors cmdtoolkit.CommandFlag[bool]
	MetadataDir         cmdtoolkit.CommandFlag[string]
	Service             cmdtoolkit.CommandFlag[string]
	Timeout             cmdtoolkit.CommandFlag[int]
}

func (rDBSettings *ResetDBSettings) ResetService(o output.Bus) (e *cmdtoolkit.ExitError) {
	if rDBSettings.Force.Value || dirty() {
		stopped, e2 := rDBSettings.StopService(o)
		if e2 != nil {
			e = e2
		}
		e2 = rDBSettings.CleanUpMetadata(o, stopped)
		e = UpdateServiceStatus(e, e2)
		MaybeClearDirty(o, e)
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

func UpdateServiceStatus(currentStatus, proposedStatus *cmdtoolkit.ExitError) *cmdtoolkit.ExitError {
	if currentStatus == nil {
		return proposedStatus
	}
	return currentStatus
}

func MaybeClearDirty(o output.Bus, e *cmdtoolkit.ExitError) {
	if e == nil {
		clearDirty(o)
	}
}

type ServiceManager interface {
	Disconnect() error
	OpenService(name string) (*mgr.Service, error)
	ListServices() ([]string, error)
}

type ServiceRep interface {
	Close() error
	Control(c svc.Cmd) (svc.Status, error)
	Query() (svc.Status, error)
}

func openService(manager ServiceManager, serviceName string) (ServiceRep, error) {
	if manager == nil || reflect.ValueOf(manager).IsNil() {
		return nil, fmt.Errorf("nil manager")
	}
	return manager.OpenService(serviceName)
}

func (rDBSettings *ResetDBSettings) StopService(o output.Bus) (bool, *cmdtoolkit.ExitError) {
	if manager, connectErr := connect(); connectErr != nil {
		e := cmdtoolkit.NewExitSystemError(resetDBCommandName)
		o.WriteCanonicalError("An attempt to connect with the service manager failed; error"+
			" is '%v'", connectErr)
		OutputSystemErrorCause(o)
		o.Log(output.Error, "service manager connect failed", map[string]any{"error": connectErr})
		return false, e
	} else {
		return rDBSettings.DisableService(o, manager)
	}
}

func OutputSystemErrorCause(o output.Bus) {
	if !ProcessIsElevated() {
		o.WriteCanonicalError("Why?\nThis failure is likely to be due to lack of permissions")
		o.WriteCanonicalError("What to do:\n" +
			"If you can, try running this command as an administrator.")
	}
}

func listServices(manager ServiceManager) ([]string, error) {
	if manager == nil || reflect.ValueOf(manager).IsNil() {
		return nil, fmt.Errorf("nil manager")
	}
	return manager.ListServices()
}

func (rDBSettings *ResetDBSettings) DisableService(o output.Bus, manager ServiceManager) (ok bool,
	e *cmdtoolkit.ExitError) {
	service, serviceError := openService(manager, rDBSettings.Service.Value)
	if serviceError != nil {
		e = cmdtoolkit.NewExitSystemError(resetDBCommandName)
		o.WriteCanonicalError("The service %q cannot be opened: %v", rDBSettings.Service.Value,
			serviceError)
		o.Log(output.Error, "service problem", map[string]any{
			"service": rDBSettings.Service.Value,
			"trigger": "OpenService",
			"error":   serviceError,
		})
		serviceList, listError := listServices(manager)
		switch listError {
		case nil:
			ListServices(o, manager, serviceList)
		default:
			o.Log(output.Error, "service problem", map[string]any{
				"trigger": "ListServices",
				"error":   listError,
			})
		}
		disconnectManager(manager)
		return
	}
	ok, e = rDBSettings.StopFoundService(o, manager, service)
	return
}

func disconnectManager(manager ServiceManager) {
	if !reflect.ValueOf(manager).IsNil() && manager != nil {
		_ = manager.Disconnect()
	}
}

func ListServices(o output.Bus, manager ServiceManager, services []string) {
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
			AddServiceState(m, service, serviceName)
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

func AddServiceState(m map[string][]string, s ServiceRep, serviceName string) {
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

func runQuery(s ServiceRep) (svc.Status, error) {
	if reflect.ValueOf(s).IsNil() {
		return svc.Status{}, fmt.Errorf("no service")
	}
	return s.Query()
}

func closeService(s ServiceRep) {
	if !reflect.ValueOf(s).IsNil() {
		_ = s.Close()
	}
}

func (rDBSettings *ResetDBSettings) StopFoundService(o output.Bus, manager ServiceManager,
	service ServiceRep) (ok bool, e *cmdtoolkit.ExitError) {
	defer func() {
		_ = manager.Disconnect()
		closeService(service)
	}()
	status, svcErr := runQuery(service)
	if svcErr != nil {
		e = cmdtoolkit.NewExitSystemError(resetDBCommandName)
		o.WriteCanonicalError("An error occurred while trying to stop service %q: %v",
			rDBSettings.Service.Value, svcErr)
		rDBSettings.ReportServiceQueryError(o, svcErr)
		return
	}
	if status.State == svc.Stopped {
		rDBSettings.ReportServiceStopped(o)
		ok = true
		return
	}
	status, svcErr = service.Control(svc.Stop)
	if svcErr == nil {
		if status.State == svc.Stopped {
			rDBSettings.ReportServiceStopped(o)
			ok = true
			return
		}
		timeout := time.Now().Add(time.Duration(rDBSettings.Timeout.Value) * time.Second)
		ok, e = rDBSettings.WaitForStop(o, service, timeout, 100*time.Millisecond)
		return
	}
	e = cmdtoolkit.NewExitSystemError(resetDBCommandName)
	o.WriteCanonicalError("The service %q cannot be stopped: %v", rDBSettings.Service.Value, svcErr)
	o.Log(output.Error, "service problem", map[string]any{
		"service": rDBSettings.Service.Value,
		"trigger": "Stop",
		"error":   svcErr,
	})
	return
}

func (rDBSettings *ResetDBSettings) ReportServiceQueryError(o output.Bus, svcErr error) {
	o.Log(output.Error, "service query error", map[string]any{
		"service": rDBSettings.Service.Value,
		"error":   svcErr,
	})
}

func (rDBSettings *ResetDBSettings) ReportServiceStopped(o output.Bus) {
	o.Log(output.Info, "service stopped", map[string]any{"service": rDBSettings.Service.Value})
}

func (rDBSettings *ResetDBSettings) WaitForStop(o output.Bus, s ServiceRep, expiration time.Time,
	checkInterval time.Duration) (bool, *cmdtoolkit.ExitError) {
	for {
		if expiration.Before(time.Now()) {
			o.WriteCanonicalError(
				"The service %q could not be stopped within the %d second timeout",
				rDBSettings.Service.Value, rDBSettings.Timeout.Value)
			o.Log(output.Error, "service problem", map[string]any{
				"service": rDBSettings.Service.Value,
				"trigger": "Stop",
				"error":   "timed out",
				"timeout": rDBSettings.Timeout.Value,
			})
			return false, cmdtoolkit.NewExitSystemError(resetDBCommandName)
		}
		time.Sleep(checkInterval)
		status, svcErr := runQuery(s)
		if svcErr != nil {
			o.WriteCanonicalError(
				"An error occurred while attempting to stop the service %q: %v",
				rDBSettings.Service.Value, svcErr)
			rDBSettings.ReportServiceQueryError(o, svcErr)
			return false, cmdtoolkit.NewExitSystemError(resetDBCommandName)
		}
		if status.State == svc.Stopped {
			rDBSettings.ReportServiceStopped(o)
			return true, nil
		}
	}
}

func (rDBSettings *ResetDBSettings) CleanUpMetadata(o output.Bus, stopped bool) *cmdtoolkit.ExitError {
	if !stopped {
		if !rDBSettings.IgnoreServiceErrors.Value {
			o.WriteCanonicalError("Metadata files will not be deleted")
			o.WriteCanonicalError(
				"Why?\nThe music service %q could not be stopped, and %q is false",
				rDBSettings.Service.Value, resetDBIgnoreServiceErrorsFlag)
			o.WriteCanonicalError("What to do:\nRerun this command with %q set to true",
				resetDBIgnoreServiceErrorsFlag)
			return cmdtoolkit.NewExitUserError(resetDBCommandName)
		}
	}
	// either stopped or service errors are ignored
	metadataFiles, filesOk := readDirectory(o, rDBSettings.MetadataDir.Value)
	if !filesOk {
		return nil
	}
	pathsToDelete := rDBSettings.FilterMetadataFiles(metadataFiles)
	if len(pathsToDelete) > 0 {
		return rDBSettings.DeleteMetadataFiles(o, pathsToDelete)
	}
	o.WriteCanonicalConsole("No metadata files were found in %q", rDBSettings.MetadataDir.Value)
	o.Log(output.Info, "no files found", map[string]any{
		"directory": rDBSettings.MetadataDir.Value,
		"extension": rDBSettings.Extension.Value,
	})
	return nil
}

func (rDBSettings *ResetDBSettings) FilterMetadataFiles(entries []fs.FileInfo) []string {
	paths := make([]string, 0)
	for _, file := range entries {
		if strings.HasSuffix(file.Name(), rDBSettings.Extension.Value) {
			path := filepath.Join(rDBSettings.MetadataDir.Value, file.Name())
			if plainFileExists(path) {
				paths = append(paths, path)
			}
		}
	}
	return paths
}

func (rDBSettings *ResetDBSettings) DeleteMetadataFiles(o output.Bus, paths []string) (e *cmdtoolkit.ExitError) {
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
		rDBSettings.MetadataDir.Value)
	return
}

func ProcessResetDBFlags(o output.Bus, values map[string]*cmdtoolkit.CommandFlag[any]) (*ResetDBSettings, bool) {
	var flagErr error
	result := &ResetDBSettings{}
	flagsOk := true // optimistic
	result.Timeout, flagErr = cmdtoolkit.GetInt(o, values, resetDBTimeout)
	if flagErr != nil {
		flagsOk = false
	}
	result.Service, flagErr = cmdtoolkit.GetString(o, values, resetDBService)
	if flagErr != nil {
		flagsOk = false
	}
	result.MetadataDir, flagErr = cmdtoolkit.GetString(o, values, resetDBMetadataDir)
	if flagErr != nil {
		flagsOk = false
	}
	result.Extension, flagErr = cmdtoolkit.GetString(o, values, resetDBExtension)
	if flagErr != nil {
		flagsOk = false
	}
	result.Force, flagErr = cmdtoolkit.GetBool(o, values, resetDBForce)
	if flagErr != nil {
		flagsOk = false
	}
	result.IgnoreServiceErrors, flagErr = cmdtoolkit.GetBool(o, values, resetDBIgnoreServiceErrors)
	if flagErr != nil {
		flagsOk = false
	}
	return result, flagsOk
}

func init() {
	RootCmd.AddCommand(ResetDatabaseCmd)
	addDefaults(ResetDatabaseFlags)
	o := getBus()
	c := getConfiguration()
	cmdtoolkit.AddFlags(o, c, ResetDatabaseCmd.Flags(), ResetDatabaseFlags)
}
