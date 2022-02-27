package files

import (
	"mp3/internal"
	"path/filepath"
)

// NOTE: the functions in this file are strictly for testing purposes. Do not
// call them from production code.

func CreateAllOddArtistsWithEvenAlbums(topDir string) []*Artist {
	var artists []*Artist
	for k := 1; k < 10; k += 2 {
		artistName := internal.CreateArtistName(k)
		artistDir := filepath.Join(topDir, artistName)
		artist := &Artist{Name: artistName}
		for n := 0; n < 10; n += 2 {
			albumName := internal.CreateAlbumName(n)
			albumDir := filepath.Join(artistDir, albumName)
			album := &Album{
				Name:            albumName,
				RecordingArtist: artist,
			}
			for p := 0; p < 10; p++ {
				trackName := internal.CreateTrackName(p)
				name, _, _ := ParseTrackName(trackName, albumName, artistName, DefaultFileExtension)
				track := &Track{
					fullPath:        filepath.Join(albumDir, trackName),
					fileName:        trackName,
					Name:            name,
					TrackNumber:     p,
					ContainingAlbum: album,
				}
				album.Tracks = append(album.Tracks, track)
			}
			artist.Albums = append(artist.Albums, album)
		}
		artists = append(artists, artist)
	}
	return artists
}

func CreateAllArtists(topDir string) []*Artist {
	var artists []*Artist
	for k := 0; k < 10; k++ {
		artistName := internal.CreateArtistName(k)
		artistDir := filepath.Join(topDir, artistName)
		artist := &Artist{Name: artistName}
		for n := 0; n < 10; n++ {
			albumName := internal.CreateAlbumName(n)
			albumDir := filepath.Join(artistDir, albumName)
			album := &Album{
				Name:            albumName,
				RecordingArtist: artist,
			}
			for p := 0; p < 10; p++ {
				trackName := internal.CreateTrackName(p)
				name, _, _ := ParseTrackName(trackName, albumName, artistName, DefaultFileExtension)
				track := &Track{
					fullPath:        filepath.Join(albumDir, trackName),
					fileName:        trackName,
					Name:            name,
					TrackNumber:     p,
					ContainingAlbum: album,
				}
				album.Tracks = append(album.Tracks, track)
			}
			artist.Albums = append(artist.Albums, album)
		}
		artists = append(artists, artist)
	}
	return artists
}