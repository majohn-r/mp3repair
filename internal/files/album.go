package files

import (
	"io/fs"
	"path/filepath"

	"github.com/bogem/id3v2/v2"
)

// Album encapsulates information about a music album
type Album struct {
	Tracks          []*Track
	FilePath        string
	RecordingArtist *Artist
	Title           string
	// the following fields are recorded in each track's metadata
	CanonicalGenre    string
	CanonicalTitle    string
	CanonicalYear     string
	MusicCDIdentifier id3v2.UnknownFrame
}

func NewAlbumFromFile(file fs.FileInfo, ar *Artist) *Album {
	albumName := file.Name()
	return AlbumMaker{
		Title:  albumName,
		Artist: ar,
		Path:   ar.subDirectory(albumName),
	}.NewAlbum()
}

func (a *Album) Copy(ar *Artist, includeTracks bool) *Album {
	a2 := AlbumMaker{Title: a.Title, Artist: ar, Path: a.FilePath}.NewAlbum()
	if includeTracks {
		for _, t := range a.Tracks {
			a2.AddTrack(t.Copy(a2))
		}
	}
	a2.CanonicalGenre = a.CanonicalGenre
	a2.CanonicalYear = a.CanonicalYear
	a2.CanonicalTitle = a.CanonicalTitle
	a2.MusicCDIdentifier = a.MusicCDIdentifier
	return a2
}

type AlbumMaker struct {
	Title  string
	Artist *Artist
	Path   string
}

// NewAlbum creates a new Album instance
func (maker AlbumMaker) NewAlbum() *Album {
	return &Album{
		Title:           maker.Title,
		RecordingArtist: maker.Artist,
		FilePath:        maker.Path,
		CanonicalTitle:  maker.Title}
}

// BackupDirectory gets the path for the album's backup directory
func (a *Album) BackupDirectory() string {
	return a.subDirectory(backupDirName)
}

// RecordingArtistName returns the name of the album's recording artist
func (a *Album) RecordingArtistName() (s string) {
	if a.RecordingArtist != nil {
		s = a.RecordingArtist.Name
	}
	return
}

// AddTrack adds a new track to the album
func (a *Album) AddTrack(t *Track) {
	a.Tracks = append(a.Tracks, t)
}

// HasTracks returns true if the album has tracks
func (a *Album) HasTracks() bool {
	return len(a.Tracks) != 0
}

func (a *Album) subDirectory(s string) string {
	return filepath.Join(a.FilePath, s)
}
