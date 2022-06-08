package internal

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func TestCreateAlbumNameForTesting(t *testing.T) {
	fnName := "CreateAlbumNameForTesting()"
	type args struct {
		k int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "negative value", args: args{k: -1}, want: "Test Album -1"},
		{name: "zero", args: args{k: 0}, want: "Test Album 0"},
		{name: "positive value", args: args{k: 1}, want: "Test Album 1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CreateAlbumNameForTesting(tt.args.k); got != tt.want {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
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
		args args
		want string
	}{
		{name: "negative value", args: args{k: -1}, want: "Test Artist -1"},
		{name: "zero", args: args{k: 0}, want: "Test Artist 0"},
		{name: "positive value", args: args{k: 1}, want: "Test Artist 1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CreateArtistNameForTesting(tt.args.k); got != tt.want {
				t.Errorf("%q = %v, want %v", fnName, got, tt.want)
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
		args args
		want string
	}{
		{name: "zero", args: args{k: 0}, want: "00-Test Track[00].mp3"},
		{name: "odd positive value", args: args{k: 1}, want: "01 Test Track[01].mp3"},
		{name: "even positive value", args: args{k: 2}, want: "02-Test Track[02].mp3"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CreateTrackNameForTesting(tt.args.k); got != tt.want {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
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
		args args
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
		name    string
		args    args
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
	tests := []struct {
		name     string
		preTest  func(t *testing.T)
		postTest func(t *testing.T)
		wantErr  bool
	}{
		{
			name: "dir blocked",
			preTest: func(t *testing.T) {
				if err := CreateFileForTestingWithContent(".", "mp3", "oops"); err != nil {
					t.Errorf("CreateDefaultYamlFile() 'dir blocked': failed to create file ./mp3: %v", err)
				}
			},
			postTest: func(t *testing.T) {
				if err := os.Remove("./mp3"); err != nil {
					t.Errorf("CreateDefaultYamlFile() 'dir blocked': failed to delete ./mp3: %v", err)
				}
			},
			wantErr: true,
		},
		{
			name: "file exists",
			preTest: func(t *testing.T) {
				if err := Mkdir("./mp3"); err != nil {
					t.Errorf("CreateDefaultYamlFile() 'file exists': failed to create directory ./mp3: %v", err)
				}
				if err := CreateFileForTestingWithContent("./mp3", DefaultConfigFileName, "who cares?"); err != nil {
					t.Errorf("CreateDefaultYamlFile() 'file exists': failed to create %q: %v", DefaultConfigFileName, err)
				}
			},
			postTest: func(t *testing.T) {
				if err := os.RemoveAll("./mp3"); err != nil {
					t.Errorf("CreateDefaultYamlFile() 'file exists': failed to remove directory ./mp3: %v", err)
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
				v := viper.New()
				v.SetConfigName(defaultConfigFileBaseName)
				v.AddConfigPath("./mp3")
				if err := v.ReadInConfig(); err != nil {
					t.Errorf("CreateDefaultYamlFile() 'good test': error reading defaults configuration file: %v", err)
				}
				if common := v.Sub("common"); common == nil {
					t.Error("CreateDefaultYamlFile() 'good test': viper does not contain common subtree")
				} else {
					if got := common.Get("topDir"); got != "." {
						t.Errorf("CreateDefaultYamlFile() 'good test': common.topDir got %s want %s", got, ".")
					}
					if got := common.Get("ext"); got != ".mpeg" {
						t.Errorf("CreateDefaultYamlFile() 'good test': common.ext got %s want %s", got, ".mpeg")
					}
					if got := common.Get("albums"); got != "^.*$" {
						t.Errorf("CreateDefaultYamlFile() 'good test': common.albums got %s want %s", got, "^.*$")
					}
					if got := common.Get("artists"); got != "^.*$" {
						t.Errorf("CreateDefaultYamlFile() 'good test': common.artists got %s want %s", got, "^.*$")
					}
				}
				if ls := v.Sub("ls"); ls == nil {
					t.Error("CreateDefaultYamlFile() 'good test': viper does not contain ls subtree")
				} else {
					if got := ls.Get("album"); got != false {
						t.Errorf("CreateDefaultYamlFile() 'good test': ls.album got %t want %t", got, false)
					}
					if got := ls.Get("artist"); got != false {
						t.Errorf("CreateDefaultYamlFile() 'good test': ls.artist got %t want %t", got, false)
					}
					if got := ls.Get("track"); got != true {
						t.Errorf("CreateDefaultYamlFile() 'good test': ls.track got %t want %t", got, true)
					}
					if got := ls.Get("annotate"); got != true {
						t.Errorf("CreateDefaultYamlFile() 'good test': ls.annotate got %t want %t", got, true)
					}
					if got := ls.Get("sort"); got != "alpha" {
						t.Errorf("CreateDefaultYamlFile() 'good test': ls.sort got %s want %s", got, "alpha")
					}
				}
				if check := v.Sub("check"); check == nil {
					t.Error("CreateDefaultYamlFile() 'good test': viper does not contain check subtree")
				} else {
					if got := check.Get("empty"); got != true {
						t.Errorf("CreateDefaultYamlFile() 'good test': check.empty got %t want %t", got, true)
					}
					if got := check.Get("gaps"); got != true {
						t.Errorf("CreateDefaultYamlFile() 'good test': check.gaps got %t want %t", got, true)
					}
					if got := check.Get("integrity"); got != false {
						t.Errorf("CreateDefaultYamlFile() 'good test': check.integrity got %t want %t", got, false)
					}
				}
				if repair := v.Sub("repair"); repair == nil {
					t.Error("CreateDefaultYamlFile() 'good test': viper does not contain repair subtree")
				} else {
					if got := repair.Get("dryRun"); got != true {
						t.Errorf("CreateDefaultYamlFile() 'good test': repair.DryRun got %t want %t", got, true)
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
				t.Errorf("CreateDefaultYamlFile() error = %v, wantErr %v", err, tt.wantErr)
			}
			tt.postTest(t)
		})
	}
}
