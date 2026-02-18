/*
Copyright © 2026 Marc Johnson (marc.johnson27591@gmail.com)
*/
package files

import (
	"io/fs"
	"path/filepath"
	"sort"

	"github.com/bogem/id3v2/v2"
)

// Album encapsulates information about a music album
type Album struct {
	tracks          []*Track
	directory       string
	recordingArtist *Artist
	title           string
	// the following fields are recorded in each track's metadata
	genre          string
	canonicalTitle string
	year           string
	cdIdentifier   id3v2.UnknownFrame
}

// Title returns the album's title
func (a *Album) Title() string { return a.title }

// Directory returns the path representing the album
func (a *Album) Directory() string { return a.directory }

// Tracks returns the album's slice of *Track
func (a *Album) Tracks() []*Track { return a.tracks }

// NewAlbumFromFile creates a new Album primarily from file data
func NewAlbumFromFile(file fs.FileInfo, ar *Artist) *Album {
	albumName := file.Name()
	return AlbumMaker{
		Title:     albumName,
		Artist:    ar,
		Directory: ar.subDirectory(albumName),
	}.NewAlbum(true)
}

func (a *Album) Copy(ar *Artist, includeTracks, addToArtist bool) *Album {
	a2 := AlbumMaker{Title: a.title, Artist: ar, Directory: a.directory}.NewAlbum(addToArtist)
	if includeTracks {
		for _, t := range a.tracks {
			a2.addTrack(t.Copy(a2, false))
		}
	}
	a2.genre = a.genre
	a2.year = a.year
	a2.canonicalTitle = a.canonicalTitle
	a2.cdIdentifier = a.cdIdentifier
	return a2
}

type AlbumMaker struct {
	Title     string
	Artist    *Artist
	Directory string
}

// NewAlbum creates a new Album instance
func (maker AlbumMaker) NewAlbum(addToArtist bool) *Album {
	a := &Album{
		title:           maker.Title,
		recordingArtist: maker.Artist,
		directory:       maker.Directory,
		canonicalTitle:  maker.Title,
	}
	if addToArtist {
		maker.Artist.addAlbum(a)
	}
	return a
}

// BackupDirectory gets the path for the album's backup directory
func (a *Album) BackupDirectory() string {
	return a.subDirectory(backupDirName)
}

// RecordingArtistName returns the name of the album's recording artist
func (a *Album) RecordingArtistName() (s string) {
	if a.recordingArtist != nil {
		s = a.recordingArtist.Name()
	}
	return
}

func (a *Album) addTrack(t *Track) {
	a.tracks = append(a.tracks, t)
}

// HasTracks returns true if the album has tracks
func (a *Album) HasTracks() bool {
	return len(a.tracks) != 0
}

func (a *Album) subDirectory(s string) string {
	return filepath.Join(a.directory, s)
}

// SortAlbums sorts albums by title; if titles match, then by artist name
func SortAlbums(albums []*Album) {
	sort.Slice(albums, func(i, j int) bool {
		if albums[i].Title() == albums[j].Title() {
			return albums[i].RecordingArtistName() < albums[j].RecordingArtistName()
		}
		return albums[i].Title() < albums[j].Title()
	})
}
