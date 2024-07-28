package files

import (
	"io/fs"
	"path/filepath"
)

// Artist encapsulates information about a recording artist (a solo performer, a
// duo, a band, etc.)
type Artist struct {
	albums     []*Album
	name       string
	directory  string
	sharedName string
}

func (a *Artist) canonicalName() string { return a.sharedName }

func (a *Artist) Directory() string { return a.directory }

func (a *Artist) Name() string { return a.name }

func (a *Artist) Albums() []*Album { return a.albums }

func NewArtistFromFile(f fs.FileInfo, dir string) *Artist {
	artistName := f.Name()
	return NewArtist(artistName, filepath.Join(dir, artistName))
}

func (a *Artist) Copy() *Artist {
	a2 := NewArtist(a.name, a.directory)
	a2.sharedName = a.sharedName
	return a2
}

// NewArtist creates a new instance of Artist
func NewArtist(name, directory string) *Artist {
	return &Artist{name: name, directory: directory, sharedName: name}
}

func (a *Artist) subDirectory(s string) string {
	return filepath.Join(a.directory, s)
}

func (a *Artist) addAlbum(album *Album) {
	a.albums = append(a.albums, album)
}

// HasAlbums returns true if there are any albums associated with the artist
func (a *Artist) HasAlbums() bool {
	return len(a.albums) != 0
}
