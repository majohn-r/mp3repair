/*
Copyright © 2021 Marc Johnson (marc.johnson27591@gmail.com)
*/
package cmd_test

import (
	"fmt"
	"io/fs"
	"mp3/cmd"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	cmd_toolkit "github.com/majohn-r/cmd-toolkit"
	"github.com/majohn-r/output"
	"github.com/spf13/cobra"
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
					"level='error' error='no results to extract flag values from' msg='internal error'\n" +
					"level='error' error='no results to extract flag values from' msg='internal error'\n" +
					"level='error' error='no results to extract flag values from' msg='internal error'\n" +
					"level='error' error='no results to extract flag values from' msg='internal error'\n" +
					"level='error' error='no results to extract flag values from' msg='internal error'\n" +
					"level='error' error='no results to extract flag values from' msg='internal error'\n",
			},
		},
		"good results": {
			values: map[string]*cmd.FlagValue{
				"extension":           {ValueType: cmd.StringType, Value: ".foo"},
				"force":               {ValueType: cmd.BoolType, Value: true},
				"ignoreServiceErrors": {ValueType: cmd.BoolType, Value: true},
				"metadataDir":         {ValueType: cmd.StringType, Value: "metadata"},
				"service":             {ValueType: cmd.StringType, Value: "music service"},
				"timeout":             {ValueType: cmd.IntType, Value: 5},
			},
			want: &cmd.ResetDBSettings{
				Extension:           ".foo",
				Force:               true,
				IgnoreServiceErrors: true,
				MetadataDir:         "metadata",
				Service:             "music service",
				Timeout:             5,
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
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("ProcessResetDBFlags() %s", issue)
				}
			}
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
		wantOk bool
		output.WantedRecording
	}{
		"already timed out": {
			rdbs: &cmd.ResetDBSettings{Service: "my service", Timeout: 10},
			args: args{expiration: time.Now().Add(time.Duration(-1) * time.Second)},
			WantedRecording: output.WantedRecording{
				Error: "The service \"my service\" could not be stopped within the 10 second timeout.\n",
				Log: "" +
					"level='error'" +
					" error='timed out'" +
					" service='my service'" +
					" timeout='10'" +
					" trigger='Stop'" +
					" msg='service issue'\n",
			},
		},
		"query error": {
			rdbs: &cmd.ResetDBSettings{Service: "my service", Timeout: 10},
			args: args{
				s:             newTestService(),
				expiration:    time.Now().Add(time.Duration(1) * time.Second),
				checkInterval: 1 * time.Millisecond,
			},
			WantedRecording: output.WantedRecording{
				Error: "An error occurred while attempting to stop the service \"my service\": no results from query.\n",
				Log: "" +
					"level='error'" +
					" error='no results from query'" +
					" service='my service'" +
					" msg='service query error'\n",
			},
		},
		"stops correctly": {
			rdbs: &cmd.ResetDBSettings{Service: "my service", Timeout: 10},
			args: args{
				s: newTestService(
					svc.Status{State: svc.Running},
					svc.Status{State: svc.Running},
					svc.Status{State: svc.Stopped}),
				expiration:    time.Now().Add(time.Duration(1) * time.Second),
				checkInterval: 1 * time.Millisecond,
			},
			wantOk: true,
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
			if gotOk := tt.rdbs.WaitForStop(o, tt.args.s, tt.args.expiration, tt.args.checkInterval); gotOk != tt.wantOk {
				t.Errorf("ResetDBSettings.WaitForStop() = %v, want %v", gotOk, tt.wantOk)
			}
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("ResetDBSettings.WaitForStop() %s", issue)
				}
			}
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
	if service, ok := tm.serviceMap[name]; ok {
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
		wantOk bool
		output.WantedRecording
	}{
		"defective service": {
			rdbs: &cmd.ResetDBSettings{Service: "my service", Timeout: 10},
			args: args{
				manager: newTestManager(nil, nil),
				service: newTestService(),
			},
			WantedRecording: output.WantedRecording{
				Error: "An error occurred while trying to stop service \"my service\": no results from query.\n",
				Log: "" +
					"level='error' " +
					"error='no results from query' " +
					"service='my service' " +
					"msg='service query error'\n",
			},
		},
		"already stopped": {
			rdbs: &cmd.ResetDBSettings{Service: "my service", Timeout: 10},
			args: args{
				manager: newTestManager(nil, nil),
				service: newTestService(svc.Status{State: svc.Stopped}),
			},
			wantOk: true,
			WantedRecording: output.WantedRecording{
				Log: "" +
					"level='info'" +
					" service='my service'" +
					" msg='service stopped'\n",
			},
		},
		"stopped easily": {
			rdbs: &cmd.ResetDBSettings{Service: "my service", Timeout: 10},
			args: args{
				manager: newTestManager(nil, nil),
				service: newTestService(
					svc.Status{State: svc.Paused},
					svc.Status{State: svc.Stopped}),
			},
			wantOk: true,
			WantedRecording: output.WantedRecording{
				Log: "" +
					"level='info'" +
					" service='my service'" +
					" msg='service stopped'\n",
			},
		},
		"stopped with a little more difficulty": {
			rdbs: &cmd.ResetDBSettings{Service: "my service", Timeout: 10},
			args: args{
				manager: newTestManager(nil, nil),
				service: newTestService(
					svc.Status{State: svc.Paused},
					svc.Status{State: svc.Paused},
					svc.Status{State: svc.Stopped}),
			},
			wantOk: true,
			WantedRecording: output.WantedRecording{
				Log: "" +
					"level='info'" +
					" service='my service'" +
					" msg='service stopped'\n",
			},
		},
		"cannot be stopped": {
			rdbs: &cmd.ResetDBSettings{Service: "my service", Timeout: 10},
			args: args{
				manager: newTestManager(nil, nil),
				service: newTestService(svc.Status{State: svc.Paused}),
			},
			WantedRecording: output.WantedRecording{
				Error: "The service \"my service\" cannot be stopped: no results from query.\n",
				Log: "" +
					"level='error'" +
					" error='no results from query'" +
					" service='my service'" +
					" trigger='Stop'" +
					" msg='service issue'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			if gotOk := tt.rdbs.StopFoundService(o, tt.args.manager, tt.args.service); gotOk != tt.wantOk {
				t.Errorf("ResetDBSettings.StopFoundService() = %v, want %v", gotOk, tt.wantOk)
			}
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("ResetDBSettings.WaitForStop() %s", issue)
				}
			}
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
			m:           map[string][]string{"no results from query": {"some other bad service"}},
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
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("ListServices() %s", issue)
				}
			}
		})
	}
}

