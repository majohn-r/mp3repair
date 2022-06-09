package files

import (
	"io/fs"
	"path/filepath"
)

type Album struct {
	name            string
	Tracks          []*Track
	RecordingArtist *Artist
	Path            string
}

func newAlbumFromFile(file fs.FileInfo, artist *Artist) *Album {
	dirName := file.Name()
	return NewAlbum(dirName, artist, filepath.Join(artist.Path, dirName))
}

func copyAlbum(a *Album, artist *Artist) *Album {
	return &Album{name: a.name, RecordingArtist: artist, Path: a.Path}
}

func NewAlbum(title string, artist *Artist, path string) *Album {
	return &Album{name: title, RecordingArtist: artist, Path: path}
}

func (a *Album) Name() string {
	return a.name
}
