package internal

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	cmd_toolkit "github.com/majohn-r/cmd-toolkit"
)

// NOTE: the functions in this file are strictly for testing purposes. Do not
// call them from production code.

var (
	// ID3V1DataSet1 is a sample ID3V1 tag from an existing file
	ID3V1DataSet1 = []byte{
		'T', 'A', 'G',
		'R', 'i', 'n', 'g', 'o', ' ', '-', ' ', 'P', 'o', 'p', ' ', 'P', 'r', 'o', 'f', 'i', 'l', 'e', ' ', '[', 'I', 'n', 't', 'e', 'r', 'v', 'i', 'e', 'w',
		'T', 'h', 'e', ' ', 'B', 'e', 'a', 't', 'l', 'e', 's', 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		'O', 'n', ' ', 'A', 'i', 'r', ':', ' ', 'L', 'i', 'v', 'e', ' ', 'A', 't', ' ', 'T', 'h', 'e', ' ', 'B', 'B', 'C', ',', ' ', 'V', 'o', 'l', 'u', 'm',
		'2', '0', '1', '3',
		' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
		0,
		29,
		12,
	}
	// ID3V1DataSet2 is a sample ID3V1 tag from an existing file
	ID3V1DataSet2 = []byte{
		'T', 'A', 'G',
		'J', 'u', 'l', 'i', 'a', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
		'T', 'h', 'e', ' ', 'B', 'e', 'a', 't', 'l', 'e', 's', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
		'T', 'h', 'e', ' ', 'W', 'h', 'i', 't', 'e', ' ', 'A', 'l', 'b', 'u', 'm', ' ', '[', 'D', 'i', 's', 'c', ' ', '1', ']', ' ', ' ', ' ', ' ', ' ', ' ',
		'1', '9', '6', '8',
		' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
		0,
		17,
		17,
	}
)

// CreateAlbumNameForTesting creates a suitable album name
func CreateAlbumNameForTesting(k int) string {
	return fmt.Sprintf("Test Album %d", k)
}

// CreateArtistNameForTesting creates a suitable artist name.
func CreateArtistNameForTesting(k int) string {
	return fmt.Sprintf("Test Artist %d", k)
}

// CreateTrackNameForTesting creates a suitable track name.
func CreateTrackNameForTesting(k int) string {
	switch k % 2 {
	case 0:
		return fmt.Sprintf("%02d-Test Track[%02d].mp3", k, k)
	default:
		return fmt.Sprintf("%02d Test Track[%02d].mp3", k, k)
	}
}

// DestroyDirectoryForTesting destroys a directory and its contents.
func DestroyDirectoryForTesting(fnName, dirName string) {
	if err := os.RemoveAll(dirName); err != nil {
		fmt.Fprintf(os.Stderr, "%s error destroying test directory %q: %v", fnName, dirName, err)
	}
}

// PopulateTopDirForTesting populates a specified directory with a collection of
// artists, albums, and tracks.
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

func createAlbumDirForTesting(artistDir string, albumNumber, tracks int) error {
	albumDir := filepath.Join(artistDir, CreateAlbumNameForTesting(albumNumber))
	dummyDir := filepath.Join(albumDir, "ignore this folder")
	directories := []string{albumDir, dummyDir}
	for _, directory := range directories {
		if err := cmd_toolkit.Mkdir(directory); err != nil {
			return err
		}
	}
	var names []string
	for k := 0; k < tracks; k++ {
		names = append(names, CreateTrackNameForTesting(k))
	}
	names = append(names, "album cover.jpeg")
	for _, name := range names {
		if err := CreateFileForTesting(albumDir, name); err != nil {
			return err
		}
	}
	return nil
}

func createArtistDirForTesting(topDir string, artistNumber int, withContent bool) error {
	artistDir := filepath.Join(topDir, CreateArtistNameForTesting(artistNumber))
	if err := cmd_toolkit.Mkdir(artistDir); err != nil {
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

// CreateNamedFileForTesting creates a specified name with the specified content.
func CreateNamedFileForTesting(fileName string, content []byte) (err error) {
	_, err = os.Stat(fileName)
	if err == nil {
		err = fmt.Errorf("file %q already exists", fileName)
	} else if errors.Is(err, os.ErrNotExist) {
		err = os.WriteFile(fileName, content, cmd_toolkit.StdFilePermissions)
	}
	return
}

// CreateFileForTestingWithContent creates a file in a specified directory.
func CreateFileForTestingWithContent(dir, name string, content []byte) error {
	fileName := filepath.Join(dir, name)
	return CreateNamedFileForTesting(fileName, content)
}

// CreateFileForTesting creates a file in a specified directory with
// standardized content
func CreateFileForTesting(dir, name string) (err error) {
	return CreateFileForTestingWithContent(dir, name, []byte("file contents for "+name))
}

// CreateDefaultYamlFileForTesting creates a yaml file with different defaults
// than the prescribed values
func CreateDefaultYamlFileForTesting() error {
	path := "./mp3"
	if err := cmd_toolkit.Mkdir(path); err != nil {
		return err
	}
	yamlInput := `---
common:
    topDir: .      # %HOMEPATH%\Music
    ext: .mpeg     # .mp3
    albumFilter: ^.*$   # .*
    artistFilter: ^.*$  # .* 
list:
    includeAlbums: false  # true
    includeArtists: false # true
    includeTracks: true   # false
    sort: alpha           # numeric
    annotate: true        # false
check:
    empty: true      # false
    gaps: true       # false
    integrity: false # true
unused:
    value: 1.25
repair:
    dryRun: true # false`
	return CreateFileForTestingWithContent(path, cmd_toolkit.DefaultConfigFileName(), []byte(yamlInput))
}