func TestResetDBSettings_HandleService(t *testing.T) {
	tests := map[string]struct {
		rdbs    *cmd.ResetDBSettings
		manager cmd.ServiceManager
		wantOk  bool
		output.WantedRecording
	}{
		"defective manager #1": {
			rdbs:    &cmd.ResetDBSettings{Service: "my service", Timeout: 1},
			manager: newTestManager(nil, []string{"my service"}),
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
					" msg='service issue'\n",
			},
		},
		"defective manager #2": {
			rdbs:    &cmd.ResetDBSettings{Service: "my service", Timeout: 1},
			manager: newTestManager(nil, nil),
			WantedRecording: output.WantedRecording{
				Error: "The service \"my service\" cannot be opened: no such service.\n",
				Log: "" +
					"level='error'" +
					" error='no such service'" +
					" service='my service'" +
					" trigger='OpenService'" +
					" msg='service issue'\n" +
					"level='error'" +
					" error='no services'" +
					" trigger='ListServices'" +
					" msg='service issue'\n",
			},
		},
		"defective manager #3": {
			rdbs:    &cmd.ResetDBSettings{Service: "my service", Timeout: 1},
			manager: newTestManager(map[string]*mgr.Service{"my service": nil}, nil),
			WantedRecording: output.WantedRecording{
				Error: "An error occurred while trying to stop service \"my service\": no service.\n",
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
			if gotOk := tt.rdbs.HandleService(o, tt.manager); gotOk != tt.wantOk {
				t.Errorf("ResetDBSettings.HandleService() = %v, want %v", gotOk, tt.wantOk)
			}
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("ResetDBSettings.HandleService() %s", issue)
				}
			}
		})
	}
}

