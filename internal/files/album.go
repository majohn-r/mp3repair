package files

import (
	"io/fs"
	"path/filepath"
)

// Album encapsulates information about a music album
type Album struct {
	name            string
	tracks          []*Track
	recordingArtist *Artist
	path            string
}

func newAlbumFromFile(file fs.FileInfo, artist *Artist) *Album {
	dirName := file.Name()
	return NewAlbum(dirName, artist, filepath.Join(artist.Path, dirName))
}

func copyAlbum(a *Album, artist *Artist) *Album {
	a2 := NewAlbum(a.name, artist, a.path)
	for _, t := range a.tracks {
		a2.AddTrack(&Track{
			Path:            t.Path,
			Name:            t.Name,
			TrackNumber:     t.TrackNumber,
			TaggedAlbum:     t.TaggedAlbum,
			TaggedArtist:    t.TaggedArtist,
			TaggedTitle:     t.TaggedTitle,
			TaggedTrack:     t.TaggedTrack,
			ContainingAlbum: a2,
		})
	}
	return a2
}

// NewAlbum creates a new Album instance
func NewAlbum(title string, artist *Artist, albumPath string) *Album {
	return &Album{name: title, recordingArtist: artist, path: albumPath}
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

// Path returns the name of the album's path
func (a *Album) Path() string {
	return a.path
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
