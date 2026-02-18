/*
Copyright © 2026 Marc Johnson (marc.johnson27591@gmail.com)
*/
package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	cmdtoolkit "github.com/majohn-r/cmd-toolkit"
	"github.com/majohn-r/output"
	"github.com/spf13/cobra"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

func Test_processResetDatabaseFlags(t *testing.T) {
	tests := map[string]struct {
		values map[string]*cmdtoolkit.CommandFlag[any]
		want   *resetDatabaseSettings
		want1  bool
		output.WantedRecording
	}{
		"massive errors": {
			values: nil,
			want:   &resetDatabaseSettings{},
			want1:  false,
			WantedRecording: output.WantedRecording{
				Error: "" +
					"An internal error occurred: no flag values exist.\n" +
					"An internal error occurred: no flag values exist.\n" +
					"An internal error occurred: no flag values exist.\n",
				Log: "" +
					"level='error'" +
					" error='no results to extract flag values from'" +
					" msg='internal error'\n" +
					"level='error'" +
					" error='no results to extract flag values from'" +
					" msg='internal error'\n" +
					"level='error'" +
					" error='no results to extract flag values from'" +
					" msg='internal error'\n",
			},
		},
		"good results": {
			values: map[string]*cmdtoolkit.CommandFlag[any]{
				"extension":           {Value: ".foo"},
				"force":               {Value: true},
				"ignoreServiceErrors": {Value: true},
				"metadataDir":         {Value: "metadata"},
				"service":             {Value: "music service"},
				"timeout":             {Value: 5},
			},
			want: &resetDatabaseSettings{
				force:               cmdtoolkit.CommandFlag[bool]{Value: true},
				ignoreServiceErrors: cmdtoolkit.CommandFlag[bool]{Value: true},
				timeout:             cmdtoolkit.CommandFlag[int]{Value: 5},
			},
			want1: true,
		},
		"timeout too low": {
			values: map[string]*cmdtoolkit.CommandFlag[any]{
				"extension":           {Value: ".foo"},
				"force":               {Value: true},
				"ignoreServiceErrors": {Value: true},
				"metadataDir":         {Value: "metadata"},
				"service":             {Value: "music service"},
				"timeout":             {Value: minTimeout - 1},
			},
			want: &resetDatabaseSettings{
				force:               cmdtoolkit.CommandFlag[bool]{Value: true},
				ignoreServiceErrors: cmdtoolkit.CommandFlag[bool]{Value: true},
				timeout:             cmdtoolkit.CommandFlag[int]{Value: minTimeout},
			},
			want1: true,
			WantedRecording: output.WantedRecording{
				Log: "level='warning'" +
					" flag='timeout'" +
					" providedValue='0'" +
					" replacedBy='1'" +
					" msg='user-supplied value replaced'\n",
			},
		},
		"timeout too high": {
			values: map[string]*cmdtoolkit.CommandFlag[any]{
				"extension":           {Value: ".foo"},
				"force":               {Value: true},
				"ignoreServiceErrors": {Value: true},
				"metadataDir":         {Value: "metadata"},
				"service":             {Value: "music service"},
				"timeout":             {Value: maxTimeout + 1},
			},
			want: &resetDatabaseSettings{
				force:               cmdtoolkit.CommandFlag[bool]{Value: true},
				ignoreServiceErrors: cmdtoolkit.CommandFlag[bool]{Value: true},
				timeout:             cmdtoolkit.CommandFlag[int]{Value: maxTimeout},
			},
			want1: true,
			WantedRecording: output.WantedRecording{
				Log: "level='warning'" +
					" flag='timeout'" +
					" providedValue='61'" +
					" replacedBy='60'" +
					" msg='user-supplied value replaced'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			got, got1 := processResetDatabaseFlags(o, tt.values)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("processResetDatabaseFlags() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("processResetDatabaseFlags() got1 = %v, want %v", got1, tt.want1)
			}
			o.Report(t, "processResetDatabaseFlags()", tt.WantedRecording)
		})
	}
}

type testService struct {
	queries  int
	statuses []svc.Status
}

func (ts *testService) Close() error {
	return nil
}

func (ts *testService) Query() (svc.Status, error) {
	if ts.queries >= len(ts.statuses) {
		return svc.Status{}, fmt.Errorf("no results from query")
	}
	status := ts.statuses[ts.queries]
	ts.queries++
	return status, nil
}

func (ts *testService) Control(_ svc.Cmd) (svc.Status, error) {
	return ts.Query()
}

func newTestService(values ...svc.Status) *testService {
	ts := &testService{
		queries:  0,
		statuses: values,
	}
	return ts
}