func TestResetDBSettings_StopService(t *testing.T) {
	originalConnect := cmd.Connect
	defer func() {
		cmd.Connect = originalConnect
	}()
	tests := map[string]struct {
		connect func() (*mgr.Mgr, error)
		rdbs    *cmd.ResetDBSettings
		wantOk  bool
		output.WantedRecording
	}{
		"connect fails": {
			connect: func() (*mgr.Mgr, error) {
				return nil, fmt.Errorf("no manager available")
			},
			rdbs: &cmd.ResetDBSettings{Service: "my service", Timeout: 1},
			WantedRecording: output.WantedRecording{
				Error: "" +
					"An attempt to connect with the service manager failed; error is no manager available.\n" +
					"Why?\n" +
					"This often fails due to lack of permissions.\n" +
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
			rdbs: &cmd.ResetDBSettings{Service: "my service", Timeout: 1},
			WantedRecording: output.WantedRecording{
				Error: "The service \"my service\" cannot be opened: nil manager.\n",
				Log: "" +
					"level='error'" +
					" error='nil manager'" +
					" service='my service'" +
					" trigger='OpenService'" +
					" msg='service issue'\n" +
					"level='error'" +
					" error='nil manager'" +
					" trigger='ListServices'" +
					" msg='service issue'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			cmd.Connect = tt.connect
			o := output.NewRecorder()
			if gotOk := tt.rdbs.StopService(o); gotOk != tt.wantOk {
				t.Errorf("ResetDBSettings.StopService() = %v, want %v", gotOk, tt.wantOk)
			}
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("ResetDBSettings.StopService() %s", issue)
				}
			}
		})
	}
}

