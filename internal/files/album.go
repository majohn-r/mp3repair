package files

import (
	"io"
	"io/fs"
	"path/filepath"
)

// Album encapsulates information about a music album
type Album struct {
	name            string
	tracks          []*Track
	recordingArtist *Artist
	path            string
}

func newAlbumFromFile(file fs.FileInfo, artist *Artist) *Album {
	dirName := file.Name()
	return NewAlbum(dirName, artist, artist.subDirectory(dirName))
}

func copyAlbum(a *Album, artist *Artist) *Album {
	a2 := NewAlbum(a.name, artist, a.path)
	for _, t := range a.tracks {
		a2.AddTrack(copyTrack(t, a2))
	}
	return a2
}

// NewAlbum creates a new Album instance
func NewAlbum(title string, artist *Artist, albumPath string) *Album {
	return &Album{name: title, recordingArtist: artist, path: albumPath}
}

func (a *Album) BackupDirectory() string {
	return a.subDirectory(backupDirName)
}

// TODO [#77] use OutputBus
func (a *Album) contents(wErr io.Writer) ([]fs.FileInfo, bool) {
	return readDirectory(wErr, a.path)
}

// Name returns the album's name
func (a *Album) Name() string {
	return a.name
}

// RecordingArtistName returns the name of the album's recording artist
func (a *Album) RecordingArtistName() string {
	if a.recordingArtist == nil {
		return ""
	}
	return a.recordingArtist.Name()
}

// AddTrack adds a new track to the album
func (a *Album) AddTrack(t *Track) {
	a.tracks = append(a.tracks, t)
}

// HasTracks returns true if the album has tracks
func (a *Album) HasTracks() bool {
	return len(a.tracks) != 0
}

// Tracks returns the slice of Tracks
func (a *Album) Tracks() []*Track {
	return a.tracks
}

func (a *Album) subDirectory(s string) string {
	return filepath.Join(a.path, s)
}
