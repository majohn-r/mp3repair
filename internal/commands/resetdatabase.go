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

	"github.com/majohn-r/output"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

func init() {
	addCommandData(resetDatabaseCommandName, commandData{isDefault: false, init: newResetDatabase})
	defaultMetadataPath = filepath.Join("%USERPROFILE%", "AppData", "Local", "Microsoft", "Media Player")
	addDefaultMapping(resetDatabaseCommandName, map[string]any{
		extensionFlag: defaultExtension,
		metadataFlag:  defaultMetadataPath,
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
)

var (
	defaultMetadataPath string
	timeoutError        = fmt.Errorf("operation timed out")
	stateToStatus       = map[svc.State]string{
		svc.Stopped:         "stopped",
		svc.StartPending:    "start pending",
		svc.StopPending:     "stop pending",
		svc.Running:         "running",
		svc.ContinuePending: "continue pending",
		svc.PausePending:    "pause pending",
		svc.Paused:          "paused",
	}
)

func newResetDatabase(o output.Bus, c *internal.Configuration, fSet *flag.FlagSet) (CommandProcessor, bool) {
	return newResetDatabaseCommand(o, c, fSet)
}

type resetDatabaseDefaults struct {
	timeout   int
	service   string
	metadata  string
	extension string
}

func newResetDatabaseCommand(o output.Bus, c *internal.Configuration, fSet *flag.FlagSet) (*resetDatabase, bool) {
	defaults, defaultsOk := evaluateResetDatabaseDefaults(o, c.SubConfiguration(resetDatabaseCommandName))
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

func evaluateResetDatabaseDefaults(o output.Bus, c *internal.Configuration) (defaults resetDatabaseDefaults, ok bool) {
	defaults = resetDatabaseDefaults{}
	ok = true
	var err error
	if defaults.timeout, err = c.IntDefault(timeoutFlag, internal.NewIntBounds(minTimeout, defaultTimeout, maxTimeout)); err != nil {
		reportBadDefault(o, resetDatabaseCommandName, err)
		ok = false
	}
	if defaults.service, err = c.StringDefault(serviceFlag, defaultService); err != nil {
		reportBadDefault(o, resetDatabaseCommandName, err)
		ok = false
	}
	if defaults.metadata, err = c.StringDefault(metadataFlag, defaultMetadataPath); err != nil {
		reportBadDefault(o, resetDatabaseCommandName, err)
		ok = false
	}
	if defaults.extension, err = c.StringDefault(extensionFlag, defaultExtension); err != nil {
		reportBadDefault(o, resetDatabaseCommandName, err)
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

func (r *resetDatabase) Exec(o output.Bus, args []string) (ok bool) {
	if internal.ProcessArgs(o, r.f, args) {
		if dirty() {
			if ok = r.runCommand(o, func() (serviceGateway, error) {
				m, err := mgr.Connect()
				if err != nil {
					return nil, err
				}
				return &sysMgr{m: m}, err
			}); ok {
				clearDirty(o)
			}
		} else {
			o.WriteCanonicalConsole("Running %q is not necessary, as no track files have been edited", resetDatabaseCommandName)
			ok = true // no harm, no foul
		}
	}
	return
}

func (r *resetDatabase) runCommand(o output.Bus, connect func() (serviceGateway, error)) (ok bool) {
	logStart(o, resetDatabaseCommandName, map[string]any{
		"-" + serviceFlag:   *r.service,
		"-" + timeoutFlag:   *r.timeout,
		"-" + metadataFlag:  *r.metadata,
		"-" + extensionFlag: *r.extension,
	})
	if !r.stop(o, connect) {
		return
	}
	return r.deleteMetadata(o)
}

func (r *resetDatabase) deleteMetadata(o output.Bus) bool {
	if files, ok := internal.ReadDirectory(o, *r.metadata); !ok {
		return false
	} else {
		pathsToDelete := r.filterMetadataFiles(files)
		if len(pathsToDelete) > 0 {
			return r.deleteMetadataFiles(o, pathsToDelete)
		}
		o.WriteCanonicalConsole("No metadata files were found in %q", *r.metadata)
		o.Log(output.Info, "no files found", map[string]any{"directory": *r.metadata, "extension": *r.extension})
		return true
	}
}

func (r *resetDatabase) deleteMetadataFiles(o output.Bus, paths []string) bool {
	var count int
	for _, path := range paths {
		if err := os.Remove(path); err != nil {
			reportFileDeletionFailure(o, path, err)
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
func (r *resetDatabase) stop(o output.Bus, connect func() (serviceGateway, error)) bool {
	// this is a privileged operation and fails if the user is not an administrator
	if sM, s := r.open(o, connect); s == nil {
		// something unhappy happened, but, fine, we're done and we're not preventing progress
		return true
	} else {
		defer func() {
			_ = sM.manager().Disconnect()
			_ = s.Close()
		}()
		if status, err := s.Query(); err != nil {
			r.reportServiceQueryIssue(o, err)
			return true
		} else {
			if status.State == svc.Stopped {
				r.reportServiceStopped(o)
				return true
			}
			ok := status.State != svc.Running
			if status, err = s.Control(svc.Stop); err == nil {
				timeout := time.Now().Add(time.Duration(*r.timeout) * time.Second)
				if stopped := r.waitForStop(o, s, status, timeout, 100*time.Millisecond); stopped {
					ok = true
				}
			} else {
				o.WriteCanonicalError("The service %q cannot be stopped: %v", *r.service, err)
				logServiceIssue(o, r.makeServiceErrorFields("stop service", err))
			}
			return ok
		}
	}
}

func (r *resetDatabase) reportServiceQueryIssue(o output.Bus, e error) {
	o.WriteCanonicalError("The status for the service %q cannot be obtained: %v", *r.service, e)
	logServiceIssue(o, r.makeServiceErrorFields("query service status", e))
}

func logServiceIssue(o output.Bus, fields map[string]any) {
	o.Log(output.Error, "service issue", fields)
}

func (r *resetDatabase) makeServiceErrorFields(s string, e error) map[string]any {
	return map[string]any{
		"error":     e,
		"service":   *r.service,
		"operation": s,
	}
}

func (r *resetDatabase) reportServiceStopped(o output.Bus) {
	o.Log(output.Info, "service status", map[string]any{
		"service": *r.service,
		"status":  "stopped",
	})
}

func (r *resetDatabase) open(o output.Bus, connect func() (serviceGateway, error)) (sM serviceGateway, s service) {
	var err error
	if sM, err = connect(); err != nil {
		o.WriteCanonicalError("The service manager cannot be accessed. Try running the program again as an administrator. Error: %v", err)
		logServiceManagerIssue(o, "connect to service manager", err)
	} else {
		if s, err = sM.open(*r.service); err != nil {
			o.WriteCanonicalError("The service %q cannot be opened: %v", *r.service, err)
			logServiceIssue(o, r.makeServiceErrorFields("open service", err))
			if services, err := sM.manager().ListServices(); err != nil {
				o.WriteCanonicalError("The list of available services cannot be obtained: %v", err)
				logServiceManagerIssue(o, "list services", err)
			} else {
				listServices(o, sM, services)
			}
			_ = sM.manager().Disconnect()
			sM = nil
			s = nil
		}
	}
	return
}

func logServiceManagerIssue(o output.Bus, operation string, e error) {
	o.Log(output.Error, "service manager issue", map[string]any{
		"error":     e,
		"operation": operation,
	})
}

func (r *resetDatabase) waitForStop(o output.Bus, s service, status svc.Status, timeout time.Time, checkFreq time.Duration) (ok bool) {
	if status.State == svc.Stopped {
		r.reportServiceStopped(o)
		ok = true
		return
	}
	for !ok {
		if timeout.Before(time.Now()) {
			o.WriteCanonicalError("The service %q could not be stopped within the %d second timeout", *r.service, *r.timeout)
			m := r.makeServiceErrorFields("stop service", timeoutError)
			m["timeout in seconds"] = *r.timeout
			logServiceIssue(o, m)
			break
		}
		time.Sleep(checkFreq)
		if status, err := s.Query(); err != nil {
			r.reportServiceQueryIssue(o, err)
			break
		} else if status.State == svc.Stopped {
			r.reportServiceStopped(o)
			ok = true
		}
	}
	return
}

func listServices(o output.Bus, sM serviceGateway, services []string) {
	o.WriteConsole("The following services are available:\n")
	if len(services) == 0 {
		o.WriteConsole("  - none -\n")
		return
	}
	sort.Strings(services)
	m := map[string][]string{}
	for _, svc := range services {
		if s, err := sM.open(svc); err == nil {
			if status, err := s.Query(); err == nil {
				key := stateToStatus[status.State]
				m[key] = append(m[key], svc)
			} else {
				e := err.Error()
				m[e] = append(m[e], svc)
			}
			s.Close()
		} else {
			e := err.Error()
			m[e] = append(m[e], svc)
		}
	}
	var states []string
	for k := range m {
		states = append(states, k)
	}
	sort.Strings(states)
	for _, state := range states {
		o.WriteConsole("  State %q:\n", state)
		for _, svc := range m[state] {
			o.WriteConsole("    %q\n", svc)
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
	open(string) (service, error)
	manager() manager
}

type sysMgr struct {
	m *mgr.Mgr
}

func (s *sysMgr) open(name string) (service, error) {
	return s.m.OpenService(name)
}

func (s *sysMgr) manager() manager {
	return s.m
}
