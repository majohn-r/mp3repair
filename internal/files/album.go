package files

import (
	"io/fs"
	"path/filepath"
)

type Album struct {
	Name            string
	Tracks          []*Track
	RecordingArtist *Artist
	Path            string
}

func newAlbum(file fs.FileInfo, artist *Artist) *Album {
	dirName := file.Name()
	return &Album{Name: dirName, RecordingArtist: artist, Path: filepath.Join(artist.Path, dirName)}
}

func copyAlbum(a *Album, artist *Artist) *Album {
	return &Album{Name: a.Name, RecordingArtist: artist, Path: a.Path}
}
