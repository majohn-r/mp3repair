package commands

import (
	"flag"
	"fmt"
	"io/fs"
	"mp3/internal"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/majohn-r/output"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

type testService struct {
	wantQueryStatus   svc.Status
	wantQueryError    error
	wantControlStatus svc.Status
	wantControlError  error
}

func (tS *testService) Close() error {
	return nil
}

func (tS *testService) Query() (svc.Status, error) {
	return tS.wantQueryStatus, tS.wantQueryError
}

func (tS *testService) Control(c svc.Cmd) (svc.Status, error) {
	return tS.wantControlStatus, tS.wantControlError
}

type testManager struct {
	m         map[string]service
	wantError error
}

func (tM *testManager) Disconnect() error {
	return nil
}

func (tM *testManager) ListServices() ([]string, error) {
	if tM.wantError != nil {
		return nil, tM.wantError
	}
	var svcs []string
	for k := range tM.m {
		svcs = append(svcs, k)
	}
	sort.Strings(svcs)
	return svcs, nil
}

func (tM *testManager) manager() manager {
	return tM
}

func (tM *testManager) openService(name string) (service, error) {
	if s, ok := tM.m[name]; ok {
		return s, nil
	}
	return nil, fmt.Errorf("access denied")
}

func Test_listAvailableServices(t *testing.T) {
	const fnName = "listAvailableServices()"
	type args struct {
		sM       serviceGateway
		services []string
	}
	tests := map[string]struct {
		args
		output.WantedRecording
	}{
		"no services available": {
			args: args{},
			WantedRecording: output.WantedRecording{
				Console: "The following services are available:\n" +
					"  - none -\n",
			},
		},
		"several services available": {
			args: args{
				sM: &testManager{
					m: map[string]service{
						"svc1": &testService{wantQueryStatus: svc.Status{State: svc.Running}},
						"svc2": &testService{wantQueryError: fmt.Errorf("access denied")},
					},
				},
				services: []string{"svc1", "svc2", "svc3"},
			},
			WantedRecording: output.WantedRecording{
				Console: "The following services are available:\n" +
					"  State \"access denied\":\n" +
					"    \"svc2\"\n" +
					"    \"svc3\"\n" +
					"  State \"running\":\n" +
					"    \"svc1\"\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			listAvailableServices(o, tt.args.sM, tt.args.services)
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
		})
	}
}

func Test_resetDatabase_waitForStop(t *testing.T) {
	const fnName = "resetDatabase.waitForStop()"
	svcName := "test service"
	timeout := 10
	type args struct {
		s         service
		status    svc.Status
		timeout   time.Time
		checkFreq time.Duration
	}
	tests := map[string]struct {
		r *resetDatabase
		args
		wantOk bool
		output.WantedRecording
	}{
		"already stopped": {
			r: &resetDatabase{
				service: &svcName,
			},
			args:   args{status: svc.Status{State: svc.Stopped}},
			wantOk: true,
			WantedRecording: output.WantedRecording{
				Log: "level='info' service='test service' status='stopped' msg='service status'\n",
			},
		},
		"timed out": {
			r: &resetDatabase{service: &svcName, timeout: &timeout},
			args: args{
				status:  svc.Status{State: svc.Running},
				timeout: time.Now().Add(-1 * time.Second),
			},
			WantedRecording: output.WantedRecording{
				Error: "The service \"test service\" could not be stopped within the 10 second timeout.\n",
				Log:   "level='error' error='operation timed out' operation='stop service' service='test service' timeout in seconds='10' msg='service issue'\n",
			},
		},
		"stopped": {
			r: &resetDatabase{service: &svcName, timeout: &timeout},
			args: args{
				s:         &testService{wantQueryStatus: svc.Status{State: svc.Stopped}},
				status:    svc.Status{State: svc.Running},
				timeout:   time.Now().Add(1 * time.Second),
				checkFreq: 1 * time.Millisecond,
			},
			wantOk: true,
			WantedRecording: output.WantedRecording{
				Log: "level='info' service='test service' status='stopped' msg='service status'\n",
			},
		},
		"query failure": {
			r: &resetDatabase{service: &svcName, timeout: &timeout},
			args: args{
				s:         &testService{wantQueryError: fmt.Errorf("access denied")},
				status:    svc.Status{State: svc.Running},
				timeout:   time.Now().Add(1 * time.Second),
				checkFreq: 1 * time.Millisecond,
			},
			WantedRecording: output.WantedRecording{
				Error: "The status for the service \"test service\" cannot be obtained: access denied.\n",
				Log:   "level='error' error='access denied' operation='query service status' service='test service' msg='service issue'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			if gotOk := tt.r.waitForStop(o, tt.args.s, tt.args.status, tt.args.timeout, tt.args.checkFreq); gotOk != tt.wantOk {
				t.Errorf("%s = %v, want %v", fnName, gotOk, tt.wantOk)
			}
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
		})
	}
}

