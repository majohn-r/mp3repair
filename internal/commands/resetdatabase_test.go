package commands

import (
	"flag"
	"fmt"
	"io/fs"
	"mp3/internal"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
	"time"

	"golang.org/x/sys/windows/svc"
)

type testService struct {
	desiredQueryStatus   svc.Status
	desiredQueryError    error
	desiredControlStatus svc.Status
	desiredControlError  error
}

func (t *testService) Close() error {
	return nil
}

func (t *testService) Query() (svc.Status, error) {
	return t.desiredQueryStatus, t.desiredQueryError
}

func (t *testService) Control(c svc.Cmd) (svc.Status, error) {
	return t.desiredControlStatus, t.desiredControlError
}

type testManager struct {
	serviceMap   map[string]service
	desiredError error
}

func (t *testManager) Disconnect() error {
	return nil
}

func (m *testManager) ListServices() ([]string, error) {
	if m.desiredError != nil {
		return nil, m.desiredError
	}
	var services []string
	for k := range m.serviceMap {
		services = append(services, k)
	}
	sort.Strings(services)
	return services, nil
}

func (m *testManager) manager() manager {
	return m
}

func (m *testManager) openService(name string) (service, error) {
	if s, ok := m.serviceMap[name]; ok {
		return s, nil
	}
	return nil, fmt.Errorf("access denied")
}

func Test_listAvailableServices(t *testing.T) {
	fnName := "listAvailableServices()"
	type args struct {
		sM       serviceGateway
		services []string
	}
	tests := []struct {
		name string
		args
		internal.WantedOutput
	}{
		{
			name: "no services available",
			args: args{},
			WantedOutput: internal.WantedOutput{
				WantConsoleOutput: "The following services are available:\n" +
					"  - none -\n",
			},
		},
		{
			name: "several services available",
			args: args{
				sM: &testManager{
					serviceMap: map[string]service{
						"svc1": &testService{desiredQueryStatus: svc.Status{State: svc.Running}},
						"svc2": &testService{desiredQueryError: fmt.Errorf("access denied")},
					},
				},
				services: []string{"svc1", "svc2", "svc3"},
			},
			WantedOutput: internal.WantedOutput{
				WantConsoleOutput: "The following services are available:\n" +
					"  State \"access denied\":\n" +
					"    \"svc2\"\n" +
					"    \"svc3\"\n" +
					"  State \"running\":\n" +
					"    \"svc1\"\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := internal.NewOutputDeviceForTesting()
			listAvailableServices(o, tt.args.sM, tt.args.services)
			if issues, ok := o.CheckOutput(tt.WantedOutput); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
		})
	}
}

func Test_resetDatabase_waitForStop(t *testing.T) {
	fnName := "resetDatabase.waitForStop()"
	svcName := "test service"
	timeout := 10
	type args struct {
		s         service
		status    svc.Status
		timeout   time.Time
		checkFreq time.Duration
	}
	tests := []struct {
		name   string
		r      *resetDatabase
		args   args
		wantOk bool
		internal.WantedOutput
	}{
		{
			name: "already stopped",
			r: &resetDatabase{
				service: &svcName,
			},
			args:   args{status: svc.Status{State: svc.Stopped}},
			wantOk: true,
			WantedOutput: internal.WantedOutput{
				WantLogOutput: "level='info' service='test service' status='stopped' msg='service status'\n",
			},
		},
		{
			name: "timed out",
			r:    &resetDatabase{service: &svcName, timeout: &timeout},
			args: args{
				status:  svc.Status{State: svc.Running},
				timeout: time.Now().Add(-1 * time.Second),
			},
			WantedOutput: internal.WantedOutput{
				WantErrorOutput: "The service \"test service\" could not be stopped within the 10 second timeout.\n",
				WantLogOutput:   "level='warn' error='operation timed out' operation='stop service' service='test service' timeout in seconds='10' msg='service issue'\n",
			},
		},
		{
			name: "stopped",
			r:    &resetDatabase{service: &svcName, timeout: &timeout},
			args: args{
				s:         &testService{desiredQueryStatus: svc.Status{State: svc.Stopped}},
				status:    svc.Status{State: svc.Running},
				timeout:   time.Now().Add(1 * time.Second),
				checkFreq: 1 * time.Millisecond,
			},
			wantOk: true,
			WantedOutput: internal.WantedOutput{
				WantLogOutput: "level='info' service='test service' status='stopped' msg='service status'\n",
			},
		},
		{
			name: "query failure",
			r:    &resetDatabase{service: &svcName, timeout: &timeout},
			args: args{
				s:         &testService{desiredQueryError: fmt.Errorf("access denied")},
				status:    svc.Status{State: svc.Running},
				timeout:   time.Now().Add(1 * time.Second),
				checkFreq: 1 * time.Millisecond,
			},
			WantedOutput: internal.WantedOutput{
				WantErrorOutput: "The status for the service \"test service\" cannot be obtained: access denied.\n",
				WantLogOutput:   "level='warn' error='access denied' operation='query service status' service='test service' msg='service issue'\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := internal.NewOutputDeviceForTesting()
			if gotOk := tt.r.waitForStop(o, tt.args.s, tt.args.status, tt.args.timeout, tt.args.checkFreq); gotOk != tt.wantOk {
				t.Errorf("%s = %v, want %v", fnName, gotOk, tt.wantOk)
			}
			if issues, ok := o.CheckOutput(tt.WantedOutput); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
		})
	}
}

