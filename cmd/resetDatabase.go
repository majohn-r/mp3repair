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
		RunE: ResetDBExec,
	}
	ResetDatabaseFlags = NewSectionFlags().WithSectionName(resetDBCommandName).WithFlags(
		map[string]*FlagDetails{
			resetDBTimeout: NewFlagDetails().WithAbbreviatedName(
				resetDBTimeoutAbbr).WithUsage(fmt.Sprintf(
				"timeout in seconds (minimum %d, maximum %d) for stopping the media player"+
					" service", minTimeout, maxTimeout)).WithExpectedType(
				IntType).WithDefaultValue(
				cmd_toolkit.NewIntBounds(minTimeout, defaultTimeout, maxTimeout)),
			resetDBService: NewFlagDetails().WithUsage(
				"name of the media player service").WithExpectedType(
				StringType).WithDefaultValue("WMPNetworkSVC"),
			resetDBMetadataDir: NewFlagDetails().WithUsage(
				"directory where the media player service metadata files are stored").WithExpectedType(
				StringType).WithDefaultValue(
				filepath.Join("%USERPROFILE%", "AppData", "Local", "Microsoft", "Media Player")),
			resetDBExtension: NewFlagDetails().WithUsage(
				"extension for metadata files").WithExpectedType(
				StringType).WithDefaultValue(".wmdb"),
			resetDBForce: NewFlagDetails().WithAbbreviatedName(
				resetDBForceAbbr).WithUsage(
				"if set, force a database reset").WithExpectedType(
				BoolType).WithDefaultValue(false),
			resetDBIgnoreServiceErrors: NewFlagDetails().WithAbbreviatedName(
				resetDBIgnoreServiceErrorsAbbr).WithUsage(
				"if set, ignore service errors and delete the media player service" +
					" metadata files").WithExpectedType(BoolType).WithDefaultValue(false),
		},
	)
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

func ResetDBExec(cmd *cobra.Command, _ []string) error {
	exitError := NewExitSystemError(resetDBCommandName)
	o := getBus()
	values, eSlice := ReadFlags(cmd.Flags(), ResetDatabaseFlags)
	if ProcessFlagErrors(o, eSlice) {
		rdbs, ok := ProcessResetDBFlags(o, values)
		if ok {
			LogCommandStart(o, resetDBCommandName, map[string]any{
				resetDBTimeoutFlag:     rdbs.timeout,
				resetDBServiceFlag:     rdbs.service,
				resetDBMetadataDirFlag: rdbs.metadataDir,
				resetDBExtensionFlag:   rdbs.extension,
				resetDBForceFlag:       rdbs.force,
			})
			exitError = rdbs.ResetService(o)
		}
	}
	return ToErrorInterface(exitError)
}

type ResetDBSettings struct {
	extension           string
	force               bool
	ignoreServiceErrors bool
	metadataDir         string
	service             string
	timeout             int
}

func NewResetDBSettings() *ResetDBSettings {
	return &ResetDBSettings{}
}

func (rdbs *ResetDBSettings) WithExtension(s string) *ResetDBSettings {
	rdbs.extension = s
	return rdbs
}

func (rdbs *ResetDBSettings) WithForce(b bool) *ResetDBSettings {
	rdbs.force = b
	return rdbs
}

func (rdbs *ResetDBSettings) WithIgnoreServiceErrors(b bool) *ResetDBSettings {
	rdbs.ignoreServiceErrors = b
	return rdbs
}

func (rdbs *ResetDBSettings) WithMetadataDir(s string) *ResetDBSettings {
	rdbs.metadataDir = s
	return rdbs
}

func (rdbs *ResetDBSettings) WithService(s string) *ResetDBSettings {
	rdbs.service = s
	return rdbs
}

func (rdbs *ResetDBSettings) WithTimeout(i int) *ResetDBSettings {
	rdbs.timeout = i
	return rdbs
}

