package files

import (
	"io/fs"
	"path/filepath"

	"github.com/bogem/id3v2/v2"
)

// Album encapsulates information about a music album
type Album struct {
	tracks []*Track
	path   string
	artist *Artist
	title  string
	// the following fields are recorded in each track's metadata
	canonicalGenre    string
	canonicalTitle    string
	canonicalYear     string
	musicCDIdentifier id3v2.UnknownFrame
}

func NewEmptyAlbum() *Album {
	return &Album{}
}

func (a *Album) GetArtist() *Artist {
	return a.artist
}

func (a *Album) WithTracks(t []*Track) *Album {
	a.tracks = t
	return a
}

func (a *Album) WithArtist(ar *Artist) *Album {
	a.artist = ar
	return a
}

func (a *Album) WithTitle(s string) *Album {
	a.title = s
	return a
}

func (a *Album) WithCanonicalGenre(s string) *Album {
	a.canonicalGenre = s
	return a
}

func (a *Album) WithCanonicalTitle(s string) *Album {
	a.canonicalTitle = s
	return a
}

func (a *Album) WithCanonicalYear(s string) *Album {
	a.canonicalYear = s
	return a
}

func (a *Album) WithMusicCDIdentifier(b []byte) *Album {
	a.musicCDIdentifier = id3v2.UnknownFrame{Body: b}
	return a
}

func NewAlbumFromFile(file fs.DirEntry, ar *Artist) *Album {
	albumName := file.Name()
	return NewAlbum(albumName, ar, ar.subDirectory(albumName))
}

func (a *Album) Copy(ar *Artist, includeTracks bool) *Album {
	a2 := NewAlbum(a.title, ar, a.path)
	if includeTracks {
		for _, t := range a.tracks {
			a2.AddTrack(t.Copy(a2))
		}
	}
	a2.canonicalGenre = a.canonicalGenre
	a2.canonicalYear = a.canonicalYear
	a2.canonicalTitle = a.canonicalTitle
	a2.musicCDIdentifier = a.musicCDIdentifier
	return a2
}

// NewAlbum creates a new Album instance
func NewAlbum(s string, ar *Artist, p string) *Album {
	return &Album{title: s, artist: ar, path: p, canonicalTitle: s}
}

// BackupDirectory gets the path for the album's backup directory
func (a *Album) BackupDirectory() string {
	return a.subDirectory(backupDirName)
}

func (a *Album) Path() string {
	return a.path
}

// Name returns the album's name
func (a *Album) Name() string {
	return a.title
}

// RecordingArtistName returns the name of the album's recording artist
func (a *Album) RecordingArtistName() (s string) {
	if a.artist != nil {
		s = a.artist.Name()
	}
	return
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