func Test_resetDatabaseSettings_waitForStop(t *testing.T) {
	type args struct {
		s             serviceRep
		expiration    time.Time
		checkInterval time.Duration
	}
	tests := map[string]struct {
		resetDatabaseSettings *resetDatabaseSettings
		args
		wantOk     bool
		wantStatus *cmdtoolkit.ExitError
		output.WantedRecording
	}{
		"already timed out": {
			resetDatabaseSettings: &resetDatabaseSettings{
				timeout: cmdtoolkit.CommandFlag[int]{Value: 10},
			},
			args:       args{expiration: time.Now().Add(time.Duration(-1) * time.Second)},
			wantStatus: cmdtoolkit.NewExitSystemError("resetDatabase"),
			WantedRecording: output.WantedRecording{
				Error: "The service \"WMPNetworkSVC\" could not be stopped within the 10 second timeout.\n",
				Log: "" +
					"level='error'" +
					" error='timed out'" +
					" service='WMPNetworkSVC'" +
					" timeout='10'" +
					" trigger='Stop'" +
					" msg='service problem'\n",
			},
		},
		"query error": {
			resetDatabaseSettings: &resetDatabaseSettings{
				timeout: cmdtoolkit.CommandFlag[int]{Value: 10},
			},
			args: args{
				s:             newTestService(),
				expiration:    time.Now().Add(time.Duration(1) * time.Second),
				checkInterval: 1 * time.Millisecond,
			},
			wantStatus: cmdtoolkit.NewExitSystemError("resetDatabase"),
			WantedRecording: output.WantedRecording{
				Error: "" +
					"An error occurred while attempting to stop the service \"WMPNetworkSVC\": " +
					"'no results from query'.\n",
				Log: "" +
					"level='error'" +
					" error='no results from query'" +
					" service='WMPNetworkSVC'" +
					" msg='service query error'\n",
			},
		},
		"stops correctly": {
			resetDatabaseSettings: &resetDatabaseSettings{
				timeout: cmdtoolkit.CommandFlag[int]{Value: 10},
			},
			args: args{
				s: newTestService(
					svc.Status{State: svc.Running},
					svc.Status{State: svc.Running},
					svc.Status{State: svc.Stopped}),
				expiration:    time.Now().Add(time.Duration(1) * time.Second),
				checkInterval: 1 * time.Millisecond,
			},
			wantOk:     true,
			wantStatus: nil,
			WantedRecording: output.WantedRecording{
				Log: "" +
					"level='info'" +
					" service='WMPNetworkSVC'" +
					" msg='service stopped'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			gotOk, gotStatus := tt.resetDatabaseSettings.waitForStop(
				o,
				tt.args.s,
				tt.args.expiration,
				tt.args.checkInterval,
			)
			if gotOk != tt.wantOk {
				t.Errorf("resetDatabaseSettings.waitForStop() = %t, want %t", gotOk, tt.wantOk)
			}
			if !compareExitErrors(gotStatus, tt.wantStatus) {
				t.Errorf("resetDatabaseSettings.waitForStop() = %s, want %s", gotStatus, tt.wantStatus)
			}
			o.Report(t, "resetDatabaseSettings.waitForStop()", tt.WantedRecording)
		})
	}
}

type testManager struct {
	serviceMap  map[string]*mgr.Service
	serviceList []string
}

func (tm *testManager) Disconnect() error {
	return nil
}

func (tm *testManager) OpenService(name string) (*mgr.Service, error) {
	if service, found := tm.serviceMap[name]; found {
		return service, nil
	}
	return nil, fmt.Errorf("no such service")
}

func (tm *testManager) ListServices() ([]string, error) {
	if len(tm.serviceList) == 0 {
		return nil, fmt.Errorf("no services")
	}
	return tm.serviceList, nil
}

func newTestManager(m map[string]*mgr.Service, list []string) *testManager {
	return &testManager{
		serviceMap:  m,
		serviceList: list,
	}
}

func Test_resetDatabaseSettings_stopFoundService(t *testing.T) {
	type args struct {
		manager serviceManager
		service serviceRep
	}
	tests := map[string]struct {
		resetDatabaseSettings *resetDatabaseSettings
		args
		wantOk     bool
		wantStatus *cmdtoolkit.ExitError
		output.WantedRecording
	}{
		"defective service": {
			resetDatabaseSettings: &resetDatabaseSettings{
				timeout: cmdtoolkit.CommandFlag[int]{Value: 10},
			},
			args: args{
				manager: newTestManager(nil, nil),
				service: newTestService(),
			},
			wantStatus: cmdtoolkit.NewExitSystemError("resetDatabase"),
			WantedRecording: output.WantedRecording{
				Error: "An error occurred while trying to stop service \"WMPNetworkSVC\": 'no results from query'.\n",
				Log: "" +
					"level='error' " +
					"error='no results from query' " +
					"service='WMPNetworkSVC' " +
					"msg='service query error'\n",
			},
		},
		"already stopped": {
			resetDatabaseSettings: &resetDatabaseSettings{
				timeout: cmdtoolkit.CommandFlag[int]{Value: 10},
			},
			args: args{
				manager: newTestManager(nil, nil),
				service: newTestService(svc.Status{State: svc.Stopped}),
			},
			wantOk:     true,
			wantStatus: nil,
			WantedRecording: output.WantedRecording{
				Log: "" +
					"level='info'" +
					" service='WMPNetworkSVC'" +
					" msg='service stopped'\n",
			},
		},
		"stopped easily": {
			resetDatabaseSettings: &resetDatabaseSettings{
				timeout: cmdtoolkit.CommandFlag[int]{Value: 10},
			},
			args: args{
				manager: newTestManager(nil, nil),
				service: newTestService(
					svc.Status{State: svc.Paused},
					svc.Status{State: svc.Stopped}),
			},
			wantOk:     true,
			wantStatus: nil,
			WantedRecording: output.WantedRecording{
				Log: "" +
					"level='info'" +
					" service='WMPNetworkSVC'" +
					" msg='service stopped'\n",
			},
		},
		"stopped with a little more difficulty": {
			resetDatabaseSettings: &resetDatabaseSettings{
				timeout: cmdtoolkit.CommandFlag[int]{Value: 10},
			},
			args: args{
				manager: newTestManager(nil, nil),
				service: newTestService(
					svc.Status{State: svc.Paused},
					svc.Status{State: svc.Paused},
					svc.Status{State: svc.Stopped}),
			},
			wantOk:     true,
			wantStatus: nil,
			WantedRecording: output.WantedRecording{
				Log: "" +
					"level='info'" +
					" service='WMPNetworkSVC'" +
					" msg='service stopped'\n",
			},
		},
		"cannot be stopped": {
			resetDatabaseSettings: &resetDatabaseSettings{
				timeout: cmdtoolkit.CommandFlag[int]{Value: 10},
			},
			args: args{
				manager: newTestManager(nil, nil),
				service: newTestService(svc.Status{State: svc.Paused}),
			},
			wantStatus: cmdtoolkit.NewExitSystemError("resetDatabase"),
			WantedRecording: output.WantedRecording{
				Error: "The service \"WMPNetworkSVC\" cannot be stopped: 'no results from query'.\n",
				Log: "" +
					"level='error'" +
					" error='no results from query'" +
					" service='WMPNetworkSVC'" +
					" trigger='Stop'" +
					" msg='service problem'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			gotOk, gotStatus := tt.resetDatabaseSettings.stopFoundService(o, tt.args.manager,
				tt.args.service)
			if gotOk != tt.wantOk {
				t.Errorf("resetDatabaseSettings.stopFoundService() = %v, want %v", gotOk, tt.wantOk)
			}
			if !compareExitErrors(gotStatus, tt.wantStatus) {
				t.Errorf("resetDatabaseSettings.stopFoundService() = %v, want %v", gotStatus, tt.wantStatus)
			}
			o.Report(t, "resetDatabaseSettings.stopFoundService()", tt.WantedRecording)
		})
	}
}

