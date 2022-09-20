package internal

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestCreateAlbumNameForTesting(t *testing.T) {
	fnName := "CreateAlbumNameForTesting()"
	type args struct {
		k int
	}
	tests := []struct {
		name string
		args
		want string
	}{
		{name: "negative value", args: args{k: -1}, want: "Test Album -1"},
		{name: "zero", args: args{k: 0}, want: "Test Album 0"},
		{name: "positive value", args: args{k: 1}, want: "Test Album 1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CreateAlbumNameForTesting(tt.args.k); got != tt.want {
				t.Errorf("%s = %q, want %q", fnName, got, tt.want)
			}
		})
	}
}

func TestCreateArtistNameForTesting(t *testing.T) {
	fnName := "CreateArtistNameForTesting()"
	type args struct {
		k int
	}
	tests := []struct {
		name string
		args
		want string
	}{
		{name: "negative value", args: args{k: -1}, want: "Test Artist -1"},
		{name: "zero", args: args{k: 0}, want: "Test Artist 0"},
		{name: "positive value", args: args{k: 1}, want: "Test Artist 1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CreateArtistNameForTesting(tt.args.k); got != tt.want {
				t.Errorf("%s = %q, want %q", fnName, got, tt.want)
			}
		})
	}
}

func TestCreateTrackNameForTesting(t *testing.T) {
	fnName := "CreateTrackNameForTesting()"
	type args struct {
		k int
	}
	tests := []struct {
		name string
		args
		want string
	}{
		{name: "zero", args: args{k: 0}, want: "00-Test Track[00].mp3"},
		{name: "odd positive value", args: args{k: 1}, want: "01 Test Track[01].mp3"},
		{name: "even positive value", args: args{k: 2}, want: "02-Test Track[02].mp3"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CreateTrackNameForTesting(tt.args.k); got != tt.want {
				t.Errorf("%s = %q, want %q", fnName, got, tt.want)
			}
		})
	}
}

func TestDestroyDirectoryForTesting(t *testing.T) {
	fnName := "DestroyDirectoryForTesting()"
	type args struct {
		fnName  string
		dirName string
	}
	testDirName := "testDir"
	if err := Mkdir(testDirName); err != nil {
		t.Errorf("%s: error creating %q: %v", fnName, testDirName, err)
	}
	tests := []struct {
		name string
		args
	}{
		{name: "no error", args: args{fnName: fnName, dirName: testDirName}},
		{name: "error", args: args{fnName: fnName, dirName: "."}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			DestroyDirectoryForTesting(tt.args.fnName, tt.args.dirName)
		})
	}
}

func TestPopulateTopDirForTesting(t *testing.T) {
	fnName := "PopulateTopDirForTesting()"
	cleanDirName := "testDir0"
	forceEarlyErrorDirName := "testDir1"
	albumDirErrName := "testDir2"
	badTrackFileName := "testDir3"
	defer func() {
		type results struct {
			dirName string
			e       error
		}
		output := []results{}
		if err := os.RemoveAll(cleanDirName); err != nil {
			output = append(output, results{dirName: cleanDirName, e: err})
		}
		if err := os.RemoveAll(forceEarlyErrorDirName); err != nil {
			output = append(output, results{dirName: forceEarlyErrorDirName, e: err})
		}
		if err := os.RemoveAll(albumDirErrName); err != nil {
			output = append(output, results{dirName: albumDirErrName, e: err})
		}
		if err := os.RemoveAll(badTrackFileName); err != nil {
			output = append(output, results{dirName: badTrackFileName, e: err})
		}
		if len(output) != 0 {
			t.Errorf("%s errors deleting test directories %v", fnName, output)
		}
	}()
	if err := Mkdir(cleanDirName); err != nil {
		t.Errorf("%s error creating directory %q: %v", fnName, cleanDirName, err)
	}
	if err := Mkdir(forceEarlyErrorDirName); err != nil {
		t.Errorf("%s error creating directory %q: %v", fnName, forceEarlyErrorDirName, err)
	}
	artistDirName := CreateArtistNameForTesting(0)
	if err := CreateFileForTesting(forceEarlyErrorDirName, artistDirName); err != nil {
		t.Errorf("%s error creating file %q: %v", fnName, artistDirName, err)
	}

	// create an artist with a file that is named the same as an expected album name
	if err := Mkdir(albumDirErrName); err != nil {
		t.Errorf("%s error creating directory %q: %v", fnName, albumDirErrName, err)
	}
	artistFileName := filepath.Join(albumDirErrName, CreateArtistNameForTesting(0))
	if err := Mkdir(artistFileName); err != nil {
		t.Errorf("%s error creating test directory %q: %v", fnName, artistFileName, err)
	}
	albumFileName := CreateAlbumNameForTesting(0)
	if err := CreateFileForTesting(artistFileName, albumFileName); err != nil {
		t.Errorf("%s error creating file %q: %v", fnName, albumFileName, err)
	}

	// create an album with a pre-existing track name
	if err := Mkdir(badTrackFileName); err != nil {
		t.Errorf("%s error creating directory %q: %v", fnName, badTrackFileName, err)
	}
	artistFileName = filepath.Join(badTrackFileName, CreateArtistNameForTesting(0))
	if err := Mkdir(artistFileName); err != nil {
		t.Errorf("%s error creating test directory %q: %v", fnName, artistFileName, err)
	}
	albumFileName = filepath.Join(artistFileName, CreateAlbumNameForTesting(0))
	if err := Mkdir(albumFileName); err != nil {
		t.Errorf("%s error creating test directory %q: %v", fnName, albumFileName, err)
	}
	trackName := CreateTrackNameForTesting(0)
	if err := CreateFileForTesting(albumFileName, trackName); err != nil {
		t.Errorf("%s error creating track %q: %v", fnName, trackName, err)
	}

	type args struct {
		topDir string
	}
	tests := []struct {
		name string
		args
		wantErr bool
	}{
		{name: "success", args: args{topDir: cleanDirName}, wantErr: false},
		{name: "force early failure", args: args{topDir: forceEarlyErrorDirName}, wantErr: true},
		{name: "bad album name", args: args{topDir: albumDirErrName}, wantErr: true},
		{name: "bad track name", args: args{topDir: badTrackFileName}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := PopulateTopDirForTesting(tt.args.topDir); (err != nil) != tt.wantErr {
				t.Errorf("%s error = %v, wantErr %v", fnName, err, tt.wantErr)
			}
		})
	}
}