func Test_resetDatabase_stopService(t *testing.T) {
	const fnName = "resetDatabase.stopService()"
	serviceName := "mp3 management service"
	fastTimeout := -1
	type args struct {
		connect func() (serviceGateway, error)
	}
	tests := map[string]struct {
		r    *resetDatabase
		want bool
		args
		output.WantedRecording
	}{
		"connect failure": {
			r:    &resetDatabase{},
			want: true,
			args: args{
				connect: func() (serviceGateway, error) {
					return nil, fmt.Errorf("access denied")
				},
			},
			WantedRecording: output.WantedRecording{
				Error: "The service manager cannot be accessed. Try running the program again as an administrator. Error: access denied.\n",
				Log:   "level='error' error='access denied' operation='connect to service manager' msg='service manager issue'\n",
			},
		},
		"connect successful, failure to open service": {
			r: &resetDatabase{
				service: &serviceName,
			},
			want: true,
			args: args{
				connect: func() (serviceGateway, error) {
					return &testManager{
						m: map[string]service{},
					}, nil
				},
			},
			WantedRecording: output.WantedRecording{
				Console: "The following services are available:\n  - none -\n",
				Error:   "The service \"mp3 management service\" cannot be opened: access denied.\n",
				Log:     "level='error' error='access denied' operation='open service' service='mp3 management service' msg='service issue'\n",
			},
		},
		"service opens but cannot be queried": {
			r: &resetDatabase{
				service: &serviceName,
			},
			want: true,
			args: args{
				connect: func() (serviceGateway, error) {
					return &testManager{
						m: map[string]service{
							serviceName: &testService{
								wantQueryError: fmt.Errorf("query failure"),
							},
						},
					}, nil
				},
			},
			WantedRecording: output.WantedRecording{
				Error: "The status for the service \"mp3 management service\" cannot be obtained: query failure.\n",
				Log:   "level='error' error='query failure' operation='query service status' service='mp3 management service' msg='service issue'\n",
			},
		},
		"service already stopped": {
			r: &resetDatabase{
				service: &serviceName,
			},
			want: true,
			args: args{
				connect: func() (serviceGateway, error) {
					return &testManager{
						m: map[string]service{
							serviceName: &testService{
								wantQueryStatus: svc.Status{
									State: svc.Stopped,
								},
							},
						},
					}, nil
				},
			},
			WantedRecording: output.WantedRecording{
				Log: "level='info' service='mp3 management service' status='stopped' msg='service status'\n",
			},
		},
		"service paused, fails to take stop command": {
			r: &resetDatabase{
				service: &serviceName,
			},
			want: true,
			args: args{
				connect: func() (serviceGateway, error) {
					return &testManager{
						m: map[string]service{
							serviceName: &testService{
								wantQueryStatus: svc.Status{
									State: svc.Paused,
								},
								wantControlError: fmt.Errorf("stop command rejected"),
							},
						},
					}, nil
				},
			},
			WantedRecording: output.WantedRecording{
				Error: "The service \"mp3 management service\" cannot be stopped: stop command rejected.\n",
				Log:   "level='error' error='stop command rejected' operation='stop service' service='mp3 management service' msg='service issue'\n",
			},
		},
		"service running, fails to take stop command": {
			r: &resetDatabase{
				service: &serviceName,
			},
			args: args{
				connect: func() (serviceGateway, error) {
					return &testManager{
						m: map[string]service{
							serviceName: &testService{
								wantQueryStatus: svc.Status{
									State: svc.Running,
								},
								wantControlError: fmt.Errorf("stop command rejected"),
							},
						},
					}, nil
				},
			},
			WantedRecording: output.WantedRecording{
				Error: "The service \"mp3 management service\" cannot be stopped: stop command rejected.\n",
				Log:   "level='error' error='stop command rejected' operation='stop service' service='mp3 management service' msg='service issue'\n",
			},
		},
		"service paused, times out waiting for stop": {
			r: &resetDatabase{
				service: &serviceName,
				timeout: &fastTimeout,
			},
			want: true,
			args: args{
				connect: func() (serviceGateway, error) {
					return &testManager{
						m: map[string]service{
							serviceName: &testService{
								wantQueryStatus: svc.Status{
									State: svc.Paused,
								},
								wantControlStatus: svc.Status{
									State: svc.Paused,
								},
							},
						},
					}, nil
				},
			},
			WantedRecording: output.WantedRecording{
				Error: "The service \"mp3 management service\" could not be stopped within the -1 second timeout.\n",
				Log:   "level='error' error='operation timed out' operation='stop service' service='mp3 management service' timeout in seconds='-1' msg='service issue'\n",
			},
		},
		"service running, times out stopping": {
			r: &resetDatabase{
				service: &serviceName,
				timeout: &fastTimeout,
			},
			args: args{
				connect: func() (serviceGateway, error) {
					return &testManager{
						m: map[string]service{
							serviceName: &testService{
								wantQueryStatus: svc.Status{
									State: svc.Running,
								},
								wantControlStatus: svc.Status{
									State: svc.Running,
								},
							},
						},
					}, nil
				},
			},
			WantedRecording: output.WantedRecording{
				Error: "The service \"mp3 management service\" could not be stopped within the -1 second timeout.\n",
				Log:   "level='error' error='operation timed out' operation='stop service' service='mp3 management service' timeout in seconds='-1' msg='service issue'\n",
			},
		},
		"service paused, successfully stopped": {
			r: &resetDatabase{
				service: &serviceName,
				timeout: &fastTimeout,
			},
			want: true,
			args: args{
				connect: func() (serviceGateway, error) {
					return &testManager{
						m: map[string]service{
							serviceName: &testService{
								wantQueryStatus: svc.Status{
									State: svc.Paused,
								},
								wantControlStatus: svc.Status{
									State: svc.Stopped,
								},
							},
						},
					}, nil
				},
			},
			WantedRecording: output.WantedRecording{
				Log: "level='info' service='mp3 management service' status='stopped' msg='service status'\n",
			},
		},
		"service running, stopped": {
			r: &resetDatabase{
				service: &serviceName,
				timeout: &fastTimeout,
			},
			want: true,
			args: args{
				connect: func() (serviceGateway, error) {
					return &testManager{
						m: map[string]service{
							serviceName: &testService{
								wantQueryStatus: svc.Status{
									State: svc.Running,
								},
								wantControlStatus: svc.Status{
									State: svc.Stopped,
								},
							},
						},
					}, nil
				},
			},
			WantedRecording: output.WantedRecording{
				Log: "level='info' service='mp3 management service' status='stopped' msg='service status'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			if got := tt.r.stopService(o, tt.args.connect); got != tt.want {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
		})
	}
}

