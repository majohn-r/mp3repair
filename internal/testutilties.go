package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// NOTE: the functions in this file are strictly for testing purposes. Do not
// call them from production code.

func CreateAlbum(t *testing.T, artistDir string, n int) {
	albumDir := filepath.Join(artistDir, CreateAlbumName(n))
	if err := Mkdir(t, "createAlbum", albumDir); err == nil {
		for k := 0; k < 10; k++ {
			CreateTrack(t, albumDir, k)
		}
		CreateFile(t, albumDir, "album cover.jpeg")
		dummyDir := filepath.Join(albumDir, "ignore this folder")
		_ = Mkdir(t, "createAlbum", dummyDir)
	}
}


func CreateAlbumName(k int) string {
	return fmt.Sprintf("Test Album %d", k)
}

func CreateArtist(t *testing.T, topDir string, k int) {
	artistDir := filepath.Join(topDir, CreateArtistName(k))
	if err := Mkdir(t, "createArtist", artistDir); err == nil {
		for n := 0; n < 10; n++ {
			CreateAlbum(t, artistDir, n)
		}
		CreateFile(t, artistDir, "dummy file to be ignored.txt")
	}
}

func CreateArtistName(k int) string {
	return fmt.Sprintf("Test Artist %d", k)
}

func CreateFile(t *testing.T, dir, s string) {
	fileName := filepath.Join(dir, s)
	if err := os.WriteFile(fileName, []byte("file contents for "+s), 0644); err != nil {
		t.Errorf("createFile() error creating %q: %v", fileName, err)
	}
}

func CreateTrack(t *testing.T, artistDir string, k int) {
	CreateFile(t, artistDir, CreateTrackName(k))
}

func CreateTrackName(k int) string {
	return fmt.Sprintf("%02d Test Track[%02d].mp3", k, k)
}

func Mkdir(t *testing.T, fnName string, dirName string) error {
	if err := os.Mkdir(dirName, 0755); err != nil {
		t.Errorf("%s error creating directory %q: %v", fnName, dirName, err)
		return err
	}
	return nil
}