func TestCreateDefaultYamlFileForTesting(t *testing.T) {
	fnName := "CreateDefaultYamlFileForTesting()"
	tests := []struct {
		name     string
		preTest  func(t *testing.T)
		postTest func(t *testing.T)
		wantErr  bool
	}{
		{
			name: "dir blocked",
			preTest: func(t *testing.T) {
				if err := CreateFileForTestingWithContent(".", "mp3", []byte("oops")); err != nil {
					t.Errorf("%s 'dir blocked': failed to create file ./mp3: %v", fnName, err)
				}
			},
			postTest: func(t *testing.T) {
				if err := os.Remove("./mp3"); err != nil {
					t.Errorf("%s 'dir blocked': failed to delete ./mp3: %v", fnName, err)
				}
			},
			wantErr: true,
		},
		{
			name: "file exists",
			preTest: func(t *testing.T) {
				if err := Mkdir("./mp3"); err != nil {
					t.Errorf("%s 'file exists': failed to create directory ./mp3: %v", fnName, err)
				}
				if err := CreateFileForTestingWithContent("./mp3", DefaultConfigFileName, []byte("who cares?")); err != nil {
					t.Errorf("%s 'file exists': failed to create %q: %v", fnName, DefaultConfigFileName, err)
				}
			},
			postTest: func(t *testing.T) {
				if err := os.RemoveAll("./mp3"); err != nil {
					t.Errorf("%s 'file exists': failed to remove directory ./mp3: %v", fnName, err)
				}
			},
			wantErr: true,
		},
		{
			name: "good test",
			preTest: func(t *testing.T) {
				// nothing to do
			},
			postTest: func(t *testing.T) {
				savedState := SaveEnvVarForTesting(appDataVar)
				os.Setenv(appDataVar, SecureAbsolutePathForTesting("."))
				defer func() {
					savedState.RestoreForTesting()
				}()
				c, _ := ReadConfigurationFile(NewOutputDeviceForTesting())
				if common := c.cMap["common"]; common == nil {
					t.Errorf("%s 'good test': configuration does not contain common subtree", fnName)
				} else {
					if got := common.sMap["topDir"]; got != "." {
						t.Errorf("%s 'good test': common.topDir got %q want %q", fnName, got, ".")
					}
					if got := common.sMap["ext"]; got != ".mpeg" {
						t.Errorf("%s 'good test': common.ext got %q want %q", fnName, got, ".mpeg")
					}
					if got := common.sMap["albumFilter"]; got != "^.*$" {
						t.Errorf("%s 'good test': common.albums got %q want %q", fnName, got, "^.*$")
					}
					if got := common.sMap["artistFilter"]; got != "^.*$" {
						t.Errorf("%s 'good test': common.artists got %q want %q", fnName, got, "^.*$")
					}
				}
				if list := c.cMap["list"]; list == nil {
					t.Errorf("%s 'good test': configuration does not contain list subtree", fnName)
				} else {
					if got := list.bMap["includeAlbums"]; got != false {
						t.Errorf("%s 'good test': list.album got %t want %t", fnName, got, false)
					}
					if got := list.bMap["includeArtists"]; got != false {
						t.Errorf("%s 'good test': list.artist got %t want %t", fnName, got, false)
					}
					if got := list.bMap["includeTracks"]; got != true {
						t.Errorf("%s 'good test': list.track got %t want %t", fnName, got, true)
					}
					if got := list.bMap["annotate"]; got != true {
						t.Errorf("%s 'good test': list.annotate got %t want %t", fnName, got, true)
					}
					if got := list.sMap["sort"]; got != "alpha" {
						t.Errorf("%s 'good test': list.sort got %s want %s", fnName, got, "alpha")
					}
				}
				if check := c.cMap["check"]; check == nil {
					t.Errorf("%s 'good test': configuration does not contain check subtree", fnName)
				} else {
					if got := check.bMap["empty"]; got != true {
						t.Errorf("%s 'good test': check.empty got %t want %t", fnName, got, true)
					}
					if got := check.bMap["gaps"]; got != true {
						t.Errorf("%s 'good test': check.gaps got %t want %t", fnName, got, true)
					}
					if got := check.bMap["integrity"]; got != false {
						t.Errorf("%s 'good test': check.integrity got %t want %t", fnName, got, false)
					}
				}
				if repair := c.cMap["repair"]; repair == nil {
					t.Errorf("%s 'good test': configuration does not contain repair subtree", fnName)
				} else {
					if got := repair.bMap["dryRun"]; got != true {
						t.Errorf("%s 'good test': repair.DryRun got %t want %t", fnName, got, true)
					}
				}
				DestroyDirectoryForTesting("CreateDefaultYamlFile()", "./mp3")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.preTest(t)
			if err := CreateDefaultYamlFileForTesting(); (err != nil) != tt.wantErr {
				t.Errorf("%s error = %v, wantErr %v", fnName, err, tt.wantErr)
			}
			tt.postTest(t)
		})
	}
}

