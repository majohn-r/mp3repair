/*
Copyright © 2021 Marc Johnson (marc.johnson27591@gmail.com)
*/
package cmd

import (
	"fmt"
	"io/fs"
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

func Test_processResetDBFlags(t *testing.T) {
	tests := map[string]struct {
		values map[string]*cmdtoolkit.CommandFlag[any]
		want   *ResetDBSettings
		want1  bool
		output.WantedRecording
	}{
		"massive errors": {
			values: nil,
			want:   &ResetDBSettings{},
			want1:  false,
			WantedRecording: output.WantedRecording{
				Error: "" +
					"An internal error occurred: no flag values exist.\n" +
					"An internal error occurred: no flag values exist.\n" +
					"An internal error occurred: no flag values exist.\n" +
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
					" msg='internal error'\n" +
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
			want: &ResetDBSettings{
				Extension:           cmdtoolkit.CommandFlag[string]{Value: ".foo"},
				Force:               cmdtoolkit.CommandFlag[bool]{Value: true},
				IgnoreServiceErrors: cmdtoolkit.CommandFlag[bool]{Value: true},
				MetadataDir:         cmdtoolkit.CommandFlag[string]{Value: "metadata"},
				Service:             cmdtoolkit.CommandFlag[string]{Value: "music service"},
				Timeout:             cmdtoolkit.CommandFlag[int]{Value: 5},
			},
			want1: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			got, got1 := ProcessResetDBFlags(o, tt.values)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("processResetDBFlags() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("processResetDBFlags() got1 = %v, want %v", got1, tt.want1)
			}
			o.Report(t, "processResetDBFlags()", tt.WantedRecording)
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

func Test_resetDBSettings_waitForStop(t *testing.T) {
	type args struct {
		s             ServiceRep
		expiration    time.Time
		checkInterval time.Duration
	}
	tests := map[string]struct {
		resetDBSettings *ResetDBSettings
		args
		wantOk     bool
		wantStatus *cmdtoolkit.ExitError
		output.WantedRecording
	}{
		"already timed out": {
			resetDBSettings: &ResetDBSettings{
				Service: cmdtoolkit.CommandFlag[string]{Value: "my service"},
				Timeout: cmdtoolkit.CommandFlag[int]{Value: 10},
			},
			args:       args{expiration: time.Now().Add(time.Duration(-1) * time.Second)},
			wantStatus: cmdtoolkit.NewExitSystemError("resetDatabase"),
			WantedRecording: output.WantedRecording{
				Error: "The service \"my service\" could not be stopped within the 10" +
					" second timeout.\n",
				Log: "" +
					"level='error'" +
					" error='timed out'" +
					" service='my service'" +
					" timeout='10'" +
					" trigger='Stop'" +
					" msg='service problem'\n",
			},
		},
		"query error": {
			resetDBSettings: &ResetDBSettings{
				Service: cmdtoolkit.CommandFlag[string]{Value: "my service"},
				Timeout: cmdtoolkit.CommandFlag[int]{Value: 10},
			},
			args: args{
				s:             newTestService(),
				expiration:    time.Now().Add(time.Duration(1) * time.Second),
				checkInterval: 1 * time.Millisecond,
			},
			wantStatus: cmdtoolkit.NewExitSystemError("resetDatabase"),
			WantedRecording: output.WantedRecording{
				Error: "An error occurred while attempting to stop the service " +
					"\"my service\": no results from query.\n",
				Log: "" +
					"level='error'" +
					" error='no results from query'" +
					" service='my service'" +
					" msg='service query error'\n",
			},
		},
		"stops correctly": {
			resetDBSettings: &ResetDBSettings{
				Service: cmdtoolkit.CommandFlag[string]{Value: "my service"},
				Timeout: cmdtoolkit.CommandFlag[int]{Value: 10},
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
					" service='my service'" +
					" msg='service stopped'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			gotOk, gotStatus := tt.resetDBSettings.WaitForStop(o, tt.args.s, tt.args.expiration, tt.args.checkInterval)
			if gotOk != tt.wantOk {
				t.Errorf("resetDBSettings.waitForStop() = %t, want %t", gotOk, tt.wantOk)
			}
			if !compareExitErrors(gotStatus, tt.wantStatus) {
				t.Errorf("resetDBSettings.waitForStop() = %s, want %s", gotStatus, tt.wantStatus)
			}
			o.Report(t, "resetDBSettings.waitForStop()", tt.WantedRecording)
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

func Test_resetDBSettings_stopFoundService(t *testing.T) {
	type args struct {
		manager ServiceManager
		service ServiceRep
	}
	tests := map[string]struct {
		resetDBSettings *ResetDBSettings
		args
		wantOk     bool
		wantStatus *cmdtoolkit.ExitError
		output.WantedRecording
	}{
		"defective service": {
			resetDBSettings: &ResetDBSettings{
				Service: cmdtoolkit.CommandFlag[string]{Value: "my service"},
				Timeout: cmdtoolkit.CommandFlag[int]{Value: 10},
			},
			args: args{
				manager: newTestManager(nil, nil),
				service: newTestService(),
			},
			wantStatus: cmdtoolkit.NewExitSystemError("resetDatabase"),
			WantedRecording: output.WantedRecording{
				Error: "An error occurred while trying to stop service \"my service\":" +
					" no results from query.\n",
				Log: "" +
					"level='error' " +
					"error='no results from query' " +
					"service='my service' " +
					"msg='service query error'\n",
			},
		},
		"already stopped": {
			resetDBSettings: &ResetDBSettings{
				Service: cmdtoolkit.CommandFlag[string]{Value: "my service"},
				Timeout: cmdtoolkit.CommandFlag[int]{Value: 10},
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
					" service='my service'" +
					" msg='service stopped'\n",
			},
		},
		"stopped easily": {
			resetDBSettings: &ResetDBSettings{
				Service: cmdtoolkit.CommandFlag[string]{Value: "my service"},
				Timeout: cmdtoolkit.CommandFlag[int]{Value: 10},
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
					" service='my service'" +
					" msg='service stopped'\n",
			},
		},
		"stopped with a little more difficulty": {
			resetDBSettings: &ResetDBSettings{
				Service: cmdtoolkit.CommandFlag[string]{Value: "my service"},
				Timeout: cmdtoolkit.CommandFlag[int]{Value: 10},
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
					" service='my service'" +
					" msg='service stopped'\n",
			},
		},
		"cannot be stopped": {
			resetDBSettings: &ResetDBSettings{
				Service: cmdtoolkit.CommandFlag[string]{Value: "my service"},
				Timeout: cmdtoolkit.CommandFlag[int]{Value: 10},
			},
			args: args{
				manager: newTestManager(nil, nil),
				service: newTestService(svc.Status{State: svc.Paused}),
			},
			wantStatus: cmdtoolkit.NewExitSystemError("resetDatabase"),
			WantedRecording: output.WantedRecording{
				Error: "The service \"my service\" cannot be stopped:" +
					" no results from query.\n",
				Log: "" +
					"level='error'" +
					" error='no results from query'" +
					" service='my service'" +
					" trigger='Stop'" +
					" msg='service problem'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			gotOk, gotStatus := tt.resetDBSettings.StopFoundService(o, tt.args.manager,
				tt.args.service)
			if gotOk != tt.wantOk {
				t.Errorf("resetDBSettings.stopFoundService() = %v, want %v", gotOk, tt.wantOk)
			}
			if !compareExitErrors(gotStatus, tt.wantStatus) {
				t.Errorf("resetDBSettings.stopFoundService() = %v, want %v", gotStatus, tt.wantStatus)
			}
			o.Report(t, "resetDBSettings.stopFoundService()", tt.WantedRecording)
		})
	}
}

func Test_addServiceState(t *testing.T) {
	tests := map[string]struct {
		m           map[string][]string
		s           ServiceRep
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
			AddServiceState(tt.m, tt.s, tt.serviceName)
			if !reflect.DeepEqual(tt.m, tt.want) {
				t.Errorf("addServiceState() = %v, want %v", tt.m, tt.want)
			}
		})
	}
}

func Test_listServices(t *testing.T) {
	type args struct {
		manager  ServiceManager
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
			ListServices(o, tt.args.manager, tt.args.services)
			o.Report(t, "listServices()", tt.WantedRecording)
		})
	}
}

func Test_resetDBSettings_disableService(t *testing.T) {
	tests := map[string]struct {
		resetDBSettings *ResetDBSettings
		manager         ServiceManager
		wantOk          bool
		wantStatus      *cmdtoolkit.ExitError
		output.WantedRecording
	}{
		"defective manager #1": {
			resetDBSettings: &ResetDBSettings{
				Service: cmdtoolkit.CommandFlag[string]{Value: "my service"},
				Timeout: cmdtoolkit.CommandFlag[int]{Value: 1},
			},
			manager:    newTestManager(nil, []string{"my service"}),
			wantStatus: cmdtoolkit.NewExitSystemError("resetDatabase"),
			WantedRecording: output.WantedRecording{
				Error: "" +
					"The service \"my service\" cannot be opened: no such service.\n" +
					"The following services are available:\n" +
					"  State \"no such service\":\n" +
					"    \"my service\"\n",
				Log: "" +
					"level='error'" +
					" error='no such service'" +
					" service='my service'" +
					" trigger='OpenService'" +
					" msg='service problem'\n",
			},
		},
		"defective manager #2": {
			resetDBSettings: &ResetDBSettings{
				Service: cmdtoolkit.CommandFlag[string]{Value: "my service"},
				Timeout: cmdtoolkit.CommandFlag[int]{Value: 1},
			},
			manager:    newTestManager(nil, nil),
			wantStatus: cmdtoolkit.NewExitSystemError("resetDatabase"),
			WantedRecording: output.WantedRecording{
				Error: "The service \"my service\" cannot be opened: no such service.\n",
				Log: "" +
					"level='error'" +
					" error='no such service'" +
					" service='my service'" +
					" trigger='OpenService'" +
					" msg='service problem'\n" +
					"level='error'" +
					" error='no services'" +
					" trigger='ListServices'" +
					" msg='service problem'\n",
			},
		},
		"defective manager #3": {
			resetDBSettings: &ResetDBSettings{
				Service: cmdtoolkit.CommandFlag[string]{Value: "my service"},
				Timeout: cmdtoolkit.CommandFlag[int]{Value: 1},
			},
			manager:    newTestManager(map[string]*mgr.Service{"my service": nil}, nil),
			wantStatus: cmdtoolkit.NewExitSystemError("resetDatabase"),
			WantedRecording: output.WantedRecording{
				Error: "An error occurred while trying to stop service \"my service\":" +
					" no service.\n",
				Log: "" +
					"level='error'" +
					" error='no service'" +
					" service='my service'" +
					" msg='service query error'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			gotOk, gotStatus := tt.resetDBSettings.DisableService(o, tt.manager)
			if gotOk != tt.wantOk {
				t.Errorf("resetDBSettings.disableService() = %v, want %v", gotOk, tt.wantOk)
			}
			if !compareExitErrors(gotStatus, tt.wantStatus) {
				t.Errorf("resetDBSettings.disableService() = %v, want %v", gotStatus, tt.wantStatus)
			}
			o.Report(t, "resetDBSettings.disableService()", tt.WantedRecording)
		})
	}
}

func Test_resetDBSettings_stopService(t *testing.T) {
	originalConnect := Connect
	originalProcessIsElevated := ProcessIsElevated
	defer func() {
		Connect = originalConnect
		ProcessIsElevated = originalProcessIsElevated
	}()
	ProcessIsElevated = func() bool { return false }
	tests := map[string]struct {
		connect         func() (*mgr.Mgr, error)
		resetDBSettings *ResetDBSettings
		wantOk          bool
		wantStatus      *cmdtoolkit.ExitError
		output.WantedRecording
	}{
		"connect fails": {
			connect: func() (*mgr.Mgr, error) {
				return nil, fmt.Errorf("no manager available")
			},
			resetDBSettings: &ResetDBSettings{
				Service: cmdtoolkit.CommandFlag[string]{Value: "my service"},
				Timeout: cmdtoolkit.CommandFlag[int]{Value: 1},
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
			resetDBSettings: &ResetDBSettings{
				Service: cmdtoolkit.CommandFlag[string]{Value: "my service"},
				Timeout: cmdtoolkit.CommandFlag[int]{Value: 1},
			},
			wantStatus: cmdtoolkit.NewExitSystemError("resetDatabase"),
			WantedRecording: output.WantedRecording{
				Error: "The service \"my service\" cannot be opened: nil manager.\n",
				Log: "" +
					"level='error'" +
					" error='nil manager'" +
					" service='my service'" +
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
			Connect = tt.connect
			o := output.NewRecorder()
			gotOk, gotStatus := tt.resetDBSettings.StopService(o)
			if gotOk != tt.wantOk {
				t.Errorf("resetDBSettings.stopService() = %v, want %v", gotOk, tt.wantOk)
			}
			if !compareExitErrors(gotStatus, tt.wantStatus) {
				t.Errorf("resetDBSettings.stopService() = %v, want %v", gotStatus, tt.wantStatus)
			}
			o.Report(t, "resetDBSettings.stopService()", tt.WantedRecording)
		})
	}
}

func Test_resetDBSettings_deleteMetadataFiles(t *testing.T) {
	originalRemove := Remove
	defer func() {
		Remove = originalRemove
	}()
	tests := map[string]struct {
		remove          func(string) error
		resetDBSettings *ResetDBSettings
		paths           []string
		want            *cmdtoolkit.ExitError
		output.WantedRecording
	}{
		"no files": {
			resetDBSettings: &ResetDBSettings{MetadataDir: cmdtoolkit.CommandFlag[string]{Value: "metadata/dir"}},
			paths:           nil,
			want:            nil,
		},
		"locked files": {
			remove:          func(_ string) error { return fmt.Errorf("cannot remove file") },
			resetDBSettings: &ResetDBSettings{MetadataDir: cmdtoolkit.CommandFlag[string]{Value: "metadata/dir"}},
			paths:           []string{"file1", "file2"},
			want:            cmdtoolkit.NewExitSystemError("resetDatabase"),
			WantedRecording: output.WantedRecording{
				Console: "0 out of 2 metadata files have been deleted from" +
					" \"metadata/dir\".\n",
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
			remove:          func(_ string) error { return nil },
			resetDBSettings: &ResetDBSettings{MetadataDir: cmdtoolkit.CommandFlag[string]{Value: "metadata/dir"}},
			paths:           []string{"file1", "file2"},
			want:            nil,
			WantedRecording: output.WantedRecording{
				Console: "2 out of 2 metadata files have been deleted from" +
					" \"metadata/dir\".\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			Remove = tt.remove
			o := output.NewRecorder()
			if got := tt.resetDBSettings.DeleteMetadataFiles(o, tt.paths); !compareExitErrors(got, tt.want) {
				t.Errorf("resetDBSettings.deleteMetadataFiles() %s want %s", got, tt.want)
			}
			o.Report(t, "resetDBSettings.deleteMetadataFiles()", tt.WantedRecording)
		})
	}
}

func Test_resetDBSettings_filterMetadataFiles(t *testing.T) {
	originalPlainFileExists := PlainFileExists
	defer func() {
		PlainFileExists = originalPlainFileExists
	}()
	tests := map[string]struct {
		plainFileExists func(string) bool
		resetDBSettings *ResetDBSettings
		entries         []fs.FileInfo
		want            []string
	}{
		"no entries": {want: []string{}},
		"mixed entries": {
			plainFileExists: func(s string) bool { return !strings.Contains(s, "dir.") },
			resetDBSettings: &ResetDBSettings{
				MetadataDir: cmdtoolkit.CommandFlag[string]{Value: "metadata"},
				Extension:   cmdtoolkit.CommandFlag[string]{Value: ".db"},
			},
			entries: []fs.FileInfo{
				newTestFile("dir. foo.db", nil),
				newTestFile("foo.db", nil),
				newTestFile("foo", nil),
			},
			want: []string{filepath.Join("metadata", "foo.db")},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			PlainFileExists = tt.plainFileExists
			if got := tt.resetDBSettings.FilterMetadataFiles(tt.entries); !reflect.DeepEqual(got,
				tt.want) {
				t.Errorf("resetDBSettings.filterMetadataFiles() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_resetDBSettings_cleanUpMetadata(t *testing.T) {
	originalReadDirectory := ReadDirectory
	originalPlainFileExists := PlainFileExists
	originalRemove := Remove
	defer func() {
		ReadDirectory = originalReadDirectory
		PlainFileExists = originalPlainFileExists
		Remove = originalRemove
	}()
	tests := map[string]struct {
		readDirectory   func(output.Bus, string) ([]fs.FileInfo, bool)
		plainFileExists func(string) bool
		remove          func(string) error
		resetDBSettings *ResetDBSettings
		stopped         bool
		want            *cmdtoolkit.ExitError
		output.WantedRecording
	}{
		"did not stop, cannot ignore it": {
			resetDBSettings: &ResetDBSettings{
				Service:             cmdtoolkit.CommandFlag[string]{Value: "musicService"},
				IgnoreServiceErrors: cmdtoolkit.CommandFlag[bool]{Value: false},
			},
			stopped: false,
			want:    cmdtoolkit.NewExitUserError("resetDatabase"),
			WantedRecording: output.WantedRecording{
				Error: "" +
					"Metadata files will not be deleted.\n" +
					"Why?\n" +
					"The music service \"musicService\" could not be stopped, and" +
					" \"--ignoreServiceErrors\" is false.\n" +
					"What to do:\n" +
					"Rerun this command with \"--ignoreServiceErrors\" set to true.\n",
			},
		},
		"stopped, no metadata": {
			readDirectory: func(_ output.Bus, _ string) ([]fs.FileInfo, bool) {
				return nil, true
			},
			resetDBSettings: &ResetDBSettings{MetadataDir: cmdtoolkit.CommandFlag[string]{Value: "metadata"}},
			stopped:         true,
			want:            nil,
			WantedRecording: output.WantedRecording{
				Console: "No metadata files were found in \"metadata\".\n",
				Log: "" +
					"level='info'" +
					" directory='metadata'" +
					" extension=''" +
					" msg='no files found'\n",
			},
		},
		"not stopped but ignored, no metadata": {
			readDirectory: func(_ output.Bus, _ string) ([]fs.FileInfo, bool) {
				return nil, true
			},
			resetDBSettings: &ResetDBSettings{
				MetadataDir:         cmdtoolkit.CommandFlag[string]{Value: "metadata"},
				IgnoreServiceErrors: cmdtoolkit.CommandFlag[bool]{Value: true},
			},
			stopped: false,
			want:    nil,
			WantedRecording: output.WantedRecording{
				Console: "No metadata files were found in \"metadata\".\n",
				Log: "" +
					"level='info'" +
					" directory='metadata'" +
					" extension=''" +
					" msg='no files found'\n",
			},
		},
		"not stopped but ignored, cannot read metadata directory": {
			readDirectory: func(_ output.Bus, _ string) ([]fs.FileInfo, bool) {
				return nil, false
			},
			resetDBSettings: &ResetDBSettings{
				MetadataDir:         cmdtoolkit.CommandFlag[string]{Value: "metadata"},
				IgnoreServiceErrors: cmdtoolkit.CommandFlag[bool]{Value: true},
			},
			stopped:         false,
			want:            nil,
			WantedRecording: output.WantedRecording{},
		},
		"work to do": {
			readDirectory: func(_ output.Bus, _ string) ([]fs.FileInfo, bool) {
				return []fs.FileInfo{newTestFile("foo.db", nil)}, true
			},
			plainFileExists: func(_ string) bool { return true },
			remove:          func(_ string) error { return nil },
			resetDBSettings: &ResetDBSettings{
				MetadataDir: cmdtoolkit.CommandFlag[string]{Value: "metadata"},
				Extension:   cmdtoolkit.CommandFlag[string]{Value: ".db"},
			},
			stopped: true,
			want:    nil,
			WantedRecording: output.WantedRecording{
				Console: "1 out of 1 metadata files have been deleted from \"metadata\".\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ReadDirectory = tt.readDirectory
			PlainFileExists = tt.plainFileExists
			Remove = tt.remove
			o := output.NewRecorder()
			if got := tt.resetDBSettings.CleanUpMetadata(o, tt.stopped); !compareExitErrors(got, tt.want) {
				t.Errorf("resetDBSettings.cleanUpMetadata() %s want %s", got, tt.want)
			}
			o.Report(t, "resetDBSettings.cleanUpMetadata()", tt.WantedRecording)
		})
	}
}

func Test_resetDBSettings_resetService(t *testing.T) {
	originalDirty := Dirty
	originalClearDirty := ClearDirty
	originalConnect := Connect
	originalProcessIsElevated := ProcessIsElevated
	defer func() {
		Dirty = originalDirty
		ClearDirty = originalClearDirty
		Connect = originalConnect
		ProcessIsElevated = originalProcessIsElevated
	}()
	ProcessIsElevated = func() bool { return false }
	ClearDirty = func(_ output.Bus) {}
	Connect = func() (*mgr.Mgr, error) { return nil, fmt.Errorf("access denied") }
	tests := map[string]struct {
		dirty           func() bool
		resetDBSettings *ResetDBSettings
		want            *cmdtoolkit.ExitError
		output.WantedRecording
	}{
		"not dirty, no force": {
			dirty:           func() bool { return false },
			resetDBSettings: &ResetDBSettings{Force: cmdtoolkit.CommandFlag[bool]{Value: false}},
			want:            cmdtoolkit.NewExitUserError("resetDatabase"),
			WantedRecording: output.WantedRecording{
				Error: "" +
					"The \"resetDatabase\" command has no work to perform.\n" +
					"Why?\n" +
					"The \"mp3repair\" program has not made any changes to any mp3 files\n" +
					"since the last successful database reset.\n" +
					"What to do:\n" +
					"If you believe the Windows database needs to be reset, run this command\n" +
					"again and use the \"--force\" flag.\n",
			},
		},
		"not dirty, force": {
			dirty:           func() bool { return false },
			resetDBSettings: &ResetDBSettings{Force: cmdtoolkit.CommandFlag[bool]{Value: true}},
			want:            cmdtoolkit.NewExitSystemError("resetDatabase"),
			WantedRecording: output.WantedRecording{
				Error: "" +
					"An attempt to connect with the service manager failed;" +
					" error is 'access denied'.\n" +
					"Why?\n" +
					"This failure is likely to be due to lack of permissions.\n" +
					"What to do:\n" +
					"If you can, try running this command as an administrator.\n" +
					"Metadata files will not be deleted.\n" +
					"Why?\n" +
					"The music service \"\" could not be stopped, and" +
					" \"--ignoreServiceErrors\" is false.\n" +
					"What to do:\n" +
					"Rerun this command with \"--ignoreServiceErrors\" set to true.\n",
				Log: "" +
					"level='error'" +
					" error='access denied'" +
					" msg='service manager connect failed'\n",
			},
		},
		"dirty, not force": {
			dirty:           func() bool { return true },
			resetDBSettings: &ResetDBSettings{},
			want:            cmdtoolkit.NewExitSystemError("resetDatabase"),
			WantedRecording: output.WantedRecording{
				Error: "" +
					"An attempt to connect with the service manager failed;" +
					" error is 'access denied'.\n" +
					"Why?\n" +
					"This failure is likely to be due to lack of permissions.\n" +
					"What to do:\n" +
					"If you can, try running this command as an administrator.\n" +
					"Metadata files will not be deleted.\n" +
					"Why?\n" +
					"The music service \"\" could not be stopped, and" +
					" \"--ignoreServiceErrors\" is false.\n" +
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
			Dirty = tt.dirty
			o := output.NewRecorder()
			if got := tt.resetDBSettings.ResetService(o); !compareExitErrors(got, tt.want) {
				t.Errorf("resetDBSettings.resetService() got %s want %s", got, tt.want)
			}
			o.Report(t, "resetDBSettings.resetService()", tt.WantedRecording)
		})
	}
}

func Test_resetDBRun(t *testing.T) {
	InitGlobals()
	originalBus := Bus
	originalDirty := Dirty
	defer func() {
		Bus = originalBus
		Dirty = originalDirty
	}()
	Dirty = func() bool { return false }
	flags := &cmdtoolkit.FlagSet{
		Name: "resetDatabase",
		Details: map[string]*cmdtoolkit.FlagDetails{
			"timeout": {
				AbbreviatedName: "t",
				Usage:           fmt.Sprintf("timeout in seconds (minimum %d, maximum %d) for stopping the media player service", 1, 60),
				ExpectedType:    cmdtoolkit.IntType,
				DefaultValue:    cmdtoolkit.NewIntBounds(1, 10, 60),
			},
			"service": {
				Usage:        "name of the media player service",
				ExpectedType: cmdtoolkit.StringType,
				DefaultValue: "WMPNetworkSVC",
			},
			"metadataDir": {
				Usage:        "directory where the media player service metadata files are" + " stored",
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
				Usage:           "if set, ignore service errors and delete the media player service metadata files",
				ExpectedType:    cmdtoolkit.BoolType,
				DefaultValue:    false,
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
					"If you believe the Windows database needs to be reset, run this" +
					" command\n" +
					"again and use the \"--force\" flag.\n",
				Log: "" +
					"level='info'" +
					" --extension='.wmdb'" +
					" --force='false'" +
					" --ignoreServiceErrors='false'" +
					" --metadataDir='AppData\\Local\\Microsoft\\Media Player'" +
					" --service='WMPNetworkSVC'" +
					" --timeout='10'" +
					" command='resetDatabase'" +
					" msg='executing command'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			Bus = o
			_ = ResetDBRun(tt.cmd, tt.in1)
			o.Report(t, "resetDBRun()", tt.WantedRecording)
		})
	}
}

func Test_resetDatabase_Help(t *testing.T) {
	commandUnderTest := cloneCommand(ResetDatabaseCmd)
	flagMap := map[string]*cmdtoolkit.FlagDetails{}
	for k, v := range ResetDatabaseFlags.Details {
		switch k {
		case "metadataDir":
			flagMap[k] = v.Copy()
			flagMap[k].DefaultValue = "[USERPROFILE]/AppData/Local/Microsoft/Media Player"
		default:
			flagMap[k] = v
		}
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
					"\"resetDatabase\" resets the Windows music database\n" +
					"\n" +
					"The changes made by the 'repair' command make the mp3 files" +
					" inconsistent with the\n" +
					"database Windows uses to organize the files into albums and artists." +
					" This command\n" +
					"resets that database, which it accomplishes by deleting the database" +
					" files.\n" +
					"\n" +
					"Prior to deleting the files, the resetDatabase command attempts to" +
					" stop the Windows\n" +
					"media player service. If there is such an active service, this" +
					" command will need to be\n" +
					"run as administrator. If, for whatever reasons, the service cannot be" +
					" stopped, using the\n" +
					"--ignoreServiceErrors flag allows the database files to be deleted, if" +
					" possible.\n" +
					"\n" +
					"This command does nothing if it determines that the repair command has not made any\n" +
					"changes, unless the --force flag is set.\n" +
					"\n" +
					"Usage:\n" +
					"  resetDatabase [--timeout seconds] [--service name]" +
					" [--metadataDir dir] [--extension string] [--force]" +
					" [--ignoreServiceErrors]\n" +
					"\n" +
					"Flags:\n" +
					"      --extension string      " +
					"extension for metadata files (default \".wmdb\")\n" +
					"  -f, --force                 " +
					"if set, force a database reset (default false)\n" +
					"  -i, --ignoreServiceErrors   " +
					"if set, ignore service errors and delete the media player service" +
					" metadata files (default false)\n" +
					"      --metadataDir string    " +
					"directory where the media player service metadata files are stored" +
					" (default \"[USERPROFILE]/AppData/Local/Microsoft/Media Player\")\n" +
					"      --service string        " +
					"name of the media player service (default \"WMPNetworkSVC\")\n" +
					"  -t, --timeout int           " +
					"timeout in seconds (minimum 1, maximum 60) for stopping the media" +
					" player service (default 10)\n",
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
			if got := UpdateServiceStatus(tt.args.currentStatus, tt.args.proposedStatus); !compareExitErrors(got, tt.want) {
				t.Errorf("updateServiceStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_maybeClearDirty(t *testing.T) {
	originalClearDirty := ClearDirty
	defer func() {
		ClearDirty = originalClearDirty
	}()
	var clearDirtyCalled bool
	ClearDirty = func(_ output.Bus) {
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
			MaybeClearDirty(output.NewNilBus(), tt.status)
			if got := clearDirtyCalled; got != tt.want {
				t.Errorf("maybeClearDirty = %t want %t", got, tt.want)
			}
		})
	}
}

func Test_outputSystemErrorCause(t *testing.T) {
	originalProcessIsElevated := ProcessIsElevated
	defer func() {
		ProcessIsElevated = originalProcessIsElevated
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
			ProcessIsElevated = tt.isElevated
			o := output.NewRecorder()
			OutputSystemErrorCause(o)
			o.Report(t, "outputSystemErrorCause()", tt.WantedRecording)
		})
	}
}
