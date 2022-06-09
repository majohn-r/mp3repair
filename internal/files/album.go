package files

import (
	"io/fs"
	"path/filepath"
)

type Album struct {
	name            string
	Tracks          []*Track
	recordingArtist *Artist
	path            string
}

func newAlbumFromFile(file fs.FileInfo, artist *Artist) *Album {
	dirName := file.Name()
	return NewAlbum(dirName, artist, filepath.Join(artist.Path, dirName))
}

func copyAlbum(a *Album, artist *Artist) *Album {
	return NewAlbum(a.name, artist, a.path)
}

// NewAlbum creates a new Album instance
func NewAlbum(title string, artist *Artist, albumPath string) *Album {
	return &Album{name: title, recordingArtist: artist, path: albumPath}
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
	return a.recordingArtist.Name
}

// Path returns the name of the album's path
func (a *Album) Path() string {
	return a.path
}
