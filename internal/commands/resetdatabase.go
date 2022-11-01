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

func init() {
	addCommandData(resetDatabaseCommandName, commandData{isDefault: false, initFunction: newResetDatabase})
	defaultMetadata = filepath.Join("%USERPROFILE%", "AppData", "Local", "Microsoft", "Media Player")
	addDefaultMapping(resetDatabaseCommandName, map[string]any{
		extensionFlag: defaultExtension,
		metadataFlag:  defaultMetadata,
		serviceFlag:   defaultService,
		timeoutFlag:   defaultTimeout,
	})
}

const (
	resetDatabaseCommandName = "resetDatabase"

	timeoutFlag   = "timeout"
	serviceFlag   = "service"
	metadataFlag  = "metadata"
	extensionFlag = "extension"

	minTimeout     = 1
	defaultTimeout = 10
	maxTimeout     = 60

	defaultService   = "WMPNetworkSVC" // Windows Media Player Network Sharing Service
	defaultExtension = ".wmdb"

	fieldKeyServiceFlag   = "-" + serviceFlag
	fieldKeyTimeoutFlag   = "-" + timeoutFlag
	fieldKeyMetadataFlag  = "-" + metadataFlag
	fieldKeyExtensionFlag = "-" + extensionFlag
	fieldKeyFileExtension = "file extension"
	fieldKeyOperation     = "operation"
	fieldKeyService       = "service"
	fieldKeyServiceStatus = "status"
	fieldKeyTimeout       = "timeout in seconds"

	operationConnect      = "connect to service manager"
	operationListServices = "list services"
	operationOpenService  = "open service"
	operationQueryService = "query service status"
	operationStopService  = "stop service"

	statusContinuePending = "continue pending"
	statusPaused          = "paused"
	statusPausePending    = "pause pending"
	statusRunning         = "running"
	statusStartPending    = "start pending"
	statusStopped         = "stopped"
	statusStopPending     = "stop pending"

	errTimeout = "operation timed out"
)

