package files

import (
	"io/fs"
	"path/filepath"

	"github.com/bogem/id3v2/v2"
)

// Album encapsulates information about a music album
type Album struct {
	Title           string
	Contents        []*Track
	RecordingArtist *Artist
	path            string
	// the following fields are recorded in each track's metadata
	CanonicalGenre    string
	CanonicalYear     string
	CanonicalTitle    string
	MusicCDIdentifier id3v2.UnknownFrame
}

func NewAlbumFromFile(file fs.DirEntry, ar *Artist) *Album {
	albumName := file.Name()
	return NewAlbum(albumName, ar, ar.subDirectory(albumName))
}

func (a *Album) Copy(ar *Artist, includeTracks bool) *Album {
	a2 := NewAlbum(a.Title, ar, a.path)
	if includeTracks {
		for _, t := range a.Contents {
			a2.AddTrack(t.Copy(a2))
		}
	}
	a2.CanonicalGenre = a.CanonicalGenre
	a2.CanonicalYear = a.CanonicalYear
	a2.CanonicalTitle = a.CanonicalTitle
	a2.MusicCDIdentifier = a.MusicCDIdentifier
	return a2
}

// NewAlbum creates a new Album instance
func NewAlbum(s string, ar *Artist, p string) *Album {
	return &Album{Title: s, RecordingArtist: ar, path: p, CanonicalTitle: s}
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
	return a.Title
}

// RecordingArtistName returns the name of the album's recording artist
func (a *Album) RecordingArtistName() (s string) {
	if a.RecordingArtist != nil {
		s = a.RecordingArtist.Name()
	}
	return
}

// AddTrack adds a new track to the album
func (a *Album) AddTrack(t *Track) {
	a.Contents = append(a.Contents, t)
}

// HasTracks returns true if the album has tracks
func (a *Album) HasTracks() bool {
	return len(a.Contents) != 0
}

// Tracks returns the slice of Tracks
func (a *Album) Tracks() []*Track {
	return a.Contents
}

func (a *Album) subDirectory(s string) string {
	return filepath.Join(a.path, s)
}
