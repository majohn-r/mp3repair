package files

import (
	"io/fs"
	"io/ioutil"
	"path/filepath"
	"regexp"

	"github.com/sirupsen/logrus"
)

type Search struct {
	topDirectory    string
	targetExtension string
	albumFilter     *regexp.Regexp
	artistFilter    *regexp.Regexp
}

func (s *Search) TopDirectory() string {
	return s.topDirectory
}

func (s *Search) TargetExtension() string {
	return s.targetExtension
}

func (s *Search) LoadUnfilteredData() (artists []*Artist) {
	logrus.WithFields(logrus.Fields{
		"topDirectory":  s.topDirectory,
		"fileExtension": s.targetExtension,
	}).Info("load raw data from file system")
	// read top directory
	artistFiles, err := readDirectory(s.topDirectory)
	if err == nil {
		for _, artistFile := range artistFiles {
			// we only care about directories, which correspond to artists
			if artistFile.IsDir() {
				artist := &Artist{
					Name: artistFile.Name(),
				}
				// look for albums for the current artist
				artistDir := filepath.Join(s.topDirectory, artistFile.Name())
				albumFiles, err := readDirectory(artistDir)
				if err == nil {
					for _, albumFile := range albumFiles {
						// skip over non-directories
						if !albumFile.IsDir() {
							continue
						}
						album := &Album{
							Name:            albumFile.Name(),
							RecordingArtist: artist,
						}
						// look for tracks in the current album
						albumDir := filepath.Join(artistDir, album.Name)
						trackFiles, err := readDirectory(albumDir)
						if err == nil {
							// process tracks
							for _, trackFile := range trackFiles {
								if trackFile.IsDir() || !trackNameRegex.MatchString(trackFile.Name()) {
									continue
								}
								if simpleName, trackNumber, valid := ParseTrackName(trackFile.Name(), album.Name, artist.Name, s.targetExtension); valid {
									track := &Track{
										fullPath:        filepath.Join(albumDir, trackFile.Name()),
										fileName:        trackFile.Name(),
										Name:            simpleName,
										TrackNumber:     trackNumber,
										ContainingAlbum: album,
									}
									album.Tracks = append(album.Tracks, track)
								}
							}
						}
						artist.Albums = append(artist.Albums, album)
					}
				}
				artists = append(artists, artist)
			}
		}
	}
	return
}

func (s *Search) FilterArtists(unfilteredArtists []*Artist) (artists []*Artist) {
	logrus.WithFields(logrus.Fields{
		"topDirectory":  s.topDirectory,
		"fileExtension": s.targetExtension,
		"albumFilter":   s.albumFilter,
		"artistFilter":  s.artistFilter,
	}).Info("filter artists")
	for _, unfilteredArtist := range unfilteredArtists {
		if s.artistFilter.MatchString(unfilteredArtist.Name) {
			artist := &Artist{
				Name: unfilteredArtist.Name,
			}
			for _, album := range unfilteredArtist.Albums {
				if s.albumFilter.MatchString(album.Name) {
					if len(album.Tracks) != 0 {
						newAlbum := &Album{
							Name:            album.Name,
							RecordingArtist: artist,
						}
						for _, track := range album.Tracks {
							newTrack := &Track{
								fullPath:        track.fullPath,
								fileName:        track.fileName,
								Name:            track.Name,
								TrackNumber:     track.TrackNumber,
								ContainingAlbum: newAlbum,
							}
							newAlbum.Tracks = append(newAlbum.Tracks, newTrack)
						}
						artist.Albums = append(artist.Albums, newAlbum)
					}
				}
			}
			if len(artist.Albums) != 0 {
				artists = append(artists, artist)
			}
		}
	}
	return
}

func (s *Search) LoadData() (artists []*Artist) {
	logrus.WithFields(logrus.Fields{
		"topDirectory":  s.topDirectory,
		"fileExtension": s.targetExtension,
		"albumFilter":   s.albumFilter,
		"artistFilter":  s.artistFilter,
	}).Info("load data from file system")
	// read top directory
	artistFiles, err := readDirectory(s.topDirectory)
	if err == nil {
		for _, artistFile := range artistFiles {
			// we only care about directories, which correspond to artists
			if !artistFile.IsDir() || !s.artistFilter.MatchString(artistFile.Name()) {
				continue
			}
			artist := &Artist{
				Name: artistFile.Name(),
			}
			// look for albums for the current artist
			artistDir := filepath.Join(s.topDirectory, artistFile.Name())
			albumFiles, err := readDirectory(artistDir)
			if err == nil {
				for _, albumFile := range albumFiles {
					// skip over non-directories or directories whose name does not match the album filter
					if !albumFile.IsDir() || !s.albumFilter.MatchString(albumFile.Name()) {
						continue
					}
					album := &Album{
						Name:            albumFile.Name(),
						RecordingArtist: artist,
					}
					// look for tracks in the current album
					albumDir := filepath.Join(artistDir, album.Name)
					trackFiles, err := readDirectory(albumDir)
					if err == nil {
						// process tracks
						for _, trackFile := range trackFiles {
							if trackFile.IsDir() || !trackNameRegex.MatchString(trackFile.Name()) {
								continue
							}
							if simpleName, trackNumber, valid := ParseTrackName(trackFile.Name(), album.Name, artist.Name, s.targetExtension); valid {
								track := &Track{
									fullPath:        filepath.Join(albumDir, trackFile.Name()),
									fileName:        trackFile.Name(),
									Name:            simpleName,
									TrackNumber:     trackNumber,
									ContainingAlbum: album,
								}
								album.Tracks = append(album.Tracks, track)
							}
						}
					}
					if len(album.Tracks) != 0 {
						artist.Albums = append(artist.Albums, album)
					}
				}
			}
			if len(artist.Albums) != 0 {
				artists = append(artists, artist)
			}
		}
	}
	return
}

func readDirectory(dir string) (files []fs.FileInfo, err error) {
	files, err = ioutil.ReadDir(dir)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"directory": dir,
			"error":     err,
		}).Error("problem reading directory")
	}
	return
}