func Test_resetDatabase_stopService(t *testing.T) {
	fnName := "resetDatabase.stopService()"
	serviceName := "mp3 management service"
	fastTimeout := -1
	type args struct {
		connect func() (serviceGateway, error)
	}
	tests := []struct {
		name string
		r    *resetDatabase
		want bool
		args
		internal.WantedOutput
	}{
		{
			name: "connect failure",
			r:    &resetDatabase{},
			want: true,
			args: args{
				connect: func() (serviceGateway, error) {
					return nil, fmt.Errorf("access denied")
				},
			},
			WantedOutput: internal.WantedOutput{
				WantErrorOutput: "The service manager cannot be accessed. Try running the program again as an administrator. Error: access denied.\n",
				WantLogOutput:   "level='warn' error='access denied' operation='connect to service manager' msg='service manager issue'\n",
			},
		},
		{
			name: "connect successful, failure to open service",
			r: &resetDatabase{
				service: &serviceName,
			},
			want: true,
			args: args{
				connect: func() (serviceGateway, error) {
					return &testManager{
						serviceMap: map[string]service{},
					}, nil
				},
			},
			WantedOutput: internal.WantedOutput{
				WantConsoleOutput: "The service \"mp3 management service\" cannot be opened: access denied.\n" +
					"The following services are available:\n  - none -\n",
				WantLogOutput: "level='warn' error='access denied' operation='open service' service='mp3 management service' msg='service issue'\n",
			},
		},
		{
			name: "service opens but cannot be queried",
			r: &resetDatabase{
				service: &serviceName,
			},
			want: true,
			args: args{
				connect: func() (serviceGateway, error) {
					return &testManager{
						serviceMap: map[string]service{
							serviceName: &testService{
								desiredQueryError: fmt.Errorf("query failure"),
							},
						},
					}, nil
				},
			},
			WantedOutput: internal.WantedOutput{
				WantErrorOutput: "The status for the service \"mp3 management service\" cannot be obtained: query failure.\n",
				WantLogOutput:   "level='warn' error='query failure' operation='query service status' service='mp3 management service' msg='service issue'\n",
			},
		},
		{
			name: "service already stopped",
			r: &resetDatabase{
				service: &serviceName,
			},
			want: true,
			args: args{
				connect: func() (serviceGateway, error) {
					return &testManager{
						serviceMap: map[string]service{
							serviceName: &testService{
								desiredQueryStatus: svc.Status{
									State: svc.Stopped,
								},
							},
						},
					}, nil
				},
			},
			WantedOutput: internal.WantedOutput{
				WantLogOutput: "level='info' service='mp3 management service' status='stopped' msg='service status'\n",
			},
		},
		{
			name: "service paused, fails to take stop command",
			r: &resetDatabase{
				service: &serviceName,
			},
			want: true,
			args: args{
				connect: func() (serviceGateway, error) {
					return &testManager{
						serviceMap: map[string]service{
							serviceName: &testService{
								desiredQueryStatus: svc.Status{
									State: svc.Paused,
								},
								desiredControlError: fmt.Errorf("stop command rejected"),
							},
						},
					}, nil
				},
			},
			WantedOutput: internal.WantedOutput{
				WantErrorOutput: "The service \"mp3 management service\" cannot be stopped: stop command rejected.\n",
				WantLogOutput:   "level='warn' error='stop command rejected' operation='stop service' service='mp3 management service' msg='service issue'\n",
			},
		},
		{
			name: "service running, fails to take stop command",
			r: &resetDatabase{
				service: &serviceName,
			},
			args: args{
				connect: func() (serviceGateway, error) {
					return &testManager{
						serviceMap: map[string]service{
							serviceName: &testService{
								desiredQueryStatus: svc.Status{
									State: svc.Running,
								},
								desiredControlError: fmt.Errorf("stop command rejected"),
							},
						},
					}, nil
				},
			},
			WantedOutput: internal.WantedOutput{
				WantErrorOutput: "The service \"mp3 management service\" cannot be stopped: stop command rejected.\n",
				WantLogOutput:   "level='warn' error='stop command rejected' operation='stop service' service='mp3 management service' msg='service issue'\n",
			},
		},
		{
			name: "service paused, times out waiting for stop",
			r: &resetDatabase{
				service: &serviceName,
				timeout: &fastTimeout,
			},
			want: true,
			args: args{
				connect: func() (serviceGateway, error) {
					return &testManager{
						serviceMap: map[string]service{
							serviceName: &testService{
								desiredQueryStatus: svc.Status{
									State: svc.Paused,
								},
								desiredControlStatus: svc.Status{
									State: svc.Paused,
								},
							},
						},
					}, nil
				},
			},
			WantedOutput: internal.WantedOutput{
				WantErrorOutput: "The service \"mp3 management service\" could not be stopped within the -1 second timeout.\n",
				WantLogOutput:   "level='warn' error='operation timed out' operation='stop service' service='mp3 management service' timeout in seconds='-1' msg='service issue'\n",
			},
		},
		{
			name: "service running, times out stopping",
			r: &resetDatabase{
				service: &serviceName,
				timeout: &fastTimeout,
			},
			args: args{
				connect: func() (serviceGateway, error) {
					return &testManager{
						serviceMap: map[string]service{
							serviceName: &testService{
								desiredQueryStatus: svc.Status{
									State: svc.Running,
								},
								desiredControlStatus: svc.Status{
									State: svc.Running,
								},
							},
						},
					}, nil
				},
			},
			WantedOutput: internal.WantedOutput{
				WantErrorOutput: "The service \"mp3 management service\" could not be stopped within the -1 second timeout.\n",
				WantLogOutput:   "level='warn' error='operation timed out' operation='stop service' service='mp3 management service' timeout in seconds='-1' msg='service issue'\n",
			},
		},
		{
			name: "service paused, successfully stopped",
			r: &resetDatabase{
				service: &serviceName,
				timeout: &fastTimeout,
			},
			want: true,
			args: args{
				connect: func() (serviceGateway, error) {
					return &testManager{
						serviceMap: map[string]service{
							serviceName: &testService{
								desiredQueryStatus: svc.Status{
									State: svc.Paused,
								},
								desiredControlStatus: svc.Status{
									State: svc.Stopped,
								},
							},
						},
					}, nil
				},
			},
			WantedOutput: internal.WantedOutput{
				WantLogOutput: "level='info' service='mp3 management service' status='stopped' msg='service status'\n",
			},
		},
		{
			name: "service running, stopped",
			r: &resetDatabase{
				service: &serviceName,
				timeout: &fastTimeout,
			},
			want: true,
			args: args{
				connect: func() (serviceGateway, error) {
					return &testManager{
						serviceMap: map[string]service{
							serviceName: &testService{
								desiredQueryStatus: svc.Status{
									State: svc.Running,
								},
								desiredControlStatus: svc.Status{
									State: svc.Stopped,
								},
							},
						},
					}, nil
				},
			},
			WantedOutput: internal.WantedOutput{
				WantLogOutput: "level='info' service='mp3 management service' status='stopped' msg='service status'\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := internal.NewOutputDeviceForTesting()
			if got := tt.r.stopService(o, tt.args.connect); got != tt.want {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
			if issues, ok := o.CheckOutput(tt.WantedOutput); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
		})
	}
}

func Test_resetDatabase_filterMetadataFiles(t *testing.T) {
	fnName := "resetDatabase.filterMetadataFiles()"
	testDir := "filterMetadataFiles"
	extension := ".wmdb"
	if err := internal.Mkdir(testDir); err != nil {
		t.Errorf("%s could not create directory %q: %v", fnName, testDir, err)
	}
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, testDir)
	}()
	for k := 0; k < 8; k++ {
		var fileName string
		if k%2 == 0 {
			fileName = fmt.Sprintf("file%d%s", k, extension)
		} else {
			fileName = fmt.Sprintf("file%d%s", k, extension+"1")
		}
		if err := internal.CreateFileForTesting(testDir, fileName); err != nil {
			t.Errorf("%s failed to create file %q: %v", fnName, fileName, err)
		}
	}
	subDir := "file8" + extension
	if err := internal.Mkdir(filepath.Join(testDir, subDir)); err != nil {
		t.Errorf("%s could not create directory %q: %v", fnName, subDir, err)
	}
	files, _ := internal.ReadDirectory(internal.NewOutputDeviceForTesting(), testDir)
	type args struct {
		files []fs.FileInfo
	}
	tests := []struct {
		name string
		r    *resetDatabase
		args
		want []string
	}{
		{
			name: "complete test",
			r: &resetDatabase{
				metadata:  &testDir,
				extension: &extension,
			},
			args: args{files: files},
			want: []string{
				filepath.Join(testDir, "file0.wmdb"),
				filepath.Join(testDir, "file2.wmdb"),
				filepath.Join(testDir, "file4.wmdb"),
				filepath.Join(testDir, "file6.wmdb"),
			},
		},
		{
			name: "nil test",
			r: &resetDatabase{
				metadata:  &testDir,
				extension: &extension,
			},
			args: args{files: nil},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.r.filterMetadataFiles(tt.args.files); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func Test_resetDatabase_deleteMetadataFiles(t *testing.T) {
	fnName := "resetDatabase.deleteMetadataFiles()"
	testDir := "deleteMetadataFiles"
	extension := ".wmdb"
	if err := internal.Mkdir(testDir); err != nil {
		t.Errorf("%s could not create directory %q: %v", fnName, testDir, err)
	}
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, testDir)
	}()
	for k := 0; k < 8; k++ {
		fileName := fmt.Sprintf("file%d%s", k, extension)
		if err := internal.CreateFileForTesting(testDir, fileName); err != nil {
			t.Errorf("%s failed to create file %q: %v", fnName, fileName, err)
		}
	}
	subDir := filepath.Join(testDir, "file8"+extension)
	if err := internal.Mkdir(filepath.Join(subDir)); err != nil {
		t.Errorf("%s could not create directory %q: %v", fnName, subDir, err)
	}
	// make file8 impossible to trivially remove
	if err := internal.CreateFileForTesting(subDir, "placeholder.txt"); err != nil {
		t.Errorf("%s failed to create file %q: %v", fnName, "placeholder.txt", err)
	}
	type args struct {
		paths []string
	}
	tests := []struct {
		name string
		r    *resetDatabase
		args
		want bool
		internal.WantedOutput
	}{
		{
			name: "complete test",
			r:    &resetDatabase{metadata: &testDir},
			args: args{paths: []string{
				filepath.Join(testDir, "file0.wmdb"),
				filepath.Join(testDir, "file1.wmdb"),
				filepath.Join(testDir, "file2.wmdb"),
				filepath.Join(testDir, "file3.wmdb"),
				filepath.Join(testDir, "file4.wmdb"),
				filepath.Join(testDir, "file5.wmdb"),
				filepath.Join(testDir, "file6.wmdb"),
				filepath.Join(testDir, "file7.wmdb"),
				filepath.Join(testDir, "file8.wmdb"),
			}},
			want: false,
			WantedOutput: internal.WantedOutput{
				WantConsoleOutput: "8 out of 9 metadata files have been deleted from \"deleteMetadataFiles\".\n",
				WantErrorOutput:   "The file \"deleteMetadataFiles\\\\file8.wmdb\" cannot be deleted: remove deleteMetadataFiles\\file8.wmdb: The directory is not empty.\n",
				WantLogOutput:     "level='warn' error='remove deleteMetadataFiles\\file8.wmdb: The directory is not empty.' fileName='deleteMetadataFiles\\file8.wmdb' msg='cannot delete file'\n",
			},
		},
		{
			name: "nil test",
			r:    &resetDatabase{metadata: &testDir},
			args: args{paths: nil},
			want: true,
			WantedOutput: internal.WantedOutput{
				WantConsoleOutput: "0 out of 0 metadata files have been deleted from \"deleteMetadataFiles\".\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := internal.NewOutputDeviceForTesting()
			if gotOk := tt.r.deleteMetadataFiles(o, tt.args.paths); gotOk != tt.want {
				t.Errorf("%s gotOK %t want %t", fnName, gotOk, tt.want)
			}
			if issues, ok := o.CheckOutput(tt.WantedOutput); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
		})
	}
}