func Test_resetDatabase_filterMetadataFiles(t *testing.T) {
	const fnName = "resetDatabase.filterMetadataFiles()"
	testDir := "filterMetadataFiles"
	extension := ".wmdb"
	if err := internal.Mkdir(testDir); err != nil {
		t.Errorf("%s could not create directory %q: %v", fnName, testDir, err)
	}
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
	files, _ := internal.ReadDirectory(output.NewNilBus(), testDir)
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, testDir)
	}()
	type args struct {
		files []fs.DirEntry
	}
	tests := map[string]struct {
		r *resetDatabase
		args
		want []string
	}{
		"complete test": {
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
		"nil test": {
			r: &resetDatabase{
				metadata:  &testDir,
				extension: &extension,
			},
			args: args{files: nil},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.r.filterMetadataFiles(tt.args.files); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func Test_resetDatabase_deleteMetadataFiles(t *testing.T) {
	const fnName = "resetDatabase.deleteMetadataFiles()"
	testDir := "deleteMetadataFiles"
	extension := ".wmdb"
	if err := internal.Mkdir(testDir); err != nil {
		t.Errorf("%s could not create directory %q: %v", fnName, testDir, err)
	}
	for k := 0; k < 8; k++ {
		fileName := fmt.Sprintf("file%d%s", k, extension)
		if err := internal.CreateFileForTesting(testDir, fileName); err != nil {
			t.Errorf("%s failed to create file %q: %v", fnName, fileName, err)
		}
	}
	subDir := filepath.Join(testDir, "file8"+extension)
	if err := internal.Mkdir(subDir); err != nil {
		t.Errorf("%s could not create directory %q: %v", fnName, subDir, err)
	}
	// make file8 impossible to trivially remove
	if err := internal.CreateFileForTesting(subDir, "placeholder.txt"); err != nil {
		t.Errorf("%s failed to create file %q: %v", fnName, "placeholder.txt", err)
	}
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, testDir)
	}()
	type args struct {
		paths []string
	}
	tests := map[string]struct {
		r *resetDatabase
		args
		want bool
		output.WantedRecording
	}{
		"complete test": {
			r: &resetDatabase{metadata: &testDir},
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
			WantedRecording: output.WantedRecording{
				Console: "8 out of 9 metadata files have been deleted from \"deleteMetadataFiles\".\n",
				Error:   "The file \"deleteMetadataFiles\\\\file8.wmdb\" cannot be deleted: remove deleteMetadataFiles\\file8.wmdb: The directory is not empty.\n",
				Log:     "level='error' error='remove deleteMetadataFiles\\file8.wmdb: The directory is not empty.' fileName='deleteMetadataFiles\\file8.wmdb' msg='cannot delete file'\n",
			},
		},
		"nil test": {
			r:    &resetDatabase{metadata: &testDir},
			args: args{paths: nil},
			want: true,
			WantedRecording: output.WantedRecording{
				Console: "0 out of 0 metadata files have been deleted from \"deleteMetadataFiles\".\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			if gotOk := tt.r.deleteMetadataFiles(o, tt.args.paths); gotOk != tt.want {
				t.Errorf("%s gotOK %t want %t", fnName, gotOk, tt.want)
			}
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
		})
	}
}

