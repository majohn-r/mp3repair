package files

import (
	"io"
	"io/fs"
	"path/filepath"
)

// Artist encapsulates information about a recording artist (a solo performer, a
// duo, a band, etc.)
type Artist struct {
	name   string
	albums []*Album
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

// TODO [#77] use OutputBus
func (a *Artist) contents(wErr io.Writer) ([]fs.FileInfo, bool) {
	return readDirectory(wErr, a.path)
}

// Name returns the artist's name
func (a *Artist) Name() string {
	return a.name
}

func (a *Artist) subDirectory(s string) string {
	return filepath.Join(a.path, s)
}

// AddAlbum adds an album to the artist's slice of albums
func (a *Artist) AddAlbum(album *Album) {
	a.albums = append(a.albums, album)
}

// Albums returns the slice of this artist's albums
func (a *Artist) Albums() []*Album {
	return a.albums
}

// HasAlbums returns true if there any albums associated with the artist
func (a *Artist) HasAlbums() bool {
	return len(a.albums) != 0
}
