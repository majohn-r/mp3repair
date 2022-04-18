package internal

import (
	"os"
	"path/filepath"
	"testing"
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
			e error
		}
		output := []results{}
		if err := os.RemoveAll(cleanDirName); err != nil {
			output = append(output, results{ dirName: cleanDirName, e: err})
		}
		if err := os.RemoveAll(forceEarlyErrorDirName); err != nil {
			output = append(output, results{ dirName: forceEarlyErrorDirName, e: err})
		}
		if err := os.RemoveAll(albumDirErrName); err != nil {
			output = append(output, results{ dirName: albumDirErrName, e: err})
		}
		if err := os.RemoveAll(badTrackFileName); err != nil {
			output = append(output, results{ dirName: badTrackFileName, e: err})
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
	if err := createFileForTesting(forceEarlyErrorDirName, artistDirName); err != nil {
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
	if err := createFileForTesting(artistFileName, albumFileName) ; err != nil {
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
	if err := Mkdir(albumFileName) ; err != nil {
		t.Errorf("%s error creating test directory %q: %v", fnName, albumFileName, err)
	}
	trackName := CreateTrackNameForTesting(0)
	if err := createFileForTesting(albumFileName, trackName) ; err != nil {
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
