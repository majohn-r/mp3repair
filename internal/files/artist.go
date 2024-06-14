package files

import (
	"io/fs"
	"path/filepath"
)

// Artist encapsulates information about a recording artist (a solo performer, a
// duo, a band, etc.)
type Artist struct {
	Albums   []*Album
	Name     string
	FilePath string
	// artist name as recorded in the metadata for each track on each album
	CanonicalName string
}

func NewArtistFromFile(f fs.FileInfo, dir string) *Artist {
	artistName := f.Name()
	return NewArtist(artistName, filepath.Join(dir, artistName))
}

func (a *Artist) Copy() *Artist {
	a2 := NewArtist(a.Name, a.FilePath)
	a2.CanonicalName = a.CanonicalName
	return a2
}

// NewArtist creates a new instance of Artist
func NewArtist(n, p string) *Artist {
	return &Artist{Name: n, FilePath: p, CanonicalName: n}
}

func (a *Artist) subDirectory(s string) string {
	return filepath.Join(a.FilePath, s)
}

// AddAlbum adds an album to the artist's slice of albums
func (a *Artist) AddAlbum(album *Album) {
	a.Albums = append(a.Albums, album)
}

// HasAlbums returns true if there are any albums associated with the artist
func (a *Artist) HasAlbums() bool {
	return len(a.Albums) != 0
}