func Test_addServiceState(t *testing.T) {
	tests := map[string]struct {
		m           map[string][]string
		s           serviceRep
		serviceName string
		want        map[string][]string
	}{
		"error": {
			m: map[string][]string{"no results from query": {
				"some other bad service",
			}},
			s:           newTestService(),
			serviceName: "bad service",
			want: map[string][]string{
				"no results from query": {"some other bad service", "bad service"},
			},
		},
		"success": {
			m:           map[string][]string{"stopped": {"some other service"}},
			s:           newTestService(svc.Status{State: svc.Stopped}),
			serviceName: "happy service",
			want: map[string][]string{
				"stopped": {"some other service", "happy service"},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			addServiceState(tt.m, tt.s, tt.serviceName)
			if !reflect.DeepEqual(tt.m, tt.want) {
				t.Errorf("addServiceState() = %v, want %v", tt.m, tt.want)
			}
		})
	}
}

func Test_listAvailableServices(t *testing.T) {
	type args struct {
		manager  serviceManager
		services []string
	}
	tests := map[string]struct {
		args
		output.WantedRecording
	}{
		"no services": {
			args: args{},
			WantedRecording: output.WantedRecording{
				Error: "" +
					"The following services are available:\n" +
					"  - none -\n",
			},
		},
		"some services": {
			args: args{
				manager: newTestManager(map[string]*mgr.Service{
					"service1": nil,
					"service2": nil,
				}, nil),
				services: []string{"service2", "service1", "service4"},
			},
			WantedRecording: output.WantedRecording{
				Error: "" +
					"The following services are available:\n" +
					"  State \"no service\":\n" +
					"    \"service1\"\n" +
					"    \"service2\"\n" +
					"  State \"no such service\":\n" +
					"    \"service4\"\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			listAvailableServices(o, tt.args.manager, tt.args.services)
			o.Report(t, "listAvailableServices()", tt.WantedRecording)
		})
	}
}

func Test_resetDatabaseSettings_disableService(t *testing.T) {
	tests := map[string]struct {
		resetDatabaseSettings *resetDatabaseSettings
		manager               serviceManager
		wantOk                bool
		wantStatus            *cmdtoolkit.ExitError
		output.WantedRecording
	}{
		"defective manager #1": {
			resetDatabaseSettings: &resetDatabaseSettings{
				timeout: cmdtoolkit.CommandFlag[int]{Value: 1},
			},
			manager:    newTestManager(nil, []string{"WMPNetworkSVC"}),
			wantStatus: cmdtoolkit.NewExitSystemError("resetDatabase"),
			WantedRecording: output.WantedRecording{
				Error: "" +
					"The service \"WMPNetworkSVC\" cannot be opened: 'no such service'.\n" +
					"The following services are available:\n" +
					"  State \"no such service\":\n" +
					"    \"WMPNetworkSVC\"\n",
				Log: "" +
					"level='error'" +
					" error='no such service'" +
					" service='WMPNetworkSVC'" +
					" trigger='OpenService'" +
					" msg='service problem'\n",
			},
		},
		"defective manager #2": {
			resetDatabaseSettings: &resetDatabaseSettings{
				timeout: cmdtoolkit.CommandFlag[int]{Value: 1},
			},
			manager:    newTestManager(nil, nil),
			wantStatus: cmdtoolkit.NewExitSystemError("resetDatabase"),
			WantedRecording: output.WantedRecording{
				Error: "The service \"WMPNetworkSVC\" cannot be opened: 'no such service'.\n",
				Log: "" +
					"level='error'" +
					" error='no such service'" +
					" service='WMPNetworkSVC'" +
					" trigger='OpenService'" +
					" msg='service problem'\n" +
					"level='error'" +
					" error='no services'" +
					" trigger='ListServices'" +
					" msg='service problem'\n",
			},
		},
		"defective manager #3": {
			resetDatabaseSettings: &resetDatabaseSettings{
				timeout: cmdtoolkit.CommandFlag[int]{Value: 1},
			},
			manager:    newTestManager(map[string]*mgr.Service{"WMPNetworkSVC": nil}, nil),
			wantStatus: cmdtoolkit.NewExitSystemError("resetDatabase"),
			WantedRecording: output.WantedRecording{
				Error: "An error occurred while trying to stop service \"WMPNetworkSVC\": 'no service'.\n",
				Log: "" +
					"level='error'" +
					" error='no service'" +
					" service='WMPNetworkSVC'" +
					" msg='service query error'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			gotOk, gotStatus := tt.resetDatabaseSettings.disableService(o, tt.manager)
			if gotOk != tt.wantOk {
				t.Errorf("resetDatabaseSettings.disableService() = %v, want %v", gotOk, tt.wantOk)
			}
			if !compareExitErrors(gotStatus, tt.wantStatus) {
				t.Errorf("resetDatabaseSettings.disableService() = %v, want %v", gotStatus, tt.wantStatus)
			}
			o.Report(t, "resetDatabaseSettings.disableService()", tt.WantedRecording)
		})
	}
}