func Test_resetDatabase_deleteMetadata(t *testing.T) {
	const fnName = "resetDatabase.deleteMetadata()"
	testDir := "deleteMetadata"
	extension := ".wmbd"
	if err := internal.Mkdir(testDir); err != nil {
		t.Errorf("%s cannot create %q: %v", fnName, testDir, err)
	}
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
	fakeDir := fnName
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, testDir)
	}()
	tests := map[string]struct {
		r    *resetDatabase
		want bool
		output.WantedRecording
	}{
		"dir read failure": {
			r: &resetDatabase{metadata: &fakeDir},
			WantedRecording: output.WantedRecording{
				Error: "The directory \"resetDatabase.deleteMetadata()\" cannot be read: open resetDatabase.deleteMetadata(): The system cannot find the file specified.\n",
				Log:   "level='error' directory='resetDatabase.deleteMetadata()' error='open resetDatabase.deleteMetadata(): The system cannot find the file specified.' msg='cannot read directory'\n",
			},
		},
		"empty dir": {
			r:    &resetDatabase{metadata: &emptyDir, extension: &extension},
			want: true,
			WantedRecording: output.WantedRecording{
				Console: "No metadata files were found in \"deleteMetadata\\\\empty\".\n",
				Log:     "level='info' directory='deleteMetadata\\empty' extension='.wmbd' msg='no files found'\n",
			},
		},
		"full dir": {
			r:    &resetDatabase{metadata: &fullDir, extension: &extension},
			want: true,
			WantedRecording: output.WantedRecording{
				Console: "10 out of 10 metadata files have been deleted from \"deleteMetadata\\\\full\".\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			if got := tt.r.deleteMetadata(o); got != tt.want {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
		})
	}
}