func Test_resetDatabase_deleteMetadata(t *testing.T) {
	fnName := "resetDatabase.deleteMetadata()"
	testDir := "deleteMetadata"
	extension := ".wmbd"
	if err := internal.Mkdir(testDir); err != nil {
		t.Errorf("%s cannot create %q: %v", fnName, testDir, err)
	}
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, testDir)
	}()
	emptyDir := filepath.Join(testDir, "empty")
	if err := internal.Mkdir(emptyDir); err != nil {
		t.Errorf("%s cannot create %q: %v", fnName, emptyDir, err)
	}
	fullDir := filepath.Join(testDir, "full")
	if err := internal.Mkdir(fullDir); err != nil {
		t.Errorf("%s cannot create %q: %v", fnName, fullDir, err)
	}
	for k := 0; k < 10; k++ {
		fileName := fmt.Sprintf("file%d%s", k, extension)
		if err := internal.CreateFileForTesting(fullDir, fileName); err != nil {
			t.Errorf("%s cannot create file %q: %v", fnName, fileName, err)
		}
	}
	tests := []struct {
		name string
		r    *resetDatabase
		want bool
		internal.WantedOutput
	}{
		{
			name: "dir read failure",
			r:    &resetDatabase{metadata: &fnName},
			WantedOutput: internal.WantedOutput{
				WantErrorOutput: "The directory \"resetDatabase.deleteMetadata()\" cannot be read: open resetDatabase.deleteMetadata(): The system cannot find the file specified.\n",
				WantLogOutput:   "level='warn' directory='resetDatabase.deleteMetadata()' error='open resetDatabase.deleteMetadata(): The system cannot find the file specified.' msg='cannot read directory'\n",
			},
		},
		{
			name: "empty dir",
			r:    &resetDatabase{metadata: &emptyDir, extension: &extension},
			want: true,
			WantedOutput: internal.WantedOutput{
				WantConsoleOutput: "No metadata files were found in \"deleteMetadata\\\\empty\".\n",
				WantLogOutput:     "level='info' directory='deleteMetadata\\empty' file extension='.wmbd' msg='no files found'\n",
			},
		},
		{
			name: "full dir",
			r:    &resetDatabase{metadata: &fullDir, extension: &extension},
			want: true,
			WantedOutput: internal.WantedOutput{
				WantConsoleOutput: "10 out of 10 metadata files have been deleted from \"deleteMetadata\\\\full\".\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := internal.NewOutputDeviceForTesting()
			if got := tt.r.deleteMetadata(o); got != tt.want {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
			if issues, ok := o.CheckOutput(tt.WantedOutput); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
		})
	}
}

