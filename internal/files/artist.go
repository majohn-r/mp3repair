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

func (a *Artist) contents() ([]fs.FileInfo, error) {
	return readDirectory(a.path)
}

// Name returns the artist's name
func (a *Artist) Name() string {
	return a.name
}

func (a *Artist) subDirectory(s string) string {
	return filepath.Join(a.path, s)
}
