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

The changes made by the '` + repairCommandName + `' command make the music files inconsistent with the
database Windows uses to organize the files into albums and artists. This command
resets that database, which it accomplishes by deleting the database files.

Prior to deleting the files, the ` + resetDBCommandName + ` command attempts to stop the Windows
media player service. If there is such an active service, this command will need to be
run as administrator. If, for whatever reasons, the service cannot be stopped, using the` +
			"\n" + resetDBIgnoreServiceErrorsFlag + ` flag allows the database files to be deleted, if possible.

This command does nothing if it determines that the repair command has not made any
changes, unless the ` + resetDBForceFlag + ` flag is set.`,
		Run: ResetDBExec,
	}
	ResetDatabaseFlags = NewSectionFlags().WithSectionName(resetDBCommandName).WithFlags(
		map[string]*FlagDetails{
			resetDBTimeout:             NewFlagDetails().WithAbbreviatedName(resetDBTimeoutAbbr).WithUsage(fmt.Sprintf("timeout in seconds (minimum %d, maximum %d) for stopping the media player service", minTimeout, maxTimeout)).WithExpectedType(IntType).WithDefaultValue(cmd_toolkit.NewIntBounds(minTimeout, defaultTimeout, maxTimeout)),
			resetDBService:             NewFlagDetails().WithUsage("name of the media player service").WithExpectedType(StringType).WithDefaultValue("WMPNetworkSVC"),
			resetDBMetadataDir:         NewFlagDetails().WithUsage("directory where the media player service metadata files are stored").WithExpectedType(StringType).WithDefaultValue(filepath.Join("%USERPROFILE%", "AppData", "Local", "Microsoft", "Media Player")),
			resetDBExtension:           NewFlagDetails().WithUsage("extension for metadata files").WithExpectedType(StringType).WithDefaultValue(".wmdb"),
			resetDBForce:               NewFlagDetails().WithAbbreviatedName(resetDBForceAbbr).WithUsage("if set, force a database reset").WithExpectedType(BoolType).WithDefaultValue(false),
			resetDBIgnoreServiceErrors: NewFlagDetails().WithAbbreviatedName(resetDBIgnoreServiceErrorsAbbr).WithUsage("if set, ignore service errors and delete the media player service metadata files").WithExpectedType(BoolType).WithDefaultValue(false),
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

func ResetDBExec(cmd *cobra.Command, _ []string) {
	status := ProgramError
	o := getBus()
	values, eSlice := ReadFlags(cmd.Flags(), ResetDatabaseFlags)
	if ProcessFlagErrors(o, eSlice) {
		rdbs, ok := ProcessResetDBFlags(o, values)
		if ok {
			LogCommandStart(o, resetDBCommandName, map[string]any{
				resetDBTimeoutFlag:     rdbs.Timeout,
				resetDBServiceFlag:     rdbs.Service,
				resetDBMetadataDirFlag: rdbs.MetadataDir,
				resetDBExtensionFlag:   rdbs.Extension,
				resetDBForceFlag:       rdbs.Force,
			})
			status = rdbs.ResetService(o)
		}
	}
	Exit(status)
}

type ResetDBSettings struct {
	Timeout             int
	Service             string
	MetadataDir         string
	Extension           string
	Force               bool
	IgnoreServiceErrors bool
}

func (rdbs *ResetDBSettings) ResetService(o output.Bus) (status int) {
	if rdbs.Force || Dirty() {
		stopped, state := rdbs.StopService(o)
		if state != Success {
			status = state
		}
		state = rdbs.DeleteMetadataFiles(o, stopped)
		status = UpdateServiceStatus(status, state)
		MaybeClearDirty(o, status)
	} else {
		status = UserError
		o.WriteCanonicalError("The %q command has no work to perform.", resetDBCommandName)
		o.WriteCanonicalError("Why?")
		o.WriteCanonicalError("The %q program has not made any changes to any mp3 files\nsince the last successful database reset.", appName)
		o.WriteError("What to do:\n")
		o.WriteCanonicalError("If you believe the Windows database needs to be reset, run this command\nagain and use the %q flag.", resetDBForceFlag)
	}
	return
}

func UpdateServiceStatus(currentStatus, proposedStatus int) int {
	if currentStatus == Success && proposedStatus != Success {
		currentStatus = proposedStatus
	}
	return currentStatus
}

func MaybeClearDirty(o output.Bus, status int) {
	if status == Success {
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

func (rdbs *ResetDBSettings) StopService(o output.Bus) (ok bool, status int) {
	var manager ServiceManager
	var err error
	if manager, err = Connect(); err != nil {
		status = SystemError
		o.WriteCanonicalError("An attempt to connect with the service manager failed; error is %v", err)
		o.WriteCanonicalError("Why?\nThis often fails due to lack of permissions")
		o.WriteCanonicalError("What to do:\nIf you can, try running this command as an administrator.")
		o.Log(output.Error, "service manager connect failed", map[string]any{"error": err})
	} else {
		ok, status = rdbs.HandleService(o, manager)
	}
	return
}

func listServices(manager ServiceManager) ([]string, error) {
	if manager == nil || reflect.ValueOf(manager).IsNil() {
		return nil, fmt.Errorf("nil manager")
	}
	return manager.ListServices()
}

func (rdbs *ResetDBSettings) HandleService(o output.Bus, manager ServiceManager) (ok bool, status int) {
	status = Success
	if service, serviceError := openService(manager, rdbs.Service); serviceError != nil {
		status = SystemError
		o.WriteCanonicalError("The service %q cannot be opened: %v", rdbs.Service, serviceError)
		o.Log(output.Error, "service issue", map[string]any{
			"service": rdbs.Service,
			"trigger": "OpenService",
			"error":   serviceError,
		})
		if serviceList, listError := listServices(manager); listError != nil {
			o.Log(output.Error, "service issue", map[string]any{
				"trigger": "ListServices",
				"error":   listError,
			})
		} else {
			ListServices(o, manager, serviceList)
		}
		disconnectManager(manager)
	} else {
		ok, status = rdbs.StopFoundService(o, manager, service)
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
		states := []string{}
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

func (rdbs *ResetDBSettings) StopFoundService(o output.Bus, manager ServiceManager, service ServiceRep) (ok bool, funcStatus int) {
	funcStatus = Success
	defer func() {
		_ = manager.Disconnect()
		closeService(service)
	}()
	if status, err := runQuery(service); err != nil {
		funcStatus = SystemError
		o.WriteCanonicalError("An error occurred while trying to stop service %q: %v", rdbs.Service, err)
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
				timeout := time.Now().Add(time.Duration(rdbs.Timeout) * time.Second)
				ok, funcStatus = rdbs.WaitForStop(o, service, timeout, 100*time.Millisecond)
			}
		} else {
			funcStatus = SystemError
			o.WriteCanonicalError("The service %q cannot be stopped: %v", rdbs.Service, err)
			o.Log(output.Error, "service issue", map[string]any{
				"service": rdbs.Service,
				"trigger": "Stop",
				"error":   err,
			})
		}
	}
	return
}

func (rdbs *ResetDBSettings) ReportServiceQueryError(o output.Bus, err error) {
	o.Log(output.Error, "service query error", map[string]any{
		"service": rdbs.Service,
		"error":   err,
	})
}

func (rdbs *ResetDBSettings) ReportServiceStopped(o output.Bus) {
	o.Log(output.Info, "service stopped", map[string]any{"service": rdbs.Service})
}

func (rdbs *ResetDBSettings) WaitForStop(o output.Bus, s ServiceRep, expiration time.Time, checkInterval time.Duration) (ok bool, funcStatus int) {
	funcStatus = Success
	for {
		if expiration.Before(time.Now()) {
			funcStatus = SystemError
			o.WriteCanonicalError("The service %q could not be stopped within the %d second timeout", rdbs.Service, rdbs.Timeout)
			o.Log(output.Error, "service issue", map[string]any{
				"service": rdbs.Service,
				"trigger": "Stop",
				"error":   "timed out",
				"timeout": rdbs.Timeout,
			})
			break
		}
		time.Sleep(checkInterval)
		if status, err := runQuery(s); err != nil {
			funcStatus = SystemError
			o.WriteCanonicalError("An error occurred while attempting to stop the service %q: %v", rdbs.Service, err)
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

func (rdbs *ResetDBSettings) DeleteMetadataFiles(o output.Bus, stopped bool) (status int) {
	status = Success
	if !stopped {
		if !rdbs.IgnoreServiceErrors {
			status = UserError
			o.WriteCanonicalError("Metadata files will not be deleted")
			o.WriteCanonicalError("Why?\nThe music service %q could not be stopped, and %q is false", rdbs.Service, resetDBIgnoreServiceErrorsFlag)
			o.WriteCanonicalError("What to do:\nRerun this command with %q set to true", resetDBIgnoreServiceErrorsFlag)
			return
		}
	}
	// either stopped or service errors are ignored
	if metadataFiles, ok := ReadDirectory(o, rdbs.MetadataDir); ok {
		pathsToDelete := rdbs.FilterMetadataFiles(metadataFiles)
		if len(pathsToDelete) > 0 {
			status = rdbs.DeleteFiles(o, pathsToDelete)
		} else {
			o.WriteCanonicalConsole("No metadata files were found in %q", rdbs.MetadataDir)
			o.Log(output.Info, "no files found", map[string]any{
				"directory": rdbs.MetadataDir,
				"extension": rdbs.Extension,
			})
		}
	}
	return
}

func (rdbs *ResetDBSettings) FilterMetadataFiles(entries []fs.DirEntry) []string {
	paths := []string{}
	for _, file := range entries {
		if strings.HasSuffix(file.Name(), rdbs.Extension) {
			path := filepath.Join(rdbs.MetadataDir, file.Name())
			if PlainFileExists(path) {
				paths = append(paths, path)
			}
		}
	}
	return paths
}

func (rdbs *ResetDBSettings) DeleteFiles(o output.Bus, paths []string) (status int) {
	status = Success
	if len(paths) != 0 {
		var count int
		for _, path := range paths {
			if err := Remove(path); err != nil {
				cmd_toolkit.LogFileDeletionFailure(o, path, err)
				status = SystemError
			} else {
				count++
			}
		}
		o.WriteCanonicalConsole("%d out of %d metadata files have been deleted from %q", count, len(paths), rdbs.MetadataDir)
	}
	return
}

func ProcessResetDBFlags(o output.Bus, values map[string]*FlagValue) (*ResetDBSettings, bool) {
	var err error
	result := &ResetDBSettings{}
	ok := true // optimistic
	result.Timeout, _, err = GetInt(o, values, resetDBTimeout)
	if err != nil {
		ok = false
	}
	result.Service, _, err = GetString(o, values, resetDBService)
	if err != nil {
		ok = false
	}
	result.MetadataDir, _, err = GetString(o, values, resetDBMetadataDir)
	if err != nil {
		ok = false
	}
	result.Extension, _, err = GetString(o, values, resetDBExtension)
	if err != nil {
		ok = false
	}
	result.Force, _, err = GetBool(o, values, resetDBForce)
	if err != nil {
		ok = false
	}
	result.IgnoreServiceErrors, _, err = GetBool(o, values, resetDBIgnoreServiceErrors)
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
