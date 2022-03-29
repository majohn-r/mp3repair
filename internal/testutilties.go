package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// NOTE: the functions in this file are strictly for testing purposes. Do not
// call them from production code.

func CreateAlbumName(k int) string {
	return fmt.Sprintf("Test Album %d", k)
}

func CreateArtistName(k int) string {
	return fmt.Sprintf("Test Artist %d", k)
}

func CreateTrackName(k int) string {
	switch k % 2 {
	case 0:
		return fmt.Sprintf("%02d-Test Track[%02d].mp3", k, k)
	default:
		return fmt.Sprintf("%02d Test Track[%02d].mp3", k, k)
	}
}

func Mkdir(t *testing.T, fnName string, dirName string) error {
	if err := os.Mkdir(dirName, 0755); err != nil {
		t.Errorf("%s: error creating directory %q: %v", fnName, dirName, err)
		return err
	}
	return nil
}

func PopulateTopDir(t *testing.T, fnName string, topDir string) {
	internalFnName := fmt.Sprintf("%s->PopulateTopDir()", fnName)
	for k := 0; k < 10; k++ {
		createArtistDir(t, internalFnName, topDir, k, true)
	}
	createArtistDir(t, internalFnName, topDir, 999, false)
}

func createAlbumDir(t *testing.T, fnName string, artistDir string, n int, tracks int) {
	internalFnName := fmt.Sprintf("%s->createAlbumDir()", fnName)
	albumDir := filepath.Join(artistDir, CreateAlbumName(n))
	if err := Mkdir(t, "createAlbumDir()", albumDir); err == nil {
		for k := 0; k < tracks; k++ {
			createTrackFile(t, internalFnName, albumDir, k)
		}
		createFile(t, internalFnName, albumDir, "album cover.jpeg")
		dummyDir := filepath.Join(albumDir, "ignore this folder")
		_ = Mkdir(t, internalFnName, dummyDir)
	}
}

func createArtistDir(t *testing.T, fnName string, topDir string, k int, withContent bool) {
	internalFnName := fmt.Sprintf("%s->createArtistDir()", fnName)
	artistDir := filepath.Join(topDir, CreateArtistName(k))
	if err := Mkdir(t, "createArtistDir()", artistDir); err == nil {
		if withContent {
			for n := 0; n < 10; n++ {
				createAlbumDir(t, internalFnName, artistDir, n, 10)
			}
			createAlbumDir(t, internalFnName, artistDir, 999, 0) // create album with no tracks
			createFile(t, internalFnName, artistDir, "dummy file to be ignored.txt")
		}
	}
}

func createFile(t *testing.T, fnName, dir, s string) {
	internalFnName := fmt.Sprintf("%s->createFile()", fnName)
	fileName := filepath.Join(dir, s)
	if err := os.WriteFile(fileName, []byte("file contents for "+s), 0644); err != nil {
		t.Errorf("%s error creating %q: %v", internalFnName, fileName, err)
	}
}

func createTrackFile(t *testing.T, fnName string, artistDir string, k int) {
	internalFnName := fmt.Sprintf("%s->createTrackFile()", fnName)
	createFile(t, internalFnName, artistDir, CreateTrackName(k))
}
