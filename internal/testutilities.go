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

func PopulateTopDirForTesting(topDir string) {
	for k := 0; k < 10; k++ {
		createArtistDirForTesting(topDir, k, true)
	}
	createArtistDirForTesting(topDir, 999, false)
}

func createAlbumDirForTesting(artistDir string, n int, tracks int) {
	albumDir := filepath.Join(artistDir, CreateAlbumNameForTesting(n))
	if err := Mkdir(albumDir); err == nil {
		for k := 0; k < tracks; k++ {
			createTrackFile(albumDir, k)
		}
		createFileForTesting(albumDir, "album cover.jpeg")
		dummyDir := filepath.Join(albumDir, "ignore this folder")
		_ = Mkdir(dummyDir)
	}
}

func createArtistDirForTesting(topDir string, k int, withContent bool) {
	artistDir := filepath.Join(topDir, CreateArtistNameForTesting(k))
	if err := Mkdir(artistDir); err == nil {
		if withContent {
			for n := 0; n < 10; n++ {
				createAlbumDirForTesting(artistDir, n, 10)
			}
			createAlbumDirForTesting(artistDir, 999, 0) // create album with no tracks
			createFileForTesting(artistDir, "dummy file to be ignored.txt")
		}
	}
}

func createFileForTesting(dir, s string) {
	fileName := filepath.Join(dir, s)
	_ = os.WriteFile(fileName, []byte("file contents for "+s), 0644)
}

func createTrackFile(artistDir string, k int) {
	createFileForTesting(artistDir, CreateTrackNameForTesting(k))
}
