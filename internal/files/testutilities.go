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
		artist := &Artist{Name: artistName}
		for n := 0; n < 10; n += 2 {
			albumName := internal.CreateAlbumNameForTesting(n)
			albumDir := filepath.Join(artistDir, albumName)
			album := &Album{
				Name:            albumName,
				RecordingArtist: artist,
			}
			for p := 0; p < 10; p++ {
				trackName := internal.CreateTrackNameForTesting(p)
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

func CreateAllArtistsForTesting(topDir string, addExtras bool) []*Artist {
	var artists []*Artist
	for k := 0; k < 10; k++ {
		artistName := internal.CreateArtistNameForTesting(k)
		artistDir := filepath.Join(topDir, artistName)
		artist := &Artist{Name: artistName}
		for n := 0; n < 10; n++ {
			albumName := internal.CreateAlbumNameForTesting(n)
			albumDir := filepath.Join(artistDir, albumName)
			album := &Album{
				Name:            albumName,
				RecordingArtist: artist,
			}
			for p := 0; p < 10; p++ {
				trackName := internal.CreateTrackNameForTesting(p)
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
		if addExtras {
			albumName := internal.CreateAlbumNameForTesting(999)
			album := &Album{
				Name:            albumName,
				RecordingArtist: artist,
			}
			artist.Albums = append(artist.Albums, album)
		}
		artists = append(artists, artist)
	}
	if addExtras {
		artistName := internal.CreateArtistNameForTesting(999)
		artist := &Artist{Name: artistName}
		artists = append(artists, artist)
	}
	return artists
}