func Test_resetDatabase_runCommand(t *testing.T) {
	fnName := "resetDatabase.runCommand()"
	fastTimeout := -1
	serviceName := "mp3 service"
	testDir := "runCommand"
	nonexistentDir := "resetdatabase_test.go"
	ext := ".wmdb"
	if err := internal.Mkdir(testDir); err != nil {
		t.Errorf("%s cannot create %q: %v", fnName, testDir, err)
	}
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, testDir)
	}()
	for k := 0; k < 10; k++ {
		fileName := fmt.Sprintf("file%d%s", k, ext)
		if err := internal.CreateFileForTesting(testDir, fileName); err != nil {
			t.Errorf("%s cannot create file %q: %v", fnName, fileName, err)
		}
	}
	type args struct {
		connect func() (serviceGateway, error)
	}
	tests := []struct {
		name string
		r    *resetDatabase
		args
		wantOk bool
		internal.WantedOutput
	}{
		{
			name: "fail to stop service",
			r: &resetDatabase{
				n:         "resetDatabase",
				timeout:   &fastTimeout,
				service:   &serviceName,
				metadata:  &testDir,
				extension: &ext,
			},
			args: args{
				connect: func() (serviceGateway, error) {
					return &testManager{
						serviceMap: map[string]service{
							serviceName: &testService{
								desiredQueryStatus: svc.Status{
									State: svc.Running,
								},
								desiredControlError: fmt.Errorf("stop command rejected"),
							},
						},
					}, nil
				},
			},
			WantedOutput: internal.WantedOutput{
				WantErrorOutput: "The service \"mp3 service\" cannot be stopped: stop command rejected.\n",
				WantLogOutput: "level='info' -extension='.wmdb' -metadata='runCommand' -service='mp3 service' -timeout='-1' command='resetDatabase' msg='executing command'\n" +
					"level='warn' error='stop command rejected' operation='stop service' service='mp3 service' msg='service issue'\n",
			},
		},
		{
			name: "fail to delete metadata",
			r: &resetDatabase{
				n:         "resetDatabase",
				timeout:   &fastTimeout,
				service:   &serviceName,
				metadata:  &nonexistentDir,
				extension: &ext,
			},
			args: args{
				connect: func() (serviceGateway, error) {
					return nil, fmt.Errorf("access denied")
				},
			},
			WantedOutput: internal.WantedOutput{
				WantErrorOutput: "The service manager cannot be accessed. Try running the program again as an administrator. Error: access denied.\n" +
					"The directory \"resetdatabase_test.go\" cannot be read: readdir resetdatabase_test.go: The system cannot find the path specified.\n",
				WantLogOutput: "level='info' -extension='.wmdb' -metadata='resetdatabase_test.go' -service='mp3 service' -timeout='-1' command='resetDatabase' msg='executing command'\n" +
					"level='warn' error='access denied' operation='connect to service manager' msg='service manager issue'\n" +
					"level='warn' directory='resetdatabase_test.go' error='readdir resetdatabase_test.go: The system cannot find the path specified.' msg='cannot read directory'\n",
			},
		},
		{
			name: "success",
			r: &resetDatabase{
				n:         "resetDatabase",
				timeout:   &fastTimeout,
				service:   &serviceName,
				metadata:  &testDir,
				extension: &ext,
			},
			args: args{
				connect: func() (serviceGateway, error) {
					return nil, fmt.Errorf("access denied")
				},
			},
			wantOk: true,
			WantedOutput: internal.WantedOutput{
				WantConsoleOutput: "10 out of 10 metadata files have been deleted from \"runCommand\".\n",
				WantErrorOutput:   "The service manager cannot be accessed. Try running the program again as an administrator. Error: access denied.\n",
				WantLogOutput: "level='info' -extension='.wmdb' -metadata='runCommand' -service='mp3 service' -timeout='-1' command='resetDatabase' msg='executing command'\n" +
					"level='warn' error='access denied' operation='connect to service manager' msg='service manager issue'\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := internal.NewOutputDeviceForTesting()
			if gotOk := tt.r.runCommand(o, tt.args.connect); gotOk != tt.wantOk {
				t.Errorf("%s = %v, want %v", fnName, gotOk, tt.wantOk)
			}
			if issues, ok := o.CheckOutput(tt.WantedOutput); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
		})
	}
}