func Test_resetDatabase_runCommand(t *testing.T) {
	const fnName = "resetDatabase.runCommand()"
	fastTimeout := -1
	serviceName := "mp3 service"
	testDir := "runCommand"
	nonexistentDir := "resetdatabase_test.go"
	ext := ".wmdb"
	if err := internal.Mkdir(testDir); err != nil {
		t.Errorf("%s cannot create %q: %v", fnName, testDir, err)
	}
	for k := 0; k < 10; k++ {
		fileName := fmt.Sprintf("file%d%s", k, ext)
		if err := internal.CreateFileForTesting(testDir, fileName); err != nil {
			t.Errorf("%s cannot create file %q: %v", fnName, fileName, err)
		}
	}
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, testDir)
	}()
	type args struct {
		connect func() (serviceGateway, error)
	}
	tests := map[string]struct {
		r *resetDatabase
		args
		wantOk bool
		output.WantedRecording
	}{
		"fail to stop service": {
			r: &resetDatabase{
				timeout:   &fastTimeout,
				service:   &serviceName,
				metadata:  &testDir,
				extension: &ext,
			},
			args: args{
				connect: func() (serviceGateway, error) {
					return &testManager{
						m: map[string]service{
							serviceName: &testService{
								wantQueryStatus: svc.Status{
									State: svc.Running,
								},
								wantControlError: fmt.Errorf("stop command rejected"),
							},
						},
					}, nil
				},
			},
			WantedRecording: output.WantedRecording{
				Error: "The service \"mp3 service\" cannot be stopped: stop command rejected.\n",
				Log: "level='info' -extension='.wmdb' -metadata='runCommand' -service='mp3 service' -timeout='-1' command='resetDatabase' msg='executing command'\n" +
					"level='error' error='stop command rejected' operation='stop service' service='mp3 service' msg='service issue'\n",
			},
		},
		"fail to delete metadata": {
			r: &resetDatabase{
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
			WantedRecording: output.WantedRecording{
				Error: "The service manager cannot be accessed. Try running the program again as an administrator. Error: access denied.\n" +
					"The directory \"resetdatabase_test.go\" cannot be read: readdir resetdatabase_test.go: The system cannot find the path specified.\n",
				Log: "level='info' -extension='.wmdb' -metadata='resetdatabase_test.go' -service='mp3 service' -timeout='-1' command='resetDatabase' msg='executing command'\n" +
					"level='error' error='access denied' operation='connect to service manager' msg='service manager issue'\n" +
					"level='error' directory='resetdatabase_test.go' error='readdir resetdatabase_test.go: The system cannot find the path specified.' msg='cannot read directory'\n",
			},
		},
		"success": {
			r: &resetDatabase{
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
			WantedRecording: output.WantedRecording{
				Console: "10 out of 10 metadata files have been deleted from \"runCommand\".\n",
				Error:   "The service manager cannot be accessed. Try running the program again as an administrator. Error: access denied.\n",
				Log: "level='info' -extension='.wmdb' -metadata='runCommand' -service='mp3 service' -timeout='-1' command='resetDatabase' msg='executing command'\n" +
					"level='error' error='access denied' operation='connect to service manager' msg='service manager issue'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			if gotOk := tt.r.runCommand(o, tt.args.connect); gotOk != tt.wantOk {
				t.Errorf("%s = %v, want %v", fnName, gotOk, tt.wantOk)
			}
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
		})
	}
}

func newResetDatabaseCommandForTesting() *resetDatabase {
	r, _ := newResetDatabaseCommand(output.NewNilBus(), internal.EmptyConfiguration(), flag.NewFlagSet("resetDatabase", flag.ContinueOnError))
	return r
}