func (rdbs *ResetDBSettings) ResetService(o output.Bus) (e *ExitError) {
	if rdbs.force || Dirty() {
		stopped, e2 := rdbs.StopService(o)
		if e2 != nil {
			e = e2
		}
		e2 = rdbs.DeleteMetadataFiles(o, stopped)
		e = UpdateServiceStatus(e, e2)
		MaybeClearDirty(o, e)
	} else {
		e = NewExitUserError(resetDBCommandName)
		o.WriteCanonicalError("The %q command has no work to perform.", resetDBCommandName)
		o.WriteCanonicalError("Why?")
		o.WriteCanonicalError("The %q program has not made any changes to any mp3 files\n"+
			"since the last successful database reset.", appName)
		o.WriteError("What to do:\n")
		o.WriteCanonicalError("If you believe the Windows database needs to be reset, run"+
			" this command\nagain and use the %q flag.", resetDBForceFlag)
	}
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

func openService(manager ServiceManager, serviceName string) (rep ServiceRep, err error) {
	if manager == nil || reflect.ValueOf(manager).IsNil() {
		err = fmt.Errorf("nil manager")
	} else {
		rep, err = manager.OpenService(serviceName)
	}
	return
}

func (rdbs *ResetDBSettings) StopService(o output.Bus) (ok bool, e *ExitError) {
	var manager ServiceManager
	var err error
	if manager, err = Connect(); err != nil {
		e = NewExitSystemError(resetDBCommandName)
		o.WriteCanonicalError("An attempt to connect with the service manager failed; error"+
			" is %v", err)
		o.WriteCanonicalError("Why?\nThis often fails due to lack of permissions")
		o.WriteCanonicalError("What to do:\n" +
			"If you can, try running this command as an administrator.")
		o.Log(output.Error, "service manager connect failed", map[string]any{"error": err})
	} else {
		ok, e = rdbs.HandleService(o, manager)
	}
	return
}

func listServices(manager ServiceManager) ([]string, error) {
	if manager == nil || reflect.ValueOf(manager).IsNil() {
		return nil, fmt.Errorf("nil manager")
	}
	return manager.ListServices()
}

func (rdbs *ResetDBSettings) HandleService(o output.Bus, manager ServiceManager) (ok bool,
	e *ExitError) {
	if service, serviceError := openService(manager, rdbs.service); serviceError != nil {
		e = NewExitSystemError(resetDBCommandName)
		o.WriteCanonicalError("The service %q cannot be opened: %v", rdbs.service,
			serviceError)
		o.Log(output.Error, "service problem", map[string]any{
			"service": rdbs.service,
			"trigger": "OpenService",
			"error":   serviceError,
		})
		if serviceList, listError := listServices(manager); listError != nil {
			o.Log(output.Error, "service problem", map[string]any{
				"trigger": "ListServices",
				"error":   listError,
			})
		} else {
			ListServices(o, manager, serviceList)
		}
		disconnectManager(manager)
	} else {
		ok, e = rdbs.StopFoundService(o, manager, service)
	}
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
	} else {
		slices.Sort(services)
		m := map[string][]string{}
		for _, serviceName := range services {
			if s, err := openService(manager, serviceName); err == nil {
				AddServiceState(m, s, serviceName)
				closeService(s)
			} else {
				e := err.Error()
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
}

func AddServiceState(m map[string][]string, s ServiceRep, serviceName string) {
	if status, err := runQuery(s); err == nil {
		key := stateToStatus[status.State]
		m[key] = append(m[key], serviceName)
	} else {
		e := err.Error()
		m[e] = append(m[e], serviceName)
	}
}

func runQuery(s ServiceRep) (status svc.Status, err error) {
	if reflect.ValueOf(s).IsNil() {
		status, err = svc.Status{}, fmt.Errorf("no service")
	} else {
		status, err = s.Query()
	}
	return
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
	if status, err := runQuery(service); err != nil {
		e = NewExitSystemError(resetDBCommandName)
		o.WriteCanonicalError("An error occurred while trying to stop service %q: %v",
			rdbs.service, err)
		rdbs.ReportServiceQueryError(o, err)
	} else {
		if status.State == svc.Stopped {
			rdbs.ReportServiceStopped(o)
			ok = true
		} else if status, err = service.Control(svc.Stop); err == nil {
			if status.State == svc.Stopped {
				rdbs.ReportServiceStopped(o)
				ok = true
			} else {
				timeout := time.Now().Add(time.Duration(rdbs.timeout) * time.Second)
				ok, e = rdbs.WaitForStop(o, service, timeout, 100*time.Millisecond)
			}
		} else {
			e = NewExitSystemError(resetDBCommandName)
			o.WriteCanonicalError("The service %q cannot be stopped: %v", rdbs.service, err)
			o.Log(output.Error, "service problem", map[string]any{
				"service": rdbs.service,
				"trigger": "Stop",
				"error":   err,
			})
		}
	}
	return
}

func (rdbs *ResetDBSettings) ReportServiceQueryError(o output.Bus, err error) {
	o.Log(output.Error, "service query error", map[string]any{
		"service": rdbs.service,
		"error":   err,
	})
}

func (rdbs *ResetDBSettings) ReportServiceStopped(o output.Bus) {
	o.Log(output.Info, "service stopped", map[string]any{"service": rdbs.service})
}

func (rdbs *ResetDBSettings) WaitForStop(o output.Bus, s ServiceRep, expiration time.Time,
	checkInterval time.Duration) (ok bool, e *ExitError) {
	for {
		if expiration.Before(time.Now()) {
			e = NewExitSystemError(resetDBCommandName)
			o.WriteCanonicalError(
				"The service %q could not be stopped within the %d second timeout",
				rdbs.service, rdbs.timeout)
			o.Log(output.Error, "service problem", map[string]any{
				"service": rdbs.service,
				"trigger": "Stop",
				"error":   "timed out",
				"timeout": rdbs.timeout,
			})
			break
		}
		time.Sleep(checkInterval)
		if status, err := runQuery(s); err != nil {
			e = NewExitSystemError(resetDBCommandName)
			o.WriteCanonicalError(
				"An error occurred while attempting to stop the service %q: %v",
				rdbs.service, err)
			rdbs.ReportServiceQueryError(o, err)
			break
		} else if status.State == svc.Stopped {
			rdbs.ReportServiceStopped(o)
			ok = true
			break
		}
	}
	return
}

func (rdbs *ResetDBSettings) DeleteMetadataFiles(o output.Bus, stopped bool) (e *ExitError) {
	if !stopped {
		if !rdbs.ignoreServiceErrors {
			e = NewExitUserError(resetDBCommandName)
			o.WriteCanonicalError("Metadata files will not be deleted")
			o.WriteCanonicalError(
				"Why?\nThe music service %q could not be stopped, and %q is false",
				rdbs.service, resetDBIgnoreServiceErrorsFlag)
			o.WriteCanonicalError("What to do:\nRerun this command with %q set to true",
				resetDBIgnoreServiceErrorsFlag)
			return
		}
	}
	// either stopped or service errors are ignored
	if metadataFiles, ok := ReadDirectory(o, rdbs.metadataDir); ok {
		pathsToDelete := rdbs.FilterMetadataFiles(metadataFiles)
		if len(pathsToDelete) > 0 {
			e = rdbs.DeleteFiles(o, pathsToDelete)
		} else {
			o.WriteCanonicalConsole("No metadata files were found in %q", rdbs.metadataDir)
			o.Log(output.Info, "no files found", map[string]any{
				"directory": rdbs.metadataDir,
				"extension": rdbs.extension,
			})
		}
	}
	return
}

func (rdbs *ResetDBSettings) FilterMetadataFiles(entries []fs.DirEntry) []string {
	paths := []string{}
	for _, file := range entries {
		if strings.HasSuffix(file.Name(), rdbs.extension) {
			path := filepath.Join(rdbs.metadataDir, file.Name())
			if PlainFileExists(path) {
				paths = append(paths, path)
			}
		}
	}
	return paths
}

func (rdbs *ResetDBSettings) DeleteFiles(o output.Bus, paths []string) (e *ExitError) {
	if len(paths) != 0 {
		var count int
		for _, path := range paths {
			if err := Remove(path); err != nil {
				cmd_toolkit.LogFileDeletionFailure(o, path, err)
				e = NewExitSystemError(resetDBCommandName)
			} else {
				count++
			}
		}
		o.WriteCanonicalConsole(
			"%d out of %d metadata files have been deleted from %q", count, len(paths),
			rdbs.metadataDir)
	}
	return
}

func ProcessResetDBFlags(o output.Bus, values map[string]*FlagValue) (*ResetDBSettings, bool) {
	var err error
	result := &ResetDBSettings{}
	ok := true // optimistic
	result.timeout, _, err = GetInt(o, values, resetDBTimeout)
	if err != nil {
		ok = false
	}
	result.service, _, err = GetString(o, values, resetDBService)
	if err != nil {
		ok = false
	}
	result.metadataDir, _, err = GetString(o, values, resetDBMetadataDir)
	if err != nil {
		ok = false
	}
	result.extension, _, err = GetString(o, values, resetDBExtension)
	if err != nil {
		ok = false
	}
	result.force, _, err = GetBool(o, values, resetDBForce)
	if err != nil {
		ok = false
	}
	result.ignoreServiceErrors, _, err = GetBool(o, values, resetDBIgnoreServiceErrors)
	if err != nil {
		ok = false
	}
	return result, ok
}

func init() {
	RootCmd.AddCommand(ResetDatabaseCmd)
	addDefaults(ResetDatabaseFlags)
	o := getBus()
	c := getConfiguration()
	AddFlags(o, c, ResetDatabaseCmd.Flags(), ResetDatabaseFlags)
}