func TestSaveEnvVarForTesting(t *testing.T) {
	fnName := "SaveEnvVarForTesting()"
	vars := os.Environ()
	firstVar := vars[0]
	i := strings.Index(firstVar, "=")
	firstVar = firstVar[0:i]
	testVar1Exists := false
	name1 := "MP3TEST1"
	testVar2Exists := false
	name2 := "MP3TEST2"
	for _, v := range vars {
		if strings.HasPrefix(v, name1+"=") {
			testVar1Exists = true
		}
		if strings.HasPrefix(v, name2+"=") {
			testVar2Exists = true
		}
	}
	firstSaveState := &SavedEnvVar{Name: firstVar, Value: os.Getenv(firstVar), Set: true}
	os.Unsetenv(firstVar)
	var saveState1 *SavedEnvVar
	if testVar1Exists {
		saveState1 = &SavedEnvVar{Name: name1, Value: os.Getenv(name1), Set: true}
	} else {
		saveState1 = &SavedEnvVar{Name: name1}
	}
	var saveState2 *SavedEnvVar
	if testVar2Exists {
		saveState2 = &SavedEnvVar{Name: name2, Value: os.Getenv(name2), Set: true}
	} else {
		saveState2 = &SavedEnvVar{Name: name2}
	}
	defer func() {
		firstSaveState.RestoreForTesting()
		saveState1.RestoreForTesting()
		saveState2.RestoreForTesting()
		if !reflect.DeepEqual(vars, os.Environ()) {
			t.Errorf("%s environment was not safely restored", fnName)
		}
	}()
	os.Setenv(name1, "value1")
	os.Unsetenv(name2)
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args
		want *SavedEnvVar
	}{
		{name: "set", args: args{name1}, want: &SavedEnvVar{Name: name1, Value: "value1", Set: true}},
		{name: "unset", args: args{name2}, want: &SavedEnvVar{Name: name2}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SaveEnvVarForTesting(tt.args.name); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func TestSecureAbsolutePathForTesting(t *testing.T) {
	fnName := "SecureAbsolutePathForTesting()"
	type args struct {
		path string
	}
	tests := []struct {
		name string
		args
		want bool
	}{
		{name: "simple", args: args{path: "."}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SecureAbsolutePathForTesting(tt.args.path)
			if tt.want && len(got) == 0 {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func TestNewOutputDeviceForTesting(t *testing.T) {
	fnName := "NewOutputDeviceForTesting()"
	tests := []struct {
		name string
		want *OutputDeviceForTesting
	}{
		{
			name: "standard",
			want: &OutputDeviceForTesting{
				consoleWriter: &bytes.Buffer{},
				errorWriter:   &bytes.Buffer{},
				logWriter:     testLogger{writer: &bytes.Buffer{}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got *OutputDeviceForTesting
			if got = NewOutputDeviceForTesting(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
			var o any = got
			if _, ok := o.(OutputBus); !ok {
				t.Errorf("%s does not implement OutputBus", fnName)
			}
		})
	}
}

func TestOutputDeviceForTesting_ErrorWriter(t *testing.T) {
	fnName := "OutputDeviceForTesting.ErrorWriter()"
	tests := []struct {
		name string
		o    *OutputDeviceForTesting
		want io.Writer
	}{
		{name: "standard", o: NewOutputDeviceForTesting(), want: &bytes.Buffer{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.o.ErrorWriter(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
			fmt.Fprintf(tt.o.ErrorWriter(), "test message")
			if gotConsoleOutput := tt.o.ConsoleOutput(); gotConsoleOutput != "" {
				t.Errorf("%s console output = %q, want %q", fnName, gotConsoleOutput, "")
			}
			if gotErrorOutput := tt.o.ErrorOutput(); gotErrorOutput != "test message" {
				t.Errorf("%s error output = %q, want %q", fnName, gotErrorOutput, "test message")
			}
			if gotLogOutput := tt.o.LogOutput(); gotLogOutput != "" {
				t.Errorf("%s log output = %q, want %q", fnName, gotLogOutput, "")
			}
		})
	}
}

func TestCheckOutput(t *testing.T) {
	fnName := "CheckOutput()"
	type args struct {
		o *OutputDeviceForTesting
		w WantedOutput
	}
	tests := []struct {
		name string
		args
		wantIssues []string
		wantOk     bool
	}{
		{name: "normal", args: args{o: NewOutputDeviceForTesting(), w: WantedOutput{}}, wantOk: true},
		{
			name: "errors",
			args: args{
				o: NewOutputDeviceForTesting(),
				w: WantedOutput{
					WantConsoleOutput: "unexpected console output",
					WantErrorOutput:   "unexpected error output",
					WantLogOutput:     "unexpected log output",
				},
			},
			wantIssues: []string{
				"console output = \"\", want \"unexpected console output\"",
				"error output = \"\", want \"unexpected error output\"",
				"log output = \"\", want \"unexpected log output\"",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotIssues, gotOk := tt.args.o.CheckOutput(tt.args.w)
			if !reflect.DeepEqual(gotIssues, tt.wantIssues) {
				t.Errorf("%s gotIssues = %v, want %v", fnName, gotIssues, tt.wantIssues)
			}
			if gotOk != tt.wantOk {
				t.Errorf("%s gotOk = %v, want %v", fnName, gotOk, tt.wantOk)
			}
		})
	}
}

func TestOutputDeviceForTesting_WriteError(t *testing.T) {
	fnName := "OutputDeviceForTesting.WriteError()"
	type args struct {
		format string
		a      []any
	}
	tests := []struct {
		name string
		o    *OutputDeviceForTesting
		args
		want string
	}{
		{
			name: "broad test",
			o:    NewOutputDeviceForTesting(),
			args: args{
				format: "test format %d %q %v\n\n\n\n",
				a:      []any{25, "foo", 1.245},
			},
			want: "Test format 25 \"foo\" 1.245.\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.o.WriteError(tt.args.format, tt.args.a...)
			if got := tt.o.errorWriter.String(); got != tt.want {
				t.Errorf("%s got %q want %q", fnName, got, tt.want)
			}
		})
	}
}

func TestOutputDeviceForTesting_WriteConsole(t *testing.T) {
	fnName := "OutputDeviceForTesting.WriteConsole()"
	type args struct {
		strict bool
		format string
		a      []any
	}
	tests := []struct {
		name string
		o    *OutputDeviceForTesting
		args
		want string
	}{
		{
			name: "strict",
			o:    NewOutputDeviceForTesting(),
			args: args{
				strict: true,
				format: "test easy %s...!",
				a:      []any{"hah!"},
			},
			want: "Test easy hah!\n",
		},
		{
			name: "lax",
			o:    NewOutputDeviceForTesting(),
			args: args{
				strict: false,
				format: "test easy %s...!",
				a:      []any{"hah!"},
			},
			want: "test easy hah!...!",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.o.WriteConsole(tt.args.strict, tt.args.format, tt.args.a...)
			if got := tt.o.consoleWriter.String(); got != tt.want {
				t.Errorf("%s got %q want %q", fnName, got, tt.want)
			}
		})
	}
}