func Test_resetDatabase_Exec(t *testing.T) {
	const fnName = "resetDatabase.Exec()"
	testDir := "Exec"
	if err := internal.Mkdir(testDir); err != nil {
		t.Errorf("%s error creating %s: %v", fnName, testDir, err)
	}
	savedUserProfile := internal.SaveEnvVarForTesting("USERPROFILE")
	testAppPath := "appPath"
	if err := internal.Mkdir(testAppPath); err != nil {
		t.Errorf("%s error creating %s: %v", fnName, testAppPath, err)
	}
	oldAppPath := internal.SetApplicationPathForTesting(testAppPath)
	userProfile := internal.SavedEnvVar{
		Name:  "USERPROFILE",
		Value: "C:\\Users\\The User",
		Set:   true,
	}
	userProfile.RestoreForTesting()
	// depending on the environment, a connection to the service manager may or
	// may not be possible. Therefore, check whether a connection is possible,
	// and tailor the wanted recordings accordingly.
	var connectionsPossible bool
	if m, err := mgr.Connect(); err != nil {
		connectionsPossible = false
	} else {
		_ = m.Disconnect()
		connectionsPossible = true
	}
	defer func() {
		savedUserProfile.RestoreForTesting()
		internal.DestroyDirectoryForTesting(fnName, testDir)
		internal.DestroyDirectoryForTesting(fnName, testAppPath)
		internal.SetApplicationPathForTesting(oldAppPath)
	}()
	type args struct {
		args []string
	}
	tests := map[string]struct {
		r                 *resetDatabase
		markMetadataDirty bool
		args
		wantOk            bool
		withConnection    output.WantedRecording
		withoutConnection output.WantedRecording
	}{
		"help": {
			r: newResetDatabaseCommandForTesting(),
			args: args{
				args: []string{"-help"},
			},
			withConnection: output.WantedRecording{
				Error: "Usage of resetDatabase:\n" +
					"  -extension extension\n" +
					"    \textension for metadata files (default \".wmdb\")\n" +
					"  -metadata directory\n" +
					"    \tdirectory where the media player service metadata files are stored (default \"C:\\\\Users\\\\The User\\\\AppData\\\\Local\\\\Microsoft\\\\Media Player\")\n" +
					"  -service service\n" +
					"    \tname of the media player service (default \"WMPNetworkSVC\")\n" +
					"  -timeout int\n" +
					"    \ttimeout in seconds (minimum 1, maximum 60) for stopping the media player service (default 10)\n",
				Log: "level='error' arguments='[-help]' msg='flag: help requested'\n",
			},
			withoutConnection: output.WantedRecording{
				Error: "Usage of resetDatabase:\n" +
					"  -extension extension\n" +
					"    \textension for metadata files (default \".wmdb\")\n" +
					"  -metadata directory\n" +
					"    \tdirectory where the media player service metadata files are stored (default \"C:\\\\Users\\\\The User\\\\AppData\\\\Local\\\\Microsoft\\\\Media Player\")\n" +
					"  -service service\n" +
					"    \tname of the media player service (default \"WMPNetworkSVC\")\n" +
					"  -timeout int\n" +
					"    \ttimeout in seconds (minimum 1, maximum 60) for stopping the media player service (default 10)\n",
				Log: "level='error' arguments='[-help]' msg='flag: help requested'\n",
			},
		},
		"runCommand fails but is short-circuited": {
			r: newResetDatabaseCommandForTesting(),
			args: args{
				args: []string{"-metadata", "no such dir"},
			},
			wantOk: true,
			withConnection: output.WantedRecording{
				Console: "Running \"resetDatabase\" is not necessary, as no track files have been edited.\n",
			},
			withoutConnection: output.WantedRecording{
				Console: "Running \"resetDatabase\" is not necessary, as no track files have been edited.\n",
			},
		},
		"runCommand fails": {
			r:                 newResetDatabaseCommandForTesting(),
			markMetadataDirty: true,
			args: args{
				args: []string{"-metadata", "no such dir"},
			},
			withConnection: output.WantedRecording{
				Error: "The directory \"no such dir\" cannot be read: open no such dir: The system cannot find the file specified.\n",
				Log: "level='info' -extension='.wmdb' -metadata='no such dir' -service='WMPNetworkSVC' -timeout='10' command='resetDatabase' msg='executing command'\n" +
					"level='info' service='WMPNetworkSVC' status='stopped' msg='service status'\n" +
					"level='error' directory='no such dir' error='open no such dir: The system cannot find the file specified.' msg='cannot read directory'\n",
			},
			withoutConnection: output.WantedRecording{
				Error: "The service manager cannot be accessed. Try running the program again as an administrator. Error: Access is denied.\n" +
					"The directory \"no such dir\" cannot be read: open no such dir: The system cannot find the file specified.\n",
				Log: "level='info' -extension='.wmdb' -metadata='no such dir' -service='WMPNetworkSVC' -timeout='10' command='resetDatabase' msg='executing command'\n" +
					"level='error' error='Access is denied.' operation='connect to service manager' msg='service manager issue'\n" +
					"level='error' directory='no such dir' error='open no such dir: The system cannot find the file specified.' msg='cannot read directory'\n",
			},
		},
		"success, though no metadata has been written": {
			r: newResetDatabaseCommandForTesting(),
			args: args{
				args: []string{"-metadata", testDir},
			},
			wantOk: true,
			withConnection: output.WantedRecording{
				Console: "Running \"resetDatabase\" is not necessary, as no track files have been edited.\n",
			},
			withoutConnection: output.WantedRecording{
				Console: "Running \"resetDatabase\" is not necessary, as no track files have been edited.\n",
			},
		},
		"success after metadata written": {
			r:                 newResetDatabaseCommandForTesting(),
			markMetadataDirty: true,
			args: args{
				args: []string{"-metadata", testDir},
			},
			wantOk: true,
			withConnection: output.WantedRecording{
				Console: "No metadata files were found in \"Exec\".\n",
				Log: "level='info' -extension='.wmdb' -metadata='Exec' -service='WMPNetworkSVC' -timeout='10' command='resetDatabase' msg='executing command'\n" +
					"level='info' service='WMPNetworkSVC' status='stopped' msg='service status'\n" +
					"level='info' directory='Exec' extension='.wmdb' msg='no files found'\n" +
					"level='info' fileName='appPath\\metadata.dirty' msg='metadata dirty file deleted'\n",
			},
			withoutConnection: output.WantedRecording{
				Console: "No metadata files were found in \"Exec\".\n",
				Error:   "The service manager cannot be accessed. Try running the program again as an administrator. Error: Access is denied.\n",
				Log: "level='info' -extension='.wmdb' -metadata='Exec' -service='WMPNetworkSVC' -timeout='10' command='resetDatabase' msg='executing command'\n" +
					"level='error' error='Access is denied.' operation='connect to service manager' msg='service manager issue'\n" +
					"level='info' directory='Exec' extension='.wmdb' msg='no files found'\n" +
					"level='info' fileName='appPath\\metadata.dirty' msg='metadata dirty file deleted'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if tt.markMetadataDirty {
				markDirty(output.NewNilBus(), resetDatabaseCommandName)
			} else {
				clearDirty(output.NewNilBus())
			}
			o := output.NewRecorder()
			if gotOk := tt.r.Exec(o, tt.args.args); gotOk != tt.wantOk {
				t.Errorf("%s = %v, want %v", fnName, gotOk, tt.wantOk)
			}
			if connectionsPossible {
				if issues, ok := o.Verify(tt.withConnection); !ok {
					for _, issue := range issues {
						t.Errorf("%s %s", fnName, issue)
					}
				}
			} else {
				if issues, ok := o.Verify(tt.withoutConnection); !ok {
					for _, issue := range issues {
						t.Errorf("%s %s", fnName, issue)
					}
				}

			}
			if tt.markMetadataDirty {
				clearDirty(output.NewNilBus())
			}
		})
	}
}

