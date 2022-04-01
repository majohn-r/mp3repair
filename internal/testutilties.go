package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// NOTE: the functions in this file are strictly for testing purposes. Do not
// call them from production code.

func CreateAlbumNameForTesting(k int) string {
	return fmt.Sprintf("Test Album %d", k)
}

func CreateArtistNameForTesting(k int) string {
	return fmt.Sprintf("Test Artist %d", k)
}

func CreateTrackNameForTesting(k int) string {
	switch k % 2 {
	case 0:
		return fmt.Sprintf("%02d-Test Track[%02d].mp3", k, k)
	default:
		return fmt.Sprintf("%02d Test Track[%02d].mp3", k, k)
	}
}

func MkdirForTesting(t *testing.T, fnName string, dirName string) error {
	if err := os.Mkdir(dirName, 0755); err != nil {
		t.Errorf("%s: error creating directory %q: %v", fnName, dirName, err)
		return err
	}
	return nil
}

func PopulateTopDirForTesting(t *testing.T, fnName string, topDir string) {
	internalFnName := fmt.Sprintf("%s->PopulateTopDir()", fnName)
	for k := 0; k < 10; k++ {
		createArtistDirForTesting(t, internalFnName, topDir, k, true)
	}
	createArtistDirForTesting(t, internalFnName, topDir, 999, false)
}

func createAlbumDirForTesting(t *testing.T, fnName string, artistDir string, n int, tracks int) {
	internalFnName := fmt.Sprintf("%s->createAlbumDir()", fnName)
	albumDir := filepath.Join(artistDir, CreateAlbumNameForTesting(n))
	if err := MkdirForTesting(t, "createAlbumDir()", albumDir); err == nil {
		for k := 0; k < tracks; k++ {
			createTrackFile(t, internalFnName, albumDir, k)
		}
		createFileForTesting(t, internalFnName, albumDir, "album cover.jpeg")
		dummyDir := filepath.Join(albumDir, "ignore this folder")
		_ = MkdirForTesting(t, internalFnName, dummyDir)
	}
}

func createArtistDirForTesting(t *testing.T, fnName string, topDir string, k int, withContent bool) {
	internalFnName := fmt.Sprintf("%s->createArtistDir()", fnName)
	artistDir := filepath.Join(topDir, CreateArtistNameForTesting(k))
	if err := MkdirForTesting(t, "createArtistDir()", artistDir); err == nil {
		if withContent {
			for n := 0; n < 10; n++ {
				createAlbumDirForTesting(t, internalFnName, artistDir, n, 10)
			}
			createAlbumDirForTesting(t, internalFnName, artistDir, 999, 0) // create album with no tracks
			createFileForTesting(t, internalFnName, artistDir, "dummy file to be ignored.txt")
		}
	}
}

func createFileForTesting(t *testing.T, fnName, dir, s string) {
	internalFnName := fmt.Sprintf("%s->createFile()", fnName)
	fileName := filepath.Join(dir, s)
	if err := os.WriteFile(fileName, []byte("file contents for "+s), 0644); err != nil {
		t.Errorf("%s error creating %q: %v", internalFnName, fileName, err)
	}
}

func createTrackFile(t *testing.T, fnName string, artistDir string, k int) {
	internalFnName := fmt.Sprintf("%s->createTrackFile()", fnName)
	createFileForTesting(t, internalFnName, artistDir, CreateTrackNameForTesting(k))
}
