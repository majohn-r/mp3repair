package files

import (
	"io/fs"
	"path/filepath"

	"github.com/bogem/id3v2/v2"
	cmd "github.com/majohn-r/cmd-toolkit"
	"github.com/majohn-r/output"
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

func newAlbumFromFile(file fs.DirEntry, ar *Artist) *Album {
	albumName := file.Name()
	return NewAlbum(albumName, ar, ar.subDirectory(albumName))
}

func (a *Album) copy(ar *Artist) *Album {
	a2 := NewAlbum(a.name, ar, a.path)
	for _, t := range a.tracks {
		a2.AddTrack(t.copy(a2))
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

func (a *Album) contents(o output.Bus) ([]fs.DirEntry, bool) {
	return cmd.ReadDirectory(o, a.path)
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
