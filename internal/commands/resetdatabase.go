package commands

import (
	"flag"
	"fmt"
	"io/fs"
	"mp3/internal"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

const (
	timeoutFlag           = "timeout"
	minTimeout            = 1
	defaultTimeout        = 10
	maxTimeout            = 60
	serviceFlag           = "service"
	defaultService        = "WMPNetworkSVC" // Windows Media Player Network Sharing Service
	metadataFlag          = "metadata"
	extensionFlag         = "extension"
	defaultExtension      = ".wmdb"
	fkServiceFlag         = "-" + serviceFlag
	fkTimeoutFlag         = "-" + timeoutFlag
	fkMetadataFlag        = "-" + metadataFlag
	fkExtensionFlag       = "-" + extensionFlag
	fkFileExtension       = "file extension"
	fkOperation           = "operation"
	fkService             = "service"
	fkServiceStatus       = "status"
	fkTimeout             = "timeout in seconds"
	opConnect             = "connect to service manager"
	opListServices        = "list services"
	opOpenService         = "open service"
	opQueryService        = "query service status"
	opStopService         = "stop service"
	statusContinuePending = "continue pending"
	statusPaused          = "paused"
	statusPausePending    = "pause pending"
	statusRunning         = "running"
	statusStartPending    = "start pending"
	statusStopped         = "stopped"
	statusStopPending     = "stop pending"
	errTimeout            = "operation timed out"
)

var defaultMetadata = filepath.Join("%Userprofile%", "AppData", "Local", "Microsoft", "Media Player")

var stateToStatus = map[svc.State]string{
	svc.Stopped:         statusStopped,
	svc.StartPending:    statusStartPending,
	svc.StopPending:     statusStopPending,
	svc.Running:         statusRunning,
	svc.ContinuePending: statusContinuePending,
	svc.PausePending:    statusPausePending,
	svc.Paused:          statusPaused,
}

func newResetDatabase(o internal.OutputBus, c *internal.Configuration, fSet *flag.FlagSet) (CommandProcessor, bool) {
	return newResetDatabaseCommand(o, c, fSet)
}

type resetDatabaseDefaults struct {
	timeout   int
	service   string
	metadata  string
	extension string
}

func newResetDatabaseCommand(o internal.OutputBus, c *internal.Configuration, fSet *flag.FlagSet) (*resetDatabase, bool) {
	name := fSet.Name()
	defaults, defaultsOk := evaluateResetDatabaseDefaults(o, c.SubConfiguration(name), name)
	if defaultsOk {
		timeoutDescription := fmt.Sprintf(
			"timeout in seconds (minimum %d, maximum %d) for stopping the media player service", minTimeout, maxTimeout)
		return &resetDatabase{
			n:         name,
			timeout:   fSet.Int(timeoutFlag, defaults.timeout, timeoutDescription),
			service:   fSet.String(serviceFlag, defaults.service, "name of the media player service"),
			metadata:  fSet.String(metadataFlag, defaults.metadata, "directory where the media player service metadata files are stored"),
			extension: fSet.String(extensionFlag, defaults.extension, "extension for metadata files"),
			f:         fSet,
		}, true
	}
	return nil, false
}

func evaluateResetDatabaseDefaults(o internal.OutputBus, c *internal.Configuration, name string) (defaults resetDatabaseDefaults, ok bool) {
	defaults = resetDatabaseDefaults{}
	ok = true
	var err error
	defaults.timeout, err = c.IntDefault(timeoutFlag, internal.NewIntBounds(minTimeout, defaultTimeout, maxTimeout))
	if err != nil {
		reportBadDefault(o, name, err)
		ok = false
	}
	defaults.service, err = c.StringDefault(serviceFlag, defaultService)
	if err != nil {
		reportBadDefault(o, name, err)
		ok = false
	}
	defaults.metadata, err = c.StringDefault(metadataFlag, defaultMetadata)
	if err != nil {
		reportBadDefault(o, name, err)
		ok = false
	}
	defaults.extension, err = c.StringDefault(extensionFlag, defaultExtension)
	if err != nil {
		reportBadDefault(o, name, err)
		ok = false
	}
	return
}

type resetDatabase struct {
	n         string
	timeout   *int
	service   *string
	metadata  *string
	extension *string
	f         *flag.FlagSet
}

func (r *resetDatabase) name() string {
	return r.n
}

func (r *resetDatabase) Exec(o internal.OutputBus, args []string) (ok bool) {
	if internal.ProcessArgs(o, r.f, args) {
		ok = r.runCommand(o, func() (serviceGateway, error) {
			m, err := mgr.Connect()
			if err != nil {
				return nil, err
			}
			return &sysMgr{m: m}, err
		})
	}
	return
}

func (r *resetDatabase) runCommand(o internal.OutputBus, connect func() (serviceGateway, error)) (ok bool) {
	o.LogWriter().Info(internal.LI_EXECUTING_COMMAND, map[string]interface{}{
		fkCommandName:   r.name(),
		fkServiceFlag:   *r.service,
		fkTimeoutFlag:   *r.timeout,
		fkMetadataFlag:  *r.metadata,
		fkExtensionFlag: *r.extension,
	})
	if !r.stopService(o, connect) {
		return
	}
	return r.deleteMetadata(o)
}

func (r *resetDatabase) deleteMetadata(o internal.OutputBus) bool {
	var files []fs.DirEntry
	var ok bool
	if files, ok = internal.ReadDirectory(o, *r.metadata); !ok {
		return false
	}
	pathsToDelete := r.filterMetadataFiles(files)
	if len(pathsToDelete) > 0 {
		return r.deleteMetadataFiles(o, pathsToDelete)
	}
	o.WriteConsole(true, "No metadata files were found in %q", *r.metadata)
	o.LogWriter().Info(internal.LI_NO_FILES_FOUND, map[string]interface{}{
		internal.FK_DIRECTORY: *r.metadata,
		fkFileExtension:       *r.extension,
	})
	return true
}

func (r *resetDatabase) deleteMetadataFiles(o internal.OutputBus, paths []string) bool {
	var count int
	for _, path := range paths {
		if err := os.Remove(path); err != nil {
			o.WriteError(internal.USER_CANNOT_DELETE_FILE, path, err)
			o.LogWriter().Error(internal.LE_CANNOT_DELETE_FILE, map[string]interface{}{
				internal.FK_FILE_NAME: path,
				internal.FK_ERROR:     err,
			})
		} else {
			count++
		}
	}
	o.WriteConsole(true, "%d out of %d metadata files have been deleted from %q", count, len(paths), *r.metadata)
	return count == len(paths)
}

func (r *resetDatabase) filterMetadataFiles(files []fs.DirEntry) []string {
	var paths []string
	for _, file := range files {
		if strings.HasSuffix(file.Name(), *r.extension) {
			path := filepath.Join(*r.metadata, file.Name())
			if internal.PlainFileExists(path) {
				paths = append(paths, path)
			}
		}
	}
	return paths
}

// returns true unless the service was detected in a running state and could not
// be stopped within the specified timeout
func (r *resetDatabase) stopService(o internal.OutputBus, connect func() (serviceGateway, error)) bool {
	// this is a privileged operation and fails if the user is not an administrator
	sM, s := r.openService(o, connect)
	if s == nil {
		// something unhappy happened, but, fine, we're done and we're not preventing progress
		return true
	}
	defer func() {
		_ = sM.manager().Disconnect()
		_ = s.Close()
	}()
	status, err := s.Query()
	if err != nil {
		o.WriteError(internal.USER_CANNOT_QUERY_SERVICE, *r.service, err)
		o.LogWriter().Error(internal.LE_SERVICE_ISSUE, map[string]interface{}{
			internal.FK_ERROR: err,
			fkService:         *r.service,
			fkOperation:       opQueryService,
		})
		return true
	}
	if status.State == svc.Stopped {
		r.logServiceStopped(o)
		return true
	}
	ok := status.State != svc.Running
	status, err = s.Control(svc.Stop)
	if err == nil {
		timeout := time.Now().Add(time.Duration(*r.timeout) * time.Second)
		if stopped := r.waitForStop(o, s, status, timeout, 100*time.Millisecond); stopped {
			ok = true
		}
	} else {
		o.WriteError(internal.USER_CANNOT_STOP_SERVICE, *r.service, err)
		o.LogWriter().Error(internal.LE_SERVICE_ISSUE, map[string]interface{}{
			internal.FK_ERROR: err,
			fkService:         *r.service,
			fkOperation:       opStopService,
		})
	}
	return ok
}

func (r *resetDatabase) logServiceStopped(o internal.OutputBus) {
	o.LogWriter().Info(internal.LI_SERVICE_STATUS, map[string]interface{}{
		fkService:       *r.service,
		fkServiceStatus: statusStopped,
	})
}

func (r *resetDatabase) openService(o internal.OutputBus, connect func() (serviceGateway, error)) (sM serviceGateway, s service) {
	sM, err := connect()
	if err != nil {
		o.WriteError(internal.USER_SERVICE_MGR_CONNECION_FAILED, err)
		o.LogWriter().Error(internal.LE_SERVICE_MANAGER_ISSUE, map[string]interface{}{
			internal.FK_ERROR: err,
			fkOperation:       opConnect,
		})
	} else {
		s, err = sM.openService(*r.service)
		if err != nil {
			o.WriteError(internal.USER_CANNOT_OPEN_SERVICE, *r.service, err)
			o.LogWriter().Error(internal.LE_SERVICE_ISSUE, map[string]interface{}{
				internal.FK_ERROR: err,
				fkService:         *r.service,
				fkOperation:       opOpenService,
			})
			services, err := sM.manager().ListServices()
			if err != nil {
				o.WriteError(internal.USER_CANNOT_LIST_SERVICES, err)
				o.LogWriter().Error(internal.LE_SERVICE_MANAGER_ISSUE, map[string]interface{}{
					internal.FK_ERROR: err,
					fkOperation:       opListServices,
				})
			} else {
				listAvailableServices(o, sM, services)
			}
			_ = sM.manager().Disconnect()
			sM = nil
			s = nil
		}
	}
	return
}

func (r *resetDatabase) waitForStop(o internal.OutputBus, s service, status svc.Status, timeout time.Time, checkFreq time.Duration) (ok bool) {
	if status.State == svc.Stopped {
		r.logServiceStopped(o)
		ok = true
		return
	}
	for !ok {
		if timeout.Before(time.Now()) {
			o.WriteError(internal.USER_SERVICE_STOP_TIMED_OUT, *r.service, *r.timeout)
			o.LogWriter().Error(internal.LE_SERVICE_ISSUE, map[string]interface{}{
				fkService:         *r.service,
				fkTimeout:         *r.timeout,
				fkOperation:       opStopService,
				internal.FK_ERROR: errTimeout,
			})
			break
		}
		time.Sleep(checkFreq)
		status, err := s.Query()
		if err != nil {
			o.WriteError(internal.USER_CANNOT_QUERY_SERVICE, *r.service, err)
			o.LogWriter().Error(internal.LE_SERVICE_ISSUE, map[string]interface{}{
				internal.FK_ERROR: err,
				fkService:         *r.service,
				fkOperation:       opQueryService,
			})
			break
		}
		if status.State == svc.Stopped {
			r.logServiceStopped(o)
			ok = true
		}
	}
	return
}

func listAvailableServices(o internal.OutputBus, sM serviceGateway, services []string) {
	o.WriteConsole(false, "The following services are available:\n")
	if len(services) == 0 {
		o.WriteConsole(false, "  - none -\n")
		return
	}
	sort.Strings(services)
	sMap := make(map[string][]string)
	for _, service := range services {
		if s, err := sM.openService(service); err == nil {
			if stat, err := s.Query(); err == nil {
				key := stateToStatus[stat.State]
				sMap[key] = append(sMap[key], service)
			} else {
				e := fmt.Sprintf("%v", err)
				sMap[e] = append(sMap[e], service)
			}
			s.Close()
		} else {
			e := fmt.Sprintf("%v", err)
			sMap[e] = append(sMap[e], service)
		}
	}
	var states []string
	for k := range sMap {
		states = append(states, k)
	}
	sort.Strings(states)
	for _, state := range states {
		o.WriteConsole(false, "  State %q:\n", state)
		for _, service := range sMap[state] {
			o.WriteConsole(false, "    %q\n", service)
		}
	}
}

// interface for methods on a service - allows for real services and for test
// implementations
type service interface {
	Close() error
	Query() (svc.Status, error)
	Control(c svc.Cmd) (svc.Status, error)
}

// interface for methods on a service manager - allows for real manager and for
// test implementations
type manager interface {
	Disconnect() error
	ListServices() ([]string, error)
}

// interface to obtain a manager and to open a service. The real manager returns
// a specific struct and its OpenService call cannot be easily forced into a
// generic call
type serviceGateway interface {
	openService(string) (service, error)
	manager() manager
}

type sysMgr struct {
	m *mgr.Mgr
}

func (m *sysMgr) openService(name string) (service, error) {
	return m.m.OpenService(name)
}

func (m *sysMgr) manager() manager {
	return m.m
}
