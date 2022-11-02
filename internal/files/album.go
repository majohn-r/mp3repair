package files

import (
	"io/fs"
	"mp3/internal"
	"mp3/internal/output"
	"path/filepath"

	"github.com/bogem/id3v2/v2"
)

// Album encapsulates information about a music album
type Album struct {
	name              string
	tracks            []*Track
	recordingArtist   *Artist
	path              string
	canonicalGenre    string
	canonicalYear     string
	canonicalTitle    string             // this is what the tracks will record as their album title
	musicCDIdentifier id3v2.UnknownFrame // MCDI frame value
}

func newAlbumFromFile(file fs.DirEntry, artist *Artist) *Album {
	dirName := file.Name()
	return NewAlbum(dirName, artist, artist.subDirectory(dirName))
}

func copyAlbum(a *Album, artist *Artist) *Album {
	a2 := NewAlbum(a.name, artist, a.path)
	for _, t := range a.tracks {
		a2.AddTrack(copyTrack(t, a2))
	}
	a2.canonicalGenre = a.canonicalGenre
	a2.canonicalYear = a.canonicalYear
	a2.canonicalTitle = a.canonicalTitle
	a2.musicCDIdentifier = a.musicCDIdentifier
	return a2
}

// NewAlbum creates a new Album instance
func NewAlbum(title string, artist *Artist, albumPath string) *Album {
	return &Album{name: title, recordingArtist: artist, path: albumPath, canonicalTitle: title}
}

// BackupDirectory gets the path for the album's backup directory
func (a *Album) BackupDirectory() string {
	return a.subDirectory(backupDirName)
}

func (a *Album) contents(o output.Bus) ([]fs.DirEntry, bool) {
	return internal.ReadDirectory(o, a.path)
}

// Name returns the album's name
func (a *Album) Name() string {
	return a.name
}

// RecordingArtistName returns the name of the album's recording artist
func (a *Album) RecordingArtistName() string {
	if a.recordingArtist == nil {
		return ""
	}
	return a.recordingArtist.Name()
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
