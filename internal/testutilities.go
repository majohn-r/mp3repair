package internal

import (
	"errors"
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
	creationParameters := []struct {
		artist        int
		createContent bool
	}{
		{artist: 0, createContent: true},
		{artist: 1, createContent: true},
		{artist: 2, createContent: true},
		{artist: 3, createContent: true},
		{artist: 4, createContent: true},
		{artist: 5, createContent: true},
		{artist: 6, createContent: true},
		{artist: 7, createContent: true},
		{artist: 8, createContent: true},
		{artist: 9, createContent: true},
		{artist: 999, createContent: false},
	}
	for _, params := range creationParameters {
		if err := createArtistDirForTesting(topDir, params.artist, params.createContent); err != nil {
			return err
		}
	}
	return nil
}

func createAlbumDirForTesting(artistDir string, n int, tracks int) error {
	albumDir := filepath.Join(artistDir, CreateAlbumNameForTesting(n))
	dummyDir := filepath.Join(albumDir, "ignore this folder")
	directories := []string{albumDir, dummyDir}
	for _, directory := range directories {
		if err := Mkdir(directory); err != nil {
			return err
		}
	}
	var fileNames []string
	for k := 0; k < tracks; k++ {
		fileNames = append(fileNames, CreateTrackNameForTesting(k))
	}
	fileNames = append(fileNames, "album cover.jpeg")
	for _, fileName := range fileNames {
		if err := CreateFileForTesting(albumDir, fileName); err != nil {
			return err
		}
	}
	return nil
}

func createArtistDirForTesting(topDir string, k int, withContent bool) error {
	artistDir := filepath.Join(topDir, CreateArtistNameForTesting(k))
	if err := Mkdir(artistDir); err != nil {
		return err
	}
	if withContent {
		creationParams := []struct {
			id     int
			tracks int
		}{
			{id: 0, tracks: 10},
			{id: 1, tracks: 10},
			{id: 2, tracks: 10},
			{id: 3, tracks: 10},
			{id: 4, tracks: 10},
			{id: 5, tracks: 10},
			{id: 6, tracks: 10},
			{id: 7, tracks: 10},
			{id: 8, tracks: 10},
			{id: 9, tracks: 10},
			{id: 999, tracks: 0},
		}
		for _, p := range creationParams {
			if err := createAlbumDirForTesting(artistDir, p.id, p.tracks); err != nil {
				return err
			}
		}
		return CreateFileForTesting(artistDir, "dummy file to be ignored.txt")
	}
	return nil
}

func CreateFileForTesting(dir, name string) (err error) {
	fileName := filepath.Join(dir, name)
	_, err = os.Stat(fileName)
	if err == nil {
		err = fmt.Errorf("file %q already exists", fileName)
	} else {
		if errors.Is(err, os.ErrNotExist) {
			err = os.WriteFile(fileName, []byte("file contents for "+name), 0644)
		}
	}
	return
}