func Test_resetDatabaseSettings_stopService(t *testing.T) {
	originalConnect := connect
	originalProcessIsElevated := processIsElevated
	defer func() {
		connect = originalConnect
		processIsElevated = originalProcessIsElevated
	}()
	processIsElevated = func() bool { return false }
	tests := map[string]struct {
		connect               func() (*mgr.Mgr, error)
		resetDatabaseSettings *resetDatabaseSettings
		wantOk                bool
		wantStatus            *cmdtoolkit.ExitError
		output.WantedRecording
	}{
		"connect fails": {
			connect: func() (*mgr.Mgr, error) {
				return nil, fmt.Errorf("no manager available")
			},
			resetDatabaseSettings: &resetDatabaseSettings{
				timeout: cmdtoolkit.CommandFlag[int]{Value: 1},
			},
			wantStatus: cmdtoolkit.NewExitSystemError("resetDatabase"),
			WantedRecording: output.WantedRecording{
				Error: "" +
					"An attempt to connect with the service manager failed;" +
					" error is 'no manager available'.\n" +
					"Why?\n" +
					"This failure is likely to be due to lack of permissions.\n" +
					"What to do:\n" +
					"If you can, try running this command as an administrator.\n",
				Log: "" +
					"level='error'" +
					" error='no manager available'" +
					" msg='service manager connect failed'\n",
			},
		},
		"connect sort of works": {
			connect: func() (*mgr.Mgr, error) {
				return nil, nil
			},
			resetDatabaseSettings: &resetDatabaseSettings{
				timeout: cmdtoolkit.CommandFlag[int]{Value: 1},
			},
			wantStatus: cmdtoolkit.NewExitSystemError("resetDatabase"),
			WantedRecording: output.WantedRecording{
				Error: "The service \"WMPNetworkSVC\" cannot be opened: 'nil manager'.\n",
				Log: "" +
					"level='error'" +
					" error='nil manager'" +
					" service='WMPNetworkSVC'" +
					" trigger='OpenService'" +
					" msg='service problem'\n" +
					"level='error'" +
					" error='nil manager'" +
					" trigger='ListServices'" +
					" msg='service problem'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			connect = tt.connect
			o := output.NewRecorder()
			gotOk, gotStatus := tt.resetDatabaseSettings.stopService(o)
			if gotOk != tt.wantOk {
				t.Errorf("resetDatabaseSettings.stopService() = %v, want %v", gotOk, tt.wantOk)
			}
			if !compareExitErrors(gotStatus, tt.wantStatus) {
				t.Errorf("resetDatabaseSettings.stopService() = %v, want %v", gotStatus, tt.wantStatus)
			}
			o.Report(t, "resetDatabaseSettings.stopService()", tt.WantedRecording)
		})
	}
}

func Test_deleteMetadataFiles(t *testing.T) {
	originalRemove := remove
	originalUserProfile := cmdtoolkit.NewEnvVarMemento("USERPROFILE")
	_ = os.Setenv("USERPROFILE", "dummyProfile")
	defer func() {
		remove = originalRemove
		originalUserProfile.Restore()
	}()
	tests := map[string]struct {
		remove func(string) error
		paths  []string
		want   *cmdtoolkit.ExitError
		output.WantedRecording
	}{
		"no files": {
			paths: nil,
			want:  nil,
		},
		"locked files": {
			remove: func(_ string) error { return fmt.Errorf("cannot remove file") },
			paths:  []string{"file1", "file2"},
			want:   cmdtoolkit.NewExitSystemError("resetDatabase"),
			WantedRecording: output.WantedRecording{
				Console: "0 out of 2 metadata files have been deleted from" +
					" \"dummyProfile\\\\AppData\\\\Local\\\\Microsoft\\\\Media Player\".\n",
				Log: "" +
					"level='error'" +
					" error='cannot remove file'" +
					" fileName='file1'" +
					" msg='cannot delete file'\n" +
					"level='error'" +
					" error='cannot remove file'" +
					" fileName='file2'" +
					" msg='cannot delete file'\n",
			},
		},
		"deletable files": {
			remove: func(_ string) error { return nil },
			paths:  []string{"file1", "file2"},
			want:   nil,
			WantedRecording: output.WantedRecording{
				Console: "2 out of 2 metadata files have been deleted from" +
					" \"dummyProfile\\\\AppData\\\\Local\\\\Microsoft\\\\Media Player\".\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			remove = tt.remove
			o := output.NewRecorder()
			if got := deleteMetadataFiles(o, tt.paths); !compareExitErrors(got, tt.want) {
				t.Errorf("deleteMetadataFiles() %s want %s", got, tt.want)
			}
			o.Report(t, "deleteMetadataFiles()", tt.WantedRecording)
		})
	}
}

