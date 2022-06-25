package internal

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// NOTE: the functions in this file are strictly for testing purposes. Do not
// call them from production code.

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
func DestroyDirectoryForTesting(fnName string, dirName string) {
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

// CreateNamedFileForTesting creates a specified name with the specified content.
func CreateNamedFileForTesting(fileName, content string) (err error) {
	_, err = os.Stat(fileName)
	if err == nil {
		err = fmt.Errorf("file %q already exists", fileName)
	} else {
		if errors.Is(err, os.ErrNotExist) {
			err = os.WriteFile(fileName, []byte(content), 0644)
		}
	}
	return
}

// CreateFileForTestingWithContent creates a file in a specified directory.
func CreateFileForTestingWithContent(dir, name, content string) error {
	fileName := filepath.Join(dir, name)
	return CreateNamedFileForTesting(fileName, content)
}

func CreateFileForTesting(dir, name string) (err error) {
	return CreateFileForTestingWithContent(dir, name, "file contents for "+name)
}

func CreateDefaultYamlFileForTesting() error {
	path := "./mp3"
	if err := Mkdir(path); err != nil {
		return err
	}
	return CreateFileForTestingWithContent(path, defaultConfigFileName,
		`---
common:
    topDir: .      # $HOMEPATH/Music
    ext: .mpeg     # .mp3
    albumFilter: ^.*$   # .*
    artistFilter: ^.*$  # .* 
ls:
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
    value: 1
repair:
    dryRun: true # false`)
}

type SavedEnvVar struct {
	Name  string
	Value string
	Set   bool
}

func SaveEnvVarForTesting(name string) *SavedEnvVar {
	s := &SavedEnvVar{Name: name}
	if value, ok := os.LookupEnv(name); ok {
		s.Value = value
		s.Set = true
	}
	return s
}

func (e *SavedEnvVar) RestoreForTesting() {
	if e.Set {
		os.Setenv(e.Name, e.Value)
	} else {
		os.Unsetenv(e.Name)
	}
}

func SecureAbsolutePathForTesting(path string) string {
	absPath, _ := filepath.Abs(path)
	return absPath
}

// testing solution

type OutputDeviceForTesting struct {
	wOut *bytes.Buffer
	wErr *bytes.Buffer
	wLog *bytes.Buffer
}

func NewOutputDeviceForTesting() *OutputDeviceForTesting {
	return &OutputDeviceForTesting{
		wOut: &bytes.Buffer{},
		wErr: &bytes.Buffer{},
		wLog: &bytes.Buffer{},
	}
}

func (o *OutputDeviceForTesting) OutputWriter() io.Writer {
	return o.wOut
}

func (o *OutputDeviceForTesting) ErrorWriter() io.Writer {
	return o.wErr
}

func (o *OutputDeviceForTesting) Log(l LogLevel, msg string, fields map[string]interface{}) {
	var parts []string
	for k, v := range fields {
		parts = append(parts, fmt.Sprintf("%s='%v'", k, v))
	}
	sort.Strings(parts)
	var level string
	switch l {
	case INFO:
		level = "info"
	case WARN:
		level = "warn"
	case ERROR:
		level = "error"
	default:
		level = fmt.Sprintf("level unknown (%d)", l)
	}
	fmt.Fprintf(o.wLog, "level='%s' %s msg='%s'\n", level, strings.Join(parts, " "), msg)
}

func (o *OutputDeviceForTesting) Stdout() string {
	return o.wOut.String()
}

func (o *OutputDeviceForTesting) Stderr() string {
	return o.wErr.String()
}

func (o *OutputDeviceForTesting) LogOutput() string {
	return o.wLog.String()
}
