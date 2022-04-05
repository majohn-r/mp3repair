package internal

import (
	"fmt"
	"os"
	"path/filepath"
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

func DestroyDirectoryForTesting(fnName string, dirName string) {
	if err := os.RemoveAll(dirName); err != nil {
		fmt.Fprintf(os.Stderr, "%s error destroying test directory %q: %v", fnName, dirName, err)
	}
}

func PopulateTopDirForTesting(topDir string) error {
	for k := 0; k < 10; k++ {
		if err := createArtistDirForTesting(topDir, k, true); err != nil {
			return err
		}
	}
	return createArtistDirForTesting(topDir, 999, false)
}

func createAlbumDirForTesting(artistDir string, n int, tracks int) error {
	albumDir := filepath.Join(artistDir, CreateAlbumNameForTesting(n))
	if err := Mkdir(albumDir); err != nil {
		return err
	}
	for k := 0; k < tracks; k++ {
		if err := createTrackFile(albumDir, k); err != nil {
			return err
		}
	}
	if err := createFileForTesting(albumDir, "album cover.jpeg"); err != nil {
		return err
	}
	dummyDir := filepath.Join(albumDir, "ignore this folder")
	return Mkdir(dummyDir)
}

func createArtistDirForTesting(topDir string, k int, withContent bool) error {
	artistDir := filepath.Join(topDir, CreateArtistNameForTesting(k))
	if err := Mkdir(artistDir); err != nil {
		return err
	}
	if withContent {
		for n := 0; n < 10; n++ {
			if err := createAlbumDirForTesting(artistDir, n, 10); err != nil {
				return err
			}
		}
		if err := createAlbumDirForTesting(artistDir, 999, 0); err != nil {
			return err
		} // create album with no tracks
		return createFileForTesting(artistDir, "dummy file to be ignored.txt")
	}
	return nil
}

func createFileForTesting(dir, s string) error {
	fileName := filepath.Join(dir, s)
	return os.WriteFile(fileName, []byte("file contents for "+s), 0644)
}

func createTrackFile(artistDir string, k int) error {
	return createFileForTesting(artistDir, CreateTrackNameForTesting(k))
}
