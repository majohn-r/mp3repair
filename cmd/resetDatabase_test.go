/*
Copyright © 2021 Marc Johnson (marc.johnson27591@gmail.com)
*/
package cmd_test

import (
	"fmt"
	"io/fs"
	"mp3repair/cmd"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	cmd_toolkit "github.com/majohn-r/cmd-toolkit"
	"github.com/majohn-r/output"
	"github.com/spf13/cobra"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

func TestProcessResetDBFlags(t *testing.T) {
	tests := map[string]struct {
		values map[string]*cmd.FlagValue
		want   *cmd.ResetDBSettings
		want1  bool
		output.WantedRecording
	}{
		"massive errors": {
			values: nil,
			want:   &cmd.ResetDBSettings{},
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
			values: map[string]*cmd.FlagValue{
				"extension":           {Value: ".foo"},
				"force":               {Value: true},
				"ignoreServiceErrors": {Value: true},
				"metadataDir":         {Value: "metadata"},
				"service":             {Value: "music service"},
				"timeout":             {Value: 5},
			},
			want: &cmd.ResetDBSettings{
				Extension:           cmd.StringValue{Value: ".foo"},
				Force:               cmd.BoolValue{Value: true},
				IgnoreServiceErrors: cmd.BoolValue{Value: true},
				MetadataDir:         cmd.StringValue{Value: "metadata"},
				Service:             cmd.StringValue{Value: "music service"},
				Timeout:             cmd.IntValue{Value: 5},
			},
			want1: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			got, got1 := cmd.ProcessResetDBFlags(o, tt.values)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ProcessResetDBFlags() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("ProcessResetDBFlags() got1 = %v, want %v", got1, tt.want1)
			}
			o.Report(t, "ProcessResetDBFlags()", tt.WantedRecording)
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

func (ts *testService) Control(c svc.Cmd) (svc.Status, error) {
	return ts.Query()
}

func newTestService(values ...svc.Status) *testService {
	ts := &testService{
		queries:  0,
		statuses: values,
	}
	return ts
}

func TestResetDBSettings_WaitForStop(t *testing.T) {
	type args struct {
		s             cmd.ServiceRep
		expiration    time.Time
		checkInterval time.Duration
	}
	tests := map[string]struct {
		rdbs *cmd.ResetDBSettings
		args
		wantOk     bool
		wantStatus *cmd.ExitError
		output.WantedRecording
	}{
		"already timed out": {
			rdbs: &cmd.ResetDBSettings{
				Service: cmd.StringValue{Value: "my service"},
				Timeout: cmd.IntValue{Value: 10},
			},
			args:       args{expiration: time.Now().Add(time.Duration(-1) * time.Second)},
			wantStatus: cmd.NewExitSystemError("resetDatabase"),
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
			rdbs: &cmd.ResetDBSettings{
				Service: cmd.StringValue{Value: "my service"},
				Timeout: cmd.IntValue{Value: 10},
			},
			args: args{
				s:             newTestService(),
				expiration:    time.Now().Add(time.Duration(1) * time.Second),
				checkInterval: 1 * time.Millisecond,
			},
			wantStatus: cmd.NewExitSystemError("resetDatabase"),
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
			rdbs: &cmd.ResetDBSettings{
				Service: cmd.StringValue{Value: "my service"},
				Timeout: cmd.IntValue{Value: 10},
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
			gotOk, gotStatus := tt.rdbs.WaitForStop(o, tt.args.s, tt.args.expiration, tt.args.checkInterval)
			if gotOk != tt.wantOk {
				t.Errorf("ResetDBSettings.WaitForStop() = %t, want %t", gotOk, tt.wantOk)
			}
			if !compareExitErrors(gotStatus, tt.wantStatus) {
				t.Errorf("ResetDBSettings.WaitForStop() = %s, want %s", gotStatus, tt.wantStatus)
			}
			o.Report(t, "ResetDBSettings.WaitForStop()", tt.WantedRecording)
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

func TestResetDBSettings_StopFoundService(t *testing.T) {
	type args struct {
		manager cmd.ServiceManager
		service cmd.ServiceRep
	}
	tests := map[string]struct {
		rdbs *cmd.ResetDBSettings
		args
		wantOk     bool
		wantStatus *cmd.ExitError
		output.WantedRecording
	}{
		"defective service": {
			rdbs: &cmd.ResetDBSettings{
				Service: cmd.StringValue{Value: "my service"},
				Timeout: cmd.IntValue{Value: 10},
			},
			args: args{
				manager: newTestManager(nil, nil),
				service: newTestService(),
			},
			wantStatus: cmd.NewExitSystemError("resetDatabase"),
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
			rdbs: &cmd.ResetDBSettings{
				Service: cmd.StringValue{Value: "my service"},
				Timeout: cmd.IntValue{Value: 10},
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
			rdbs: &cmd.ResetDBSettings{
				Service: cmd.StringValue{Value: "my service"},
				Timeout: cmd.IntValue{Value: 10},
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
			rdbs: &cmd.ResetDBSettings{
				Service: cmd.StringValue{Value: "my service"},
				Timeout: cmd.IntValue{Value: 10},
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
			rdbs: &cmd.ResetDBSettings{
				Service: cmd.StringValue{Value: "my service"},
				Timeout: cmd.IntValue{Value: 10},
			},
			args: args{
				manager: newTestManager(nil, nil),
				service: newTestService(svc.Status{State: svc.Paused}),
			},
			wantStatus: cmd.NewExitSystemError("resetDatabase"),
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
			gotOk, gotStatus := tt.rdbs.StopFoundService(o, tt.args.manager,
				tt.args.service)
			if gotOk != tt.wantOk {
				t.Errorf("ResetDBSettings.StopFoundService() = %v, want %v", gotOk, tt.wantOk)
			}
			if !compareExitErrors(gotStatus, tt.wantStatus) {
				t.Errorf("ResetDBSettings.StopFoundService() = %v, want %v", gotStatus, tt.wantStatus)
			}
			o.Report(t, "ResetDBSettings.WaitForStop()", tt.WantedRecording)
		})
	}
}

func TestAddServiceState(t *testing.T) {
	tests := map[string]struct {
		m           map[string][]string
		s           cmd.ServiceRep
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
			cmd.AddServiceState(tt.m, tt.s, tt.serviceName)
			if !reflect.DeepEqual(tt.m, tt.want) {
				t.Errorf("AddServiceState() = %v, want %v", tt.m, tt.want)
			}
		})
	}
}

func Test_listServices(t *testing.T) {
	type args struct {
		manager  cmd.ServiceManager
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
			cmd.ListServices(o, tt.args.manager, tt.args.services)
			o.Report(t, "ListServices()", tt.WantedRecording)
		})
	}
}

func TestResetDBSettings_DisableService(t *testing.T) {
	tests := map[string]struct {
		rdbs       *cmd.ResetDBSettings
		manager    cmd.ServiceManager
		wantOk     bool
		wantStatus *cmd.ExitError
		output.WantedRecording
	}{
		"defective manager #1": {
			rdbs: &cmd.ResetDBSettings{
				Service: cmd.StringValue{Value: "my service"},
				Timeout: cmd.IntValue{Value: 1},
			},
			manager:    newTestManager(nil, []string{"my service"}),
			wantStatus: cmd.NewExitSystemError("resetDatabase"),
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
			rdbs: &cmd.ResetDBSettings{
				Service: cmd.StringValue{Value: "my service"},
				Timeout: cmd.IntValue{Value: 1},
			},
			manager:    newTestManager(nil, nil),
			wantStatus: cmd.NewExitSystemError("resetDatabase"),
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
			rdbs: &cmd.ResetDBSettings{
				Service: cmd.StringValue{Value: "my service"},
				Timeout: cmd.IntValue{Value: 1},
			},
			manager:    newTestManager(map[string]*mgr.Service{"my service": nil}, nil),
			wantStatus: cmd.NewExitSystemError("resetDatabase"),
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
			gotOk, gotStatus := tt.rdbs.DisableService(o, tt.manager)
			if gotOk != tt.wantOk {
				t.Errorf("ResetDBSettings.DisableService() = %v, want %v", gotOk, tt.wantOk)
			}
			if !compareExitErrors(gotStatus, tt.wantStatus) {
				t.Errorf("ResetDBSettings.DisableService() = %v, want %v", gotStatus, tt.wantStatus)
			}
			o.Report(t, "ResetDBSettings.DisableService()", tt.WantedRecording)
		})
	}
}

func TestResetDBSettings_StopService(t *testing.T) {
	originalConnect := cmd.Connect
	originalIsElevated := cmd.IsElevated
	defer func() {
		cmd.Connect = originalConnect
		cmd.IsElevated = originalIsElevated
	}()
	cmd.IsElevated = func(_ windows.Token) bool { return false }
	tests := map[string]struct {
		connect    func() (*mgr.Mgr, error)
		rdbs       *cmd.ResetDBSettings
		wantOk     bool
		wantStatus *cmd.ExitError
		output.WantedRecording
	}{
		"connect fails": {
			connect: func() (*mgr.Mgr, error) {
				return nil, fmt.Errorf("no manager available")
			},
			rdbs: &cmd.ResetDBSettings{
				Service: cmd.StringValue{Value: "my service"},
				Timeout: cmd.IntValue{Value: 1},
			},
			wantStatus: cmd.NewExitSystemError("resetDatabase"),
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
			rdbs: &cmd.ResetDBSettings{
				Service: cmd.StringValue{Value: "my service"},
				Timeout: cmd.IntValue{Value: 1},
			},
			wantStatus: cmd.NewExitSystemError("resetDatabase"),
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
			cmd.Connect = tt.connect
			o := output.NewRecorder()
			gotOk, gotStatus := tt.rdbs.StopService(o)
			if gotOk != tt.wantOk {
				t.Errorf("ResetDBSettings.StopService() = %v, want %v", gotOk, tt.wantOk)
			}
			if !compareExitErrors(gotStatus, tt.wantStatus) {
				t.Errorf("ResetDBSettings.StopService() = %v, want %v", gotStatus, tt.wantStatus)
			}
			o.Report(t, "ResetDBSettings.StopService()", tt.WantedRecording)
		})
	}
}

func TestResetDBSettings_DeleteMetadataFiles(t *testing.T) {
	originalRemove := cmd.Remove
	defer func() {
		cmd.Remove = originalRemove
	}()
	tests := map[string]struct {
		remove func(string) error
		rdbs   *cmd.ResetDBSettings
		paths  []string
		want   *cmd.ExitError
		output.WantedRecording
	}{
		"no files": {
			rdbs:  &cmd.ResetDBSettings{MetadataDir: cmd.StringValue{Value: "metadata/dir"}},
			paths: nil,
			want:  nil,
		},
		"undeletable files": {
			remove: func(_ string) error { return fmt.Errorf("cannot remove file") },
			rdbs:   &cmd.ResetDBSettings{MetadataDir: cmd.StringValue{Value: "metadata/dir"}},
			paths:  []string{"file1", "file2"},
			want:   cmd.NewExitSystemError("resetDatabase"),
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
			remove: func(_ string) error { return nil },
			rdbs:   &cmd.ResetDBSettings{MetadataDir: cmd.StringValue{Value: "metadata/dir"}},
			paths:  []string{"file1", "file2"},
			want:   nil,
			WantedRecording: output.WantedRecording{
				Console: "2 out of 2 metadata files have been deleted from" +
					" \"metadata/dir\".\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			cmd.Remove = tt.remove
			o := output.NewRecorder()
			if got := tt.rdbs.DeleteMetadataFiles(o, tt.paths); !compareExitErrors(got, tt.want) {
				t.Errorf("ResetDBSettings.DeleteMetadataFiles() %s want %s", got, tt.want)
			}
			o.Report(t, "ResetDBSettings.DeleteMetadataFiles()", tt.WantedRecording)
		})
	}
}

func TestResetDBSettings_FilterMetadataFiles(t *testing.T) {
	originalPlainFileExists := cmd.PlainFileExists
	defer func() {
		cmd.PlainFileExists = originalPlainFileExists
	}()
	tests := map[string]struct {
		plainFileExists func(string) bool
		rdbs            *cmd.ResetDBSettings
		entries         []fs.FileInfo
		want            []string
	}{
		"no entries": {want: []string{}},
		"mixed entries": {
			plainFileExists: func(s string) bool { return !strings.Contains(s, "dir.") },
			rdbs: &cmd.ResetDBSettings{
				MetadataDir: cmd.StringValue{Value: "metadata"},
				Extension:   cmd.StringValue{Value: ".db"},
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
			cmd.PlainFileExists = tt.plainFileExists
			if got := tt.rdbs.FilterMetadataFiles(tt.entries); !reflect.DeepEqual(got,
				tt.want) {
				t.Errorf("ResetDBSettings.FilterMetadataFiles() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResetDBSettings_CleanUpMetadata(t *testing.T) {
	originalReadDirectory := cmd.ReadDirectory
	originalPlainFileExists := cmd.PlainFileExists
	originalRemove := cmd.Remove
	defer func() {
		cmd.ReadDirectory = originalReadDirectory
		cmd.PlainFileExists = originalPlainFileExists
		cmd.Remove = originalRemove
	}()
	tests := map[string]struct {
		readDirectory   func(output.Bus, string) ([]fs.FileInfo, bool)
		plainFileExists func(string) bool
		remove          func(string) error
		rdbs            *cmd.ResetDBSettings
		stopped         bool
		want            *cmd.ExitError
		output.WantedRecording
	}{
		"did not stop, cannot ignore it": {
			rdbs: &cmd.ResetDBSettings{
				Service:             cmd.StringValue{Value: "musicService"},
				IgnoreServiceErrors: cmd.BoolValue{Value: false},
			},
			stopped: false,
			want:    cmd.NewExitUserError("resetDatabase"),
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
			rdbs:    &cmd.ResetDBSettings{MetadataDir: cmd.StringValue{Value: "metadata"}},
			stopped: true,
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
		"not stopped but ignored, no metadata": {
			readDirectory: func(_ output.Bus, _ string) ([]fs.FileInfo, bool) {
				return nil, true
			},
			rdbs: &cmd.ResetDBSettings{
				MetadataDir:         cmd.StringValue{Value: "metadata"},
				IgnoreServiceErrors: cmd.BoolValue{Value: true},
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
			rdbs: &cmd.ResetDBSettings{
				MetadataDir:         cmd.StringValue{Value: "metadata"},
				IgnoreServiceErrors: cmd.BoolValue{Value: true},
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
			rdbs: &cmd.ResetDBSettings{
				MetadataDir: cmd.StringValue{Value: "metadata"},
				Extension:   cmd.StringValue{Value: ".db"},
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
			cmd.ReadDirectory = tt.readDirectory
			cmd.PlainFileExists = tt.plainFileExists
			cmd.Remove = tt.remove
			o := output.NewRecorder()
			if got := tt.rdbs.CleanUpMetadata(o, tt.stopped); !compareExitErrors(got, tt.want) {
				t.Errorf("ResetDBSettings.CleanUpMetadata() %s want %s", got, tt.want)
			}
			o.Report(t, "ResetDBSettings.CleanUpMetadata()", tt.WantedRecording)
		})
	}
}

func TestResetDBSettings_ResetService(t *testing.T) {
	originalDirty := cmd.Dirty
	originalClearDirty := cmd.ClearDirty
	originalConnect := cmd.Connect
	originalIsElevated := cmd.IsElevated
	defer func() {
		cmd.Dirty = originalDirty
		cmd.ClearDirty = originalClearDirty
		cmd.Connect = originalConnect
		cmd.IsElevated = originalIsElevated
	}()
	cmd.IsElevated = func(_ windows.Token) bool { return false }
	cmd.ClearDirty = func(_ output.Bus) {}
	cmd.Connect = func() (*mgr.Mgr, error) { return nil, fmt.Errorf("access denied") }
	tests := map[string]struct {
		dirty func() bool
		rdbs  *cmd.ResetDBSettings
		want  *cmd.ExitError
		output.WantedRecording
	}{
		"not dirty, no force": {
			dirty: func() bool { return false },
			rdbs:  &cmd.ResetDBSettings{Force: cmd.BoolValue{Value: false}},
			want:  cmd.NewExitUserError("resetDatabase"),
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
			dirty: func() bool { return false },
			rdbs:  &cmd.ResetDBSettings{Force: cmd.BoolValue{Value: true}},
			want:  cmd.NewExitSystemError("resetDatabase"),
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
			dirty: func() bool { return true },
			rdbs:  &cmd.ResetDBSettings{},
			want:  cmd.NewExitSystemError("resetDatabase"),
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
			cmd.Dirty = tt.dirty
			o := output.NewRecorder()
			if got := tt.rdbs.ResetService(o); !compareExitErrors(got, tt.want) {
				t.Errorf("ResetDBSettings.ResetService() got %s want %s", got, tt.want)
			}
			o.Report(t, "ResetDBSettings.ResetService()", tt.WantedRecording)
		})
	}
}

func TestResetDBRun(t *testing.T) {
	cmd.InitGlobals()
	originalBus := cmd.Bus
	originalDirty := cmd.Dirty
	defer func() {
		cmd.Bus = originalBus
		cmd.Dirty = originalDirty
	}()
	cmd.Dirty = func() bool { return false }
	flags := &cmd.SectionFlags{
		SectionName: "resetDatabase",
		Details: map[string]*cmd.FlagDetails{
			"timeout": {
				AbbreviatedName: "t",
				Usage:           fmt.Sprintf("timeout in seconds (minimum %d, maximum %d) for stopping the media player service", 1, 60),
				ExpectedType:    cmd.IntType,
				DefaultValue:    cmd_toolkit.NewIntBounds(1, 10, 60),
			},
			"service": {
				Usage:        "name of the media player service",
				ExpectedType: cmd.StringType,
				DefaultValue: "WMPNetworkSVC",
			},
			"metadataDir": {
				Usage:        "directory where the media player service metadata files are" + " stored",
				ExpectedType: cmd.StringType,
				DefaultValue: filepath.Join("AppData", "Local", "Microsoft", "Media Player"),
			},
			"extension": {
				Usage:        "extension for metadata files",
				ExpectedType: cmd.StringType,
				DefaultValue: ".wmdb",
			},
			"force": {
				AbbreviatedName: "f",
				Usage:           "if set, force a database reset",
				ExpectedType:    cmd.BoolType,
				DefaultValue:    false,
			},
			"ignoreServiceErrors": {
				AbbreviatedName: "i",
				Usage:           "if set, ignore service errors and delete the media player service metadata files",
				ExpectedType:    cmd.BoolType,
				DefaultValue:    false,
			},
		},
	}
	myCommand := &cobra.Command{}
	cmd.AddFlags(output.NewNilBus(), cmd_toolkit.EmptyConfiguration(), myCommand.Flags(),
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
			cmd.Bus = o
			cmd.ResetDBRun(tt.cmd, tt.in1)
			o.Report(t, "ResetDBRun()", tt.WantedRecording)
		})
	}
}

func TestResetDatabaseHelp(t *testing.T) {
	commandUnderTest := cloneCommand(cmd.ResetDatabaseCmd)
	flagMap := map[string]*cmd.FlagDetails{}
	for k, v := range cmd.ResetDatabaseFlags.Details {
		switch k {
		case "metadataDir":
			flagMap[k] = v.Copy()
			flagMap[k].DefaultValue = "[USERPROFILE]/AppData/Local/Microsoft/Media Player"
		default:
			flagMap[k] = v
		}
	}
	flagCopy := &cmd.SectionFlags{SectionName: "resetDatabase", Details: flagMap}
	cmd.AddFlags(output.NewNilBus(), cmd_toolkit.EmptyConfiguration(),
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
			command.Help()
			o.Report(t, "resetDatabase Help()", tt.WantedRecording)
		})
	}
}

func TestUpdateServiceStatus(t *testing.T) {
	type args struct {
		currentStatus  *cmd.ExitError
		proposedStatus *cmd.ExitError
	}
	tests := map[string]struct {
		args
		want *cmd.ExitError
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
				proposedStatus: cmd.NewExitUserError("resetDatabase"),
			},
			want: cmd.NewExitUserError("resetDatabase"),
		},
		"success, program error": {
			args: args{
				currentStatus:  nil,
				proposedStatus: cmd.NewExitProgrammingError("resetDatabase"),
			},
			want: cmd.NewExitProgrammingError("resetDatabase"),
		},
		"success, system error": {
			args: args{
				currentStatus:  nil,
				proposedStatus: cmd.NewExitSystemError("resetDatabase"),
			},
			want: cmd.NewExitSystemError("resetDatabase"),
		},
		"user error, success": {
			args: args{
				currentStatus:  cmd.NewExitUserError("resetDatabase"),
				proposedStatus: nil,
			},
			want: cmd.NewExitUserError("resetDatabase"),
		},
		"user error, user error": {
			args: args{
				currentStatus:  cmd.NewExitUserError("resetDatabase"),
				proposedStatus: cmd.NewExitUserError("resetDatabase"),
			},
			want: cmd.NewExitUserError("resetDatabase"),
		},
		"user error, program error": {
			args: args{
				currentStatus:  cmd.NewExitUserError("resetDatabase"),
				proposedStatus: cmd.NewExitProgrammingError("resetDatabase"),
			},
			want: cmd.NewExitUserError("resetDatabase"),
		},
		"user error, system error": {
			args: args{
				currentStatus:  cmd.NewExitUserError("resetDatabase"),
				proposedStatus: cmd.NewExitSystemError("resetDatabase"),
			},
			want: cmd.NewExitUserError("resetDatabase"),
		},
		"program error, success": {
			args: args{
				currentStatus:  cmd.NewExitProgrammingError("resetDatabase"),
				proposedStatus: nil,
			},
			want: cmd.NewExitProgrammingError("resetDatabase"),
		},
		"program error, user error": {
			args: args{
				currentStatus:  cmd.NewExitProgrammingError("resetDatabase"),
				proposedStatus: cmd.NewExitUserError("resetDatabase"),
			},
			want: cmd.NewExitProgrammingError("resetDatabase"),
		},
		"program error, program error": {
			args: args{
				currentStatus:  cmd.NewExitProgrammingError("resetDatabase"),
				proposedStatus: cmd.NewExitProgrammingError("resetDatabase"),
			},
			want: cmd.NewExitProgrammingError("resetDatabase"),
		},
		"program error, system error": {
			args: args{
				currentStatus:  cmd.NewExitProgrammingError("resetDatabase"),
				proposedStatus: cmd.NewExitSystemError("resetDatabase"),
			},
			want: cmd.NewExitProgrammingError("resetDatabase"),
		},
		"system error, success": {
			args: args{
				currentStatus:  cmd.NewExitSystemError("resetDatabase"),
				proposedStatus: nil,
			},
			want: cmd.NewExitSystemError("resetDatabase"),
		},
		"system error, user error": {
			args: args{
				currentStatus:  cmd.NewExitSystemError("resetDatabase"),
				proposedStatus: cmd.NewExitUserError("resetDatabase"),
			},
			want: cmd.NewExitSystemError("resetDatabase"),
		},
		"system error, program error": {
			args: args{
				currentStatus:  cmd.NewExitSystemError("resetDatabase"),
				proposedStatus: cmd.NewExitProgrammingError("resetDatabase"),
			},
			want: cmd.NewExitSystemError("resetDatabase"),
		},
		"system error, system error": {
			args: args{
				currentStatus:  cmd.NewExitSystemError("resetDatabase"),
				proposedStatus: cmd.NewExitSystemError("resetDatabase"),
			},
			want: cmd.NewExitSystemError("resetDatabase"),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := cmd.UpdateServiceStatus(tt.args.currentStatus, tt.args.proposedStatus); !compareExitErrors(got, tt.want) {
				t.Errorf("UpdateServiceStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMaybeClearDirty(t *testing.T) {
	originalClearDirty := cmd.ClearDirty
	defer func() {
		cmd.ClearDirty = originalClearDirty
	}()
	var clearDirtyCalled bool
	cmd.ClearDirty = func(_ output.Bus) {
		clearDirtyCalled = true
	}
	tests := map[string]struct {
		status *cmd.ExitError
		want   bool
	}{
		"success": {
			status: nil,
			want:   true,
		},
		"user error": {
			status: cmd.NewExitUserError("resetDatabase"),
			want:   false,
		},
		"program error": {
			status: cmd.NewExitProgrammingError("resetDatabase"),
			want:   false,
		},
		"system error": {
			status: cmd.NewExitSystemError("resetDatabase"),
			want:   false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			clearDirtyCalled = false
			cmd.MaybeClearDirty(output.NewNilBus(), tt.status)
			if got := clearDirtyCalled; got != tt.want {
				t.Errorf("MaybeClearDirty = %t want %t", got, tt.want)
			}
		})
	}
}

func TestOutputSystemErrorCause(t *testing.T) {
	originalIsElevated := cmd.IsElevated
	defer func() {
		cmd.IsElevated = originalIsElevated
	}()
	tests := map[string]struct {
		isElevated func(windows.Token) bool
		output.WantedRecording
	}{
		"elevated": {
			isElevated: func(_ windows.Token) bool { return true },
		},
		"ordinary": {
			isElevated: func(_ windows.Token) bool { return false },
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
			cmd.IsElevated = tt.isElevated
			o := output.NewRecorder()
			cmd.OutputSystemErrorCause(o)
			o.Report(t, "OutputSystemErrorCause()", tt.WantedRecording)
		})
	}
}