func Test_resetDatabase_Exec(t *testing.T) {
	fnName := "resetDatabase.Exec()"
	testDir := "Exec"
	if err := internal.Mkdir(testDir); err != nil {
		t.Errorf("%s error creating %s: %v", fnName, testDir, err)
	}
	savedUserProfile := internal.SaveEnvVarForTesting("Userprofile")
	defer func() {
		savedUserProfile.RestoreForTesting()
		internal.DestroyDirectoryForTesting(fnName, testDir)
	}()
	userProfile := internal.SavedEnvVar{
		Name:  "Userprofile",
		Value: "C:\\Users\\The User",
		Set:   true,
	}
	userProfile.RestoreForTesting()
	type args struct {
		args []string
	}
	tests := []struct {
		name string
		r    *resetDatabase
		args
		wantOk bool
		internal.WantedOutput
	}{
		{
			name: "bad arguments",
			r: newResetDatabaseCommand(
				internal.EmptyConfiguration(),
				flag.NewFlagSet("resetDatabase", flag.ContinueOnError)),
			args: args{
				args: []string{"-help"},
			},
			WantedOutput: internal.WantedOutput{
				WantErrorOutput: "Usage of resetDatabase:\n" +
					"  -extension string\n" +
					"    \textension for metadata files (default \".wmdb\")\n" +
					"  -metadata string\n" +
					"    \tdirectory where the media player service metadata files are stored (default \"C:\\\\Users\\\\The User\\\\AppData\\\\Local\\\\Microsoft\\\\Media Player\")\n" +
					"  -service string\n" +
					"    \tname of the media player service (default \"WMPNetworkSVC\")\n" +
					"  -timeout int\n" +
					"    \ttimeout in seconds (minimum 1, maximum 60) for stopping the media player service (default 10)\n",
				WantLogOutput: "level='error' arguments='[-help]' msg='flag: help requested'\n",
			},
		},
		{
			name: "runCommand fails",
			r: newResetDatabaseCommand(
				internal.EmptyConfiguration(),
				flag.NewFlagSet("resetDatabase", flag.ContinueOnError)),
			args: args{
				args: []string{"-metadata", "no such dir"},
			},
			WantedOutput: internal.WantedOutput{
				WantErrorOutput: "The service manager cannot be accessed. Try running the program again as an administrator. Error: Access is denied.\n" +
					"The directory \"no such dir\" cannot be read: open no such dir: The system cannot find the file specified.\n",
				WantLogOutput: "level='info' -extension='.wmdb' -metadata='no such dir' -service='WMPNetworkSVC' -timeout='10' command='resetDatabase' msg='executing command'\n" +
					"level='warn' error='Access is denied.' operation='connect to service manager' msg='service manager issue'\n" +
					"level='warn' directory='no such dir' error='open no such dir: The system cannot find the file specified.' msg='cannot read directory'\n",
			},
		},
		{
			name: "success",
			r: newResetDatabaseCommand(
				internal.EmptyConfiguration(),
				flag.NewFlagSet("resetDatabase", flag.ContinueOnError)),
			args: args{
				args: []string{"-metadata", testDir},
			},
			wantOk: true,
			WantedOutput: internal.WantedOutput{
				WantConsoleOutput: "No metadata files were found in \"Exec\".\n",
				WantErrorOutput:   "The service manager cannot be accessed. Try running the program again as an administrator. Error: Access is denied.\n",
				WantLogOutput: "level='info' -extension='.wmdb' -metadata='Exec' -service='WMPNetworkSVC' -timeout='10' command='resetDatabase' msg='executing command'\n" +
					"level='warn' error='Access is denied.' operation='connect to service manager' msg='service manager issue'\n" +
					"level='info' directory='Exec' file extension='.wmdb' msg='no files found'\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := internal.NewOutputDeviceForTesting()
			if gotOk := tt.r.Exec(o, tt.args.args); gotOk != tt.wantOk {
				t.Errorf("%s = %v, want %v", fnName, gotOk, tt.wantOk)
			}
			if issues, ok := o.CheckOutput(tt.WantedOutput); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
		})
	}
}

