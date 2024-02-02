package files

import (
	"io/fs"
	"path/filepath"
)

// Artist encapsulates information about a recording artist (a solo performer, a
// duo, a band, etc.)
type Artist struct {
	FileName string
	Contents []*Album
	path     string
	// artist name as recorded in the metadata for each track in each album
	CanonicalName string
}

func NewArtistFromFile(f fs.DirEntry, dir string) *Artist {
	artistName := f.Name()
	return NewArtist(artistName, filepath.Join(dir, artistName))
}

func (a *Artist) Copy() *Artist {
	a2 := NewArtist(a.FileName, a.Path())
	a2.CanonicalName = a.CanonicalName
	return a2
}

// NewArtist creates a new instance of Artist
func NewArtist(n, p string) *Artist {
	return &Artist{FileName: n, path: p, CanonicalName: n}
}

func (a *Artist) Path() string {
	return a.path
}

// Name returns the artist's name
func (a *Artist) Name() string {
	return a.FileName
}

func (a *Artist) subDirectory(s string) string {
	return filepath.Join(a.path, s)
}

// AddAlbum adds an album to the artist's slice of albums
func (a *Artist) AddAlbum(album *Album) {
	a.Contents = append(a.Contents, album)
}

// Albums returns the slice of this artist's albums
func (a *Artist) Albums() []*Album {
	return a.Contents
}

// HasAlbums returns true if there any albums associated with the artist
func (a *Artist) HasAlbums() bool {
	return len(a.Contents) != 0
}
