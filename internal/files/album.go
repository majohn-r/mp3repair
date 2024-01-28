package files

import (
	"io/fs"
	"path/filepath"

	"github.com/bogem/id3v2/v2"
)

// Album encapsulates information about a music album
type Album struct {
	name            string
	tracks          []*Track
	recordingArtist *Artist
	path            string
	// the following fields are recorded in each track's metadata
	canonicalGenre    string
	canonicalYear     string
	canonicalTitle    string
	musicCDIdentifier id3v2.UnknownFrame
}

func NewAlbumFromFile(file fs.DirEntry, ar *Artist) *Album {
	albumName := file.Name()
	return NewAlbum(albumName, ar, ar.subDirectory(albumName))
}

func (a *Album) Copy(ar *Artist, includeTracks bool) *Album {
	a2 := NewAlbum(a.name, ar, a.path)
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
	return &Album{name: s, recordingArtist: ar, path: p, canonicalTitle: s}
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
	return a.name
}

// RecordingArtistName returns the name of the album's recording artist
func (a *Album) RecordingArtistName() (s string) {
	if a.recordingArtist != nil {
		s = a.recordingArtist.Name()
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
