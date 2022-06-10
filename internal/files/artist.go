package files

import (
	"io/fs"
	"path/filepath"
)

// Artist encapsulates information about a recording artist (a solo performer, a
// duo, a band, etc.)
type Artist struct {
	name   string
	Albums []*Album
	path   string
}

func newArtistFromFile(f fs.FileInfo, dir string) *Artist {
	artistName := f.Name()
	return NewArtist(artistName, filepath.Join(dir, artistName))
}

func copyArtist(a *Artist) *Artist {
	return NewArtist(a.name, a.path)
}

// NewArtist creates a new instance of Artist
func NewArtist(n, p string) *Artist {
	return &Artist{name: n, path: p}
}

// Name returns the artist's name
func (a *Artist) Name() string {
	return a.name
}

// Path returns the artist's underlying path
func (a *Artist) Path() string {
	return a.path
}