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
func CreateNamedFileForTesting(fileName string, content []byte) (err error) {
	_, err = os.Stat(fileName)
	if err == nil {
		err = fmt.Errorf("file %q already exists", fileName)
	} else {
		if errors.Is(err, os.ErrNotExist) {
			err = os.WriteFile(fileName, content, 0644)
		}
	}
	return
}

// CreateFileForTestingWithContent creates a file in a specified directory.
func CreateFileForTestingWithContent(dir string, name string, content []byte) error {
	fileName := filepath.Join(dir, name)
	return CreateNamedFileForTesting(fileName, content)
}

func CreateFileForTesting(dir, name string) (err error) {
	return CreateFileForTestingWithContent(dir, name, []byte("file contents for "+name))
}

func CreateDefaultYamlFileForTesting() error {
	path := "./mp3"
	if err := Mkdir(path); err != nil {
		return err
	}
	json := `---
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
    value: 1.25
repair:
    dryRun: true # false`
	return CreateFileForTestingWithContent(path, defaultConfigFileName, []byte(json))
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
	consoleWriter *bytes.Buffer
	errorWriter   *bytes.Buffer
	logWriter     testLogger
}

func NewOutputDeviceForTesting() *OutputDeviceForTesting {
	return &OutputDeviceForTesting{
		consoleWriter: &bytes.Buffer{},
		errorWriter:   &bytes.Buffer{},
		logWriter:     testLogger{writer: &bytes.Buffer{}},
	}
}

func (o *OutputDeviceForTesting) ConsoleWriter() io.Writer {
	return o.consoleWriter
}

func (o *OutputDeviceForTesting) ErrorWriter() io.Writer {
	return o.errorWriter
}

func (o *OutputDeviceForTesting) LogWriter() Logger {
	return o.logWriter
}

type testLogger struct {
	writer *bytes.Buffer
}

func (tl testLogger) Info(msg string, fields map[string]interface{}) {
	tl.log("info", msg, fields)
}

func (tl testLogger) log(level string, msg string, fields map[string]interface{}) {
	var parts []string
	for k, v := range fields {
		parts = append(parts, fmt.Sprintf("%s='%v'", k, v))
	}
	sort.Strings(parts)
	fmt.Fprintf(tl.writer, "level='%s' %s msg='%s'\n", level, strings.Join(parts, " "), msg)
}

func (tl testLogger) Warn(msg string, fields map[string]interface{}) {
	tl.log("warn", msg, fields)
}

func (tl testLogger) Error(msg string, fields map[string]interface{}) {
	tl.log("error", msg, fields)
}

func (o *OutputDeviceForTesting) ConsoleOutput() string {
	return o.consoleWriter.String()
}

func (o *OutputDeviceForTesting) ErrorOutput() string {
	return o.errorWriter.String()
}

func (o *OutputDeviceForTesting) LogOutput() string {
	return o.logWriter.writer.String()
}

type WantedOutput struct {
	WantConsoleOutput string
	WantErrorOutput   string
	WantLogOutput     string
}

func (o *OutputDeviceForTesting) CheckOutput(w WantedOutput) (issues []string, ok bool) {
	ok = true
	if gotConsoleOutput := o.consoleWriter.String(); gotConsoleOutput != w.WantConsoleOutput {
		issues = append(issues, fmt.Sprintf("console output = %q, want %q", gotConsoleOutput, w.WantConsoleOutput))
		ok = false
	}
	if gotErrorOutput := o.errorWriter.String(); gotErrorOutput != w.WantErrorOutput {
		issues = append(issues, fmt.Sprintf("error output = %q, want %q", gotErrorOutput, w.WantErrorOutput))
		ok = false
	}
	if gotLogOutput := o.logWriter.writer.String(); gotLogOutput != w.WantLogOutput {
		issues = append(issues, fmt.Sprintf("log output = %q, want %q", gotLogOutput, w.WantLogOutput))
		ok = false
	}
	return
}