func TestResetDBSettings_DeleteFiles(t *testing.T) {
	originalRemove := cmd.Remove
	defer func() {
		cmd.Remove = originalRemove
	}()
	tests := map[string]struct {
		remove func(string) error
		rdbs   *cmd.ResetDBSettings
		paths  []string
		output.WantedRecording
	}{
		"no files": {
			rdbs:  &cmd.ResetDBSettings{MetadataDir: "metadata/dir"},
			paths: nil,
		},
		"undeletable files": {
			remove: func(_ string) error { return fmt.Errorf("cannot remove file") },
			rdbs:   &cmd.ResetDBSettings{MetadataDir: "metadata/dir"},
			paths:  []string{"file1", "file2"},
			WantedRecording: output.WantedRecording{
				Console: "0 out of 2 metadata files have been deleted from \"metadata/dir\".\n",
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
			rdbs:   &cmd.ResetDBSettings{MetadataDir: "metadata/dir"},
			paths:  []string{"file1", "file2"},
			WantedRecording: output.WantedRecording{
				Console: "2 out of 2 metadata files have been deleted from \"metadata/dir\".\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			cmd.Remove = tt.remove
			o := output.NewRecorder()
			tt.rdbs.DeleteFiles(o, tt.paths)
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("ResetDBSettings.DeleteFiles() %s", issue)
				}
			}
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
		entries         []fs.DirEntry
		want            []string
	}{
		"no entries": {want: []string{}},
		"mixed entries": {
			plainFileExists: func(s string) bool { return !strings.Contains(s, "dir.") },
			rdbs:            &cmd.ResetDBSettings{MetadataDir: "metadata", Extension: ".db"},
			entries: []fs.DirEntry{
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
			if got := tt.rdbs.FilterMetadataFiles(tt.entries); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ResetDBSettings.FilterMetadataFiles() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResetDBSettings_DeleteMetadataFiles(t *testing.T) {
	originalReadDirectory := cmd.ReadDirectory
	originalPlainFileExists := cmd.PlainFileExists
	originalRemove := cmd.Remove
	defer func() {
		cmd.ReadDirectory = originalReadDirectory
		cmd.PlainFileExists = originalPlainFileExists
		cmd.Remove = originalRemove
	}()
	tests := map[string]struct {
		readDirectory   func(output.Bus, string) ([]fs.DirEntry, bool)
		plainFileExists func(string) bool
		remove          func(string) error
		rdbs            *cmd.ResetDBSettings
		stopped         bool
		output.WantedRecording
	}{
		"did not stop, cannot ignore it": {
			rdbs:    &cmd.ResetDBSettings{Service: "musicService", IgnoreServiceErrors: false},
			stopped: false,
			WantedRecording: output.WantedRecording{
				Error: "" +
					"Metadata files will not be deleted.\n" +
					"Why?\n" +
					"The music service \"musicService\" could not be stopped, and \"--ignoreServiceErrors\" is false.\n" +
					"What to do:\n" +
					"Rerun this command with \"--ignoreServiceErrors\" set to true.\n",
			},
		},
		"stopped, no metadata": {
			readDirectory: func(_ output.Bus, _ string) ([]fs.DirEntry, bool) {
				return nil, true
			},
			rdbs:    &cmd.ResetDBSettings{MetadataDir: "metadata"},
			stopped: true,
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
			readDirectory: func(_ output.Bus, _ string) ([]fs.DirEntry, bool) {
				return nil, true
			},
			rdbs: &cmd.ResetDBSettings{
				MetadataDir:         "metadata",
				IgnoreServiceErrors: true,
			},
			stopped: false,
			WantedRecording: output.WantedRecording{
				Console: "No metadata files were found in \"metadata\".\n",
				Log: "" +
					"level='info'" +
					" directory='metadata'" +
					" extension=''" +
					" msg='no files found'\n",
			},
		},
		"work to do": {
			readDirectory: func(_ output.Bus, _ string) ([]fs.DirEntry, bool) {
				return []fs.DirEntry{newTestFile("foo.db", nil)}, true
			},
			plainFileExists: func(_ string) bool { return true },
			remove:          func(_ string) error { return nil },
			rdbs: &cmd.ResetDBSettings{
				MetadataDir: "metadata",
				Extension:   ".db",
			},
			stopped: true,
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
			tt.rdbs.DeleteMetadataFiles(o, tt.stopped)
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("ResetDBSettings.DeleteMetadataFiles() %s", issue)
				}
			}
		})
	}
}

func TestResetDBSettings_ResetService(t *testing.T) {
	originalDirty := cmd.Dirty
	originalClearDirty := cmd.ClearDirty
	originalConnect := cmd.Connect
	defer func() {
		cmd.Dirty = originalDirty
		cmd.ClearDirty = originalClearDirty
		cmd.Connect = originalConnect
	}()
	cmd.ClearDirty = func(_ output.Bus) {}
	cmd.Connect = func() (*mgr.Mgr, error) { return nil, fmt.Errorf("access denied") }
	tests := map[string]struct {
		dirty func() bool
		rdbs  *cmd.ResetDBSettings
		output.WantedRecording
	}{
		"not dirty, no force": {
			dirty: func() bool { return false },
			rdbs:  &cmd.ResetDBSettings{Force: false},
			WantedRecording: output.WantedRecording{
				Error: "" +
					"The \"resetDatabase\" command has no work to perform.\n" +
					"Why?\n" +
					"The \"mp3\" program has not made any changes to any mp3 files\n" +
					"since the last successful database reset.\n" +
					"What to do:\n" +
					"If you believe the Windows database needs to be reset, run this command\n" +
					"again and use the \"--force\" flag.\n",
			},
		},
		"not dirty, force": {
			dirty: func() bool { return false },
			rdbs:  &cmd.ResetDBSettings{Force: true},
			WantedRecording: output.WantedRecording{
				Error: "" +
					"An attempt to connect with the service manager failed; error is access denied.\n" +
					"Why?\n" +
					"This often fails due to lack of permissions.\n" +
					"What to do:\n" +
					"If you can, try running this command as an administrator.\n" +
					"Metadata files will not be deleted.\n" +
					"Why?\n" +
					"The music service \"\" could not be stopped, and \"--ignoreServiceErrors\" is false.\n" +
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
			WantedRecording: output.WantedRecording{
				Error: "" +
					"An attempt to connect with the service manager failed; error is access denied.\n" +
					"Why?\n" +
					"This often fails due to lack of permissions.\n" +
					"What to do:\n" +
					"If you can, try running this command as an administrator.\n" +
					"Metadata files will not be deleted.\n" +
					"Why?\n" +
					"The music service \"\" could not be stopped, and \"--ignoreServiceErrors\" is false.\n" +
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
			tt.rdbs.ResetService(o)
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("ResetDBSettings.ResetService() %s", issue)
				}
			}
		})
	}
}

func TestResetDBExec(t *testing.T) {
	cmd.InitGlobals()
	originalBus := cmd.Bus
	originalDirty := cmd.Dirty
	defer func() {
		cmd.Bus = originalBus
		cmd.Dirty = originalDirty
	}()
	cmd.Dirty = func() bool { return false }
	flags := cmd.SectionFlags{
		SectionName: "resetDatabase",
		Flags: map[string]*cmd.FlagDetails{
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
				Usage:        "directory where the media player service metadata files are stored",
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
	cmd.AddFlags(output.NewNilBus(), cmd_toolkit.EmptyConfiguration(), myCommand.Flags(), flags, false)
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
					"The \"mp3\" program has not made any changes to any mp3 files\n" +
					"since the last successful database reset.\n" +
					"What to do:\n" +
					"If you believe the Windows database needs to be reset, run this command\n" +
					"again and use the \"--force\" flag.\n",
				Log: "" +
					"level='info'" +
					" --extension='.wmdb'" +
					" --force='false'" +
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
			cmd.ResetDBExec(tt.cmd, tt.in1)
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("ResetDBExec() %s", issue)
				}
			}
		})
	}
}