func Test_resetDatabase_openService(t *testing.T) {
	const fnName = "resetDatabase.openService()"
	serviceName := "mp3 management service"
	fastTimeout := -1
	type args struct {
		connect func() (serviceGateway, error)
	}
	tests := map[string]struct {
		r *resetDatabase
		args
		wantM bool
		wantS bool
		output.WantedRecording
	}{
		"fail to connect to manager": {
			r: &resetDatabase{service: &serviceName, timeout: &fastTimeout},
			args: args{
				connect: func() (serviceGateway, error) {
					return nil, fmt.Errorf("access denied")
				},
			},
			WantedRecording: output.WantedRecording{
				Error: "The service manager cannot be accessed. Try running the program again as an administrator. Error: access denied.\n",
				Log:   "level='error' error='access denied' operation='connect to service manager' msg='service manager issue'\n",
			},
		},
		"connected to manager, cannot connect to service or list services": {
			r: &resetDatabase{service: &serviceName, timeout: &fastTimeout},
			args: args{
				connect: func() (serviceGateway, error) {
					return &testManager{
						m:         map[string]service{},
						wantError: fmt.Errorf("cannot list services"),
					}, nil
				},
			},
			WantedRecording: output.WantedRecording{
				Error: "The service \"mp3 management service\" cannot be opened: access denied.\n" +
					"The list of available services cannot be obtained: cannot list services.\n",
				Log: "level='error' error='access denied' operation='open service' service='mp3 management service' msg='service issue'\n" +
					"level='error' error='cannot list services' operation='list services' msg='service manager issue'\n",
			},
		},
		"connected to manager, cannot connect to service, but can list services": {
			r: &resetDatabase{service: &serviceName, timeout: &fastTimeout},
			args: args{
				connect: func() (serviceGateway, error) {
					return &testManager{
						m: map[string]service{
							"other service": &testService{
								wantQueryStatus: svc.Status{
									State: svc.Running,
								},
							},
						},
					}, nil
				},
			},
			WantedRecording: output.WantedRecording{
				Console: "The following services are available:\n" +
					"  State \"running\":\n" +
					"    \"other service\"\n",
				Error: "The service \"mp3 management service\" cannot be opened: access denied.\n",
				Log:   "level='error' error='access denied' operation='open service' service='mp3 management service' msg='service issue'\n",
			},
		},
		"open manager and specified service": {
			r: &resetDatabase{service: &serviceName, timeout: &fastTimeout},
			args: args{
				connect: func() (serviceGateway, error) {
					return &testManager{
						m: map[string]service{
							serviceName: &testService{
								wantQueryStatus: svc.Status{
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
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			returnedM, returnedS := tt.r.openService(o, tt.args.connect)
			gotM := returnedM != nil
			gotS := returnedS != nil
			if gotM != tt.wantM {
				t.Errorf("%s gotM = %t, want %t", fnName, gotM, tt.wantM)
			}
			if gotS != tt.wantS {
				t.Errorf("%s gotS = %t, want %t", fnName, gotS, tt.wantS)
			}
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
		})
	}
}

func Test_newResetDatabaseCommand(t *testing.T) {
	const fnName = "newResetDatabaseCommand()"
	savedFoo := internal.SaveEnvVarForTesting("FOO")
	os.Unsetenv("FOO")
	defer func() {
		savedFoo.RestoreForTesting()
	}()
	type args struct {
		c *internal.Configuration
	}
	tests := map[string]struct {
		args
		wantOk bool
		output.WantedRecording
	}{
		"normal case": {
			args:   args{c: internal.EmptyConfiguration()},
			wantOk: true,
		},
		"bad default timeout": {
			args: args{
				c: internal.CreateConfiguration(output.NewNilBus(), map[string]any{
					"resetDatabase": map[string]any{
						"timeout": "forever",
					},
				}),
			},
			WantedRecording: output.WantedRecording{
				Error: "The configuration file \"defaults.yaml\" contains an invalid value for \"resetDatabase\": invalid value \"forever\" for flag -timeout: parse error.\n",
				Log:   "level='error' error='invalid value \"forever\" for flag -timeout: parse error' section='resetDatabase' msg='invalid content in configuration file'\n",
			},
		},
		"bad default service": {
			args: args{
				c: internal.CreateConfiguration(output.NewNilBus(), map[string]any{
					"resetDatabase": map[string]any{
						"service": "Win$FOO",
					},
				}),
			},
			WantedRecording: output.WantedRecording{
				Error: "The configuration file \"defaults.yaml\" contains an invalid value for \"resetDatabase\": invalid value \"Win$FOO\" for flag -service: missing environment variables: [FOO].\n",
				Log:   "level='error' error='invalid value \"Win$FOO\" for flag -service: missing environment variables: [FOO]' section='resetDatabase' msg='invalid content in configuration file'\n",
			},
		},
		"bad default metadata": {
			args: args{
				c: internal.CreateConfiguration(output.NewNilBus(), map[string]any{
					"resetDatabase": map[string]any{
						"metadata": "%FOO%/data",
					},
				}),
			},
			WantedRecording: output.WantedRecording{
				Error: "The configuration file \"defaults.yaml\" contains an invalid value for \"resetDatabase\": invalid value \"%FOO%/data\" for flag -metadata: missing environment variables: [FOO].\n",
				Log:   "level='error' error='invalid value \"%FOO%/data\" for flag -metadata: missing environment variables: [FOO]' section='resetDatabase' msg='invalid content in configuration file'\n",
			},
		},
		"bad default extension": {
			args: args{
				c: internal.CreateConfiguration(output.NewNilBus(), map[string]any{
					"resetDatabase": map[string]any{
						"extension": ".%FOO%",
					},
				}),
			},
			WantedRecording: output.WantedRecording{
				Error: "The configuration file \"defaults.yaml\" contains an invalid value for \"resetDatabase\": invalid value \".%FOO%\" for flag -extension: missing environment variables: [FOO].\n",
				Log:   "level='error' error='invalid value \".%FOO%\" for flag -extension: missing environment variables: [FOO]' section='resetDatabase' msg='invalid content in configuration file'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			got, gotOk := newResetDatabaseCommand(o, tt.args.c, flag.NewFlagSet("resetDatabase", flag.ContinueOnError))
			if gotOk != tt.wantOk {
				t.Errorf("%s got1 = %v, want %v", fnName, gotOk, tt.wantOk)
			}
			if gotOk && got == nil {
				t.Errorf("%s got nil instance", fnName)
			}
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
		})
	}
}