func Test_resetDatabase_openService(t *testing.T) {
	fnName := "resetDatabase.openService()"
	serviceName := "mp3 management service"
	fastTimeout := -1
	type args struct {
		connect func() (serviceGateway, error)
	}
	tests := []struct {
		name string
		r    *resetDatabase
		args
		wantM bool
		wantS bool
		internal.WantedOutput
	}{
		{
			name: "fail to connect to manager",
			r:    &resetDatabase{service: &serviceName, timeout: &fastTimeout},
			args: args{
				connect: func() (serviceGateway, error) {
					return nil, fmt.Errorf("access denied")
				},
			},
			WantedOutput: internal.WantedOutput{
				WantErrorOutput: "The service manager cannot be accessed. Try running the program again as an administrator. Error: access denied.\n",
				WantLogOutput:   "level='warn' error='access denied' operation='connect to service manager' msg='service manager issue'\n",
			},
		},
		{
			name: "connected to manager, cannot connect to service or list services",
			r:    &resetDatabase{service: &serviceName, timeout: &fastTimeout},
			args: args{
				connect: func() (serviceGateway, error) {
					return &testManager{
						serviceMap:   map[string]service{},
						desiredError: fmt.Errorf("cannot list services"),
					}, nil
				},
			},
			WantedOutput: internal.WantedOutput{
				WantConsoleOutput: "The service \"mp3 management service\" cannot be opened: access denied.\n",
				WantErrorOutput:   "The list of available services cannot be obtained: cannot list services.\n",
				WantLogOutput: "level='warn' error='access denied' operation='open service' service='mp3 management service' msg='service issue'\n" +
					"level='warn' error='cannot list services' operation='list services' msg='service manager issue'\n",
			},
		},
		{
			name: "connected to manager, cannot connect to service, but can list services",
			r:    &resetDatabase{service: &serviceName, timeout: &fastTimeout},
			args: args{
				connect: func() (serviceGateway, error) {
					return &testManager{
						serviceMap: map[string]service{
							"other service": &testService{
								desiredQueryStatus: svc.Status{
									State: svc.Running,
								},
							},
						},
					}, nil
				},
			},
			WantedOutput: internal.WantedOutput{
				WantConsoleOutput: "The service \"mp3 management service\" cannot be opened: access denied.\n" +
					"The following services are available:\n" +
					"  State \"running\":\n" +
					"    \"other service\"\n",
				WantLogOutput: "level='warn' error='access denied' operation='open service' service='mp3 management service' msg='service issue'\n",
			},
		},
		{
			name: "open manager and specified service",
			r:    &resetDatabase{service: &serviceName, timeout: &fastTimeout},
			args: args{
				connect: func() (serviceGateway, error) {
					return &testManager{
						serviceMap: map[string]service{
							serviceName: &testService{
								desiredQueryStatus: svc.Status{
									State: svc.Running,
								},
							},
						},
					}, nil
				},
			},
			wantM: true,
			wantS: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := internal.NewOutputDeviceForTesting()
			returnedM, returnedS := tt.r.openService(o, tt.args.connect)
			gotM := returnedM != nil
			gotS := returnedS != nil
			if gotM != tt.wantM {
				t.Errorf("%s gotM = %t, want %t", fnName, gotM, tt.wantM)
			}
			if gotS != tt.wantS {
				t.Errorf("%s gotS = %t, want %t", fnName, gotS, tt.wantS)
			}
			if issues, ok := o.CheckOutput(tt.WantedOutput); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
		})
	}
}
