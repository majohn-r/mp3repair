package files

import (
	"mp3/internal"
	"path/filepath"
)

// NOTE: the functions in this file are strictly for testing purposes. Do not
// call them from production code.

func CreateAllOddArtistsWithEvenAlbumsForTesting(topDir string) []*Artist {
	var artists []*Artist
	for k := 1; k < 10; k += 2 {
		artistName := internal.CreateArtistNameForTesting(k)
		artistDir := filepath.Join(topDir, artistName)
		artist := &Artist{Name: artistName, Path: artistDir}
		for n := 0; n < 10; n += 2 {
			albumName := internal.CreateAlbumNameForTesting(n)
			albumDir := filepath.Join(artistDir, albumName)
			album := NewAlbum(albumName, artist, albumDir)
			for p := 0; p < 10; p++ {
				trackName := internal.CreateTrackNameForTesting(p)
				name, _, _ := ParseTrackName(trackName, albumName, artistName, DefaultFileExtension)
				track := &Track{
					Path:            filepath.Join(albumDir, trackName),
					Name:            name,
					TrackNumber:     p,
					ContainingAlbum: album,
					TaggedTrack:     trackUnknownTagsNotRead,
				}
				album.Tracks = append(album.Tracks, track)
			}
			artist.Albums = append(artist.Albums, album)
		}
		artists = append(artists, artist)
	}
	return artists
}

func CreateAllArtistsForTesting(topDir string, addExtras bool) []*Artist {
	var artists []*Artist
	for k := 0; k < 10; k++ {
		artistName := internal.CreateArtistNameForTesting(k)
		artistDir := filepath.Join(topDir, artistName)
		artist := &Artist{Name: artistName, Path: artistDir}
		for n := 0; n < 10; n++ {
			albumName := internal.CreateAlbumNameForTesting(n)
			albumDir := filepath.Join(artistDir, albumName)
			album := NewAlbum(albumName, artist, albumDir)
			for p := 0; p < 10; p++ {
				trackName := internal.CreateTrackNameForTesting(p)
				name, trackNo, _ := ParseTrackName(trackName, albumName, artistName, DefaultFileExtension)
				track := &Track{
					Path:            filepath.Join(albumDir, trackName),
					Name:            name,
					TrackNumber:     trackNo,
					ContainingAlbum: album,
					TaggedTrack:     trackUnknownTagsNotRead,
				}
				album.Tracks = append(album.Tracks, track)
			}
			artist.Albums = append(artist.Albums, album)
		}
		if addExtras {
			albumName := internal.CreateAlbumNameForTesting(999)
			album := NewAlbum(albumName, artist, filepath.Join(artist.Path, albumName))
			artist.Albums = append(artist.Albums, album)
		}
		artists = append(artists, artist)
	}
	if addExtras {
		artistName := internal.CreateArtistNameForTesting(999)
		artist := &Artist{Name: artistName, Path: filepath.Join(topDir, artistName)}
		artists = append(artists, artist)
	}
	return artists
}

// CreateTaggedData creates tagged content. This code is based on reading
// https://id3.org/id3v2.3.0 and on looking at a hex dump of a real mp3 file.
func CreateTaggedData(payload []byte, frames map[string]string) []byte {
	content := make([]byte, 0)
	// block off tag header
	content = append(content, []byte("ID3")...)
	content = append(content, []byte{3, 0, 0, 0, 0, 0, 0}...)
	// add some text frames
	for name, value := range frames {
		content = append(content, makeTextFrame(name, value)...)
	}
	contentLength := len(content) - 10
	factor := 128 * 128 * 128
	for k := 0; k < 4; k++ {
		content[6+k] = byte(contentLength / factor)
		contentLength = contentLength % factor
		factor = factor / 128
	}
	// add payload
	content = append(content, payload...)
	return content
}

func makeTextFrame(id string, content string) []byte {
	frame := make([]byte, 0)
	frame = append(frame, []byte(id)...)
	contentSize := 1 + len(content)
	factor := 256 * 256 * 256
	for k := 0; k < 4; k++ {
		frame = append(frame, byte(contentSize/factor))
		contentSize = contentSize % factor
		factor = factor / 256
	}
	frame = append(frame, []byte{0, 0, 0}...)
	frame = append(frame, []byte(content)...)
	return frame
}