func Test_filterMetadataFiles(t *testing.T) {
	originalPlainFileExists := plainFileExists
	originalUserProfile := cmdtoolkit.NewEnvVarMemento("USERPROFILE")
	_ = os.Setenv("USERPROFILE", "dummyProfile")
	defer func() {
		plainFileExists = originalPlainFileExists
		originalUserProfile.Restore()
	}()
	tests := map[string]struct {
		plainFileExists func(string) bool
		entries         []fs.FileInfo
		want            []string
	}{
		"no entries": {want: []string{}},
		"mixed entries": {
			plainFileExists: func(s string) bool { return !strings.Contains(s, "dir.") },
			entries: []fs.FileInfo{
				newTestFile("dir. foo.wmdb", nil),
				newTestFile("foo.wmdb", nil),
				newTestFile("foo", nil),
			},
			want: []string{
				filepath.Join("dummyProfile", "AppData", "Local", "Microsoft", "Media Player", "foo.wmdb"),
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			plainFileExists = tt.plainFileExists
			if got := filterMetadataFiles(tt.entries); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("filterMetadataFiles() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_resetDatabaseSettings_cleanUpMetadata(t *testing.T) {
	originalReadDirectory := readDirectory
	originalPlainFileExists := plainFileExists
	originalRemove := remove
	originalUserProfile := cmdtoolkit.NewEnvVarMemento("USERPROFILE")
	_ = os.Setenv("USERPROFILE", "dummyProfile")
	defer func() {
		readDirectory = originalReadDirectory
		plainFileExists = originalPlainFileExists
		remove = originalRemove
		originalUserProfile.Restore()
	}()
	tests := map[string]struct {
		readDirectory         func(output.Bus, string) ([]fs.FileInfo, bool)
		plainFileExists       func(string) bool
		remove                func(string) error
		resetDatabaseSettings *resetDatabaseSettings
		stopped               bool
		want                  *cmdtoolkit.ExitError
		output.WantedRecording
	}{
		"did not stop, cannot ignore it": {
			resetDatabaseSettings: &resetDatabaseSettings{
				ignoreServiceErrors: cmdtoolkit.CommandFlag[bool]{Value: false},
			},
			stopped: false,
			want:    cmdtoolkit.NewExitUserError("resetDatabase"),
			WantedRecording: output.WantedRecording{
				Error: "" +
					"Metadata files will not be deleted.\n" +
					"Why?\n" +
					"The Windows Media Player sharing service \"WMPNetworkSVC\" could not be stopped, and" +
					" \"--ignoreServiceErrors\" is false.\n" +
					"What to do:\n" +
					"Rerun this command with \"--ignoreServiceErrors\" set to true.\n",
			},
		},
		"stopped, no metadata": {
			readDirectory: func(_ output.Bus, _ string) ([]fs.FileInfo, bool) {
				return nil, true
			},
			resetDatabaseSettings: &resetDatabaseSettings{},
			stopped:               true,
			want:                  nil,
			WantedRecording: output.WantedRecording{
				Console: "No metadata files were found in " +
					"\"dummyProfile\\\\AppData\\\\Local\\\\Microsoft\\\\Media Player\".\n",
				Log: "" +
					"level='info'" +
					" directory='dummyProfile\\AppData\\Local\\Microsoft\\Media Player'" +
					" extension='.wmdb'" +
					" msg='no files found'\n",
			},
		},
		"not stopped but ignored, no metadata": {
			readDirectory: func(_ output.Bus, _ string) ([]fs.FileInfo, bool) {
				return nil, true
			},
			resetDatabaseSettings: &resetDatabaseSettings{
				ignoreServiceErrors: cmdtoolkit.CommandFlag[bool]{Value: true},
			},
			stopped: false,
			want:    nil,
			WantedRecording: output.WantedRecording{
				Console: "No metadata files were found in " +
					"\"dummyProfile\\\\AppData\\\\Local\\\\Microsoft\\\\Media Player\".\n",
				Log: "" +
					"level='info'" +
					" directory='dummyProfile\\AppData\\Local\\Microsoft\\Media Player'" +
					" extension='.wmdb'" +
					" msg='no files found'\n",
			},
		},
		"not stopped but ignored, cannot read metadata directory": {
			readDirectory: func(_ output.Bus, _ string) ([]fs.FileInfo, bool) {
				return nil, false
			},
			resetDatabaseSettings: &resetDatabaseSettings{
				ignoreServiceErrors: cmdtoolkit.CommandFlag[bool]{Value: true},
			},
			stopped:         false,
			want:            nil,
			WantedRecording: output.WantedRecording{},
		},
		"work to do": {
			readDirectory: func(_ output.Bus, _ string) ([]fs.FileInfo, bool) {
				return []fs.FileInfo{newTestFile("foo.wmdb", nil)}, true
			},
			plainFileExists:       func(_ string) bool { return true },
			remove:                func(_ string) error { return nil },
			resetDatabaseSettings: &resetDatabaseSettings{},
			stopped:               true,
			want:                  nil,
			WantedRecording: output.WantedRecording{
				Console: "1 out of 1 metadata files have been deleted from " +
					"\"dummyProfile\\\\AppData\\\\Local\\\\Microsoft\\\\Media Player\".\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			readDirectory = tt.readDirectory
			plainFileExists = tt.plainFileExists
			remove = tt.remove
			o := output.NewRecorder()
			if got := tt.resetDatabaseSettings.cleanUpMetadata(o, tt.stopped); !compareExitErrors(got, tt.want) {
				t.Errorf("resetDatabaseSettings.cleanUpMetadata() %s want %s", got, tt.want)
			}
			o.Report(t, "resetDatabaseSettings.cleanUpMetadata()", tt.WantedRecording)
		})
	}
}

func Test_resetDatabaseSettings_resetService(t *testing.T) {
	originalDirty := dirty
	originalClearDirty := clearDirty
	originalConnect := connect
	originalProcessIsElevated := processIsElevated
	originalApplicationName := applicationName
	defer func() {
		dirty = originalDirty
		clearDirty = originalClearDirty
		connect = originalConnect
		processIsElevated = originalProcessIsElevated
		applicationName = originalApplicationName
	}()
	applicationName = "mp3repair"
	processIsElevated = func() bool { return false }
	clearDirty = func(_ output.Bus) {}
	connect = func() (*mgr.Mgr, error) { return nil, fmt.Errorf("access denied") }
	tests := map[string]struct {
		dirty                 func(o output.Bus) bool
		resetDatabaseSettings *resetDatabaseSettings
		want                  *cmdtoolkit.ExitError
		output.WantedRecording
	}{
		"not dirty, no force": {
			dirty:                 func(_ output.Bus) bool { return false },
			resetDatabaseSettings: &resetDatabaseSettings{force: cmdtoolkit.CommandFlag[bool]{Value: false}},
			want:                  cmdtoolkit.NewExitUserError("resetDatabase"),
			WantedRecording: output.WantedRecording{
				Error: "" +
					"The \"resetDatabase\" command has no work to perform.\n" +
					"Why?\n" +
					"The \"mp3repair\" program has not made any changes to any mp3 files\n" +
					"since the last successful database reset.\n" +
					"What to do:\n" +
					"If you believe the Windows Media Player database needs to be reset, run this command\n" +
					"again and use the \"--force\" flag.\n",
			},
		},
		"not dirty, force": {
			dirty:                 func(_ output.Bus) bool { return false },
			resetDatabaseSettings: &resetDatabaseSettings{force: cmdtoolkit.CommandFlag[bool]{Value: true}},
			want:                  cmdtoolkit.NewExitSystemError("resetDatabase"),
			WantedRecording: output.WantedRecording{
				Error: "" +
					"An attempt to connect with the service manager failed; error is 'access denied'.\n" +
					"Why?\n" +
					"This failure is likely to be due to lack of permissions.\n" +
					"What to do:\n" +
					"If you can, try running this command as an administrator.\n" +
					"Metadata files will not be deleted.\n" +
					"Why?\n" +
					"The Windows Media Player sharing service \"WMPNetworkSVC\" could not be stopped, and " +
					"\"--ignoreServiceErrors\" is false.\n" +
					"What to do:\n" +
					"Rerun this command with \"--ignoreServiceErrors\" set to true.\n",
				Log: "" +
					"level='error'" +
					" error='access denied'" +
					" msg='service manager connect failed'\n",
			},
		},
		"dirty, not force": {
			dirty:                 func(_ output.Bus) bool { return true },
			resetDatabaseSettings: &resetDatabaseSettings{},
			want:                  cmdtoolkit.NewExitSystemError("resetDatabase"),
			WantedRecording: output.WantedRecording{
				Error: "" +
					"An attempt to connect with the service manager failed; error is 'access denied'.\n" +
					"Why?\n" +
					"This failure is likely to be due to lack of permissions.\n" +
					"What to do:\n" +
					"If you can, try running this command as an administrator.\n" +
					"Metadata files will not be deleted.\n" +
					"Why?\n" +
					"The Windows Media Player sharing service \"WMPNetworkSVC\" could not be stopped, and " +
					"\"--ignoreServiceErrors\" is false.\n" +
					"What to do:\n" +
					"Rerun this command with \"--ignoreServiceErrors\" set to true.\n",
				Log: "" +
					"level='error'" +
					" error='access denied'" +
					" msg='service manager connect failed'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			dirty = tt.dirty
			o := output.NewRecorder()
			if got := tt.resetDatabaseSettings.resetService(o); !compareExitErrors(got, tt.want) {
				t.Errorf("resetDatabaseSettings.resetService() got %s want %s", got, tt.want)
			}
			o.Report(t, "resetDatabaseSettings.resetService()", tt.WantedRecording)
		})
	}
}

func Test_resetDatabaseRun(t *testing.T) {
	initGlobals()
	originalApplicationName := applicationName
	originalBus := bus
	originalDirty := dirty
	defer func() {
		bus = originalBus
		dirty = originalDirty
		applicationName = originalApplicationName
	}()
	applicationName = "mp3repair"
	dirty = func(_ output.Bus) bool { return false }
	flags := &cmdtoolkit.FlagSet{
		Name: "resetDatabase",
		Details: map[string]*cmdtoolkit.FlagDetails{
			"timeout": {
				AbbreviatedName: "t",
				Usage: fmt.Sprintf("timeout in seconds (minimum %d, maximum %d) for stopping the "+
					"media player service", 1, 60),
				ExpectedType: cmdtoolkit.IntType,
				DefaultValue: cmdtoolkit.NewIntBounds(1, 10, 60),
			},
			"service": {
				Usage:        "name of the Windows Media Player service",
				ExpectedType: cmdtoolkit.StringType,
				DefaultValue: "WMPNetworkSVC",
			},
			"metadataDir": {
				Usage:        "directory where the Windows Media Player service metadata files are" + " stored",
				ExpectedType: cmdtoolkit.StringType,
				DefaultValue: filepath.Join("AppData", "Local", "Microsoft", "Media Player"),
			},
			"extension": {
				Usage:        "extension for metadata files",
				ExpectedType: cmdtoolkit.StringType,
				DefaultValue: ".wmdb",
			},
			"force": {
				AbbreviatedName: "f",
				Usage:           "if set, force a database reset",
				ExpectedType:    cmdtoolkit.BoolType,
				DefaultValue:    false,
			},
			"ignoreServiceErrors": {
				AbbreviatedName: "i",
				Usage: "if set, ignore service errors and delete the Windows Media Player service's " +
					"metadata files",
				ExpectedType: cmdtoolkit.BoolType,
				DefaultValue: false,
			},
		},
	}
	myCommand := &cobra.Command{}
	cmdtoolkit.AddFlags(output.NewNilBus(), cmdtoolkit.EmptyConfiguration(), myCommand.Flags(),
		flags)
	tests := map[string]struct {
		cmd *cobra.Command
		in1 []string
		output.WantedRecording
	}{
		"simple": {
			cmd: myCommand,
			WantedRecording: output.WantedRecording{
				Error: "" +
					"The \"resetDatabase\" command has no work to perform.\n" +
					"Why?\n" +
					"The \"mp3repair\" program has not made any changes to any mp3 files\n" +
					"since the last successful database reset.\n" +
					"What to do:\n" +
					"If you believe the Windows Media Player database needs to be reset, run this" +
					" command\n" +
					"again and use the \"--force\" flag.\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			bus = o
			_ = resetDatabaseRun(tt.cmd, tt.in1)
			o.Report(t, "resetDatabaseRun()", tt.WantedRecording)
		})
	}
}

func Test_resetDatabase_Help(t *testing.T) {
	commandUnderTest := cloneCommand(resetDatabaseCmd)
	flagMap := map[string]*cmdtoolkit.FlagDetails{}
	for k, v := range resetDatabaseFlags.Details {
		flagMap[k] = v
	}
	flagCopy := &cmdtoolkit.FlagSet{Name: "resetDatabase", Details: flagMap}
	cmdtoolkit.AddFlags(output.NewNilBus(), cmdtoolkit.EmptyConfiguration(),
		commandUnderTest.Flags(), flagCopy)
	tests := map[string]struct {
		output.WantedRecording
	}{
		"good": {
			WantedRecording: output.WantedRecording{
				Console: "" +
					"\"resetDatabase\" resets the Windows Media Player database\n" +
					"\n" +
					"The changes made by the 'rewrite' command make the mp3 files inconsistent with the\n" +
					"Windows Media Player database which organizes the files into albums and artists. This command\n" +
					"resets that database, which it accomplishes by deleting the database files.\n" +
					"\n" +
					"Prior to deleting the files, the resetDatabase command attempts to stop the Windows\n" +
					"Media Player service, which allows Windows Media Player to share its database with a network. " +
					"If\n" +
					"there is such an active service, this command will need to be run as administrator. If, for\n" +
					"whatever reasons, the service cannot be stopped, using the\n" +
					"--ignoreServiceErrors flag allows the database files to be deleted, if possible.\n" +
					"\n" +
					"This command does nothing if it determines that the rewrite command has not made any\n" +
					"changes, unless the --force flag is set.\n" +
					"\n" +
					"Usage:\n" +
					"  resetDatabase [--timeout seconds] [--force] [--ignoreServiceErrors]\n" +
					"\n" +
					"Flags:\n" +
					"  -f, --force                 if set, force a database reset (default false)\n" +
					"  -i, --ignoreServiceErrors   if set, ignore service errors and delete the Windows Media Player " +
					"service's metadata files (default false)\n" +
					"  -t, --timeout int           timeout in seconds (minimum 1, maximum 60) for stopping the media " +
					"player service (default 10)\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			command := commandUnderTest
			enableCommandRecording(o, command)
			_ = command.Help()
			o.Report(t, "resetDatabase Help()", tt.WantedRecording)
		})
	}
}

func Test_updateServiceStatus(t *testing.T) {
	type args struct {
		currentStatus  *cmdtoolkit.ExitError
		proposedStatus *cmdtoolkit.ExitError
	}
	tests := map[string]struct {
		args
		want *cmdtoolkit.ExitError
	}{
		"success, success": {
			args: args{
				currentStatus:  nil,
				proposedStatus: nil,
			},
			want: nil,
		},
		"success, user error": {
			args: args{
				currentStatus:  nil,
				proposedStatus: cmdtoolkit.NewExitUserError("resetDatabase"),
			},
			want: cmdtoolkit.NewExitUserError("resetDatabase"),
		},
		"success, program error": {
			args: args{
				currentStatus:  nil,
				proposedStatus: cmdtoolkit.NewExitProgrammingError("resetDatabase"),
			},
			want: cmdtoolkit.NewExitProgrammingError("resetDatabase"),
		},
		"success, system error": {
			args: args{
				currentStatus:  nil,
				proposedStatus: cmdtoolkit.NewExitSystemError("resetDatabase"),
			},
			want: cmdtoolkit.NewExitSystemError("resetDatabase"),
		},
		"user error, success": {
			args: args{
				currentStatus:  cmdtoolkit.NewExitUserError("resetDatabase"),
				proposedStatus: nil,
			},
			want: cmdtoolkit.NewExitUserError("resetDatabase"),
		},
		"user error, user error": {
			args: args{
				currentStatus:  cmdtoolkit.NewExitUserError("resetDatabase"),
				proposedStatus: cmdtoolkit.NewExitUserError("resetDatabase"),
			},
			want: cmdtoolkit.NewExitUserError("resetDatabase"),
		},
		"user error, program error": {
			args: args{
				currentStatus:  cmdtoolkit.NewExitUserError("resetDatabase"),
				proposedStatus: cmdtoolkit.NewExitProgrammingError("resetDatabase"),
			},
			want: cmdtoolkit.NewExitUserError("resetDatabase"),
		},
		"user error, system error": {
			args: args{
				currentStatus:  cmdtoolkit.NewExitUserError("resetDatabase"),
				proposedStatus: cmdtoolkit.NewExitSystemError("resetDatabase"),
			},
			want: cmdtoolkit.NewExitUserError("resetDatabase"),
		},
		"program error, success": {
			args: args{
				currentStatus:  cmdtoolkit.NewExitProgrammingError("resetDatabase"),
				proposedStatus: nil,
			},
			want: cmdtoolkit.NewExitProgrammingError("resetDatabase"),
		},
		"program error, user error": {
			args: args{
				currentStatus:  cmdtoolkit.NewExitProgrammingError("resetDatabase"),
				proposedStatus: cmdtoolkit.NewExitUserError("resetDatabase"),
			},
			want: cmdtoolkit.NewExitProgrammingError("resetDatabase"),
		},
		"program error, program error": {
			args: args{
				currentStatus:  cmdtoolkit.NewExitProgrammingError("resetDatabase"),
				proposedStatus: cmdtoolkit.NewExitProgrammingError("resetDatabase"),
			},
			want: cmdtoolkit.NewExitProgrammingError("resetDatabase"),
		},
		"program error, system error": {
			args: args{
				currentStatus:  cmdtoolkit.NewExitProgrammingError("resetDatabase"),
				proposedStatus: cmdtoolkit.NewExitSystemError("resetDatabase"),
			},
			want: cmdtoolkit.NewExitProgrammingError("resetDatabase"),
		},
		"system error, success": {
			args: args{
				currentStatus:  cmdtoolkit.NewExitSystemError("resetDatabase"),
				proposedStatus: nil,
			},
			want: cmdtoolkit.NewExitSystemError("resetDatabase"),
		},
		"system error, user error": {
			args: args{
				currentStatus:  cmdtoolkit.NewExitSystemError("resetDatabase"),
				proposedStatus: cmdtoolkit.NewExitUserError("resetDatabase"),
			},
			want: cmdtoolkit.NewExitSystemError("resetDatabase"),
		},
		"system error, program error": {
			args: args{
				currentStatus:  cmdtoolkit.NewExitSystemError("resetDatabase"),
				proposedStatus: cmdtoolkit.NewExitProgrammingError("resetDatabase"),
			},
			want: cmdtoolkit.NewExitSystemError("resetDatabase"),
		},
		"system error, system error": {
			args: args{
				currentStatus:  cmdtoolkit.NewExitSystemError("resetDatabase"),
				proposedStatus: cmdtoolkit.NewExitSystemError("resetDatabase"),
			},
			want: cmdtoolkit.NewExitSystemError("resetDatabase"),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := updateServiceStatus(
				tt.args.currentStatus,
				tt.args.proposedStatus,
			); !compareExitErrors(got, tt.want) {
				t.Errorf("updateServiceStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_maybeClearDirty(t *testing.T) {
	originalClearDirty := clearDirty
	defer func() {
		clearDirty = originalClearDirty
	}()
	var clearDirtyCalled bool
	clearDirty = func(_ output.Bus) {
		clearDirtyCalled = true
	}
	tests := map[string]struct {
		status *cmdtoolkit.ExitError
		want   bool
	}{
		"success": {
			status: nil,
			want:   true,
		},
		"user error": {
			status: cmdtoolkit.NewExitUserError("resetDatabase"),
			want:   false,
		},
		"program error": {
			status: cmdtoolkit.NewExitProgrammingError("resetDatabase"),
			want:   false,
		},
		"system error": {
			status: cmdtoolkit.NewExitSystemError("resetDatabase"),
			want:   false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			clearDirtyCalled = false
			maybeClearDirty(output.NewNilBus(), tt.status)
			if got := clearDirtyCalled; got != tt.want {
				t.Errorf("maybeClearDirty = %t want %t", got, tt.want)
			}
		})
	}
}

func Test_outputSystemErrorCause(t *testing.T) {
	originalProcessIsElevated := processIsElevated
	defer func() {
		processIsElevated = originalProcessIsElevated
	}()
	tests := map[string]struct {
		isElevated func() bool
		output.WantedRecording
	}{
		"elevated": {
			isElevated: func() bool { return true },
		},
		"ordinary": {
			isElevated: func() bool { return false },
			WantedRecording: output.WantedRecording{
				Error: "" +
					"Why?\n" +
					"This failure is likely to be due to lack of permissions.\n" +
					"What to do:\n" +
					"If you can, try running this command as an administrator.\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			processIsElevated = tt.isElevated
			o := output.NewRecorder()
			outputSystemErrorCause(o)
			o.Report(t, "outputSystemErrorCause()", tt.WantedRecording)
		})
	}
}
