/*
Copyright © 2021 Marc Johnson (marc.johnson27591@gmail.com)
*/
package cmd

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"time"

	cmd_toolkit "github.com/majohn-r/cmd-toolkit"
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
	ResetDatabaseFlags = &SectionFlags{
		SectionName: resetDBCommandName,
		Details: map[string]*FlagDetails{
			resetDBTimeout: {
				AbbreviatedName: resetDBTimeoutAbbr,
				Usage:           fmt.Sprintf("timeout in seconds (minimum %d, maximum %d) for stopping the media player service", minTimeout, maxTimeout),
				ExpectedType:    IntType,
				DefaultValue:    cmd_toolkit.NewIntBounds(minTimeout, defaultTimeout, maxTimeout),
			},
			resetDBService: {
				Usage:        "name of the media player service",
				ExpectedType: StringType,
				DefaultValue: "WMPNetworkSVC",
			},
			resetDBMetadataDir: {
				Usage:        "directory where the media player service metadata files are stored",
				ExpectedType: StringType,
				DefaultValue: filepath.Join("%USERPROFILE%", "AppData", "Local", "Microsoft", "Media Player"),
			},
			resetDBExtension: {
				Usage:        "extension for metadata files",
				ExpectedType: StringType,
				DefaultValue: ".wmdb",
			},
			resetDBForce: {
				AbbreviatedName: resetDBForceAbbr,
				Usage:           "if set, force a database reset",
				ExpectedType:    BoolType,
				DefaultValue:    false,
			},
			resetDBIgnoreServiceErrors: {
				AbbreviatedName: resetDBIgnoreServiceErrorsAbbr,
				Usage:           "if set, ignore service errors and delete the media player service metadata files",
				ExpectedType:    BoolType,
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

func ResetDBRun(cmd *cobra.Command, _ []string) error {
	exitError := NewExitSystemError(resetDBCommandName)
	o := getBus()
	values, eSlice := ReadFlags(cmd.Flags(), ResetDatabaseFlags)
	if ProcessFlagErrors(o, eSlice) {
		rdbs, flagsOk := ProcessResetDBFlags(o, values)
		if flagsOk {
			LogCommandStart(o, resetDBCommandName, map[string]any{
				resetDBTimeoutFlag:             rdbs.Timeout.Value,
				resetDBServiceFlag:             rdbs.Service.Value,
				resetDBMetadataDirFlag:         rdbs.MetadataDir.Value,
				resetDBExtensionFlag:           rdbs.Extension.Value,
				resetDBForceFlag:               rdbs.Force.Value,
				resetDBIgnoreServiceErrorsFlag: rdbs.IgnoreServiceErrors.Value,
			})
			exitError = rdbs.ResetService(o)
		}
	}
	return ToErrorInterface(exitError)
}

type ResetDBSettings struct {
	Extension           StringValue
	Force               BoolValue
	IgnoreServiceErrors BoolValue
	MetadataDir         StringValue
	Service             StringValue
	Timeout             IntValue
}

func (rdbs *ResetDBSettings) ResetService(o output.Bus) (e *ExitError) {
	if rdbs.Force.Value || Dirty() {
		stopped, e2 := rdbs.StopService(o)
		if e2 != nil {
			e = e2
		}
		e2 = rdbs.CleanUpMetadata(o, stopped)
		e = UpdateServiceStatus(e, e2)
		MaybeClearDirty(o, e)
		return
	}
	e = NewExitUserError(resetDBCommandName)
	o.WriteCanonicalError("The %q command has no work to perform.", resetDBCommandName)
	o.WriteCanonicalError("Why?")
	o.WriteCanonicalError("The %q program has not made any changes to any mp3 files\n"+
		"since the last successful database reset.", appName)
	o.WriteError("What to do:\n")
	o.WriteCanonicalError("If you believe the Windows database needs to be reset, run"+
		" this command\nagain and use the %q flag.", resetDBForceFlag)
	return
}

func UpdateServiceStatus(currentStatus, proposedStatus *ExitError) *ExitError {
	if currentStatus == nil {
		return proposedStatus
	}
	return currentStatus
}

func MaybeClearDirty(o output.Bus, e *ExitError) {
	if e == nil {
		ClearDirty(o)
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

func (rdbs *ResetDBSettings) StopService(o output.Bus) (bool, *ExitError) {
	if manager, connectErr := Connect(); connectErr != nil {
		e := NewExitSystemError(resetDBCommandName)
		o.WriteCanonicalError("An attempt to connect with the service manager failed; error"+
			" is '%v'", connectErr)
		OutputSystemErrorCause(o)
		o.Log(output.Error, "service manager connect failed", map[string]any{"error": connectErr})
		return false, e
	} else {
		return rdbs.DisableService(o, manager)
	}
}

func OutputSystemErrorCause(o output.Bus) {
	if !processIsElevated() {
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

func (rdbs *ResetDBSettings) DisableService(o output.Bus, manager ServiceManager) (ok bool,
	e *ExitError) {
	service, serviceError := openService(manager, rdbs.Service.Value)
	if serviceError != nil {
		e = NewExitSystemError(resetDBCommandName)
		o.WriteCanonicalError("The service %q cannot be opened: %v", rdbs.Service.Value,
			serviceError)
		o.Log(output.Error, "service problem", map[string]any{
			"service": rdbs.Service.Value,
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
	ok, e = rdbs.StopFoundService(o, manager, service)
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
		for _, svc := range m[state] {
			o.WriteError("    %q\n", svc)
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

func (rdbs *ResetDBSettings) StopFoundService(o output.Bus, manager ServiceManager,
	service ServiceRep) (ok bool, e *ExitError) {
	defer func() {
		_ = manager.Disconnect()
		closeService(service)
	}()
	status, svcErr := runQuery(service)
	if svcErr != nil {
		e = NewExitSystemError(resetDBCommandName)
		o.WriteCanonicalError("An error occurred while trying to stop service %q: %v",
			rdbs.Service.Value, svcErr)
		rdbs.ReportServiceQueryError(o, svcErr)
		return
	}
	if status.State == svc.Stopped {
		rdbs.ReportServiceStopped(o)
		ok = true
		return
	}
	status, svcErr = service.Control(svc.Stop)
	if svcErr == nil {
		if status.State == svc.Stopped {
			rdbs.ReportServiceStopped(o)
			ok = true
			return
		}
		timeout := time.Now().Add(time.Duration(rdbs.Timeout.Value) * time.Second)
		ok, e = rdbs.WaitForStop(o, service, timeout, 100*time.Millisecond)
		return
	}
	e = NewExitSystemError(resetDBCommandName)
	o.WriteCanonicalError("The service %q cannot be stopped: %v", rdbs.Service.Value, svcErr)
	o.Log(output.Error, "service problem", map[string]any{
		"service": rdbs.Service.Value,
		"trigger": "Stop",
		"error":   svcErr,
	})
	return
}

func (rdbs *ResetDBSettings) ReportServiceQueryError(o output.Bus, svcErr error) {
	o.Log(output.Error, "service query error", map[string]any{
		"service": rdbs.Service.Value,
		"error":   svcErr,
	})
}

func (rdbs *ResetDBSettings) ReportServiceStopped(o output.Bus) {
	o.Log(output.Info, "service stopped", map[string]any{"service": rdbs.Service.Value})
}

func (rdbs *ResetDBSettings) WaitForStop(o output.Bus, s ServiceRep, expiration time.Time,
	checkInterval time.Duration) (bool, *ExitError) {
	for {
		if expiration.Before(time.Now()) {
			o.WriteCanonicalError(
				"The service %q could not be stopped within the %d second timeout",
				rdbs.Service.Value, rdbs.Timeout.Value)
			o.Log(output.Error, "service problem", map[string]any{
				"service": rdbs.Service.Value,
				"trigger": "Stop",
				"error":   "timed out",
				"timeout": rdbs.Timeout.Value,
			})
			return false, NewExitSystemError(resetDBCommandName)
		}
		time.Sleep(checkInterval)
		status, svcErr := runQuery(s)
		if svcErr != nil {
			o.WriteCanonicalError(
				"An error occurred while attempting to stop the service %q: %v",
				rdbs.Service.Value, svcErr)
			rdbs.ReportServiceQueryError(o, svcErr)
			return false, NewExitSystemError(resetDBCommandName)
		}
		if status.State == svc.Stopped {
			rdbs.ReportServiceStopped(o)
			return true, nil
		}
	}
}

func (rdbs *ResetDBSettings) CleanUpMetadata(o output.Bus, stopped bool) *ExitError {
	if !stopped {
		if !rdbs.IgnoreServiceErrors.Value {
			o.WriteCanonicalError("Metadata files will not be deleted")
			o.WriteCanonicalError(
				"Why?\nThe music service %q could not be stopped, and %q is false",
				rdbs.Service.Value, resetDBIgnoreServiceErrorsFlag)
			o.WriteCanonicalError("What to do:\nRerun this command with %q set to true",
				resetDBIgnoreServiceErrorsFlag)
			return NewExitUserError(resetDBCommandName)
		}
	}
	// either stopped or service errors are ignored
	metadataFiles, filesOk := ReadDirectory(o, rdbs.MetadataDir.Value)
	if !filesOk {
		return nil
	}
	pathsToDelete := rdbs.FilterMetadataFiles(metadataFiles)
	if len(pathsToDelete) > 0 {
		return rdbs.DeleteMetadataFiles(o, pathsToDelete)
	}
	o.WriteCanonicalConsole("No metadata files were found in %q", rdbs.MetadataDir.Value)
	o.Log(output.Info, "no files found", map[string]any{
		"directory": rdbs.MetadataDir.Value,
		"extension": rdbs.Extension.Value,
	})
	return nil
}

func (rdbs *ResetDBSettings) FilterMetadataFiles(entries []fs.FileInfo) []string {
	paths := []string{}
	for _, file := range entries {
		if strings.HasSuffix(file.Name(), rdbs.Extension.Value) {
			path := filepath.Join(rdbs.MetadataDir.Value, file.Name())
			if PlainFileExists(path) {
				paths = append(paths, path)
			}
		}
	}
	return paths
}

func (rdbs *ResetDBSettings) DeleteMetadataFiles(o output.Bus, paths []string) (e *ExitError) {
	if len(paths) == 0 {
		return
	}
	var count int
	for _, path := range paths {
		fileErr := Remove(path)
		switch {
		case fileErr != nil:
			cmd_toolkit.LogFileDeletionFailure(o, path, fileErr)
			e = NewExitSystemError(resetDBCommandName)
		default:
			count++
		}
	}
	o.WriteCanonicalConsole(
		"%d out of %d metadata files have been deleted from %q", count, len(paths),
		rdbs.MetadataDir.Value)
	return
}

func ProcessResetDBFlags(o output.Bus, values map[string]*FlagValue) (*ResetDBSettings, bool) {
	var flagErr error
	result := &ResetDBSettings{}
	flagsOk := true // optimistic
	result.Timeout, flagErr = GetInt(o, values, resetDBTimeout)
	if flagErr != nil {
		flagsOk = false
	}
	result.Service, flagErr = GetString(o, values, resetDBService)
	if flagErr != nil {
		flagsOk = false
	}
	result.MetadataDir, flagErr = GetString(o, values, resetDBMetadataDir)
	if flagErr != nil {
		flagsOk = false
	}
	result.Extension, flagErr = GetString(o, values, resetDBExtension)
	if flagErr != nil {
		flagsOk = false
	}
	result.Force, flagErr = GetBool(o, values, resetDBForce)
	if flagErr != nil {
		flagsOk = false
	}
	result.IgnoreServiceErrors, flagErr = GetBool(o, values, resetDBIgnoreServiceErrors)
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
	AddFlags(o, c, ResetDatabaseCmd.Flags(), ResetDatabaseFlags)
}