var defaultMetadata string

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
	defaults, defaultsOk := evaluateResetDatabaseDefaults(o, c.SubConfiguration(resetDatabaseCommandName), resetDatabaseCommandName)
	if defaultsOk {
		timeoutDescription := fmt.Sprintf(
			"timeout in seconds (minimum %d, maximum %d) for stopping the media player service", minTimeout, maxTimeout)
		timeoutUsage := internal.DecorateIntFlagUsage(timeoutDescription, defaults.timeout)
		serviceUsage := internal.DecorateStringFlagUsage("name of the media player `service`", defaults.service)
		metadataUsage := internal.DecorateStringFlagUsage("`directory` where the media player service metadata files are stored", defaults.metadata)
		extensionUsage := internal.DecorateStringFlagUsage("`extension` for metadata files", defaults.extension)
		return &resetDatabase{
			timeout:   fSet.Int(timeoutFlag, defaults.timeout, timeoutUsage),
			service:   fSet.String(serviceFlag, defaults.service, serviceUsage),
			metadata:  fSet.String(metadataFlag, defaults.metadata, metadataUsage),
			extension: fSet.String(extensionFlag, defaults.extension, extensionUsage),
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
	timeout   *int
	service   *string
	metadata  *string
	extension *string
	f         *flag.FlagSet
}

func (r *resetDatabase) Exec(o internal.OutputBus, args []string) (ok bool) {
	if internal.ProcessArgs(o, r.f, args) {
		if Dirty() {
			ok = r.runCommand(o, func() (serviceGateway, error) {
				m, err := mgr.Connect()
				if err != nil {
					return nil, err
				}
				return &sysMgr{m: m}, err
			})
			if ok {
				ClearDirty(o)
			}
		} else {
			o.WriteCanonicalConsole("Running %q is not necessary, as no track files have been edited", resetDatabaseCommandName)
			ok = true // no harm, no foul
		}
	}
	return
}

func (r *resetDatabase) runCommand(o internal.OutputBus, connect func() (serviceGateway, error)) (ok bool) {
	o.Log(internal.Info, internal.LogInfoExecutingCommand, map[string]any{
		fieldKeyCommandName:   resetDatabaseCommandName,
		fieldKeyServiceFlag:   *r.service,
		fieldKeyTimeoutFlag:   *r.timeout,
		fieldKeyMetadataFlag:  *r.metadata,
		fieldKeyExtensionFlag: *r.extension,
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
	o.WriteCanonicalConsole("No metadata files were found in %q", *r.metadata)
	o.Log(internal.Info, internal.LogInfoNoFilesFound, map[string]any{
		internal.FieldKeyDirectory: *r.metadata,
		fieldKeyFileExtension:      *r.extension,
	})
	return true
}

func (r *resetDatabase) deleteMetadataFiles(o internal.OutputBus, paths []string) bool {
	var count int
	for _, path := range paths {
		if err := os.Remove(path); err != nil {
			o.WriteCanonicalError(internal.UserCannotDeleteFile, path, err)
			o.Log(internal.Error, internal.LogErrorCannotDeleteFile, map[string]any{
				internal.FieldKeyFileName: path,
				internal.FieldKeyError:    err,
			})
		} else {
			count++
		}
	}
	o.WriteCanonicalConsole("%d out of %d metadata files have been deleted from %q", count, len(paths), *r.metadata)
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
		o.WriteCanonicalError(internal.UserCannotQueryService, *r.service, err)
		o.Log(internal.Error, internal.LogErrorServiceIssue, map[string]any{
			internal.FieldKeyError: err,
			fieldKeyService:        *r.service,
			fieldKeyOperation:      operationQueryService,
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
		o.WriteCanonicalError(internal.UserCannotStopService, *r.service, err)
		o.Log(internal.Error, internal.LogErrorServiceIssue, map[string]any{
			internal.FieldKeyError: err,
			fieldKeyService:        *r.service,
			fieldKeyOperation:      operationStopService,
		})
	}
	return ok
}

func (r *resetDatabase) logServiceStopped(o internal.OutputBus) {
	o.Log(internal.Info, internal.LogInfoServiceStatus, map[string]any{
		fieldKeyService:       *r.service,
		fieldKeyServiceStatus: statusStopped,
	})
}

func (r *resetDatabase) openService(o internal.OutputBus, connect func() (serviceGateway, error)) (sM serviceGateway, s service) {
	sM, err := connect()
	if err != nil {
		o.WriteCanonicalError(internal.UserServiceMgrConnectionFailed, err)
		o.Log(internal.Error, internal.LogErrorServiceManagerIssue, map[string]any{
			internal.FieldKeyError: err,
			fieldKeyOperation:      operationConnect,
		})
	} else {
		s, err = sM.openService(*r.service)
		if err != nil {
			o.WriteCanonicalError(internal.UserCannotOpenService, *r.service, err)
			o.Log(internal.Error, internal.LogErrorServiceIssue, map[string]any{
				internal.FieldKeyError: err,
				fieldKeyService:        *r.service,
				fieldKeyOperation:      operationOpenService,
			})
			services, err := sM.manager().ListServices()
			if err != nil {
				o.WriteCanonicalError(internal.UserCannotListServices, err)
				o.Log(internal.Error, internal.LogErrorServiceManagerIssue, map[string]any{
					internal.FieldKeyError: err,
					fieldKeyOperation:      operationListServices,
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
			o.WriteCanonicalError(internal.UserServiceStopTimedOut, *r.service, *r.timeout)
			o.Log(internal.Error, internal.LogErrorServiceIssue, map[string]any{
				fieldKeyService:        *r.service,
				fieldKeyTimeout:        *r.timeout,
				fieldKeyOperation:      operationStopService,
				internal.FieldKeyError: errTimeout,
			})
			break
		}
		time.Sleep(checkFreq)
		status, err := s.Query()
		if err != nil {
			o.WriteCanonicalError(internal.UserCannotQueryService, *r.service, err)
			o.Log(internal.Error, internal.LogErrorServiceIssue, map[string]any{
				internal.FieldKeyError: err,
				fieldKeyService:        *r.service,
				fieldKeyOperation:      operationQueryService,
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
	o.WriteConsole("The following services are available:\n")
	if len(services) == 0 {
		o.WriteConsole("  - none -\n")
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
		o.WriteConsole("  State %q:\n", state)
		for _, service := range sMap[state] {
			o.WriteConsole("    %q\n", service)
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
