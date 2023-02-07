package internal

import (
	"os"
	"path/filepath"
	"testing"

	tools "github.com/majohn-r/cmd-toolkit"
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
	if err := tools.Mkdir(testDirName); err != nil {
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
	if err := tools.Mkdir(cleanDirName); err != nil {
		t.Errorf("%s error creating directory %q: %v", fnName, cleanDirName, err)
	}
	if err := tools.Mkdir(forceEarlyErrorDirName); err != nil {
		t.Errorf("%s error creating directory %q: %v", fnName, forceEarlyErrorDirName, err)
	}
	artistDirName := CreateArtistNameForTesting(0)
	if err := CreateFileForTesting(forceEarlyErrorDirName, artistDirName); err != nil {
		t.Errorf("%s error creating file %q: %v", fnName, artistDirName, err)
	}

	// create an artist with a file that is named the same as an expected album name
	if err := tools.Mkdir(albumDirErrName); err != nil {
		t.Errorf("%s error creating directory %q: %v", fnName, albumDirErrName, err)
	}
	artistFileName := filepath.Join(albumDirErrName, CreateArtistNameForTesting(0))
	if err := tools.Mkdir(artistFileName); err != nil {
		t.Errorf("%s error creating test directory %q: %v", fnName, artistFileName, err)
	}
	albumFileName := CreateAlbumNameForTesting(0)
	if err := CreateFileForTesting(artistFileName, albumFileName); err != nil {
		t.Errorf("%s error creating file %q: %v", fnName, albumFileName, err)
	}

	// create an album with a pre-existing track name
	if err := tools.Mkdir(badTrackFileName); err != nil {
		t.Errorf("%s error creating directory %q: %v", fnName, badTrackFileName, err)
	}
	artistFileName = filepath.Join(badTrackFileName, CreateArtistNameForTesting(0))
	if err := tools.Mkdir(artistFileName); err != nil {
		t.Errorf("%s error creating test directory %q: %v", fnName, artistFileName, err)
	}
	albumFileName = filepath.Join(artistFileName, CreateAlbumNameForTesting(0))
	if err := tools.Mkdir(albumFileName); err != nil {
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
				if err := tools.Mkdir("./mp3"); err != nil {
					t.Errorf("%s 'file exists': failed to create directory ./mp3: %v", fnName, err)
				}
				if err := CreateFileForTestingWithContent("./mp3", tools.DefaultConfigFileName(), []byte("who cares?")); err != nil {
					t.Errorf("%s 'file exists': failed to create %q: %v", fnName, tools.DefaultConfigFileName(), err)
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
				oldAppPath := tools.SetApplicationPath("mp3")
				tools.InitApplicationPath(output.NewNilBus())
				defer func() {
					tools.SetApplicationPath(oldAppPath)
				}()
				c, _ := tools.ReadConfigurationFile(output.NewNilBus())
				if !c.HasSubConfiguration("common") {
					t.Errorf("%s 'good test': configuration does not contain common subtree", fnName)
				} else {
					common := c.SubConfiguration("common")
					if got, ok := common.StringValue("topDir"); !ok || got != "." {
						t.Errorf("%s 'good test': common.topDir got %q want %q", fnName, got, ".")
					}
					if got, ok := common.StringValue("ext"); !ok || got != ".mpeg" {
						t.Errorf("%s 'good test': common.ext got %q want %q", fnName, got, ".mpeg")
					}
					if got, ok := common.StringValue("albumFilter"); !ok || got != "^.*$" {
						t.Errorf("%s 'good test': common.albums got %q want %q", fnName, got, "^.*$")
					}
					if got, ok := common.StringValue("artistFilter"); !ok || got != "^.*$" {
						t.Errorf("%s 'good test': common.artists got %q want %q", fnName, got, "^.*$")
					}
				}
				if !c.HasSubConfiguration("list") {
					t.Errorf("%s 'good test': configuration does not contain list subtree", fnName)
				} else {
					list := c.SubConfiguration("list")
					if got, ok := list.BooleanValue("includeAlbums"); !ok || got != false {
						t.Errorf("%s 'good test': list.album got %t want %t", fnName, got, false)
					}
					if got, ok := list.BooleanValue("includeArtists"); !ok || got != false {
						t.Errorf("%s 'good test': list.artist got %t want %t", fnName, got, false)
					}
					if got, ok := list.BooleanValue("includeTracks"); !ok || got != true {
						t.Errorf("%s 'good test': list.track got %t want %t", fnName, got, true)
					}
					if got, ok := list.BooleanValue("annotate"); !ok || got != true {
						t.Errorf("%s 'good test': list.annotate got %t want %t", fnName, got, true)
					}
					if got, ok := list.StringValue("sort"); !ok || got != "alpha" {
						t.Errorf("%s 'good test': list.sort got %s want %s", fnName, got, "alpha")
					}
				}
				if !c.HasSubConfiguration("check") {
					t.Errorf("%s 'good test': configuration does not contain check subtree", fnName)
				} else {
					check := c.SubConfiguration("check")
					if got, ok := check.BooleanValue("empty"); !ok || got != true {
						t.Errorf("%s 'good test': check.empty got %t want %t", fnName, got, true)
					}
					if got, ok := check.BooleanValue("gaps"); !ok || got != true {
						t.Errorf("%s 'good test': check.gaps got %t want %t", fnName, got, true)
					}
					if got, ok := check.BooleanValue("integrity"); !ok || got != false {
						t.Errorf("%s 'good test': check.integrity got %t want %t", fnName, got, false)
					}
				}
				if !c.HasSubConfiguration("repair") {
					t.Errorf("%s 'good test': configuration does not contain repair subtree", fnName)
				} else {
					repair := c.SubConfiguration("repair")
					if got, ok := repair.BooleanValue("dryRun"); !ok || got != true {
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
