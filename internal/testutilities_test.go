package internal

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/majohn-r/output"
)

func TestCreateAlbumNameForTesting(t *testing.T) {
	const fnName = "CreateAlbumNameForTesting()"
	type args struct {
		albumNumber int
	}
	tests := map[string]struct {
		args
		want string
	}{
		"negative value": {args: args{albumNumber: -1}, want: "Test Album -1"},
		"zero":           {args: args{albumNumber: 0}, want: "Test Album 0"},
		"positive value": {args: args{albumNumber: 1}, want: "Test Album 1"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := CreateAlbumNameForTesting(tt.args.albumNumber); got != tt.want {
				t.Errorf("%s = %q, want %q", fnName, got, tt.want)
			}
		})
	}
}

func TestCreateArtistNameForTesting(t *testing.T) {
	const fnName = "CreateArtistNameForTesting()"
	type args struct {
		artistNumber int
	}
	tests := map[string]struct {
		args
		want string
	}{
		"negative value": {args: args{artistNumber: -1}, want: "Test Artist -1"},
		"zero":           {args: args{artistNumber: 0}, want: "Test Artist 0"},
		"positive value": {args: args{artistNumber: 1}, want: "Test Artist 1"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := CreateArtistNameForTesting(tt.args.artistNumber); got != tt.want {
				t.Errorf("%s = %q, want %q", fnName, got, tt.want)
			}
		})
	}
}

func TestCreateTrackNameForTesting(t *testing.T) {
	const fnName = "CreateTrackNameForTesting()"
	type args struct {
		trackNumber int
	}
	tests := map[string]struct {
		args
		want string
	}{
		"zero":                {args: args{trackNumber: 0}, want: "00-Test Track[00].mp3"},
		"odd positive value":  {args: args{trackNumber: 1}, want: "01 Test Track[01].mp3"},
		"even positive value": {args: args{trackNumber: 2}, want: "02-Test Track[02].mp3"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := CreateTrackNameForTesting(tt.args.trackNumber); got != tt.want {
				t.Errorf("%s = %q, want %q", fnName, got, tt.want)
			}
		})
	}
}

func TestDestroyDirectoryForTesting(t *testing.T) {
	const fnName = "DestroyDirectoryForTesting()"
	testDirName := "testDir"
	if err := Mkdir(testDirName); err != nil {
		t.Errorf("%s: error creating %q: %v", fnName, testDirName, err)
	}
	type args struct {
		fnName  string
		dirName string
	}
	tests := map[string]struct {
		args
	}{
		"no error": {args: args{fnName: fnName, dirName: testDirName}},
		"error":    {args: args{fnName: fnName, dirName: "."}},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			DestroyDirectoryForTesting(tt.args.fnName, tt.args.dirName)
		})
	}
}

func TestPopulateTopDirForTesting(t *testing.T) {
	const fnName = "PopulateTopDirForTesting()"
	cleanDirName := "testDir0"
	forceEarlyErrorDirName := "testDir1"
	albumDirErrName := "testDir2"
	badTrackFileName := "testDir3"
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

	defer func() {
		type results struct {
			dirName string
			e       error
		}
		listing := []results{}
		if err := os.RemoveAll(cleanDirName); err != nil {
			listing = append(listing, results{dirName: cleanDirName, e: err})
		}
		if err := os.RemoveAll(forceEarlyErrorDirName); err != nil {
			listing = append(listing, results{dirName: forceEarlyErrorDirName, e: err})
		}
		if err := os.RemoveAll(albumDirErrName); err != nil {
			listing = append(listing, results{dirName: albumDirErrName, e: err})
		}
		if err := os.RemoveAll(badTrackFileName); err != nil {
			listing = append(listing, results{dirName: badTrackFileName, e: err})
		}
		if len(listing) != 0 {
			t.Errorf("%s errors deleting test directories %v", fnName, listing)
		}
	}()
	type args struct {
		topDir string
	}
	tests := map[string]struct {
		args
		wantErr bool
	}{
		"success":             {args: args{topDir: cleanDirName}, wantErr: false},
		"force early failure": {args: args{topDir: forceEarlyErrorDirName}, wantErr: true},
		"bad album name":      {args: args{topDir: albumDirErrName}, wantErr: true},
		"bad track name":      {args: args{topDir: badTrackFileName}, wantErr: true},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if err := PopulateTopDirForTesting(tt.args.topDir); (err != nil) != tt.wantErr {
				t.Errorf("%s error = %v, wantErr %v", fnName, err, tt.wantErr)
			}
		})
	}
}

func TestCreateDefaultYamlFileForTesting(t *testing.T) {
	const fnName = "CreateDefaultYamlFileForTesting()"
	tests := map[string]struct {
		preTest  func(t *testing.T)
		postTest func(t *testing.T)
		wantErr  bool
	}{
		"dir blocked": {
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
		"file exists": {
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
		"good test": {
			preTest: func(t *testing.T) {
				// nothing to do
			},
			postTest: func(t *testing.T) {
				savedState := SaveEnvVarForTesting(appDataVar)
				os.Setenv(appDataVar, SecureAbsolutePathForTesting("."))
				oldAppPath := ApplicationPath()
				InitApplicationPath(output.NewNilBus())
				defer func() {
					savedState.RestoreForTesting()
					SetApplicationPathForTesting(oldAppPath)
				}()
				c, _ := ReadConfigurationFile(output.NewNilBus())
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
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.preTest(t)
			if err := CreateDefaultYamlFileForTesting(); (err != nil) != tt.wantErr {
				t.Errorf("%s error = %v, wantErr %v", fnName, err, tt.wantErr)
			}
			tt.postTest(t)
		})
	}
}

func TestSaveEnvVarForTesting(t *testing.T) {
	const fnName = "SaveEnvVarForTesting()"
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
	os.Setenv(name1, "value1")
	os.Unsetenv(name2)
	defer func() {
		firstSaveState.RestoreForTesting()
		saveState1.RestoreForTesting()
		saveState2.RestoreForTesting()
		if !reflect.DeepEqual(vars, os.Environ()) {
			t.Errorf("%s environment was not safely restored", fnName)
		}
	}()
	type args struct {
		name string
	}
	tests := map[string]struct {
		args
		want *SavedEnvVar
	}{
		"set":   {args: args{name1}, want: &SavedEnvVar{Name: name1, Value: "value1", Set: true}},
		"unset": {args: args{name2}, want: &SavedEnvVar{Name: name2}},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := SaveEnvVarForTesting(tt.args.name); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func TestSecureAbsolutePathForTesting(t *testing.T) {
	const fnName = "SecureAbsolutePathForTesting()"
	type args struct {
		path string
	}
	tests := map[string]struct {
		args
		want bool
	}{"simple": {args: args{path: "."}, want: true}}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := SecureAbsolutePathForTesting(tt.args.path)
			if tt.want && got == "" {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}
